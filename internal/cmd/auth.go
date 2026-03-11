package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/heyashy/bitbucket-cli/internal/auth"
	"github.com/heyashy/bitbucket-cli/internal/config"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication with Bitbucket",
	}

	cmd.AddCommand(newAuthLoginCmd())
	cmd.AddCommand(newAuthLogoutCmd())
	cmd.AddCommand(newAuthStatusCmd())

	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	var appPassword bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to Bitbucket",
		Long:  "Authenticate with Bitbucket using OAuth (browser) or an app password.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if appPassword {
				return loginWithAppPassword()
			}
			return loginWithOAuth()
		},
	}

	cmd.Flags().BoolVar(&appPassword, "app-password", false, "Use app password instead of OAuth")

	return cmd
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

func loginWithAppPassword() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Bitbucket username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("App password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	if username == "" || password == "" {
		return fmt.Errorf("validation: username and app password are required")
	}

	token := &auth.StoredToken{
		AccessToken: username + ":" + password,
		TokenType:   "apppassword",
	}

	if err := auth.SaveToken(token); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	cfg.Set("auth.method", "apppassword")
	if err := cfg.WriteGlobal(); err != nil {
		return err
	}

	fmt.Println("Logged in successfully with app password.")
	return nil
}

func newAuthLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Log out of Bitbucket",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := auth.DeleteToken(); err != nil {
				return err
			}
			fmt.Println("Logged out successfully.")
			return nil
		},
	}
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := auth.LoadToken()
			if err != nil {
				fmt.Println("Not logged in.")
				return nil
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			method := cfg.AuthMethod()
			fmt.Printf("Auth method: %s\n", method)

			if token.TokenType == "apppassword" {
				parts := strings.SplitN(token.AccessToken, ":", 2)
				if len(parts) == 2 {
					fmt.Printf("Username: %s\n", parts[0])
				}
			} else {
				if token.IsExpired() {
					fmt.Println("Status: token expired (will auto-refresh)")
				} else {
					fmt.Println("Status: authenticated")
				}
				if token.Scopes != "" {
					fmt.Printf("Scopes: %s\n", token.Scopes)
				}
			}

			return nil
		},
	}
}
