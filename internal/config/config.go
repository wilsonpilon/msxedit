package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Theme      string `json:"theme" omitzero:"true"`
	TabSize    int    `json:"tab_size" omitzero:"true"`
	ShowLineNo bool   `json:"show_line_numbers" omitzero:"true"`
	Highlight  bool   `json:"highlight" omitzero:"true"`
}

const (
	ConfigFileName = "msxedit.json"
	AppName        = "msxedit"
)

func GetDefaultConfig() Config {
	return Config{
		Theme:      "default",
		TabSize:    4,
		ShowLineNo: true,
		Highlight:  true,
	}
}

func GetConfigPath(local bool) (string, error) {
	if local {
		return ConfigFileName, nil
	}

	appData, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(appData, AppName)
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", err
		}
	}

	return filepath.Join(dir, ConfigFileName), nil
}

func Load(path string) (Config, error) {
	cfg := GetDefaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("erro ao decodificar config: %w", err)
	}

	return cfg, nil
}

func Save(path string, cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
