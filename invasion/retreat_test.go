package invasion

import (
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

func newTestRegistry(t *testing.T) *InvaderClassRegistry {
	t.Helper()
	reg := NewInvaderClassRegistry()
	err := reg.Register(InvaderClass{
		ID:               "warrior",
		Name:             "戦士",
		Element:          types.Fire,
		BaseHP:           100,
		BaseATK:          30,
		BaseDEF:          20,
		BaseSPD:          15,
		RetreatThreshold: 0.3,
		PreferredGoal:    DestroyCore,
	})
	if err != nil {
		t.Fatalf("register warrior: %v", err)
	}
	err = reg.Register(InvaderClass{
		ID:               "thief",
		Name:             "盗賊",
		Element:          types.Metal,
		BaseHP:           80,
		BaseATK:          25,
		BaseDEF:          15,
		BaseSPD:          20,
		RetreatThreshold: 0.4,
		PreferredGoal:    StealTreasure,
	})
	if err != nil {
		t.Fatalf("register thief: %v", err)
	}
	return reg
}

func newTestInvader(id int, classID string, hp, maxHP int) *Invader {
	return &Invader{
		ID:      id,
		ClassID: classID,
		HP:      hp,
		MaxHP:   maxHP,
		State:   Advancing,
	}
}

func TestRetreatEvaluator_LowHP(t *testing.T) {
	reg := newTestRegistry(t)
	eval := NewRetreatEvaluator(reg)

	tests := []struct {
		name    string
		hp      int
		maxHP   int
		classID string
		want    bool
		reason  RetreatReason
	}{
		{
			name:    "HP below threshold triggers retreat",
			hp:      20,
			maxHP:   100,
			classID: "warrior", // threshold 0.3 → 30
			want:    true,
			reason:  ReasonLowHP,
		},
		{
			name:    "HP at threshold triggers retreat",
			hp:      30,
			maxHP:   100,
			classID: "warrior", // threshold 0.3 → 30
			want:    true,
			reason:  ReasonLowHP,
		},
		{
			name:    "HP above threshold no retreat",
			hp:      31,
			maxHP:   100,
			classID: "warrior", // threshold 0.3 → 30
			want:    false,
		},
		{
			name:    "thief higher threshold",
			hp:      40,
			maxHP:   100,
			classID: "thief", // threshold 0.4 → 40
			want:    true,
			reason:  ReasonLowHP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := newTestInvader(1, tt.classID, tt.hp, tt.maxHP)
			wave := []*Invader{inv}
			got, reason := eval.ShouldRetreat(inv, wave)
			if got != tt.want {
				t.Errorf("ShouldRetreat = %v, want %v", got, tt.want)
			}
			if got && reason != tt.reason {
				t.Errorf("reason = %v, want %v", reason, tt.reason)
			}
		})
	}
}

func TestRetreatEvaluator_MoraleBroken(t *testing.T) {
	reg := newTestRegistry(t)
	eval := NewRetreatEvaluator(reg)

	tests := []struct {
		name         string
		total        int
		defeatedCount int
		want         bool
	}{
		{"half defeated triggers retreat", 4, 2, true},
		{"more than half defeated", 4, 3, true},
		{"less than half defeated no retreat", 4, 1, false},
		{"all defeated except subject", 3, 2, true},
		{"single invader never morale break", 1, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wave := make([]*Invader, tt.total)
			// Subject invader (healthy, high HP to avoid LowHP trigger).
			wave[0] = newTestInvader(0, "warrior", 100, 100)
			for i := 1; i < tt.total; i++ {
				inv := newTestInvader(i, "warrior", 100, 100)
				if i <= tt.defeatedCount {
					inv.State = Defeated
				}
				wave[i] = inv
			}
			// Adjust exact defeated count.
			defeatedActual := 0
			for _, inv := range wave {
				if inv.State == Defeated {
					defeatedActual++
				}
			}
			// Fix count if needed.
			for defeatedActual < tt.defeatedCount {
				for _, inv := range wave {
					if inv.State != Defeated && inv != wave[0] {
						inv.State = Defeated
						defeatedActual++
						break
					}
				}
			}

			got, reason := eval.ShouldRetreat(wave[0], wave)
			if got != tt.want {
				t.Errorf("ShouldRetreat = %v, want %v (defeated=%d, total=%d)", got, tt.want, tt.defeatedCount, tt.total)
			}
			if got && reason != ReasonMoraleBroken {
				t.Errorf("reason = %v, want MoraleBroken", reason)
			}
		})
	}
}

