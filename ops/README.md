# SQLite operations

The API stores its data in a local SQLite database and replicates it asynchronously
to a private Tigris bucket with Litestream.

## Runtime contract

`fly.toml` describes one shared-vCPU machine with 256 MiB in `ams`, mounting
`astro_data` at `/data`. The application database is `/data/astro.db`. Keep exactly
one application machine and one replicator for its backup prefix. Separate volumes
have separate lock files; a local file lock cannot prevent two machines from
replicating divergent databases to the same destination.

The image runs as UID/GID 65532. When preparing a volume, make the mounted data
directory and database files writable by that identity. An offline root maintenance
command can set ownership before initialization or recovery. The API and normal
maintenance commands run without root privileges.

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
executes `litestream replicate -exec /astro`. Startup rejects missing, zero-length
or corrupt files and invalid schemas. Neither the API nor the entrypoint restores
or initializes an absent database. `recover` uses the same file lock and fails with
exit code 75 while the writer holds it. Ordinary SQLite writes also retain the
engine's WAL, foreign-key and `synchronous=FULL` settings.

Fly sends SIGINT and gives shutdown up to 60 seconds. Litestream's final-sync budget
is 30 seconds. Forced stops and hardware loss can interrupt either budget. Normal
single-machine deploys have a maintenance gap; schema changes must still allow the
previous SQLite image to reopen the database if an image rollback is needed.

## Deployment

Tags and manual runs of the `build` workflow build and publish an image, exercise
that published image, then deploy it by digest. CI checks out the triggering
revision in both jobs and uses that revision's `fly.toml`. The deploy job requires
exactly one existing machine and passes `--ha=false`.

Normal releases reuse the mounted volume. Startup applies pending Goose schema
migrations under the writer lock before starting Litestream and the API. There is
no separate Fly release command because the application volume is mounted on the
serving machine. Deploying an older image does not undo database schema changes.

The GitHub production environment needs a `FLY_API_TOKEN` secret. Storage credentials
and token encryption keys live in Fly secrets. `fly.toml` supplies the database path
and backup destination. The workflow does not provision the bucket, volume, machine
or optional alert receiver.

## Initialize an empty installation

Stop the writer and use the tested image with the mounted data directory writable
by UID/GID 65532. Set `SQLITE_PATH` to an absent absolute path on that volume:

```sh
/migrate init
/migrate check
```

Initialization refuses existing files and orphaned sidecars. Use recovery for an
existing installation whose database is missing or corrupt. Verify the first
off-host backup and a remote restore when setting up a new installation.

## Recovery

Stop the writer before maintenance and keep it stopped throughout recovery. On Fly,
use the same tested image with the existing volume in an offline maintenance boot;
do not attach another writable copy to a second serving machine. Rehearse the volume
and maintenance-machine attachment as part of recurring recovery drills.

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

When `BACKUP_HEALTHCHECK_URL` is configured, the API checks the private Unix control
socket at `/tmp/litestream.sock` on startup and every minute, with one check at a
time. It requests `POST /sync` with
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
missed-ping alert delivery when configuring alerts. An idle
successful check can use cached replication positions, so it confirms observed writes reached the replica;
it does not continuously validate credentials while idle. LTX validation checks the
backup chain; recurring restore drills establish usable data.

Replication is asynchronous. The one-second sync interval is an operating setting,
not a maximum data-loss guarantee. If storage, credentials or the network fail,
the API can continue accepting writes and the off-host loss window grows. The
operator must stop writes if the accepted loss window is exceeded; automatic
refusal based on backup lag is not implemented. With alerts disabled, backup
failures require manual detection. Recovery requires an available operator. This
repository supplies the metrics/checker configuration, but it does not create the
receiver, alert routing or operator coverage, or establish a remote recovery time.

Snapshots are hourly with seven-day retention. Fly's independent daily volume
snapshots are supplementary recovery options. Perform recurring restore drills
after storage or image changes.

Before the first completed replication, immediate shutdown can leave no replica:
Litestream's final sync only runs after its database has initialized. The image
check establishes a confirmed remote sync before writing its shutdown marker.
Verify that an off-host backup exists before relying on recovery.

The [2026-10-18 commissioning record](2026-10-18-commissioning.md) preserves the
installation's original resource inventory, data verification and observed limits.

References: [Fly volumes](https://docs.fly.io/volumes/overview),
[Fly shutdown configuration](https://docs.fly.io/reference/configuration),
[Litestream configuration](https://litestream.io/reference/config/),
[Litestream restore](https://litestream.io/reference/restore/), and
[SQLite durability](https://sqlite.org/pragma.html#pragma_synchronous).
