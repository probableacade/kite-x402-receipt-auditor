package receipt

import (
	"math/big"
	"testing"
	"time"
)

func validReceipt(now time.Time) Receipt {
	return Receipt{
		ID: "pay_01", Payer: "0x1111111111111111111111111111111111111111",
		Payee: "0x2222222222222222222222222222222222222222", ChainID: 2368,
		Token: "USDC", Amount: "1.25", TxHash: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Resource: "https://api.example.com/weather", IssuedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Minute),
	}
}

func TestPolicyAcceptsValidReceipt(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	p := DefaultPolicy()
	p.Now = func() time.Time { return now }
	p.AllowedHosts["api.example.com"] = true
	d := p.Evaluate(validReceipt(now))
	if !d.Accepted || d.RiskScore != 0 {
		t.Fatalf("unexpected decision: %#v", d)
	}
}

func TestPolicyRejectsRiskyReceipt(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	p := DefaultPolicy()
	p.Now = func() time.Time { return now }
	p.MaxAmount = big.NewRat(1, 1)
	r := validReceipt(now)
	r.Amount = "9"
	r.IssuedAt = now.Add(-2 * time.Hour)
	r.ExpiresAt = now.Add(-time.Hour)
	d := p.Evaluate(r)
	if d.Accepted || len(d.Reasons) != 2 {
		t.Fatalf("expected two rejections: %#v", d)
	}
}
