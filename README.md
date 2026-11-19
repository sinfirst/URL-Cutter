# URL-Cutter

> Учебный проект курса «Продвинутый Go-разработчик» (Яндекс Практикум).

[![Go Version](https://img.shields.io/github/go-mod/go-version/sinfirst/URL-Cutter)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Сервис сокращения URL на Go: HTTP и gRPC API для создания коротких ссылок и управления ими. Реализованы базовые эндпоинты, JSON API, сжатие, несколько вариантов хранения, аутентификация, асинхронное удаление, gRPC и свой staticlint.

## Возможности

- **Сокращение URL**: создание коротких ссылок через `POST /` (text/plain) и `POST /api/shorten` (JSON).
- **Редирект**: `GET /{id}` — перенаправление на оригинальный URL с кодом 307.
- **Пакетная обработка**: `POST /api/shorten/batch` для множественного создания коротких ссылок.
- **Получение всех ссылок пользователя**: `GET /api/user/urls`.
- **Асинхронное удаление**: `DELETE /api/user/urls` — мягкое удаление ссылок.
- **Проверка соединения с БД**: `GET /ping`.
- **Статистика для доверенных подсетей**: `GET /api/internal/stats` (количество URL и пользователей).
- **Сжатие запросов/ответов**: gzip для типов `application/json` и `text/html`.
- **Хранение**:
  - в памяти (in-memory);
  - в файле (JSON);
  - в PostgreSQL (с уникальным индексом на оригинальный URL).
- **Аутентификация**: пользовательские cookie с подписью.
- **Graceful shutdown**: корректное завершение с сохранением данных.
- **gRPC**: те же операции доступны по протоколу gRPC наряду с HTTP.
- **Конфигурация**: флаги, переменные окружения, JSON-файл.
- **Версионирование бинарного файла**: вывод build version, date, commit при старте.
- **Статический анализ**: multichecker с собственным правилом (запрет `os.Exit` в `main`).
- **Тесты**: юнит-тесты и примеры в `internal/app/app_test.go` и `internal/app/example_test.go`, плюс бенчмарки.

## Архитектура

Проект организован по стандартной для Go многослойной архитектуре с разделением ответственности.

```
.
├── .github                   # GitHub Actions (CI)
├── cmd
│   ├── shortener             # точка входа HTTP- и gRPC-сервера
│   └── staticlint            # multichecker для статического анализа
├── internal
│   ├── app                   # инициализация приложения, запуск сервера
│   ├── config                # конфигурация (флаги, env, JSON)
│   ├── grpc_server           # gRPC-сервер
│   ├── handlers              # HTTP-обработчики
│   ├── middleware            # промежуточное ПО для HTTP
│   │   ├── compress          # сжатие ответов (gzip)
│   │   ├── jwtgen            # аутентификация через JWT / cookie
│   │   └── logging           # логирование запросов и ответов (zap)
│   ├── models                # общие модели данных
│   ├── router                # HTTP-роутер (chi)
│   ├── storage               # интерфейс хранилища и реализации
│   │   ├── files             # файловое хранилище (JSON)
│   │   ├── memory            # in-memory хранилище
│   │   ├── pg                # PostgreSQL
│   │   └── storage.go        # интерфейс хранилища
│   └── workers               # фоновые задачи (асинхронное удаление)
├── proto
│   └── url_cutter            # Protobuf-спецификации для gRPC API
├── .gitignore
├── README.md
├── go.mod
├── go.sum
└── staticcheck.conf
```

### Ключевые компоненты

- **cmd/shortener** — точка входа: конфигурация, подключение к хранилищу, запуск HTTP- и gRPC-серверов.
- **internal/app** — сборка зависимостей и запуск серверов, graceful shutdown.
- **internal/config** — единый источник конфигурации: флаги, переменные окружения, JSON-файл.
- **internal/handlers** — HTTP-хендлеры, работают с хранилищем через интерфейс.
- **internal/grpc_server** — gRPC-сервер с той же функциональностью, что и HTTP.
- **internal/middleware** — `compress` (gzip), `jwtgen` (подписанные cookie / JWT), `logging` (zap).
- **internal/router** — маршруты HTTP-сервера и подключение middleware.
- **internal/storage** — интерфейс `Storage` и три реализации: файл, память, PostgreSQL.
- **internal/workers** — фоновые горутины: асинхронное удаление URL, периодическое сохранение в файл.
- **proto/url_cutter** — описание gRPC API; сгенерированный код используется в `grpc_server`.
- **cmd/staticlint** — multichecker: стандартные анализаторы Go, staticcheck и собственное правило против `os.Exit` в `main`.

### Взаимодействие слоёв

1. HTTP-запрос попадает в роутер и проходит через middleware (логирование, сжатие, аутентификация).
2. Хендлер извлекает и валидирует данные, обращается к хранилищу через интерфейс `storage.Storage`.
3. Результат возвращается через хендлер; middleware при необходимости сжимает ответ и логирует результат.
4. gRPC-запрос обрабатывается аналогично: сгенерированные сервисы вызывают ту же логику.

Принципы: зависимости передаются через конструкторы, graceful shutdown обрабатывает SIGTERM / SIGINT / SIGQUIT.

## Используемые технологии

- **Go** 1.24
- **Роутер**: [chi](https://github.com/go-chi/chi)
- **База данных**: PostgreSQL (драйвер [jackc/pgx](https://github.com/jackc/pgx)), миграции [goose](https://github.com/pressly/goose)
- **Логирование**: [go.uber.org/zap](https://go.uber.org/zap)
- **Сжатие**: `compress/gzip` и middleware
- **Конфигурация**: флаги `flag`, переменные окружения, JSON-файл
- **Аутентификация**: подписанные cookie
- **gRPC**: [google.golang.org/grpc](https://google.golang.org/grpc)
- **Статический анализ**: `golang.org/x/tools/go/analysis`, staticcheck
- **Тестирование**: стандартный пакет `testing` и `httptest`

## Конфигурация

Сервер настраивается через флаги, переменные окружения или JSON-файл (флаг `-c` / `-config`). Приоритет: флаг > переменная окружения > JSON-файл > значение по умолчанию.

| Флаг             | Переменная окружения | Описание                                                  | Значение по умолчанию     |
|------------------|----------------------|-----------------------------------------------------------|---------------------------|
| `-a`             | `SERVER_ADDRESS`     | Адрес HTTP-сервера                                        | `localhost:8080`          |
| `-b`             | `BASE_URL`           | Базовый URL для сокращённых ссылок                        | `http://localhost:8080`   |
| `-f`             | `FILE_STORAGE_PATH`  | Путь к файлу хранения (пусто — отключено)                 | `/tmp/short-url-db.json`  |
| `-d`             | `DATABASE_DSN`       | Строка подключения к PostgreSQL                           | `""` (память или файл)    |
| `-s`             | `ENABLE_HTTPS`       | Включить HTTPS                                            | `false`                   |
| `-t`             | `TRUSTED_SUBNET`     | Доверенная подсеть (CIDR) для `/api/internal/stats`       | `""` (доступ запрещён)    |
| `-c` / `-config` | `CONFIG`             | Путь к JSON-файлу конфигурации                            | `""`                      |
| `-k`             | `KEY`                | Ключ для подписи cookie                                   | `""`                      |

Пример JSON-файла конфигурации:

```json
{
    "server_address": "localhost:8080",
    "base_url": "http://localhost:8080",
    "file_storage_path": "/path/to/file.db",
    "database_dsn": "postgres://user:pass@localhost/shortener?sslmode=disable",
    "enable_https": false,
    "trusted_subnet": "192.168.1.0/24"
}
```

### Информация о сборке

При старте сервер выводит в stdout версию, дату и коммит сборки (или `N/A`). Их можно задать при компиляции:

```bash
go build -ldflags "-X main.buildVersion=v1.0.0 -X main.buildDate=$(date +'%Y/%m/%d') -X main.buildCommit=$(git rev-parse HEAD)" -o shortener cmd/shortener/main.go
```

## Установка и запуск

### Требования

- Go 1.24 или новее
- PostgreSQL — только если нужен режим с базой данных

### Локальная сборка

```bash
git clone https://github.com/sinfirst/URL-Cutter.git
cd URL-Cutter
go build -o shortener cmd/shortener/main.go
./shortener -a localhost:8080 -b http://localhost:8080
```

### Запуск с PostgreSQL

Укажите DSN через флаг `-d` или переменную `DATABASE_DSN`. Сервер выполнит миграции при старте.

```bash
./shortener -a localhost:8080 -b http://localhost:8080 -d "postgres://user:pass@localhost/shortener?sslmode=disable"
```

### Пример Dockerfile

В репозитории отдельного `Dockerfile` нет, ниже — пример сборки образа:

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o shortener cmd/shortener/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/shortener .
EXPOSE 8080
CMD ["./shortener"]
```

## Примеры использования

### Сокращение URL (text/plain)

```bash
curl -X POST http://localhost:8080/ \
     -H "Content-Type: text/plain" \
     -d "https://practicum.yandex.ru"
```

Ответ: `http://localhost:8080/abc123`

### Сокращение URL (JSON)

```bash
curl -X POST http://localhost:8080/api/shorten \
     -H "Content-Type: application/json" \
     -d '{"url": "https://practicum.yandex.ru"}'
```

Ответ: `{"result":"http://localhost:8080/abc123"}`

### Получение оригинального URL

Перейдите в браузере по `http://localhost:8080/abc123` — произойдёт редирект.

### Пакетное создание

```bash
curl -X POST http://localhost:8080/api/shorten/batch \
     -H "Content-Type: application/json" \
     -d '[{"correlation_id":"1","original_url":"https://ya.ru"},{"correlation_id":"2","original_url":"https://google.com"}]'
```

### Список ссылок пользователя

```bash
curl -X GET http://localhost:8080/api/user/urls \
     -b "cookie.txt"
```

### Удаление ссылок (асинхронное)

```bash
curl -X DELETE http://localhost:8080/api/user/urls \
     -H "Content-Type: application/json" \
     -d '["abc123", "def456"]' \
     -b "cookie.txt"
```

### Проверка соединения с БД

```bash
curl http://localhost:8080/ping
```

### Внутренняя статистика (для доверенных IP)

```bash
curl -H "X-Real-IP: 192.168.1.10" http://localhost:8080/api/internal/stats
```

## Тестирование и качество кода

```bash
go test ./... -cover
go test -bench=. ./...
go run cmd/staticlint/main.go ./...
```

Юнит-тесты и примеры лежат в `internal/app`. Для профиля памяти:

```bash
go test -memprofile=profiles/base.pprof -bench=.
```

## gRPC

Сервер также поднимает gRPC-интерфейс (порт задаётся отдельно). Операции дублируют HTTP API. Пример:

```bash
./shortener -grpc-address localhost:50051
```

Клиент можно сгенерировать из protobuf-спецификации в `proto/url_cutter`.

## Лицензия

MIT. См. файл [LICENSE](LICENSE).
