#pragma once

#include "delivery.h"

#include <cstdint>
#include <optional>
#include <random>
#include <string>
#include <vector>

namespace eta_engine {

struct EtaResult {
    std::string order_id;
    int64_t     eta_seconds;
    std::string eta_display;
    std::string delivery_mode;
    bool        is_precise;
};

struct WindowResult {
    std::string delivery_mode;
    std::string display_name;
    int64_t     min_eta_seconds;
    int64_t     max_eta_seconds;
    bool        available;
    std::string unavailable_reason;
};

class Engine {
public:
    Engine(const DeliveryModes& modes, std::optional<int64_t> seed = std::nullopt);

    EtaResult calculate_eta(
        const std::string& order_id,
        const std::string& delivery_mode,
        int32_t item_count,
        bool precise_eta_enabled);

    std::vector<WindowResult> get_delivery_windows(
        const std::vector<std::string>& allowed_modes);

private:
    static std::string format_eta(int64_t seconds, bool precise);

    const DeliveryModes& modes_;
    std::mt19937_64 rng_;
};

}  // namespace eta_engine
