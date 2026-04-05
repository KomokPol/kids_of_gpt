#include "config.h"
#include <cstdlib>
#include <string>

namespace eta_engine {

namespace {

std::string env_or(const char* key, const std::string& fallback) {
    const char* val = std::getenv(key);
    return val ? std::string(val) : fallback;
}

int64_t env_int_or(const char* key, int64_t fallback) {
    const char* val = std::getenv(key);
    return val ? std::stoll(val) : fallback;
}

}  // namespace

Config Config::from_env() {
    Config c;
    c.grpc_port            = static_cast<uint16_t>(env_int_or("ETA_ENGINE_GRPC_PORT", 50052));
    c.delivery_config_path = env_or("ETA_ENGINE_DELIVERY_CONFIG_PATH", "config/delivery_modes.json");
    c.log_level            = env_or("ETA_ENGINE_LOG_LEVEL", "info");

    const char* seed_str = std::getenv("ETA_ENGINE_SEED");
    if (seed_str) {
        c.seed = std::stoll(seed_str);
    }
    return c;
}

}  // namespace eta_engine
