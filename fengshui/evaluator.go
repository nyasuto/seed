package fengshui

import (
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// Evaluator calculates feng shui scores for rooms in a cave based on chi
// levels, elemental adjacency relationships, and dragon vein connectivity.
type Evaluator struct {
	cave     *world.Cave
	registry *world.RoomTypeRegistry
	params   *ScoreParams
}

// NewEvaluator creates a new Evaluator for the given cave, room type registry,
// and scoring parameters.
func NewEvaluator(cave *world.Cave, registry *world.RoomTypeRegistry, params *ScoreParams) *Evaluator {
	return &Evaluator{
		cave:     cave,
		registry: registry,
		params:   params,
	}
}

// EvaluateRoom calculates the feng shui score for a single room.
func (ev *Evaluator) EvaluateRoom(roomID int, engine *ChiFlowEngine) FengShuiScore {
	score := FengShuiScore{RoomID: roomID}

	// 1. ChiScore = chi fill ratio × ChiRatioWeight
	rc, ok := engine.RoomChi[roomID]
	if !ok {
		return score
	}
	score.ChiScore = rc.Ratio() * ev.params.ChiRatioWeight

	// 2. AdjacencyScore = sum of elemental bonuses/penalties from neighbors
	graph := ev.cave.BuildAdjacencyGraph()
	neighbors := graph.Neighbors(roomID)
	for _, nid := range neighbors {
		nrc, nok := engine.RoomChi[nid]
		if !nok {
			continue
		}
		score.AdjacencyScore += ev.adjacencyBonus(rc.Element, nrc.Element)
	}

	// 3. DragonVeinScore = bonus if room is on any vein path
	for _, vein := range engine.Veins {
		roomIDs := vein.RoomsOnPath(ev.cave)
		for _, rid := range roomIDs {
			if rid == roomID {
				score.DragonVeinScore = ev.params.DragonVeinBonus
				break
			}
		}
		if score.DragonVeinScore > 0 {
			break
		}
	}

	// 4. Total
	score.Total = score.ChiScore + score.AdjacencyScore + score.DragonVeinScore
	return score
}

// adjacencyBonus returns the score bonus/penalty for a room with the given
// element having an adjacent room with neighborElement.
func (ev *Evaluator) adjacencyBonus(roomElem, neighborElem types.Element) float64 {
	if roomElem == neighborElem {
		return ev.params.SameElementBonus
	}
	if types.Generates(roomElem, neighborElem) {
		return ev.params.GeneratesBonus
	}
	if types.Overcomes(roomElem, neighborElem) {
		return ev.params.OvercomesPenalty
	}
	return 0
}

// EvaluateAll calculates feng shui scores for all rooms in the cave.
func (ev *Evaluator) EvaluateAll(engine *ChiFlowEngine) []FengShuiScore {
	scores := make([]FengShuiScore, 0, len(ev.cave.Rooms))
	for _, room := range ev.cave.Rooms {
		scores = append(scores, ev.EvaluateRoom(room.ID, engine))
	}
	return scores
}

// CaveTotal returns the sum of all rooms' feng shui scores.
func (ev *Evaluator) CaveTotal(engine *ChiFlowEngine) float64 {
	var total float64
	for _, room := range ev.cave.Rooms {
		s := ev.EvaluateRoom(room.ID, engine)
		total += s.Total
	}
	return total
}
