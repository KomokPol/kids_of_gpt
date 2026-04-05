# C++ Services

High-performance compute engines for ZONDEX platform.

## Services

| Service | Port | Description |
|---------|------|-------------|
| [slot-engine](slot-engine/) | 50051 | RNG и таблицы выплат для казино «Бурмалда» |
| [eta-engine](eta-engine/) | 50052 | Расчёт ETA и окон доставки для «Баланды» |

## Architecture

Both engines are **pure compute modules** — they perform calculations and return results via gRPC. They do not access wallet-ledger, progression-entitlements, or Kafka directly.

## Cross-platform build

The project builds on **macOS**, **Windows**, and **Linux** using the same CMakeLists.txt. Three ways to build:

### Way 1: Docker (recommended for production parity)

Works on any OS with Docker. Builds inside Ubuntu 24.04 using vcpkg.

```bash
cd services/cpp
docker compose up --build
```

Or individually from repo root:

```bash
docker build -f services/cpp/slot-engine/Dockerfile -t slot-engine .
docker build -f services/cpp/eta-engine/Dockerfile -t eta-engine .
```

### Way 2: vcpkg + local CMake (recommended for development)

Works on macOS, Windows, Linux. No system package manager needed.

```bash
# One-time setup: install vcpkg and all dependencies
bash services/cpp/scripts/setup-vcpkg.sh
export VCPKG_ROOT=$(pwd)/vcpkg

# Build & test both engines
bash services/cpp/scripts/build-local.sh
```

On Windows (PowerShell):

```powershell
git clone --depth 1 https://github.com/microsoft/vcpkg.git vcpkg
.\vcpkg\bootstrap-vcpkg.bat
$env:VCPKG_ROOT = "$(Get-Location)\vcpkg"

cd services\cpp\slot-engine
cmake -B build -DCMAKE_TOOLCHAIN_FILE="$env:VCPKG_ROOT\scripts\buildsystems\vcpkg.cmake"
cmake --build build --parallel
cd build; ctest --output-on-failure
```

### Way 3: System packages (quick local dev)

If you already have grpc, protobuf, etc. installed via your OS package manager:

```bash
# macOS
brew install grpc protobuf nlohmann-json spdlog googletest

# Ubuntu/Debian
apt install -y libgrpc++-dev libprotobuf-dev protobuf-compiler-grpc \
    nlohmann-json3-dev libspdlog-dev libgtest-dev cmake build-essential

# Build
cd services/cpp/slot-engine
cmake -B build -DBUILD_TESTS=ON
cmake --build build --parallel
cd build && ctest --output-on-failure
```

## Proto contracts

- `proto/slot_engine/slot_engine.proto`
- `proto/eta_engine/eta_engine.proto`

## Stack

- C++20
- gRPC / Protobuf
- nlohmann/json for config parsing
- spdlog for structured JSON logging
- Google Test for unit tests
- CMake build system
- vcpkg for cross-platform dependency management
