#!/bin/sh
set -eu

image=${1:?usage: test-image.sh IMAGE}
repository_dir=$(CDPATH='' cd "$(dirname "$0")/.." && pwd)
TOKEN_PUBLIC_KEY=$(awk '$1 == "public_key:" { print $2 }' "$repository_dir/config/config.yaml")
TOKEN_PRIVATE_KEY=$(awk '$1 == "private_key:" { print $2 }' "$repository_dir/config/config.yaml")
export TOKEN_PUBLIC_KEY TOKEN_PRIVATE_KEY
prefix="astro-image-$(date +%s)-$$"
writer="$prefix-writer"
restored_writer="$prefix-restored"
source_volume="$prefix-source"
backup_volume="$prefix-backup"
corrupt_backup_volume="$prefix-corrupt-backup"
recovery_volume="$prefix-recovery"
empty_volume="$prefix-empty"

cleanup() {
  exit_status=$?
  if [ "$exit_status" -ne 0 ]; then
    docker logs --tail 30 "$writer" 2>/dev/null || true
    docker logs --tail 30 "$restored_writer" 2>/dev/null || true
  fi

  docker rm -f "$writer" "$restored_writer" >/dev/null 2>&1 || true
  docker volume rm "$source_volume" "$backup_volume" "$corrupt_backup_volume" "$recovery_volume" "$empty_volume" >/dev/null 2>&1 || true
}
trap cleanup EXIT HUP INT TERM

for volume in "$source_volume" "$backup_volume" "$corrupt_backup_volume" "$recovery_volume" "$empty_volume"; do
  docker volume create "$volume" >/dev/null
  docker run --rm --user 0 --network none -v "$volume:/data" "$image" /bin/sh -c 'chown 65532:65532 /data'
done

expect_failure() {
  if "$@"; then
    echo 'Expected failure, but command succeeded' >&2
    exit 1
  fi
}

expect_locked() {
  status=0
  "$@" || status=$?
  if [ "$status" -ne 75 ]; then
    echo "Expected writer lock exit 75, got $status" >&2
    exit 1
  fi
}

wait_ready() {
  container=$1
  port=$(docker port "$container" 8080/tcp | sed 's/.*://')
  base_url="http://127.0.0.1:$port"
  attempts=0
  until curl --silent --fail --max-time 1 "$base_url/health" >/dev/null; do
    attempts=$((attempts + 1))
    if [ "$attempts" -ge 60 ] || [ "$(docker inspect --format '{{.State.Running}}' "$container")" != true ]; then
      docker logs --tail 30 "$container"
      echo 'Image did not become ready' >&2
      exit 1
    fi
    sleep 1
  done
}

expect_failure docker run --rm --network none -v "$empty_volume:/data" \
  -e LITESTREAM_REPLICA_URL=file:///backup/astro \
  "$image"
docker run --rm --network none -v "$empty_volume:/data" "$image" /bin/sh -c 'test ! -e "$SQLITE_PATH"'

docker run --rm --network none -v "$source_volume:/data" "$image" /migrate init
expect_failure docker run --rm --network none -v "$source_volume:/data" "$image" /migrate init
expect_failure docker run --rm --network none -v "$source_volume:/data" "$image"

docker run -d --name "$writer" --memory 256m --cpus 1 \
  -p 127.0.0.1::8080 -p 127.0.0.1::9092 -v "$source_volume:/data" -v "$backup_volume:/backup" \
  -e CONFIG_PATH= -e HTTP_PORT=8080 -e METRICS_ADDRESS=:9091 \
  -e HTTP_LIMITER_REQUESTS=120 -e HTTP_LIMITER_EXPIRATION=60s \
  -e TOKEN_PUBLIC_KEY -e TOKEN_PRIVATE_KEY \
  -e LITESTREAM_REPLICA_URL=file:///backup/astro \
  "$image" >/dev/null
wait_ready "$writer"

token=$(curl --silent --show-error --fail --max-time 5 -X POST "$base_url/token")
habit_id=$(curl --silent --show-error --fail --max-time 5 -X POST \
  -H "Authorization: $token" "$base_url/habits?name=image-smoke" | jq -er .id)
curl --silent --show-error --fail --max-time 5 -X POST \
  -H "Authorization: $token" -H 'Content-Type: application/json' \
  --data '{"description":"restore marker","date":"2026-01-02T03:04:05Z"}' \
  "$base_url/habits/$habit_id" >/dev/null
docker exec "$writer" litestream sync -socket /tmp/litestream.sock -wait -timeout 10 /data/astro.db
docker exec "$writer" litestream list -socket /tmp/litestream.sock -json | jq -e '.databases[0].last_sync_at != null' >/dev/null
metrics_port=$(docker port "$writer" 9092/tcp | sed 's/.*://')
curl --silent --show-error --fail --max-time 5 "http://127.0.0.1:$metrics_port/metrics" | \
  awk '/^litestream_txid\{/ { found = 1 } END { exit !found }'

expect_locked docker run --rm --network none -v "$source_volume:/data" \
  -e LITESTREAM_REPLICA_URL=file:///backup/astro \
  "$image"
expect_locked docker run --rm --network none -v "$source_volume:/data" \
  -e LITESTREAM_REPLICA_URL=file:///backup/astro "$image" recover

