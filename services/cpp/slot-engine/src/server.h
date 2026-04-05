#pragma once

#include "engine.h"
#include "payout_table.h"
#include "config.h"
#include "slot_engine.grpc.pb.h"

#include <grpcpp/grpcpp.h>

namespace slot_engine {

class SlotEngineServiceImpl final
    : public zondex::slot_engine::v1::SlotEngineService::Service {
public:
    SlotEngineServiceImpl(Engine& engine, const PayoutTable& table, const Config& cfg);

    grpc::Status CalculateOutcome(
        grpc::ServerContext* context,
        const zondex::slot_engine::v1::CalculateOutcomeRequest* request,
        zondex::slot_engine::v1::CalculateOutcomeResponse* response) override;

    grpc::Status GetPayoutTable(
        grpc::ServerContext* context,
        const zondex::slot_engine::v1::GetPayoutTableRequest* request,
        zondex::slot_engine::v1::GetPayoutTableResponse* response) override;

private:
    Engine& engine_;
    const PayoutTable& table_;
    const Config& cfg_;
};

}  // namespace slot_engine
