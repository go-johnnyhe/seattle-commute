package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	HomeAddress   string `json:"home_address"`
	WorkAddress   string `json:"work_address,omitempty"`
	GoogleAPIKey  string `json:"google_api_key"`
}

func GetConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".seattle-commute", "config.json")
}

func LoadConfig() (*Config, error) {
	configPath := GetConfigPath()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) Save() error {
	configPath := GetConfigPath()
	configDir := filepath.Dir(configPath)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

func (c *Config) IsValid() bool {
	return c.HomeAddress != "" && c.GoogleAPIKey != ""
}