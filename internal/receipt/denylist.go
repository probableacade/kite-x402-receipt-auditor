package receipt

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// BlacklistEntry contains revoked address metadata
type BlacklistEntry struct {
	Address   string    `json:"address"`
	Reason    string    `json:"reason"`
	RevokedAt time.Time `json:"revoked_at"`
}

// DynamicDenylist manages runtime quarantine of fraudulent payers or illicit receipt issuers
type DynamicDenylist struct {
	mu      sync.RWMutex
	blocked map[string]BlacklistEntry
}

func NewDynamicDenylist() *DynamicDenylist {
	return &DynamicDenylist{
		blocked: make(map[string]BlacklistEntry),
	}
}

// BlockAddress places an address under immediate quarantine
func (dd *DynamicDenylist) BlockAddress(address string, reason string, now time.Time) error {
	dd.mu.Lock()
	defer dd.mu.Unlock()

	addr := strings.ToLower(strings.TrimSpace(address))
	if addr == "" {
		return fmt.Errorf("cannot block empty address")
	}

	dd.blocked[addr] = BlacklistEntry{
		Address:   addr,
		Reason:    reason,
		RevokedAt: now,
	}
	return nil
}

// IsBlocked checks if either payer or payee is present in the quarantine denylist
func (dd *DynamicDenylist) IsBlocked(address string) (bool, string) {
	dd.mu.RLock()
	defer dd.mu.RUnlock()

	addr := strings.ToLower(strings.TrimSpace(address))
	if entry, found := dd.blocked[addr]; found {
		return true, entry.Reason
	}
	return false, ""
}

// UnblockAddress lifts quarantine
func (dd *DynamicDenylist) UnblockAddress(address string) bool {
	dd.mu.Lock()
	defer dd.mu.Unlock()

	addr := strings.ToLower(strings.TrimSpace(address))
	if _, found := dd.blocked[addr]; found {
		delete(dd.blocked, addr)
		return true
	}
	return false
}

// Size returns total blacklisted entries
func (dd *DynamicDenylist) Size() int {
	dd.mu.RLock()
	defer dd.mu.RUnlock()
	return len(dd.blocked)
}
