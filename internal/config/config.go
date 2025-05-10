package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"markdown-editor/internal/app"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

var (
	ErrConfigHomeDir         = errors.New("could not get user home directory")
	ErrConfigDirCreation     = errors.New("could not create config directory")
	ErrConfigReadFailed      = errors.New("failed to read config file")
	ErrConfigParseFailed     = errors.New("invalid config format")
	ErrConfigMarshalFailed   = errors.New("could not marshal config to JSON")
	ErrConfigWriteFailed     = errors.New("could not write config file")
	ErrFolderSelectionFailed = errors.New("folder selection failed or was cancelled")
	ErrNoFolderSelected      = errors.New("no folder selected by the user")
)

type Config struct {
	DefaultFolder string `json:"default_folder"`
}

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrConfigHomeDir, err)
	}

	configDir := filepath.Join(home, ".config", "markdown-editor")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("%w: %v", ErrConfigDirCreation, err)
	}

	return filepath.Join(configDir, "config.json"), nil
}

func LoadConfig(window fyne.Window) (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		app.ShowErrorNotification("Configuration Error", "Could not determine configuration path.", err)
		return nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return createDefaultConfig(window)
	}

	file, err := os.ReadFile(configPath)
	if err != nil {
		wrappedErr := fmt.Errorf("%w: %v", ErrConfigReadFailed, err)
		app.ShowErrorNotification("Configuration Error", "Failed to read the configuration file.", wrappedErr)
		return nil, wrappedErr
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		wrappedErr := fmt.Errorf("%w: %v", ErrConfigParseFailed, err)
		app.ShowErrorNotification("Configuration Error", "Configuration file format is invalid.", wrappedErr)
		return nil, wrappedErr
	}

	return &config, nil
}

func createDefaultConfig(window fyne.Window) (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		app.ShowErrorNotification("Configuration Error", "Could not determine configuration path during setup.", err)
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
		wrappedErr := fmt.Errorf("%w: %v", ErrFolderSelectionFailed, selection.err)
		app.ShowErrorNotification("Configuration Setup", "Folder selection process failed.", wrappedErr)
		return nil, wrappedErr
	}
	if selection.uri == nil {
		app.ShowErrorNotification("Configuration Setup", ErrNoFolderSelected.Error(), ErrNoFolderSelected)
		return nil, ErrNoFolderSelected
	}

	config := Config{
		DefaultFolder: selection.uri.Path(),
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		wrappedErr := fmt.Errorf("%w: %v", ErrConfigMarshalFailed, err)
		app.ShowErrorNotification("Configuration Error", "Failed to prepare configuration for saving.", wrappedErr)
		return nil, wrappedErr
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		wrappedErr := fmt.Errorf("%w: %v", ErrConfigWriteFailed, err)
		app.ShowErrorNotification("Configuration Error", "Failed to save the configuration file.", wrappedErr)
		return nil, wrappedErr
	}
	app.ShowSuccessNotification("Configuration Saved", fmt.Sprintf("Workspace configured to: %s", config.DefaultFolder))
	return &config, nil
}
