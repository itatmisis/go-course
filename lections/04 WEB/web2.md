# Лекция 4. Тестирование HTTP API, middleware, логирование и аутентификация в Go + Echo

## 1. Как тестировать HTTP API

В Go для HTTP-тестов обычно хватает стандартных пакетов `testing` и `net/http/httptest`
- `httptest.NewRequest` создаёт HTTP-запрос в памяти.
- `httptest.NewRecorder` записывает статус, заголовки и тело ответа.
- `e.ServeHTTP(rec, req)` передаёт запрос роутеру Echo **без открытия порта**.
- `t.Fatalf` завершает тест с сообщением об ошибке.

**Файл `main.go`:**

```go
package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func newRouter() *echo.Echo {
	e := echo.New()
	e.GET("/hello", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "world")
	})

	return e
}

func main() {
	e := newRouter()
	e.Logger.Fatal(e.Start(":1323"))
}
```

**Файл `main_test.go`** (рядом с `main.go`):

```go
package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHello(t *testing.T) {
	e := newRouter()
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}

	if got, want := rec.Body.String(), "\"world\"\n"; got != want {
		t.Fatalf("body: got %q, want %q", got, want)
	}
}
```

Запуск:

```bash
go test -v ./...
go run .
curl -i http://localhost:1323/hello
```

 `c.JSON` сериализует строку как JSON, добавляя кавычки и завершающий перевод строки. Если нужен простой текст без кавычек, используйте `c.String(http.StatusOK, "world")`

## 2. Повторим способы получения данных из запроса

### 2.1. Заголовки (Headers)

Заголовки передают метаданные запроса. Пример: `User-Name: Olya`.

**Фрагмент для обработчика и маршрута** (внутри функции `newRouter` вместо предыдущего `/hello`):

```go
	e.GET("/hello", func(c echo.Context) error {
		userName := c.Request().Header.Get("User-Name")
		if userName == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "User-Name header is required")
		}

		return c.JSON(http.StatusOK, "Hello "+userName)
	})
```

Нужно добавить импорт `net/http`, если его ещё нет. Заголовки HTTP регистронезависимы: `user-name` и `User-Name` эквивалентны. Отсутствующий заголовок Echo **не валидирует автоматически**

```bash
curl -i -H 'User-Name: Olya' http://localhost:1323/hello
```

Ожидаемое JSON-тело: `"Hello Olya"`.

**Тест заголовка** (можно добавить в `main_test.go`):

```go
func TestHelloHeader(t *testing.T) {
	e := newRouter()
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	req.Header.Set("User-Name", "Olya")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d", rec.Code)
	}
}
```

### 2.2. Query-параметры

Пример URL: `/hello?user_name=itam&marks=4&marks=5&marks=3`.

```go
	e.GET("/hello", func(c echo.Context) error {
		userName := c.QueryParam("user_name")
		marks := c.QueryParams()["marks"]

		if userName == "" || len(marks) == 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "user_name and marks are required")
		}

		message := fmt.Sprintf(
			"Hello %s, your marks are: %s",
			userName,
			strings.Join(marks, ","),
		)
		return c.JSON(http.StatusOK, message)
	})
```

Для этого фрагмента добавьте импорты `fmt` и `strings`. Обратите внимание: `c.QueryParam("marks")` вернёт одно значение, а `c.QueryParams()["marks"]` — `[]string` со **всеми** значениями этого ключа.

```bash
curl -G http://localhost:1323/hello \
  --data-urlencode 'user_name=itam' \
  --data-urlencode 'marks=4' \
  --data-urlencode 'marks=5' \
  --data-urlencode 'marks=3'
```

Результат — JSON-строка `"Hello itam, your marks are: 4,5,3"`.

### 2.3. Path-параметры

Маршрут содержит переменную часть пути; в Echo она объявляется через двоеточие

```
	e.GET("/users/:userId/hello", func(c echo.Context) error {
		userID := c.Param("userId")
		return c.JSON(http.StatusOK, "Hello "+userID)
	})
```

```bash
curl -i http://localhost:1323/users/123/hello
```

В Go переменная `userID` произвольна, а имя внутри `c.Param("userId")` должно совпадать с именем в маршруте

### 2.4. Тело запроса (Body)

Для JSON-тела Echo использует метод `c.Bind`, заполняющий Go-структуру. Это **декодирование**, а не полная проверка бизнес-правил.

```go
type CreateUserRequest struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// Внутри newRouter():
e.POST("/users", func(c echo.Context) error {
	var input CreateUserRequest
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON")
	}
	if input.Name == "" || input.Age <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid name or age")
	}

	return c.JSON(http.StatusCreated, input)
})
```

```bash
curl -i -X POST http://localhost:1323/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"Olya","age":20}'
```

## 3. Логирование: `log/slog`

В Go с версии 1.21 в стандартной библиотеке есть структурированный логгер `log/slog`. Он поддерживает уровни `Debug`, `Info`, `Warn` и `Error`, а также поля `key=value`.

**Самостоятельный файл `main.go`:**

```go
package main

import (
	"errors"
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	logger.Debug("незначительное событие", "step", 1)
	logger.Info("важное событие", "user_id", 42)
	logger.Warn("необычное поведение", "attempts", 3)
	logger.Error("ошибка запроса", "status", 500)

	err := errors.New("ошибка подключения")
	logger.Error("не удалось получить данные", "error", err)
}
```

```bash
go run .
```


В Go ошибки — это значения типа `error`; мы чаще возвращаем их и явно передаём в логгер, а не перехватываем исключения. В большинстве слоёв приложения **не нужно одновременно логировать и возвращать** одну и ту же ошибку: возникнут дубли в логах.

## 4. Middleware: код до и после обработчика

