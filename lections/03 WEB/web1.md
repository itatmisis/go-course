# Урок 3 — WEB 1: сети, HTTP, JSON, REST и Echo (Go)  
  
## План занятия  
  
1. Клиент и сервер. Модель OSI, TCP и UDP.  
2. HTTP: запросы, ответы, методы, заголовки и статусы.  
3. HTTP-клиент на Go с `net/http`.  
4. JSON: типы данных, `encoding/json`, теги структур.  
5. REST и устройство API.  
6. HTTP-сервер на Echo: маршруты, параметры, обработчики.  
7. Модели данных, биндинг, валидация.  
8. Практика: CRUD API.
  
---  
  
## 1. Сети: от клиента до сервера  
  
**Клиент** отправляет запрос и ожидает ответ; **сервер** принимает запрос, выполняет необходимую работу и возвращает результат. Клиентом может быть браузер, мобильное приложение или другая backend-служба. Один и тот же процесс может быть и клиентом, и сервером в разных взаимодействиях.  
  
Передача данных по сети многоуровневая. Эталонная модель **OSI** выделяет семь уровней:  
  
| № | Уровень | Пример или назначение |  
|---|---|---|  
| 7 | Прикладной | HTTP, DNS, WebSocket как прикладной протокол поверх HTTP при установлении соединения |  
| 6 | Представления | Кодирование, представление данных, шифрование как функция уровня модели |  
| 5 | Сеансовый | Управление сеансами взаимодействия |  
| 4 | Транспортный | TCP, UDP |  
| 3 | Сетевой | IP, маршрутизация |  
| 2 | Канальный | Ethernet, Wi-Fi (канальный уровень) |  
| 1 | Физический | Радиосигналы, электрические и оптические сигналы |  
  
OSI — **концептуальная модель**, а не точная схема всех реальных сетевых стеков. Для интернета часто используют более практичную модель TCP/IP.  
  
### TCP и UDP  
  
**TCP** — потоковый протокол с установлением соединения, доставкой данных в правильном порядке и повторной передачей потерянных сегментов. Его используют HTTP/1.1 и HTTP/2.  
  
**UDP** — протокол датаграмм без гарантий доставки и упорядочивания на транспортном уровне. Подходит, когда приложение само выбирает стратегию надёжности или важнее низкие задержки. **HTTP/3 использует QUIC поверх UDP**, но QUIC самостоятельно обеспечивает надёжную передачу потоков.  
  
## 2. HTTP: запрос и ответ  
  
HTTP — прикладной протокол, организующий обмен сообщениями по схеме «запрос — ответ». Им передают HTML, JSON, изображения и другие данные. HTTPS — HTTP с защитой транспортного соединения (TLS).  
  
Пример запроса:  
  
```http  
GET /users/42?verbose=true HTTP/1.1  
Host: localhost:1323  
Accept: application/json  
User-Agent: example-client/1.0  
```  
  
Части запроса:  
  
- **Метод** `GET` — желаемая операция.  
- **Path** `/users/42` — адрес ресурса.  
- **Query string** `verbose=true` — дополнительные параметры.  
- **Headers** — метаданные запроса.  
- **Body** — тело запроса (может отсутствовать).  
  
Пример ответа:  
  
```http  
HTTP/1.1 200 OK  
Content-Type: application/json; charset=UTF-8  
  
{"id":42,"username":"alice"}  
```  
  
Ответ содержит статус, заголовки и, как правило, тело. HTTP/2 и HTTP/3 передают сообщения в ином формате на уровне соединения, но сохраняют семантику методов, статусов и заголовков.  
  
### Методы HTTP  
  
| Метод | Типичное назначение | Свойства |  
|---|---|---|  
| `GET` | Получить ресурс | Safe, идемпотентный |  
| `HEAD` | Получить только заголовки ответа | Safe, идемпотентный |  
| `POST` | Создать ресурс / выполнить действие | Не обязательно идемпотентный |  
| `PUT` | Полностью создать или заменить ресурс по адресу | Идемпотентный |  
| `PATCH` | Частично изменить ресурс | Не обязательно идемпотентный |  
| `DELETE` | Удалить ресурс | Идемпотентный по семантике |  
  
