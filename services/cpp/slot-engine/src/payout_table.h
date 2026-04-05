#pragma once

#include <cstdint>
#include <optional>
#include <string>
#include <vector>

namespace slot_engine {

struct PayoutRule {
    std::string          combination_name;
    std::vector<int32_t> pattern;   // -1 = wildcard (any symbol)
    double               multiplier;
    bool                 is_jackpot;
};

struct PayoutTableConfig {
    int32_t                 num_reels;
    int32_t                 num_symbols;
    std::vector<int32_t>    symbol_weights;
    std::vector<PayoutRule> rules;
};

struct MatchResult {
    std::string combination_name;
    double      multiplier;
    bool        is_jackpot;
};

class PayoutTable {
public:
    explicit PayoutTable(PayoutTableConfig cfg);

    static PayoutTable load_from_file(const std::string& path);

    std::optional<MatchResult> match(const std::vector<int32_t>& reels) const;
    double raw_probability(const PayoutRule& rule) const;

    const PayoutTableConfig& config() const { return cfg_; }
    const std::vector<PayoutRule>& rules() const { return cfg_.rules; }

private:
    void validate() const;

    PayoutTableConfig cfg_;
};

}  // namespace slot_engine
