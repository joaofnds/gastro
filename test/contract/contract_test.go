// Package contract captures the HTTP contract of the fully-wired application as a golden
// file, so a framework migration can be diffed against it rather than reviewed by hand.
//
// The comparison covers status code, headers, and body, ignoring only:
//   - Date and Content-Length, which are omitted entirely;
//   - the value of X-RateLimit-Reset, a wall clock — its presence and name still compare;
//   - generated UUIDs in bodies, normalized to <uuid>;
//   - POST /token's body, a fresh age blob per run — its status and headers still compare.
//
// Nothing else is excluded. X-RateLimit-Remaining and Activity.CreatedAt compare exactly:
// the first is deterministic given the fixed probe sequence, and the second is populated
// from the request body rather than generated.
//
// Run with -update to rewrite the golden file. Regenerating it to make a failure go away is
// the failure mode this file exists to prevent — investigate the diff instead.
package contract

import (
	"astro/adapters/health"
	http2 "astro/adapters/http"
	"astro/adapters/logger"
	"astro/adapters/postgres"
	"astro/config"
	"astro/habit"
	"astro/test"
	"astro/test/transaction"
	"astro/token"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

var update = flag.Bool("update", false, "rewrite the golden contract file")

const goldenPath = "testdata/contract.golden.json"

func startApp(t *testing.T) string {
	t.Helper()

	var httpConfig http2.Config

	fxApp := fxtest.New(
		t,
		logger.NopLogger,
		test.NewPortAppConfig,
		habit.NopProbeProvider,
		token.NopProbeProvider,
		http2.NopProbeProvider,
		config.Module,
		health.Module,
		postgres.Module,
		http2.Module,
		habit.Module,
		token.Module,
		transaction.Module,
		LargeRoute,
		fx.Populate(&httpConfig),
	).RequireStart()

	t.Cleanup(fxApp.RequireStop)

	return fmt.Sprintf("http://localhost:%d", httpConfig.Port)
}

func TestContract(t *testing.T) {
	captured := run(t, startApp(t))

	if *update {
		writeGolden(t, captured)
		t.Logf("wrote %d entries to %s", len(captured), goldenPath)
		return
	}

	for _, difference := range diff(readGolden(t), captured) {
		t.Error(difference)
	}
}

func TestHarnessBoots(t *testing.T) {
	baseURL := startApp(t)

	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /health status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestContractLargeIsCompressed(t *testing.T) {
	baseURL := startApp(t)

	req, err := http.NewRequest(http.MethodGet, baseURL+"/contract/large", nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /contract/large: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if got := resp.Header.Get("Content-Encoding"); got != "gzip" {
		t.Errorf("Content-Encoding = %q, want %q", got, "gzip")
	}

	decompressed, err := gunzip(readAll(t, resp.Body))
	if err != nil {
		t.Fatalf("gunzipping body: %v", err)
	}
	if len(decompressed) <= 1024 {
		t.Errorf("decompressed body = %d bytes, want > 1024", len(decompressed))
	}
}

func readGolden(t *testing.T) []entry {
	t.Helper()

	raw, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("reading golden file (run with -update to capture it): %v", err)
	}

	var golden []entry
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatalf("parsing golden file: %v", err)
	}

	return golden
}

func writeGolden(t *testing.T, entries []entry) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
		t.Fatalf("creating testdata dir: %v", err)
	}

	raw, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		t.Fatalf("marshalling golden file: %v", err)
	}

	if err := os.WriteFile(goldenPath, append(raw, '\n'), 0o644); err != nil {
		t.Fatalf("writing golden file: %v", err)
	}
}

func diff(golden, captured []entry) []string {
	if len(golden) != len(captured) {
		return []string{fmt.Sprintf("probe count: golden has %d, capture has %d", len(golden), len(captured))}
	}

	var differences []string
	for i := range golden {
		differences = append(differences, diffEntry(golden[i], captured[i])...)
	}

	return differences
}

func diffEntry(want, got entry) []string {
	if want.Name != got.Name {
		return []string{fmt.Sprintf("probe order: golden has %q where the capture has %q", want.Name, got.Name)}
	}

	var differences []string
	report := func(field, w, g string) {
		if w == g || accepted(want.Name, field, w, g) {
			return
		}

		differences = append(differences, fmt.Sprintf("%s: %s: golden %q, capture %q", want.Name, field, w, g))
	}

	report("status", fmt.Sprint(want.Status), fmt.Sprint(got.Status))
	report("body_encoding", want.BodyEncoding, got.BodyEncoding)
	report("body", comparableBody(want.Body), comparableBody(got.Body))

	for _, name := range headerNames(want.Headers, got.Headers) {
		report(name, headerValue(want.Headers, name), headerValue(got.Headers, name))
	}

	return differences
}

// headerValue renders a header for comparison. Absent and empty are both "", which is what
// an acceptedDelta's From or To means when it is blank.
func headerValue(headers map[string][]string, name string) string {
	return strings.Join(headers[name], ", ")
}

func headerNames(maps ...map[string][]string) []string {
	seen := map[string]bool{}
	for _, m := range maps {
		for name := range m {
			seen[name] = true
		}
	}

	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)

	return names
}
