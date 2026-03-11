package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/heyashy/bb/internal/bitbucket"
	"github.com/heyashy/bb/internal/cmd/resolve"
)

func newPRCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pr",
		Short: "Manage pull requests",
	}

	cmd.AddCommand(newPRCreateCmd())
	cmd.AddCommand(newPRListCmd())
	cmd.AddCommand(newPRViewCmd())
	cmd.AddCommand(newPRMergeCmd())
	cmd.AddCommand(newPRApproveCmd())
	cmd.AddCommand(newPRDeclineCmd())
	cmd.AddCommand(newPRCommentCmd())
	cmd.AddCommand(newPRDiffCmd())

	return cmd
}

func newPRCreateCmd() *cobra.Command {
	var (
		title             string
		description       string
		source            string
		destination       string
		closeSourceBranch bool
		reviewers         []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a pull request",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := resolve.PRService()
			if err != nil {
				return err
			}

			if source == "" {
				source, err = currentBranch()
				if err != nil {
					return fmt.Errorf("cannot detect current branch: %w", err)
				}
			}

			if destination == "" {
				destination = "main"
			}

			if title == "" {
				return fmt.Errorf("validation: --title is required")
			}

			req := bitbucket.CreatePRRequest{
				Title:             title,
				Description:       description,
				Source:            bitbucket.Ref{Branch: bitbucket.Branch{Name: source}},
				Destination:       bitbucket.Ref{Branch: bitbucket.Branch{Name: destination}},
				CloseSourceBranch: closeSourceBranch,
			}

			for _, r := range reviewers {
				req.Reviewers = append(req.Reviewers, bitbucket.User{UUID: r})
			}

			pr, err := svc.Create(context.Background(), req)
			if err != nil {
				return err
			}

			fmt.Printf("Created PR #%d: %s\n", pr.ID, pr.Title)
			fmt.Printf("%s → %s\n", pr.Source.Branch.Name, pr.Destination.Branch.Name)
			if pr.Links.HTML != nil {
				fmt.Printf("URL: %s\n", pr.Links.HTML.Href)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&title, "title", "t", "", "PR title (required)")
	cmd.Flags().StringVarP(&description, "body", "b", "", "PR description")
	cmd.Flags().StringVarP(&source, "source", "s", "", "Source branch (default: current branch)")
	cmd.Flags().StringVarP(&destination, "dest", "d", "", "Destination branch (default: main)")
	cmd.Flags().BoolVar(&closeSourceBranch, "close-branch", false, "Close source branch after merge")
	cmd.Flags().StringSliceVarP(&reviewers, "reviewer", "r", nil, "Reviewer UUIDs")

	return cmd
}

func newPRListCmd() *cobra.Command {
	var state string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pull requests",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := resolve.PRService()
			if err != nil {
				return err
			}

			opts := bitbucket.ListPRsOpts{State: strings.ToUpper(state)}
			result, err := svc.List(context.Background(), opts)
			if err != nil {
				return err
			}

			if len(result.Values) == 0 {
				fmt.Println("No pull requests found.")
				return nil
			}

			for _, pr := range result.Values {
				fmt.Printf("#%-4d %-10s %-50s %s → %s\n",
					pr.ID,
					pr.State,
					truncate(pr.Title, 50),
					pr.Source.Branch.Name,
					pr.Destination.Branch.Name,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&state, "state", "OPEN", "Filter by state: OPEN, MERGED, DECLINED, SUPERSEDED")

	return cmd
}

func newPRViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view <id>",
		Short: "View pull request details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := resolve.PRService()
			if err != nil {
				return err
			}

			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("validation: PR id must be a number")
			}

			pr, err := svc.Get(context.Background(), id)
			if err != nil {
				return err
			}

			fmt.Printf("PR #%d: %s\n", pr.ID, pr.Title)
			fmt.Printf("State:  %s\n", pr.State)
			fmt.Printf("Author: %s\n", pr.Author.DisplayName)
			fmt.Printf("Branch: %s → %s\n", pr.Source.Branch.Name, pr.Destination.Branch.Name)
			fmt.Printf("Created: %s\n", pr.CreatedOn.Format("2006-01-02 15:04"))
			if pr.Description != "" {
				fmt.Printf("\n%s\n", pr.Description)
			}
			if pr.Links.HTML != nil {
				fmt.Printf("\nURL: %s\n", pr.Links.HTML.Href)
			}

			if len(pr.Participants) > 0 {
				fmt.Println("\nParticipants:")
				for _, p := range pr.Participants {
					status := "reviewed"
					if p.Approved {
						status = "approved"
					}
					fmt.Printf("  %s (%s) — %s\n", p.User.DisplayName, p.Role, status)
				}
			}

			return nil
		},
	}
}

func newPRMergeCmd() *cobra.Command {
	var (
		strategy    string
		closeBranch bool
		message     string
	)

	cmd := &cobra.Command{
		Use:   "merge <id>",
		Short: "Merge a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := resolve.PRService()
			if err != nil {
				return err
			}

			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("validation: PR id must be a number")
			}

			opts := bitbucket.MergeOpts{
				MergeStrategy: strategy,
				Message:       message,
			}
			if cmd.Flags().Changed("close-branch") {
				opts.CloseSourceBranch = &closeBranch
			}

			pr, err := svc.Merge(context.Background(), id, opts)
			if err != nil {
				return err
			}

			fmt.Printf("Merged PR #%d: %s\n", pr.ID, pr.Title)

			return nil
		},
	}

	cmd.Flags().StringVar(&strategy, "strategy", "", "Merge strategy: merge_commit, squash, fast_forward")
	cmd.Flags().BoolVar(&closeBranch, "close-branch", false, "Close source branch after merge")
	cmd.Flags().StringVarP(&message, "message", "m", "", "Merge commit message")

	return cmd
}

func newPRApproveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "approve <id>",
		Short: "Approve a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := resolve.PRService()
			if err != nil {
				return err
			}

			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("validation: PR id must be a number")
			}

			if err := svc.Approve(context.Background(), id); err != nil {
				return err
			}

			fmt.Printf("Approved PR #%d\n", id)
			return nil
		},
	}
}

func newPRDeclineCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "decline <id>",
		Short: "Decline a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := resolve.PRService()
			if err != nil {
				return err
			}

			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("validation: PR id must be a number")
			}

			pr, err := svc.Decline(context.Background(), id)
			if err != nil {
				return err
			}

			fmt.Printf("Declined PR #%d: %s\n", pr.ID, pr.Title)
			return nil
		},
	}
}

func newPRCommentCmd() *cobra.Command {
	var body string

	cmd := &cobra.Command{
		Use:   "comment <id>",
		Short: "Add a comment to a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := resolve.PRService()
			if err != nil {
				return err
			}

			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("validation: PR id must be a number")
			}

			if body == "" {
				return fmt.Errorf("validation: --body is required")
			}

			comment, err := svc.AddComment(context.Background(), id, body)
			if err != nil {
				return err
			}

			fmt.Printf("Added comment #%d to PR #%d\n", comment.ID, id)
			return nil
		},
	}

	cmd.Flags().StringVarP(&body, "body", "b", "", "Comment body (required)")

	return cmd
}

func newPRDiffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diff <id>",
		Short: "Show pull request diff",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := resolve.PRService()
			if err != nil {
				return err
			}

			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("validation: PR id must be a number")
			}

			diff, err := svc.Diff(context.Background(), id)
			if err != nil {
				return err
			}

			fmt.Print(diff)
			return nil
		},
	}
}

func currentBranch() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

