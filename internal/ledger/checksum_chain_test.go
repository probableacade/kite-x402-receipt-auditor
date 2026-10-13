package ledger

import (
	"testing"
	"time"

	"github.com/probableacade/kite-x402-receipt-auditor/internal/receipt"
)

func TestChecksumChain_SequentialIntegrity(t *testing.T) {
	chain := NewChecksumChain()

	e1 := Entry{
		Receipt: receipt.Receipt{
			ID:     "rcpt-1",
			TxHash: "0x1111111111111111111111111111111111111111111111111111111111111111",
			Amount: "10.0",
		},
		Decision: receipt.Decision{Accepted: true},
		StoredAt: time.Now(),
	}

	e2 := Entry{
		Receipt: receipt.Receipt{
			ID:     "rcpt-2",
			TxHash: "0x2222222222222222222222222222222222222222222222222222222222222222",
			Amount: "25.0",
		},
		Decision: receipt.Decision{Accepted: false},
		StoredAt: time.Now(),
	}

	p1 := chain.Append(e1)
	if p1.Index != 1 {
		t.Fatalf("expected index 1, got %d", p1.Index)
	}

	p2 := chain.Append(e2)
	if p2.PreviousHash != p1.CurrentHash {
		t.Fatalf("expected p2.PreviousHash to equal p1.CurrentHash")
	}

	if chain.Head() != p2.CurrentHash {
		t.Fatalf("expected head to match latest hash")
	}

	// Verify ledger valid
	if !chain.VerifyLedger([]Entry{e1, e2}) {
		t.Fatalf("expected ledger verification to succeed")
	}

	// Tampered entry check
	tampered := e1
	tampered.Receipt.Amount = "999.0"
	if chain.VerifyLedger([]Entry{tampered, e2}) {
		t.Fatalf("expected tampered ledger verification to fail")
	}
}
