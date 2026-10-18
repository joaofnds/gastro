# SQLite operations

The [2026-10-18 cutover record](2026-10-18-cutover.md) lists the production
resources, data verification and current operating limits.

## Runtime contract

`fly.toml` describes one shared-vCPU machine with 256 MiB in `ams`, mounting
`astro_data` at `/data`. The application database is `/data/astro.db`. Keep exactly
one application machine and one replicator for its backup prefix. Separate volumes
have separate lock files; a local file lock cannot prevent two machines from
replicating divergent databases to the same destination.

The image runs as UID/GID 65532. During commissioning, make the mounted data
directory and transferred database files writable by that identity. An offline
root maintenance command can set ownership before running the transfer; the API
and normal maintenance commands run without root privileges.

The container requires `SQLITE_PATH` and `LITESTREAM_REPLICA_URL`.
`SQLITE_PATH` defaults to `/data/astro.db`. The optional
`BACKUP_HEALTHCHECK_URL` enables the API's backup checker and must be an HTTPS
Healthchecks success-ping URL. Store it as a secret. The checker is disabled when
the URL is empty. The replica URL identifies the bucket and prefix without
credentials. Litestream reads storage credentials from
its supported environment variables. For S3-compatible storage, put the custom
endpoint and region in the replica URL as documented by the provider. Use secret
manager injection or `fly secrets import` through standard input; keep credentials
out of command arguments, committed files and saved shell history.

Startup takes a nonblocking file lock, runs `/migrate up` and `/migrate check`, then
executes `litestream replicate -exec /astro`. Missing, zero-length, corrupt and
uncommissioned databases fail startup. Neither the API nor the entrypoint restores
or initializes an absent database. `recover` uses the same file lock and fails with
exit code 75 while the writer holds it. Ordinary SQLite writes also retain the
engine's WAL, foreign-key and `synchronous=FULL` settings.

Fly sends SIGINT and gives shutdown up to 60 seconds. Litestream's final-sync budget
is 30 seconds. Forced stops and hardware loss can interrupt either budget. Normal
single-machine deploys have a maintenance gap; schema changes must still allow the
previous SQLite image to reopen the database if an image rollback is needed.

## Production cutover gates

Keep `SQLITE_DEPLOY_READY` unset until all of these are recorded for automated deployments:

1. A chosen off-host bucket/prefix, storage credentials and a backup-freshness alert
   destination are configured. Restore from that destination onto a fresh volume,
   validate the data, and measure recovery from detection through service readiness.
2. Invalid credentials, missing replicas, corrupt databases and interrupted writes
   have been exercised with the final image on isolated data. Confirm operator
   response, acceptable write loss, and behavior during an extended backup outage.
3. The volume is commissioned with a verified transfer, the previous API cannot
   write, and only one SQLite machine/replicator can start. Recheck the 256 MiB memory
   allocation with the application and backup process running together.
4. The source-retention plan and a verified off-host archive are recorded. Retire
   the original Postgres resources when the owner authorizes it. Record the tested
   image digest, cutover source/target counts and per-table hashes, and the
   maintenance window before accepting SQLite writes.

For the single-user deployment, manual activation can restore service after the
verified transfer, first off-host sync and successful remote restore. Record any
pending alert delivery as an operating limitation and leave the automated deploy
gate closed until the checks above are complete.

The repository does not provision the bucket, alert destination, Fly volume or
production machine. The bucket, commissioned volume and remote restore evidence
are prerequisites for manual activation. Alert delivery is also required before
automated deployment. A local file-replica test does not satisfy the first gate.

CI checks out the triggering revision in both jobs, tests the published image by
digest, and deploys that digest with the checked-out `fly.toml`. The deploy job also
requires exactly one existing machine and passes `--ha=false`. Repository variable
`SQLITE_DEPLOY_READY=true` enables that job; it does not provision or validate the
external prerequisites. There is no automatic image rollback across the Postgres
boundary.

## Transfer

Freeze all source writers before forward or return transfer. Supply the Postgres
connection through `ASTRO_TRANSFER_POSTGRES_DSN`, injected into the maintenance
process environment. The transfer process prints counts and canonical hashes, not
row contents or connection credentials.

On the stopped writer's volume, the forward command creates and verifies a new
SQLite database and refuses an existing output:

```sh
/transfer --direction postgres-to-sqlite --source-frozen --output /data/astro.db
/transfer --direction verify --sqlite-path /data/astro.db
SQLITE_PATH=/data/astro.db /migrate check
```

An intentional return after SQLite has accepted writes requires a verified return
transfer. Use an explicitly confirmed target database and a new
`astro_return_<suffix>` schema; the tool can create it alongside frozen `public`:

```sh
/transfer --direction sqlite-to-postgres --source-frozen \
  --sqlite-path /data/astro.db --target-database RETURN_DATABASE \
  --target-schema astro_return_recovery
```

