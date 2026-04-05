#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CPP_DIR="$(dirname "$SCRIPT_DIR")"
REPO_ROOT="$(cd "$CPP_DIR/../.." && pwd)"

# ── Detect vcpkg ──────────────────────────────────────────────
if [ -n "${VCPKG_ROOT:-}" ] && [ -f "$VCPKG_ROOT/scripts/buildsystems/vcpkg.cmake" ]; then
    TOOLCHAIN="$VCPKG_ROOT/scripts/buildsystems/vcpkg.cmake"
    echo "[*] Using vcpkg from VCPKG_ROOT=$VCPKG_ROOT"
elif [ -d "$REPO_ROOT/vcpkg" ] && [ -f "$REPO_ROOT/vcpkg/scripts/buildsystems/vcpkg.cmake" ]; then
    TOOLCHAIN="$REPO_ROOT/vcpkg/scripts/buildsystems/vcpkg.cmake"
    echo "[*] Using vcpkg from $REPO_ROOT/vcpkg"
else
    TOOLCHAIN=""
    echo "[*] No vcpkg found — using system-installed packages"
    echo "    To use vcpkg: export VCPKG_ROOT=/path/to/vcpkg"
fi

CMAKE_EXTRA_ARGS=""
if [ -n "$TOOLCHAIN" ]; then
    CMAKE_EXTRA_ARGS="-DCMAKE_TOOLCHAIN_FILE=$TOOLCHAIN"
fi

BUILD_TYPE="${1:-Debug}"
RUN_TESTS="${2:-ON}"

# ── Build slot-engine ─────────────────────────────────────────
echo ""
echo "═══ Building slot-engine ($BUILD_TYPE) ═══"
cd "$CPP_DIR/slot-engine"
cmake -B build \
    -DCMAKE_BUILD_TYPE="$BUILD_TYPE" \
    -DBUILD_TESTS="$RUN_TESTS" \
    $CMAKE_EXTRA_ARGS
cmake --build build --parallel

if [ "$RUN_TESTS" = "ON" ]; then
    echo "─── Running slot-engine tests ───"
    cd build && ctest --output-on-failure && cd ..
fi

# ── Build eta-engine ──────────────────────────────────────────
echo ""
echo "═══ Building eta-engine ($BUILD_TYPE) ═══"
cd "$CPP_DIR/eta-engine"
cmake -B build \
    -DCMAKE_BUILD_TYPE="$BUILD_TYPE" \
    -DBUILD_TESTS="$RUN_TESTS" \
    $CMAKE_EXTRA_ARGS
cmake --build build --parallel

if [ "$RUN_TESTS" = "ON" ]; then
    echo "─── Running eta-engine tests ───"
    cd build && ctest --output-on-failure && cd ..
fi

echo ""
echo "═══ All done ═══"
echo "  slot-engine binary: $CPP_DIR/slot-engine/build/slot_engine"
echo "  eta-engine  binary: $CPP_DIR/eta-engine/build/eta_engine"
