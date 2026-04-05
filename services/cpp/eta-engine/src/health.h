#pragma once

#include <grpcpp/grpcpp.h>
#include <grpcpp/health_check_service_interface.h>

namespace eta_engine {

void register_health_service(grpc::ServerBuilder& builder);

}  // namespace eta_engine
