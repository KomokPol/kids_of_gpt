#include "config.h"
#include <cstdlib>
#include <string>

namespace slot_engine {

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
    c.grpc_port        = static_cast<uint16_t>(env_int_or("SLOT_ENGINE_GRPC_PORT", 50051));
    c.redis_url        = env_or("SLOT_ENGINE_REDIS_URL", "redis://localhost:6379");
    c.payout_table_path = env_or("SLOT_ENGINE_PAYOUT_TABLE_PATH", "config/payout_table.json");
    c.max_stake        = env_int_or("SLOT_ENGINE_MAX_STAKE", 10000);
    c.house_edge_bps   = static_cast<int32_t>(env_int_or("SLOT_ENGINE_HOUSE_EDGE_BPS", 500));
    c.num_reels        = static_cast<int32_t>(env_int_or("SLOT_ENGINE_NUM_REELS", 3));
    c.num_symbols      = static_cast<int32_t>(env_int_or("SLOT_ENGINE_NUM_SYMBOLS", 6));
    c.log_level        = env_or("SLOT_ENGINE_LOG_LEVEL", "info");
    return c;
}

}  // namespace slot_engine
