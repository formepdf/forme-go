#!/bin/bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

echo "Building Forme WASM (wasm32-wasip1, release)..."
cargo build \
  --manifest-path "$REPO_ROOT/engine/Cargo.toml" \
  --lib \
  --target wasm32-wasip1 \
  --release \
  --features wasm-raw

cp "$REPO_ROOT/engine/target/wasm32-wasip1/release/forme.wasm" "$SCRIPT_DIR/forme.wasm"
echo "Built forme.wasm ($(wc -c < "$SCRIPT_DIR/forme.wasm") bytes)"
