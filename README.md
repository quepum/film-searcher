# Film Searcher

Микросервис на Go: синхронизирует фильмы с внешнего стенда в PostgreSQL, отдаёт их по HTTP/gRPC,
ищет по словам из названия и режиссёра через обратный индекс Redis и кэширует ответы поиска.

## Запуск

### 1. PostgreSQL и Redis

```bash
docker compose up -d
```

Поднимаются:
- PostgreSQL `localhost:5432`
- Redis `localhost:6379`

### 2. Переменные окружения

| Переменная        | Обязательна | По умолчанию                | Назначение                          |
|-------------------|-------------|-----------------------------|-------------------------------------|
| `MOVIE_BASE_URL`  | нет         | `https://homeworksite.site` | базовый URL фильм-стенда            |
| `HTTP_ADDR`       | нет         | `:8080`                     | адрес HTTP-сервера                  |
| `GRPC_ADDR`       | нет         | `:9090`                     | адрес gRPC-сервера                  |
| `DATABASE_URL`    | да          | —                           | DSN PostgreSQL                      |
| `REDIS_ADDR`      | да          | —                           | адрес Redis в формате `host:port`   |

```bash
export DATABASE_URL='postgres://postgres:postgres@127.0.0.1:5432/movies?sslmode=disable'
export REDIS_ADDR='127.0.0.1:6379'
export MOVIE_BASE_URL='https://homeworksite.site'
export HTTP_ADDR='127.0.0.1:8080'
export GRPC_ADDR='127.0.0.1:9090'
go run ./cmd/server
```