package fengshui

import (
	"encoding/json"
	"fmt"

	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// jsonPos is the JSON representation of a types.Pos.
type jsonPos struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// jsonDragonVein is the JSON representation of a DragonVein.
type jsonDragonVein struct {
	ID        int           `json:"id"`
	SourceX   int           `json:"source_x"`
	SourceY   int           `json:"source_y"`
	Element   types.Element `json:"element"`
	FlowRate  float64       `json:"flow_rate"`
	Path      []jsonPos     `json:"path"`
}

// jsonRoomChi is the JSON representation of a RoomChi.
type jsonRoomChi struct {
	RoomID   int           `json:"room_id"`
	Current  float64       `json:"current"`
	Capacity float64       `json:"capacity"`
	Element  types.Element `json:"element"`
}

// jsonChiFlowEngine is the JSON representation of a ChiFlowEngine.
type jsonChiFlowEngine struct {
	Veins   []jsonDragonVein `json:"veins"`
	RoomChi []jsonRoomChi    `json:"room_chi"`
}

// MarshalJSON serializes the ChiFlowEngine to JSON, including all dragon veins
// and room chi states.
func (e *ChiFlowEngine) MarshalJSON() ([]byte, error) {
	je := jsonChiFlowEngine{
		Veins:   marshalVeins(e.Veins),
		RoomChi: marshalRoomChi(e.RoomChi),
	}
	return json.Marshal(je)
}

// UnmarshalChiFlowEngine restores a ChiFlowEngine from JSON data.
// The cave, registry, and params must be provided to reconstruct the engine.
// Dragon vein paths are validated against the cave to ensure consistency.
func UnmarshalChiFlowEngine(data []byte, cave *world.Cave, registry *world.RoomTypeRegistry, params *FlowParams) (*ChiFlowEngine, error) {
	var je jsonChiFlowEngine
	if err := json.Unmarshal(data, &je); err != nil {
		return nil, fmt.Errorf("unmarshalling chi flow engine: %w", err)
	}

	veins := unmarshalVeins(je.Veins)

	// Validate vein paths against the cave.
	for _, vein := range veins {
		for _, pos := range vein.Path {
			if !cave.Grid.InBounds(pos) {
				return nil, fmt.Errorf("dragon vein %d has out-of-bounds path position (%d, %d)", vein.ID, pos.X, pos.Y)
			}
		}
	}

	roomChi := unmarshalRoomChiMap(je.RoomChi)

	return &ChiFlowEngine{
		Veins:    veins,
		RoomChi:  roomChi,
		Params:   params,
		cave:     cave,
		registry: registry,
	}, nil
}

func marshalVeins(veins []*DragonVein) []jsonDragonVein {
	result := make([]jsonDragonVein, len(veins))
	for i, v := range veins {
		path := make([]jsonPos, len(v.Path))
		for j, p := range v.Path {
			path[j] = jsonPos{X: p.X, Y: p.Y}
		}
		result[i] = jsonDragonVein{
			ID:       v.ID,
			SourceX:  v.SourcePos.X,
			SourceY:  v.SourcePos.Y,
			Element:  v.Element,
			FlowRate: v.FlowRate,
			Path:     path,
		}
	}
	return result
}

func unmarshalVeins(jveins []jsonDragonVein) []*DragonVein {
	veins := make([]*DragonVein, len(jveins))
	for i, jv := range jveins {
		path := make([]types.Pos, len(jv.Path))
		for j, jp := range jv.Path {
			path[j] = types.Pos{X: jp.X, Y: jp.Y}
		}
		veins[i] = &DragonVein{
			ID:        jv.ID,
			SourcePos: types.Pos{X: jv.SourceX, Y: jv.SourceY},
			Element:   jv.Element,
			FlowRate:  jv.FlowRate,
			Path:      path,
		}
	}
	return veins
}

func marshalRoomChi(roomChi map[int]*RoomChi) []jsonRoomChi {
	result := make([]jsonRoomChi, 0, len(roomChi))
	for _, rc := range roomChi {
		result = append(result, jsonRoomChi{
			RoomID:   rc.RoomID,
			Current:  rc.Current,
			Capacity: rc.Capacity,
			Element:  rc.Element,
		})
	}
	return result
}

func unmarshalRoomChiMap(jroomChi []jsonRoomChi) map[int]*RoomChi {
	result := make(map[int]*RoomChi, len(jroomChi))
	for _, jrc := range jroomChi {
		result[jrc.RoomID] = &RoomChi{
			RoomID:   jrc.RoomID,
			Current:  jrc.Current,
			Capacity: jrc.Capacity,
			Element:  jrc.Element,
		}
	}
	return result
}
