# eta-engine

High-performance gRPC-сервис для расчёта ETA и окон доставки «Баланды».

## Роль в архитектуре

`eta-engine` — чистый compute-модуль. Принимает параметры заказа от `balanda-service` (Go) по gRPC, рассчитывает время доставки и возвращает результат. Не знает про wallet, Kafka, progression — только математика ETA.

## API (gRPC)

| RPC | Описание |
|-----|----------|
| `CalculateETA` | Рассчитать ETA для конкретного заказа и delivery mode |
| `GetDeliveryWindows` | Получить доступные окна доставки с учётом allowed_delivery_modes |

Proto: `proto/eta_engine/eta_engine.proto`

## Способы доставки

| Mode | Название | Диапазон | Описание |
|------|---------|----------|----------|
| `as_is` | Как есть | 20-40 мин | Базовый, всегда доступен |
| `heated` | Подогретая | 25-45 мин | Открывается по уровню |
| `express_tunnel` | Экспресс-подкоп | 10-20 мин | Премиум, открывается позже |

## Сборка

```bash
cmake -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build --parallel
```

## Тесты

```bash
cmake -B build -DBUILD_TESTS=ON
cmake --build build --parallel
cd build && ctest --output-on-failure
```

## Запуск

```bash
ETA_ENGINE_GRPC_PORT=50052 \
ETA_ENGINE_DELIVERY_CONFIG_PATH=config/delivery_modes.json \
./build/eta_engine
```

## Docker

```bash
docker build -t eta-engine .
docker run -p 50052:50052 eta-engine
```

## Конфигурация (env vars)

| Переменная | По умолчанию | Описание |
|------------|-------------|----------|
| `ETA_ENGINE_GRPC_PORT` | `50052` | Порт gRPC сервера |
| `ETA_ENGINE_REDIS_URL` | `redis://localhost:6379` | URL Redis (зарезервировано) |
| `ETA_ENGINE_DELIVERY_CONFIG_PATH` | `config/delivery_modes.json` | Путь к конфигу доставки |
| `ETA_ENGINE_SEED` | (пусто = random) | Фиксированный seed для детерминизма |
| `ETA_ENGINE_LOG_LEVEL` | `info` | Уровень логирования |

## Формат ETA

- **precise_eta_enabled=true**: точный формат, например `"14 мин"`
- **precise_eta_enabled=false**: грубый формат, округлённый до 5 минут: `"~15 мин"`

Флаг `precise_eta_enabled` приходит от `balanda-service`, который берёт его из `entitlements snapshot`.

## Интеграция

```
balanda-service (Go)
    |
    | gRPC: CalculateETA(order_id, delivery_mode, item_count, precise_eta_enabled)
    |   или GetDeliveryWindows(allowed_delivery_modes)
    v
eta-engine (C++)
    |
    | returns: eta_seconds, eta_display, delivery windows
    v
balanda-service (Go)
    |
    | wallet-ledger.Debit(...)
    | Kafka: order.created
    v
```
