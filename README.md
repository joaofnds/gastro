# gastro

The Go HTTP API for [astro](https://github.com/joaofnds/astro). Data lives in one
SQLite file, with GORM, Goose migrations and the modernc Go driver.

## Local development

Use Go 1.27.2 or newer. Initialize a new development database once:

```sh
export CGO_ENABLED=0
export CONFIG_PATH="$(pwd)/config/config.yaml"
export SQLITE_PATH="$(pwd)/data/astro.db"
go run ./cmd/migrate init
go run ./cmd/astro
```

`init` refuses an existing path. Normal startup requires an initialized database,
applies pending migrations, and checks the schema and integrity before listening.
Deleting the file does not create a fresh database on the next start.

```sh
go run ./cmd/migrate up
go run ./cmd/migrate check
go test ./...
```

Tests create isolated temporary databases and apply the production migration SQL.
They need no database service or separate migration step. The full local gate is
`gitleaks detect --no-banner`, `golangci-lint run ./...`, and `CGO_ENABLED=0 go test ./...`.

## Container and deployment

```sh
docker build -t astro:local .
sh scripts/test-image.sh astro:local
```

The image includes `/astro`, `/migrate`, Litestream 0.5.17 and the
operational scripts. The image check uses disposable local Docker volumes, writes
a synthetic HTTP marker, and confirms that restoration preserves the response.
It also exercises strict startup, writer exclusion, corrupt/missing data and SIGINT
shutdown. It does not establish off-host backup durability or Fly resource sizing.

The optional Mage routes use the same Dockerfile and service-free tests:
`go -C magefiles run main.go build` and `go -C magefiles run main.go testDagger`.

[Deployment, backup and recovery](ops/README.md) describes normal releases and
offline maintenance. Tags and manual runs of the build workflow publish and verify
an image, then deploy that same image by digest to the existing Fly machine and
volume. Deployment requires exactly one machine and uses `--ha=false`.