**Идемпотентность:** повтор идентичного запроса, должен иметь тот же *предполагаемый эффект на состояние сервера*, что и один запрос; коды ответов могут отличаться. **Safe:** метод не предназначен для изменения состояния ресурса.  
  
### Коды ответа  
  
- `2xx` — успех: `200 OK`, `201 Created`, `204 No Content`.  
- `3xx` — перенаправление: `301 Moved Permanently`, `302 Found`, `304 Not Modified`.  
- `4xx` — ошибка на стороне запроса: `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found`, `409 Conflict`, `422 Unprocessable Content`.  
- `5xx` — ошибка сервера: `500 Internal Server Error`, `503 Service Unavailable`.  
  
`401` означает, что требуется корректная аутентификация, а `403` — что доступ запрещён.  
  
### Заголовки  
  
- `Content-Type: application/json` — формат **передаваемого тела**.  
- `Accept: application/json` — желаемый формат **ответа**.  
- `Authorization: Bearer ...` — данные авторизации.  
  
## 3. HTTP-клиент в Go: пакет `net/http`  
  
В Python часто используют библиотеку `requests`. В Go HTTP-клиент есть в стандартной библиотеке. Создадим `client.go`:  
  
```go  
package main  
  
import (  
    "context"
    "fmt"
    "io"
    "log"
    "net/http"
    "net/url"
    "time"
    )  
  
func main() {  
    client := &http.Client{Timeout: 5 * time.Second}  
    u, err := url.Parse("https://httpbin.org/get")
    if err != nil { log.Fatal(err) }
    
    q := u.Query()
    q.Set("language", "Go")
    u.RawQuery = q.Encode() 
     
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)     if err != nil { log.Fatal(err) }    
	
	req.Header.Set("Accept", "application/json")  
    resp, err := client.Do(req)    
    if err != nil { log.Fatal(err) }    
    defer resp.Body.Close()  
    
    body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))    
    if err != nil { log.Fatal(err) }  
    
    fmt.Println("Status:", resp.Status)    
    fmt.Println("Content-Type:", resp.Header.Get("Content-Type"))
    fmt.Println(string(body))}  
```  
  
Запуск: `go run client.go`.  
  
**На что обратить внимание:**  
  
1. `http.Client` желательно переиспользовать; у него есть пул соединений.  
2. `client.Do` возвращает ошибку при сетевой проблеме, **не** при ответе `404` или `500`. Статус проверяем отдельно через `resp.StatusCode`.  
3. Тело ответа — поток `io.ReadCloser`: закрываем его с `defer resp.Body.Close()`.  
4. `context.Context` позволяет отменять запросы; `Timeout` ограничивает их длительность.  
5. `url.Values` корректно экранирует query-параметры.  
  
## 5. REST API  
  
REST (Representational State Transfer) — архитектурный стиль. Важные ограничения REST: клиент–сервер, отсутствие состояния клиента между запросами на сервере (stateless), кэшируемость, единый интерфейс, многоуровневая система; опционально — передача исполняемого кода клиенту (code-on-demand).  
  
**Stateless** не означает «серверу запрещено хранить данные в БД». Это означает, что сервер не должен требовать скрытого контекста *предыдущих запросов*, чтобы понять текущий запрос.  
  
Ресурсы обычно описываются существительными в URL, а действия — HTTP-методами:  
  
| Запрос            | Значение                        | Ожидаемый ответ       |     |
| ----------------- | ------------------------------- | --------------------- | --- |
| `GET /users`      | Список пользователей            | `200` + JSON-массив   |     |
| `GET /users/1`    | Один пользователь               | `200` или `404`       |     |
| `POST /users`     | Создать пользователя            | `201` + JSON          |     |
| `PUT /users/1`    | Полностью заменить пользователя | `200`/`204` или `404` |     |
| `PATCH /users/1`  | Изменить часть полей            | `200`/`204` или `404` |     |
| `DELETE /users/1` | Удалить пользователя            | `204` или `404`       |     |
  
