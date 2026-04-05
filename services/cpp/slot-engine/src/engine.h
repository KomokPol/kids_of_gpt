#pragma once

#include "payout_table.h"

#include <cstdint>
#include <mutex>
#include <optional>
#include <random>
#include <string>
#include <vector>

namespace slot_engine {

struct SpinResult {
    std::vector<int32_t> reels;
    std::string          combination_name;
    double               multiplier{0.0};
    int64_t              delta{0};
    bool                 is_jackpot{false};
};

class Engine {
public:
    explicit Engine(const PayoutTable& table);

    SpinResult spin(int64_t stake, std::optional<int64_t> seed = std::nullopt);

private:
    std::vector<int32_t> roll_reels(std::mt19937_64& rng) const;

    const PayoutTable& table_;
    std::mt19937_64    default_rng_;
    std::mutex         rng_mutex_;
};

}  // namespace slot_engine
