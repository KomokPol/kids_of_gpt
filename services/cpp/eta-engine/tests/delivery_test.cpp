#include "delivery.h"
#include "engine.h"

#include <gtest/gtest.h>

namespace {

std::vector<eta_engine::DeliveryModeConfig> default_modes() {
    return {
        {"as_is",           "Как есть",        1200, 2400},
        {"heated",          "Подогретая",      1500, 2700},
        {"express_tunnel",  "Экспресс-подкоп",  600, 1200},
    };
}

TEST(DeliveryModesTest, FindExistingMode) {
    eta_engine::DeliveryModes modes(default_modes());
    const auto* m = modes.find("heated");
    ASSERT_NE(m, nullptr);
    EXPECT_EQ(m->display_name, "Подогретая");
    EXPECT_EQ(m->min_seconds, 1500);
    EXPECT_EQ(m->max_seconds, 2700);
}

TEST(DeliveryModesTest, FindNonExistentReturnsNull) {
    eta_engine::DeliveryModes modes(default_modes());
    EXPECT_EQ(modes.find("teleport"), nullptr);
}

TEST(DeliveryModesTest, AllModesReturned) {
    eta_engine::DeliveryModes modes(default_modes());
    EXPECT_EQ(modes.all().size(), 3u);
}

TEST(DeliveryModesTest, LoadFromFile) {
    try {
        auto modes = eta_engine::DeliveryModes::load_from_file(
            "config/delivery_modes.json");
        EXPECT_EQ(modes.all().size(), 3u);
        EXPECT_NE(modes.find("as_is"), nullptr);
        EXPECT_NE(modes.find("heated"), nullptr);
        EXPECT_NE(modes.find("express_tunnel"), nullptr);
    } catch (const std::exception&) {
        GTEST_SKIP() << "delivery_modes.json not found, skipping file load test";
    }
}

TEST(DeliveryWindowsTest, AllModesAllowed) {
    eta_engine::DeliveryModes modes(default_modes());
    eta_engine::Engine engine(modes, 42);

    auto windows = engine.get_delivery_windows({"as_is", "heated", "express_tunnel"});
    EXPECT_EQ(windows.size(), 3u);
    for (const auto& w : windows) {
        EXPECT_TRUE(w.available);
        EXPECT_TRUE(w.unavailable_reason.empty());
    }
}

TEST(DeliveryWindowsTest, OnlyAsIsAllowed) {
    eta_engine::DeliveryModes modes(default_modes());
    eta_engine::Engine engine(modes, 42);

    auto windows = engine.get_delivery_windows({"as_is"});
    EXPECT_EQ(windows.size(), 3u);

    for (const auto& w : windows) {
        if (w.delivery_mode == "as_is") {
            EXPECT_TRUE(w.available);
        } else {
            EXPECT_FALSE(w.available);
            EXPECT_EQ(w.unavailable_reason, "locked_by_level");
        }
    }
}

TEST(DeliveryWindowsTest, EmptyAllowedList) {
    eta_engine::DeliveryModes modes(default_modes());
    eta_engine::Engine engine(modes, 42);

    auto windows = engine.get_delivery_windows({});
    EXPECT_EQ(windows.size(), 3u);
    for (const auto& w : windows) {
        EXPECT_FALSE(w.available);
    }
}

}  // namespace