curl --silent --show-error --fail --max-time 5 -X POST \
  -H "Authorization: $token" -H 'Content-Type: application/json' \
  --data '{"description":"shutdown marker","date":"2026-01-03T03:04:05Z"}' \
  "$base_url/habits/$habit_id" >/dev/null
expected=$(curl --silent --show-error --fail --max-time 5 \
  -H "Authorization: $token" "$base_url/habits/$habit_id" | jq -eS .)
printf '%s\n' "$expected" | jq -e '.name == "image-smoke" and (.activities | length) == 2' >/dev/null

docker kill --signal SIGINT "$writer" >/dev/null
exit_code=$(docker wait "$writer")
test "$exit_code" = 0
docker run --rm --network none -v "$source_volume:/data" "$image" /migrate check

docker start "$writer" >/dev/null
wait_ready "$writer"
actual=$(curl --silent --show-error --fail --max-time 5 \
  -H "Authorization: $token" "$base_url/habits/$habit_id" | jq -eS .)
test "$actual" = "$expected"
docker kill --signal SIGINT "$writer" >/dev/null
exit_code=$(docker wait "$writer")
test "$exit_code" = 0
docker run --rm --network none -v "$source_volume:/data" "$image" /migrate check
echo 'Same-volume restart retained the complete HTTP response'

expect_failure docker run --rm --network none -v "$source_volume:/data" -v "$backup_volume:/backup" \
  -e LITESTREAM_REPLICA_URL=file:///backup/astro "$image" recover
expect_failure docker run --rm --network none -v "$empty_volume:/data" -v "$backup_volume:/backup" \
  -e LITESTREAM_REPLICA_URL=file:///backup/missing "$image" recover
docker run --rm --network none -v "$empty_volume:/data" "$image" /bin/sh -c 'test ! -e "$SQLITE_PATH"'

docker run --rm --network none -v "$backup_volume:/backup:ro" -v "$corrupt_backup_volume:/damaged" \
  "$image" /bin/sh -c '
    cp -a /backup/astro /damaged/astro
    object_count=$(find /damaged/astro -type f -name "*.ltx" | wc -l)
    test "$object_count" -gt 0
    find /damaged/astro -type f -name "*.ltx" -exec /bin/sh -c '\''
      for object do printf "damaged replica object\n" > "$object"; done
    '\'' sh {} +
    echo "Corrupted isolated replica LTX objects: $object_count"
  '
expect_failure docker run --rm --network none -v "$empty_volume:/data" -v "$corrupt_backup_volume:/backup:ro" \
  -e LITESTREAM_REPLICA_URL=file:///backup/astro "$image" recover
docker run --rm --network none -v "$empty_volume:/data" "$image" /bin/sh -c '
  test ! -e "$SQLITE_PATH"
  test ! -L "$SQLITE_PATH"
  test ! -e "$SQLITE_PATH-wal"
  test ! -e "$SQLITE_PATH-shm"
  test ! -e /data/.astro.db-litestream
  test -z "$(find /data -maxdepth 1 -name ".astro-restore.*" -print -quit)"
'
echo 'Corrupt-replica recovery refused publication and removed scratch data'

docker run --rm --network none -v "$recovery_volume:/data" -v "$backup_volume:/backup" \
  -e LITESTREAM_REPLICA_URL=file:///backup/astro "$image" recover
docker run --rm --network none -v "$recovery_volume:/data" "$image" /migrate check
docker run -d --name "$restored_writer" --memory 256m --cpus 1 \
  -p 127.0.0.1::8080 -v "$recovery_volume:/data" -v "$backup_volume:/backup" \
  -e CONFIG_PATH= -e HTTP_PORT=8080 -e METRICS_ADDRESS=:9091 \
  -e HTTP_LIMITER_REQUESTS=120 -e HTTP_LIMITER_EXPIRATION=60s \
  -e TOKEN_PUBLIC_KEY -e TOKEN_PRIVATE_KEY \
  -e LITESTREAM_REPLICA_URL=file:///backup/astro \
  "$image" >/dev/null
wait_ready "$restored_writer"
actual=$(curl --silent --show-error --fail --max-time 5 \
  -H "Authorization: $token" "$base_url/habits/$habit_id" | jq -eS .)
test "$actual" = "$expected"

docker kill --signal SIGINT "$restored_writer" >/dev/null
exit_code=$(docker wait "$restored_writer")
test "$exit_code" = 0

docker run --rm --network none -v "$empty_volume:/data" "$image" /bin/sh -c ': > "$SQLITE_PATH"'
expect_failure docker run --rm --network none -v "$empty_volume:/data" \
  -e LITESTREAM_REPLICA_URL=file:///backup/astro "$image"
docker run --rm --network none -v "$empty_volume:/data" "$image" /bin/sh -c 'test ! -s "$SQLITE_PATH"'

docker run --rm --network none -v "$empty_volume:/data" "$image" /bin/sh -c 'printf corrupt > "$SQLITE_PATH"'
expect_failure docker run --rm --network none -v "$empty_volume:/data" \
  -e LITESTREAM_REPLICA_URL=file:///backup/astro \
  "$image"

echo 'Image checks passed: strict startup, single writer, same-volume restart, SIGINT shutdown, corrupt-replica refusal, verified restore, and retained HTTP data'
