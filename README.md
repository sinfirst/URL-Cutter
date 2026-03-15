# URL Shortener Service

Сервис сокращения URL — это высокопроизводительное приложение на Go, предоставляющее удобный API для создания коротких ссылок и управления ими. Проект реализован в рамках учебного трека «Сервис сокращения URL» и включает все ключевые возможности, описанные в техническом задании: базовые эндпоинты, JSON API, поддержку сжатия, персистентность (файл, PostgreSQL), аутентификацию, асинхронное удаление, gRPC, статический анализ и многое другое.

## Возможности

- **Сокращение URL**: создание коротких ссылок через `POST /` (text/plain) и `POST /api/shorten` (JSON).
- **Редирект**: GET `/{id}` — перенаправление на оригинальный URL с кодом 307.
- **Пакетная обработка**: `POST /api/shorten/batch` для множественного создания коротких ссылок.
- **Получение всех ссылок пользователя**: `GET /api/user/urls`.
- **Асинхронное удаление**: `DELETE /api/user/urls` — мягкое удаление ссылок.
- **Проверка соединения с БД**: `GET /ping`.
- **Статистика для доверенных подсетей**: `GET /api/internal/stats` (количество URL и пользователей).
- **Сжатие запросов/ответов**: gzip для типов `application/json` и `text/html`.
- **Хранение**:
  - В памяти (in-memory)
  - В файле (JSON)
  - В PostgreSQL (с уникальным индексом на оригинальный URL)
- **Аутентификация**: пользовательские cookie с подписью.
- **Graceful shutdown**: корректное завершение с сохранением данных.
- **Поддержка gRPC**: все эндпоинты доступны по протоколу gRPC (наряду с HTTP).
- **Конфигурация**: флаги, переменные окружения, JSON-файл.
- **Версионирование бинарного файла**: вывод build version, date, commit при старте.
- **Статический анализ**: multichecker с собственными правилами (запрет `os.Exit` в `main`).
- **Тесты**: юнит-тесты с покрытием >40%, бенчмарки, примеры в `example_test.go`.

# URL Shortener Service

(Предыдущие разделы остаются без изменений, кроме обновлённой архитектуры ниже.)

## Архитектура

Проект организован по стандартной для Go многослойной архитектуре с чётким разделением ответственности. Ниже представлена актуальная структура директорий и назначение каждого компонента.

```
.
├── .github                   # конфигурации GitHub Actions (CI/CD)
├── cmd
│   ├── shortener             # точка входа HTTP- и gRPC-сервера
│   └── staticlint            # multichecker для статического анализа
├── internal
│   ├── app                   # инициализация приложения, DI, запуск сервера
│   ├── config                # работа с конфигурацией (флаги, env, JSON)
│   ├── grpc_server           # реализация gRPC-сервера и прото-сервисов
│   ├── handlers              # HTTP-обработчики (роутинг, валидация)
│   ├── middleware            # промежуточное ПО для HTTP
│   │   ├── compress          # сжатие ответов (gzip)
│   │   ├── jwtgen            # аутентификация через JWT / cookie
│   │   └── logging           # логирование запросов/ответов (zap)
│   ├── models                # общие модели данных (URL, пользователь, статистика)
│   ├── router                # инициализация HTTP-роутера (chi/mux)
│   ├── storage               # интерфейс хранилища и реализации
│   │   ├── files             # файловое хранилище (JSON)
│   │   ├── memory            # in-memory хранилище
│   │   ├── pg                # PostgreSQL хранилище
│   │   └── storage.go        # основной интерфейс хранилища
│   └── workers               # фоновые задачи (асинхронное удаление, сохранение)
├── proto
│   └── url_cutter            # Protobuf-спецификации для gRPC API
│       └── url_cutter.proto
├── .gitignore
├── README.md
├── go.mod
├── go.sum
└── staticcheck.conf          # конфигурация статического анализатора staticcheck
```

### Ключевые компоненты

