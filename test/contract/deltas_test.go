package contract

import "testing"

// These constraints are what stops the exception list from degenerating into a blanket
// ignore. The golden file cannot see them: exceptions live in the test code, so a `git diff`
// on testdata/ stays clean no matter how broad an entry becomes.

// C14 records only two deltas as affecting every response: Vary: Accept-Encoding (C14.1) and
// the JSON Content-Type parameter (C14.4). Every other entry must name the probe it applies to.
var refsAllowedEveryProbe = map[string]bool{"C14.1": true, "C14.4": true}

func TestAcceptedDeltas(t *testing.T) {
	t.Run("every entry cites a C14 reference", func(t *testing.T) {
		for _, d := range acceptedDeltas {
			if d.Ref == "" {
				t.Errorf("%s / %s: empty Ref", d.Scope, d.Field)
			}
		}
	})

	t.Run("every entry gives a reason", func(t *testing.T) {
		for _, d := range acceptedDeltas {
			if d.Reason == "" {
				t.Errorf("%s / %s: empty Reason", d.Scope, d.Field)
			}
		}
	})

	t.Run("every entry targets a named field", func(t *testing.T) {
		for _, d := range acceptedDeltas {
			if d.Field == "" {
				t.Errorf("%s: empty Field would match a whole header map or body", d.Scope)
			}
		}
	})

	t.Run("only C14's every-response deltas are scoped to every probe", func(t *testing.T) {
		for _, d := range acceptedDeltas {
			if d.Scope == scopeEveryProbe && !refsAllowedEveryProbe[d.Ref] {
				t.Errorf("%s scoped to every probe, but C14 does not record it as affecting every response", d.Ref)
			}
		}
	})

	t.Run("every entry names a probe that exists", func(t *testing.T) {
		names := map[string]bool{scopeEveryProbe: true}
		for _, p := range probes() {
			names[p.name] = true
		}

		for _, d := range acceptedDeltas {
			if !names[d.Scope] {
				t.Errorf("%q is not a probe name", d.Scope)
			}
		}
	})

	t.Run("no entry passes a transition it was not written for", func(t *testing.T) {
		if accepted("GET /health", "status", "200", "500") {
			t.Error("an unlisted status transition was accepted")
		}
		if accepted("GET /health", "Vary", "", "Accept-Encoding, Origin") {
			t.Error("a Vary value the list does not name was accepted")
		}
	})
}
