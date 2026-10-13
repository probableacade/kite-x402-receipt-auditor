package receipt

import (
	"fmt"
	"testing"
	"time"
)

func TestAnomalyDetector_NormalVolume(t *testing.T) {
	detector := NewAnomalyDetector(1*time.Minute, 5)
	now := time.Now()

	r := Receipt{
		ID:       "rcpt-1",
		Payer:    "0x1111111111111111111111111111111111111111",
		Payee:    "0x2222222222222222222222222222222222222222",
		IssuedAt: now,
	}

	report := detector.Observe(r, now)
	if report.IsSuspicious {
		t.Fatalf("expected normal receipt not to be suspicious")
	}
	if report.ReceiptCount != 1 {
		t.Fatalf("expected count 1, got %d", report.ReceiptCount)
	}
}

func TestAnomalyDetector_BurstSpike(t *testing.T) {
	detector := NewAnomalyDetector(1*time.Minute, 3)
	now := time.Now()
	payer := "0x1111111111111111111111111111111111111111"

	for i := 1; i <= 4; i++ {
		r := Receipt{
			ID:       fmt.Sprintf("rcpt-%d", i),
			Payer:    payer,
			Payee:    "0x2222222222222222222222222222222222222222",
			IssuedAt: now.Add(time.Duration(i) * time.Second),
		}
		report := detector.Observe(r, now.Add(time.Duration(i)*time.Second))
		if i == 4 {
			if !report.IsSuspicious {
				t.Fatalf("expected 4th receipt to trigger burst anomaly")
			}
			foundBurst := false
			for _, flag := range report.Flags {
				if flag == AnomalyBurstSpike {
					foundBurst = true
				}
			}
			if !foundBurst {
				t.Fatalf("expected BURST_VOLUME_SPIKE flag in report")
			}
		}
	}
}

func TestAnomalyDetector_DivergentPayeeClustering(t *testing.T) {
	detector := NewAnomalyDetector(5*time.Minute, 10)
	now := time.Now()
	payer := "0x1111111111111111111111111111111111111111"

	var report AnomalyReport
	for i := 1; i <= 5; i++ {
		r := Receipt{
			ID:       fmt.Sprintf("rcpt-%d", i),
			Payer:    payer,
			Payee:    fmt.Sprintf("0x%040x", i+100),
			IssuedAt: now.Add(time.Duration(i) * time.Second),
		}
		report = detector.Observe(r, now.Add(time.Duration(i)*time.Second))
	}

	if !report.IsSuspicious {
		t.Fatalf("expected divergent payee clustering to trigger anomaly")
	}
	foundDivergent := false
	for _, flag := range report.Flags {
		if flag == AnomalyDivergentPayee {
			foundDivergent = true
		}
	}
	if !foundDivergent {
		t.Fatalf("expected DIVERGENT_PAYEE_CLUSTER flag")
	}
}
