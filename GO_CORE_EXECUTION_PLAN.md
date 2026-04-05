# Go Core: execution plan

Этот план фиксирует только Go-ядро и интеграционные границы для остальных стеков.

## 1. Scope Go-ядра

В Go делаем:
- gateway-bff
- registry-auth
- profile-service
- wallet-ledger
- progression-entitlements
- sharaga-service
- burmalda-service
- barygi-service
- kinoshka-service
- balanda-service
- leaderboard-service
- catalog-service
- scheduler-worker

Не делаем в Go:
- AI диалоги (Python)
- рекомендации/поиск/аналитика (Python)
- RNG и ETA ядра (C++)

## 2. Contract-first цикл

Для каждого сервиса:
1. Обновить proto.
2. Обновить gateway OpenAPI.
3. Зафиксировать события в contracts/events.
4. Только потом писать код сервиса.

## 3. Первая реализационная очередь

### Wave A: фундамент

1. `registry-auth`
2. `wallet-ledger`
3. `progression-entitlements`
4. `profile-service`
5. `gateway-bff`

Результат Wave A:
- Регистрация и логин работают через gateway.
- Можно получить entitlement snapshot.
- Можно начислять XP и деньги через service-to-service вызовы.

### Wave B: первая вертикаль

1. `sharaga-service`
2. Связка с wallet + progression
3. Публикация `xp.granted` и `wallet.credited`

Результат Wave B:
- Happy path: register -> login -> sharaga completion -> xp/money updated.

### Wave C: продуктовые API

1. `burmalda-service`
2. `barygi-service`
3. `kinoshka-service`
4. `balanda-service`
5. `leaderboard-service`
6. `catalog-service`

### Wave D: системные задачи

1. `scheduler-worker` (interest accrual, ban timers)
2. hardening: idempotency, retries, tracing, health checks

## 4. API и данные: обязательные инварианты

- Деньги меняются только через wallet-ledger.
- XP и unlock флаги меняются только через progression-entitlements.
- Gateway не содержит доменную бизнес-логику.
- Любая мутация должна поддерживать idempotency key.
- Внешние вызовы только с timeout и correlation-id.

## 5. Минимум по каждому сервису

- `cmd/main.go`
- `internal/config`
- `internal/transport/grpc` (или http для gateway)
- `internal/service`
- `internal/repository`
- `migrations`
- `Dockerfile`
- `README.md`

## 6. Git поток для Go

- Ветка: `feat/go/<service>-<topic>`
- Один PR на один сервисный шаг.
- Контрактные PR идут отдельно и раньше реализации.

Пример:
1. `feat/go/contracts-auth-wallet-progression`
2. `feat/go/registry-auth-bootstrap`
3. `feat/go/wallet-ledger-bootstrap`

## 7. Definition of Ready и Definition of Done

Definition of Ready:
- proto согласован
- OpenAPI endpoint согласован
- event contract согласован
- миграции таблиц определены

Definition of Done:
- сервис собирается
- health/readiness живы
- есть базовые unit tests
- есть structured logs
- есть idempotency для мутаций

## 8. Текущий immediate шаг

1. Зафиксировать контракты proto/openapi (сделано в этом коммите).
2. Сразу после этого начать реализацию Wave A.