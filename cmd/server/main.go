package main

import (
	"log"
	"net/http"
	"time"

	"github.com/probableacade/kite-x402-receipt-auditor/internal/api"
	"github.com/probableacade/kite-x402-receipt-auditor/internal/config"
	"github.com/probableacade/kite-x402-receipt-auditor/internal/ledger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	l, err := ledger.Open(cfg.LedgerPath)
	if err != nil {
		log.Fatal(err)
	}
	srv := &http.Server{Addr: cfg.Address, Handler: api.New(cfg.Policy, l).Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("receipt auditor listening on %s", cfg.Address)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
