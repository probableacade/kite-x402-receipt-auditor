package ledger

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/probableacade/kite-x402-receipt-auditor/internal/receipt"
)

func TestLedgerPersistsAndBlocksReplay(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ledger.json")
	l, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	r := receipt.Receipt{ID: "pay_01", TxHash: "0xabc"}
	d := receipt.Decision{ReceiptID: r.ID, Accepted: true, CheckedAt: time.Now()}
	if err := l.Append(r, d); err != nil {
		t.Fatal(err)
	}
	if err := l.Append(r, d); !errors.Is(err, ErrReplay) {
		t.Fatalf("expected replay error, got %v", err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := reopened.List(10); len(got) != 1 || got[0].Receipt.ID != r.ID {
		t.Fatalf("unexpected entries: %#v", got)
	}
}
