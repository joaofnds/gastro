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

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

// largeBody is over 1KB so the compress middleware actually engages. Neither Fiber major
// compresses a small response, which is what makes a gzip assertion on /health vacuous.
var largeBody = strings.Repeat("astro contract probe payload.\n", 40)

var LargeRoute = fx.Invoke(func(app *fiber.App) {
	app.Get("/contract/large", func(c *fiber.Ctx) error {
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
	entries := make([]entry, 0, len(probes()))

	for _, p := range probes() {
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
