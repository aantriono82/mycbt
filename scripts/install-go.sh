#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

mkdir -p .tooling

ver_and_file="$(
  node -e '(async()=>{const r=await fetch("https://go.dev/dl/?mode=json"); const j=await r.json(); const rel=j.find(x=>x.stable); if(!rel) throw new Error("no stable release found"); const f=rel.files.find(f=>f.os==="linux"&&f.arch==="amd64"&&f.kind==="archive"); if(!f) throw new Error("no linux/amd64 archive found"); console.log(rel.version, f.filename); })().catch(e=>{console.error(e); process.exit(1);});'
)"

GO_VER="$(echo "$ver_and_file" | awk '{print $1}')"
GO_TGZ="$(echo "$ver_and_file" | awk '{print $2}')"

echo "Installing $GO_VER ($GO_TGZ) into .tooling/..."

curl -fsSLo ".tooling/$GO_TGZ" "https://go.dev/dl/$GO_TGZ"
rm -rf ".tooling/$GO_VER" ".tooling/go"
mkdir -p ".tooling/$GO_VER"
tar -C ".tooling/$GO_VER" -xzf ".tooling/$GO_TGZ"
ln -s "$GO_VER/go" ".tooling/go"

".tooling/go/bin/go" version