func TestRetreatEvaluator_GoalAchieved(t *testing.T) {
	reg := newTestRegistry(t)
	eval := NewRetreatEvaluator(reg)

	inv := newTestInvader(1, "warrior", 100, 100)
	inv.State = GoalAchieved
	wave := []*Invader{inv}

	got, reason := eval.ShouldRetreat(inv, wave)
	if !got {
		t.Error("expected retreat when goal achieved")
	}
	if reason != ReasonGoalComplete {
		t.Errorf("reason = %v, want GoalComplete", reason)
	}
}

func TestRetreatEvaluator_AlreadyRetreating(t *testing.T) {
	reg := newTestRegistry(t)
	eval := NewRetreatEvaluator(reg)

	inv := newTestInvader(1, "warrior", 10, 100) // Low HP but already retreating.
	inv.State = Retreating
	wave := []*Invader{inv}

	got, _ := eval.ShouldRetreat(inv, wave)
	if got {
		t.Error("should not re-evaluate for already retreating invader")
	}
}

func TestRetreatEvaluator_AlreadyDefeated(t *testing.T) {
	reg := newTestRegistry(t)
	eval := NewRetreatEvaluator(reg)

	inv := newTestInvader(1, "warrior", 0, 100)
	inv.State = Defeated
	wave := []*Invader{inv}

	got, _ := eval.ShouldRetreat(inv, wave)
	if got {
		t.Error("should not evaluate for defeated invader")
	}
}

func TestRetreatEvaluator_PriorityOrder(t *testing.T) {
	reg := newTestRegistry(t)
	eval := NewRetreatEvaluator(reg)

	// Invader with goal achieved AND low HP — goal achieved should win (checked first).
	inv := newTestInvader(1, "warrior", 10, 100)
	inv.State = GoalAchieved
	wave := []*Invader{inv}

	_, reason := eval.ShouldRetreat(inv, wave)
	if reason != ReasonGoalComplete {
		t.Errorf("expected GoalComplete to take priority, got %v", reason)
	}
}

// makeRetreatCave creates a linear cave: Room1 -- Room2 -- Room3 -- Room4
// Room 1 is the "entry" room. Connections: 1-2, 2-3, 3-4.
func makeRetreatCave(t *testing.T) (*world.Cave, world.AdjacencyGraph) {
	t.Helper()
	cave, err := world.NewCave(20, 10)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	_, err = cave.AddRoom("chi_chamber", types.Pos{X: 1, Y: 1}, 3, 3,
		[]world.RoomEntrance{{Pos: types.Pos{X: 3, Y: 2}, Dir: types.East}})
	if err != nil {
		t.Fatalf("AddRoom 1: %v", err)
	}

	_, err = cave.AddRoom("chi_chamber", types.Pos{X: 6, Y: 1}, 3, 3,
		[]world.RoomEntrance{
			{Pos: types.Pos{X: 6, Y: 2}, Dir: types.West},
			{Pos: types.Pos{X: 8, Y: 2}, Dir: types.East},
		})
	if err != nil {
		t.Fatalf("AddRoom 2: %v", err)
	}

	_, err = cave.AddRoom("chi_chamber", types.Pos{X: 11, Y: 1}, 3, 3,
		[]world.RoomEntrance{
			{Pos: types.Pos{X: 11, Y: 2}, Dir: types.West},
			{Pos: types.Pos{X: 13, Y: 2}, Dir: types.East},
		})
	if err != nil {
		t.Fatalf("AddRoom 3: %v", err)
	}

	_, err = cave.AddRoom("chi_chamber", types.Pos{X: 16, Y: 1}, 3, 3,
		[]world.RoomEntrance{{Pos: types.Pos{X: 16, Y: 2}, Dir: types.West}})
	if err != nil {
		t.Fatalf("AddRoom 4: %v", err)
	}

	if _, err := cave.ConnectRooms(1, 2); err != nil {
		t.Fatalf("ConnectRooms 1-2: %v", err)
	}
	if _, err := cave.ConnectRooms(2, 3); err != nil {
		t.Fatalf("ConnectRooms 2-3: %v", err)
	}
	if _, err := cave.ConnectRooms(3, 4); err != nil {
		t.Fatalf("ConnectRooms 3-4: %v", err)
	}

	graph := cave.BuildAdjacencyGraph()
	return cave, graph
}

