package receipt

import (
	"testing"
	"time"
)

func TestReplayTracker_RecordAndReject(t *testing.T) {
	tracker := NewReplayTracker(1 * time.Hour)
	now := time.Now()
	tx := "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	if err := tracker.RecordIfUnseen(tx, now); err != nil {
		t.Fatalf("expected first insert to succeed, got: %v", err)
	}

	if err := tracker.RecordIfUnseen(tx, now); err == nil {
		t.Fatalf("expected duplicate insert to fail with replay error")
	}

	if tracker.Size() != 1 {
		t.Fatalf("expected size 1, got %d", tracker.Size())
	}
}

func TestReplayTracker_PruneExpired(t *testing.T) {
	tracker := NewReplayTracker(10 * time.Minute)
	now := time.Now()

	_ = tracker.RecordIfUnseen("0x1111111111111111111111111111111111111111111111111111111111111111", now.Add(-15*time.Minute))
	_ = tracker.RecordIfUnseen("0x2222222222222222222222222222222222222222222222222222222222222222", now.Add(-5*time.Minute))

	pruned := tracker.PruneExpired(now)
	if pruned != 1 {
		t.Fatalf("expected 1 pruned entry, got %d", pruned)
	}
	if tracker.Size() != 1 {
		t.Fatalf("expected 1 active entry left, got %d", tracker.Size())
	}
}

func TestFeeRateAuditor_ZeroDeduction(t *testing.T) {
	auditor := NewFeeRateAuditor()

	if !auditor.VerifyZeroDeduction("1.00", "1.00") {
		t.Fatalf("expected exact match to pass zero deduction check")
	}
	if !auditor.VerifyZeroDeduction("1.00", "1.05") {
		t.Fatalf("expected higher settled amount to pass zero deduction check")
	}
	if auditor.VerifyZeroDeduction("1.00", "0.99") {
		t.Fatalf("expected fee-slashed settled amount to fail zero deduction check")
	}
}
