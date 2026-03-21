package invasion

import (
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
)

func newTestRegistry(t *testing.T) *InvaderClassRegistry {
	t.Helper()
	reg := NewInvaderClassRegistry()
	err := reg.Register(InvaderClass{
		ID:               "warrior",
		Name:             "戦士",
		Element:          types.Fire,
		BaseHP:           100,
		BaseATK:          30,
		BaseDEF:          20,
		BaseSPD:          15,
		RetreatThreshold: 0.3,
		PreferredGoal:    DestroyCore,
	})
	if err != nil {
		t.Fatalf("register warrior: %v", err)
	}
	err = reg.Register(InvaderClass{
		ID:               "thief",
		Name:             "盗賊",
		Element:          types.Metal,
		BaseHP:           80,
		BaseATK:          25,
		BaseDEF:          15,
		BaseSPD:          20,
		RetreatThreshold: 0.4,
		PreferredGoal:    StealTreasure,
	})
	if err != nil {
		t.Fatalf("register thief: %v", err)
	}
	return reg
}

func newTestInvader(id int, classID string, hp, maxHP int) *Invader {
	return &Invader{
		ID:      id,
		ClassID: classID,
		HP:      hp,
		MaxHP:   maxHP,
		State:   Advancing,
	}
}

func TestRetreatEvaluator_LowHP(t *testing.T) {
	reg := newTestRegistry(t)
	eval := NewRetreatEvaluator(reg)

	tests := []struct {
		name    string
		hp      int
		maxHP   int
		classID string
		want    bool
		reason  RetreatReason
	}{
		{
			name:    "HP below threshold triggers retreat",
			hp:      20,
			maxHP:   100,
			classID: "warrior", // threshold 0.3 → 30
			want:    true,
			reason:  ReasonLowHP,
		},
		{
			name:    "HP at threshold triggers retreat",
			hp:      30,
			maxHP:   100,
			classID: "warrior", // threshold 0.3 → 30
			want:    true,
			reason:  ReasonLowHP,
		},
		{
			name:    "HP above threshold no retreat",
			hp:      31,
			maxHP:   100,
			classID: "warrior", // threshold 0.3 → 30
			want:    false,
		},
		{
			name:    "thief higher threshold",
			hp:      40,
			maxHP:   100,
			classID: "thief", // threshold 0.4 → 40
			want:    true,
			reason:  ReasonLowHP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := newTestInvader(1, tt.classID, tt.hp, tt.maxHP)
			wave := []*Invader{inv}
			got, reason := eval.ShouldRetreat(inv, wave)
			if got != tt.want {
				t.Errorf("ShouldRetreat = %v, want %v", got, tt.want)
			}
			if got && reason != tt.reason {
				t.Errorf("reason = %v, want %v", reason, tt.reason)
			}
		})
	}
}

func TestRetreatEvaluator_MoraleBroken(t *testing.T) {
	reg := newTestRegistry(t)
	eval := NewRetreatEvaluator(reg)

	tests := []struct {
		name         string
		total        int
		defeatedCount int
		want         bool
	}{
		{"half defeated triggers retreat", 4, 2, true},
		{"more than half defeated", 4, 3, true},
		{"less than half defeated no retreat", 4, 1, false},
		{"all defeated except subject", 3, 2, true},
		{"single invader never morale break", 1, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wave := make([]*Invader, tt.total)
			// Subject invader (healthy, high HP to avoid LowHP trigger).
			wave[0] = newTestInvader(0, "warrior", 100, 100)
			for i := 1; i < tt.total; i++ {
				inv := newTestInvader(i, "warrior", 100, 100)
				if i <= tt.defeatedCount {
					inv.State = Defeated
				}
				wave[i] = inv
			}
			// Adjust exact defeated count.
			defeatedActual := 0
			for _, inv := range wave {
				if inv.State == Defeated {
					defeatedActual++
				}
			}
			// Fix count if needed.
			for defeatedActual < tt.defeatedCount {
				for _, inv := range wave {
					if inv.State != Defeated && inv != wave[0] {
						inv.State = Defeated
						defeatedActual++
						break
					}
				}
			}

			got, reason := eval.ShouldRetreat(wave[0], wave)
			if got != tt.want {
				t.Errorf("ShouldRetreat = %v, want %v (defeated=%d, total=%d)", got, tt.want, tt.defeatedCount, tt.total)
			}
			if got && reason != ReasonMoraleBroken {
				t.Errorf("reason = %v, want MoraleBroken", reason)
			}
		})
	}
}

func TestRetreatEvaluator_GoalAchieved(t *testing.T) {
	reg := newTestRegistry(t)
	eval := NewRetreatEvaluator(reg)

	inv := newTestInvader(1, "warrior", 100, 100)
	inv.State = GoalAchieved
	wave := []*Invader{inv}

	got, reason := eval.ShouldRetreat(inv, wave)
	if !got {
		t.Error("expected retreat when goal achieved")
	}
	if reason != ReasonGoalComplete {
		t.Errorf("reason = %v, want GoalComplete", reason)
	}
}

func TestRetreatEvaluator_AlreadyRetreating(t *testing.T) {
	reg := newTestRegistry(t)
	eval := NewRetreatEvaluator(reg)

	inv := newTestInvader(1, "warrior", 10, 100) // Low HP but already retreating.
	inv.State = Retreating
	wave := []*Invader{inv}

	got, _ := eval.ShouldRetreat(inv, wave)
	if got {
		t.Error("should not re-evaluate for already retreating invader")
	}
}

func TestRetreatEvaluator_AlreadyDefeated(t *testing.T) {
	reg := newTestRegistry(t)
	eval := NewRetreatEvaluator(reg)

	inv := newTestInvader(1, "warrior", 0, 100)
	inv.State = Defeated
	wave := []*Invader{inv}

	got, _ := eval.ShouldRetreat(inv, wave)
	if got {
		t.Error("should not evaluate for defeated invader")
	}
}

func TestRetreatEvaluator_PriorityOrder(t *testing.T) {
	reg := newTestRegistry(t)
	eval := NewRetreatEvaluator(reg)

	// Invader with goal achieved AND low HP — goal achieved should win (checked first).
	inv := newTestInvader(1, "warrior", 10, 100)
	inv.State = GoalAchieved
	wave := []*Invader{inv}

	_, reason := eval.ShouldRetreat(inv, wave)
	if reason != ReasonGoalComplete {
		t.Errorf("expected GoalComplete to take priority, got %v", reason)
	}
}

func TestRetreatReason_String(t *testing.T) {
	tests := []struct {
		reason RetreatReason
		want   string
	}{
		{ReasonLowHP, "LowHP"},
		{ReasonMoraleBroken, "MoraleBroken"},
		{ReasonGoalComplete, "GoalComplete"},
		{RetreatReason(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.reason.String(); got != tt.want {
			t.Errorf("RetreatReason(%d).String() = %q, want %q", tt.reason, got, tt.want)
		}
	}
}