REST — не синоним «любой JSON поверх HTTP». Семантика маршрутов, методов и статусов тоже важна.  
  
## 6. HTTP-сервер: знакомимся с Echo  
  
Минимальный `main.go`:  
  
```go  
package main  
  
import (  
    "net/http"    
    "github.com/labstack/echo/v4"
    )  
  
func main() {  
    e := echo.New()    
    e.GET("/hello", func(c echo.Context) error {        
	    return c.JSON(http.StatusOK, map[string]string{"message": "Hello, Go!"})     })    
	    
	e.Logger.Fatal(e.Start(":1323"))
}  
```  
  
Запускаем: `go run .`. В другом терминале:  
  
```bash  
curl -i http://localhost:1323/hello  
```  
  
**Что происходит:** `echo.New()` создаёт приложение, `e.GET()` регистрирует маршрут, функция-обработчик принимает `echo.Context`, `c.JSON()` выставляет JSON-ответ и статус, `e.Start()` запускает HTTP-сервер.  
  
### Path, query и тело запроса  
  
```go  
e.GET("/users/:id", func(c echo.Context) error {  
    id := c.Param("id") // /users/42 -> "42"    verbose := c.QueryParam("verbose") // ?verbose=true    return c.JSON(http.StatusOK, map[string]string{        "id": id, "verbose": verbose,    })})  
```  
  
`c.Param` и `c.QueryParam` возвращают **строки**. Если поле должно быть числом, используйте `strconv.Atoi` и обрабатывайте ошибку, не доверяя входу.  
  
### Middleware  
  
Middleware — обработчики, работающие *вокруг* основного обработчика: логирование, восстановление после panic, авторизация, CORS и т. д.  
  
```go  
import "github.com/labstack/echo/v4/middleware"  
  
e.Use(middleware.Logger())  
e.Use(middleware.Recover())  
```  
  
В реальном проекте можно подключать middleware глобально или к группам маршрутов.  
  
## 7.  Модели и валидация Go  
  
1. **DTO** — структуры запроса и ответа с `json`-тегами.  
2. **Биндинг** — `c.Bind(&dto)` (Echo распознаёт источники параметров согласно своим правилам; для JSON API отдельно контролируйте `Content-Type`).  
3. **Валидация** — собственная функция или библиотека, например `go-playground/validator` с подключением к `e.Validator`.  
4. **Бизнес-логика** — обработка уже проверенных данных.  
  
Пример отдельной DTO, которую безопасно отдавать клиенту:  
  
```go  
type CreateUserRequest struct {  
    Username string `json:"username"`    
    Age      int    `json:"age"`    
    Email    string `json:"email"`
    }  
  
type UserResponse struct {  
    ID       int    `json:"id"`    
    Username string `json:"username"`    
    Age      int    `json:"age"`    
    Email    string `json:"email"`
    }  
```  
  
Почему не стоит декодировать запрос сразу в структуру БД? Клиент может прислать поля, которыми не должен управлять (`id`, `role`, `is_admin`). Это называют **mass assignment / over-posting**.  
  
Простейшая ручная валидация:  
  
```go  
func validateUser(input CreateUserRequest) error {  
    if strings.TrimSpace(input.Username) == "" {        
	    return errors.New("username is required")    
	}    
	if input.Age < 18 {        
		return errors.New("user must be adult")    
	}    
	email, err := mail.ParseAddress(input.Email)    
	if err != nil || email.Address != input.Email {        
		return errors.New("invalid email")    
	}    
	
	return nil}  
```  
  
Необходимые импорты: `errors`, `net/mail`, `strings`. `mail.ParseAddress` — **учебная базовая проверка**, а не гарантия, что email существует или доставляется. Проверка пароля отдельна от хранения: **пароли нельзя возвращать в API или хранить в открытом виде**; в настоящем сервисе нужен безопасный хэш.  
  
## 8. Практика

Пишем сервис