- **cmd/shortener** — главная точка входа. Здесь происходит инициализация конфигурации, подключение к хранилищу, запуск HTTP- и gRPC-серверов.
- **internal/app** — собирает все зависимости и запускает серверы, обеспечивая graceful shutdown.
- **internal/config** — единый источник конфигурации: флаги, переменные окружения, JSON-файл. Значения нормализуются и предоставляются остальным пакетам.
- **internal/handlers** — HTTP-хендлеры, реализующие бизнес-логику эндпоинтов. Они вызывают методы сервисного слоя (который может находиться в `internal/app` или отдельно), работают с хранилищем через интерфейс.
- **internal/grpc_server** — gRPC-сервер, который использует те же сервисы, что и HTTP, обеспечивая идентичную функциональность по протоколу gRPC.
- **internal/middleware** — набор middleware для HTTP:
  - `compress` — сжатие ответов (gzip) для клиентов, поддерживающих сжатие.
  - `jwtgen` — аутентификация пользователей с помощью подписанных JWT (или cookie).
  - `logging` — логирование деталей запросов и ответов с использованием zap.
- **internal/router** — конфигурация маршрутов HTTP-сервера, подключение middleware.
- **internal/storage** — абстракция хранилища. Определяет интерфейс `Storage`, который реализуют три варианта:
  - `files` — хранение в JSON-файле (используется при отсутствии БД).
  - `memory` — хранение в оперативной памяти (для тестов и режима без сохранения).
  - `pg` — PostgreSQL с поддержкой уникальных индексов, batch-операций и транзакций.
- **internal/workers** — фоновые горутины: асинхронное удаление URL (помечаются как удалённые), периодическое сохранение в файл (если используется файловое хранилище с интервалом).
- **proto/url_cutter** — описание gRPC API в формате Protobuf. Сгенерированный код используется в `grpc_server`.
- **cmd/staticlint** — multichecker, объединяющий стандартные анализаторы Go, анализаторы staticcheck (все SA и один из других классов) и собственный анализатор, запрещающий прямой вызов `os.Exit` в функции `main` пакета `main`.
- **staticcheck.conf** — файл конфигурации для статического анализатора staticcheck (например, отключение некоторых проверок).

### Взаимодействие слоёв

1. **HTTP-запрос** попадает в роутер, проходит через цепочку middleware (логирование, сжатие, аутентификация).
2. **Хендлер** извлекает данные запроса, валидирует их и вызывает соответствующий метод **сервиса** (бизнес-логика). Сервис может находиться внутри `handlers` или в отдельном пакете `internal/app/service`.
3. **Сервис** обращается к **хранилищу** через интерфейс `storage.Storage` (in-memory, файл или PostgreSQL).
4. Результат возвращается обратно через хендлер, middleware формирует ответ (возможно, сжатый) и логирует результат.
5. **gRPC-запрос** обрабатывается аналогично, но вместо HTTP-хендлеров используются сгенерированные gRPC-серверы, которые также вызывают общие сервисы.

### Принципы

- **Интерфейсы** определяются там, где они используются (в `storage` — для хранилища).
- **Зависимости** передаются через конструкторы (явное DI), что упрощает тестирование.
- **Graceful shutdown** обрабатывает сигналы SIGTERM, SIGINT, SIGQUIT, завершая все активные запросы и сохраняя данные.
- **Тестирование** построено на использовании моков (например, через `gomock`) для хранилища и внешних зависимостей.

## Используемые технологии

