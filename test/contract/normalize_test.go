package contract

import (
	"net/http"
	"testing"
)

// The amended AC3 ignore list is exactly four items: Date and Content-Length are dropped,
// X-RateLimit-Reset's value is masked, generated UUIDs in bodies become <uuid>, and
// POST /token's body is opaque. These tests pin that list from both sides — what the
// normalizer changes, and what it must leave alone.

func TestNormalizeHeaders(t *testing.T) {
	t.Run("drops Date", func(t *testing.T) {
		got := normalizeHeaders(http.Header{"Date": {"Tue, 28 Jul 2026 22:25:26 GMT"}})

		if _, present := got["Date"]; present {
			t.Errorf("Date survived normalization: %v", got)
		}
	})

	t.Run("drops Content-Length", func(t *testing.T) {
		got := normalizeHeaders(http.Header{"Content-Length": {"22"}})

		if _, present := got["Content-Length"]; present {
			t.Errorf("Content-Length survived normalization: %v", got)
		}
	})

	t.Run("masks the X-RateLimit-Reset value", func(t *testing.T) {
		got := normalizeHeaders(http.Header{"X-Ratelimit-Reset": {"47"}})

		want := []string{resetValue}
		if len(got["X-Ratelimit-Reset"]) != 1 || got["X-Ratelimit-Reset"][0] != want[0] {
			t.Errorf("X-Ratelimit-Reset = %v, want %v", got["X-Ratelimit-Reset"], want)
		}
	})

	t.Run("keeps every other header exact", func(t *testing.T) {
		in := http.Header{
			"X-Ratelimit-Remaining": {"96"},
			"X-Ratelimit-Limit":     {"120"},
			"Content-Type":          {"application/json"},
			"Vary":                  {"Accept-Encoding"},
			"Content-Encoding":      {"gzip"},
			"Location":              {"https://github.com/joaofnds/astro"},
		}

		got := normalizeHeaders(in)

		for name, want := range in {
			if len(got[name]) != 1 || got[name][0] != want[0] {
				t.Errorf("%s = %v, want %v", name, got[name], want)
			}
		}
	})

	t.Run("sorts values within a header", func(t *testing.T) {
		got := normalizeHeaders(http.Header{"Vary": {"Origin", "Accept-Encoding"}})

		want := []string{"Accept-Encoding", "Origin"}
		if len(got["Vary"]) != 2 || got["Vary"][0] != want[0] || got["Vary"][1] != want[1] {
			t.Errorf("Vary = %v, want %v", got["Vary"], want)
		}
	})
}

func TestEncodeBody(t *testing.T) {
	t.Run("normalizes a generated UUID", func(t *testing.T) {
		encoding, body := encodeBody([]byte(`{"id":"9f1a6c2e-3b4d-4e5f-8a9b-0c1d2e3f4a5b"}`))

		if encoding != "utf8" {
			t.Errorf("encoding = %q, want %q", encoding, "utf8")
		}
		if want := `{"id":"<uuid>"}`; body != want {
			t.Errorf("body = %q, want %q", body, want)
		}
	})

	t.Run("keeps a timestamp exact", func(t *testing.T) {
		raw := `{"created_at":"2026-01-02T03:04:05Z","description":"contract probe"}`

		_, body := encodeBody([]byte(raw))

		if body != raw {
			t.Errorf("body = %q, want %q", body, raw)
		}
	})

	t.Run("keeps a non-UUID body byte-identical", func(t *testing.T) {
		raw := "Cannot GET /nope"

		_, body := encodeBody([]byte(raw))

		if body != raw {
			t.Errorf("body = %q, want %q", body, raw)
		}
	})

	t.Run("base64-encodes a body that is not valid UTF-8", func(t *testing.T) {
		encoding, body := encodeBody([]byte{0xff, 0xfe, 0x00})

		if encoding != "base64" {
			t.Errorf("encoding = %q, want %q", encoding, "base64")
		}
		if want := "//4A"; body != want {
			t.Errorf("body = %q, want %q", body, want)
		}
	})
}
