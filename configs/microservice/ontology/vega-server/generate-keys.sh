#!/bin/bash

# RSA Key Generation Script for Vega Services
# Usage: ./generate-keys.sh [output_dir]
# Default output directory: current directory
#
# Canonical key store lives under data-connection/ (gitignored & untracked),
# so it survives a `git checkout` that resets the tracked vega-server placeholders.
# This script ensures the canonical key pair, then syncs derived copies into the
# vega-server output dir (private_key.pem, public_key.pem) and rebuilds state.json.
#
# Key-pair rules (in data-connection/):
#   - both present        -> reuse as-is
#   - both missing        -> generate a fresh pair
#   - only private present -> derive the public key from the private key
#   - only public present  -> ERROR: private key cannot be recovered from a public key

set -e

OUTPUT_DIR="${1:-.}"

echo "=== RSA Key Generation Script ==="
echo "Output directory: $OUTPUT_DIR"
echo ""

# Canonical (persistent) key store
DC_DIR="$OUTPUT_DIR/data-connection"
SRC_PRIV="$DC_DIR/private_key.pem"
SRC_PUB="$DC_DIR/public_key.pem"

# Derived copies / outputs in the vega-server dir
DEST_PRIV="$OUTPUT_DIR/private_key.pem"
DEST_PUB="$OUTPUT_DIR/public_key.pem"
STATE_JSON="$OUTPUT_DIR/state.json"

# Helper: check if a file exists and is non-empty
is_non_empty() { [ -s "$1" ]; }

mkdir -p "$DC_DIR"

# --- Ensure the canonical key pair under data-connection/ ---
if is_non_empty "$SRC_PRIV" && is_non_empty "$SRC_PUB"; then
  echo "Found existing canonical key pair, reusing:"
  echo "  $SRC_PRIV"
  echo "  $SRC_PUB"
elif ! is_non_empty "$SRC_PRIV" && ! is_non_empty "$SRC_PUB"; then
  echo "No canonical keys found, generating a fresh pair..."
  openssl genrsa -out "$SRC_PRIV" 2048 2>/dev/null
  echo "  Created: $SRC_PRIV"
  openssl rsa -in "$SRC_PRIV" -pubout -out "$SRC_PUB" 2>/dev/null
  echo "  Created: $SRC_PUB"
elif is_non_empty "$SRC_PRIV"; then
  # Only the private key exists -> derive the public key from it.
  echo "Public key missing, deriving it from the existing private key..."
  openssl rsa -in "$SRC_PRIV" -pubout -out "$SRC_PUB" 2>/dev/null
  echo "  Created: $SRC_PUB"
else
  # Only the public key exists -> the private key is unrecoverable.
  echo "ERROR: private key missing in data-connection (only public_key.pem found)." >&2
  echo "A private key cannot be recovered from a public key." >&2
  echo "Delete $SRC_PUB and re-run this script to generate a fresh key pair." >&2
  exit 1
fi

# --- Sync derived copies into the vega-server dir (overwrite; reproducible) ---
echo ""
echo "Syncing derived copies into $OUTPUT_DIR ..."
cp "$SRC_PRIV" "$DEST_PRIV"
cp "$SRC_PUB" "$DEST_PUB"
chmod 644 "$DEST_PRIV" "$DEST_PUB"
echo "  $DEST_PRIV"
echo "  $DEST_PUB"

echo ""
echo "=== RSA Key Generation Complete ==="

# --- state.json (derived from the public key) ---
echo "Generating state.json from public key..."
PEM_ESCAPED=$(awk 'BEGIN { ORS="" } { sub(/\r$/, ""); if (NR>1) printf "\\n"; printf "%s", $0 }' "$SRC_PUB")
cat > "$STATE_JSON" << EOF
{
  "publicKey": "${PEM_ESCAPED}"
}
EOF
echo "  Created: $STATE_JSON"

echo ""
echo "IMPORTANT: Do NOT commit generated key/state files to version control!"
