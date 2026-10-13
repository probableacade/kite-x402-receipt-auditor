package receipt

import (
	"testing"
	"time"
)

func TestDynamicDenylist_QuarantineAndLift(t *testing.T) {
	dl := NewDynamicDenylist()
	now := time.Now()
	addr := "0x9999999999999999999999999999999999999999"

	blocked, _ := dl.IsBlocked(addr)
	if blocked {
		t.Fatalf("expected address not to be blocked initially")
	}

	err := dl.BlockAddress(addr, "suspected_sybil_drain", now)
	if err != nil {
		t.Fatalf("unexpected error blocking address: %v", err)
	}

	blocked, reason := dl.IsBlocked(addr)
	if !blocked {
		t.Fatalf("expected address to be blocked")
	}
	if reason != "suspected_sybil_drain" {
		t.Fatalf("expected reason suspected_sybil_drain, got %s", reason)
	}

	if dl.Size() != 1 {
		t.Fatalf("expected size 1, got %d", dl.Size())
	}

	lifted := dl.UnblockAddress(addr)
	if !lifted {
		t.Fatalf("expected address to be unblocked")
	}

	blocked, _ = dl.IsBlocked(addr)
	if blocked {
		t.Fatalf("expected address no longer blocked")
	}
}
