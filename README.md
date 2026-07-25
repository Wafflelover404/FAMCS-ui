# Стратиграфия — Задача 2, Перекрывающиеся прямоугольники 

React + Go приложение для задачи о перекрывающихся цветных костяшках домино
на поле n×n. Решает оба режима из условия (запас / фиксированная
последовательность), объясняет недостижимость с указанием причины, считает
число достижимых картин (обобщения 3–4), и визуализирует ответ в 2D и в 3D —
где вертикальная ось сцены есть время: чем выше слой, тем позже положена
костяшка.

Решатель — прямой порт `overlap.cpp` из корня репозитория, сверенный с ним
дифференциальным тестом на тысячах случайных партий (`cmd/oracle`).

## Быстрый старт (без Docker)

```bash
cd app/backend
go run ./cmd/api          # слушает :8080, без Postgres/Redis/Kafka работает в упрощённом режиме
```

```bash
cd app/frontend
npm install
npm run dev                # http://localhost:5173, проксирует /api и /healthz на :8080
```

## Полный стек (Docker Compose)

```bash
cd app
docker compose up -d --build
```

Поднимает Postgres, Redis, Kafka (KRaft, без ZooKeeper), API, воркер очереди
и фронтенд (nginx). Фронтенд — на `http://localhost:8081`, API — на
`http://localhost:8080`.

- Postgres хранит пресеты, историю решений/подсчётов и результаты
  верификации.
- Redis кэширует `/analyze` и `/solve` по хешу запроса.
- Очередь (по умолчанию Kafka, `QUEUE_DRIVER=redisstream` для Redis Streams)
  используется для фоновых заданий верификации (`POST /api/v1/verify`),
  которые может забрать `cmd/worker`.

## Структура

```
app/
├── backend/
│   ├── cmd/
│   │   ├── overlap/    CLI, argv-совместимый с overlap.cpp
│   │   ├── api/         HTTP API
│   │   ├── worker/      обработчик очереди (верификация)
│   │   └── oracle/      дифференциальный тест против overlap.cpp
│   └── internal/
│       ├── solver/      достижимость, паросочетание, mode1/mode2, перепись, генератор
│       ├── verify/      пакетная верификация (порт stress.py + сверка паросочетания)
│       ├── store/       Postgres (pgx)
│       ├── cache/       Redis
│       ├── queue/       интерфейс очереди + драйверы redisstream/kafka
│       └── api/         маршруты, обработчики
└── frontend/
    └── src/
        ├── components/  Studio (конструктор, 2D/3D, разбор, ходы, плеер), Census
        ├── store/        Zustand: поле, решение, интерфейс
        └── lib/          типы, API-клиент, геометрия, палитра
```

## Проверка решателя

```bash
cd app/backend
go test ./...                                    # модульные и property-тесты
go build -o /tmp/overlap ./cmd/overlap
g++ -O2 -std=c++17 -o /tmp/overlap_cpp ../../overlap.cpp
go run ./cmd/oracle -bin /tmp/overlap_cpp -trials 10000   # сверка с оригиналом
```

Зафиксированные значения из `Afanasyev_Z2.tex` §7: `count1 2×2 {1:1,2:1}` →
`28`; `count2 2×2 [1,2]` → `16`. Фигура из условия (`examples/pdf_pattern.txt`,
запас `2 0 1 1 2 1 1`) — достижима.

## API

| Метод | Путь | Назначение |
|---|---|---|
| POST | `/api/v1/analyze` | достижимость, паросочетание по цветам, дефицит запаса |
| POST | `/api/v1/solve` | найти последовательность ходов (режим 1 или 2) |
| POST | `/api/v1/count` | обобщения 3–4 |
| POST | `/api/v1/simulate` | воспроизвести ходы → кадры, история покрытия клеток |
| POST | `/api/v1/random` | случайная достижимая партия |
| GET | `/api/v1/presets` | примеры, включая фигуру из условия |
| POST | `/api/v1/verify` | фоновая партия верификации (нужны Postgres + очередь) |
| GET | `/api/v1/jobs/{id}` | статус фонового задания |
| GET | `/api/v1/ops/queue` | глубина очереди |
| GET | `/api/v1/benchmarks` | история замеров |
