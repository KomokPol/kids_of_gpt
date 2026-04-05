#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
VCPKG_DIR="$REPO_ROOT/vcpkg"

if [ -d "$VCPKG_DIR" ] && [ -f "$VCPKG_DIR/vcpkg" ]; then
    echo "[*] vcpkg already installed at $VCPKG_DIR"
else
    echo "[*] Cloning vcpkg..."
    git clone --depth 1 https://github.com/microsoft/vcpkg.git "$VCPKG_DIR"
    echo "[*] Bootstrapping vcpkg..."
    "$VCPKG_DIR/bootstrap-vcpkg.sh" -disableMetrics
fi

echo "[*] Installing dependencies for slot-engine..."
cd "$REPO_ROOT/services/cpp/slot-engine"
"$VCPKG_DIR/vcpkg" install

echo "[*] Installing dependencies for eta-engine..."
cd "$REPO_ROOT/services/cpp/eta-engine"
"$VCPKG_DIR/vcpkg" install

echo ""
echo "[*] Done. Set this in your shell:"
echo "    export VCPKG_ROOT=$VCPKG_DIR"
echo ""
echo "    Then build with:"
echo "    bash services/cpp/scripts/build-local.sh"
