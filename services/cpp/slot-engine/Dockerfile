# Build context: repository root (kids_of_gpt/)
FROM ubuntu:24.04
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
        build-essential cmake pkg-config \
        protobuf-compiler-grpc libgrpc++-dev libprotobuf-dev \
        nlohmann-json3-dev libspdlog-dev libgtest-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY proto/slot_engine/                       proto/slot_engine/
COPY services/cpp/slot-engine/src/            services/cpp/slot-engine/src/
COPY services/cpp/slot-engine/tests/          services/cpp/slot-engine/tests/
COPY services/cpp/slot-engine/config/         services/cpp/slot-engine/config/
COPY services/cpp/slot-engine/CMakeLists.txt  services/cpp/slot-engine/

WORKDIR /src/services/cpp/slot-engine
RUN cmake -B build -DCMAKE_BUILD_TYPE=Release -DBUILD_TESTS=ON \
    && cmake --build build --parallel "$(nproc)" \
    && cd build && ctest --output-on-failure

RUN useradd --system --no-create-home appuser
USER appuser

ENV SLOT_ENGINE_GRPC_PORT=50051
ENV SLOT_ENGINE_PAYOUT_TABLE_PATH=/src/services/cpp/slot-engine/config/payout_table.json
EXPOSE 50051
ENTRYPOINT ["/src/services/cpp/slot-engine/build/slot_engine"]
