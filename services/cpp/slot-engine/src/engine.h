#pragma once

#include "payout_table.h"

#include <cstdint>
#include <optional>
#include <random>
#include <string>
#include <vector>

namespace slot_engine {

struct SpinResult {
    std::vector<int32_t> reels;
    std::string          combination_name;
    double               multiplier;
    int64_t              delta;
    bool                 is_jackpot;
};

class Engine {
public:
    explicit Engine(const PayoutTable& table);

    SpinResult spin(int64_t stake, std::optional<int64_t> seed = std::nullopt);

private:
    std::vector<int32_t> roll_reels(std::mt19937_64& rng) const;

    const PayoutTable& table_;
    std::mt19937_64    default_rng_;
};

}  // namespace slot_engine
