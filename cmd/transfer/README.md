# Database transfer

Build the tool with `go build -o transfer ./cmd/transfer`. It reads the Postgres DSN
from `ASTRO_TRANSFER_POSTGRES_DSN`. Use a secret manager or an existing protected
environment variable to supply it. Output contains row counts and SHA-256 hashes;
connection errors and customer rows are omitted.

Stop all writers before either transfer, including API machines, maintenance jobs,
and manual sessions. `--source-frozen` records that operator acknowledgment; the
tool cannot stop writers itself. Verification compares snapshots and is useful for
cutover only while both stores remain frozen. Keep the original store available
until activation and recovery checks pass.

## Postgres to SQLite

```sh
./transfer --direction postgres-to-sqlite \
  --postgres-schema public \
  --output /data/astro-cutover.db \
  --source-frozen --timeout 5m

./transfer --direction verify \
  --postgres-schema public \
  --sqlite-path /data/astro-cutover.db --timeout 5m
```

The source uses a repeatable-read, read-only Postgres transaction. The tool copies
all fields from `habits`, `groups`, `activities`, and `groups_habits` into a new
staged SQLite database. It checks schema, foreign keys, full SQLite integrity,
canonical UUIDs, microsecond timestamp precision, and ordered field hashes before
publishing the database. It checkpoints and closes the WAL before publishing the
main file. File creation uses a hard link in the destination filesystem so an
existing file, including a symlink, cannot be replaced. The destination directory
must already exist on a filesystem that supports hard links and directory fsync.

Choose a new output path for each attempt. A failure before publication removes
the tool's staging directory. A crash can leave a hidden `.astro-transfer-*`
directory; inspect and remove it only after confirming no transfer process owns it.
If publication succeeded but the process stopped before reporting success, rerun
`verify` against that output. A repeated transfer refuses to replace it.

Activation is a separate operator action: configure the SQLite service to use the
verified path, start it, and check reads and writes. The transfer tool does not
start a service, alter routing, or delete Postgres data.

## SQLite to an isolated Postgres return schema

Freeze SQLite writes before returning. Supply a DSN for the intended Postgres
database and confirm its exact database name. The schema must be a new name
beginning with `astro_return_`; the tool refuses any existing schema. It can create
that isolated schema alongside the frozen `public` schema without changing it.

```sh
./transfer --direction sqlite-to-postgres \
  --sqlite-path /data/astro.db \
  --target-database astro \
  --target-schema astro_return_20261018 \
  --source-frozen --timeout 5m

./transfer --direction verify \
  --postgres-schema astro_return_20261018 \
  --sqlite-path /data/astro.db --timeout 5m
```

Schema creation, table creation, copying, and field verification occur in one
serializable Postgres transaction. Failure rolls that transaction back. The tool
performs no deletes, truncation, upserts, or replacement of existing schemas.
SQLite values with precision finer than a microsecond, invalid UUIDs, or unsupported
timestamps are refused. Other values Postgres cannot represent, such as embedded
NUL characters, cause a rolled-back return transfer.

To activate the previous Postgres application, configure its connection's
`search_path=astro_return_20261018,public`, using the exact verified schema name.
For a URL DSN this is the `search_path` query parameter; for a lib/pq keyword DSN
it is a `search_path` option. Put the return schema first. The copied tables retain
UUID defaults supported by Postgres 14. If the application connects as a different
role from the transfer operator, grant that role schema usage and the table access
it needs before activation. Recheck the effective search path through that role.
Starting the old application against unchanged `public` would omit writes made
after the SQLite cutover, so routing and connection configuration must be changed
explicitly after verification. The tool does not activate that application.

## Checks

Default tests require only SQLite:

```sh
go test -count=1 ./internal/transfer ./cmd/transfer
```

The separate integration job requires a disposable Postgres 14.4 service with
schema creation privileges. Set `ASTRO_TRANSFER_TEST_POSTGRES_DSN` to that service,
then run:

```sh
go test -count=1 -tags=transfer_integration ./internal/transfer ./cmd/transfer
```

The tagged suite fails if the DSN is absent. It creates and removes unique schemas,
exercises a complete Postgres to SQLite to Postgres round trip, detects changed
fields, preserves existing outputs, and observes transactional rollback after a
Postgres text-representation error. It must use a disposable database.
