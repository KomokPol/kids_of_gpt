#pragma once

#include "engine.h"
#include "delivery.h"
#include "config.h"
#include "eta_engine.grpc.pb.h"

#include <grpcpp/grpcpp.h>

namespace eta_engine {

class EtaEngineServiceImpl final
    : public zondex::eta_engine::v1::EtaEngineService::Service {
public:
    EtaEngineServiceImpl(Engine& engine, const DeliveryModes& modes, const Config& cfg);

    grpc::Status CalculateETA(
        grpc::ServerContext* context,
        const zondex::eta_engine::v1::CalculateETARequest* request,
        zondex::eta_engine::v1::CalculateETAResponse* response) override;

    grpc::Status GetDeliveryWindows(
        grpc::ServerContext* context,
        const zondex::eta_engine::v1::GetDeliveryWindowsRequest* request,
        zondex::eta_engine::v1::GetDeliveryWindowsResponse* response) override;

private:
    Engine& engine_;
    const DeliveryModes& modes_;
    const Config& cfg_;
};

}  // namespace eta_engine