Middleware оборачивает обработчик: может измерять время, проверять заголовки, логировать, ограничивать доступ. В Echo тип middleware — `echo.MiddlewareFunc`: функция принимает следующий `echo.HandlerFunc` и возвращает новый.

```go
package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
)

func processTimeMiddleware(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)
			elapsed := time.Since(start)

			logger.Info("request finished",
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
				"duration", elapsed,
			)

			return err
		}
	}
}

func newRouter() *echo.Echo {
	e := echo.New()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	e.Use(processTimeMiddleware(logger))

	e.GET("/hello", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "world")
	})

	return e
}

func main() {
	e := newRouter()
	e.Logger.Fatal(e.Start(":1323"))
}
```

**Проверяем middleware через `httptest`** (файл `main_test.go`):

```go
package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHelloWithMiddleware(t *testing.T) {
	e := newRouter()
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}
}
```

Для готовых middleware у Echo есть пакет `github.com/labstack/echo/v4/middleware` (например, `Recover`, `RequestID`, `Logger`).

## 6. Проверка API-ключа

```go
package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func apiKeyMiddleware(expected string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Header.Get("X-Api-Key") != expected {
				return c.JSON(http.StatusForbidden, map[string]string{
					"detail": "wrong api key",
				})
			}

			return next(c)
		}
	}
}

func newRouter() *echo.Echo {
	e := echo.New()
	e.GET("/hello", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "world")
	}, apiKeyMiddleware("42"))

	return e
}

func main() {
	e := newRouter()
	e.Logger.Fatal(e.Start(":1323"))
}
```

```bash
curl -i -H 'X-Api-Key: 10' http://localhost:1323/hello
curl -i -H 'X-Api-Key: 42' http://localhost:1323/hello
```

Ожидаем соответственно `403` и `200`. **Строка `42` — только учебный ключ**; реальный секрет нельзя жёстко зашивать в исходники. Используйте окружение / секрет-хранилище, HTTPS и механизм смены ключей.


```go
package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIKey(t *testing.T) {
	cases := []struct {
		name       string
		key        string
		wantStatus int
	}{
		{name: "missing", key: "", wantStatus: http.StatusForbidden},
		{name: "incorrect", key: "10", wantStatus: http.StatusForbidden},
		{name: "correct", key: "42", wantStatus: http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/hello", nil)
			req.Header.Set("X-Api-Key", tc.key)
			rec := httptest.NewRecorder()

			newRouter().ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("got %d, want %d", rec.Code, tc.wantStatus)
			}
		})
	}
}
```

В Echo middleware можно навесить на одну ручку, группу роутов (`group.Use(...)`) либо всё приложение (`e.Use(...)`).

## 7. JWT: проверяем подпись и достаём данные пользователя

JWT (JSON Web Token) — компактный токен из трёх частей, разделённых точками: `header.payload.signature`. **Payload закодирован, но не зашифрован.** При использовании HMAC (HS256) подпись обеспечивает целостность и подтверждает знание секрета, если проверять её корректно.


```go
package main

import (
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func jwtMiddleware(secret []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			raw := c.Request().Header.Get("X-Id-Token")
			if raw == "" {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"detail": "X-Id-Token header is required",
				})
			}
			
			claims := new(Claims)
			token, err := jwt.ParseWithClaims(
				raw,
				claims,
				func(token *jwt.Token) (any, error) {
					return secret, nil
				},
				jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
				jwt.WithExpirationRequired(),
			)
			if err != nil || !token.Valid || claims.Username == "" {
				return c.JSON(http.StatusForbidden, map[string]string{
					"detail": "bad token",
				})
			}

			c.Set("username", claims.Username)
			return next(c)
		}
	}
}

func newRouter(secret []byte) *echo.Echo {
	e := echo.New()
	e.GET("/hello", func(c echo.Context) error {
		username, ok := c.Get("username").(string)
		if !ok {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		return c.JSON(http.StatusOK, "Hello "+username)
	}, jwtMiddleware(secret))

	return e
}

func main() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET is required")
	}

	e := newRouter([]byte(secret))
	e.Logger.Fatal(e.Start(":1323"))
}
```

Секрет для HS256 должен быть случайным и достаточно длинным; для учебного запуска можно сгенерировать случайные байты и передать их в окружении. В реальном приложении дополнительно проверяют `iss`, `aud`, срок действия и схему передачи токена (обычно `Authorization: Bearer ...`).

**Файл `main_test.go`** — тест создаёт настоящий подписанный токен со сроком действия:

```go
package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWT(t *testing.T) {
	secret := []byte("this-is-a-test-secret-only-not-for-production")
	claims := Claims{
		Username: "tainella",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("missing token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/hello", nil)
		rec := httptest.NewRecorder()
		newRouter(secret).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("got %d, want 400", rec.Code)
		}
	})

	t.Run("valid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/hello", nil)
		req.Header.Set("X-Id-Token", raw)
		rec := httptest.NewRecorder()
		newRouter(secret).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got %d, want 200: %s", rec.Code, rec.Body.String())
		}
		if got, want := rec.Body.String(), "\"Hello tainella\"\n"; got != want {
			t.Fatalf("got body %q, want %q", got, want)
		}
	})

	t.Run("invalid signature", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/hello", nil)
		req.Header.Set("X-Id-Token", raw)
		rec := httptest.NewRecorder()
		newRouter([]byte("different-secret")).ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("got %d, want 403", rec.Code)
		}
	})
}
```

В учебной реализации отсутствие токена возвращает `400`, плохой токен — `403`, чтобы соответствовать смыслу оригинального примера. На практике для отсутствующей или некорректной аутентификации обычно используют **`401 Unauthorized`**, а **`403 Forbidden`** — для аутентифицированного пользователя без нужных прав.