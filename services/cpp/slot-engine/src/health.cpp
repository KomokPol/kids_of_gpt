#include "health.h"

namespace slot_engine {

void register_health_service(grpc::ServerBuilder& builder) {
    grpc::EnableDefaultHealthCheckService(true);
}

}  // namespace slot_engine
