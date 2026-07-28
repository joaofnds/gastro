# AGENTS.md

The HTTP API behind [astro](https://github.com/joaofnds/astro). Go module `astro`, go 1.25,
deployed to Fly as `gastro`.

**This service is a downstream of the Go service template and has fallen behind it.** The
template's own `AGENTS.md` is the house Go stack — read it first when the template is checked
out locally (`~/code/go/template/AGENTS.md` on the author's machine). If it isn't available,
this file plus `~/.agents/rules/coding_style_go.md` cover everything except a Harness rewrite.

This file records only what astro does *differently*, and most of those differences are debt
rather than choices. Last substantive commit: 2026-03-30; the template moved on after that.

## Stack

Shares the template's stack — fx, Fiber, GORM over Postgres, Viper, Zap, Prometheus,
Ginkgo/Gomega, Goose — minus everything the template added since: no Watermill, Asynq,
Casbin/Casdoor, OpenTelemetry, Mongo, Redis, feature flags, or `validator/v10`.

Fly runs `release_command = "/migrate up"` before the new machines start, with
`auto_rollback = true`. A migration therefore runs against production while the previous
version is still serving, and a rollback reverts the code but not the schema.

## Divergences from the template

| | template | astro |
| --- | --- | --- |
| Infrastructure dir | `adapter/` | `adapters/` |
| Domain layout | package family — `user/` with nested `user_adapter/`, `user_http/`, `user_module/` | flat — `habit/habit_service.go`, `controller.go`, `module.go` in one package |
| Repository file | `postgres_repository.go` in `_adapter` | `<entity>_sql_repository.go` in the domain package |
| Module visibility | `fx.Private` for anything not consumed outside the module | everything public in the graph |
| Observability | `PromProbe` subscribes to domain events over Watermill | `Probe` interface called directly by the service |
| Test bootstrap | `test/harness` | eleven hand-wired `fxtest.New` sites |
| Doubles | hand-written real implementations | `token/mock.go`, mockgen output over owned ports |
| Receiver naming | named for the type — `repository`, `controller`, `driver` | mostly single letters — `a`, `d`, `c`, `p` are 128 of 171 receivers |
| Build | Task (`Taskfile.yaml`) | Mage (`magefiles/`) |
| Lint config | `.golangci.yaml` | none, and no format target — `magefiles/pipe.go:28` runs `golangci-lint run` only |

Prefer the template's shape for new code here. Converting existing code is a migration, not a
drive-by: do it when the task already touches that area, and say so. **When a row is closed,
delete it** — a row that no longer describes astro is worse than no row.

## Layout as it stands

Domain packages at the root (`habit/`, `token/`), infrastructure under `adapters/`
(`http`, `postgres`, `logger`, `metrics`, `health`), configuration in `config/`, entry
points in `cmd/astro` and `cmd/migrate`, shared test utilities in `test/`.

Within a domain package: `<entity>.go` for the concept, `<entity>_service.go`,
`<entity>_sql_repository.go`, `controller.go`, `module.go`, `config.go`, `error.go`,
`prom_probe.go` / `nop_probe.go`. `habit/` also has `sql_repository.go`, which holds the
shared `resultErr` and `translateError` helpers — astro's equivalent of the template's
`gormErr`.

## Testing — the target and the debt

The target is the template's Harness: `~/.agents/rules/testing/01-architecture-and-harness.md`
§5, implemented at `~/code/go/template/test/harness/`.

**astro does not have it.** Eleven hand-wired bootstraps across ten test files
(`token/age_encrypter_test.go`, `token/service_test.go`, `token/controller_test.go`,
`habit/controller_test.go`, `habit/activity_service_test.go`, `habit/group_service_test.go`,
`habit/habit_service_test.go`, `adapters/metrics/metrics_test.go`,
`adapters/health/controller_test.go` ×2, `adapters/http/fiber_test.go`) list modules in
`BeforeEach`. Four of those sites also register their own routes with `fx.Invoke` —
`habit/controller_test.go:52`, `token/controller_test.go:40`,
`adapters/health/controller_test.go:58` and `:94`. A test that has to register routes is
telling you the production module doesn't; fix the module, not the test.

Those bootstraps predate the rule and are known debt — do not copy the pattern into a new
test, and do not read them as sanctioning it.

**What to copy, and what not to.** Copy the Harness's shape: a `Setup` returning a driver,
overrides as functional options, isolation owned by the harness rather than the test. Do
*not* copy how it holds its module list — the template keeps a second copy by hand in
`defaultOptions`, parallel to `cmd/app/app.go`, and the two have already drifted. astro's
`cmd/astro/astro.go` has the same inline list, so building the Harness here is a chance to
extract one shared `fx.Options` that both the entry point and the Harness consume. That is
the part the template still owes.

What astro already has to build on: `test/driver` (a `Driver` plus a low-level `API`),
`test/req`, `test/matchers` (`Must2[T]`), `test/transaction` (rollback via `fx.Decorate`),
and `test/new_port_app_config.go` (dynamic ports for parallel suites). Those are the
Harness's parts, unassembled.

`token/mock.go` is mockgen output for four owned `token` ports — `IDGenerator`, `Encrypter`,
`Encoder`, `Instrumentation` — all consumed by `token/service_test.go`. Mocking owned
interfaces is the case `~/.agents/rules/testing/02-mocking-roles.md` bans. It stays until that
suite is rewritten.

## Boundaries

- **Do not rewrite the eleven bootstraps unasked.** Migrating the suite to a Harness is its
  own piece of work with its own review, not a side effect of a feature task.
- **Do not regenerate `token/mock.go` naively.** The file declares `package token` while its
  generated header says `mock_token`; it was hand-edited after generation, and a plain
  `mockgen` run restores the header's package and breaks the build. Do not add another mock.
- **Do not edit other generated files.** Re-run the generator.
