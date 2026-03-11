package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type StoredToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scopes       string    `json:"scopes,omitempty"`
}

func (t *StoredToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt.Add(-30 * time.Second))
}

func TokenFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("infra: cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "bb", "tokens.json"), nil
}

func LoadToken() (*StoredToken, error) {
	path, err := TokenFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("infra: cannot read token file: %w", err)
	}

	var token StoredToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("infra: cannot parse token file: %w", err)
	}

	return &token, nil
}

func SaveToken(token *StoredToken) error {
	path, err := TokenFilePath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("infra: cannot create config directory: %w", err)
	}

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("infra: cannot marshal token: %w", err)
	}

	return os.WriteFile(path, data, 0o600)
}

func DeleteToken() error {
	path, err := TokenFilePath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("infra: cannot delete token file: %w", err)
	}
	return nil
}
