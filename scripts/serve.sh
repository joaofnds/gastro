#!/bin/sh
set -eu
umask 077

/migrate up
/migrate check

exec litestream replicate -config /etc/litestream.yml -exec /astro
