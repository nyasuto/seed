package save

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// DefaultConfigPath returns the default config file path (~/.chaosforge/config.json).
func DefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("user home dir: %w", err)
	}
	return filepath.Join(home, ".chaosforge", "config.json"), nil
}

// GameConfig holds user-configurable game settings.
type GameConfig struct {
	WindowWidth      int `json:"window_width"`
	WindowHeight     int `json:"window_height"`
	FastForwardSpeed int `json:"fast_forward_speed"`
}

// DefaultGameConfig returns the default game configuration.
func DefaultGameConfig() GameConfig {
	return GameConfig{
		WindowWidth:      1088,
		WindowHeight:     728,
		FastForwardSpeed: 10,
	}
}

// SaveConfig writes the game configuration to the given path.
// Parent directories are created if needed.
func SaveConfig(path string, cfg GameConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}

// LoadConfig reads the game configuration from the given path.
// Returns the default configuration if the file does not exist.
func LoadConfig(path string) (GameConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultGameConfig(), nil
		}
		return GameConfig{}, fmt.Errorf("read config file: %w", err)
	}
	var cfg GameConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return GameConfig{}, fmt.Errorf("unmarshal config: %w", err)
	}
	return cfg, nil
}
