package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	for _, key := range []string{"LISTEN_ADDR", "LEDGER_PATH", "ALLOWED_CHAIN_IDS", "ALLOWED_TOKENS", "ALLOWED_RESOURCE_HOSTS", "MAX_PAYMENT_AMOUNT", "CLOCK_SKEW"} {
		t.Setenv(key, "")
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Address != ":8080" || !cfg.Policy.AllowedChains[2368] || !cfg.Policy.AllowedTokens["USDC"] {
		t.Fatalf("unexpected config: %#v", cfg)
	}
}

func TestRejectsInvalidMaximum(t *testing.T) {
	t.Setenv("MAX_PAYMENT_AMOUNT", "many")
	if _, err := Load(); err == nil {
		t.Fatal("expected invalid decimal error")
	}
}
