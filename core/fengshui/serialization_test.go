package fengshui

import (
	"math"
	"testing"

	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

func TestChiFlowEngineSerialization_SaveRestore(t *testing.T) {
	cave, source := buildTwoRoomCave(t, "wood_room", "fire_room")
	reg := testRegistry()
	params := DefaultFlowParams()

	vein, err := BuildDragonVein(cave, source, types.Wood, 10.0)
	if err != nil {
		t.Fatalf("BuildDragonVein: %v", err)
	}

	engine := NewChiFlowEngine(cave, []*DragonVein{vein}, reg, params)

	// Run a few ticks to get non-zero chi values.
	for range 5 {
		engine.Tick()
	}

	// Serialize.
	data, err := engine.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	// Deserialize.
	restored, err := UnmarshalChiFlowEngine(data, cave, reg, params)
	if err != nil {
		t.Fatalf("UnmarshalChiFlowEngine: %v", err)
	}

	// Verify veins.
	if len(restored.Veins) != len(engine.Veins) {
		t.Fatalf("veins count: got %d, want %d", len(restored.Veins), len(engine.Veins))
	}
	for i, orig := range engine.Veins {
		rest := restored.Veins[i]
		if rest.ID != orig.ID {
			t.Errorf("vein %d ID: got %d, want %d", i, rest.ID, orig.ID)
		}
		if rest.SourcePos != orig.SourcePos {
			t.Errorf("vein %d SourcePos: got %v, want %v", i, rest.SourcePos, orig.SourcePos)
		}
		if rest.Element != orig.Element {
			t.Errorf("vein %d Element: got %v, want %v", i, rest.Element, orig.Element)
		}
		if rest.FlowRate != orig.FlowRate {
			t.Errorf("vein %d FlowRate: got %v, want %v", i, rest.FlowRate, orig.FlowRate)
		}
		if len(rest.Path) != len(orig.Path) {
			t.Errorf("vein %d path length: got %d, want %d", i, len(rest.Path), len(orig.Path))
			continue
		}
		for j, op := range orig.Path {
			if rest.Path[j] != op {
				t.Errorf("vein %d path[%d]: got %v, want %v", i, j, rest.Path[j], op)
			}
		}
	}

	// Verify room chi.
	if len(restored.RoomChi) != len(engine.RoomChi) {
		t.Fatalf("room chi count: got %d, want %d", len(restored.RoomChi), len(engine.RoomChi))
	}
	for rid, orig := range engine.RoomChi {
		rest, ok := restored.RoomChi[rid]
		if !ok {
			t.Errorf("room chi %d: missing in restored", rid)
			continue
		}
		if rest.RoomID != orig.RoomID {
			t.Errorf("room chi %d RoomID: got %d, want %d", rid, rest.RoomID, orig.RoomID)
		}
		if math.Abs(rest.Current-orig.Current) > 1e-9 {
			t.Errorf("room chi %d Current: got %v, want %v", rid, rest.Current, orig.Current)
		}
		if rest.Capacity != orig.Capacity {
			t.Errorf("room chi %d Capacity: got %v, want %v", rid, rest.Capacity, orig.Capacity)
		}
		if rest.Element != orig.Element {
			t.Errorf("room chi %d Element: got %v, want %v", rid, rest.Element, orig.Element)
		}
	}

	// Verify that the restored engine has its cave and registry set.
	// Run a tick to confirm it works without panic.
	restored.Tick()
}

func TestChiFlowEngineSerialization_Empty(t *testing.T) {
	cave, err := newTestCave(8, 6)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}
	reg := testRegistry()
	params := DefaultFlowParams()

	engine := NewChiFlowEngine(cave, nil, reg, params)

	// Serialize.
	data, err := engine.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	// Deserialize.
	restored, err := UnmarshalChiFlowEngine(data, cave, reg, params)
	if err != nil {
		t.Fatalf("UnmarshalChiFlowEngine: %v", err)
	}

	if len(restored.Veins) != 0 {
		t.Errorf("veins: got %d, want 0", len(restored.Veins))
	}
	if len(restored.RoomChi) != 0 {
		t.Errorf("room chi: got %d, want 0", len(restored.RoomChi))
	}

	// Should be able to tick without panic.
	restored.Tick()
}

func TestUnmarshalChiFlowEngine_InvalidJSON(t *testing.T) {
	cave, err := newTestCave(8, 6)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}
	reg := testRegistry()
	params := DefaultFlowParams()

	_, err = UnmarshalChiFlowEngine([]byte("not json"), cave, reg, params)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestUnmarshalChiFlowEngine_OutOfBoundsPath(t *testing.T) {
	cave, err := newTestCave(8, 6)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}
	reg := testRegistry()
	params := DefaultFlowParams()

	// JSON with a vein whose path is out of bounds.
	data := []byte(`{"veins":[{"id":1,"source_x":0,"source_y":0,"element":0,"flow_rate":5,"path":[{"x":100,"y":100}]}],"room_chi":[]}`)
	_, err = UnmarshalChiFlowEngine(data, cave, reg, params)
	if err == nil {
		t.Error("expected error for out-of-bounds path, got nil")
	}
}

// newTestCave is a helper that creates a simple empty cave for tests.
func newTestCave(w, h int) (*world.Cave, error) {
	return world.NewCave(w, h)
}
