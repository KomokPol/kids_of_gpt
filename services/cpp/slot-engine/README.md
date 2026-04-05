# slot-engine

High-performance gRPC-сервис для расчёта результатов спинов казино «Бурмалда».

## Роль в архитектуре

`slot-engine` — чистый compute-модуль. Принимает параметры спина от `burmalda-service` (Go) по gRPC, рассчитывает выпавшую комбинацию и возвращает результат. Не знает про wallet, Kafka, progression — только математика.

## API (gRPC)

| RPC | Описание |
|-----|----------|
| `CalculateOutcome` | Рассчитать результат спина по ставке и (опционально) seed |
| `GetPayoutTable` | Получить текущую таблицу выплат |

Proto: `proto/slot_engine/slot_engine.proto`

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
SLOT_ENGINE_GRPC_PORT=50051 \
SLOT_ENGINE_PAYOUT_TABLE_PATH=config/payout_table.json \
./build/slot_engine
```

## Docker

```bash
docker build -t slot-engine .
docker run -p 50051:50051 slot-engine
```

## Конфигурация (env vars)

| Переменная | По умолчанию | Описание |
|------------|-------------|----------|
| `SLOT_ENGINE_GRPC_PORT` | `50051` | Порт gRPC сервера |
| `SLOT_ENGINE_REDIS_URL` | `redis://localhost:6379` | URL Redis (зарезервировано) |
| `SLOT_ENGINE_PAYOUT_TABLE_PATH` | `config/payout_table.json` | Путь к таблице выплат |
| `SLOT_ENGINE_MAX_STAKE` | `10000` | Максимальная ставка |
| `SLOT_ENGINE_HOUSE_EDGE_BPS` | `500` | House edge в базисных пунктах (500 = 5%) |
| `SLOT_ENGINE_NUM_REELS` | `3` | Количество барабанов |
| `SLOT_ENGINE_NUM_SYMBOLS` | `6` | Количество символов |
| `SLOT_ENGINE_LOG_LEVEL` | `info` | Уровень логирования |

## Детерминированный режим

Для тестов передайте `seed` в `CalculateOutcomeRequest`. При одинаковом seed результат всегда идентичен.

## Интеграция

```
burmalda-service (Go)
    |
    | gRPC: CalculateOutcome(spin_id, user_id, stake, seed?)
    v
slot-engine (C++)
    |
    | returns: reels, combination, multiplier, delta, is_jackpot
    v
burmalda-service (Go)
    |
    | wallet-ledger.Credit/Debit(delta)
    | Kafka: burmalda.spin_completed
    v
```
