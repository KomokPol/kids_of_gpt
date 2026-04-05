#include "engine.h"
#include <algorithm>
#include <cmath>
#include <unordered_set>

namespace eta_engine {

Engine::Engine(const DeliveryModes& modes, std::optional<int64_t> seed)
    : modes_(modes)
    , rng_(seed.has_value()
           ? std::mt19937_64(static_cast<uint64_t>(seed.value()))
           : std::mt19937_64(std::random_device{}()))
{}

EtaResult Engine::calculate_eta(
    const std::string& order_id,
    const std::string& delivery_mode,
    int32_t item_count,
    bool precise_eta_enabled)
{
    const auto* mode_cfg = modes_.find(delivery_mode);

    int64_t min_s = mode_cfg ? mode_cfg->min_seconds : 1200;
    int64_t max_s = mode_cfg ? mode_cfg->max_seconds : 2400;

    int64_t raw_eta;
    {
        std::lock_guard<std::mutex> lock(rng_mutex_);

        std::uniform_int_distribution<int64_t> dist(min_s, max_s);
        int64_t base_time = dist(rng_);

        double item_factor = 1.0 + std::max(0, item_count - 1) * 0.05;
        raw_eta = static_cast<int64_t>(
            std::round(static_cast<double>(base_time) * item_factor));

        int64_t jitter_range = raw_eta / 10;
        if (jitter_range > 0) {
            std::uniform_int_distribution<int64_t> jitter_dist(-jitter_range, jitter_range);
            raw_eta += jitter_dist(rng_);
        }
    }

    raw_eta = std::max(raw_eta, int64_t{60});

    return EtaResult{
        .order_id      = order_id,
        .eta_seconds   = raw_eta,
        .eta_display   = format_eta(raw_eta, precise_eta_enabled),
        .delivery_mode = delivery_mode,
        .is_precise    = precise_eta_enabled,
    };
}

std::vector<WindowResult> Engine::get_delivery_windows(
    const std::vector<std::string>& allowed_modes) const
{
    std::unordered_set<std::string> allowed_set(
        allowed_modes.begin(), allowed_modes.end());

    std::vector<WindowResult> windows;
    windows.reserve(modes_.all().size());

    for (const auto& m : modes_.all()) {
        WindowResult w;
        w.delivery_mode    = m.delivery_mode;
        w.display_name     = m.display_name;
        w.min_eta_seconds  = m.min_seconds;
        w.max_eta_seconds  = m.max_seconds;

        if (allowed_set.contains(m.delivery_mode)) {
            w.available = true;
        } else {
            w.available = false;
            w.unavailable_reason = "locked_by_level";
        }
        windows.push_back(std::move(w));
    }
    return windows;
}

std::string Engine::format_eta(int64_t seconds, bool precise) {
    int64_t minutes = seconds / 60;
    if (minutes < 1) minutes = 1;

    if (precise) {
        return std::to_string(minutes) + " мин";
    }

    int64_t rounded = ((minutes + 4) / 5) * 5;
    if (rounded < 5) rounded = 5;
    return "~" + std::to_string(rounded) + " мин";
}

}  // namespace eta_engine
