package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/heyashy/bitbucket-cli/internal/bitbucket"
	"github.com/heyashy/bitbucket-cli/internal/cmd/resolve"
	"github.com/heyashy/bitbucket-cli/internal/ui"
)

func newPRCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pr",
		Short: "Manage pull requests",
		Long: `Create, list, view, merge, approve, decline, comment on, and diff pull requests
on Bitbucket Cloud.

Most commands accept an optional [id] argument. When omitted, bb automatically
finds the open pull request for your current git branch.

The workspace and repository are auto-detected from your git remote URL.
Override with: bb config set workspace <name> / bb config set repo_slug <name>

COMMANDS
  bb pr create               Create a new PR from the current branch
  bb pr list                 List PRs (default: open)
  bb pr view [id]            Show PR details, reviewers, and status
  bb pr merge [id]           Merge a PR (supports squash/fast-forward)
  bb pr approve [id]         Approve a PR
  bb pr decline [id]         Decline a PR
  bb pr comment [id] -b msg  Add a comment to a PR
  bb pr diff [id]            Show the full diff of a PR

EXAMPLES
  bb pr create -t "Fix bug" -d main
  bb pr create -t "Feature" -b "Description here" --close-branch
  bb pr list --state MERGED
  bb pr view                    View the PR for the current branch
  bb pr merge --strategy squash
  bb pr approve 42
  bb pr comment -b "LGTM"`,
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

// resolvePRID gets a PR ID from args, or finds the open PR for the current branch.
func resolvePRID(args []string, svc *bitbucket.PRService) (int, error) {
	if len(args) > 0 {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return 0, fmt.Errorf("validation: PR id must be a number, got '%s'", args[0])
		}
		return id, nil
	}

	branch, err := currentBranch()
	if err != nil {
		return 0, fmt.Errorf("validation: provide a PR id or run from a feature branch")
	}

	result, err := svc.List(context.Background(), bitbucket.ListPRsOpts{State: "OPEN"})
	if err != nil {
		return 0, err
	}

	for _, pr := range result.Values {
		if pr.Source.Branch.Name == branch {
			return pr.ID, nil
		}
	}

	return 0, fmt.Errorf("domain: no open PR found for branch '%s' — provide a PR id", branch)
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
		Long: `Create a new pull request on Bitbucket Cloud.

Source branch defaults to your current git branch.
Destination branch defaults to "main".

FLAGS
  -t, --title        PR title (required)
  -b, --body         PR description/body text
  -s, --source       Source branch (default: current git branch)
  -d, --dest         Destination branch (default: main)
  -r, --reviewer     Reviewer UUID (repeatable for multiple reviewers)
      --close-branch Delete source branch after merge

EXAMPLES
  bb pr create -t "Fix login bug"
  bb pr create -t "Add feature" -b "Detailed description" -d develop
  bb pr create -t "Hotfix" --close-branch -r "{uuid-of-reviewer}"`,
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

			fmt.Println()
			fmt.Printf("  %s Created PR %s\n", ui.CheckMark(), ui.PRNumber.Render(fmt.Sprintf("#%d", pr.ID)))
			fmt.Printf("  %s\n", ui.Bold.Render(pr.Title))
			fmt.Printf("  %s %s %s\n", ui.PRBranch.Render(pr.Source.Branch.Name), ui.Arrow(), ui.PRBranch.Render(pr.Destination.Branch.Name))
			if pr.Links.HTML != nil {
				fmt.Printf("  %s\n", ui.Faint.Render(pr.Links.HTML.Href))
			}
			fmt.Println()

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
		Long: `List pull requests for the current repository.

Defaults to showing open PRs. Use --state to filter.

FLAGS
  --state   Filter by state: OPEN, MERGED, DECLINED, SUPERSEDED (default: OPEN)

EXAMPLES
  bb pr list
  bb pr list --state MERGED
  bb pr list --state DECLINED`,
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
				fmt.Println(ui.Faint.Render("  No pull requests found."))
				return nil
			}

			fmt.Println()
			for _, pr := range result.Values {
				number := ui.PRNumber.Render(fmt.Sprintf("#%d", pr.ID))
				badge := ui.StatusBadge(pr.State)
				title := ui.PRTitle.Render(truncate(pr.Title, 50))
				branch := fmt.Sprintf("%s %s %s",
					ui.PRBranch.Render(pr.Source.Branch.Name),
					ui.Arrow(),
					ui.PRBranch.Render(pr.Destination.Branch.Name),
				)
				author := ui.PRAuthor.Render(pr.Author.DisplayName)

				fmt.Printf("  %s %s %s\n", number, badge, title)
				fmt.Printf("         %s  %s\n", branch, author)
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&state, "state", "OPEN", "Filter by state: OPEN, MERGED, DECLINED, SUPERSEDED")

	return cmd
}

func newPRViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view [id]",
		Short: "View pull request details (defaults to current branch PR)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := resolve.PRService()
			if err != nil {
				return err
			}

			id, err := resolvePRID(args, svc)
			if err != nil {
				return err
			}

			pr, err := svc.Get(context.Background(), id)
			if err != nil {
				return err
			}

			fmt.Println()
			fmt.Printf("  %s %s %s\n",
				ui.PRNumber.Render(fmt.Sprintf("#%d", pr.ID)),
				ui.StatusBadge(pr.State),
				ui.Bold.Render(pr.Title),
			)
			fmt.Println(ui.Divider.String())

			fmt.Printf("  %s %s\n", ui.Key.Render("Author:"), ui.PRAuthor.Render(pr.Author.DisplayName))
			fmt.Printf("  %s %s %s %s\n",
				ui.Key.Render("Branch:"),
				ui.Value.Render(pr.Source.Branch.Name),
				ui.Arrow(),
				ui.Value.Render(pr.Destination.Branch.Name),
			)
			fmt.Printf("  %s %s\n", ui.Key.Render("Created:"), ui.Faint.Render(pr.CreatedOn.Format("2006-01-02 15:04")))

			if pr.Description != "" {
				fmt.Println()
				fmt.Printf("  %s\n", pr.Description)
			}

			if len(pr.Participants) > 0 {
				fmt.Println()
				fmt.Println(ui.Divider.String())
				fmt.Println(ui.Bold.Render("  Reviewers"))
				for _, p := range pr.Participants {
					var status string
					if p.Approved {
						status = ui.Success.Render("approved")
					} else {
						status = ui.Faint.Render(p.State)
					}
					fmt.Printf("  %s %s (%s)\n",
						ui.PRAuthor.Render(p.User.DisplayName),
						ui.Faint.Render(p.Role),
						status,
					)
				}
			}

			if pr.Links.HTML != nil {
				fmt.Println()
				fmt.Printf("  %s\n", ui.Faint.Render(pr.Links.HTML.Href))
			}
			fmt.Println()

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
		Use:   "merge [id]",
		Short: "Merge a pull request (defaults to current branch PR)",
		Long: `Merge a pull request. Defaults to the open PR for the current branch.

FLAGS
  --strategy      Merge strategy: merge_commit, squash, fast_forward
  --close-branch  Delete source branch after merge
  -m, --message   Custom merge commit message

EXAMPLES
  bb pr merge                            Merge current branch PR
  bb pr merge 42
  bb pr merge --strategy squash
  bb pr merge --close-branch -m "Ship it"`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := resolve.PRService()
			if err != nil {
				return err
			}

			id, err := resolvePRID(args, svc)
			if err != nil {
				return err
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

			fmt.Printf("\n  %s Merged PR %s %s\n\n",
				ui.CheckMark(),
				ui.PRNumber.Render(fmt.Sprintf("#%d", pr.ID)),
				ui.Bold.Render(pr.Title),
			)

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
		Use:   "approve [id]",
		Short: "Approve a pull request (defaults to current branch PR)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := resolve.PRService()
			if err != nil {
				return err
			}

			id, err := resolvePRID(args, svc)
			if err != nil {
				return err
			}

			if err := svc.Approve(context.Background(), id); err != nil {
				return err
			}

			fmt.Printf("\n  %s Approved PR %s\n\n", ui.CheckMark(), ui.PRNumber.Render(fmt.Sprintf("#%d", id)))
			return nil
		},
	}
}

func newPRDeclineCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "decline [id]",
		Short: "Decline a pull request (defaults to current branch PR)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := resolve.PRService()
			if err != nil {
				return err
			}

			id, err := resolvePRID(args, svc)
			if err != nil {
				return err
			}

			pr, err := svc.Decline(context.Background(), id)
			if err != nil {
				return err
			}

			fmt.Printf("\n  %s Declined PR %s %s\n\n",
				ui.CrossMark(),
				ui.PRNumber.Render(fmt.Sprintf("#%d", pr.ID)),
				ui.Bold.Render(pr.Title),
			)
			return nil
		},
	}
}

func newPRCommentCmd() *cobra.Command {
	var body string

	cmd := &cobra.Command{
		Use:   "comment [id]",
		Short: "Add a comment to a pull request (defaults to current branch PR)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := resolve.PRService()
			if err != nil {
				return err
			}

			id, err := resolvePRID(args, svc)
			if err != nil {
				return err
			}

			if body == "" {
				return fmt.Errorf("validation: --body is required")
			}

			comment, err := svc.AddComment(context.Background(), id, body)
			if err != nil {
				return err
			}

			fmt.Printf("\n  %s Added comment #%d to PR %s\n\n",
				ui.CheckMark(),
				comment.ID,
				ui.PRNumber.Render(fmt.Sprintf("#%d", id)),
			)
			return nil
		},
	}

	cmd.Flags().StringVarP(&body, "body", "b", "", "Comment body (required)")

	return cmd
}

func newPRDiffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diff [id]",
		Short: "Show pull request diff (defaults to current branch PR)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := resolve.PRService()
			if err != nil {
				return err
			}

			id, err := resolvePRID(args, svc)
			if err != nil {
				return err
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
