package scenario

import (
	"testing"
)

func TestEventEngine_OneShotFiresOnce(t *testing.T) {
	engine := NewEventEngine()

	events := []EventDef{
		{
			ID: "bonus_chi",
			Condition: ConditionDef{
				Type:   "survive_until",
				Params: map[string]any{"ticks": 10.0},
			},
			Commands: []CommandDef{
				{Type: "modify_chi", Params: map[string]any{"amount": 100.0}},
			},
			OneShot: true,
		},
	}

	snapshot := GameSnapshot{Tick: 15, CoreHP: 10}

	// First tick: should fire.
	cmds, err := engine.Tick(snapshot, events)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cmds) != 1 {
		t.Fatalf("want 1 command, got %d", len(cmds))
	}

	// Second tick: should not fire again (one-shot).
	cmds, err = engine.Tick(snapshot, events)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cmds) != 0 {
		t.Fatalf("want 0 commands on second fire, got %d", len(cmds))
	}
}

func TestEventEngine_ConditionNotMet(t *testing.T) {
	engine := NewEventEngine()

	events := []EventDef{
		{
			ID: "early_warning",
			Condition: ConditionDef{
				Type:   "survive_until",
				Params: map[string]any{"ticks": 100.0},
			},
			Commands: []CommandDef{
				{Type: "message", Params: map[string]any{"text": "warning!"}},
			},
			OneShot: false,
		},
	}

	snapshot := GameSnapshot{Tick: 50, CoreHP: 10}

	cmds, err := engine.Tick(snapshot, events)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cmds) != 0 {
		t.Fatalf("want 0 commands when condition not met, got %d", len(cmds))
	}
}

func TestEventEngine_MultipleEventsSimultaneous(t *testing.T) {
	engine := NewEventEngine()

	events := []EventDef{
		{
			ID: "event_a",
			Condition: ConditionDef{
				Type:   "survive_until",
				Params: map[string]any{"ticks": 10.0},
			},
			Commands: []CommandDef{
				{Type: "message", Params: map[string]any{"text": "event A fired"}},
			},
			OneShot: false,
		},
		{
			ID: "event_b",
			Condition: ConditionDef{
				Type:   "chi_pool",
				Params: map[string]any{"threshold": 50.0},
			},
			Commands: []CommandDef{
				{Type: "modify_chi", Params: map[string]any{"amount": -10.0}},
				{Type: "message", Params: map[string]any{"text": "event B fired"}},
			},
			OneShot: true,
		},
	}

	snapshot := GameSnapshot{Tick: 20, CoreHP: 10, ChiPoolBalance: 100.0}

	cmds, err := engine.Tick(snapshot, events)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// event_a: 1 command, event_b: 2 commands = 3 total
	if len(cmds) != 3 {
		t.Fatalf("want 3 commands, got %d", len(cmds))
	}
}

func TestEventEngine_FiredEventsPersistence(t *testing.T) {
	engine := NewEventEngine()

	oneShotEvent := EventDef{
		ID: "once_only",
		Condition: ConditionDef{
			Type:   "survive_until",
			Params: map[string]any{"ticks": 5.0},
		},
		Commands: []CommandDef{
			{Type: "message", Params: map[string]any{"text": "hello"}},
		},
		OneShot: true,
	}

	repeatingEvent := EventDef{
		ID: "repeating",
		Condition: ConditionDef{
			Type:   "survive_until",
			Params: map[string]any{"ticks": 5.0},
		},
		Commands: []CommandDef{
			{Type: "message", Params: map[string]any{"text": "again"}},
		},
		OneShot: false,
	}

	events := []EventDef{oneShotEvent, repeatingEvent}
	snapshot := GameSnapshot{Tick: 10, CoreHP: 10}

	// Tick 3 times.
	for i := 0; i < 3; i++ {
		cmds, err := engine.Tick(snapshot, events)
		if err != nil {
			t.Fatalf("tick %d: unexpected error: %v", i, err)
		}
		if i == 0 {
			// First tick: both fire (2 commands).
			if len(cmds) != 2 {
				t.Fatalf("tick 0: want 2 commands, got %d", len(cmds))
			}
		} else {
			// Subsequent ticks: only repeating fires (1 command).
			if len(cmds) != 1 {
				t.Fatalf("tick %d: want 1 command, got %d", i, len(cmds))
			}
		}
	}

	// Verify FiredEvents map state.
	if !engine.FiredEvents["once_only"] {
		t.Error("once_only should be in FiredEvents")
	}
	if engine.FiredEvents["repeating"] {
		t.Error("repeating should not be in FiredEvents")
	}
}

func TestEventEngine_NonOneShotFiresRepeatedly(t *testing.T) {
	engine := NewEventEngine()

	events := []EventDef{
		{
			ID: "repeater",
			Condition: ConditionDef{
				Type:   "survive_until",
				Params: map[string]any{"ticks": 1.0},
			},
			Commands: []CommandDef{
				{Type: "message", Params: map[string]any{"text": "tick"}},
			},
			OneShot: false,
		},
	}

	snapshot := GameSnapshot{Tick: 5, CoreHP: 10}

	for i := 0; i < 5; i++ {
		cmds, err := engine.Tick(snapshot, events)
		if err != nil {
			t.Fatalf("tick %d: unexpected error: %v", i, err)
		}
		if len(cmds) != 1 {
			t.Fatalf("tick %d: want 1 command, got %d", i, len(cmds))
		}
	}
}

func TestEventEngine_EmptyEvents(t *testing.T) {
	engine := NewEventEngine()
	snapshot := GameSnapshot{Tick: 10, CoreHP: 10}

	cmds, err := engine.Tick(snapshot, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cmds) != 0 {
		t.Fatalf("want 0 commands for empty events, got %d", len(cmds))
	}
}
