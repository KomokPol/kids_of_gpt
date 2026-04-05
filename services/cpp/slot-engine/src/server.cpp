#include "server.h"

#include <chrono>
#include <spdlog/spdlog.h>

namespace slot_engine {

SlotEngineServiceImpl::SlotEngineServiceImpl(
    Engine& engine, const PayoutTable& table, const Config& cfg)
    : engine_(engine), table_(table), cfg_(cfg)
{}

grpc::Status SlotEngineServiceImpl::CalculateOutcome(
    grpc::ServerContext* /*context*/,
    const zondex::slot_engine::v1::CalculateOutcomeRequest* request,
    zondex::slot_engine::v1::CalculateOutcomeResponse* response)
{
    auto start = std::chrono::steady_clock::now();
    const auto& corr_id = request->correlation_id();

    if (request->stake() <= 0) {
        spdlog::warn(
            R"({{"service":"slot-engine","correlation_id":"{}","msg":"invalid stake: {}"}})",
            corr_id, request->stake());
        return {grpc::INVALID_ARGUMENT, "stake must be positive"};
    }
    if (request->stake() > cfg_.max_stake) {
        spdlog::warn(
            R"({{"service":"slot-engine","correlation_id":"{}","msg":"stake exceeds max: {}"}})",
            corr_id, request->stake());
        return {grpc::INVALID_ARGUMENT, "stake exceeds maximum allowed"};
    }
    if (request->spin_id().empty()) {
        return {grpc::INVALID_ARGUMENT, "spin_id is required"};
    }

    std::optional<int64_t> seed;
    if (request->has_seed()) {
        seed = request->seed();
    }

    SpinResult result;
    try {
        result = engine_.spin(request->stake(), seed);
    } catch (const std::exception& e) {
        spdlog::error(
            R"({{"service":"slot-engine","correlation_id":"{}","msg":"engine error: {}"}})",
            corr_id, e.what());
        return {grpc::INTERNAL, "engine failure"};
    }

    response->set_spin_id(request->spin_id());
    for (auto v : result.reels) {
        response->add_reels(v);
    }
    response->set_combination_name(result.combination_name);
    response->set_multiplier(result.multiplier);
    response->set_delta(result.delta);
    response->set_is_jackpot(result.is_jackpot);

    auto elapsed_us = std::chrono::duration_cast<std::chrono::microseconds>(
        std::chrono::steady_clock::now() - start).count();

    spdlog::info(
        R"({{"service":"slot-engine","correlation_id":"{}",)"
        R"("spin_id":"{}","stake":{},"delta":{},"combination":"{}",)"
        R"("is_jackpot":{},"latency_us":{}}})",
        corr_id, request->spin_id(), request->stake(),
        result.delta, result.combination_name,
        result.is_jackpot, elapsed_us);

    return grpc::Status::OK;
}

grpc::Status SlotEngineServiceImpl::GetPayoutTable(
    grpc::ServerContext* /*context*/,
    const zondex::slot_engine::v1::GetPayoutTableRequest* /*request*/,
    zondex::slot_engine::v1::GetPayoutTableResponse* response)
{
    for (const auto& rule : table_.rules()) {
        auto* proto_rule = response->add_rules();
        proto_rule->set_combination_name(rule.combination_name);
        for (auto p : rule.pattern) {
            proto_rule->add_pattern(p);
        }
        proto_rule->set_multiplier(rule.multiplier);
        proto_rule->set_probability(0.0);
    }
    return grpc::Status::OK;
}

}  // namespace slot_engine
