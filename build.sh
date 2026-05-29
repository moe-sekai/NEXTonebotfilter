#!/usr/bin/env bash
# ===============================================================
#  NEXTonebotfilter - build script (Linux / macOS)
#  Produces a single self-contained nextonebotfilter binary with
#  the Next.js console embedded.
# ===============================================================
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

EMBED_DIR="backend/internal/server/web"

echo "[1/3] building Next.js console (static export)"
( cd console && [ -d node_modules ] || npm install --no-audit --no-fund )
( cd console && npm run build:export )

echo "[2/3] copying console/out -> ${EMBED_DIR}"
rm -rf "${EMBED_DIR}"
mkdir -p "${EMBED_DIR}"
cp -R console/out/. "${EMBED_DIR}/"

echo "[3/3] building Go binary (CGO disabled, pure-Go SQLite)"
case "$(uname -s)" in
  MINGW*|MSYS*|CYGWIN*) OUT="nextonebotfilter.exe" ;;
  *)                    OUT="nextonebotfilter" ;;
esac
( cd backend && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "${OUT}" ./cmd/nextonebotfilter )

echo
echo "build OK -> backend/${OUT}"
echo "run: ./start.sh"
