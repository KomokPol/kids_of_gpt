#include "engine.h"
#include <numeric>

namespace slot_engine {

Engine::Engine(const PayoutTable& table)
    : table_(table)
    , default_rng_(std::random_device{}())
{}

std::vector<int32_t> Engine::roll_reels(std::mt19937_64& rng) const {
    const auto& cfg = table_.config();
    const auto& weights = cfg.symbol_weights;
    int32_t total_weight = std::accumulate(weights.begin(), weights.end(), 0);

    std::vector<int32_t> reels;
    reels.reserve(cfg.num_reels);

    std::uniform_int_distribution<int32_t> dist(0, total_weight - 1);

    for (int32_t r = 0; r < cfg.num_reels; ++r) {
        int32_t roll = dist(rng);
        int32_t cumulative = 0;
        int32_t symbol = 0;
        for (int32_t s = 0; s < cfg.num_symbols; ++s) {
            cumulative += weights[s];
            if (roll < cumulative) {
                symbol = s;
                break;
            }
        }
        reels.push_back(symbol);
    }
    return reels;
}

SpinResult Engine::spin(int64_t stake, std::optional<int64_t> seed) {
    std::mt19937_64 rng = seed.has_value()
        ? std::mt19937_64(static_cast<uint64_t>(seed.value()))
        : default_rng_;

    auto reels = roll_reels(rng);

    if (!seed.has_value()) {
        default_rng_ = rng;
    }

    auto match = table_.match(reels);

    SpinResult result;
    result.reels = std::move(reels);

    if (match.has_value()) {
        result.combination_name = match->combination_name;
        result.multiplier       = match->multiplier;
        result.is_jackpot       = match->is_jackpot;
        result.delta            = static_cast<int64_t>(
            static_cast<double>(stake) * match->multiplier) - stake;
    } else {
        result.combination_name = "loss";
        result.multiplier       = 0.0;
        result.is_jackpot       = false;
        result.delta            = -stake;
    }
    return result;
}

}  // namespace slot_engine
