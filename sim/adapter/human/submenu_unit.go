package human

import (
	"fmt"

	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
)

// SummonOption represents an available element for beast summoning.
type SummonOption struct {
	// Element is the five-element attribute.
	Element types.Element
	// Cost is the chi cost to summon a beast of this element.
	Cost float64
}

// UpgradeOption represents an existing room eligible for upgrade.
type UpgradeOption struct {
	// ID is the room ID.
	ID int
	// Name is the display name.
	Name string
	// TypeID is the room type identifier.
	TypeID string
	// Level is the current room level.
	Level int
	// UpgradeCost is the chi cost to upgrade this room.
	UpgradeCost float64
}

// UnitContext holds the context needed by unit (summon/upgrade) submenus.
type UnitContext struct {
	// SummonOptions lists available elements for summoning with their costs.
	SummonOptions []SummonOption
	// UpgradeOptions lists rooms eligible for upgrade.
	UpgradeOptions []UpgradeOption
	// Rooms lists all existing rooms in the cave.
	Rooms []RoomInfo
	// ChiBalance is the current chi pool balance.
	ChiBalance float64
}

// ShowSummonBeastMenu displays the beast summoning submenu and returns a SummonBeastAction.
// Returns nil action when the player chooses to go back.
func ShowSummonBeastMenu(ir *InputReader, ctx UnitContext) (simulation.PlayerAction, error) {
	if len(ctx.SummonOptions) == 0 {
		fmt.Fprintln(ir.out, "召喚可能な属性がありません。")
		return nil, nil
	}

	// Display summon choices.
	fmt.Fprintln(ir.out, "")
	fmt.Fprintln(ir.out, "=== 仙獣を召喚する ===")
	fmt.Fprintln(ir.out, "  0. 戻る")
	for i, opt := range ctx.SummonOptions {
		costWarning := ""
		if opt.Cost > ctx.ChiBalance {
			costWarning = " [コスト不足!]"
		}
		fmt.Fprintf(ir.out, "  %d. %s - コスト: %.1f%s\n",
			i+1, opt.Element, opt.Cost, costWarning)
	}

	// Read element choice.
	choice, err := ir.ReadIntInRange("属性> ", 0, len(ctx.SummonOptions))
	if err != nil {
		return nil, err
	}
	if choice == 0 {
		return nil, nil
	}

	selected := ctx.SummonOptions[choice-1]

	// Warn if cost is insufficient but still allow proceeding.
	if selected.Cost > ctx.ChiBalance {
		fmt.Fprintf(ir.out, "警告: コスト不足です（必要: %.1f, 残高: %.1f）\n",
			selected.Cost, ctx.ChiBalance)
	}

	return simulation.SummonBeastAction{
		Element: selected.Element,
	}, nil
}

// ShowUpgradeRoomMenu displays the room upgrade submenu and returns an UpgradeRoomAction.
// Returns nil action when the player chooses to go back.
func ShowUpgradeRoomMenu(ir *InputReader, ctx UnitContext) (simulation.PlayerAction, error) {
	if len(ctx.UpgradeOptions) == 0 {
		fmt.Fprintln(ir.out, "アップグレード可能な部屋がありません。")
		return nil, nil
	}

	// Display upgrade choices.
	fmt.Fprintln(ir.out, "")
	fmt.Fprintln(ir.out, "=== 部屋をアップグレードする ===")
	fmt.Fprintln(ir.out, "  0. 戻る")
	for i, opt := range ctx.UpgradeOptions {
		costWarning := ""
		if opt.UpgradeCost > ctx.ChiBalance {
			costWarning = " [コスト不足!]"
		}
		fmt.Fprintf(ir.out, "  %d. ID:%d %s (%s) Lv%d - コスト: %.1f%s\n",
			i+1, opt.ID, opt.Name, opt.TypeID, opt.Level, opt.UpgradeCost, costWarning)
	}

	// Read room choice.
	choice, err := ir.ReadIntInRange("部屋> ", 0, len(ctx.UpgradeOptions))
	if err != nil {
		return nil, err
	}
	if choice == 0 {
		return nil, nil
	}

	selected := ctx.UpgradeOptions[choice-1]

	// Warn if cost is insufficient but still allow proceeding.
	if selected.UpgradeCost > ctx.ChiBalance {
		fmt.Fprintf(ir.out, "警告: コスト不足です（必要: %.1f, 残高: %.1f）\n",
			selected.UpgradeCost, ctx.ChiBalance)
	}

	return simulation.UpgradeRoomAction{
		RoomID: selected.ID,
	}, nil
}
