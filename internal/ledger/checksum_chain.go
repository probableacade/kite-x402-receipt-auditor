package ledger

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
)

// AuditProof represents a cryptographic inclusion or chronological state proof of ledger entries
type AuditProof struct {
	Index        int    `json:"index"`
	ReceiptID    string `json:"receipt_id"`
	PreviousHash string `json:"previous_hash"`
	CurrentHash  string `json:"current_hash"`
	Verified     bool   `json:"verified"`
}

// ChecksumChain provides tamper-evident sequential chaining of ledger entries
type ChecksumChain struct {
	mu     sync.RWMutex
	hashes []string
}

func NewChecksumChain() *ChecksumChain {
	return &ChecksumChain{
		hashes: []string{"0000000000000000000000000000000000000000000000000000000000000000"},
	}
}

// ComputeEntryHash produces a deterministic SHA-256 digest of an entry and its predecessor
func (cc *ChecksumChain) ComputeEntryHash(prevHash string, entry Entry) string {
	hasher := sha256.New()
	hasher.Write([]byte(prevHash))
	hasher.Write([]byte(entry.Receipt.ID))
	hasher.Write([]byte(entry.Receipt.TxHash))
	hasher.Write([]byte(entry.Receipt.Amount))
	hasher.Write([]byte(fmt.Sprintf("%t", entry.Decision.Accepted)))
	return hex.EncodeToString(hasher.Sum(nil))
}

// Append calculates the next link in the checksum chain
func (cc *ChecksumChain) Append(entry Entry) AuditProof {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	prev := cc.hashes[len(cc.hashes)-1]
	curr := cc.ComputeEntryHash(prev, entry)
	cc.hashes = append(cc.hashes, curr)

	return AuditProof{
		Index:        len(cc.hashes) - 1,
		ReceiptID:    entry.Receipt.ID,
		PreviousHash: prev,
		CurrentHash:  curr,
		Verified:     true,
	}
}

// Head returns the latest cryptographic state digest
func (cc *ChecksumChain) Head() string {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.hashes[len(cc.hashes)-1]
}

// VerifyLedger audits the complete chain against given entries for tamper detection
func (cc *ChecksumChain) VerifyLedger(entries []Entry) bool {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	if len(entries) != len(cc.hashes)-1 {
		return false
	}

	curr := "0000000000000000000000000000000000000000000000000000000000000000"
	for i, entry := range entries {
		nextHash := cc.ComputeEntryHash(curr, entry)
		if nextHash != cc.hashes[i+1] {
			return false
		}
		curr = nextHash
	}
	return true
}
