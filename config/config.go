package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

type Config struct {
	DefaultFolder string `json:"default_folder"`
}

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "markdown-editor")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("could not create config directory: %w", err)
	}

	return filepath.Join(configDir, "config.json"), nil
}

func LoadConfig(window fyne.Window) (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return createDefaultConfig(window)
	}

	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("invalid config format: %w", err)
	}

	return &config, nil
}

func createDefaultConfig(window fyne.Window) (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	result := make(chan struct {
		uri fyne.ListableURI
		err error
	}, 1)

	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		result <- struct {
			uri fyne.ListableURI
			err error
		}{uri, err}
	}, window)

	selection := <-result
	if selection.err != nil {
		return nil, fmt.Errorf("folder selection failed: %w", selection.err)
	}
	if selection.uri == nil {
		return nil, fmt.Errorf("no folder selected")
	}

	config := Config{
		DefaultFolder: selection.uri.Path(),
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("could not marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return nil, fmt.Errorf("could not write config: %w", err)
	}

	return &config, nil
}
