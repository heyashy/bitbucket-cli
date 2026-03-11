package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/heyashy/bitbucket-cli/internal/bitbucket"
	"github.com/heyashy/bitbucket-cli/internal/cmd/resolve"
	"github.com/heyashy/bitbucket-cli/internal/git"
)

func newAPICmd() *cobra.Command {
	var (
		method      string
		body        string
		queryParams []string
		fields      []string
	)

	cmd := &cobra.Command{
		Use:   "api <endpoint>",
		Short: "Make an authenticated request to the Bitbucket Cloud API",
		Long: `Make an authenticated request to any Bitbucket Cloud REST API v2.0 endpoint.

The base URL (https://api.bitbucket.org/2.0) is prepended automatically.
Use {workspace} and {repo} as placeholders — they are replaced with values
detected from the current git remote.

METHODS
  Default is GET. Use -X to specify POST, PUT, DELETE, PATCH.

BODY
  Use --body for raw JSON, or -f key=value to build a JSON object from fields.

QUERY PARAMETERS
  Use -q key=value to add query string parameters (repeatable).

EXAMPLES
  bb api /user
  bb api /repositories/{workspace}/{repo}
  bb api /repositories/{workspace}/{repo}/pullrequests
  bb api /repositories/{workspace}/{repo}/pullrequests -q state=OPEN
  bb api /repositories/{workspace}/{repo}/pullrequests/42
  bb api /repositories/{workspace}/{repo}/pullrequests/42/comments
  bb api /repositories/{workspace}/{repo}/pipelines
  bb api -X POST /repositories/{workspace}/{repo}/pullrequests/42/approve
  bb api -X POST /repositories/{workspace}/{repo}/pullrequests \
    -f title="My PR" -f source.branch.name=feature/x -f destination.branch.name=main
  bb api -X POST /repositories/{workspace}/{repo}/pullrequests/42/comments \
    --body '{"content":{"raw":"LGTM"}}'

COMMON ENDPOINTS
  /user                                          Current authenticated user
  /repositories/{workspace}/{repo}               Repository details
  /repositories/{workspace}/{repo}/pullrequests  List pull requests
  /repositories/{workspace}/{repo}/pullrequests/{id}           PR details
  /repositories/{workspace}/{repo}/pullrequests/{id}/diff      PR diff
  /repositories/{workspace}/{repo}/pullrequests/{id}/comments  PR comments
  /repositories/{workspace}/{repo}/pullrequests/{id}/approve   Approve PR (POST)
  /repositories/{workspace}/{repo}/pullrequests/{id}/merge     Merge PR (POST)
  /repositories/{workspace}/{repo}/pullrequests/{id}/decline   Decline PR (POST)
  /repositories/{workspace}/{repo}/pipelines                   List pipelines
  /repositories/{workspace}/{repo}/commit/{sha}                Commit details
  /repositories/{workspace}/{repo}/refs/branches               List branches
  /repositories/{workspace}/{repo}/refs/tags                   List tags
  /repositories/{workspace}/{repo}/src/{sha}/{path}            File contents

Full API reference: https://developer.atlassian.com/cloud/bitbucket/rest/intro/`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint := args[0]

			// Ensure leading slash
			if !strings.HasPrefix(endpoint, "/") {
				endpoint = "/" + endpoint
			}

			// Replace {workspace} and {repo} placeholders
			endpoint = replacePlaceholders(endpoint)

			authProvider, err := resolve.AuthProvider()
			if err != nil {
				return err
			}

			client := bitbucket.NewClient(authProvider)
			ctx := context.Background()

			// Build query params
			var query url.Values
			if len(queryParams) > 0 {
				query = url.Values{}
				for _, q := range queryParams {
					parts := strings.SplitN(q, "=", 2)
					if len(parts) == 2 {
						query.Set(parts[0], parts[1])
					}
				}
			}

			// Build body from -f fields
			var requestBody interface{}
			if body != "" {
				var parsed interface{}
				if err := json.Unmarshal([]byte(body), &parsed); err != nil {
					return fmt.Errorf("validation: --body must be valid JSON: %w", err)
				}
				requestBody = parsed
			} else if len(fields) > 0 {
				requestBody = buildNestedJSON(fields)
			}

			var resp *bitbucket.RawResponse
			switch strings.ToUpper(method) {
			case "GET":
				httpResp, err := client.Get(ctx, endpoint, query)
				if err != nil {
					return err
				}
				resp = &bitbucket.RawResponse{Response: httpResp}
			case "POST":
				httpResp, err := client.Post(ctx, endpoint, requestBody)
				if err != nil {
					return err
				}
				resp = &bitbucket.RawResponse{Response: httpResp}
			case "PUT":
				httpResp, err := client.Put(ctx, endpoint, requestBody)
				if err != nil {
					return err
				}
				resp = &bitbucket.RawResponse{Response: httpResp}
			case "DELETE":
				httpResp, err := client.Delete(ctx, endpoint)
				if err != nil {
					return err
				}
				resp = &bitbucket.RawResponse{Response: httpResp}
			default:
				return fmt.Errorf("validation: unsupported method: %s", method)
			}

			defer resp.Response.Body.Close()

			respBody, err := io.ReadAll(resp.Response.Body)
			if err != nil {
				return fmt.Errorf("infra: cannot read response: %w", err)
			}

			// Pretty-print JSON if possible
			if len(respBody) > 0 {
				var pretty json.RawMessage
				if json.Unmarshal(respBody, &pretty) == nil {
					formatted, err := json.MarshalIndent(pretty, "", "  ")
					if err == nil {
						fmt.Println(string(formatted))
						return nil
					}
				}
				// Not JSON, print raw
				fmt.Print(string(respBody))
			}

			if resp.Response.StatusCode >= 400 {
				return fmt.Errorf("API returned HTTP %d", resp.Response.StatusCode)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&method, "method", "X", "GET", "HTTP method: GET, POST, PUT, DELETE")
	cmd.Flags().StringVar(&body, "body", "", "Raw JSON request body")
	cmd.Flags().StringSliceVarP(&queryParams, "query", "q", nil, "Query parameters (key=value, repeatable)")
	cmd.Flags().StringSliceVarP(&fields, "field", "f", nil, "JSON body fields (key=value, supports nesting: source.branch.name=main)")

	return cmd
}

func replacePlaceholders(endpoint string) string {
	if !strings.Contains(endpoint, "{workspace}") && !strings.Contains(endpoint, "{repo}") {
		return endpoint
	}

	info, err := git.DetectRepo()
	if err != nil {
		return endpoint
	}

	endpoint = strings.ReplaceAll(endpoint, "{workspace}", info.Workspace)
	endpoint = strings.ReplaceAll(endpoint, "{repo}", info.RepoSlug)
	return endpoint
}

// buildNestedJSON builds a nested map from key=value pairs.
// e.g. "source.branch.name=main" becomes {"source":{"branch":{"name":"main"}}}
func buildNestedJSON(fields []string) map[string]interface{} {
	result := make(map[string]interface{})

	for _, f := range fields {
		parts := strings.SplitN(f, "=", 2)
		if len(parts) != 2 {
			continue
		}

		keys := strings.Split(parts[0], ".")
		value := parts[1]

		current := result
		for i, key := range keys {
			if i == len(keys)-1 {
				current[key] = value
			} else {
				if _, ok := current[key]; !ok {
					current[key] = make(map[string]interface{})
				}
				current = current[key].(map[string]interface{})
			}
		}
	}

	return result
}
