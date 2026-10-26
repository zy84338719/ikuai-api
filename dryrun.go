package ikuaiapi

// Dry-run previews: a single implementation shared by every verb, so the
// preview shape cannot drift between Do/Get/doPayload.

import (
	"encoding/json"
	"net/url"
	"strings"
)

// dryRunPreview renders the JSON preview returned in DryRun mode: the method
// and path always, plus the encoded query (when params is non-empty) and the
// marshalled body (when body is non-nil).
func dryRunPreview(method, p string, body any, params map[string]string) []byte {
	preview := map[string]any{
		"dry_run": true,
		"method":  strings.ToUpper(method),
		"path":    p,
	}
	if len(params) > 0 {
		q := url.Values{}
		for k, v := range params {
			q.Set(k, v)
		}
		preview["query"] = q.Encode()
	}
	if body != nil {
		bb, _ := json.Marshal(body)
		preview["body"] = json.RawMessage(bb)
	}
	raw, _ := json.Marshal(preview)
	return raw
}
