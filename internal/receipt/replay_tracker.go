package receipt

import (
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"
)

// ReplayTracker guarantees that previously verified transaction hashes cannot be reused
type ReplayTracker struct {
	mu           sync.RWMutex
	settledHashes map[string]time.Time
	retention    time.Duration
}

func NewReplayTracker(retention time.Duration) *ReplayTracker {
	return &ReplayTracker{
		settledHashes: make(map[string]time.Time),
		retention:    retention,
	}
}

// RecordIfUnseen checks if txHash has already been observed. If seen, returns an error.
func (rt *ReplayTracker) RecordIfUnseen(txHash string, settledAt time.Time) error {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	hash := strings.ToLower(txHash)
	if existing, found := rt.settledHashes[hash]; found {
		return fmt.Errorf("transaction %s already settled at %s (replay rejected)", hash, existing.Format(time.RFC3339))
	}

	rt.settledHashes[hash] = settledAt
	return nil
}

// Size returns total recorded replay protections
func (rt *ReplayTracker) Size() int {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return len(rt.settledHashes)
}

// PruneExpired evicts settled transactions older than retention window
func (rt *ReplayTracker) PruneExpired(now time.Time) int {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	cutoff := now.Add(-rt.retention)
	pruned := 0
	for h, ts := range rt.settledHashes {
		if ts.Before(cutoff) {
			delete(rt.settledHashes, h)
			pruned++
		}
	}
	return pruned
}

// FeeRateAuditor verifies that micropayment transaction fee margin is strictly non-negative
type FeeRateAuditor struct {
	MinFeeRatio *big.Rat
}

func NewFeeRateAuditor() *FeeRateAuditor {
	return &FeeRateAuditor{
		MinFeeRatio: big.NewRat(0, 1),
	}
}

func (f *FeeRateAuditor) VerifyZeroDeduction(claimedAmount, settledAmount string) bool {
	claimed, ok1 := new(big.Rat).SetString(claimedAmount)
	settled, ok2 := new(big.Rat).SetString(settledAmount)
	if !ok1 || !ok2 {
		return false
	}
	return settled.Cmp(claimed) >= 0
}
