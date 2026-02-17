package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error for missing config file, got: %v", err)
	}
	if cfg.HubURL != "http://localhost:3000" {
		t.Errorf("expected default HubURL='http://localhost:3000', got '%s'", cfg.HubURL)
	}
}

func TestLoad_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfgDir := filepath.Join(tmpDir, ".config", "agenthq")
	if err := os.MkdirAll(cfgDir, 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	cfgData := Config{
		HubURL:   "https://example.com",
		APIKey:   "ahq_testkey123",
		JWTToken: "jwt-token-abc",
		OrgID:    "org-123",
		AgentID:  "agent-456",
	}
	data, _ := json.Marshal(cfgData)
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), data, 0600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.HubURL != "https://example.com" {
		t.Errorf("expected HubURL='https://example.com', got '%s'", cfg.HubURL)
	}
	if cfg.APIKey != "ahq_testkey123" {
		t.Errorf("expected APIKey='ahq_testkey123', got '%s'", cfg.APIKey)
	}
	if cfg.JWTToken != "jwt-token-abc" {
		t.Errorf("expected JWTToken='jwt-token-abc', got '%s'", cfg.JWTToken)
	}
	if cfg.OrgID != "org-123" {
		t.Errorf("expected OrgID='org-123', got '%s'", cfg.OrgID)
	}
	if cfg.AgentID != "agent-456" {
		t.Errorf("expected AgentID='agent-456', got '%s'", cfg.AgentID)
	}
}

func TestLoad_EmptyHubURL(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfgDir := filepath.Join(tmpDir, ".config", "agenthq")
	if err := os.MkdirAll(cfgDir, 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	cfgData := Config{
		HubURL: "",
		APIKey: "ahq_key",
	}
	data, _ := json.Marshal(cfgData)
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), data, 0600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.HubURL != "http://localhost:3000" {
		t.Errorf("expected default HubURL when empty, got '%s'", cfg.HubURL)
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg := &Config{
		HubURL: "https://example.com",
		APIKey: "ahq_key",
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("unexpected error saving config: %v", err)
	}

	cfgDir := filepath.Join(tmpDir, ".config", "agenthq")
	info, err := os.Stat(cfgDir)
	if err != nil {
		t.Fatalf("config directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("expected config path to be a directory")
	}
}

func TestSave_WritesJSON(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg := &Config{
		HubURL:   "https://hub.example.com",
		APIKey:   "ahq_mykey",
		JWTToken: "my-jwt",
		OrgID:    "org-abc",
		AgentID:  "agent-xyz",
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("unexpected error saving config: %v", err)
	}

	cfgPath := filepath.Join(tmpDir, ".config", "agenthq", "config.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("failed to read saved config file: %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("saved config is not valid JSON: %v", err)
	}

	if loaded.HubURL != "https://hub.example.com" {
		t.Errorf("expected HubURL='https://hub.example.com', got '%s'", loaded.HubURL)
	}
	if loaded.APIKey != "ahq_mykey" {
		t.Errorf("expected APIKey='ahq_mykey', got '%s'", loaded.APIKey)
	}
	if loaded.OrgID != "org-abc" {
		t.Errorf("expected OrgID='org-abc', got '%s'", loaded.OrgID)
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	original := &Config{
		HubURL:   "https://roundtrip.example.com",
		APIKey:   "ahq_round",
		JWTToken: "jwt-round",
		OrgID:    "org-round",
		AgentID:  "agent-round",
	}

	if err := Save(original); err != nil {
		t.Fatalf("unexpected error saving config: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if loaded.HubURL != original.HubURL {
		t.Errorf("HubURL mismatch: got '%s', want '%s'", loaded.HubURL, original.HubURL)
	}
	if loaded.APIKey != original.APIKey {
		t.Errorf("APIKey mismatch: got '%s', want '%s'", loaded.APIKey, original.APIKey)
	}
	if loaded.JWTToken != original.JWTToken {
		t.Errorf("JWTToken mismatch: got '%s', want '%s'", loaded.JWTToken, original.JWTToken)
	}
	if loaded.OrgID != original.OrgID {
		t.Errorf("OrgID mismatch: got '%s', want '%s'", loaded.OrgID, original.OrgID)
	}
	if loaded.AgentID != original.AgentID {
		t.Errorf("AgentID mismatch: got '%s', want '%s'", loaded.AgentID, original.AgentID)
	}
}

func TestGetAuthToken_APIKey(t *testing.T) {
	cfg := &Config{
		APIKey: "ahq_myapikey",
	}

	token := cfg.GetAuthToken()
	if token != "ahq_myapikey" {
		t.Errorf("expected 'ahq_myapikey', got '%s'", token)
	}
}

func TestGetAuthToken_JWT(t *testing.T) {
	cfg := &Config{
		JWTToken: "jwt-token-value",
	}

	token := cfg.GetAuthToken()
	if token != "jwt-token-value" {
		t.Errorf("expected 'jwt-token-value', got '%s'", token)
	}
}

func TestGetAuthToken_Empty(t *testing.T) {
	cfg := &Config{}

	token := cfg.GetAuthToken()
	if token != "" {
		t.Errorf("expected empty string, got '%s'", token)
	}
}

func TestGetAuthToken_APIKeyPriority(t *testing.T) {
	cfg := &Config{
		APIKey:   "ahq_priority",
		JWTToken: "jwt-should-not-be-used",
	}

	token := cfg.GetAuthToken()
	if token != "ahq_priority" {
		t.Errorf("expected APIKey to take priority, got '%s'", token)
	}
}
