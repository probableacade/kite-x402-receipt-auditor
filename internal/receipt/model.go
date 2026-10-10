package receipt

import "time"

// Receipt is the normalized payment evidence emitted after an x402 settlement.
type Receipt struct {
	ID        string    `json:"id"`
	Payer     string    `json:"payer"`
	Payee     string    `json:"payee"`
	ChainID   uint64    `json:"chainId"`
	Token     string    `json:"token"`
	Amount    string    `json:"amount"`
	TxHash    string    `json:"txHash"`
	Resource  string    `json:"resource"`
	IssuedAt  time.Time `json:"issuedAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// Decision captures a deterministic audit result.
type Decision struct {
	ReceiptID string   `json:"receiptId"`
	Accepted  bool     `json:"accepted"`
	Reasons   []string `json:"reasons"`
	RiskScore int      `json:"riskScore"`
	CheckedAt time.Time `json:"checkedAt"`
}
