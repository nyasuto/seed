package senju

import (
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// setupAIIntegrationCave creates a 5-room cave with corridors for AI testing.
//
// Layout (32x32):
//
//	Room 1 (senju_room/Wood)  at (2,2)   3x3  ← Guard beast
//	Room 2 (senju_room/Fire)  at (8,2)   3x3  ← Patrol beast
//	Room 3 (senju_room/Earth) at (14,2)  3x3
//	Room 4 (senju_room/Metal) at (20,2)  3x3  ← Chase beast
//	Room 5 (recovery_room/Water) at (26,2) 3x3
//
// Corridors: 1↔2, 2↔3, 3↔4, 4↔5
func setupAIIntegrationCave(t *testing.T) (*world.Cave, world.AdjacencyGraph, *world.RoomTypeRegistry, map[int]*world.Room) {
	t.Helper()

	reg := world.NewRoomTypeRegistry()
	for _, rt := range []world.RoomType{
		{ID: "senju_room", Name: "仙獣部屋", Element: types.Wood, BaseChiCapacity: 100, MaxBeasts: 3},
		{ID: "senju_fire", Name: "仙獣部屋火", Element: types.Fire, BaseChiCapacity: 100, MaxBeasts: 3},
		{ID: "senju_earth", Name: "仙獣部屋土", Element: types.Earth, BaseChiCapacity: 100, MaxBeasts: 3},
		{ID: "senju_metal", Name: "仙獣部屋金", Element: types.Metal, BaseChiCapacity: 100, MaxBeasts: 3},
		{ID: "recovery_room", Name: "回復室", Element: types.Water, BaseChiCapacity: 50, MaxBeasts: 3},
	} {
		if err := reg.Register(rt); err != nil {
			t.Fatal(err)
		}
	}

	cave, err := world.NewCave(32, 32)
	if err != nil {
		t.Fatal(err)
	}

	type roomDef struct {
		typeID    string
		pos       types.Pos
		entrances []world.RoomEntrance
	}
	defs := []roomDef{
		{"senju_room", types.Pos{X: 2, Y: 2}, []world.RoomEntrance{
			{Pos: types.Pos{X: 4, Y: 3}, Dir: types.East},
		}},
		{"senju_fire", types.Pos{X: 8, Y: 2}, []world.RoomEntrance{
			{Pos: types.Pos{X: 8, Y: 3}, Dir: types.West},
			{Pos: types.Pos{X: 10, Y: 3}, Dir: types.East},
		}},
		{"senju_earth", types.Pos{X: 14, Y: 2}, []world.RoomEntrance{
			{Pos: types.Pos{X: 14, Y: 3}, Dir: types.West},
			{Pos: types.Pos{X: 16, Y: 3}, Dir: types.East},
		}},
		{"senju_metal", types.Pos{X: 20, Y: 2}, []world.RoomEntrance{
			{Pos: types.Pos{X: 20, Y: 3}, Dir: types.West},
			{Pos: types.Pos{X: 22, Y: 3}, Dir: types.East},
		}},
		{"recovery_room", types.Pos{X: 26, Y: 2}, []world.RoomEntrance{
			{Pos: types.Pos{X: 26, Y: 3}, Dir: types.West},
		}},
	}

	var roomIDs []int
	for _, d := range defs {
		r, err := cave.AddRoom(d.typeID, d.pos, 3, 3, d.entrances)
		if err != nil {
			t.Fatalf("AddRoom %s: %v", d.typeID, err)
		}
		roomIDs = append(roomIDs, r.ID)
	}

	// Connect rooms linearly: 1↔2, 2↔3, 3↔4, 4↔5
	for i := 0; i < len(roomIDs)-1; i++ {
		if _, err := cave.ConnectRooms(roomIDs[i], roomIDs[i+1]); err != nil {
			t.Fatalf("ConnectRooms %d↔%d: %v", roomIDs[i], roomIDs[i+1], err)
		}
	}

	ag := cave.BuildAdjacencyGraph()
	rooms := make(map[int]*world.Room)
	for _, id := range roomIDs {
		rooms[id] = cave.RoomByID(id)
	}

	return cave, ag, reg, rooms
}

// TestAIIntegration_20TickSimulation verifies the full AI behavior lifecycle:
//   - Guard beast stays in its assigned room
//   - Patrol beast visits multiple rooms
//   - Invader placement triggers Chase movement toward invader
//   - HP reduction triggers Flee transition
func TestAIIntegration_20TickSimulation(t *testing.T) {
	cave, ag, reg, rooms := setupAIIntegrationCave(t)

	speciesReg, err := LoadDefaultSpecies()
	if err != nil {
		t.Fatalf("LoadDefaultSpecies: %v", err)
	}

	// Collect room IDs in order
	var roomIDs []int
	for id := range rooms {
		roomIDs = append(roomIDs, id)
	}
	// Sort for deterministic ordering
	for i := 0; i < len(roomIDs); i++ {
		for j := i + 1; j < len(roomIDs); j++ {
			if roomIDs[i] > roomIDs[j] {
				roomIDs[i], roomIDs[j] = roomIDs[j], roomIDs[i]
			}
		}
	}

	room1ID := roomIDs[0] // Wood senju_room
	room2ID := roomIDs[1] // Fire senju_fire
	room4ID := roomIDs[3] // Metal senju_metal

	// Create 3 beasts
	suiryu, _ := speciesReg.Get("suiryu")   // Wood
	enhou, _ := speciesReg.Get("enhou")      // Fire
	kinrou, _ := speciesReg.Get("kinrou")    // Metal

	guardBeast := NewBeast(1, suiryu, 0)
	patrolBeast := NewBeast(2, enhou, 0)
	chaseBeast := NewBeast(3, kinrou, 0)

	// Place beasts
	rt1, _ := reg.Get("senju_room")
	rt2, _ := reg.Get("senju_fire")
	rt4, _ := reg.Get("senju_metal")

	if err := PlaceBeast(guardBeast, rooms[room1ID], rt1); err != nil {
		t.Fatalf("PlaceBeast guard: %v", err)
	}
	if err := PlaceBeast(patrolBeast, rooms[room2ID], rt2); err != nil {
		t.Fatalf("PlaceBeast patrol: %v", err)
	}
	if err := PlaceBeast(chaseBeast, rooms[room4ID], rt4); err != nil {
		t.Fatalf("PlaceBeast chase: %v", err)
	}

	// Create behavior engine and assign behaviors
	engine := NewBehaviorEngine(cave, ag, reg, nil)
	engine.AssignBehavior(guardBeast, Guard)
	engine.AssignBehavior(patrolBeast, Patrol)
	engine.AssignBehavior(chaseBeast, Chase)

	beasts := []*Beast{guardBeast, patrolBeast, chaseBeast}
	noInvaders := map[int][]int{}

	// --- Phase 1: 20 ticks with no invaders ---
	// Track patrol beast room visits
	patrolVisited := map[int]bool{}

	for tick := 0; tick < 20; tick++ {
		actions := engine.Tick(beasts, noInvaders, nil)
		if err := ApplyActions(beasts, rooms, reg, actions); err != nil {
			t.Fatalf("ApplyActions tick %d: %v", tick, err)
		}
		patrolVisited[patrolBeast.RoomID] = true
	}

	// Guard beast must stay in room 1
	if guardBeast.RoomID != room1ID {
		t.Errorf("Guard beast should stay in room %d, got room %d", room1ID, guardBeast.RoomID)
	}

	// Patrol beast should have visited multiple rooms
	if len(patrolVisited) < 2 {
		t.Errorf("Patrol beast should visit at least 2 rooms, visited %d: %v", len(patrolVisited), patrolVisited)
	}

	// --- Phase 2: Introduce invader, verify Chase behavior ---
	// Place invader in room 3 (adjacent to chase beast's current area)
	room3ID := roomIDs[2]
	invaders := map[int][]int{
		room3ID: {99}, // invader ID 99 in room 3
	}

	// Move chase beast back to room 4 for controlled test
	if chaseBeast.RoomID != room4ID {
		MoveBeast(chaseBeast, rooms[chaseBeast.RoomID], rooms[room4ID], rt4)
	}
	// Assign fresh Chase targeting the invader
	engine.SetBehavior(chaseBeast.ID, NewChaseBehavior(99, 10))

	initialChaseRoom := chaseBeast.RoomID
	chaseMoved := false

	for tick := 0; tick < 10; tick++ {
		actions := engine.Tick(beasts, invaders, nil)
		if err := ApplyActions(beasts, rooms, reg, actions); err != nil {
			t.Fatalf("ApplyActions chase tick %d: %v", tick, err)
		}
		if chaseBeast.RoomID != initialChaseRoom {
			chaseMoved = true
		}
		// If chase beast reached invader room, verify attack
		if chaseBeast.RoomID == room3ID {
			break
		}
	}

	if !chaseMoved {
		t.Error("Chase beast should have moved toward invader but stayed in place")
	}

	// Chase beast should have moved toward room 3 (the invader room)
	// Room 4 is adjacent to room 3, so it should reach in 1 tick
	if chaseBeast.RoomID != room3ID {
		t.Errorf("Chase beast should reach invader room %d, got room %d", room3ID, chaseBeast.RoomID)
	}

	// --- Phase 3: HP reduction → Flee transition ---
	// Lower guard beast HP to trigger flee
	guardBeast.HP = guardBeast.MaxHP / 5 // 20% HP, below 25% threshold

	// Place invader in guard beast's room to make flee meaningful
	fleeInvaders := map[int][]int{
		room1ID: {100},
	}

	// Run a few ticks — guard should transition to Flee
	for tick := 0; tick < 5; tick++ {
		actions := engine.Tick(beasts, fleeInvaders, nil)
		if err := ApplyActions(beasts, rooms, reg, actions); err != nil {
			t.Fatalf("ApplyActions flee tick %d: %v", tick, err)
		}
	}

	// Guard beast should have fled from room 1
	if guardBeast.RoomID == room1ID {
		t.Error("Guard beast with low HP should have fled from room 1")
	}

	// Verify the beast transitioned to fleeing or recovering state
	if guardBeast.State != Recovering && guardBeast.State != Idle {
		// Flee behavior moves beast away; state depends on whether recovery room reached
		// At minimum, beast should not be in Fighting state
		if guardBeast.State == Fighting {
			t.Errorf("Fleeing beast should not be Fighting, got state %d", guardBeast.State)
		}
	}

	// --- Verify Guard beast's flee behavior is active ---
	b := engine.GetBehavior(guardBeast.ID)
	if b == nil {
		t.Fatal("Guard beast should have a behavior assigned")
	}
	if b.Type() != Flee {
		t.Errorf("Guard beast with low HP should have Flee behavior, got %d", b.Type())
	}
}
