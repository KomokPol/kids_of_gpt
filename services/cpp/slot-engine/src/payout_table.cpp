#include "payout_table.h"
#include <fstream>
#include <stdexcept>
#include <nlohmann/json.hpp>

namespace slot_engine {

PayoutTable::PayoutTable(PayoutTableConfig cfg) : cfg_(std::move(cfg)) {}

PayoutTable PayoutTable::load_from_file(const std::string& path) {
    std::ifstream ifs(path);
    if (!ifs.is_open()) {
        throw std::runtime_error("Cannot open payout table: " + path);
    }
    auto j = nlohmann::json::parse(ifs);

    PayoutTableConfig cfg;
    cfg.num_reels   = j.at("num_reels").get<int32_t>();
    cfg.num_symbols = j.at("num_symbols").get<int32_t>();
    cfg.symbol_weights = j.at("symbol_weights").get<std::vector<int32_t>>();

    for (auto& jr : j.at("rules")) {
        PayoutRule r;
        r.combination_name = jr.at("combination_name").get<std::string>();
        r.pattern          = jr.at("pattern").get<std::vector<int32_t>>();
        r.multiplier       = jr.at("multiplier").get<double>();
        r.is_jackpot       = jr.at("is_jackpot").get<bool>();
        cfg.rules.push_back(std::move(r));
    }
    return PayoutTable(std::move(cfg));
}

std::optional<MatchResult> PayoutTable::match(const std::vector<int32_t>& reels) const {
    for (const auto& rule : cfg_.rules) {
        if (rule.pattern.size() != reels.size()) continue;

        bool matched = true;
        for (size_t i = 0; i < reels.size(); ++i) {
            if (rule.pattern[i] == -1) continue;  // wildcard
            if (rule.pattern[i] != reels[i]) {
                matched = false;
                break;
            }
        }
        if (matched) {
            return MatchResult{rule.combination_name, rule.multiplier, rule.is_jackpot};
        }
    }
    return std::nullopt;
}

}  // namespace slot_engine
