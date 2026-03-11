package resolve

import (
	"fmt"

	"github.com/heyashy/bb/internal/auth"
	"github.com/heyashy/bb/internal/bitbucket"
	"github.com/heyashy/bb/internal/config"
	"github.com/heyashy/bb/internal/git"
)

func AuthProvider() (auth.Provider, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("cannot load config: %w", err)
	}

	method := cfg.AuthMethod()

	switch method {
	case "apppassword":
		token, err := auth.LoadToken()
		if err != nil {
			return nil, fmt.Errorf("domain: not logged in — run 'bb auth login': %w", err)
		}
		parts := splitOnce(token.AccessToken, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("domain: invalid stored credentials — run 'bb auth login --app-password'")
		}
		return auth.NewAppPasswordProvider(parts[0], parts[1]), nil

	case "oauth", "":
		clientID := cfg.OAuthClientID()
		clientSecret := cfg.OAuthClientSecret()
		if clientID == "" || clientSecret == "" {
			return nil, fmt.Errorf("validation: OAuth not configured — run 'bb auth login --app-password' or configure OAuth")
		}
		return auth.NewOAuthProvider(clientID, clientSecret), nil

	default:
		return nil, fmt.Errorf("validation: unknown auth method: %s", method)
	}
}

func PRService() (*bitbucket.PRService, error) {
	authProvider, err := AuthProvider()
	if err != nil {
		return nil, err
	}

	client := bitbucket.NewClient(authProvider)

	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	workspace := cfg.Workspace()
	repoSlug := cfg.RepoSlug()

	if workspace == "" || repoSlug == "" {
		info, err := git.DetectRepo()
		if err != nil {
			return nil, fmt.Errorf("cannot detect repository — set workspace and repo_slug in config or ensure a Bitbucket remote exists: %w", err)
		}
		workspace = info.Workspace
		repoSlug = info.RepoSlug
	}

	return bitbucket.NewPRService(client, workspace, repoSlug), nil
}

func splitOnce(s, sep string) []string {
	for i := 0; i < len(s); i++ {
		if s[i:i+len(sep)] == sep {
			return []string{s[:i], s[i+len(sep):]}
		}
	}
	return []string{s}
}
