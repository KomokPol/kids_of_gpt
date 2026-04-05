#include "payout_table.h"

#include <gtest/gtest.h>

namespace {

slot_engine::PayoutTableConfig make_config() {
    slot_engine::PayoutTableConfig cfg;
    cfg.num_reels   = 3;
    cfg.num_symbols = 6;
    cfg.symbol_weights = {30, 25, 20, 12, 8, 5};
    cfg.rules = {
        {"seven_triple",  {5, 5, 5}, 50.0, true},
        {"bar_triple",    {4, 4, 4}, 20.0, false},
        {"cherry_triple", {2, 2, 2},  5.0, false},
        {"cherry_double", {2, 2, -1}, 1.5, false},
    };
    return cfg;
}

TEST(PayoutTableTest, ExactMatchTriple) {
    slot_engine::PayoutTable table(make_config());
    auto result = table.match({5, 5, 5});
    ASSERT_TRUE(result.has_value());
    EXPECT_EQ(result->combination_name, "seven_triple");
    EXPECT_DOUBLE_EQ(result->multiplier, 50.0);
    EXPECT_TRUE(result->is_jackpot);
}

TEST(PayoutTableTest, WildcardMatch) {
    slot_engine::PayoutTable table(make_config());
    auto result = table.match({2, 2, 3});
    ASSERT_TRUE(result.has_value());
    EXPECT_EQ(result->combination_name, "cherry_double");
    EXPECT_DOUBLE_EQ(result->multiplier, 1.5);
}

TEST(PayoutTableTest, ExactMatchTakesPriorityOverWildcard) {
    slot_engine::PayoutTable table(make_config());
    auto result = table.match({2, 2, 2});
    ASSERT_TRUE(result.has_value());
    EXPECT_EQ(result->combination_name, "cherry_triple");
}

TEST(PayoutTableTest, NoMatchReturnsNullopt) {
    slot_engine::PayoutTable table(make_config());
    auto result = table.match({0, 1, 3});
    EXPECT_FALSE(result.has_value());
}

TEST(PayoutTableTest, WrongReelCountNoMatch) {
    slot_engine::PayoutTable table(make_config());
    auto result = table.match({5, 5});
    EXPECT_FALSE(result.has_value());
}

TEST(PayoutTableTest, LoadFromFile) {
    // This test requires the config file to exist at the expected path
    // It's an integration-level test
    try {
        auto table = slot_engine::PayoutTable::load_from_file(
            "config/payout_table.json");
        EXPECT_GT(table.rules().size(), 0u);
        EXPECT_EQ(table.config().num_reels, 3);
        EXPECT_EQ(table.config().num_symbols, 6);
    } catch (const std::exception&) {
        GTEST_SKIP() << "payout_table.json not found, skipping file load test";
    }
}

}  // namespace
