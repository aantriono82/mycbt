#!/usr/bin/env sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
FRONTEND_DIR="$ROOT_DIR/frontend"
PKG_JSON="$FRONTEND_DIR/package.json"

if [ ! -f "$PKG_JSON" ]; then
  echo "Error: file tidak ditemukan: $PKG_JSON" >&2
  exit 1
fi

usage() {
  cat <<'EOF'
Usage:
  scripts/version.sh dump
  scripts/version.sh bump patch|minor|major
  scripts/version.sh bump set <x.y.z>

Contoh:
  scripts/version.sh dump
  scripts/version.sh bump patch
  scripts/version.sh bump set 4.2.0
EOF
}

current_version() {
  node -p "require('$PKG_JSON').version"
}

dump_version() {
  printf "frontend/package.json version: %s\n" "$(current_version)"
}

bump_semver() {
  part="$1"
  npm --prefix "$FRONTEND_DIR" version "$part" --no-git-tag-version >/dev/null
  printf "Version berhasil di-bump (%s): %s\n" "$part" "$(current_version)"
}

set_semver() {
  ver="$1"
  npm --prefix "$FRONTEND_DIR" version "$ver" --no-git-tag-version >/dev/null
  printf "Version berhasil di-set: %s\n" "$(current_version)"
}

cmd="${1:-}"
case "$cmd" in
  dump)
    dump_version
    ;;
  bump)
    sub="${2:-}"
    case "$sub" in
      patch|minor|major)
        bump_semver "$sub"
        ;;
      set)
        ver="${3:-}"
        if [ -z "$ver" ]; then
          echo "Error: versi baru wajib diisi. Contoh: scripts/version.sh bump set 4.2.0" >&2
          exit 1
        fi
        set_semver "$ver"
        ;;
      *)
        echo "Error: mode bump tidak valid: $sub" >&2
        usage
        exit 1
        ;;
    esac
    ;;
  *)
    usage
    exit 1
    ;;
esac
