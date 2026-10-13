package receipt

import (
	"math/big"
	"testing"
	"time"
)

func TestBudgetLimiter_AllowsSpendWithinLimit(t *testing.T) {
	limiter := NewBudgetLimiter(big.NewRat(100, 1), 1*time.Hour)
	now := time.Now()
	payer := "0x1111111111111111111111111111111111111111"

	err := limiter.CheckAndCharge(payer, "30.5", now)
	if err != nil {
		t.Fatalf("expected spend allowed, got: %v", err)
	}

	err = limiter.CheckAndCharge(payer, "40.0", now.Add(5*time.Minute))
	if err != nil {
		t.Fatalf("expected second spend allowed, got: %v", err)
	}

	current, maxCap, err := limiter.GetSpend(payer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if current != "70.50" {
		t.Fatalf("expected current spend 70.50, got %s", current)
	}
	if maxCap != "100.00" {
		t.Fatalf("expected cap 100.00, got %s", maxCap)
	}
}

func TestBudgetLimiter_RejectsOverspend(t *testing.T) {
	limiter := NewBudgetLimiter(big.NewRat(50, 1), 1*time.Hour)
	now := time.Now()
	payer := "0x2222222222222222222222222222222222222222"

	err := limiter.CheckAndCharge(payer, "40.0", now)
	if err != nil {
		t.Fatalf("expected initial spend allowed, got: %v", err)
	}

	err = limiter.CheckAndCharge(payer, "15.0", now.Add(2*time.Minute))
	if err == nil {
		t.Fatalf("expected overspend error, got nil")
	}
}

func TestBudgetLimiter_ResetsPeriodAutomatically(t *testing.T) {
	limiter := NewBudgetLimiter(big.NewRat(50, 1), 1*time.Hour)
	now := time.Now()
	payer := "0x3333333333333333333333333333333333333333"

	_ = limiter.CheckAndCharge(payer, "45.0", now)

	// Advance past 1 hour window
	future := now.Add(65 * time.Minute)
	err := limiter.CheckAndCharge(payer, "20.0", future)
	if err != nil {
		t.Fatalf("expected new period reset to permit spend, got: %v", err)
	}

	current, _, _ := limiter.GetSpend(payer)
	if current != "20.00" {
		t.Fatalf("expected current spend reset to 20.00, got %s", current)
	}
}
