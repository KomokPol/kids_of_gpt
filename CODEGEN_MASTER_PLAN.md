# ZONDEX: Master Plan для генерации кода

Этот файл фиксирует рабочую спецификацию для поэтапной генерации кода без архитектурного дрейфа.
Использовать как единственный источник правды во время реализации.

## 1) Цель и рамки

Цель MVP: поднять единый супер-апп с общей экономикой, уровнями, долгом и unlock-механиками для сервисов Шарага, Бурмалда, Барыги, Киношка, Баланда.

Жесткие рамки:
- Не проектировать функциональность, обучающую реальному мошенничеству.
- Режимы с манипуляцией переводятся в anti-scam/interrogation (распознавание манипуляций, психологическая устойчивость).
- Контракты (proto/openapi/events) фиксируются до активной генерации кода.
- Один сервис за раз, без параллельного редактирования одной папки разными агентами.

## 2) Нефункциональные требования (базовые)

- Снаружи: REST через gateway-bff.
- Внутри: gRPC между сервисами.
- Межсервисные события: Kafka/Redpanda.
- Идемпотентность денежных и прогрессионных операций обязательна.
- Ledger-модель денег append-only.
- Любой UI unlock дублируется серверной валидацией.
- Все долгие/периодические расчеты через scheduler-worker, не в runtime API.

## 3) Главные доменные оси

1. Identity / Registration
2. Economy / Wallet / Debt
3. Progression / XP / Unlocks / Bans
4. Content / Catalog / Notifications / Analytics

Ключевая идея: progression-entitlements выдает единый snapshot прав пользователя, и фронт/бэк опираются на него одинаково.

## 4) Ключевые инварианты платформы

1. Баланс меняется только через wallet-ledger, прямых update баланса в других сервисах нет.
2. Любое начисление/списание имеет source, correlation_id, idempotency_key.
3. XP меняется только через progression-entitlements.
4. Блокировки/баны считаются централизованно в progression-entitlements.
5. Если entitlement запрещает действие, продуктовый сервис возвращает отказ даже если фронт скрыл ограничение.
6. Проценты по долгу начисляет только scheduler-worker по расписанию.
7. Порог долга для этапирования: debt <= -5000 -> бан на 12 часов.
8. События в шину публикуются после фиксации транзакции (outbox-паттерн).

## 5) Сервисы и их точная ответственность

### 5.1 Платформенные

gateway-bff:
- Единая REST-точка для фронта.
- Агрегация ответов, auth middleware, rate limit.
- Не содержит бизнес-логику денег/XP/долгов.

registry-auth:
- Регистрация по форме, вход по личному номеру + PIN.
- Хеш PIN, выпуск access/refresh JWT.
- Оркестрация: profile create + wallet init + entitlements init.
- Генерация inmate number и barcode asset.

profile-service:
- Профиль пользователя: фото, кличка, статья, камера, срок.
- Хранение медиа-ссылок (файлы в MinIO).

wallet-ledger:
- Единственная бухгалтерия бабок/долга.
- Append-only проводки (credit/debit/debt-accrual).
- Расчет current balance из ledger + кэш snapshot в Redis.
- Поддержка idempotency key.

progression-entitlements:
- XP/уровень/пересчет unlock-состояния.
- Snapshot прав по всем продуктам.
- Бан/разбан, флаги досрочного разбана через задачи.
- API для: GetSnapshot, GrantXP, RecalculateRestrictions.

notification-service:
- In-app/system уведомления.
- Шаблоны для onboarding, банов, заказов, выплат.

leaderboard-service:
- Топы по авторитету, топ должников, недельные лидерборды.

catalog-service:
- Нормализованный каталог товаров/фильмов/челленджей.

search-service:
- Поиск по каталогу (OpenSearch), фильтры, ранжирование.

recommendation-service:
- Рекомендации на основе событий и поведения.

analytics-ingest / analytics-service:
- Прием событий и аналитические агрегации в ClickHouse.

admin-cms:
- Ручное управление контентом, заданиями, банами, корректировками.

scheduler-worker:
- Периодические задачи: проценты долга, daily spin, reset флагов, таймеры банов.

### 5.2 Продуктовые

