package auth

import (
	"encoding/base64"
	"fmt"
	"net/http"
)

type AppPasswordProvider struct {
	Username    string
	AppPassword string
}

func NewAppPasswordProvider(username, appPassword string) *AppPasswordProvider {
	return &AppPasswordProvider{
		Username:    username,
		AppPassword: appPassword,
	}
}

func (p *AppPasswordProvider) AuthStyle() AuthStyle {
	return AuthStyleBasic
}

func (p *AppPasswordProvider) Token() (string, error) {
	if p.Username == "" || p.AppPassword == "" {
		return "", fmt.Errorf("validation: username and app password are required")
	}
	return base64.StdEncoding.EncodeToString(
		[]byte(p.Username + ":" + p.AppPassword),
	), nil
}

func (p *AppPasswordProvider) ApplyAuth(req *http.Request) error {
	req.SetBasicAuth(p.Username, p.AppPassword)
	return nil
}
