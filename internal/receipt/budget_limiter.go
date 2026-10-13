package receipt

import (
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"
)

// PayerBudget tracks dynamic cumulative spend caps for individual agents or accounts
type PayerBudget struct {
	Payer           string    `json:"payer"`
	MaxPeriodSpend  *big.Rat  `json:"max_period_spend"`
	CurrentSpend    *big.Rat  `json:"current_spend"`
	PeriodDuration  time.Duration `json:"period_duration"`
	PeriodStartedAt time.Time `json:"period_started_at"`
}

// BudgetLimiter provides tenant-level spending controls preventing rogue agent overspend
type BudgetLimiter struct {
	mu             sync.RWMutex
	defaultCap     *big.Rat
	periodDuration time.Duration
	budgets        map[string]*PayerBudget
}

func NewBudgetLimiter(defaultCap *big.Rat, periodDuration time.Duration) *BudgetLimiter {
	if defaultCap == nil || defaultCap.Sign() <= 0 {
		defaultCap = big.NewRat(500, 1) // 500 tokens default
	}
	if periodDuration <= 0 {
		periodDuration = 24 * time.Hour
	}
	return &BudgetLimiter{
		defaultCap:     defaultCap,
		periodDuration: periodDuration,
		budgets:        make(map[string]*PayerBudget),
	}
}

// CheckAndCharge verifies if payer has sufficient remaining budget and deducts amount
func (bl *BudgetLimiter) CheckAndCharge(payer string, amountStr string, now time.Time) error {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	amt, ok := new(big.Rat).SetString(amountStr)
	if !ok || amt.Sign() <= 0 {
		return fmt.Errorf("invalid spend amount: %s", amountStr)
	}

	p := strings.ToLower(payer)
	budget, exists := bl.budgets[p]
	if !exists {
		budget = &PayerBudget{
			Payer:           p,
			MaxPeriodSpend:  new(big.Rat).Set(bl.defaultCap),
			CurrentSpend:    new(big.Rat),
			PeriodDuration:  bl.periodDuration,
			PeriodStartedAt: now,
		}
		bl.budgets[p] = budget
	}

	// Reset window if expired
	if now.Sub(budget.PeriodStartedAt) >= budget.PeriodDuration {
		budget.CurrentSpend = new(big.Rat)
		budget.PeriodStartedAt = now
	}

	newTotal := new(big.Rat).Add(budget.CurrentSpend, amt)
	if newTotal.Cmp(budget.MaxPeriodSpend) > 0 {
		return fmt.Errorf("payer %s exceeded period budget cap (spend: %s, max: %s)",
			p, newTotal.FloatString(2), budget.MaxPeriodSpend.FloatString(2))
	}

	budget.CurrentSpend = newTotal
	return nil
}

// GetSpend returns current spend and period cap
func (bl *BudgetLimiter) GetSpend(payer string) (current string, cap string, err error) {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	p := strings.ToLower(payer)
	budget, exists := bl.budgets[p]
	if !exists {
		return "0", bl.defaultCap.FloatString(2), nil
	}
	return budget.CurrentSpend.FloatString(2), budget.MaxPeriodSpend.FloatString(2), nil
}
