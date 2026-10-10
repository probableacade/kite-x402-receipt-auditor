package ledger

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/probableacade/kite-x402-receipt-auditor/internal/receipt"
)

var ErrReplay = errors.New("receipt or transaction has already been processed")

type Entry struct {
	Receipt  receipt.Receipt  `json:"receipt"`
	Decision receipt.Decision `json:"decision"`
	StoredAt time.Time        `json:"storedAt"`
}

type Ledger struct {
	mu     sync.RWMutex
	path   string
	items  []Entry
	byID   map[string]struct{}
	byHash map[string]struct{}
}

type Stats struct {
	Total    int `json:"total"`
	Accepted int `json:"accepted"`
	Rejected int `json:"rejected"`
}

func Open(path string) (*Ledger, error) {
	l := &Ledger{path: path, byID: map[string]struct{}{}, byHash: map[string]struct{}{}}
	if path == "" {
		return l, nil
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return l, nil
	}
	if err != nil {
		return nil, err
	}
	if len(b) > 0 {
		if err := json.Unmarshal(b, &l.items); err != nil {
			return nil, err
		}
		for _, item := range l.items {
			l.byID[item.Receipt.ID] = struct{}{}
			l.byHash[item.Receipt.TxHash] = struct{}{}
		}
	}
	return l, nil
}

func (l *Ledger) Append(r receipt.Receipt, d receipt.Decision) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.byID[r.ID]; ok {
		return ErrReplay
	}
	if _, ok := l.byHash[r.TxHash]; ok {
		return ErrReplay
	}
	entry := Entry{Receipt: r, Decision: d, StoredAt: time.Now().UTC()}
	l.items = append(l.items, entry)
	l.byID[r.ID] = struct{}{}
	l.byHash[r.TxHash] = struct{}{}
	if err := l.persist(); err != nil {
		l.items = l.items[:len(l.items)-1]
		delete(l.byID, r.ID)
		delete(l.byHash, r.TxHash)
		return err
	}
	return nil
}

func (l *Ledger) List(limit int) []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if limit <= 0 || limit > len(l.items) {
		limit = len(l.items)
	}
	out := make([]Entry, limit)
	for i := 0; i < limit; i++ {
		out[i] = l.items[len(l.items)-1-i]
	}
	return out
}

func (l *Ledger) Stats() Stats {
	l.mu.RLock()
	defer l.mu.RUnlock()
	stats := Stats{Total: len(l.items)}
	for _, item := range l.items {
		if item.Decision.Accepted {
			stats.Accepted++
		} else {
			stats.Rejected++
		}
	}
	return stats
}

func (l *Ledger) persist() error {
	if l.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(l.path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(l.items, "", "  ")
	if err != nil {
		return err
	}
	tmp := l.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, l.path)
}
