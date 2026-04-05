# Hackathon Fit Check

Этот документ фиксирует, насколько текущий Go-контур соответствует условиям задачи.

## 1. Бизнес-логика с зависимостью UX от активности

Статус: сделано.

Что есть:
- Прогресс и уровень через XP.
- Режимы UX:
  - `repair` (плохой UX в начале, улучшается при активности).
  - `punish` (нормальный UX в начале, ухудшается при неактивности).
- Инактивность влияет на XP и UI burden score.

Файлы:
- [services/go/platform/progression/main.go](services/go/platform/progression/main.go)
- [services/go/platform/registry-auth/main.go](services/go/platform/registry-auth/main.go)

## 2. Геймификация и unlock-механики

Статус: сделано.

Что есть:
- Уровни и XP.
- Ограничения/разблокировки: `cart_limit`, `delivery_modes`, `precise_eta`, `search`, `filters`.
- Санкции при долге (`banned_until`).
- Смешные UI-механики: `ui_burden_score`, `captcha_enabled`, `forced_action_delay_ms`, `ads_intensity`, `meme_badge`, `funny_message`.

## 3. Progress bar

Статус: сделано.

Что есть:
- `progress_percent`.
- `xp_to_next_level`.
- Отдельный gateway endpoint: `/api/v1/progression/progress-bar`.

Файлы:
- [gateway/bff-go/main.go](gateway/bff-go/main.go)
- [api/openapi/gateway.yaml](api/openapi/gateway.yaml)

## 4. Отслеживание действий пользователя

Статус: сделано.

Что фиксируется:
- XP-активность и стрик.
- Последняя активность.
- Баланс/долг.
- Идемпотентность операций.

## 5. Согласованность между сервисами

Статус: улучшено.

Что выровнено:
- Wallet синхронизирует `wallet_balance/debt_balance` в progression (`/internal/progression/sync-wallet`).
- Register инициирует progression mode и стартовые сущности.
- Gateway проксирует единые API-эндпоинты.

## 6. Контракты

Статус: обновлены.

Обновлено:
- `proto/registry/auth.proto` — `experience_mode`.
- `proto/entitlements/progression.proto` — progress bar/UX burden fields + sync wallet RPC.
- `proto/wallet/ledger.proto` — debtors RPC.
- OpenAPI — endpoint progress bar и новые поля snapshot.

## 7. Что желательно сделать следующим шагом

1. Перевести in-memory state в Postgres/Redis.
2. Добавить rate-limit на login/spin.
3. Добавить интеграционные e2e тесты через gateway.
4. Добавить UI-компонент progress bar во frontend и отображение funny-модификаторов.