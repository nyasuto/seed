package view

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/nyasuto/seed/core/fengshui"
	"github.com/nyasuto/seed/core/invasion"
	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/senju"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/world"
	"github.com/nyasuto/seed/game/asset"
)

const (
	infoPanelWidth   = 220
	infoPanelPadding = 6
)

// InfoPanelData holds the text lines to display in the info panel.
type InfoPanelData struct {
	Title string
	Lines []string
}

// InfoPanel renders a detail panel on the right side of the screen.
type InfoPanel struct {
	// selectedRoomID is the currently selected room, or 0 for none.
	selectedRoomID int
}

// NewInfoPanel creates a new InfoPanel.
func NewInfoPanel() *InfoPanel {
	return &InfoPanel{}
}

// SelectRoom sets the selected room ID for display.
func (ip *InfoPanel) SelectRoom(roomID int) {
	ip.selectedRoomID = roomID
}

// ClearSelection resets the selection.
func (ip *InfoPanel) ClearSelection() {
	ip.selectedRoomID = 0
}

// SelectedRoomID returns the currently selected room ID.
func (ip *InfoPanel) SelectedRoomID() int {
	return ip.selectedRoomID
}

// BuildRoomInfo generates info panel lines for a room.
func BuildRoomInfo(room *world.Room, rt world.RoomType, roomChi *fengshui.RoomChi, beasts []*senju.Beast) InfoPanelData {
	data := InfoPanelData{
		Title: fmt.Sprintf("Room #%d", room.ID),
	}
	data.Lines = append(data.Lines, fmt.Sprintf("Element: %s  Lv%d", rt.Element.String(), room.Level))

	if roomChi != nil {
		data.Lines = append(data.Lines, fmt.Sprintf("Chi: %.0f/%.0f", roomChi.Current, roomChi.Capacity))
	}

	if rt.BaseCoreHP > 0 {
		data.Lines = append(data.Lines, fmt.Sprintf("CoreHP: %d", room.CoreHP))
	}

	if len(beasts) > 0 {
		data.Lines = append(data.Lines, "---")
		data.Lines = append(data.Lines, fmt.Sprintf("Beasts (%d):", len(beasts)))
		for _, b := range beasts {
			data.Lines = append(data.Lines, fmt.Sprintf("  %s Lv%d %s", b.Name, b.Level, b.State.String()))
		}
	} else {
		data.Lines = append(data.Lines, "Beasts: none")
	}

	return data
}

// BuildBeastInfo generates info panel lines for a beast.
func BuildBeastInfo(beast *senju.Beast) InfoPanelData {
	data := InfoPanelData{
		Title: beast.Name,
	}
	data.Lines = append(data.Lines,
		fmt.Sprintf("Species: %s", beast.SpeciesID),
		fmt.Sprintf("Element: %s  Lv%d", beast.Element.String(), beast.Level),
		fmt.Sprintf("HP: %d/%d", beast.HP, beast.MaxHP),
		fmt.Sprintf("ATK:%d DEF:%d SPD:%d", beast.ATK, beast.DEF, beast.SPD),
		fmt.Sprintf("State: %s", beast.State.String()),
	)
	return data
}

// BuildInvaderInfo generates info panel lines for an invader.
func BuildInvaderInfo(inv *invasion.Invader) InfoPanelData {
	data := InfoPanelData{
		Title: inv.Name,
	}
	data.Lines = append(data.Lines,
		fmt.Sprintf("Class: %s", inv.ClassID),
		fmt.Sprintf("Element: %s  Lv%d", inv.Element.String(), inv.Level),
		fmt.Sprintf("HP: %d/%d", inv.HP, inv.MaxHP),
		fmt.Sprintf("ATK:%d DEF:%d SPD:%d", inv.ATK, inv.DEF, inv.SPD),
		fmt.Sprintf("Goal: %s", inv.Goal.Type().String()),
		fmt.Sprintf("State: %s", inv.State.String()),
	)
	return data
}

// BuildGameInfo generates info panel lines for overall game status
// when nothing is selected.
func BuildGameInfo(state *simulation.GameState, snap scenario.GameSnapshot) InfoPanelData {
	data := InfoPanelData{
		Title: state.Scenario.Name,
	}

	// Wave status.
	activeWaves := 0
	for _, w := range state.Waves {
		if w.State == invasion.Active {
			activeWaves++
		}
	}
	data.Lines = append(data.Lines,
		fmt.Sprintf("Tick: %d", snap.Tick),
		fmt.Sprintf("Waves: %d/%d", snap.DefeatedWaves, snap.TotalWaves),
		fmt.Sprintf("Active waves: %d", activeWaves),
		fmt.Sprintf("Rooms: %d  Beasts: %d", snap.RoomCount, snap.AliveBeasts),
	)

	return data
}

// BuildInfoForRoom creates InfoPanelData for a selected room, including
// its beasts from the game state.
func BuildInfoForRoom(roomID int, state *simulation.GameState) InfoPanelData {
	cave := state.Cave
	room := cave.RoomByID(roomID)
	if room == nil {
		return InfoPanelData{Title: "Unknown Room"}
	}

	rt, err := state.RoomTypeRegistry.Get(room.TypeID)
	if err != nil {
		return InfoPanelData{Title: fmt.Sprintf("Room #%d", roomID), Lines: []string{"(unknown type)"}}
	}

	var roomChi *fengshui.RoomChi
	if state.ChiFlowEngine != nil {
		roomChi = state.ChiFlowEngine.RoomChi[roomID]
	}

	var beasts []*senju.Beast
	for _, bid := range room.BeastIDs {
		for _, b := range state.Beasts {
			if b.ID == bid {
				beasts = append(beasts, b)
				break
			}
		}
	}

	return BuildRoomInfo(room, rt, roomChi, beasts)
}

// Draw renders the info panel on the right side of the screen.
func (ip *InfoPanel) Draw(screen *ebiten.Image, data InfoPanelData) {
	sw := screen.Bounds().Dx()
	sh := screen.Bounds().Dy()

	x := sw - infoPanelWidth
	y := topBarHeight

	panelHeight := sh - topBarHeight - actionBarHeight
	if panelHeight <= 0 {
		return
	}

	// Background.
	bg := ebiten.NewImage(infoPanelWidth, panelHeight)
	bg.Fill(color.RGBA{R: 0x1A, G: 0x1A, B: 0x2E, A: 0xDD})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(bg, op)

	// Left border.
	border := ebiten.NewImage(1, panelHeight)
	border.Fill(asset.ColorUIBorder)
	bop := &ebiten.DrawImageOptions{}
	bop.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(border, bop)

	// Title.
	tx := x + infoPanelPadding
	ty := y + infoPanelPadding
	if data.Title != "" {
		DrawColoredText(screen, data.Title, tx, ty, color.RGBA{R: 0xFF, G: 0xD7, B: 0x00, A: 0xFF}, 1.0)
		ty += LineHeight + 4
	}

	// Lines.
	for _, line := range data.Lines {
		if ty+LineHeight > y+panelHeight {
			break
		}
		DrawText(screen, line, tx, ty)
		ty += LineHeight
	}
}
