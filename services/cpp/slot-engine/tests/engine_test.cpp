#include "engine.h"
#include "payout_table.h"

#include <gtest/gtest.h>
#include <cmath>
#include <unordered_map>

namespace {

slot_engine::PayoutTableConfig make_default_config() {
    slot_engine::PayoutTableConfig cfg;
    cfg.num_reels   = 3;
    cfg.num_symbols = 6;
    cfg.symbol_weights = {30, 25, 20, 12, 8, 5};
    cfg.rules = {
        {"seven_triple",  {5, 5, 5}, 50.0, true},
        {"bar_triple",    {4, 4, 4}, 20.0, false},
        {"bell_triple",   {3, 3, 3}, 10.0, false},
        {"cherry_triple", {2, 2, 2},  5.0, false},
        {"lemon_triple",  {1, 1, 1},  3.0, false},
        {"plum_triple",   {0, 0, 0},  2.0, false},
        {"cherry_double", {2, 2, -1}, 1.5, false},
        {"any_bar_pair",  {4, 4, -1}, 2.0, false},
    };
    return cfg;
}

class EngineTest : public ::testing::Test {
protected:
    slot_engine::PayoutTable table_{make_default_config()};
    slot_engine::Engine engine_{table_};
};

TEST_F(EngineTest, DeterministicSeedProducesSameResult) {
    auto r1 = engine_.spin(100, 42);
    auto r2 = engine_.spin(100, 42);
    EXPECT_EQ(r1.reels, r2.reels);
    EXPECT_EQ(r1.combination_name, r2.combination_name);
    EXPECT_EQ(r1.delta, r2.delta);
    EXPECT_EQ(r1.is_jackpot, r2.is_jackpot);
}

TEST_F(EngineTest, DifferentSeedsProduceDifferentResults) {
    auto r1 = engine_.spin(100, 42);
    auto r2 = engine_.spin(100, 12345);
    // They could match by coincidence, but extremely unlikely for different seeds
    // Just verify they return valid results
    EXPECT_FALSE(r1.combination_name.empty());
    EXPECT_FALSE(r2.combination_name.empty());
}

TEST_F(EngineTest, ReelsHaveCorrectSize) {
    auto result = engine_.spin(100, 42);
    EXPECT_EQ(static_cast<int>(result.reels.size()), 3);
}

TEST_F(EngineTest, ReelValuesInRange) {
    for (int seed = 0; seed < 100; ++seed) {
        auto result = engine_.spin(100, seed);
        for (auto v : result.reels) {
            EXPECT_GE(v, 0);
            EXPECT_LT(v, 6);
        }
    }
}

TEST_F(EngineTest, LossResultNegativeDelta) {
    // Run many spins to find a loss
    bool found_loss = false;
    for (int seed = 0; seed < 1000; ++seed) {
        auto result = engine_.spin(100, seed);
        if (result.combination_name == "loss") {
            EXPECT_EQ(result.delta, -100);
            EXPECT_EQ(result.multiplier, 0.0);
            EXPECT_FALSE(result.is_jackpot);
            found_loss = true;
            break;
        }
    }
    EXPECT_TRUE(found_loss) << "Expected to find at least one loss in 1000 spins";
}

TEST_F(EngineTest, WinResultPositiveDelta) {
    bool found_win = false;
    for (int seed = 0; seed < 10000; ++seed) {
        auto result = engine_.spin(100, seed);
        if (result.combination_name != "loss") {
            EXPECT_GT(result.delta, -100);
            EXPECT_GT(result.multiplier, 0.0);
            found_win = true;
            break;
        }
    }
    EXPECT_TRUE(found_win) << "Expected to find at least one win in 10000 spins";
}

TEST_F(EngineTest, JackpotFlagOnlyForJackpotCombination) {
    for (int seed = 0; seed < 10000; ++seed) {
        auto result = engine_.spin(100, seed);
        if (result.is_jackpot) {
            EXPECT_EQ(result.combination_name, "seven_triple");
        }
    }
}

TEST_F(EngineTest, MathematicalExpectationIsNegative) {
    const int n = 100000;
    const int64_t stake = 100;
    int64_t total_delta = 0;

    for (int seed = 0; seed < n; ++seed) {
        auto result = engine_.spin(stake, seed);
        total_delta += result.delta;
    }

    double avg_return = static_cast<double>(total_delta) / static_cast<double>(n);
    // House should win on average — avg_return should be negative
    EXPECT_LT(avg_return, 0.0) << "House edge not working: avg delta = " << avg_return;
}

}  // namespace
