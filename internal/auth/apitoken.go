package auth

import (
	"fmt"
	"net/http"
)

type APITokenProvider struct {
	Email    string
	APIToken string
}

func NewAPITokenProvider(email, apiToken string) *APITokenProvider {
	return &APITokenProvider{
		Email:    email,
		APIToken: apiToken,
	}
}

func (p *APITokenProvider) AuthStyle() AuthStyle {
	return AuthStyleBasic
}

func (p *APITokenProvider) Token() (string, error) {
	if p.Email == "" || p.APIToken == "" {
		return "", fmt.Errorf("validation: email and API token are required")
	}
	return p.APIToken, nil
}

func (p *APITokenProvider) ApplyAuth(req *http.Request) error {
	req.SetBasicAuth(p.Email, p.APIToken)
	return nil
}