func TestRetreatPathfinder_ReverseVisitOrder(t *testing.T) {
	cave, graph := makeRetreatCave(t)
	rp := NewRetreatPathfinder(cave, graph)

	// Simulate invader that entered at room 1, visited 1→2→3→4
	inv := &Invader{
		ID:            1,
		CurrentRoomID: 4,
		Memory: &ExplorationMemory{
			VisitedRooms:    map[int]types.Tick{1: 10, 2: 20, 3: 30, 4: 40},
			KnownBeastRooms: make(map[int]bool),
		},
	}

	path := rp.FindRetreatPath(inv)
	expected := []int{4, 3, 2, 1}
	if len(path) != len(expected) {
		t.Fatalf("path length = %d, want %d; path = %v", len(path), len(expected), path)
	}
	for i, v := range expected {
		if path[i] != v {
			t.Errorf("path[%d] = %d, want %d; path = %v", i, path[i], v, path)
		}
	}
}

func TestRetreatPathfinder_AlreadyAtEntry(t *testing.T) {
	cave, graph := makeRetreatCave(t)
	rp := NewRetreatPathfinder(cave, graph)

	inv := &Invader{
		ID:            1,
		CurrentRoomID: 1,
		Memory: &ExplorationMemory{
			VisitedRooms:    map[int]types.Tick{1: 10},
			KnownBeastRooms: make(map[int]bool),
		},
	}

	path := rp.FindRetreatPath(inv)
	if len(path) != 1 || path[0] != 1 {
		t.Errorf("expected [1], got %v", path)
	}
}

func TestRetreatPathfinder_FallbackToBFS(t *testing.T) {
	cave, graph := makeRetreatCave(t)
	rp := NewRetreatPathfinder(cave, graph)

	// Invader visited rooms 1 and 4 only (skipped 2 and 3 in memory).
	// Memory-based path can't connect 4→1 directly, so BFS should kick in.
	inv := &Invader{
		ID:            1,
		CurrentRoomID: 4,
		Memory: &ExplorationMemory{
			VisitedRooms:    map[int]types.Tick{1: 10, 4: 40},
			KnownBeastRooms: make(map[int]bool),
		},
	}

	path := rp.FindRetreatPath(inv)
	// BFS shortest path: 4→3→2→1
	if path == nil {
		t.Fatal("expected non-nil path")
	}
	if path[0] != 4 {
		t.Errorf("path should start at 4, got %d", path[0])
	}
	if path[len(path)-1] != 1 {
		t.Errorf("path should end at 1, got %d", path[len(path)-1])
	}
	if len(path) != 4 {
		t.Errorf("expected BFS path length 4, got %d; path = %v", len(path), path)
	}
}

func TestRetreatPathfinder_PartialMemoryPath(t *testing.T) {
	cave, graph := makeRetreatCave(t)
	rp := NewRetreatPathfinder(cave, graph)

	// Invader visited 1→2→3 but is at room 3. Should retrace 3→2→1.
	inv := &Invader{
		ID:            1,
		CurrentRoomID: 3,
		Memory: &ExplorationMemory{
			VisitedRooms:    map[int]types.Tick{1: 10, 2: 20, 3: 30},
			KnownBeastRooms: make(map[int]bool),
		},
	}

	path := rp.FindRetreatPath(inv)
	expected := []int{3, 2, 1}
	if len(path) != len(expected) {
		t.Fatalf("path = %v, want %v", path, expected)
	}
	for i, v := range expected {
		if path[i] != v {
			t.Errorf("path[%d] = %d, want %d", i, path[i], v)
		}
	}
}

func TestRetreatResult_Fields(t *testing.T) {
	result := RetreatResult{
		InvaderID: 42,
		Reason:    ReasonGoalComplete,
		StolenChi: 15.5,
	}
	if result.InvaderID != 42 {
		t.Errorf("InvaderID = %d, want 42", result.InvaderID)
	}
	if result.Reason != ReasonGoalComplete {
		t.Errorf("Reason = %v, want GoalComplete", result.Reason)
	}
	if result.StolenChi != 15.5 {
		t.Errorf("StolenChi = %f, want 15.5", result.StolenChi)
	}
}

func TestRetreatResult_ZeroStolenChi(t *testing.T) {
	result := RetreatResult{
		InvaderID: 1,
		Reason:    ReasonLowHP,
		StolenChi: 0,
	}
	if result.StolenChi != 0 {
		t.Errorf("StolenChi = %f, want 0", result.StolenChi)
	}
}

func TestRetreatReason_String(t *testing.T) {
	tests := []struct {
		reason RetreatReason
		want   string
	}{
		{ReasonLowHP, "LowHP"},
		{ReasonMoraleBroken, "MoraleBroken"},
		{ReasonGoalComplete, "GoalComplete"},
		{RetreatReason(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.reason.String(); got != tt.want {
			t.Errorf("RetreatReason(%d).String() = %q, want %q", tt.reason, got, tt.want)
		}
	}
}
