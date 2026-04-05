#pragma once

#include <cstdint>
#include <optional>
#include <string>

namespace eta_engine {

struct Config {
    uint16_t                grpc_port{50052};
    std::string             delivery_config_path{"config/delivery_modes.json"};
    std::optional<int64_t>  seed;
    std::string             log_level{"info"};

    static Config from_env();
};

}  // namespace eta_engine
