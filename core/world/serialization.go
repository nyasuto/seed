package world

import (
	"encoding/json"
	"fmt"

	"github.com/nyasuto/seed/core/types"
)

// jsonCell is the JSON representation of a Cell.
type jsonCell struct {
	Type   CellType `json:"type"`
	RoomID int      `json:"room_id"`
}

// jsonGrid is the JSON representation of a Grid.
type jsonGrid struct {
	Width  int         `json:"width"`
	Height int         `json:"height"`
	Cells  [][]jsonCell `json:"cells"`
}

// jsonRoomEntrance is the JSON representation of a RoomEntrance.
type jsonRoomEntrance struct {
	X   int             `json:"x"`
	Y   int             `json:"y"`
	Dir types.Direction `json:"dir"`
}

// jsonRoom is the JSON representation of a Room.
type jsonRoom struct {
	ID        int                `json:"id"`
	TypeID    string             `json:"type_id"`
	X         int                `json:"x"`
	Y         int                `json:"y"`
	Width     int                `json:"width"`
	Height    int                `json:"height"`
	Level     int                `json:"level"`
	Entrances []jsonRoomEntrance `json:"entrances"`
	BeastIDs  []int              `json:"beast_ids,omitempty"`
}

// jsonCorridor is the JSON representation of a Corridor.
type jsonCorridor struct {
	ID         int        `json:"id"`
	FromRoomID int        `json:"from_room_id"`
	ToRoomID   int        `json:"to_room_id"`
	Path       []jsonPos  `json:"path"`
}

// jsonPos is the JSON representation of a types.Pos.
type jsonPos struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// jsonCave is the JSON representation of a Cave.
type jsonCave struct {
	Grid           jsonGrid       `json:"grid"`
	Rooms          []jsonRoom     `json:"rooms"`
	Corridors      []jsonCorridor `json:"corridors"`
	NextRoomID     int            `json:"next_room_id"`
	NextCorridorID int            `json:"next_corridor_id"`
}

// MarshalJSON serializes the Cave to JSON, including the full grid, rooms, and corridors.
func (c *Cave) MarshalJSON() ([]byte, error) {
	jc := jsonCave{
		Grid:           marshalGrid(c.Grid),
		Rooms:          marshalRooms(c.Rooms),
		Corridors:      marshalCorridors(c.Corridors),
		NextRoomID:     c.nextRoomID,
		NextCorridorID: c.nextCorridorID,
	}
	return json.Marshal(jc)
}

// UnmarshalCave restores a Cave from JSON data.
func UnmarshalCave(data []byte) (*Cave, error) {
	var jc jsonCave
	if err := json.Unmarshal(data, &jc); err != nil {
		return nil, fmt.Errorf("unmarshalling cave: %w", err)
	}

	grid, err := unmarshalGrid(jc.Grid)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling grid: %w", err)
	}

	rooms := unmarshalRooms(jc.Rooms)
	corridors := unmarshalCorridors(jc.Corridors)

	return &Cave{
		Grid:           grid,
		Rooms:          rooms,
		Corridors:      corridors,
		nextRoomID:     jc.NextRoomID,
		nextCorridorID: jc.NextCorridorID,
	}, nil
}

func marshalGrid(g *Grid) jsonGrid {
	cells := make([][]jsonCell, g.Height)
	for y := 0; y < g.Height; y++ {
		cells[y] = make([]jsonCell, g.Width)
		for x := 0; x < g.Width; x++ {
			c := g.cells[y][x]
			cells[y][x] = jsonCell(c)
		}
	}
	return jsonGrid{Width: g.Width, Height: g.Height, Cells: cells}
}

func unmarshalGrid(jg jsonGrid) (*Grid, error) {
	grid, err := NewGrid(jg.Width, jg.Height)
	if err != nil {
		return nil, err
	}
	for y := 0; y < jg.Height; y++ {
		if y >= len(jg.Cells) {
			break
		}
		for x := 0; x < jg.Width; x++ {
			if x >= len(jg.Cells[y]) {
				break
			}
			jc := jg.Cells[y][x]
			grid.cells[y][x] = Cell(jc)
		}
	}
	return grid, nil
}

func marshalRooms(rooms []*Room) []jsonRoom {
	result := make([]jsonRoom, len(rooms))
	for i, r := range rooms {
		entrances := make([]jsonRoomEntrance, len(r.Entrances))
		for j, e := range r.Entrances {
			entrances[j] = jsonRoomEntrance{X: e.Pos.X, Y: e.Pos.Y, Dir: e.Dir}
		}
		result[i] = jsonRoom{
			ID:        r.ID,
			TypeID:    r.TypeID,
			X:         r.Pos.X,
			Y:         r.Pos.Y,
			Width:     r.Width,
			Height:    r.Height,
			Level:     r.Level,
			Entrances: entrances,
			BeastIDs:  r.BeastIDs,
		}
	}
	return result
}

func unmarshalRooms(jrooms []jsonRoom) []*Room {
	rooms := make([]*Room, len(jrooms))
	for i, jr := range jrooms {
		entrances := make([]RoomEntrance, len(jr.Entrances))
		for j, je := range jr.Entrances {
			entrances[j] = RoomEntrance{
				Pos: types.Pos{X: je.X, Y: je.Y},
				Dir: je.Dir,
			}
		}
		rooms[i] = &Room{
			ID:        jr.ID,
			TypeID:    jr.TypeID,
			Pos:       types.Pos{X: jr.X, Y: jr.Y},
			Width:     jr.Width,
			Height:    jr.Height,
			Level:     jr.Level,
			Entrances: entrances,
			BeastIDs:  jr.BeastIDs,
		}
	}
	return rooms
}

func marshalCorridors(corridors []Corridor) []jsonCorridor {
	result := make([]jsonCorridor, len(corridors))
	for i, c := range corridors {
		path := make([]jsonPos, len(c.Path))
		for j, p := range c.Path {
			path[j] = jsonPos{X: p.X, Y: p.Y}
		}
		result[i] = jsonCorridor{
			ID:         c.ID,
			FromRoomID: c.FromRoomID,
			ToRoomID:   c.ToRoomID,
			Path:       path,
		}
	}
	return result
}

func unmarshalCorridors(jcorridors []jsonCorridor) []Corridor {
	corridors := make([]Corridor, len(jcorridors))
	for i, jc := range jcorridors {
		path := make([]types.Pos, len(jc.Path))
		for j, jp := range jc.Path {
			path[j] = types.Pos{X: jp.X, Y: jp.Y}
		}
		corridors[i] = Corridor{
			ID:         jc.ID,
			FromRoomID: jc.FromRoomID,
			ToRoomID:   jc.ToRoomID,
			Path:       path,
		}
	}
	return corridors
}