- **Go** 1.21+
- **Роутер**: [chi](https://github.com/go-chi/chi) / [gorilla/mux](https://github.com/gorilla/mux) (на выбор)
- **База данных**: PostgreSQL 10+ (драйвер [jackc/pgx](https://github.com/jackc/pgx))
- **Логирование**: [go.uber.org/zap](https://go.uber.org/zap)
- **Сжатие**: встроенный `compress/gzip` + middleware
- **Конфигурация**: флаги `flag`, переменные окружения, JSON-файл
- **Аутентификация**: подписанные cookie (HMAC-SHA256)
- **gRPC**: [google.golang.org/grpc](https://google.golang.org/grpc)
- **Статический анализ**: `golang.org/x/tools/go/analysis`, `staticcheck.io`
- **Тестирование**: `testing`, `testify`, `gomock` (при необходимости)

## Конфигурация

Сервер настраивается через флаги командной строки, переменные окружения или JSON-файл (флаг `-c/-config`). Приоритет: флаг > переменная окружения > JSON-файл > значение по умолчанию.

### Основные параметры

| Флаг           | Переменная окружения  | Описание                                             | Значение по умолчанию        |
|----------------|------------------------|------------------------------------------------------|------------------------------|
| `-a`           | `SERVER_ADDRESS`       | Адрес HTTP-сервера                                   | `localhost:8080`             |
| `-b`           | `BASE_URL`             | Базовый URL для сокращённых ссылок                   | `http://localhost:8080`      |
| `-f`           | `FILE_STORAGE_PATH`    | Путь к файлу для хранения данных (пусто = отключено) | `/tmp/short-url-db.json`     |
| `-d`           | `DATABASE_DSN`         | Строка подключения к PostgreSQL                      | `""` (in-memory/file)        |
| `-s`           | `ENABLE_HTTPS`         | Включить HTTPS (требует сертификаты)                 | `false`                      |
| `-t`           | `TRUSTED_SUBNET`       | Доверенная подсеть (CIDR) для эндпоинта `/internal/stats` | `""` (доступ запрещён)    |
| `-c` / `-config` | `CONFIG`             | Путь к JSON-файлу конфигурации                        | `""`                         |
| `-k`           | `KEY`                  | Ключ для подписи cookie                               | `""`                         |

Пример JSON-файла конфигурации (`config.json`):
```json
{
    "server_address": "localhost:8080",
    "base_url": "http://localhost",
    "file_storage_path": "/path/to/file.db",
    "database_dsn": "postgres://user:pass@localhost/shortener?sslmode=disable",
    "enable_https": true,
    "trusted_subnet": "192.168.1.0/24"
}
```

### Информация о сборке

При старте сервер выводит в stdout:
```
Build version: <buildVersion> (или "N/A")
Build date: <buildDate> (или "N/A")
Build commit: <buildCommit> (или "N/A")
```
Переменные задаются при компиляции, например:
```bash
go build -ldflags "-X main.buildVersion=v1.0.0 -X main.buildDate=$(date +'%Y/%m/%d') -X main.buildCommit=$(git rev-parse HEAD)" -o shortener cmd/shortener/main.go
```

## Установка и запуск

### Локальная сборка

1. Клонируйте репозиторий:
   ```bash
   git clone https://github.com/username/shortener.git
   cd shortener
   ```
2. Соберите бинарный файл:
   ```bash
   go build -o shortener cmd/shortener/main.go
   ```
3. Запустите сервер:
   ```bash
   ./shortener -a localhost:8080 -b http://localhost:8080
   ```

### Docker

Можно использовать готовый Dockerfile (пример):
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o shortener cmd/shortener/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/shortener .
EXPOSE 8080
CMD ["./shortener"]
```

Сборка и запуск:
```bash
docker build -t shortener .
docker run -p 8080:8080 shortener
```

### Запуск с PostgreSQL

Укажите DSN через флаг `-d` или переменную `DATABASE_DSN`. Сервер автоматически выполнит миграции (создаст таблицы при необходимости).

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
Ответ:
```json
[
    {"correlation_id":"1","short_url":"http://localhost:8080/abc123"},
    {"correlation_id":"2","short_url":"http://localhost:8080/def456"}
]
```

### Список ссылок пользователя

```bash
curl -X GET http://localhost:8080/api/user/urls \
     -b "cookie.txt"  # после аутентификации
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

- **Юнит-тесты**: `go test ./... -cover`
- **Бенчмарки**: `go test -bench=. ./...`
- **Статический анализ**:
  ```bash
  go run cmd/staticlint/main.go ./...
  ```
- **Профилирование**:
  - Сохранение профиля памяти: `go test -memprofile=profiles/base.pprof -bench=.`
  - Сравнение профилей: `pprof -top -diff_base=profiles/base.pprof profiles/result.pprof`

Проект покрыт тестами не менее чем на 40%. Для критических компонентов написаны примеры в `example_test.go`.

## gRPC

Сервер также предоставляет gRPC-интерфейс (порт можно задать отдельно). Все хендлеры дублируют функциональность HTTP. Пример запуска gRPC-сервера (если реализовано):

```bash
./shortener -grpc-address localhost:50051
```

Клиент для gRPC может быть сгенерирован из protobuf-спецификации.
