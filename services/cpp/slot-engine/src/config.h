#pragma once

#include <cstdint>
#include <string>

namespace slot_engine {

struct Config {
    uint16_t    grpc_port{50051};
    std::string payout_table_path{"config/payout_table.json"};
    int64_t     max_stake{10000};
    std::string log_level{"info"};

    static Config from_env();
};

}  // namespace slot_engine
