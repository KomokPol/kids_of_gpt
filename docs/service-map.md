# Service Map

## Ownership

- Go owner: gateway, auth, wallet, progression, scheduler, product APIs, proto, openapi, contracts.
- Frontend owner: app shell, admin shell, UI composition, state display.
- Python owner: AI dialog, recommendations, search, analytics, notifications.
- C++ owner: slot RNG, ETA engine.

## Dependency rules

- Frontend depends on openapi and gateway.
- Product services depend on entitlements and wallet through Go contracts.
- Python services consume events and approved read models.
- C++ engines are pure compute modules.