Verify with `--direction verify --postgres-schema astro_return_recovery` and the
same SQLite path. Configure the Postgres-backed image with the verified return
schema first in its connection's search path, then validate reads before opening
writes. Deploying the old image against the frozen pre-cutover schema would discard
subsequent SQLite writes. [The transfer guide](../cmd/transfer/README.md) covers
field validation, return-schema permissions and interrupted attempts.

## Recovery

Stop the writer before maintenance and keep it stopped throughout recovery. On Fly,
use the same tested image with the existing volume in an offline maintenance boot;
do not attach another writable copy to a second serving machine. The exact volume
and maintenance-machine attachment must be rehearsed before enabling production.

For a missing/corrupt database, preserve the current database, `-wal`, `-shm`, and
`.astro.db-litestream` directory together in quarantine on the stopped volume.
Remove none of that evidence until recovery is verified. The live target and those
sidecars must be absent before publishing a restored database.

The image command is:

```sh
recover
recover --timestamp 2026-10-18T12:00:00Z
```

For example, on an offline Docker volume with credentials already present in the
process environment:

```sh
docker run --rm -v astro-data:/data \
  -e LITESTREAM_REPLICA_URL -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY \
  astro@sha256:TESTED_DIGEST recover
```

Recovery refuses a held writer lock or an existing target. It restores into a private
scratch directory on the target filesystem, with a 15-minute network-operation
limit, performs Litestream's full integrity check and `/migrate check`, syncs the
file, and publishes it using a no-overwrite hard link. Failure before publication
leaves the live path absent. Successful recovery leaves the writer stopped. Validate
the restored marker/data, then resume the normal container and check `/health`.

Recovery to an older point requires fencing the old writer and preserving the old
backup prefix. Configure a new backup prefix before starting the recovered writer,
so it cannot conflict with a replica containing later transactions. Reset or discard
neither the original data nor its backup history to make a recovery start.

## Backup monitoring and guarantees

`litestream.yml` exposes metrics on port 9092 and enables five-minute LTX
validation. Fly scrapes the API's metrics on port 9091. Litestream's metrics remain
available privately on port 9092. The deployed Machine configuration selects the
API scrape target; listing both ports for the same process does not collect both.
Through a private Fly proxy or SSH session, inspect
`litestream_sync_error_count`,
`litestream_replica_operation_errors_total`, `litestream_disk_full` and disk space.
These signals report failures; the local `litestream_txid` does not prove remote
freshness. Litestream 0.5.17 has no remote-last-success timestamp metric.

The API checks the private Unix control socket at `/tmp/litestream.sock` on startup
and every minute, with one check at a time. It requests `POST /sync` with
`wait: true` and a 30-second timeout, then requires a completed status, the configured
database path, and equal nonzero `txid` and `replicated_txid` values before sending
the Healthchecks success ping. The ping has a five-second timeout. Socket failures,
upload failures, incomplete responses and position mismatches withhold success.
The checker accepts HTTPS only, refuses redirects, and logs static errors without
the secret URL or HTTP error details. Fx shutdown cancels the current check and
waits for the checker to stop.

`litestream list -socket /tmp/litestream.sock -json` exposes `last_sync_at`, but
that timestamp can advance during partial backlog upload and does not establish
that replication has caught up.
`litestream sync -socket /tmp/litestream.sock -wait -timeout 30 /data/astro.db`
returns local and remote transaction IDs after a completed upload.
Configure a Healthchecks interval of one minute with one minute of grace and test
missed-ping alert delivery before enabling automated deployments. An idle
successful check can use cached replication positions, so it confirms observed writes reached the replica;
it does not continuously validate credentials while idle. LTX validation checks the
backup chain; recurring restore drills establish usable data.

Replication is asynchronous. The one-second sync interval is an operating setting,
not a maximum data-loss guarantee. If storage, credentials or the network fail,
the API can continue accepting writes and the off-host loss window grows. The
operating procedure requires an alerted operator to stop writes when the accepted
window is exceeded; automatic refusal based on backup lag is not implemented.
The one-hour recovery target also requires an available operator. This repository
supplies the metrics/checker configuration, but it does not create the receiver,
alert routing, operator coverage or a demonstrated remote recovery time.

Snapshots are hourly with seven-day retention. Fly's independent daily volume
snapshots are supplementary recovery options. Retain independent archival evidence
for cutover, and perform recurring restore drills after storage or image changes.

Before the first completed replication, immediate shutdown can leave no replica:
Litestream's final sync only runs after its database has initialized. The image
check establishes a confirmed remote sync before writing its shutdown marker.
The first production backup and remote restore must exist before opening writes.

References: [Fly volumes](https://docs.fly.io/volumes/overview),
[Fly shutdown configuration](https://docs.fly.io/reference/configuration),
[Litestream configuration](https://litestream.io/reference/config/),
[Litestream restore](https://litestream.io/reference/restore/), and
[SQLite durability](https://sqlite.org/pragma.html#pragma_synchronous).
