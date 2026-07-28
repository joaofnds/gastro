# Contract deltas — Fiber v2.52.14 → v3.4.0

Every difference the golden capture produced when the ported tree was diffed against
`testdata/contract.golden.json`, with its disposition. Per the migration's C14 the capture's
diff **is** the delta set: two prior rounds of enumeration by inspection each declared the set
complete and each was wrong.

The raw diff was **43 differences across 31 probes, in 16 distinct kinds**. Nothing here is
accepted by omission — §4 lists what the capture provably could not see.

## 1. Suppressed — v2's behavior reproduced by configuration

| Surface | v2 | v3 default | How it is held |
| --- | --- | --- | --- |
| `GET /` status | `302` | `303` | `c.Redirect().Status(fiber.StatusFound).To(...)` (`adapters/http/http.go`) |
| Preflight `Access-Control-Allow-Methods` set | 6 methods | adds `QUERY` | explicit `cors.Config{AllowMethods: ...}` listing v2's six (`adapters/http/fiber.go`) |
| JSON parse-failure body | decoder's own message | `bind from body: ` prefix | `messageFromBindError` at `habit/controller.go:87,117,134,188` |

These three produce **no** entry in the diff. That is the evidence they hold.

## 2. Accepted — residual deltas that ship

| # | Ref | Delta | Probes | Reason |
| --- | --- | --- | --- | --- |
| 1 | C14.1 | `Vary: Accept-Encoding` added to responses the middleware declines to compress | 18 | Unsuppressible by configuration (C14a): the escapes that skip `appendVaryAcceptEncoding` — `compress.Config.Next` returning true, `Level: LevelDisabled` — also disable compression. It is also the HTTP-correct header; v2's omission was the defect. Affects cache keys only, and nothing caches here: the TUI is the sole client and Fly puts no CDN in front of `gastro`. |
| 2 | C14.1 | Preflight `Vary` gains a 4th value, `Accept-Encoding` | 1 | The same delta as #1 observed on a different request. The cors middleware sets the same 3 values on both majors; the 4th is compress's. Not a CORS change. |
| 3 | C14.4 | `Content-Type: application/json` → `application/json; charset=utf-8` | 9 | `app/api/client.go:86-90` unmarshals the body without inspecting `Content-Type`. The parameter is also the HTTP-correct form. |
| 4 | C14.2 | `Access-Control-Allow-Methods` separator `,` → `, ` | 1 | `strings.Join(cfg.AllowMethods, ", ")` is hard-coded at `middleware/cors/cors.go:192` and `AllowMethods` feeds nothing else, so the separator is not configurable. RFC 7230 treats the two as equivalent, and there is no browser client. |
| 5 | C14.8 | `HEAD` of a compressible route loses `Content-Encoding: gzip` | 1 | v3's `shouldSkip` returns early for HEAD. A HEAD response carries no body, so no client can act on the header. |
| 6 | C14.7 | `PROPFIND /health`: `400 Invalid http method` → `501 Not Implemented` | 1 | v3 short-circuits genuinely unknown methods before the middleware chain. `501` is the correct status for an unimplemented method; `400` was not. Unreachable from the TUI, which issues only GET/POST/PATCH/DELETE. |
| 7 | C14.7 | `QUERY /health`: `400 Invalid http method` → `405 Method Not Allowed`, gaining `Allow: GET, HEAD` **and** the three `X-RateLimit-*` headers | 1 | v3 adds `MethodQuery` to `DefaultMethods`, so QUERY is routed rather than rejected and reaches the limiter. `405` + `Allow` is the correct response for a known method with no handler. Unreachable from the TUI. |

### Correction to C14.7

C14.7 states that `PROPFIND` and `QUERY` both **lose** `X-RateLimit-*` on v3. The capture shows
the opposite for `QUERY`:

- **v2** — neither carries `X-RateLimit-*`; both are rejected at the router with `400`.
- **v3** — `PROPFIND` still carries none (short-circuited); `QUERY` **gains** all three, because
  v3 routes it through the middleware chain before the router answers `405`.

The direction was backwards in the baseline. The disposition is unchanged: accepted.

This is also why the probe list puts `PROPFIND` and `QUERY` last. `QUERY` now consumes a limiter
slot on v3 and not on v2, so any probe after it would carry an `X-RateLimit-Remaining` offset —
one real delta cascading into false ones.

