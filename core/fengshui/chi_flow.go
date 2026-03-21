package fengshui

import (
	"sort"

	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

// ChiFlowEngine simulates the flow of chi energy through a cave's dragon veins
// and rooms. Each tick, chi is supplied by dragon veins, propagated between
// adjacent rooms, and reduced by natural decay.
type ChiFlowEngine struct {
	// Veins is the set of dragon veins carrying chi through the cave.
	Veins []*DragonVein
	// RoomChi maps room ID to its current chi state.
	RoomChi map[int]*RoomChi
	// Params controls flow multipliers and decay rate.
	Params *FlowParams
	// cave is the underlying cave structure (used for adjacency lookups).
	cave *world.Cave
	// registry provides room type information (element, capacity).
	registry *world.RoomTypeRegistry
}

// NewChiFlowEngine creates a new ChiFlowEngine for the given cave, dragon veins,
// room type registry, and flow parameters. It initializes a RoomChi entry for
// every room in the cave.
func NewChiFlowEngine(cave *world.Cave, veins []*DragonVein, registry *world.RoomTypeRegistry, params *FlowParams) *ChiFlowEngine {
	e := &ChiFlowEngine{
		Veins:    veins,
		RoomChi:  make(map[int]*RoomChi),
		Params:   params,
		cave:     cave,
		registry: registry,
	}
	for _, room := range cave.Rooms {
		e.ensureRoomChi(room)
	}
	return e
}

// ensureRoomChi adds a RoomChi entry for the given room if one does not exist.
func (e *ChiFlowEngine) ensureRoomChi(room *world.Room) {
	if _, ok := e.RoomChi[room.ID]; ok {
		return
	}
	rt, err := e.registry.Get(room.TypeID)
	var elem types.Element
	var cap float64
	if err == nil {
		elem = rt.Element
		cap = float64(rt.BaseChiCapacity)
	}
	e.RoomChi[room.ID] = &RoomChi{
		RoomID:   room.ID,
		Current:  0,
		Capacity: cap,
		Element:  elem,
	}
}

// SyncNewRooms registers RoomChi entries for any rooms in the cave that are
// not yet tracked. Unlike OnCaveChanged, this does not rebuild dragon veins,
// preserving existing vein paths and game behavior.
func (e *ChiFlowEngine) SyncNewRooms(cave *world.Cave) {
	for _, room := range cave.Rooms {
		e.ensureRoomChi(room)
	}
}

// elementMultiplier returns the flow multiplier for chi flowing from a source
// element into a room with the given element.
func (e *ChiFlowEngine) elementMultiplier(source, room types.Element) float64 {
	if source == room {
		return e.Params.SameElementMultiplier
	}
	if types.Generates(source, room) {
		return e.Params.GeneratesMultiplier
	}
	if types.Overcomes(source, room) {
		return e.Params.OvercomesMultiplier
	}
	return e.Params.NeutralMultiplier
}

// Tick advances the chi flow simulation by one tick. The update proceeds in
// three phases:
//  1. Dragon vein supply: each vein delivers chi to rooms on its path.
//  2. Adjacency propagation: chi flows from high to low between connected rooms.
//  3. Decay: all rooms lose a fraction of their chi.
func (e *ChiFlowEngine) Tick() {
	// Phase 1: Dragon vein supply.
	for _, vein := range e.Veins {
		roomIDs := vein.RoomsOnPath(e.cave)
		for _, rid := range roomIDs {
			rc, ok := e.RoomChi[rid]
			if !ok {
				continue
			}
			mult := e.elementMultiplier(vein.Element, rc.Element)
			rc.Current += vein.FlowRate * mult
		}
	}

	// Phase 2: Adjacency propagation.
	// Calculate deltas first, then apply (to avoid order-dependent results).
	// Iterate rooms in sorted ID order for deterministic float accumulation.
	graph := e.cave.BuildAdjacencyGraph()
	deltas := make(map[int]float64)
	sortedRoomIDs := make([]int, 0, len(e.RoomChi))
	for rid := range e.RoomChi {
		deltas[rid] = 0
		sortedRoomIDs = append(sortedRoomIDs, rid)
	}
	sort.Ints(sortedRoomIDs)

	// Track processed pairs to avoid double-counting.
	type pair struct{ a, b int }
	processed := make(map[pair]bool)

	for _, rid := range sortedRoomIDs {
		rc := e.RoomChi[rid]
		neighbors := graph.Neighbors(rid)
		for _, nid := range neighbors {
			p := pair{rid, nid}
			if rid > nid {
				p = pair{nid, rid}
			}
			if processed[p] {
				continue
			}
			processed[p] = true

			nrc, ok := e.RoomChi[nid]
			if !ok {
				continue
			}
			diff := rc.Current - nrc.Current
			if diff == 0 {
				continue
			}

			// Transfer a fraction of the difference.
			const transferRate = 0.1
			transfer := diff * transferRate

			// Apply element multiplier based on the direction of flow.
			var mult float64
			if transfer > 0 {
				// Chi flows from rc to nrc.
				mult = e.elementMultiplier(rc.Element, nrc.Element)
			} else {
				// Chi flows from nrc to rc.
				mult = e.elementMultiplier(nrc.Element, rc.Element)
			}
			transfer *= mult

			deltas[rid] -= transfer
			deltas[nid] += transfer
		}
	}

	for rid, d := range deltas {
		e.RoomChi[rid].Current += d
	}

	// Phase 3: Decay and clamp.
	for _, rc := range e.RoomChi {
		rc.Current -= rc.Current * e.Params.BaseDecayRate
		// Clamp to [0, Capacity].
		if rc.Current < 0 {
			rc.Current = 0
		}
		if rc.Capacity > 0 && rc.Current > rc.Capacity {
			rc.Current = rc.Capacity
		}
	}
}

// OnCaveChanged rebuilds all dragon veins based on the current cave state
// and adds RoomChi entries for any new rooms. Call this after modifying the
// cave (adding rooms, digging corridors, etc.).
func (e *ChiFlowEngine) OnCaveChanged(cave *world.Cave) {
	e.cave = cave

	// Rebuild all dragon veins.
	for i, vein := range e.Veins {
		rebuilt, err := RebuildDragonVein(cave, vein)
		if err != nil {
			// If rebuild fails, keep the old vein.
			continue
		}
		e.Veins[i] = rebuilt
	}

	// Add RoomChi for any new rooms.
	for _, room := range cave.Rooms {
		e.ensureRoomChi(room)
	}
}
