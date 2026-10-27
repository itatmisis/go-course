# Лекция 8. Deploy 2: Docker и Docker Compose

На прошлой лекции мы запускали сервер руками: ставили `go` на сервер, собирали бинарник, прятали его в `screen`.
Это работает, но плохо повторяется: на другой машине другая версия `go`, другой Postgres, забыли переменную
окружения — и всё сломалось. Сегодня разберём, как упаковать приложение так, чтобы оно одинаково запускалось
везде.

## 1. Docker

![image](https://upload.wikimedia.org/wikipedia/commons/4/4e/Docker_%28container_engine%29_logo.svg?uselang=ru)

Docker — система для автоматизации развёртывания и управления контейнерами.
Позволяет «упаковать» приложение со всем его окружением и зависимостями в контейнер.

### 1.1. Прелести докера

**Декларативность описания.** Используя `Dockerfile` и `docker-compose.yaml`, можно однозначно описать, как
собирать и запускать проект.

**Универсальность.** Если у вас установлен докер, вы можете запустить что угодно на чём угодно.

**Изолированность.** Контейнеры по умолчанию не могут влиять на вашу «материнскую» систему и не видят друг друга.

### 1.2. Образ и контейнер

- **Образ (image)** — «слепок» файловой системы с приложением и всем, что ему нужно. Собирается по `Dockerfile`
  или скачивается с [Docker Hub](https://hub.docker.com/). Образ неизменяем.
- **Контейнер (container)** — запущенный экземпляр образа. Из одного образа можно запустить сколько угодно
  контейнеров.

Аналогия из `go`: образ — это `struct`-тип, контейнер — переменная этого типа.

### 1.3. Основные команды

```bash
docker build -t app . # собрать образ с именем app из Dockerfile в текущей директории
docker images # список образов
docker run --name=app -p 8080:8080 app # запустить контейнер app из образа app
docker run -d ... # то же самое, но в фоне (detached)
docker ps # список запущенных контейнеров
docker ps -a # список всех контейнеров, включая остановленные
docker logs -f app # логи контейнера app
docker exec -it app sh # зайти внутрь контейнера (если в образе есть shell)
docker stop app # остановить контейнер
docker rm -f app # удалить контейнер (даже запущенный)
```

`-p 8080:8080` — пробросить порт: `<порт на хосте>:<порт в контейнере>`. Без этого до сервера в контейнере
снаружи не достучаться.

## 2. Проект: сокращатель ссылок

Рабочий пример лежит в папке [project](project). Это сервис коротких ссылок:

- `POST /link` с телом `{"link": "https://go.dev"}` — создаёт короткую ссылку и возвращает её:
  `{"link": "http://localhost:8080/aB3xZ"}`;
- `GET /aB3xZ` — редиректит (`301`) на исходную ссылку или отдаёт `404`, если такой нет.

### 2.1. Структура

Проект разбит на слои (луковая архитектура): зависимости направлены внутрь, от транспорта к данным.

```tree
project
├── main.go                  # сборка всех слоёв и запуск сервера
├── internal
│   ├── settings             # конфигурация из переменных окружения
│   ├── apis
│   │   └── httpapi          # HTTP-хендлеры на echo
│   ├── services
│   │   └── linkservice      # бизнес-логика
│   ├── repos
│   │   └── linkrepo         # работа с Postgres через pgx
│   └── utils
│       └── randutil         # генерация случайных строк
├── Dockerfile
├── docker-compose.yaml
└── Makefile
```

- `apis` — интерфейсы взаимодействия с проектом: HTTP API, чтение из очереди и т. д. Здесь только разбор
  запроса и формирование ответа.
- `services` — бизнес-логика, то есть всё, что связано с логикой самого продукта. Подключение к базе данных — не
  бизнес-логика, а генерация короткой ссылки — бизнес-логика.
- `repos` — работа с базой данных. Здесь не должно быть бизнес-логики, только простые запросы: прочитать,
  добавить, удалить.
- `utils` — утилиты, которые не относятся ни к одному слою.

Папка `internal` — особая для `go`: пакеты из неё нельзя импортировать из других модулей.

### 2.2. Конфигурация через переменные окружения

В докере конфигурацию принято передавать через переменные окружения. Читаем их библиотекой
[caarlos0/env](https://github.com/caarlos0/env):

```go
type Postgres struct {
	Database string `env:"DATABASE" envDefault:"db_main"`
	Host     string `env:"HOST"     envDefault:"localhost"`
	Port     int    `env:"PORT"     envDefault:"5432"`
	Username string `env:"USERNAME" envDefault:"db_main"`
	Password string `env:"PASSWORD" envDefault:"db_main"`
}

type Settings struct {
	Postgres Postgres `envPrefix:"PG__"`
	HTTP     HTTP     `envPrefix:"HTTP__"`
}

func Load() (Settings, error) {
	appSettings, err := env.ParseAsWithOptions[Settings](env.Options{Prefix: "APP_"})
	if err != nil {
		return Settings{}, fmt.Errorf("parse env: %w", err)
	}

	return appSettings, nil
}
```

Префиксы складываются: поле `Host` в `Postgres` читается из переменной `APP_PG__HOST`. Если переменной нет,
берётся `envDefault`. Поэтому локально всё работает без настройки, а в докере мы переопределяем только хост базы.

### 2.3. Graceful shutdown

Когда докер останавливает контейнер, он посылает процессу сигнал `SIGTERM` и ждёт 10 секунд, после чего убивает
его. Хорошее приложение должно за это время доделать текущие запросы и закрыть соединения с БД:

```go
ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
defer stop()

go func() {
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()

	shutdownErr := server.Shutdown(shutdownCtx)
	if shutdownErr != nil {
		logger.ErrorContext(ctx, "server.shutdown.failed", "error", shutdownErr)
	}
}()
```

## 3. Dockerfile

`Dockerfile` — инструкция, как собрать образ. Каждая команда создаёт новый слой образа.

```dockerfile
FROM golang:1.25 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /bin/app .

FROM gcr.io/distroless/static-debian12

COPY --from=builder /bin/app /app

EXPOSE 8080

ENTRYPOINT ["/app"]
```

| Команда | Что делает |
|---|---|
| `FROM <image>` | Базовый образ. По умолчанию берётся с [Docker Hub](https://hub.docker.com/) |
| `WORKDIR <dir>` | Сменить рабочую директорию внутри образа. Если её нет — создать |
| `COPY <src> <dst>` | Скопировать файлы из материнской системы в образ |
| `RUN <command>` | Выполнить команду во время сборки образа |
| `EXPOSE <port>` | Задокументировать, какой порт слушает приложение (сам порт не пробрасывает) |
| `ENTRYPOINT [...]` | Команда, которая выполнится при запуске контейнера |

### 3.1. Multi-stage сборка

В `Dockerfile` два `FROM` — это multi-stage сборка:

1. **Стадия `builder`** — образ `golang:1.25` (почти 1 ГБ: компилятор, стандартная библиотека, git). В нём
   собираем бинарник.
2. **Финальная стадия** — образ `distroless/static` (около 2 МБ: нет даже shell). В него копируем только
   готовый бинарник через `COPY --from=builder`.

В итоге образ весит пару десятков мегабайт вместо гигабайта, а в контейнере нет ничего лишнего, что можно
взломать. Это возможно благодаря тому, что `go` собирает один статический бинарник. `CGO_ENABLED=0` нужен, чтобы
бинарник не зависел от системной `libc`, которой в `distroless/static` нет.

### 3.2. Кэширование слоёв

Обратите внимание на порядок: сначала копируем только `go.mod` и `go.sum` и скачиваем зависимости, а уже потом
копируем весь код. Докер кэширует слои и пересобирает только те, у которых поменялись входные файлы. Код меняется
часто, зависимости — редко, поэтому при обычной правке кода `go mod download` не выполняется заново.

### 3.3. `.dockerignore`

Аналог `.gitignore` для `COPY . .`: перечисленные файлы не попадут в образ. Сюда стоит добавить `.data`, `.git`,
`.env` и собранные бинарники.

## 4. Docker Compose

Наше приложение не работает без Postgres. Запускать каждый контейнер руками через `docker run` неудобно, поэтому
используем `docker compose` — он поднимает сразу несколько контейнеров по описанию в `docker-compose.yaml`:

```yaml
services:
  backend:
    restart: always
    build:
      context: .
      dockerfile: Dockerfile
    environment:
      APP_PG__HOST: postgres
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy

  postgres:
    image: postgres:17.0
    environment:
      POSTGRES_USER: db_main
      POSTGRES_PASSWORD: db_main
      PGUSER: db_main
      PGPASSWORD: db_main
      PGDATA: /var/lib/postgresql/data/pgdata
    ports:
      - "5432:5432"
    volumes:
      - ".data/pg/:/var/lib/postgresql/data/"
    healthcheck:
      test: ["CMD", "pg_isready"]
      interval: 2s
      timeout: 2s
      retries: 15
```

| Ключ | Что делает |
|---|---|
| `services` | Список контейнеров. Имя сервиса — это и его хостнейм внутри сети compose |
| `build` | Собрать образ из `Dockerfile` |
| `image` | Скачать готовый образ с Docker Hub |
| `environment` | Переменные окружения контейнера |
| `ports` | Пробросить порты в материнскую систему |
| `volumes` | Подключить _папку_ из материнской системы в контейнер. Без этого данные БД пропадут вместе с контейнером |
| `restart: always` | Перезапускать контейнер, если он упал |
| `depends_on` | Порядок запуска: `backend` стартует после `postgres` |
| `healthcheck` | Как проверить, что сервис готов. `condition: service_healthy` ждёт, пока Postgres начнёт принимать соединения |

Обратите внимание на `APP_PG__HOST: postgres`. Внутри сети compose контейнеры видят друг друга по имени сервиса,
поэтому из `backend` база доступна по адресу `postgres:5432`, а не `localhost:5432`: `localhost` внутри
контейнера — это сам контейнер.

Команды:

```bash
docker compose up --build # собрать и запустить все сервисы
docker compose up -d # то же самое, но в фоне
docker compose up postgres # запустить только postgres
docker compose logs -f backend # логи сервиса backend
docker compose down # остановить и удалить контейнеры
```

## 5. Makefile

```make
docker-build:
	docker build -t app .

docker-rerun:
	docker rm -f app || true
	docker run --name=app -p 8080:8080 -e APP_PG__HOST=host.docker.internal app

run-infra:
	docker compose up postgres

run-all:
	docker compose up --build

install:
	go mod download

run:
	go run .
```

Типичный сценарий разработки: `make run-infra` в одном терминале поднимает Postgres в докере, `make run` в другом
запускает приложение локально. Перед деплоем — `make run-all`, чтобы проверить, что всё работает в контейнерах.

Проверяем:

```bash
curl -X POST localhost:8080/link -H 'Content-Type: application/json' -d '{"link": "https://go.dev"}'
curl -i localhost:8080/<short-link>
```

## 6. Деплой на сервер

С докером деплой сводится к трём шагам:

```bash
# Установить докер на сервер: https://docs.docker.com/engine/install/ubuntu/

# Скопировать проект на сервер (или сделать git clone)
git clone <repo-url> && cd <repo>

# Поднять всё в фоне
docker compose up -d --build
```

Больше не нужен ни `screen`, ни установка `go` на сервер: `restart: always` перезапустит упавший контейнер, а
версии `go` и Postgres зафиксированы в `Dockerfile` и `docker-compose.yaml`.

## Ссылки

- [Установка докера](https://docs.docker.com/engine/install/)
- [Официальный гайд: Docker для `go`](https://docs.docker.com/guides/golang/)
- [Справочник по Dockerfile](https://docs.docker.com/reference/dockerfile/)
- [Справочник по docker-compose.yaml](https://docs.docker.com/reference/compose-file/)
