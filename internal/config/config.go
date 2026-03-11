package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	v *viper.Viper
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.SetEnvPrefix("BB")
	v.AutomaticEnv()

	v.SetDefault("auth.method", "oauth")
	v.SetDefault("pr.merge_strategy", "merge_commit")

	home, err := os.UserHomeDir()
	if err == nil {
		globalDir := filepath.Join(home, ".config", "bb")
		v.AddConfigPath(globalDir)
	}

	v.AddConfigPath(".bb")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	repoViper := viper.New()
	repoViper.SetConfigName("config")
	repoViper.SetConfigType("toml")
	repoViper.AddConfigPath(".bb")

	if err := repoViper.ReadInConfig(); err == nil {
		for _, key := range repoViper.AllKeys() {
			v.Set(key, repoViper.Get(key))
		}
	}

	return &Config{v: v}, nil
}

func (c *Config) Workspace() string {
	return c.v.GetString("workspace")
}

func (c *Config) RepoSlug() string {
	return c.v.GetString("repo_slug")
}

func (c *Config) AuthMethod() string {
	return c.v.GetString("auth.method")
}

func (c *Config) DefaultMergeStrategy() string {
	return c.v.GetString("pr.merge_strategy")
}

func (c *Config) DefaultReviewers() []string {
	return c.v.GetStringSlice("pr.default_reviewers")
}

func (c *Config) Get(key string) interface{} {
	return c.v.Get(key)
}

func (c *Config) GetString(key string) string {
	return c.v.GetString(key)
}

func (c *Config) Set(key string, value interface{}) {
	c.v.Set(key, value)
}

func (c *Config) WriteGlobal() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dir := filepath.Join(home, ".config", "bb")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	return c.v.WriteConfigAs(filepath.Join(dir, "config.toml"))
}

func (c *Config) OAuthClientID() string {
	return c.v.GetString("auth.oauth.client_id")
}

func (c *Config) OAuthClientSecret() string {
	return c.v.GetString("auth.oauth.client_secret")
}

func GlobalConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "bb"), nil
}
