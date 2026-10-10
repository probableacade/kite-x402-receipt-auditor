package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/probableacade/kite-x402-receipt-auditor/internal/ledger"
	"github.com/probableacade/kite-x402-receipt-auditor/internal/receipt"
)

func testServer(t *testing.T) (*Server, time.Time) {
	t.Helper()
	now := time.Unix(1_700_000_000, 0).UTC()
	p := receipt.DefaultPolicy()
	p.Now = func() time.Time { return now }
	l, err := ledger.Open("")
	if err != nil {
		t.Fatal(err)
	}
	return New(p, l), now
}

func validBody(now time.Time) []byte {
	r := receipt.Receipt{ID: "pay_01", Payer: "0x1111111111111111111111111111111111111111", Payee: "0x2222222222222222222222222222222222222222", ChainID: 2368, Token: "USDC", Amount: "1.25", TxHash: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Resource: "https://api.example.com/weather", IssuedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Minute)}
	b, _ := json.Marshal(r)
	return b
}

func TestAuditAndReplay(t *testing.T) {
	s, now := testServer(t)
	first := httptest.NewRecorder()
	s.Handler().ServeHTTP(first, httptest.NewRequest(http.MethodPost, "/v1/audits", bytes.NewReader(validBody(now))))
	if first.Code != http.StatusOK {
		t.Fatalf("first status=%d body=%s", first.Code, first.Body.String())
	}
	second := httptest.NewRecorder()
	s.Handler().ServeHTTP(second, httptest.NewRequest(http.MethodPost, "/v1/audits", bytes.NewReader(validBody(now))))
	if second.Code != http.StatusConflict {
		t.Fatalf("second status=%d", second.Code)
	}
}

func TestRejectsUnknownFields(t *testing.T) {
	s, _ := testServer(t)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/v1/audits", bytes.NewBufferString(`{"extra":true}`)))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestStatsReflectAudits(t *testing.T) {
	s, now := testServer(t)
	s.Handler().ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/v1/audits", bytes.NewReader(validBody(now))))
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/stats", nil))
	if rr.Code != http.StatusOK || !bytes.Contains(rr.Body.Bytes(), []byte(`"accepted":1`)) {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}
