package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/probableacade/kite-x402-receipt-auditor/internal/dashboard"
	"github.com/probableacade/kite-x402-receipt-auditor/internal/ledger"
	"github.com/probableacade/kite-x402-receipt-auditor/internal/receipt"
)

type Server struct {
	policy receipt.Policy
	ledger *ledger.Ledger
}

func New(policy receipt.Policy, l *ledger.Ledger) *Server { return &Server{policy: policy, ledger: l} }

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("POST /v1/audits", s.audit)
	mux.HandleFunc("GET /v1/audits", s.list)
	mux.HandleFunc("GET /v1/stats", s.stats)
	mux.Handle("GET /", dashboard.Handler())
	return withHeaders(mux)
}

func (s *Server) stats(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.ledger.Stats())
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) audit(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var candidate receipt.Receipt
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&candidate); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_receipt", err.Error())
		return
	}
	decision := s.policy.Evaluate(candidate)
	if err := s.ledger.Append(candidate, decision); err != nil {
		if errors.Is(err, ledger.ErrReplay) {
			writeError(w, http.StatusConflict, "replay_detected", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "ledger_failure", "could not persist decision")
		return
	}
	status := http.StatusOK
	if !decision.Accepted {
		status = http.StatusUnprocessableEntity
	}
	writeJSON(w, status, decision)
}

func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 200 {
			writeError(w, http.StatusBadRequest, "invalid_limit", "limit must be between 1 and 200")
			return
		}
		limit = parsed
	}
	writeJSON(w, http.StatusOK, map[string]any{"audits": s.ledger.List(limit)})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}
func withHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		if !strings.HasPrefix(r.URL.Path, "/v1/") && r.URL.Path != "/healthz" && r.Method != http.MethodGet {
			writeError(w, http.StatusNotFound, "not_found", fmt.Sprintf("route %s not found", r.URL.Path))
			return
		}
		next.ServeHTTP(w, r)
	})
}
