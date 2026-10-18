# SQLite commissioning, 2026-10-18

This record describes the installation at https://gastro.fly.dev on 2026-10-18,
using SQLite on one Amsterdam Machine. Backup alerts were disabled at the owner's
request; `SQLITE_DEPLOY_READY` was unset. See the [operations guide](README.md) for
current deployment and recovery procedures.

## Resources

| Resource | Configuration |
| --- | --- |
| Application | `gastro` |
| Writer | `d891e960a466d8`, `ams`, shared CPU, 256 MiB |
| SQLite volume | `vol_4912endjeyp2y56r`, `astro_data`, 1 GB, mounted at `/data` |
| Database | `/data/astro.db`, owned by UID/GID 65532, mode 0600 |
| Volume snapshots | Daily, seven-day retention |
| Private Tigris bucket | `gastro-sqlite-backups` |
| Litestream prefix | `gastro/production`, endpoint `t3.storage.dev` |

The tested and deployed OCI image index is
`registry.fly.io/gastro@sha256:3fc32a4c57914e20cc2fb4ea50d44ea66811a39ff50ff5362115438ef742a1a9`.
Fly resolved its Linux/amd64 manifest to
`sha256:46ffb2823cd09eff61721056fb7ae59ffa80cf44a1ea13a03ab41c7908593ec4`.
The image contains application code from commit `9abeb72`.

## Recorded database verification

The commissioned database had these row counts and canonical ordered hashes:

| Table | Rows | SHA-256 |
| --- | ---: | --- |
| habits | 134 | `c1b5d441b046dfad3e7121591e795b8a46d1d882fb82a2ff2e9e81a23c060478` |
| groups | 12 | `149d45bc92e0f4f501fc3fa9c9b255f7d4945411355286a8b29cf16b521e0f1d` |
| activities | 6984 | `440bc2a0a64000f438d17cf6b0cdc5866e4617355b95eb059b3c1df47fedec27` |
| groups_habits | 9 | `ec5368c4dd2799bfce891197e463c1582d7b7e2b30eda3db47bedeb6d1f78e7d` |

## Observed behavior

The first off-host backup completed before API activation. The final image's
protected `recover` command restored Tigris data into an empty directory on the
Amsterdam volume and passed integrity and schema checks. The restored data matched
all four recorded hashes.

Public authenticated requests read existing data. A separate synthetic
identity created a habit, activity, group and group link. A fresh Tigris restore
contained those writes. Restarting the actual API Machine retained the exact HTTP
response. The synthetic records were then deleted; another remote restore matched
the counts and hashes above.

Both Fly health checks passed. `/health` returned HTTP 200 with the database up.
Unauthenticated habit access returned 401.
The combined observed API and Litestream RSS was about 79 MiB during these short
checks, with about 126 MiB available to the guest. This is a smoke observation,
not a sustained-load capacity result.

The HTTPS checks from the operator's workstation took approximately 104–193 ms,
including new client connections. Database operations use the local SQLite file
and have no API-to-database network hop.

## Operating limits

Alerts were skipped explicitly. Litestream continues asynchronous replication,
but a future replication failure will not send a backup alert. The one-second
sync interval does not bound write loss. Automated deployment was gated at that time.
Account login and missed-ping delivery were not commissioned.

The restore checks used empty directories on the production volume, not a
replacement-host drill. Host-loss recovery time and sustained peak memory remain
unmeasured.

At the [Fly rates](https://docs.fly.io/about/pricing), the Amsterdam
shared-CPU/256 MiB Machine and 1 GB volume were estimated at about $2.43 per
30 days. Backup requests, traffic and snapshots are additional usage; this is an
estimate rather than an invoice.