## 3. Open entries, decided here

### C14.5 — 404 body `Cannot GET /nope` → `Not Found` — **ACCEPTED**

C14 marked this open on the stated ground that "the TUI surfaces this verbatim in
`api.Error.Body`". **That ground is false**, and the correction is what decides the entry:

- `app/api/error.go:14-19` — `Error.Error()` formats `Method`, `Path`, `StatusCode` and
  `Message`. It never reads `Body`.
- `Body` is written at `app/api/client.go:81` and `app/api/token.go:35,68`, and read nowhere.
- `Message` is set nowhere outside `app/api/error_test.go`, so `Error()` always takes the short
  branch and never renders a body at all.
- Every path the TUI builds is a fixed literal plus a UUID or a query-escaped name
  (`app/api/habits.go:21-66`, `groups.go:18-44`, `token.go:14,47`). It cannot construct an
  unregistered path.

The body is dead data in the only client, on a path that client cannot reach. `Not Found` is
also the standard reason phrase, where `Cannot GET /x` leaked the request line back to the
caller.

### C14.6 — `GET /habitsfoo`, `GET /groupsxyz` unauthenticated: `401` → `404` — **ACCEPTED**

Accepted by the owner on 2026-07-29, after both facts below were surfaced. Not waved through:
C17 reserved this call for them.

Both facts, surfaced rather than waved through:

**It is a status change on a path the `/habits` group used to own.** A literal reading of C17
("stop if the port requires a change to a route... or response shape") makes this an abort
condition, and the grilled doc's Invariants §5 left it open for exactly that reason.

**It is a v3 correctness fix, and unreachable from the client.** `path.go:554-562` — v3's
`getMatch` calls `hasPartialMatchBoundary` on prefix routes, so a `Use`/`Group` prefix match must
end at a segment boundary. v2 had no such check, so `/habitsfoo` fell into the `/habits` group
and ran `middlewareDecodeToken`, which answered `401 missing Authorization token`. v3 does not
match the group, and the router answers `404`. The v2 behavior meant a family of paths the
service defines no route for nonetheless reported "you need a token" — advertising that a
`/habits` prefix exists. `404` leaks less.

The TUI never constructs such a path (see the C14.5 evidence above).

**Rationale for accepting:** `/habitsfoo` was never a declared route — it was accidental prefix
capture — so no route the service defines changed shape.

## 4. Known blind spots — what the capture provably cannot see

Recorded because C14's inversion is only sound with this list written down. "Uncaptured" must
be a category with members.

| Blind spot | Why | Disposition |
| --- | --- | --- |
| **C14.3 — `Content-Type: text/json` returns 400 on v3 where v2 returned 201** | **No probe sends `text/json`.** The delta is absent from the diff because the surface is uncovered, not because it went away | Accepted per C14.3 — unreachable from the TUI, which sets `application/json` on every bodied request (`app/api/client.go:62-64`). Reproducing v2 would mean re-implementing content-type sniffing. The `messageFromBindError` fallback branch is unit-tested at `habit/bind_error_test.go` precisely because no probe reaches it |
| C14.9 — sliding-window limiter rounds up where v2 truncated | Surfaces only when a client is actually 429'd; no single-shot capture can see it | Accepted. `X-RateLimit-Remaining` is compared **exactly** across every probe that carries it and showed no difference, which is the best available proxy |
| `X-RateLimit-Reset` value | A wall clock; differs between the capture run and every assert run | Accepted; presence and name still compare exactly |
| Generated UUIDs in bodies | Per-run | Accepted; the body structure around them still compares |
| `POST /token` body | A fresh age blob per run | Accepted; status and headers still compare |
| Compression **in production** | `/contract/large` is registered by the test harness and returns 404 in production | Accepted. C14.1 and C14.8 are verified locally only |
| `Content-Encoding` on probes that do not set `Accept-Encoding` | Go's transport negotiates gzip itself and transparently decompresses, stripping the header | Not a defect: deterministic, and applied identically to v2 and v3. Probes 25/26 set the header explicitly and do observe it |
| Any surface no probe touches | Residual, by definition | The realised risk of this project. The probe list's coverage is the mitigation |
