package contract

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"
)

// largeBody is over 1KB so the compress middleware actually engages. Neither Fiber major
// compresses a small response, which is what makes a gzip assertion on /health vacuous.
var largeBody = strings.Repeat("astro contract probe payload.\n", 40)

var LargeRoute = fx.Invoke(func(app *fiber.App) {
	app.Get("/contract/large", func(c fiber.Ctx) error {
		return c.SendString(largeBody)
	})
})

// probeClient must not follow redirects: probe 22 records the configured 302 that the port
// preserves, and a following client would record GitHub's 200 instead.
var probeClient = &http.Client{
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// state threads the identifiers each probe creates into the probes that use them.
type state struct {
	token      string
	habitID    string
	activityID string
	groupID    string
}

type probe struct {
	name    string
	method  string
	path    string
	auth    bool
	headers map[string]string
	body    string
	// opaque records status and headers only. POST /token returns a fresh age blob per run.
	opaque  bool
	capture func(*state, []byte)
}

// resolvePath substitutes the identifiers created earlier in the sequence. The recorded
// entry keeps the unresolved form, so the golden file is stable across runs.
func (p probe) resolvePath(s *state) string {
	r := strings.NewReplacer("{h}", s.habitID, "{a}", s.activityID, "{g}", s.groupID)
	return r.Replace(p.path)
}

type entry struct {
	Name         string              `json:"name"`
	Method       string              `json:"method"`
	Path         string              `json:"path"`
	Status       int                 `json:"status"`
	Headers      map[string][]string `json:"headers"`
	BodyEncoding string              `json:"body_encoding"`
	Body         string              `json:"body"`
}

const jsonContentType = "application/json"

var jsonHeaders = map[string]string{"Content-Type": jsonContentType}

// fixedActivityDate round-trips deterministically: Activity.CreatedAt is populated from the
// request body rather than generated, so it compares exactly (plan §4.4).
const fixedActivityDate = "2026-01-02T03:04:05Z"

func probes() []probe {
	return []probe{
		{
			name: "POST /token", method: http.MethodPost, path: "/token", opaque: true,
			capture: func(s *state, b []byte) { s.token = string(b) },
		},
		{name: "GET /tokentest", method: http.MethodGet, path: "/tokentest", auth: true},
		{name: "GET /habits (empty)", method: http.MethodGet, path: "/habits", auth: true},
		{
			name: "POST /habits?name=read", method: http.MethodPost, path: "/habits?name=read", auth: true,
			capture: func(s *state, b []byte) { s.habitID = jsonField(b, "id") },
		},
		{name: "GET /habits/{h}", method: http.MethodGet, path: "/habits/{h}", auth: true},
		{
			name: "PATCH /habits/{h}", method: http.MethodPatch, path: "/habits/{h}", auth: true,
			headers: jsonHeaders, body: `{"name":"read more"}`,
		},
		{
			name: "POST /habits/{h} (add activity)", method: http.MethodPost, path: "/habits/{h}", auth: true,
			headers: jsonHeaders,
			body:    `{"description":"contract probe","date":"` + fixedActivityDate + `"}`,
		},
		{
			name: "GET /habits/{h} (nested activities)", method: http.MethodGet, path: "/habits/{h}", auth: true,
			capture: func(s *state, b []byte) { s.activityID = firstActivityID(b) },
		},
		{name: "GET /habits (non-empty)", method: http.MethodGet, path: "/habits", auth: true},
		{
			name: "PATCH /habits/{h}/{a}", method: http.MethodPatch, path: "/habits/{h}/{a}", auth: true,
			headers: jsonHeaders, body: `{"description":"edited"}`,
		},
		{name: "GET /groups (groups empty)", method: http.MethodGet, path: "/groups", auth: true},
		{
			name: "POST /groups", method: http.MethodPost, path: "/groups", auth: true,
			headers: jsonHeaders, body: `{"name":"morning"}`,
			capture: func(s *state, b []byte) { s.groupID = jsonField(b, "id") },
		},
		{name: "POST /groups/{g}/{h}", method: http.MethodPost, path: "/groups/{g}/{h}", auth: true},
		{name: "GET /groups (non-empty)", method: http.MethodGet, path: "/groups", auth: true},
		{name: "DELETE /groups/{g}/{h}", method: http.MethodDelete, path: "/groups/{g}/{h}", auth: true},
		{name: "DELETE /groups/{g}", method: http.MethodDelete, path: "/groups/{g}", auth: true},
		{name: "DELETE /habits/{h}/{a}", method: http.MethodDelete, path: "/habits/{h}/{a}", auth: true},

		// The three JSON error paths (AC5) run while {h} still exists. Deleting it first
		// turns all three into a 404 from middlewareFindHabit and voids the coverage.
		{
			name: "PATCH /habits/{h} (malformed JSON)", method: http.MethodPatch, path: "/habits/{h}", auth: true,
			headers: jsonHeaders, body: `{`,
		},
		{
			name: "PATCH /habits/{h} (type mismatch)", method: http.MethodPatch, path: "/habits/{h}", auth: true,
			headers: jsonHeaders, body: `{"name":123}`,
		},
		{
			name: "PATCH /habits/{h} (empty body)", method: http.MethodPatch, path: "/habits/{h}", auth: true,
			headers: jsonHeaders, body: ``,
		},
		{name: "DELETE /habits/{h}", method: http.MethodDelete, path: "/habits/{h}", auth: true},

		{name: "GET / (redirect)", method: http.MethodGet, path: "/"},
		{name: "GET /health", method: http.MethodGet, path: "/health"},
		{
			name: "OPTIONS /habits (CORS preflight)", method: http.MethodOptions, path: "/habits",
			headers: map[string]string{
				"Origin":                        "https://example.com",
				"Access-Control-Request-Method": http.MethodPost,
			},
		},
		{
			name: "GET /contract/large (gzip)", method: http.MethodGet, path: "/contract/large",
			headers: map[string]string{"Accept-Encoding": "gzip"},
		},
		{
			name: "HEAD /contract/large (gzip)", method: http.MethodHead, path: "/contract/large",
			headers: map[string]string{"Accept-Encoding": "gzip"},
		},
		{name: "GET /nope (404)", method: http.MethodGet, path: "/nope"},
		{name: "GET /habitsfoo (unauthenticated)", method: http.MethodGet, path: "/habitsfoo"},
		{name: "GET /groupsxyz (unauthenticated)", method: http.MethodGet, path: "/groupsxyz"},

		// PROPFIND and QUERY run last. v3 short-circuits unknown methods before the middleware
		// chain (C14.7), so they never reach the limiter there. Anything after them would have
		// every X-RateLimit-Remaining off by two, cascading one real delta into false ones.
		{name: "PROPFIND /health", method: "PROPFIND", path: "/health"},
		{name: "QUERY /health", method: "QUERY", path: "/health"},
	}
}

func run(t *testing.T, baseURL string) []entry {
	t.Helper()

	s := &state{}
	sequence := probes()
	entries := make([]entry, 0, len(sequence))

	for _, p := range sequence {
		entries = append(entries, runProbe(t, baseURL, p, s))
	}

	return entries
}

func runProbe(t *testing.T, baseURL string, p probe, s *state) entry {
	t.Helper()

	var body io.Reader
	if p.body != "" {
		body = strings.NewReader(p.body)
	}

	req, err := http.NewRequest(p.method, baseURL+p.resolvePath(s), body)
	if err != nil {
		t.Fatalf("%s: building request: %v", p.name, err)
	}
	for k, v := range p.headers {
		req.Header.Set(k, v)
	}
	if p.auth {
		req.Header.Set("Authorization", s.token)
	}

	resp, err := probeClient.Do(req)
	if err != nil {
		t.Fatalf("%s: %v", p.name, err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw := readAll(t, resp.Body)
	if resp.Header.Get("Content-Encoding") == "gzip" && len(raw) > 0 {
		decompressed, gzErr := gunzip(raw)
		if gzErr != nil {
			t.Fatalf("%s: gunzipping body: %v", p.name, gzErr)
		}
		raw = decompressed
	}

	if p.capture != nil {
		p.capture(s, raw)
	}

	return newEntry(p, resp, raw)
}

func newEntry(p probe, resp *http.Response, raw []byte) entry {
	e := entry{
		Name:    p.name,
		Method:  p.method,
		Path:    p.path,
		Status:  resp.StatusCode,
		Headers: normalizeHeaders(resp.Header),
	}

	if p.opaque {
		e.BodyEncoding = "opaque"
		return e
	}

	e.BodyEncoding, e.Body = encodeBody(raw)

	return e
}

// acceptedDelta records one contract difference that ships, so the comparator can pass it
// without the golden file being regenerated — regeneration is the failure mode a golden file
// exists to prevent. Every entry names the C14 reference that authorised it; DELTAS.md carries
// the reasoning. An entry targets exactly one field of one probe, so it can never degenerate
// into a blanket ignore.
type acceptedDelta struct {
	// Scope is a probe name, or "*" for every probe. "*" is permitted only where C14 records
	// the delta as affecting every response.
	Scope string
	// Field is "status", "body", or a canonical header name.
	Field string
	// From is the golden value, To the observed one. "" means the header is absent.
	From string
	To   string
	// Ref is the C14 entry authorising this delta. Required.
	Ref    string
	Reason string
}

const scopeEveryProbe = "*"

var acceptedDeltas = []acceptedDelta{
	{
		Scope: scopeEveryProbe, Field: "Vary", From: "", To: "Accept-Encoding", Ref: "C14.1",
		Reason: "v3's compress middleware sets Vary even when it declines to compress. Not suppressible by config (C14a); affects cache keys only and nothing caches here.",
	},
	{
		Scope: scopeEveryProbe, Field: "Content-Type",
		From: "application/json", To: "application/json; charset=utf-8", Ref: "C14.4",
		Reason: "app/api/client.go:86-90 unmarshals without inspecting Content-Type.",
	},
	{
		Scope: "OPTIONS /habits (CORS preflight)", Field: "Vary",
		From: "Access-Control-Request-Method, Access-Control-Request-Headers, Origin",
		To:   "Access-Control-Request-Method, Access-Control-Request-Headers, Origin, Accept-Encoding",
		Ref:  "C14.1",
		Reason: "The same delta as the Vary entry above, observed on the preflight. The cors " +
			"middleware sets the same 3 values on both majors; the 4th is compress's.",
	},
	{
		Scope: "OPTIONS /habits (CORS preflight)", Field: "Access-Control-Allow-Methods",
		From: "GET,POST,HEAD,PUT,DELETE,PATCH", To: "GET, POST, HEAD, PUT, DELETE, PATCH", Ref: "C14.2",
		Reason: "strings.Join(cfg.AllowMethods, \", \") is hard-coded at middleware/cors/cors.go:192. RFC 7230 treats the two as equivalent; there is no browser client.",
	},
	{
		Scope: "HEAD /contract/large (gzip)", Field: "Content-Encoding", From: "gzip", To: "", Ref: "C14.8",
		Reason: "v3's shouldSkip returns early for HEAD. A HEAD response carries no body, so no client can act on the header.",
	},
	{
		Scope: "PROPFIND /health", Field: "status", From: "400", To: "501", Ref: "C14.7",
		Reason: "v3 short-circuits unknown methods before the middleware chain. 501 is correct for an unimplemented method; 400 was not.",
	},
	{
		Scope: "PROPFIND /health", Field: "body", From: "Invalid http method", To: "Not Implemented", Ref: "C14.7",
		Reason: "Reason phrase for the corrected status.",
	},
	{
		Scope: "QUERY /health", Field: "status", From: "400", To: "405", Ref: "C14.7",
		Reason: "v3 adds MethodQuery to DefaultMethods, so QUERY is routed rather than rejected and the router answers 405.",
	},
	{
		Scope: "QUERY /health", Field: "body", From: "Invalid http method", To: "Method Not Allowed", Ref: "C14.7",
		Reason: "Reason phrase for the corrected status.",
	},
	{
		Scope: "QUERY /health", Field: "Allow", From: "", To: "GET, HEAD", Ref: "C14.7",
		Reason: "405 requires Allow. v2 never reached the router for this method.",
	},
	{
		Scope: "QUERY /health", Field: "X-Ratelimit-Limit", From: "", To: "120", Ref: "C14.7",
		Reason: "QUERY now reaches the limiter on v3. C14.7 states both unknown methods lose X-RateLimit-*; the capture shows QUERY gains them. See DELTAS.md.",
	},
	{
		Scope: "QUERY /health", Field: "X-Ratelimit-Remaining", From: "", To: "90", Ref: "C14.7",
		Reason: "As above. Exact because QUERY runs last, so no probe follows it to inherit an offset.",
	},
	{
		Scope: "QUERY /health", Field: "X-Ratelimit-Reset", From: "", To: resetValue, Ref: "C14.7",
		Reason: "As above.",
	},
	{
		Scope: "GET /nope (404)", Field: "body", From: "Cannot GET /nope", To: "Not Found", Ref: "C14.5",
		Reason: "api.Error.Body is written and never read, and the TUI builds every path from a fixed set. Not Found is also the standard reason phrase.",
	},
	{
		Scope: "GET /habitsfoo (unauthenticated)", Field: "status", From: "401", To: "404", Ref: "C14.6",
		Reason: "v3 requires a prefix match to end at a segment boundary (path.go:554-562), so /habitsfoo no longer falls into the /habits group. Owner-accepted 2026-07-29.",
	},
	{
		Scope: "GET /habitsfoo (unauthenticated)", Field: "body",
		From: "missing Authorization token", To: "Not Found", Ref: "C14.6",
		Reason: "As above. v2 advertised that a /habits prefix existed; 404 leaks less.",
	},
	{
		Scope: "GET /groupsxyz (unauthenticated)", Field: "status", From: "401", To: "404", Ref: "C14.6",
		Reason: "As above, for the /groups group.",
	},
	{
		Scope: "GET /groupsxyz (unauthenticated)", Field: "body",
		From: "missing Authorization token", To: "Not Found", Ref: "C14.6",
		Reason: "As above, for the /groups group.",
	},
}

// accepted reports whether some entry authorises this exact golden-to-capture transition on
// this exact probe and field.
func accepted(probeName, field, golden, captured string) bool {
	for _, d := range acceptedDeltas {
		if d.Scope != scopeEveryProbe && d.Scope != probeName {
			continue
		}

		if d.Field == field && d.From == golden && d.To == captured {
			return true
		}
	}

	return false
}

// resetValue stands in for X-RateLimit-Reset, a wall clock that counts down within the
// limiter window and so differs between the capture run and every assert run. Its presence
// and name still compare exactly. X-RateLimit-Remaining is deliberately NOT masked: it is
// deterministic given the fixed probe sequence, and it is the only header through which the
// sliding-window rounding change (C14.9) could ever surface.
const resetValue = "<reset>"

var (
	droppedHeaders = map[string]bool{"Date": true, "Content-Length": true}
	maskedHeaders  = map[string]string{"X-Ratelimit-Reset": resetValue}
	uuidPattern    = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
)

func normalizeHeaders(h http.Header) map[string][]string {
	out := make(map[string][]string, len(h))

	for name, values := range h {
		if droppedHeaders[name] {
			continue
		}

		if mask, ok := maskedHeaders[name]; ok {
			out[name] = []string{mask}
			continue
		}

		sorted := append([]string(nil), values...)
		sort.Strings(sorted)
		out[name] = sorted
	}

	return out
}

// parseErrorPattern matches sonic's rendering of a decode failure, which the controllers put
// straight into the response body.
//
// That rendering is architecture-dependent: sonic has an amd64 JIT path and a fallback for
// everything else, and they disagree on both the message and the ParsingError code. For the
// input `{`, sonic v1.15.2 returns "expect a json key" on arm64 and "eof" on amd64; for an
// empty body, Code:0x0 Msg:"invalid chars" versus Code:0x1 Msg:"". Same version, same input —
// verified by calling sonic.Unmarshal directly under both GOARCH values, with no Fiber
// involved. So this substring cannot be golden under any probe design that has to pass on a
// developer's machine and in CI both.
//
// It is normalized at comparison time rather than at capture, so the committed golden file
// keeps the bytes it recorded from the v2 tree.
//
// What survives normalization is what the migration actually changed: v3 wraps decode failures
// in *BindError, whose Error() prepends "bind from body: ". That prefix sits before the match
// and is still compared exactly, so the unwrap at habit/controller.go stays under test.
var parseErrorPattern = regexp.MustCompile(`(?s)"?(Syntax error|Mismatch type).*`)

const parseErrorPlaceholder = "<parse-error>"

// comparableBody removes the parts of a body that no capture can hold stable. Applied to the
// golden entry and the observed one alike.
func comparableBody(body string) string {
	return parseErrorPattern.ReplaceAllString(body, parseErrorPlaceholder)
}

// encodeBody keeps the golden file reviewable: valid UTF-8 is stored verbatim so a human can
// read the T11 diff, and only genuinely binary bodies fall back to base64.
func encodeBody(raw []byte) (string, string) {
	if !utf8.Valid(raw) {
		return "base64", base64.StdEncoding.EncodeToString(raw)
	}

	return "utf8", uuidPattern.ReplaceAllString(string(raw), "<uuid>")
}

func jsonField(b []byte, field string) string {
	m := regexp.MustCompile(`"` + field + `"\s*:\s*"([^"]*)"`).FindSubmatch(b)
	if m == nil {
		return ""
	}

	return string(m[1])
}

func firstActivityID(b []byte) string {
	_, activities, found := strings.Cut(string(b), `"activities":`)
	if !found {
		return ""
	}

	return jsonField([]byte(activities), "id")
}

func readAll(t *testing.T, r io.Reader) []byte {
	t.Helper()

	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}

	return b
}

func gunzip(b []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()

	return io.ReadAll(r)
}