sharaga-service:
- Режимы: понятия, ситуации/загадки, interrogation/anti-scam.
- Результат -> GrantXP + CreditWallet + leaderboard update.

interrogation-sim:
- AI-сценарии допроса/anti-scam с безопасной логикой.

burmalda-service:
- Ставки, спины, кредит, формирование долга.
- Outcome через slot-engine.
- Денежный результат фиксируется в wallet-ledger.

slot-engine (C++):
- RNG и таблицы выплат.
- Детерминируемый режим для тестов по seed.

barygi-service:
- Маркетплейс, корзина, level-gating, ограничения покупки.
- Опирается на entitlement snapshot.

kinoshka-service:
- Тарифы/подписки/поштучная покупка, доступ к контенту.

media-service:
- Метаданные медиа, постеры, трейлеры.

balanda-service:
- Меню, заказы, способы доставки, unlock-опции.

eta-engine (C++):
- ETA, окна доставки, доступность режимов доставки.

## 6) Entitlements snapshot (канонический минимум)

```json
{
  "level": 2,
  "xp": 340,
  "wallet_balance": 480,
  "debt_balance": -1700,
  "search_enabled": false,
  "filters_enabled": false,
  "cart_limit": 1,
  "delivery_modes": ["as_is"],
  "precise_eta_enabled": false,
  "menu_choice_count": 2,
  "daily_spin_available": true,
  "banned_until": null,
  "can_early_unban_via_tasks": true,
  "film_subscription_tier": "pervohod"
}
```

Правило: любые новые поля добавлять только через версионирование контракта и changelog.

## 7) Минимальные события (event contracts)

Обязательные темы:
- user.events
- auth.events
- xp.events
- wallet.events
- debt.events
- catalog.events
- order.events
- delivery.events
- notification.events
- analytics.raw

Минимальные типы событий:
- user.registered
- auth.logged_in
- xp.granted
- wallet.credited
- wallet.debited
- debt.accrued
- debt.threshold_reached
- ban.applied
- ban.lifted
- sharaga.completed
- burmalda.spin_completed
- order.created
- order.status_changed

Единый envelope события:
- event_id (uuid)
- event_type
- aggregate_type
- aggregate_id
- occurred_at (UTC)
- producer
- correlation_id
- payload (json)

## 8) База данных: минимальный стартовый набор таблиц

registry-auth:
- users
- auth_identities
- refresh_tokens

profile-service:
- profiles
- profile_media

wallet-ledger:
- wallet_accounts
- wallet_entries (append-only)
- wallet_idempotency

progression-entitlements:
- progression_state
- level_rules
- entitlement_snapshots
- bans

burmalda-service:
- spins
- debt_state

sharaga-service:
- challenge_runs
- challenge_results

balanda-service:
- menu_items
- orders
- order_events

kinoshka-service:
- films
- subscriptions
- purchases

## 9) Сквозные сценарии (как должны работать)

### Регистрация
1. Front -> gateway -> registry-auth Register.
2. registry-auth валидирует данные, хеширует PIN, создает user.
3. Вызывает profile-service CreateProfile.
4. Вызывает wallet-ledger InitWallet.
5. Вызывает progression-entitlements InitProgress(level=0, xp=0).
6. Публикует user.registered.
7. Возвращает inmate number, barcode, tokens.

### Прохождение задания в Шараге
1. Front -> gateway -> sharaga-service.
2. sharaga считает score/reward.
3. progression-entitlements GrantXP.
4. wallet-ledger Credit.
5. leaderboard update.
6. Публикация sharaga.completed, xp.granted, wallet.credited.

### Долг и этапирование
1. Spin в burmalda -> outcome из slot-engine.
2. wallet-ledger фиксирует результат.
3. Если баланс < 0: debt_state обновлен.
4. scheduler-worker каждые 10 минут: debt.accrued (+5%).
5. При debt <= -5000: progression бан до now+12h.
6. notification-service отправляет уведомление о бане.

### Заказ в Баланде
1. balanda-service получает entitlement snapshot.
2. Проверяет доступные режимы/опции.
3. Запрашивает ETA у eta-engine.
4. wallet-ledger Debit.
5. Создает order и отправляет события.

