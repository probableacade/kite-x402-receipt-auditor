package receipt

import (
	"testing"
	"time"
)

func TestTelemetryReporter_MetricsAggregation(t *testing.T) {
	reporter := NewTelemetryReporter()
	now := time.Now().UTC()

	d1 := Decision{
		ReceiptID: "pay_01",
		Accepted:  true,
		RiskScore: 0,
		CheckedAt: now,
	}
	d2 := Decision{
		ReceiptID: "pay_02",
		Accepted:  false,
		RiskScore: 60,
		CheckedAt: now,
	}

	reporter.RecordDecision(d1)
	reporter.RecordDecision(d2)
	reporter.RecordReplay()

	metrics := reporter.GetMetrics()
	if metrics.TotalEvaluated != 2 {
		t.Fatalf("expected total 2, got %d", metrics.TotalEvaluated)
	}
	if metrics.AcceptedCount != 1 || metrics.RejectedCount != 1 {
		t.Fatalf("expected 1 accepted and 1 rejected")
	}
	if metrics.ReplayAttempts != 1 {
		t.Fatalf("expected 1 replay attempt")
	}
	if metrics.AverageRiskScore != 30.0 {
		t.Fatalf("expected average risk 30.0, got %f", metrics.AverageRiskScore)
	}

	jsonData, err := reporter.ExportJSON()
	if err != nil || len(jsonData) == 0 {
		t.Fatalf("failed to export JSON: %v", err)
	}
}
