package receipt

import (
	"fmt"
	"math/big"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	evmAddress = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
	txHash     = regexp.MustCompile(`^0x[0-9a-fA-F]{64}$`)
)

type Policy struct {
	AllowedChains map[uint64]bool
	AllowedTokens map[string]bool
	AllowedHosts  map[string]bool
	MaxAmount     *big.Rat
	ClockSkew     time.Duration
	Now           func() time.Time
}

func DefaultPolicy() Policy {
	return Policy{
		AllowedChains: map[uint64]bool{2368: true},
		AllowedTokens: map[string]bool{"USDC": true},
		AllowedHosts:  map[string]bool{},
		MaxAmount:     big.NewRat(100, 1),
		ClockSkew:     30 * time.Second,
		Now:           time.Now,
	}
}

func (p Policy) Evaluate(r Receipt) Decision {
	reasons := make([]string, 0)
	risk := 0
	if strings.TrimSpace(r.ID) == "" {
		reasons = append(reasons, "missing_receipt_id")
		risk += 30
	}
	if !evmAddress.MatchString(r.Payer) {
		reasons = append(reasons, "invalid_payer")
		risk += 25
	}
	if !evmAddress.MatchString(r.Payee) {
		reasons = append(reasons, "invalid_payee")
		risk += 25
	}
	if !txHash.MatchString(r.TxHash) {
		reasons = append(reasons, "invalid_transaction_hash")
		risk += 25
	}
	if !p.AllowedChains[r.ChainID] {
		reasons = append(reasons, "chain_not_allowed")
		risk += 40
	}
	if !p.AllowedTokens[strings.ToUpper(r.Token)] {
		reasons = append(reasons, "token_not_allowed")
		risk += 40
	}
	if amount, ok := new(big.Rat).SetString(r.Amount); !ok || amount.Sign() <= 0 {
		reasons = append(reasons, "invalid_amount")
		risk += 35
	} else if p.MaxAmount != nil && amount.Cmp(p.MaxAmount) > 0 {
		reasons = append(reasons, "amount_exceeds_limit")
		risk += 35
	}
	if u, err := url.Parse(r.Resource); err != nil || u.Scheme != "https" || u.Hostname() == "" {
		reasons = append(reasons, "invalid_resource_url")
		risk += 20
	} else if len(p.AllowedHosts) > 0 && !p.AllowedHosts[strings.ToLower(u.Hostname())] {
		reasons = append(reasons, "resource_host_not_allowed")
		risk += 35
	}
	now := p.Now().UTC()
	if r.IssuedAt.After(now.Add(p.ClockSkew)) {
		reasons = append(reasons, "issued_in_future")
		risk += 25
	}
	if r.ExpiresAt.Before(now.Add(-p.ClockSkew)) {
		reasons = append(reasons, "receipt_expired")
		risk += 45
	}
	if !r.ExpiresAt.After(r.IssuedAt) {
		reasons = append(reasons, "invalid_validity_window")
		risk += 30
	}
	if risk > 100 {
		risk = 100
	}
	return Decision{ReceiptID: r.ID, Accepted: len(reasons) == 0, Reasons: reasons, RiskScore: risk, CheckedAt: now}
}

func (p Policy) Validate() error {
	if len(p.AllowedChains) == 0 {
		return fmt.Errorf("at least one chain must be allowed")
	}
	if len(p.AllowedTokens) == 0 {
		return fmt.Errorf("at least one token must be allowed")
	}
	if p.MaxAmount == nil || p.MaxAmount.Sign() <= 0 {
		return fmt.Errorf("max amount must be positive")
	}
	return nil
}
