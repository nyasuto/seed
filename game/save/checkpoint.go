package save

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nyasuto/seed/core/simulation"
)

// DefaultSaveDir returns the default save directory path (~/.chaosforge/saves/).
func DefaultSaveDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("user home dir: %w", err)
	}
	return filepath.Join(home, ".chaosforge", "saves"), nil
}

// SaveFile is the on-disk representation of a save file.
// It bundles the checkpoint with the scenario JSON needed to restore.
type SaveFile struct {
	ScenarioJSON []byte                `json:"scenario_json"`
	Checkpoint   *simulation.Checkpoint `json:"checkpoint"`
	SavedAt      time.Time             `json:"saved_at"`
	ScenarioID   string                `json:"scenario_id,omitempty"`
}

// GenerateFilename returns a save filename with the current timestamp.
func GenerateFilename() string {
	return fmt.Sprintf("save_%s.json", time.Now().Format("20060102_150405"))
}

// SaveCheckpoint saves the engine state and scenario to the given path.
// Parent directories are created if needed.
func SaveCheckpoint(path string, engine *simulation.SimulationEngine, scenarioJSON []byte, scenarioID string) error {
	cp, err := simulation.CreateCheckpoint(engine)
	if err != nil {
		return fmt.Errorf("create checkpoint: %w", err)
	}

	sf := SaveFile{
		ScenarioJSON: scenarioJSON,
		Checkpoint:   cp,
		SavedAt:      time.Now(),
		ScenarioID:   scenarioID,
	}

	data, err := json.Marshal(sf)
	if err != nil {
		return fmt.Errorf("marshal save file: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create save directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write save file: %w", err)
	}
	return nil
}

// LoadSaveFile reads a save file from the given path.
func LoadSaveFile(path string) (*SaveFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read save file: %w", err)
	}
	var sf SaveFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return nil, fmt.Errorf("unmarshal save file: %w", err)
	}
	return &sf, nil
}

// SaveEntry holds metadata about a save file for listing.
type SaveEntry struct {
	Path       string
	Filename   string
	SavedAt    time.Time
	ScenarioID string
}

// ListSaves returns all save files in the given directory, sorted newest first.
// Returns an empty slice (not error) if the directory does not exist.
func ListSaves(dir string) ([]SaveEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read save directory: %w", err)
	}

	var saves []SaveEntry
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		sf, err := LoadSaveFile(path)
		if err != nil {
			continue // skip invalid files
		}
		saves = append(saves, SaveEntry{
			Path:       path,
			Filename:   e.Name(),
			SavedAt:    sf.SavedAt,
			ScenarioID: sf.ScenarioID,
		})
	}

	sort.Slice(saves, func(i, j int) bool {
		return saves[i].SavedAt.After(saves[j].SavedAt)
	})

	return saves, nil
}
