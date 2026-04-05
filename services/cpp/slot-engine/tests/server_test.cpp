#include "config.h"
#include "engine.h"
#include "payout_table.h"
#include "server.h"

#include <gtest/gtest.h>
#include <grpcpp/grpcpp.h>
#include "slot_engine.grpc.pb.h"

#include <memory>
#include <thread>

namespace {

slot_engine::PayoutTableConfig make_config() {
    slot_engine::PayoutTableConfig cfg;
    cfg.num_reels   = 3;
    cfg.num_symbols = 6;
    cfg.symbol_weights = {30, 25, 20, 12, 8, 5};
    cfg.rules = {
        {"seven_triple",  {5, 5, 5}, 50.0, true},
        {"cherry_triple", {2, 2, 2},  5.0, false},
    };
    return cfg;
}

class ServerTest : public ::testing::Test {
protected:
    void SetUp() override {
        table_ = std::make_unique<slot_engine::PayoutTable>(make_config());
        engine_ = std::make_unique<slot_engine::Engine>(*table_);
        service_ = std::make_unique<slot_engine::SlotEngineServiceImpl>(
            *engine_, *table_, cfg_);

        grpc::ServerBuilder builder;
        builder.AddListeningPort("localhost:0", grpc::InsecureServerCredentials(), &port_);
        builder.RegisterService(service_.get());
        server_ = builder.BuildAndStart();

        auto channel = grpc::CreateChannel(
            "localhost:" + std::to_string(port_),
            grpc::InsecureChannelCredentials());
        stub_ = zondex::slot_engine::v1::SlotEngineService::NewStub(channel);
    }

    void TearDown() override {
        server_->Shutdown();
    }

    slot_engine::Config cfg_;
    std::unique_ptr<slot_engine::PayoutTable> table_;
    std::unique_ptr<slot_engine::Engine> engine_;
    std::unique_ptr<slot_engine::SlotEngineServiceImpl> service_;
    std::unique_ptr<grpc::Server> server_;
    std::unique_ptr<zondex::slot_engine::v1::SlotEngineService::Stub> stub_;
    int port_ = 0;
};

TEST_F(ServerTest, CalculateOutcomeSuccess) {
    zondex::slot_engine::v1::CalculateOutcomeRequest req;
    req.set_spin_id("test-spin-1");
    req.set_user_id("user-1");
    req.set_stake(100);
    req.set_correlation_id("corr-1");
    req.set_seed(42);

    zondex::slot_engine::v1::CalculateOutcomeResponse resp;
    grpc::ClientContext ctx;
    auto status = stub_->CalculateOutcome(&ctx, req, &resp);

    EXPECT_TRUE(status.ok()) << status.error_message();
    EXPECT_EQ(resp.spin_id(), "test-spin-1");
    EXPECT_EQ(resp.reels_size(), 3);
    EXPECT_FALSE(resp.combination_name().empty());
}

TEST_F(ServerTest, CalculateOutcomeZeroStake) {
    zondex::slot_engine::v1::CalculateOutcomeRequest req;
    req.set_spin_id("test-spin-2");
    req.set_user_id("user-1");
    req.set_stake(0);
    req.set_correlation_id("corr-2");

    zondex::slot_engine::v1::CalculateOutcomeResponse resp;
    grpc::ClientContext ctx;
    auto status = stub_->CalculateOutcome(&ctx, req, &resp);

    EXPECT_FALSE(status.ok());
    EXPECT_EQ(status.error_code(), grpc::INVALID_ARGUMENT);
}

TEST_F(ServerTest, CalculateOutcomeExceedsMaxStake) {
    zondex::slot_engine::v1::CalculateOutcomeRequest req;
    req.set_spin_id("test-spin-3");
    req.set_user_id("user-1");
    req.set_stake(999999);
    req.set_correlation_id("corr-3");

    zondex::slot_engine::v1::CalculateOutcomeResponse resp;
    grpc::ClientContext ctx;
    auto status = stub_->CalculateOutcome(&ctx, req, &resp);

    EXPECT_FALSE(status.ok());
    EXPECT_EQ(status.error_code(), grpc::INVALID_ARGUMENT);
}

TEST_F(ServerTest, CalculateOutcomeMissingSpinId) {
    zondex::slot_engine::v1::CalculateOutcomeRequest req;
    req.set_user_id("user-1");
    req.set_stake(100);
    req.set_correlation_id("corr-4");

    zondex::slot_engine::v1::CalculateOutcomeResponse resp;
    grpc::ClientContext ctx;
    auto status = stub_->CalculateOutcome(&ctx, req, &resp);

    EXPECT_FALSE(status.ok());
    EXPECT_EQ(status.error_code(), grpc::INVALID_ARGUMENT);
}

TEST_F(ServerTest, DeterministicResultsMatchOverGrpc) {
    auto make_request = [](int64_t seed) {
        zondex::slot_engine::v1::CalculateOutcomeRequest req;
        req.set_spin_id("det-spin");
        req.set_user_id("user-1");
        req.set_stake(100);
        req.set_correlation_id("corr-det");
        req.set_seed(seed);
        return req;
    };

    zondex::slot_engine::v1::CalculateOutcomeResponse resp1, resp2;
    {
        grpc::ClientContext ctx;
        auto req = make_request(42);
        stub_->CalculateOutcome(&ctx, req, &resp1);
    }
    {
        grpc::ClientContext ctx;
        auto req = make_request(42);
        stub_->CalculateOutcome(&ctx, req, &resp2);
    }

    EXPECT_EQ(resp1.combination_name(), resp2.combination_name());
    EXPECT_EQ(resp1.delta(), resp2.delta());
    EXPECT_EQ(resp1.reels_size(), resp2.reels_size());
    for (int i = 0; i < resp1.reels_size(); ++i) {
        EXPECT_EQ(resp1.reels(i), resp2.reels(i));
    }
}

TEST_F(ServerTest, GetPayoutTableReturnsRules) {
    zondex::slot_engine::v1::GetPayoutTableRequest req;
    zondex::slot_engine::v1::GetPayoutTableResponse resp;
    grpc::ClientContext ctx;

    auto status = stub_->GetPayoutTable(&ctx, req, &resp);

    EXPECT_TRUE(status.ok()) << status.error_message();
    EXPECT_EQ(resp.rules_size(), 2);
}

}  // namespace
