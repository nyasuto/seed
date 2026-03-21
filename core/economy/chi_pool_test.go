package economy

import (
	"errors"
	"testing"

	"github.com/nyasuto/seed/core/types"
)

func TestChiPool_DepositWithdraw(t *testing.T) {
	pool := NewChiPool(100.0)

	if pool.Balance() != 0 {
		t.Fatalf("expected initial balance 0, got %f", pool.Balance())
	}

	// Deposit
	if err := pool.Deposit(50, Supply, "vein supply", 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pool.Balance() != 50 {
		t.Fatalf("expected balance 50, got %f", pool.Balance())
	}

	// Withdraw
	if err := pool.Withdraw(20, RoomMaintenance, "room upkeep", 2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pool.Balance() != 30 {
		t.Fatalf("expected balance 30, got %f", pool.Balance())
	}

	// CanAfford
	if !pool.CanAfford(30) {
		t.Fatal("expected CanAfford(30) to be true")
	}
	if pool.CanAfford(31) {
		t.Fatal("expected CanAfford(31) to be false")
	}
}

func TestChiPool_PartialWithdraw(t *testing.T) {
	pool := NewChiPool(100.0)
	_ = pool.Deposit(20, Supply, "supply", 1)

	err := pool.Withdraw(50, Construction, "build room", 2)
	if err == nil {
		t.Fatal("expected error for insufficient chi")
	}
	if !errors.Is(err, ErrInsufficientChi) {
		t.Fatalf("expected ErrInsufficientChi, got %v", err)
	}
	if pool.Balance() != 0 {
		t.Fatalf("expected balance 0 after partial withdraw, got %f", pool.Balance())
	}

	// Check last transaction records actual withdrawn amount
	last := pool.History[len(pool.History)-1]
	if last.Amount != 20 {
		t.Fatalf("expected transaction amount 20 (actual withdrawn), got %f", last.Amount)
	}
	if last.BalanceAfter != 0 {
		t.Fatalf("expected BalanceAfter 0, got %f", last.BalanceAfter)
	}
}

func TestChiPool_CapClamp(t *testing.T) {
	pool := NewChiPool(50.0)

	_ = pool.Deposit(80, Supply, "large supply", 1)
	if pool.Balance() != 50 {
		t.Fatalf("expected balance clamped to cap 50, got %f", pool.Balance())
	}

	// Second deposit should still be clamped
	_ = pool.Deposit(10, Supply, "extra", 2)
	if pool.Balance() != 50 {
		t.Fatalf("expected balance still at cap 50, got %f", pool.Balance())
	}
}

func TestChiPool_TransactionHistory(t *testing.T) {
	pool := NewChiPool(100.0)

	_ = pool.Deposit(40, Supply, "supply tick 1", 1)
	_ = pool.Withdraw(15, BeastMaintenance, "beast upkeep", 2)
	_ = pool.Deposit(10, Reward, "invasion reward", 3)

	if len(pool.History) != 3 {
		t.Fatalf("expected 3 transactions, got %d", len(pool.History))
	}

	tests := []struct {
		idx          int
		tick         types.Tick
		amount       float64
		txType       TransactionType
		reason       string
		balanceAfter float64
	}{
		{0, 1, 40, Supply, "supply tick 1", 40},
		{1, 2, 15, BeastMaintenance, "beast upkeep", 25},
		{2, 3, 10, Reward, "invasion reward", 35},
	}

	for _, tt := range tests {
		tx := pool.History[tt.idx]
		if tx.Tick != tt.tick {
			t.Errorf("tx[%d] tick: got %d, want %d", tt.idx, tx.Tick, tt.tick)
		}
		if tx.Amount != tt.amount {
			t.Errorf("tx[%d] amount: got %f, want %f", tt.idx, tx.Amount, tt.amount)
		}
		if tx.Type != tt.txType {
			t.Errorf("tx[%d] type: got %d, want %d", tt.idx, tx.Type, tt.txType)
		}
		if tx.Reason != tt.reason {
			t.Errorf("tx[%d] reason: got %q, want %q", tt.idx, tx.Reason, tt.reason)
		}
		if tx.BalanceAfter != tt.balanceAfter {
			t.Errorf("tx[%d] balanceAfter: got %f, want %f", tt.idx, tx.BalanceAfter, tt.balanceAfter)
		}
	}
}

func TestChiPool_NegativeAmounts(t *testing.T) {
	pool := NewChiPool(100.0)

	if err := pool.Deposit(-10, Supply, "bad", 1); err == nil {
		t.Fatal("expected error for negative deposit")
	}
	if err := pool.Withdraw(-5, Construction, "bad", 2); err == nil {
		t.Fatal("expected error for negative withdraw")
	}
}

func TestChiPool_BalanceAfterAccuracy(t *testing.T) {
	pool := NewChiPool(1000.0)

	// Run a sequence and verify BalanceAfter matches actual balance at each step
	ops := []struct {
		deposit bool
		amount  float64
	}{
		{true, 100},
		{false, 30},
		{true, 50},
		{false, 80},
		{true, 200},
		{false, 240}, // exactly all remaining
	}

	for i, op := range ops {
		if op.deposit {
			_ = pool.Deposit(op.amount, Supply, "op", types.Tick(i))
		} else {
			_ = pool.Withdraw(op.amount, RoomMaintenance, "op", types.Tick(i))
		}
		last := pool.History[len(pool.History)-1]
		if last.BalanceAfter != pool.Balance() {
			t.Errorf("op %d: BalanceAfter %f != Balance() %f", i, last.BalanceAfter, pool.Balance())
		}
	}
}
