#!/bin/sh
set -eu
umask 077

: "${SQLITE_PATH:?SQLITE_PATH must name the commissioned database}"
case "$SQLITE_PATH" in
  /*) ;;
  *) echo 'SQLITE_PATH must be absolute inside the container' >&2; exit 1 ;;
esac

case "${1:-}" in
  serve)
    : "${LITESTREAM_REPLICA_URL:?Configure the backup destination before serving}"
    exec flock --exclusive --nonblock --conflict-exit-code 75 --no-fork "$SQLITE_PATH.writer.lock" /scripts/serve.sh
    ;;
  recover)
    shift
    : "${LITESTREAM_REPLICA_URL:?Configure the restore source before recovery}"
    exec flock --exclusive --nonblock --conflict-exit-code 75 --no-fork "$SQLITE_PATH.writer.lock" /scripts/recover.sh "$@"
    ;;
  *) exec "$@" ;;
esac
