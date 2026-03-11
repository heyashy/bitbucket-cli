package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/browser"
)

const (
	bitbucketAuthorizeURL = "https://bitbucket.org/site/oauth2/authorize"
	bitbucketTokenURL     = "https://bitbucket.org/site/oauth2/access_token"
	callbackPath          = "/callback"
)

type OAuthProvider struct {
	ClientID     string
	ClientSecret string
	token        *StoredToken
}

func NewOAuthProvider(clientID, clientSecret string) *OAuthProvider {
	return &OAuthProvider{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
}

func (p *OAuthProvider) AuthStyle() AuthStyle {
	return AuthStyleBearer
}

func (p *OAuthProvider) Token() (string, error) {
	if p.token == nil {
		stored, err := LoadToken()
		if err != nil {
			return "", fmt.Errorf("domain: not logged in — run 'bb auth login': %w", err)
		}
		p.token = stored
	}

	if p.token.IsExpired() {
		if err := p.refresh(); err != nil {
			return "", fmt.Errorf("infra: token refresh failed: %w", err)
		}
	}

	return p.token.AccessToken, nil
}

func (p *OAuthProvider) ApplyAuth(req *http.Request) error {
	token, err := p.Token()
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (p *OAuthProvider) Login() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("infra: cannot start callback server: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d%s", port, callbackPath)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errMsg := r.URL.Query().Get("error_description")
			if errMsg == "" {
				errMsg = "no authorization code received"
			}
			errCh <- fmt.Errorf("domain: %s", errMsg)
			fmt.Fprint(w, "<html><body><h2>Authentication failed.</h2><p>You can close this tab.</p></body></html>")
			return
		}
		codeCh <- code
		fmt.Fprint(w, "<html><body><h2>Authentication successful!</h2><p>You can close this tab.</p></body></html>")
	})

	server := &http.Server{Handler: mux}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	authURL := fmt.Sprintf("%s?client_id=%s&response_type=code&redirect_uri=%s",
		bitbucketAuthorizeURL,
		url.QueryEscape(p.ClientID),
		url.QueryEscape(redirectURI),
	)

	fmt.Printf("Opening browser for authentication...\n")
	fmt.Printf("If the browser doesn't open, visit:\n%s\n\n", authURL)

	if err := browser.OpenURL(authURL); err != nil {
		fmt.Printf("Could not open browser automatically: %v\n", err)
	}

	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		server.Shutdown(context.Background())
		return err
	case <-time.After(5 * time.Minute):
		server.Shutdown(context.Background())
		return fmt.Errorf("domain: authentication timed out after 5 minutes")
	}

	server.Shutdown(context.Background())

	token, err := p.exchangeCode(code, redirectURI)
	if err != nil {
		return err
	}

	p.token = token
	return SaveToken(token)
}

func (p *OAuthProvider) exchangeCode(code, redirectURI string) (*StoredToken, error) {
	data := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {redirectURI},
	}

	req, err := http.NewRequest("POST", bitbucketTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("infra: cannot create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(p.ClientID, p.ClientSecret)

	return doTokenRequest(req)
}

func (p *OAuthProvider) refresh() error {
	if p.token == nil || p.token.RefreshToken == "" {
		return fmt.Errorf("domain: no refresh token available — run 'bb auth login'")
	}

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {p.token.RefreshToken},
	}

	req, err := http.NewRequest("POST", bitbucketTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("infra: cannot create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(p.ClientID, p.ClientSecret)

	token, err := doTokenRequest(req)
	if err != nil {
		return err
	}

	p.token = token
	return SaveToken(token)
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scopes       string `json:"scopes"`
}

func doTokenRequest(req *http.Request) (*StoredToken, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("infra: token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("infra: cannot read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("domain: token exchange failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, fmt.Errorf("infra: cannot parse token response: %w", err)
	}

	return &StoredToken{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		TokenType:    tr.TokenType,
		ExpiresAt:    time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second),
		Scopes:       tr.Scopes,
	}, nil
}
