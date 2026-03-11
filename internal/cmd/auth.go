package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/heyashy/bitbucket-cli/internal/auth"
	"github.com/heyashy/bitbucket-cli/internal/bitbucket"
	"github.com/heyashy/bitbucket-cli/internal/cmd/resolve"
	"github.com/heyashy/bitbucket-cli/internal/config"
	"github.com/heyashy/bitbucket-cli/internal/ui"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication with Bitbucket",
		Long: `Manage authentication credentials for the Bitbucket Cloud API.

bb supports two authentication methods:

  API Token (default):
    Interactive login that prompts for your registered email and API token.
    Tokens are created at: https://id.atlassian.com/manage-profile/security/api-tokens
    Select "Bitbucket" as the product and grant pull request + repository scopes.
    Credentials are stored at ~/.config/bb/tokens.json (0600 permissions).

  OAuth (browser flow):
    Opens your browser for Bitbucket OAuth consent. Requires an OAuth consumer
    to be configured first (bb config set --global auth.oauth.client_id ...).

EXAMPLES
  bb auth login              Log in with an API token (interactive prompt)
  bb auth login --oauth      Log in via OAuth browser flow
  bb auth status             Check if credentials are valid
  bb auth logout             Remove stored credentials`,
	}

	cmd.AddCommand(newAuthLoginCmd())
	cmd.AddCommand(newAuthLogoutCmd())
	cmd.AddCommand(newAuthStatusCmd())

	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	var useOAuth bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to Bitbucket",
		Long:  "Authenticate with Bitbucket using an API token (default) or OAuth (browser flow).",
		RunE: func(cmd *cobra.Command, args []string) error {
			if useOAuth {
				return loginWithOAuth()
			}
			return loginWithAPIToken()
		},
	}

	cmd.Flags().BoolVar(&useOAuth, "oauth", false, "Use OAuth browser flow instead of API token")

	return cmd
}

func loginWithAPIToken() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println(ui.Faint.Render("  Create an API token at: https://id.atlassian.com/manage-profile/security/api-tokens"))
	fmt.Println(ui.Faint.Render("  Select 'Bitbucket' as the product and grant the scopes you need."))
	fmt.Println()
	fmt.Print(ui.Info.Render("Registered email: "))
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print(ui.Info.Render("API token: "))
	token, _ := reader.ReadString('\n')
	token = strings.TrimSpace(token)

	if email == "" || token == "" {
		return fmt.Errorf("validation: email and API token are required")
	}

	stored := &auth.StoredToken{
		AccessToken: email + ":" + token,
		TokenType:   "apitoken",
	}

	if err := auth.SaveToken(stored); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	cfg.Set("auth.method", "apitoken")
	if err := cfg.WriteGlobal(); err != nil {
		return err
	}

	fmt.Println(ui.Success.Render("  Logged in successfully with API token."))
	return nil
}

func loginWithOAuth() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("cannot load config: %w", err)
	}

	clientID := cfg.OAuthClientID()
	clientSecret := cfg.OAuthClientSecret()

	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("validation: OAuth not configured — set auth.oauth.client_id and auth.oauth.client_secret in config\nRun: bb config set --global auth.oauth.client_id YOUR_CLIENT_ID\nRun: bb config set --global auth.oauth.client_secret YOUR_CLIENT_SECRET")
	}

	provider := auth.NewOAuthProvider(clientID, clientSecret)
	return provider.Login()
}

func newAuthLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Log out of Bitbucket",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := auth.DeleteToken(); err != nil {
				return err
			}
			fmt.Println(ui.Success.Render("  Logged out successfully."))
			return nil
		},
	}
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status and verify credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := auth.LoadToken()
			if err != nil {
				fmt.Println(ui.Warning.Render("  Not logged in."))
				return nil
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			method := cfg.AuthMethod()
			fmt.Println(ui.Title.Render("Authentication Status"))
			fmt.Println(ui.Divider.String())
			fmt.Printf("%s %s\n", ui.Key.Render("Method:"), ui.Value.Render(method))

			switch token.TokenType {
			case "apitoken":
				parts := strings.SplitN(token.AccessToken, ":", 2)
				if len(parts) == 2 {
					fmt.Printf("%s %s\n", ui.Key.Render("Email:"), ui.Value.Render(parts[0]))
				}
			case "apppassword":
				parts := strings.SplitN(token.AccessToken, ":", 2)
				if len(parts) == 2 {
					fmt.Printf("%s %s\n", ui.Key.Render("Username:"), ui.Value.Render(parts[0]))
				}
			default:
				if token.IsExpired() {
					fmt.Printf("%s %s\n", ui.Key.Render("Status:"), ui.Warning.Render("token expired (will auto-refresh)"))
				} else {
					fmt.Printf("%s %s\n", ui.Key.Render("Status:"), ui.Success.Render("authenticated (OAuth)"))
				}
				if token.Scopes != "" {
					fmt.Printf("%s %s\n", ui.Key.Render("Scopes:"), token.Scopes)
				}
			}

			fmt.Println(ui.Divider.String())
			fmt.Printf("%s Verifying credentials... ", ui.Spinner())

			provider, err := resolve.AuthProvider()
			if err != nil {
				fmt.Println(ui.CrossMark())
				return err
			}

			client := bitbucket.NewClient(provider)
			resp, err := client.Get(cmd.Context(), "/user", nil)
			if err != nil {
				fmt.Println(ui.CrossMark())
				return fmt.Errorf("cannot reach Bitbucket API: %w", err)
			}

			var user bitbucket.User
			if err := bitbucket.DecodeResponse(resp, &user); err != nil {
				fmt.Println(ui.CrossMark())
				return err
			}

			fmt.Println(ui.CheckMark())
			fmt.Printf("%s %s (%s)\n", ui.Key.Render("Logged in as:"), ui.Value.Render(user.DisplayName), ui.PRAuthor.Render(user.Nickname))
			return nil
		},
	}
}
