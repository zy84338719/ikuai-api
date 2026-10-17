// Package ikuaiapi is a Go SDK for the iKuai v4.0 REST API — the local
// management API exposed by routers running iKuai OS 4.x.
//
// # Getting started
//
// Obtain an API token in the router web UI (System -> Auth -> API Token),
// then build a client. Everything else hangs off the typed service layer.
//
//	client, err := ikuaiapi.NewClient("https://192.168.1.1",
//	    ikuaiapi.WithToken(token),
//	    ikuaiapi.WithTimeout(15*time.Second),
//	)
//	if err != nil {
//	    return err
//	}
//	defer client.Close(context.Background())
//
//	api := service.NewAPIClient(client)
//	sys, err := api.Monitoring().GetMonitoringSystem(ctx)
//
// Credentials can also come from a session file or the environment; see
// the session package.
//
// # The v4 API is not uniform
//
// Two facts shape most of this SDK's design, and both are learned the hard
// way from real hardware:
//
//   - The page-size query parameter is named limit, not page_size. Sending
//     page_size makes the router ignore the whole paging query, silently,
//     so a "paged" read quietly returns everything.
//   - Collections do not all speak the same query dialect. Most take
//     page/limit/filter/order/order_by, but several monitoring endpoints
//     want an uppercase TYPE/LIMIT/OFFSET form, /auth/online-users mixes
//     lowercase paging with uppercase search keys, and a family of
//     endpoints uses starttime/stoptime. Sending the wrong shape is also
//     silently ignored.
//
// Both are encoded in the catalog rather than left to the caller. Each
// endpoint records its dialect in V4Endpoint.Query, and the generated
// service layer exposes a typed Query<Name> helper for the endpoints that
// need one. See v4_catalog.go and skills/SKILL.md.
//
// # Writing
//
// The firmware validates writes strictly and, for most rule-like resources,
// rejects a partial update rather than filling the gaps with its own
// defaults — which silently reconfigures whatever the caller left out.
// V4Schemas records each endpoint's create defaults, update allowlist and
// required fields; CreateBody and UpdatePayload apply them. See
// fields.go and skills/writing.md.
//
// # Errors
//
// Errors follow the oekit contract, so an error raised here is understood
// by an oekit service without an adapter. Every APIError carries an
// oekit BizError: the iKuai code becomes the business code, and an expired
// token (3007) answers 401, a permission failure (10014) answers 403, and a
// bad payload (3001) answers 400. A handler calling errors.As, including
// oekit's middleware.Fail, sees those codes rather than a generic 500.
//
//	var apiErr *ikuaiapi.APIError
//	if errors.As(err, &apiErr) {
//	    switch {
//	    case apiErr.IsAuthFailure():       // refresh the token
//	    case apiErr.IsPermissionFailure(): // needs a privileged account
//	    case apiErr.IsBadRequest():        // fix the payload
//	    }
//	    log.Println(apiErr.Message, "->", apiErr.Remedy())
//	}
//
// SDK-local validation failures use the 4xxx business-code band, so a
// caller can tell "the SDK refused to send this" from "the router said no".
//
// # Logging
//
// The SDK logs through the oekit logx facade with a lazily resolved
// package logger, so it never picks a backend or a service name on behalf
// of its consumer: call logx.SetDefault, logx.UseSlog, or install an
// hlog backend in main, and this package follows. SetLogLevel controls the
// floor for the SDK's own events.
//
// # Observability
//
// Metrics collects request counts and latency without forcing a metrics
// backend on a consumer. Retry honours Retry-After, and never retries a
// write on a transport error, because that write may already have reached
// the router.
//
// # Packages
//
//   - service — the generated typed client, one service per endpoint group.
//   - session — credential persistence and resolution.
//   - output  — table and JSON rendering for responses.
package ikuaiapi
