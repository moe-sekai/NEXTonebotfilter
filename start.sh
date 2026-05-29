#!/usr/bin/env bash
# ===============================================================
#  NEXTonebotfilter - one-click launcher (Linux / macOS)
#  Single binary, single log stream. Run ./build.sh first.
# ===============================================================
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

PORT="${PORT:-8787}"
LOG="data/nextonebotfilter.log"

case "$(uname -s)" in
  MINGW*|MSYS*|CYGWIN*) BIN="backend/nextonebotfilter.exe" ;;
  *)                    BIN="backend/nextonebotfilter" ;;
esac

if [[ ! -x "${BIN}" ]]; then
  echo "[setup] ${BIN} not found - running ./build.sh first..."
  ./build.sh
fi

mkdir -p data

cat <<EOF
----------------------------------------------------------------
 NEXTonebotfilter
 console + API : http://localhost:${PORT}
 log file      : ${LOG}
----------------------------------------------------------------
 press Ctrl+C to stop
----------------------------------------------------------------
EOF

( sleep 2
  if   command -v xdg-open >/dev/null 2>&1; then xdg-open "http://localhost:${PORT}" >/dev/null 2>&1 || true
  elif command -v open    >/dev/null 2>&1; then open    "http://localhost:${PORT}" >/dev/null 2>&1 || true
  fi
) &

exec "${BIN}" -db data/nextonebotfilter.db -console ":${PORT}" -log "${LOG}" "$@"
