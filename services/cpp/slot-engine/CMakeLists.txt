cmake_minimum_required(VERSION 3.20)
project(slot_engine LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

# Protobuf: CONFIG mode (Homebrew/vcpkg) or MODULE mode (Ubuntu apt)
find_package(Protobuf REQUIRED)
find_package(gRPC CONFIG REQUIRED)
find_package(nlohmann_json CONFIG REQUIRED)
find_package(spdlog CONFIG REQUIRED)

set(PROTO_DIR "${CMAKE_SOURCE_DIR}/../../../proto/slot_engine")
set(PROTO_FILE "${PROTO_DIR}/slot_engine.proto")
set(PROTO_GEN_DIR "${CMAKE_BINARY_DIR}/generated")
file(MAKE_DIRECTORY ${PROTO_GEN_DIR})

find_program(GRPC_CPP_PLUGIN grpc_cpp_plugin REQUIRED)
find_program(PROTOC protoc REQUIRED)

add_custom_command(
    OUTPUT
        "${PROTO_GEN_DIR}/slot_engine.pb.cc"
        "${PROTO_GEN_DIR}/slot_engine.pb.h"
        "${PROTO_GEN_DIR}/slot_engine.grpc.pb.cc"
        "${PROTO_GEN_DIR}/slot_engine.grpc.pb.h"
    COMMAND ${PROTOC}
        "--proto_path=${PROTO_DIR}"
        "--cpp_out=${PROTO_GEN_DIR}"
        "--grpc_out=${PROTO_GEN_DIR}"
        "--plugin=protoc-gen-grpc=${GRPC_CPP_PLUGIN}"
        "${PROTO_FILE}"
    DEPENDS "${PROTO_FILE}"
    COMMENT "Generating slot_engine proto stubs"
)

add_library(slot_engine_proto
    "${PROTO_GEN_DIR}/slot_engine.pb.cc"
    "${PROTO_GEN_DIR}/slot_engine.grpc.pb.cc"
)
target_include_directories(slot_engine_proto PUBLIC ${PROTO_GEN_DIR})
target_link_libraries(slot_engine_proto PUBLIC
    protobuf::libprotobuf
    gRPC::grpc++
)

add_library(slot_engine_lib
    src/config.cpp
    src/payout_table.cpp
    src/engine.cpp
    src/server.cpp
    src/health.cpp
)
target_include_directories(slot_engine_lib PUBLIC src)
target_link_libraries(slot_engine_lib PUBLIC
    slot_engine_proto
    gRPC::grpc++
    nlohmann_json::nlohmann_json
    spdlog::spdlog
)

add_executable(slot_engine src/main.cpp)
target_link_libraries(slot_engine PRIVATE slot_engine_lib)

option(BUILD_TESTS "Build unit tests" ON)
if(BUILD_TESTS)
    enable_testing()
    find_package(GTest CONFIG REQUIRED)
    include(GoogleTest)

    add_executable(slot_engine_tests
        tests/engine_test.cpp
        tests/payout_table_test.cpp
        tests/server_test.cpp
    )
    target_link_libraries(slot_engine_tests PRIVATE
        slot_engine_lib
        GTest::gtest
        GTest::gtest_main
    )
    gtest_discover_tests(slot_engine_tests)
endif()
