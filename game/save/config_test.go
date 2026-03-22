package save

import (
	"path/filepath"
	"testing"
)

func TestSaveConfig_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := GameConfig{
		WindowWidth:      800,
		WindowHeight:     600,
		FastForwardSpeed: 5,
	}

	if err := SaveConfig(path, cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if loaded.WindowWidth != cfg.WindowWidth {
		t.Errorf("WindowWidth = %d, want %d", loaded.WindowWidth, cfg.WindowWidth)
	}
	if loaded.WindowHeight != cfg.WindowHeight {
		t.Errorf("WindowHeight = %d, want %d", loaded.WindowHeight, cfg.WindowHeight)
	}
	if loaded.FastForwardSpeed != cfg.FastForwardSpeed {
		t.Errorf("FastForwardSpeed = %d, want %d", loaded.FastForwardSpeed, cfg.FastForwardSpeed)
	}
}

func TestLoadConfig_NonexistentFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json")
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	defaults := DefaultGameConfig()
	if cfg.WindowWidth != defaults.WindowWidth {
		t.Errorf("WindowWidth = %d, want default %d", cfg.WindowWidth, defaults.WindowWidth)
	}
	if cfg.WindowHeight != defaults.WindowHeight {
		t.Errorf("WindowHeight = %d, want default %d", cfg.WindowHeight, defaults.WindowHeight)
	}
	if cfg.FastForwardSpeed != defaults.FastForwardSpeed {
		t.Errorf("FastForwardSpeed = %d, want default %d", cfg.FastForwardSpeed, defaults.FastForwardSpeed)
	}
}

func TestDefaultGameConfig(t *testing.T) {
	cfg := DefaultGameConfig()
	if cfg.WindowWidth <= 0 {
		t.Errorf("default WindowWidth = %d, want > 0", cfg.WindowWidth)
	}
	if cfg.WindowHeight <= 0 {
		t.Errorf("default WindowHeight = %d, want > 0", cfg.WindowHeight)
	}
	if cfg.FastForwardSpeed <= 0 {
		t.Errorf("default FastForwardSpeed = %d, want > 0", cfg.FastForwardSpeed)
	}
}
