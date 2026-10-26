# Лекция 6. Работа с PostgreSQL из Go и проектирование схемы БД

## 1. ORM и прямые SQL-запросы

### 1.1. Что такое ORM

**ORM (Object–Relational Mapping)** — подход, который связывает объекты (или структуры) языка программирования с таблицами реляционной базы данных. В Python часто используют SQLAlchemy; в Go существуют GORM и другие библиотеки.

Идея ORM — позволить описывать многие операции с данными на уровне моделей и методов библиотеки, скрыв часть SQL. При этом реальная БД продолжает выполнять SQL-запросы.

| Работа напрямую с SQL                                       | ORM                                                       |
| ----------------------------------------------------------- | --------------------------------------------------------- |
| Явно пишем `SELECT ... FROM tasks WHERE id = $1`            | Вызываем ORM для выборки модели `Task`                    |
| SQL-запрос легко скопировать в `psql` / консоль             | Необходимо понимать, какой SQL сгенерирует ORM            |
| Больше ручного кода для сканирования результатов            | Меньше повторяющегося кода для простых операций           |
| Полный контроль над `JOIN`, `RETURNING`, сложными запросами | Удобное API, но сложные запросы иногда требуют сырого SQL |
|                                                             |                                                           |
Еще существуют sql-builder'ы и они довольно распространены в го, но..... часто команды пишут их под себя, поэтому смысла рассматривать их я не вижу, так как они будут слишком сильно отличаться

**В нашем курсе ORM не используем**

### 1.2. Как SQL выглядит в `pgx`


```go
var title string

err := pool.QueryRow(
	ctx,
	`SELECT title FROM tasks WHERE id = $1`,
	taskID,
).Scan(&title)
```

`$1` — место для первого параметра. `pgx` передаёт аргумент отдельно от текста SQL. **Не собирайте SQL из пользовательского ввода конкатенацией строк:** это может привести к SQL-инъекции.

---

## 2. Подключение к PostgreSQL из Go

Используем драйвер `pgx`

### 2.1. Отдельное соединение и пул соединений

`pgx.Conn` — одно соединение с PostgreSQL. Одно и то же соединение **нельзя одновременно использовать из нескольких горутин**.

`pgxpool.Pool` — пул соединений. Он самостоятельно выдаёт свободное соединение обработчику и возвращает его в пул после завершения операции. 

Схема работы:

```text
HTTP GET /tasks/1 ----\
HTTP GET /tasks/2 -----+--> Echo handlers --> pgxpool --> PostgreSQL
HTTP POST /tasks -----/
```

Устанавливаем зависимости в папке проекта:

```bash
go get github.com/jackc/pgx/v5/pgxpool
```


Переменные окружения можно задать через yaml или json файлик и распарсить его, пока будем просто выставлять переменную в ОС:

```bash
export DATABASE_URL='postgres://tracker:tracker@localhost:5432/tracker?sslmode=disable'
```

