#include "delivery.h"
#include <fstream>
#include <stdexcept>
#include <nlohmann/json.hpp>

namespace eta_engine {

DeliveryModes::DeliveryModes(std::vector<DeliveryModeConfig> modes)
    : modes_(std::move(modes))
{
    for (size_t i = 0; i < modes_.size(); ++i) {
        index_[modes_[i].delivery_mode] = i;
    }
}

DeliveryModes DeliveryModes::load_from_file(const std::string& path) {
    std::ifstream ifs(path);
    if (!ifs.is_open()) {
        throw std::runtime_error("Cannot open delivery config: " + path);
    }
    auto j = nlohmann::json::parse(ifs);

    std::vector<DeliveryModeConfig> modes;
    for (auto& jm : j.at("modes")) {
        DeliveryModeConfig m;
        m.delivery_mode = jm.at("delivery_mode").get<std::string>();
        m.display_name  = jm.at("display_name").get<std::string>();
        m.min_seconds   = jm.at("min_seconds").get<int64_t>();
        m.max_seconds   = jm.at("max_seconds").get<int64_t>();
        modes.push_back(std::move(m));
    }
    return DeliveryModes(std::move(modes));
}

const DeliveryModeConfig* DeliveryModes::find(const std::string& mode) const {
    auto it = index_.find(mode);
    if (it == index_.end()) return nullptr;
    return &modes_[it->second];
}

}  // namespace eta_engine
