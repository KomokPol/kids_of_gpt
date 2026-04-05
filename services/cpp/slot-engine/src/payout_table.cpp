#include "payout_table.h"
#include <fstream>
#include <numeric>
#include <stdexcept>
#include <nlohmann/json.hpp>

namespace slot_engine {

PayoutTable::PayoutTable(PayoutTableConfig cfg) : cfg_(std::move(cfg)) {
    validate();
}

void PayoutTable::validate() const {
    if (cfg_.num_reels <= 0)
        throw std::invalid_argument("num_reels must be positive");
    if (cfg_.num_symbols <= 0)
        throw std::invalid_argument("num_symbols must be positive");
    if (static_cast<int32_t>(cfg_.symbol_weights.size()) != cfg_.num_symbols)
        throw std::invalid_argument("symbol_weights size must equal num_symbols");

    int32_t total = std::accumulate(
        cfg_.symbol_weights.begin(), cfg_.symbol_weights.end(), 0);
    if (total <= 0)
        throw std::invalid_argument("total symbol weight must be positive");

    for (const auto& rule : cfg_.rules) {
        if (static_cast<int32_t>(rule.pattern.size()) != cfg_.num_reels)
            throw std::invalid_argument(
                "pattern size must equal num_reels: " + rule.combination_name);
        for (auto p : rule.pattern) {
            if (p != -1 && (p < 0 || p >= cfg_.num_symbols))
                throw std::invalid_argument(
                    "invalid symbol index in pattern: " + rule.combination_name);
        }
        if (rule.multiplier < 0.0)
            throw std::invalid_argument(
                "multiplier must be non-negative: " + rule.combination_name);
    }
}

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
            if (rule.pattern[i] == -1) continue;
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

double PayoutTable::raw_probability(const PayoutRule& rule) const {
    int32_t total = std::accumulate(
        cfg_.symbol_weights.begin(), cfg_.symbol_weights.end(), 0);
    if (total == 0) return 0.0;

    double prob = 1.0;
    for (auto p : rule.pattern) {
        if (p == -1) continue;
        if (p < 0 || p >= static_cast<int32_t>(cfg_.symbol_weights.size()))
            return 0.0;
        prob *= static_cast<double>(cfg_.symbol_weights[p])
              / static_cast<double>(total);
    }
    return prob;
}

}  // namespace slot_engine
