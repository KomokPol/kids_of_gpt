# Build context: repository root (kids_of_gpt/)
FROM ubuntu:24.04
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
        build-essential cmake pkg-config \
        protobuf-compiler-grpc libgrpc++-dev libprotobuf-dev \
        nlohmann-json3-dev libspdlog-dev libgtest-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY proto/eta_engine/                        proto/eta_engine/
COPY services/cpp/eta-engine/src/             services/cpp/eta-engine/src/
COPY services/cpp/eta-engine/tests/           services/cpp/eta-engine/tests/
COPY services/cpp/eta-engine/config/          services/cpp/eta-engine/config/
COPY services/cpp/eta-engine/CMakeLists.txt   services/cpp/eta-engine/

WORKDIR /src/services/cpp/eta-engine
RUN cmake -B build -DCMAKE_BUILD_TYPE=Release -DBUILD_TESTS=ON \
    && cmake --build build --parallel "$(nproc)" \
    && cd build && ctest --output-on-failure

RUN useradd --system --no-create-home appuser
USER appuser

ENV ETA_ENGINE_GRPC_PORT=50052
ENV ETA_ENGINE_DELIVERY_CONFIG_PATH=/src/services/cpp/eta-engine/config/delivery_modes.json
EXPOSE 50052
ENTRYPOINT ["/src/services/cpp/eta-engine/build/eta_engine"]
