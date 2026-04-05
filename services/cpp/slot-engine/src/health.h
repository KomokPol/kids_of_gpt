#pragma once

#include <grpcpp/grpcpp.h>
#include <grpcpp/health_check_service_interface.h>

namespace slot_engine {

void register_health_service(grpc::ServerBuilder& builder);

}  // namespace slot_engine
