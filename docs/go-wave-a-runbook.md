# Go Wave A Runbook

Этот файл нужен как пошаговая инструкция, как запускать Go-ядро и в какие моменты тебе делать коммиты.

## 1. Что уже реализовано

- Общий Go модуль: `go.mod`.
- Платформенные сервисы Wave A:
  - gateway-bff
  - registry-auth
  - profile
  - wallet
  - progression
- Первый продуктовый сервис: sharaga.
- Унифицированные health endpoint у всех остальных Go сервисов.

## 2. Порты по умолчанию

- gateway: 8080
- registry-auth: 8101
- profile: 8102
- wallet: 8103
- progression: 8104
- sharaga: 8105
- leaderboard: 8108
- notification: 8109
- catalog: 8110
- scheduler: 8111
- burmalda: 8112
- barygi: 8113
- kinoshka: 8114
- balanda: 8115

## 3. Быстрая локальная проверка

1. Проверить сборку:

```bash
go test ./...
```

2. Поднять сервисы в отдельных терминалах:

```bash
go run ./services/go/platform/profile
go run ./services/go/platform/wallet
go run ./services/go/platform/progression
go run ./services/go/platform/registry-auth
go run ./services/go/product/sharaga
go run ./gateway/bff-go
```

3. Проверить health:

```bash
curl http://localhost:8080/api/v1/health
curl http://localhost:8101/healthz
curl http://localhost:8102/healthz
curl http://localhost:8103/healthz
curl http://localhost:8104/healthz
curl http://localhost:8105/healthz
```

## 4. Минимальный e2e сценарий

1. Регистрация:

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{
    "nickname":"test",
    "article":"111",
    "term":"5y",
    "cell":"A-1",
    "pin":"1234",
    "photoUrl":"/tmp/p.png",
    "experienceMode":"repair",
    "acceptedRules":true,
    "idempotencyKey":"reg-1"
  }'
```

2. Логин:

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"inmateNumber":"INM-000001","pin":"1234"}'
```

3. Сохранить `accessToken` и сделать защищенные запросы:

```bash
curl http://localhost:8080/api/v1/profile/me -H "Authorization: Bearer <TOKEN>"
curl http://localhost:8080/api/v1/wallet/balance -H "Authorization: Bearer <TOKEN>"
curl http://localhost:8080/api/v1/progression/snapshot -H "Authorization: Bearer <TOKEN>"
curl http://localhost:8080/api/v1/progression/progress-bar -H "Authorization: Bearer <TOKEN>"
```

4. Пройти Sharaga:

```bash
curl -X POST http://localhost:8080/api/v1/sharaga/challenges/ch-1/run \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"mode":"quiz","idempotencyKey":"run-1"}'
```

## 5. Когда тебе делать коммиты

Коммит 1: после зеленого `go test ./...` и успешных health checks.

Коммит 2: после успешного e2e сценария регистрации/логина и чтения profile/wallet/progression.

Коммит 3: после успешного запуска `sharaga` через gateway и проверки idempotency повторным запросом.

Коммит 4: после следующего шага hardening:
- rate limit на login,
- сильный PIN hash (Argon2/bcrypt),
- structured logging,
- persistent storage вместо in-memory.

## 6. Важное замечание

Текущая реализация сделана для быстрого рабочего контура Wave A.
Для production надо заменить in-memory хранилища и простой токен/хеш на полноценный security stack.