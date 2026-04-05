#include "server.h"

#include <chrono>
#include <spdlog/spdlog.h>

namespace eta_engine {

EtaEngineServiceImpl::EtaEngineServiceImpl(
    Engine& engine, const DeliveryModes& modes, const Config& cfg)
    : engine_(engine), modes_(modes), cfg_(cfg)
{}

grpc::Status EtaEngineServiceImpl::CalculateETA(
    grpc::ServerContext* /*context*/,
    const zondex::eta_engine::v1::CalculateETARequest* request,
    zondex::eta_engine::v1::CalculateETAResponse* response)
{
    auto start = std::chrono::steady_clock::now();
    const auto& corr_id = request->correlation_id();

    if (request->order_id().empty()) {
        return {grpc::INVALID_ARGUMENT, "order_id is required"};
    }
    if (request->delivery_mode().empty()) {
        return {grpc::INVALID_ARGUMENT, "delivery_mode is required"};
    }
    if (request->item_count() <= 0) {
        return {grpc::INVALID_ARGUMENT, "item_count must be positive"};
    }

    if (!modes_.find(request->delivery_mode())) {
        spdlog::warn(
            R"({{"service":"eta-engine","correlation_id":"{}","msg":"unknown delivery_mode: {}"}})",
            corr_id, request->delivery_mode());
        return {grpc::INVALID_ARGUMENT, "unknown delivery_mode"};
    }

    EtaResult result;
    try {
        result = engine_.calculate_eta(
            request->order_id(),
            request->delivery_mode(),
            request->item_count(),
            request->precise_eta_enabled());
    } catch (const std::exception& e) {
        spdlog::error(
            R"({{"service":"eta-engine","correlation_id":"{}","msg":"engine error: {}"}})",
            corr_id, e.what());
        return {grpc::INTERNAL, "engine failure"};
    }

    response->set_order_id(result.order_id);
    response->set_eta_seconds(result.eta_seconds);
    response->set_eta_display(result.eta_display);
    response->set_delivery_mode(result.delivery_mode);
    response->set_is_precise(result.is_precise);

    auto elapsed_us = std::chrono::duration_cast<std::chrono::microseconds>(
        std::chrono::steady_clock::now() - start).count();

    spdlog::info(
        R"({{"service":"eta-engine","correlation_id":"{}",)"
        R"("order_id":"{}","delivery_mode":"{}","item_count":{},)"
        R"("eta_seconds":{},"is_precise":{},"latency_us":{}}})",
        corr_id, result.order_id, result.delivery_mode,
        request->item_count(), result.eta_seconds,
        result.is_precise, elapsed_us);

    return grpc::Status::OK;
}

grpc::Status EtaEngineServiceImpl::GetDeliveryWindows(
    grpc::ServerContext* /*context*/,
    const zondex::eta_engine::v1::GetDeliveryWindowsRequest* request,
    zondex::eta_engine::v1::GetDeliveryWindowsResponse* response)
{
    auto start = std::chrono::steady_clock::now();
    const auto& corr_id = request->correlation_id();

    std::vector<std::string> allowed;
    for (const auto& m : request->allowed_delivery_modes()) {
        allowed.push_back(m);
    }

    auto windows = engine_.get_delivery_windows(allowed);

    for (const auto& w : windows) {
        auto* pw = response->add_windows();
        pw->set_delivery_mode(w.delivery_mode);
        pw->set_display_name(w.display_name);
        pw->set_min_eta_seconds(w.min_eta_seconds);
        pw->set_max_eta_seconds(w.max_eta_seconds);
        pw->set_available(w.available);
        if (!w.unavailable_reason.empty()) {
            pw->set_unavailable_reason(w.unavailable_reason);
        }
    }

    auto elapsed_us = std::chrono::duration_cast<std::chrono::microseconds>(
        std::chrono::steady_clock::now() - start).count();

    spdlog::info(
        R"({{"service":"eta-engine","correlation_id":"{}",)"
        R"("rpc":"GetDeliveryWindows","allowed_count":{},"total_windows":{},"latency_us":{}}})",
        corr_id, allowed.size(), windows.size(), elapsed_us);

    return grpc::Status::OK;
}

}  // namespace eta_engine
