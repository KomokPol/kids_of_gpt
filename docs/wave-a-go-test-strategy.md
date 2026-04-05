# Wave A Go Test Strategy (Minimal)

## Scope

Wave A services:
- gateway-bff
- registry-auth
- wallet-ledger
- progression-entitlements
- profile-service

Primary risk focus for first test pack:
- security
- idempotency
- progression rules

## Minimal Framework Setup

### 1) Repository baseline

1. Keep tests close to service code for fast ownership:
   - gateway/bff-go
   - services/go/platform/registry-auth
   - services/go/platform/wallet
   - services/go/platform/progression
   - services/go/platform/profile
2. Add a shared test helper package:
   - services/go/platform/testkit
3. Add deterministic fixtures and schemas:
   - tests/testdata/events
   - tests/testdata/openapi

### 2) Required tooling only

- Standard library:
  - testing
  - net/http/httptest
  - context
  - time
- Assertion helpers:
  - github.com/stretchr/testify/require
- Mocking (minimal and explicit):
  - go.uber.org/mock/gomock
- gRPC transport integration without network ports:
  - google.golang.org/grpc/test/bufconn
- OpenAPI contract checks for gateway:
  - github.com/getkin/kin-openapi/openapi3
- Event JSON schema checks:
  - github.com/santhosh-tekuri/jsonschema/v5

### 3) Minimal conventions

- Table-driven tests for all business rules.
- Deterministic clock and id generator passed via interfaces.
- Every mutation test must check idempotency behavior.
- Every external call test must assert correlation_id propagation.
- Keep first suite split:
  - Unit: no network, no db.
  - Transport integration: in-memory only (httptest or bufconn).
  - Contract validation: schema-level only.

### 4) Test command surface

- Fast local suite:
  - go test ./... -short
- Full suite:
  - go test ./... -race -count=1
- CI gate for Wave A (first stage):
  - run unit + transport integration + contract validation
  - do not block on long e2e yet

## First 10 Test Cases

Priority scale:
- P0: must pass before merge
- P1: should pass in same sprint

### TC-01 (P0) Register requires rules acceptance

- Area: Security
- Service: registry-auth via gateway /auth/register
- Goal: prevent account creation when acceptedRules is false.
- Steps:
  1. Send register request with valid fields but acceptedRules=false.
  2. Include unique idempotencyKey.
- Expected:
  - Request is rejected (4xx).
  - No profile bootstrap side effect.
  - No wallet/progression init side effect.

### TC-02 (P0) Login fails on wrong PIN

- Area: Security
- Service: registry-auth via gateway /auth/login
- Goal: enforce credential validation.
- Steps:
  1. Seed user with known inmateNumber and pin.
  2. Attempt login with wrong pin.
- Expected:
  - Request is rejected (401 or mapped auth error).
  - No access token and no refresh token returned.

### TC-03 (P0) Protected endpoint rejects missing token

- Area: Security
- Service: gateway /profile/me
- Goal: ensure auth guard on protected read paths.
- Steps:
  1. Call /profile/me without Authorization header.
- Expected:
  - Request is rejected (401).
  - No downstream profile call is executed.

### TC-04 (P0) Register idempotency returns stable identity

- Area: Idempotency
- Service: registry-auth Register
- Goal: duplicate submission must not create duplicate user.
- Steps:
  1. Send valid register request with idempotencyKey=K.
  2. Replay the same request with same key K.
- Expected:
  - Same user_id and inmate_number in both responses.
  - Only one identity record exists.
  - Bootstrap actions (profile/wallet/progression init) happen once.

### TC-05 (P0) Wallet Credit idempotency applies amount once

- Area: Idempotency
- Service: wallet-ledger Credit
- Goal: duplicate credits do not duplicate money.
- Steps:
  1. Credit amount=100 with idempotencyKey=K.
  2. Replay same Credit call with same key K.
- Expected:
  - Final balance increased by exactly 100 once.
  - Ledger has exactly one credit entry for key K.

### TC-06 (P0) Progression GrantXP idempotency applies XP once

- Area: Idempotency + Progression
- Service: progression-entitlements GrantXP
- Goal: duplicate XP grant does not over-level user.
- Steps:
  1. GrantXP xp=150 with idempotencyKey=K.
  2. Replay same GrantXP with key K.
- Expected:
  - Snapshot XP increased once.
  - Level transition evaluated once.
  - If event is emitted, only one xp.granted for key K.

### TC-07 (P0) Progression monotonic level rule

- Area: Progression rules
- Service: progression-entitlements
- Goal: level never decreases when XP increases.
- Steps:
  1. Start from known level and XP.
  2. Apply positive XP grants in sequence.
- Expected:
  - Level is monotonic non-decreasing.
  - XP is monotonic non-decreasing.

### TC-08 (P1) Restriction recalculation reflects debt/balance inputs

- Area: Progression rules
- Service: progression-entitlements RecalculateRestrictions
- Goal: unlock flags are derived only from progression rule engine.
- Steps:
  1. Call RecalculateRestrictions with healthy balance/debt values.
  2. Call with severe debt values.
- Expected:
  - Entitlement flags change according to rule table.
  - Product-facing flags are sourced from snapshot only.

### TC-09 (P1) Manual ban and lift ban consistency

- Area: Security + Progression rules
- Service: progression-entitlements ApplyManualBan and LiftBan
- Goal: ban lifecycle is explicit and reversible by policy.
- Steps:
  1. ApplyManualBan with future banned_until_unix.
  2. Verify snapshot shows active ban.
  3. LiftBan and verify snapshot.
- Expected:
  - banned_until_unix set after ban.
  - banned_until_unix cleared or reset after lift.
  - can_early_unban_via_tasks remains policy-consistent.

### TC-10 (P0) Gateway propagates correlation_id end-to-end

- Area: Security observability baseline
- Service: gateway + Wave A downstream calls
- Goal: preserve request traceability and forensic audit path.
- Steps:
  1. Send request with correlation id header.
  2. Verify outgoing call metadata to downstream services.
- Expected:
  - Same correlation id appears in downstream metadata/log context.
  - Any emitted event includes matching correlation_id.

## Execution Order for Sprint 1

1. TC-01
2. TC-02
3. TC-03
4. TC-04
5. TC-05
6. TC-06
7. TC-10
8. TC-07
9. TC-08
10. TC-09

## Coverage Mapping

- Security: TC-01, TC-02, TC-03, TC-09, TC-10
- Idempotency: TC-04, TC-05, TC-06
- Progression rules: TC-06, TC-07, TC-08, TC-09

## Definition of Ready for implementing these tests

- Service interfaces extracted from main package into internal/service for unit tests.
- Auth middleware shape fixed in gateway for unauthorized test assertions.
- Idempotency storage contract fixed per service (repo method signatures).
- Progression rule table versioned and test fixtures frozen.

## Definition of Done for first test pack

- All P0 tests are green in CI with race detector.
- Flaky rate below 1 percent over 20 reruns.
- Each failed assertion points to explicit invariant (security, idempotency, progression).
- Report includes pass/fail by service and by risk category.
