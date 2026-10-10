package config

import (
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/probableacade/kite-x402-receipt-auditor/internal/receipt"
)

type Config struct {
	Address    string
	LedgerPath string
	Policy     receipt.Policy
}

func Load() (Config, error) {
	p := receipt.DefaultPolicy()
	p.AllowedChains = map[uint64]bool{}
	for _, raw := range csv(env("ALLOWED_CHAIN_IDS", "2368")) {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return Config{}, fmt.Errorf("invalid chain id %q: %w", raw, err)
		}
		p.AllowedChains[id] = true
	}
	p.AllowedTokens = stringSet(csv(env("ALLOWED_TOKENS", "USDC")), strings.ToUpper)
	p.AllowedHosts = stringSet(csv(os.Getenv("ALLOWED_RESOURCE_HOSTS")), strings.ToLower)
	max, ok := new(big.Rat).SetString(env("MAX_PAYMENT_AMOUNT", "100"))
	if !ok {
		return Config{}, fmt.Errorf("MAX_PAYMENT_AMOUNT must be a decimal")
	}
	p.MaxAmount = max
	skew, err := time.ParseDuration(env("CLOCK_SKEW", "30s"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid CLOCK_SKEW: %w", err)
	}
	p.ClockSkew = skew
	if err := p.Validate(); err != nil {
		return Config{}, err
	}
	return Config{Address: env("LISTEN_ADDR", ":8080"), LedgerPath: env("LEDGER_PATH", "data/ledger.json"), Policy: p}, nil
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
func csv(value string) []string {
	var out []string
	for _, v := range strings.Split(value, ",") {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}
func stringSet(values []string, normalize func(string) string) map[string]bool {
	out := map[string]bool{}
	for _, v := range values {
		out[normalize(v)] = true
	}
	return out
}
