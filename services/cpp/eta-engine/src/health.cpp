#include "health.h"

namespace eta_engine {

void register_health_service(grpc::ServerBuilder& builder) {
    grpc::EnableDefaultHealthCheckService(true);
}

}  // namespace eta_engine
