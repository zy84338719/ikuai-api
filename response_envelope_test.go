package ikuaiapi

import (
	"encoding/json"
	"testing"
)

// The firmware answers three different envelope shapes. All three must
// unwrap to the first object.
// TestEnvelopeShapesUnwrap locks the three wrapping shapes the v4 API
// answers with, since a read-modify-write depends on finding the row
// rather than the wrapper that carries it.
func TestEnvelopeShapesUnwrap(t *testing.T) {
	cases := map[string]string{
		"data":       `{"data":[{"id":1,"tagname":"a"}]}`,
		"results":    `{"results":[{"id":2,"tagname":"b"}]}`,
		"group_data": `{"ip_data":[{"id":3,"tagname":"c"}],"ip_total":1}`,
		"static":     `{"static_data":[{"id":4,"tagname":"d"}],"static_total":7}`,
	}
	for name, body := range cases {
		got, err := FirstObject(json.RawMessage(body))
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if got["tagname"] == nil {
			t.Errorf("%s: got %#v, want the inner row", name, got)
		}
	}
}
