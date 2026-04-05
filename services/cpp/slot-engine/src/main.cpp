#include "config.h"
#include "engine.h"
#include "health.h"
#include "payout_table.h"
#include "server.h"

#include <grpcpp/grpcpp.h>
#include <spdlog/spdlog.h>
#include <spdlog/sinks/stdout_color_sinks.h>

#include <chrono>
#include <csignal>
#include <memory>
#include <string>

namespace {

std::unique_ptr<grpc::Server> g_server;

void shutdown_handler(int /*sig*/) {
    spdlog::info(R"({"service":"slot-engine","msg":"SIGTERM received, shutting down"})");
    if (g_server) {
        auto deadline = std::chrono::system_clock::now() + std::chrono::seconds(5);
        g_server->Shutdown(deadline);
    }
}

}  // namespace

int main() {
    auto cfg = slot_engine::Config::from_env();

    spdlog::set_default_logger(spdlog::stdout_color_mt("slot-engine"));
    spdlog::set_level(spdlog::level::from_str(cfg.log_level));

    spdlog::info(
        R"({{"service":"slot-engine","msg":"loading payout table","path":"{}"}})",
        cfg.payout_table_path);

    auto table = slot_engine::PayoutTable::load_from_file(cfg.payout_table_path);

    spdlog::info(
        R"({{"service":"slot-engine","msg":"payout table loaded","rules":{}}})",
        table.rules().size());

    slot_engine::Engine engine(table);
    slot_engine::SlotEngineServiceImpl service(engine, table, cfg);

    std::string addr = "0.0.0.0:" + std::to_string(cfg.grpc_port);
    grpc::ServerBuilder builder;
    builder.AddListeningPort(addr, grpc::InsecureServerCredentials());
    builder.RegisterService(&service);

    slot_engine::register_health_service(builder);

    g_server = builder.BuildAndStart();
    if (!g_server) {
        spdlog::critical(R"({"service":"slot-engine","msg":"failed to start server"})");
        return 1;
    }

    std::signal(SIGTERM, shutdown_handler);
    std::signal(SIGINT, shutdown_handler);

    spdlog::info(
        R"({{"service":"slot-engine","msg":"listening","addr":"{}"}})", addr);

    g_server->Wait();

    spdlog::info(R"({"service":"slot-engine","msg":"server stopped"})");
    return 0;
}
