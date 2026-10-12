package receipt

import (
	"encoding/json"
	"time"
)


// MetricSnapshot stores aggregated verification health stats for telemetry scrapers
type MetricSnapshot struct {
	TotalEvaluated   uint64    `json:"total_evaluated"`
	AcceptedCount    uint64    `json:"accepted_count"`
	RejectedCount    uint64    `json:"rejected_count"`
	ReplayAttempts   uint64    `json:"replay_attempts"`
	AverageRiskScore float64   `json:"average_risk_score"`
	LastAuditedAt    time.Time `json:"last_audited_at"`
}

// TelemetryReporter accumulates and outputs Prometheus/JSON diagnostic feeds
type TelemetryReporter struct {
	metrics MetricSnapshot
}

func NewTelemetryReporter() *TelemetryReporter {
	return &TelemetryReporter{}
}

func (tr *TelemetryReporter) RecordDecision(d Decision) {
	tr.metrics.TotalEvaluated++
	if d.Accepted {
		tr.metrics.AcceptedCount++
	} else {
		tr.metrics.RejectedCount++
	}
	tr.metrics.LastAuditedAt = d.CheckedAt

	// Rolling average risk score
	prevTotal := float64(tr.metrics.TotalEvaluated - 1)
	tr.metrics.AverageRiskScore = (tr.metrics.AverageRiskScore*prevTotal + float64(d.RiskScore)) / float64(tr.metrics.TotalEvaluated)
}

func (tr *TelemetryReporter) RecordReplay() {
	tr.metrics.ReplayAttempts++
}

func (tr *TelemetryReporter) ExportJSON() ([]byte, error) {
	return json.MarshalIndent(tr.metrics, "", "  ")
}

func (tr *TelemetryReporter) GetMetrics() MetricSnapshot {
	return tr.metrics
}
