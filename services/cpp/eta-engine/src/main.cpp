#include "config.h"
#include "delivery.h"
#include "engine.h"
#include "health.h"
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
    spdlog::info(R"({"service":"eta-engine","msg":"SIGTERM received, shutting down"})");
    if (g_server) {
        auto deadline = std::chrono::system_clock::now() + std::chrono::seconds(5);
        g_server->Shutdown(deadline);
    }
}

}  // namespace

int main() {
    auto cfg = eta_engine::Config::from_env();

    spdlog::set_default_logger(spdlog::stdout_color_mt("eta-engine"));
    spdlog::set_level(spdlog::level::from_str(cfg.log_level));

    spdlog::info(
        R"({{"service":"eta-engine","msg":"loading delivery config","path":"{}"}})",
        cfg.delivery_config_path);

    auto modes = eta_engine::DeliveryModes::load_from_file(cfg.delivery_config_path);

    spdlog::info(
        R"({{"service":"eta-engine","msg":"delivery modes loaded","count":{}}})",
        modes.all().size());

    eta_engine::Engine engine(modes, cfg.seed);
    eta_engine::EtaEngineServiceImpl service(engine, modes, cfg);

    std::string addr = "0.0.0.0:" + std::to_string(cfg.grpc_port);
    grpc::ServerBuilder builder;
    builder.AddListeningPort(addr, grpc::InsecureServerCredentials());
    builder.RegisterService(&service);

    eta_engine::register_health_service(builder);

    g_server = builder.BuildAndStart();
    if (!g_server) {
        spdlog::critical(R"({"service":"eta-engine","msg":"failed to start server"})");
        return 1;
    }

    std::signal(SIGTERM, shutdown_handler);
    std::signal(SIGINT, shutdown_handler);

    spdlog::info(
        R"({{"service":"eta-engine","msg":"listening","addr":"{}"}})", addr);

    g_server->Wait();

    spdlog::info(R"({"service":"eta-engine","msg":"server stopped"})");
    return 0;
}
