package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

var (
	config     Config
	configOnce sync.Once
	configErr  error
	configPath string
)

// loadConfig initializes the config from file
func loadConfig() {
	exePath, err := os.Executable()
	configPath := os.Getenv("CONFIG_PATH")
	if err != nil {
		configErr = err
		return
	}
	if configPath == "" {
		exeDir := filepath.Dir(exePath)
		configPath = filepath.Join(exeDir, "config.json")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		configErr = err
		return
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		configErr = err
		return
	}
}

// GetConfig returns the singleton instance of the config.
// It's safe for concurrent use and loads the config only once per process lifetime.
func GetConfig() (*Config, error) {
	configOnce.Do(loadConfig)
	return &config, configErr
}

// ConfigPath returns the path to the loaded config file (useful for logging or reload later if needed)
func ConfigPath() string {
	// Make sure config is at least attempted to be loaded
	configOnce.Do(loadConfig)
	return configPath
}
