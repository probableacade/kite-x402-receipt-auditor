package receipt

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// AnomalyFlag represents detected anomalous behavior in receipt emission
type AnomalyFlag string

const (
	AnomalyBurstSpike       AnomalyFlag = "BURST_VOLUME_SPIKE"
	AnomalyDivergentPayee   AnomalyFlag = "DIVERGENT_PAYEE_CLUSTER"
	AnomalyAbnormalVelocity AnomalyFlag = "UNEXPECTED_VELOCITY"
)

// AnomalyReport contains detected behavioral anomalies for a given payer/service
type AnomalyReport struct {
	Payer           string        `json:"payer"`
	EvaluatedWindow time.Duration `json:"evaluated_window"`
	ReceiptCount    int           `json:"receipt_count"`
	Flags           []AnomalyFlag `json:"flags"`
	IsSuspicious    bool          `json:"is_suspicious"`
}

// AnomalyDetector tracks temporal issuance patterns per payer to detect automated draining or burst attacks
type AnomalyDetector struct {
	mu           sync.RWMutex
	window       time.Duration
	burstLimit   int
	payerHistory map[string][]Receipt
}

func NewAnomalyDetector(window time.Duration, burstLimit int) *AnomalyDetector {
	if burstLimit <= 0 {
		burstLimit = 20
	}
	return &AnomalyDetector{
		window:       window,
		burstLimit:   burstLimit,
		payerHistory: make(map[string][]Receipt),
	}
}

// Observe records a receipt and evaluates anomalous behavior
func (ad *AnomalyDetector) Observe(r Receipt, now time.Time) AnomalyReport {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	payer := strings.ToLower(r.Payer)
	history := ad.payerHistory[payer]

	// Prune history older than window
	cutoff := now.Add(-ad.window)
	filtered := make([]Receipt, 0, len(history)+1)
	for _, item := range history {
		if item.IssuedAt.After(cutoff) {
			filtered = append(filtered, item)
		}
	}
	filtered = append(filtered, r)
	ad.payerHistory[payer] = filtered

	flags := make([]AnomalyFlag, 0)

	// Check 1: Burst volume spike
	if len(filtered) > ad.burstLimit {
		flags = append(flags, AnomalyBurstSpike)
	}

	// Check 2: Divergent payee clustering (single payer paying to > 5 distinct payees within small window)
	payeeSet := make(map[string]bool)
	for _, item := range filtered {
		payeeSet[strings.ToLower(item.Payee)] = true
	}
	if len(payeeSet) >= 5 {
		flags = append(flags, AnomalyDivergentPayee)
	}

	return AnomalyReport{
		Payer:           payer,
		EvaluatedWindow: ad.window,
		ReceiptCount:    len(filtered),
		Flags:           flags,
		IsSuspicious:    len(flags) > 0,
	}
}

// InspectPayer returns current anomaly report without mutating state
func (ad *AnomalyDetector) InspectPayer(payer string, now time.Time) (int, error) {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	p := strings.ToLower(payer)
	history, ok := ad.payerHistory[p]
	if !ok {
		return 0, fmt.Errorf("payer %s has no recorded receipts", payer)
	}
	return len(history), nil
}