### 2.2. Создаём пул и проверяем соединение

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	var first, second int
	err = pool.Query(ctx, `SELECT 1, 2`).Scan(&first, &second)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(first, second) // 1, 2
}
```

**Что здесь происходит:**

- `context.WithTimeout` ограничивает время на начальные операции с БД.
- `pgxpool.New` создаёт объект пула; **`Ping`** отдельно проверяет доступность БД.
- `QueryRow` выполняет запрос, ожидающий одну строку результата.
- `Scan` переносит значения столбцов в Go-переменные в том же порядке.
- `defer pool.Close()` освобождает ресурсы при завершении программы.

### 2.3. Основные методы `pgxpool`

| Метод | Для чего используем |
|---|---|
| `QueryRow(ctx, sql, args...).Scan(...)` | Одна строка: `SELECT` по ID, `INSERT ... RETURNING`, `UPDATE ... RETURNING` |
| `Query(ctx, sql, args...)` | Несколько строк: список задач |
| `Exec(ctx, sql, args...)` | Запрос, когда значения строк результата не нужны |
| `Ping(ctx)` | Проверить соединение с БД |
| `Begin(ctx)` | Начать транзакцию |

При работе с `Query` всегда закрывайте `rows` и проверяйте `rows.Err()` после цикла `rows.Next()`.

---

## 3. Проектирование таблиц

### 3.1. Имена таблиц и столбцов

Рекомендации:

- `snake_case` для таблиц и столбцов (`created_at`, `assignee_id`);
- один стиль именования таблиц **во всём проекте** — у нас множественное число: `users`, `tasks`;
- понятные имена (`tasks` вместо `t1`);
- `id` для первичного ключа, `<сущность>_id` для ссылки на другую сущность (`assignee_id` указывает на `users.id`).

**Единственное и множественное число — вопрос соглашения**, не ограничение PostgreSQL, лучше следовать стилю какой вам нравится либо какой уже принят в вашей команде

НО `user` и `order` могут приводить к путанице и необходимости кавычек в отдельных запросах, если они написаны в единственном числе

### 3.2.  BIGINT vs UUID

| `BIGINT GENERATED ... AS IDENTITY` | `UUID`                                                                    |
| ---------------------------------- | ------------------------------------------------------------------------- |
| Компактный и удобный для обучения  | Можно генерировать вне БД, полезно для некоторых распределённых сценариев |
| Значение обычно назначает база     | Не раскрывает простой последовательный номер записи                       |
| Последовательные идентификаторы    | Индексы и характер вставок зависят от версии и способа генерации UUID     |

## 4. Служебные поля: `created_at` и `updated_at`

Для записей часто нужны поля:

```sql
created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
```

**`TIMESTAMPTZ`** — тип PostgreSQL для момента времени; `now()` даёт время начала текущей транзакции. Использование `TIMESTAMP WITH TIME ZONE` не означает, что PostgreSQL сохраняет исходное название часового пояса.

Важный нюанс: `DEFAULT now()` действует **при вставке строки**, но **не обновляет** `updated_at` при каждом `UPDATE`.

Поэтому:

```sql
UPDATE tasks
SET status = $2,
    updated_at = now()
WHERE id = $1;
```


---

## 5. Soft Delete

В обычном варианте `DELETE` безвозвратно удаляет строку из таблицы (если нет специальных механизмов восстановления):

```sql
DELETE FROM tasks WHERE id = $1;
```

Для Task Tracker иногда разумнее использовать **soft delete** — пометить задачу удалённой, но оставить в БД.

Добавляем столбец `deleted_at`:

```sql
ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
```

**Смысл значения:**

- `deleted_at IS NULL` — задача активна;
- `deleted_at IS NOT NULL` — задача скрыта (удалена логически).

Удаляем задачу логически:

```sql
UPDATE tasks
SET deleted_at = now(),
    updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;
```

Выбираем только активные задачи:

```sql
SELECT id, title, status
FROM tasks
WHERE deleted_at IS NULL
ORDER BY created_at DESC, id DESC;
```

При необходимости восстанавливаем задачу:

```sql
UPDATE tasks
SET deleted_at = NULL,
    updated_at = now()
WHERE id = $1 AND deleted_at IS NOT NULL;
```

### 6. Миграция

обычно есть папочка `/migrations` где лежат все ваши миграции за всю историю проекта, в строгой последовательности

```sql
ALTER TABLE tasks
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_tasks_active_status_created_at
    ON tasks (status, created_at DESC, id DESC)
    WHERE deleted_at IS NULL;
```

Запускаем миграцию:

```bash
docker exec -i task-tracker-db psql -v ON_ERROR_STOP=1 -U tracker -d tracker \
  < migrations/002_soft_delete.sql
```

Индекс здесь **частичный** (`partial index`): содержит только активные задачи, потому что наши ручки по умолчанию работают с ними. 