package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	HubURL   string `json:"hub_url"`
	APIKey   string `json:"api_key"`
	JWTToken string `json:"jwt_token"`
	OrgID    string `json:"org_id"`
	AgentID  string `json:"agent_id"`
}

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "agenthq")
}

func configPath() string {
	return filepath.Join(configDir(), "config.json")
}

func Load() (*Config, error) {
	data, err := os.ReadFile(configPath())
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{HubURL: "http://localhost:3000"}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.HubURL == "" {
		cfg.HubURL = "http://localhost:3000"
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0600)
}

func (c *Config) GetAuthToken() string {
	if c.APIKey != "" {
		return c.APIKey
	}
	return c.JWTToken
}
