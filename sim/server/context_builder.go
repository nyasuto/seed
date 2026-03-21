package server

import (
	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/types"
	"github.com/nyasuto/seed/sim/adapter/human"
)

// GameContextBuilder implements human.ContextBuilder by reading the live
// GameState from a GameServer's engine.
type GameContextBuilder struct {
	gs *GameServer
}

// NewGameContextBuilder creates a GameContextBuilder for the given GameServer.
func NewGameContextBuilder(gs *GameServer) *GameContextBuilder {
	return &GameContextBuilder{gs: gs}
}

// BuildCtx returns the context needed by build submenus.
func (b *GameContextBuilder) BuildCtx(_ scenario.GameSnapshot) human.BuildContext {
	engine := b.gs.Engine()
	if engine == nil {
		return human.BuildContext{}
	}
	s := engine.State

	// Collect room type options.
	allTypes := s.RoomTypeRegistry.All()
	opts := make([]human.RoomTypeOption, 0, len(allTypes))
	for _, rt := range allTypes {
		cost := s.EconomyEngine.Construction.CalcRoomCost(rt.ID)
		opts = append(opts, human.RoomTypeOption{
			TypeID:  rt.ID,
			Name:    rt.Name,
			Element: rt.Element,
			Cost:    cost,
		})
	}

	// Collect existing rooms.
	rooms := make([]human.RoomInfo, 0, len(s.Cave.Rooms))
	for _, r := range s.Cave.Rooms {
		rt, _ := s.RoomTypeRegistry.Get(r.TypeID)
		rooms = append(rooms, human.RoomInfo{
			ID:     r.ID,
			TypeID: r.TypeID,
			Name:   rt.Name,
			Pos:    r.Pos,
		})
	}

	return human.BuildContext{
		RoomTypes:  opts,
		Rooms:      rooms,
		ChiBalance: s.EconomyEngine.ChiPool.Balance(),
		CaveWidth:  s.Cave.Grid.Width,
		CaveHeight: s.Cave.Grid.Height,
	}
}

// UnitCtx returns the context needed by unit submenus.
func (b *GameContextBuilder) UnitCtx(_ scenario.GameSnapshot) human.UnitContext {
	engine := b.gs.Engine()
	if engine == nil {
		return human.UnitContext{}
	}
	s := engine.State

	// Summon options: one per element with cost.
	summonOpts := make([]human.SummonOption, 0, types.ElementCount)
	for e := range types.Element(types.ElementCount) {
		cost := s.EconomyEngine.Beast.CalcSummonCost(e)
		if cost > 0 {
			summonOpts = append(summonOpts, human.SummonOption{
				Element: e,
				Cost:    cost,
			})
		}
	}

	// Collect existing rooms for display.
	rooms := make([]human.RoomInfo, 0, len(s.Cave.Rooms))
	for _, r := range s.Cave.Rooms {
		rt, _ := s.RoomTypeRegistry.Get(r.TypeID)
		rooms = append(rooms, human.RoomInfo{
			ID:     r.ID,
			TypeID: r.TypeID,
			Name:   rt.Name,
			Pos:    r.Pos,
		})
	}

	// Upgrade options: rooms eligible for upgrade.
	upgradeOpts := make([]human.UpgradeOption, 0)
	for _, r := range s.Cave.Rooms {
		rt, err := s.RoomTypeRegistry.Get(r.TypeID)
		if err != nil {
			continue
		}
		cost := s.EconomyEngine.Construction.CalcUpgradeCost(r.TypeID, r.Level)
		upgradeOpts = append(upgradeOpts, human.UpgradeOption{
			ID:          r.ID,
			Name:        rt.Name,
			TypeID:      r.TypeID,
			Level:       r.Level,
			UpgradeCost: cost,
		})
	}

	return human.UnitContext{
		SummonOptions:  summonOpts,
		UpgradeOptions: upgradeOpts,
		Rooms:          rooms,
		ChiBalance:     s.EconomyEngine.ChiPool.Balance(),
	}
}

// serverCheckpointOps implements human.CheckpointOps by delegating to a GameServer.
type serverCheckpointOps struct {
	gs *GameServer
}

// NewServerCheckpointOps creates a CheckpointOps backed by a GameServer.
func NewServerCheckpointOps(gs *GameServer) human.CheckpointOps {
	return &serverCheckpointOps{gs: gs}
}

// SaveCheckpoint saves the current game state.
func (o *serverCheckpointOps) SaveCheckpoint(path string) error {
	return o.gs.SaveCheckpointTo(path)
}

// LoadCheckpoint loads a checkpoint and restores the engine.
func (o *serverCheckpointOps) LoadCheckpoint(path string) error {
	return o.gs.LoadCheckpointFrom(path)
}

// SaveReplay saves a replay of the current game.
func (o *serverCheckpointOps) SaveReplay(path string) error {
	return o.gs.SaveReplayTo(path)
}

