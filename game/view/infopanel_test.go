package view

import (
	"strings"
	"testing"

	"github.com/nyasuto/seed/core/fengshui"
	"github.com/nyasuto/seed/core/invasion"
	"github.com/nyasuto/seed/core/senju"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/core/world"
)

func TestBuildRoomInfo_BasicRoom(t *testing.T) {
	room := &world.Room{
		ID:       1,
		TypeID:   "fire_room",
		Level:    2,
		BeastIDs: nil,
	}
	rt := world.RoomType{
		ID:              "fire_room",
		Name:            "火の間",
		Element:         types.Fire,
		BaseChiCapacity: 50,
		MaxBeasts:       2,
	}
	roomChi := &fengshui.RoomChi{
		RoomID:   1,
		Current:  30,
		Capacity: 50,
	}

	data := BuildRoomInfo(room, rt, roomChi, nil)

	if data.Title != "Room #1" {
		t.Errorf("title: got %q, want %q", data.Title, "Room #1")
	}

	joined := strings.Join(data.Lines, "\n")
	if !strings.Contains(joined, "Fire") {
		t.Errorf("should contain element Fire, got:\n%s", joined)
	}
	if !strings.Contains(joined, "Lv2") {
		t.Errorf("should contain Lv2, got:\n%s", joined)
	}
	if !strings.Contains(joined, "Chi: 30/50") {
		t.Errorf("should contain Chi: 30/50, got:\n%s", joined)
	}
	if !strings.Contains(joined, "Beasts: none") {
		t.Errorf("should contain 'Beasts: none', got:\n%s", joined)
	}
}

func TestBuildRoomInfo_WithBeasts(t *testing.T) {
	room := &world.Room{
		ID:       2,
		TypeID:   "water_room",
		Level:    1,
		BeastIDs: []int{10, 20},
	}
	rt := world.RoomType{
		ID:      "water_room",
		Name:    "水の間",
		Element: types.Water,
	}
	beasts := []*senju.Beast{
		{ID: 10, Name: "Kirin", Level: 3, State: senju.Patrolling},
		{ID: 20, Name: "Byakko", Level: 1, State: senju.Fighting},
	}

	data := BuildRoomInfo(room, rt, nil, beasts)

	joined := strings.Join(data.Lines, "\n")
	if !strings.Contains(joined, "Beasts (2):") {
		t.Errorf("should show beast count, got:\n%s", joined)
	}
	if !strings.Contains(joined, "Kirin Lv3 Patrolling") {
		t.Errorf("should list Kirin, got:\n%s", joined)
	}
	if !strings.Contains(joined, "Byakko Lv1 Fighting") {
		t.Errorf("should list Byakko, got:\n%s", joined)
	}
}

func TestBuildRoomInfo_CoreRoom(t *testing.T) {
	room := &world.Room{
		ID:     1,
		TypeID: "dragon_hole",
		Level:  1,
		CoreHP: 80,
	}
	rt := world.RoomType{
		ID:         "dragon_hole",
		Name:       "龍穴",
		Element:    types.Earth,
		BaseCoreHP: 100,
	}

	data := BuildRoomInfo(room, rt, nil, nil)
	joined := strings.Join(data.Lines, "\n")
	if !strings.Contains(joined, "CoreHP: 80") {
		t.Errorf("should contain CoreHP: 80, got:\n%s", joined)
	}
}

func TestBuildBeastInfo(t *testing.T) {
	beast := &senju.Beast{
		ID:        1,
		SpeciesID: "kirin",
		Name:      "Kirin",
		Element:   types.Wood,
		Level:     5,
		HP:        40,
		MaxHP:     60,
		ATK:       15,
		DEF:       10,
		SPD:       8,
		State:     senju.Chasing,
	}

	data := BuildBeastInfo(beast)

	if data.Title != "Kirin" {
		t.Errorf("title: got %q, want %q", data.Title, "Kirin")
	}

	joined := strings.Join(data.Lines, "\n")
	if !strings.Contains(joined, "Species: kirin") {
		t.Errorf("should contain species, got:\n%s", joined)
	}
	if !strings.Contains(joined, "HP: 40/60") {
		t.Errorf("should contain HP, got:\n%s", joined)
	}
	if !strings.Contains(joined, "ATK:15 DEF:10 SPD:8") {
		t.Errorf("should contain stats, got:\n%s", joined)
	}
	if !strings.Contains(joined, "State: Chasing") {
		t.Errorf("should contain state, got:\n%s", joined)
	}
}

func TestBuildInvaderInfo(t *testing.T) {
	inv := &invasion.Invader{
		ID:      1,
		ClassID: "warrior",
		Name:    "Warrior",
		Element: types.Metal,
		Level:   3,
		HP:      25,
		MaxHP:   30,
		ATK:     12,
		DEF:     8,
		SPD:     5,
		Goal:    &invasion.DestroyCoreGoal{},
		State:   invasion.Advancing,
	}

	data := BuildInvaderInfo(inv)

	if data.Title != "Warrior" {
		t.Errorf("title: got %q, want %q", data.Title, "Warrior")
	}

	joined := strings.Join(data.Lines, "\n")
	if !strings.Contains(joined, "Class: warrior") {
		t.Errorf("should contain class, got:\n%s", joined)
	}
	if !strings.Contains(joined, "HP: 25/30") {
		t.Errorf("should contain HP, got:\n%s", joined)
	}
	if !strings.Contains(joined, "Goal: DestroyCore") {
		t.Errorf("should contain goal, got:\n%s", joined)
	}
	if !strings.Contains(joined, "State: Advancing") {
		t.Errorf("should contain state, got:\n%s", joined)
	}
}

func TestInfoPanel_Selection(t *testing.T) {
	ip := NewInfoPanel()

	if ip.SelectedRoomID() != 0 {
		t.Errorf("initial selection should be 0, got %d", ip.SelectedRoomID())
	}

	ip.SelectRoom(5)
	if ip.SelectedRoomID() != 5 {
		t.Errorf("selected room should be 5, got %d", ip.SelectedRoomID())
	}

	ip.ClearSelection()
	if ip.SelectedRoomID() != 0 {
		t.Errorf("after clear, selection should be 0, got %d", ip.SelectedRoomID())
	}
}
