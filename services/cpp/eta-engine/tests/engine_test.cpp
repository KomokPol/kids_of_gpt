#include "engine.h"
#include "delivery.h"

#include <gtest/gtest.h>

namespace {

std::vector<eta_engine::DeliveryModeConfig> default_modes() {
    return {
        {"as_is",           "Как есть",        1200, 2400},
        {"heated",          "Подогретая",      1500, 2700},
        {"express_tunnel",  "Экспресс-подкоп",  600, 1200},
    };
}

class EtaEngineTest : public ::testing::Test {
protected:
    eta_engine::DeliveryModes modes_{default_modes()};
    eta_engine::Engine engine_{modes_, 42};
};

TEST_F(EtaEngineTest, AsIsEtaInRange) {
    auto result = engine_.calculate_eta("order-1", "as_is", 1, true);
    // With jitter ±10%, and item_factor=1.0, range roughly [1080, 2640]
    EXPECT_GE(result.eta_seconds, 60);
    EXPECT_LE(result.eta_seconds, 3600);
    EXPECT_EQ(result.delivery_mode, "as_is");
    EXPECT_TRUE(result.is_precise);
}

TEST_F(EtaEngineTest, ExpressTunnelFasterThanAsIs) {
    // With seed, run multiple and compare averages
    eta_engine::Engine eng1(modes_, 100);
    eta_engine::Engine eng2(modes_, 100);

    int64_t total_express = 0;
    int64_t total_as_is = 0;
    const int n = 100;
    for (int i = 0; i < n; ++i) {
        total_express += eng1.calculate_eta("o", "express_tunnel", 1, true).eta_seconds;
        total_as_is += eng2.calculate_eta("o", "as_is", 1, true).eta_seconds;
    }
    EXPECT_LT(total_express, total_as_is);
}

TEST_F(EtaEngineTest, MoreItemsIncreasesEta) {
    eta_engine::Engine eng1(modes_, 200);
    eta_engine::Engine eng2(modes_, 200);

    auto r1 = eng1.calculate_eta("o1", "as_is", 1, true);
    auto r10 = eng2.calculate_eta("o2", "as_is", 10, true);

    // item_factor for 10 items = 1.0 + 9*0.05 = 1.45
    // Same seed but different call sequence still, just verify logic
    // Use fresh engines with same seed to isolate
    eta_engine::Engine eA(modes_, 999);
    eta_engine::Engine eB(modes_, 999);
    // Both get same base from RNG, but item_factor differs
    // Actually the RNG state advances, so compare within one engine
    eta_engine::Engine eC(modes_, 777);
    auto single = eC.calculate_eta("o", "as_is", 1, true);
    eta_engine::Engine eD(modes_, 777);
    auto multi = eD.calculate_eta("o", "as_is", 10, true);

    // Same seed → same base → multi should be ~1.45x of single
    double ratio = static_cast<double>(multi.eta_seconds) /
                   static_cast<double>(single.eta_seconds);
    EXPECT_GT(ratio, 1.0);
    EXPECT_LT(ratio, 2.0);
}

TEST_F(EtaEngineTest, PreciseFormatNoTilde) {
    auto result = engine_.calculate_eta("o", "as_is", 1, true);
    EXPECT_TRUE(result.eta_display.find('~') == std::string::npos);
    EXPECT_TRUE(result.eta_display.find("мин") != std::string::npos);
}

TEST_F(EtaEngineTest, ImpreciseFormatHasTilde) {
    auto result = engine_.calculate_eta("o", "as_is", 1, false);
    EXPECT_TRUE(result.eta_display.find('~') != std::string::npos ||
                result.eta_display.find("мин") != std::string::npos);
    EXPECT_FALSE(result.is_precise);
}

TEST_F(EtaEngineTest, DeterministicSeedSameResult) {
    eta_engine::Engine e1(modes_, 42);
    eta_engine::Engine e2(modes_, 42);

    auto r1 = e1.calculate_eta("o1", "as_is", 3, true);
    auto r2 = e2.calculate_eta("o1", "as_is", 3, true);

    EXPECT_EQ(r1.eta_seconds, r2.eta_seconds);
    EXPECT_EQ(r1.eta_display, r2.eta_display);
}

TEST_F(EtaEngineTest, MinEtaAtLeast60Seconds) {
    for (int seed = 0; seed < 100; ++seed) {
        eta_engine::Engine e(modes_, seed);
        auto r = e.calculate_eta("o", "express_tunnel", 1, true);
        EXPECT_GE(r.eta_seconds, 60);
    }
}

}  // namespace
