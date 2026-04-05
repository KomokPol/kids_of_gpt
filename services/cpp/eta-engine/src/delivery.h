#pragma once

#include <cstdint>
#include <string>
#include <unordered_map>
#include <vector>

namespace eta_engine {

struct DeliveryModeConfig {
    std::string delivery_mode;
    std::string display_name;
    int64_t     min_seconds;
    int64_t     max_seconds;
};

class DeliveryModes {
public:
    explicit DeliveryModes(std::vector<DeliveryModeConfig> modes);

    static DeliveryModes load_from_file(const std::string& path);

    const DeliveryModeConfig* find(const std::string& mode) const;
    const std::vector<DeliveryModeConfig>& all() const { return modes_; }

private:
    std::vector<DeliveryModeConfig> modes_;
    std::unordered_map<std::string, size_t> index_;
};

}  // namespace eta_engine
