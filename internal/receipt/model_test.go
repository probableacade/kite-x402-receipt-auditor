package receipt

import (
	"encoding/json"
	"testing"
	"time"
)

func TestReceiptJSONRoundTrip(t *testing.T) {
	want := Receipt{ID: "pay_01", ChainID: 2368, Amount: "1.25", IssuedAt: time.Unix(1_700_000_000, 0).UTC()}
	b, err := json.Marshal(want)
	if err != nil { t.Fatal(err) }
	var got Receipt
	if err := json.Unmarshal(b, &got); err != nil { t.Fatal(err) }
	if got.ID != want.ID || got.ChainID != want.ChainID || got.Amount != want.Amount { t.Fatalf("round trip mismatch: %#v", got) }
}
