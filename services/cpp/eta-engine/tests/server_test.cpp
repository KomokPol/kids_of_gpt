#include "config.h"
#include "delivery.h"
#include "engine.h"
#include "server.h"

#include <gtest/gtest.h>
#include <grpcpp/grpcpp.h>
#include "eta_engine.grpc.pb.h"

#include <memory>

namespace {

std::vector<eta_engine::DeliveryModeConfig> default_modes() {
    return {
        {"as_is",           "Как есть",        1200, 2400},
        {"heated",          "Подогретая",      1500, 2700},
        {"express_tunnel",  "Экспресс-подкоп",  600, 1200},
    };
}

class EtaServerTest : public ::testing::Test {
protected:
    void SetUp() override {
        modes_ = std::make_unique<eta_engine::DeliveryModes>(default_modes());
        engine_ = std::make_unique<eta_engine::Engine>(*modes_, int64_t{42});
        service_ = std::make_unique<eta_engine::EtaEngineServiceImpl>(
            *engine_, *modes_, cfg_);

        grpc::ServerBuilder builder;
        builder.AddListeningPort("localhost:0", grpc::InsecureServerCredentials(), &port_);
        builder.RegisterService(service_.get());
        server_ = builder.BuildAndStart();

        auto channel = grpc::CreateChannel(
            "localhost:" + std::to_string(port_),
            grpc::InsecureChannelCredentials());
        stub_ = zondex::eta_engine::v1::EtaEngineService::NewStub(channel);
    }

    void TearDown() override {
        server_->Shutdown();
    }

    eta_engine::Config cfg_;
    std::unique_ptr<eta_engine::DeliveryModes> modes_;
    std::unique_ptr<eta_engine::Engine> engine_;
    std::unique_ptr<eta_engine::EtaEngineServiceImpl> service_;
    std::unique_ptr<grpc::Server> server_;
    std::unique_ptr<zondex::eta_engine::v1::EtaEngineService::Stub> stub_;
    int port_ = 0;
};

TEST_F(EtaServerTest, CalculateETASuccess) {
    zondex::eta_engine::v1::CalculateETARequest req;
    req.set_order_id("order-1");
    req.set_user_id("user-1");
    req.set_delivery_mode("as_is");
    req.set_item_count(2);
    req.set_precise_eta_enabled(true);
    req.set_correlation_id("corr-1");

    zondex::eta_engine::v1::CalculateETAResponse resp;
    grpc::ClientContext ctx;
    auto status = stub_->CalculateETA(&ctx, req, &resp);

    EXPECT_TRUE(status.ok()) << status.error_message();
    EXPECT_EQ(resp.order_id(), "order-1");
    EXPECT_GT(resp.eta_seconds(), 0);
    EXPECT_FALSE(resp.eta_display().empty());
    EXPECT_EQ(resp.delivery_mode(), "as_is");
    EXPECT_TRUE(resp.is_precise());
}

TEST_F(EtaServerTest, CalculateETAMissingOrderId) {
    zondex::eta_engine::v1::CalculateETARequest req;
    req.set_delivery_mode("as_is");
    req.set_item_count(1);
    req.set_correlation_id("corr-2");

    zondex::eta_engine::v1::CalculateETAResponse resp;
    grpc::ClientContext ctx;
    auto status = stub_->CalculateETA(&ctx, req, &resp);

    EXPECT_FALSE(status.ok());
    EXPECT_EQ(status.error_code(), grpc::INVALID_ARGUMENT);
}

TEST_F(EtaServerTest, CalculateETAZeroItems) {
    zondex::eta_engine::v1::CalculateETARequest req;
    req.set_order_id("order-2");
    req.set_delivery_mode("as_is");
    req.set_item_count(0);
    req.set_correlation_id("corr-3");

    zondex::eta_engine::v1::CalculateETAResponse resp;
    grpc::ClientContext ctx;
    auto status = stub_->CalculateETA(&ctx, req, &resp);

    EXPECT_FALSE(status.ok());
    EXPECT_EQ(status.error_code(), grpc::INVALID_ARGUMENT);
}

TEST_F(EtaServerTest, CalculateETAUnknownMode) {
    zondex::eta_engine::v1::CalculateETARequest req;
    req.set_order_id("order-3");
    req.set_delivery_mode("teleport");
    req.set_item_count(1);
    req.set_correlation_id("corr-4");

    zondex::eta_engine::v1::CalculateETAResponse resp;
    grpc::ClientContext ctx;
    auto status = stub_->CalculateETA(&ctx, req, &resp);

    EXPECT_FALSE(status.ok());
    EXPECT_EQ(status.error_code(), grpc::INVALID_ARGUMENT);
}

TEST_F(EtaServerTest, GetDeliveryWindowsAll) {
    zondex::eta_engine::v1::GetDeliveryWindowsRequest req;
    req.set_user_id("user-1");
    req.add_allowed_delivery_modes("as_is");
    req.add_allowed_delivery_modes("heated");
    req.add_allowed_delivery_modes("express_tunnel");
    req.set_correlation_id("corr-5");

    zondex::eta_engine::v1::GetDeliveryWindowsResponse resp;
    grpc::ClientContext ctx;
    auto status = stub_->GetDeliveryWindows(&ctx, req, &resp);

    EXPECT_TRUE(status.ok()) << status.error_message();
    EXPECT_EQ(resp.windows_size(), 3);

    for (const auto& w : resp.windows()) {
        EXPECT_TRUE(w.available());
    }
}

TEST_F(EtaServerTest, GetDeliveryWindowsLocked) {
    zondex::eta_engine::v1::GetDeliveryWindowsRequest req;
    req.set_user_id("user-1");
    req.add_allowed_delivery_modes("as_is");
    req.set_correlation_id("corr-6");

    zondex::eta_engine::v1::GetDeliveryWindowsResponse resp;
    grpc::ClientContext ctx;
    auto status = stub_->GetDeliveryWindows(&ctx, req, &resp);

    EXPECT_TRUE(status.ok());
    EXPECT_EQ(resp.windows_size(), 3);

    int locked = 0;
    for (const auto& w : resp.windows()) {
        if (!w.available()) {
            ++locked;
            EXPECT_EQ(w.unavailable_reason(), "locked_by_level");
        }
    }
    EXPECT_EQ(locked, 2);
}

}  // namespace
