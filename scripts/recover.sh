#!/bin/sh
set -eu
umask 077

if [ "$#" -ne 0 ] && { [ "$#" -ne 2 ] || [ "$1" != '--timestamp' ]; }; then
  echo 'usage: recover [--timestamp RFC3339]' >&2
  exit 1
fi

database_dir=$(dirname "$SQLITE_PATH")
database_name=$(basename "$SQLITE_PATH")
for existing in "$SQLITE_PATH" "$SQLITE_PATH-wal" "$SQLITE_PATH-shm" "$database_dir/.$database_name-litestream"; do
  if [ -e "$existing" ] || [ -L "$existing" ]; then
    echo 'Recovery requires an absent database and sidecars; quarantine existing files while the writer is stopped' >&2
    exit 1
  fi
done

restore_dir=$(mktemp -d "$database_dir/.astro-restore.XXXXXX")
trap 'rm -rf "$restore_dir"' EXIT HUP INT TERM
restore_path="$restore_dir/astro.db"

if [ "$#" -eq 2 ]; then
  timeout --signal=TERM --kill-after=10s 15m litestream restore -config /etc/litestream.yml -integrity-check full -timestamp "$2" -o "$restore_path" "$SQLITE_PATH"
else
  timeout --signal=TERM --kill-after=10s 15m litestream restore -config /etc/litestream.yml -integrity-check full -o "$restore_path" "$SQLITE_PATH"
fi

SQLITE_PATH="$restore_path" /migrate check
sync "$restore_path"
ln -T "$restore_path" "$SQLITE_PATH"
sync "$SQLITE_PATH" "$database_dir"

echo 'Validated database published; recovery did not start the writer'
