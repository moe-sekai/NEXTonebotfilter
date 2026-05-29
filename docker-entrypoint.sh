#!/bin/sh
# Ensure the data directory exists even when no volume is mounted, then exec
# the main process so signals reach it (tini already gives us pid 1 niceties).
set -e

mkdir -p /app/data

exec "$@"