## 10) Порядок генерации кода (строгий)

Этап 1. Контракты:
1. proto platform services
2. proto product services
3. openapi gateway
4. event schemas

Этап 2. Фундамент:
1. gateway-bff
2. registry-auth
3. profile-service
4. wallet-ledger
5. progression-entitlements

Этап 3. Первый вертикальный путь:
1. регистрация
2. логин
3. Sharaga completion
4. начисление XP и бабок

Этап 4. Казино и долг:
1. burmalda-service
2. slot-engine
3. debt accrual
4. бан/разбан

Этап 5. Витрины:
1. barygi-service
2. kinoshka-service
3. balanda-service
4. eta-engine

Этап 6. Метаслой:
1. notification
2. leaderboard
3. search
4. recommendation
5. analytics
6. admin-cms

## 11) Definition of Done для каждого сервиса

Каждый сервис считается готовым, если есть:
- grpc/openapi контракт, совпадающий со спецификацией.
- Миграции БД.
- Repository + service layer.
- Health/readiness endpoints.
- Structured logging + correlation_id.
- Таймауты на внешние вызовы.
- Идемпотентность для повторных запросов (где нужна).
- Unit tests на критичную бизнес-логику.
- Dockerfile и локальный запуск.
- Минимальный README сервиса.

## 12) Чеклист безопасности и корректности

- PIN хранится только в виде hash + salt.
- JWT с ротацией refresh token.
- Не доверять entitlement-флагам с фронта.
- Денежные операции только транзакционно.
- Защита от двойного списания (idempotency).
- Audit trail для админских операций.
- Rate limit на login/spin/critical endpoints.
- PII/чувствительные поля не логировать.

## 13) Чеклист наблюдаемости

- Метрики: RPS, p95/p99 latency, error rate.
- Бизнес-метрики: выручка, долг, XP issuance, conversion by service.
- Трассировка межсервисных вызовов (trace_id).
- Алерты: рост debt threshold событий, ошибки wallet, падение auth.

## 14) Риски и анти-паттерны

Риски:
- Расползание логики unlock по продуктовым сервисам.
- Расчет процентов долга в нескольких местах.
- Несогласованность snapshot и фактического state.
- Параллельная генерация кода с конфликтами контрактов.

Что запрещено:
- Менять proto/events без отдельного решения.
- Писать бизнес-логику денег в gateway.
- Обновлять баланс прямыми update без ledger entries.
- Пытаться ускорить MVP ценой удаления idempotency в платежных местах.

## 15) Шаблон запроса к coding-agent на один сервис

Использовать этот шаблон без вольных формулировок:

1. Роль сервиса.
2. Контракты (какие proto/openapi уже зафиксированы).
3. Схема таблиц (минимум для запуска).
4. Внешние зависимости (какие сервисы вызывает).
5. Ограничения (что нельзя менять).
6. Требуемый результат:
   - структура проекта
   - grpc/rest handlers
   - service layer
   - repository
   - migrations
   - tests
   - Dockerfile
   - README

## 16) Первая практическая итерация (что делать сразу)

1. Создать каркас monorepo директорий.
2. Зафиксировать proto для:
   - registry-auth
   - wallet-ledger
   - progression-entitlements
   - sharaga-service
3. Зафиксировать event schemas: user.registered, xp.granted, wallet.credited.
4. Реализовать registry-auth + wallet-ledger + progression-entitlements.
5. Поднять первый e2e сценарий: register -> login -> sharaga reward -> wallet/xp updated.

## 17) Контрольный список перед началом генерации каждого нового сервиса

- Контракт существует и заморожен.
- Понятен владелец данных (source of truth).
- Определены idempotency точки.
- Определены события, которые сервис публикует.
- Прописаны таймауты и retry-стратегия вызовов.
- Прописаны тест-кейсы на критичные ветки.

## 18) Примечание по стилю разработки

- Предпочитать небольшие инкрементальные PR.
- Сначала happy path, затем edge cases.
- Любой спорный момент фиксировать в ADR.
- Никаких silent-изменений контракта.

---

Этот план использовать как дорожную карту для всех последующих генераций кода.