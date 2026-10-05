#  **GO**

### Темы:
- Переменные
- Примитивы
- Управление потоком
- Функции

Кстати - эту лекцию удобнее всего читать через obsidian с плагином Go Playground, он позволяет запускать код прямо в этом файлике

#### Переменные
```go
package main  
  
import "fmt"  

var (
	something int64 = 14
	somethingElse = 22
)

const IAmImmutable = "hehe"
  
func main() {  
    var a int  
    B := 4  
    c, python, java := true, 1, "no!"
    
    fmt.Println("global variables:", something, somethingElse)
	fmt.Println("constant:", IAmImmutable) // global too
    fmt.Println("regular:", a, B)  
    fmt.Println("multiple:", c, python, java)
    fmt.Printf("%T\n", c)

}
```

```golang-run-result
global variables: 14 22
constant: hehe
regular: 0 4
multiple: true 1 no!
bool

```

Глобальные переменные с заглавной буквы - экспортируются, с строчной - нет
Внутри функции все переменные находятся лишь в скоупе функции (ее поле видимости), вне ф-ции к ней обратиться нельзя
В примере выше только переменная`IAmImmutable` экспортнута

Если захотите поиграться с типами и тд, можно проверять их таким образом:
`fmt.Printf("Type: %T Value: %v\n", a, a)`

Тут видно что к переменной `а` не было присвоено значение, но вывелось `0` - это `zero value` для численных типов, так же как и `""` для `string` и `false` для `bool`

#### Примитивы
```
bool

string

int  int8  int16  int32  int64
uint uint8 uint16 uint32 uint64 uintptr

byte // alias for uint8

rune // alias for int32
     // represents a Unicode code point

float32 float64

complex64 complex128
```

#### Функции

```go
package main

import "fmt"

func add(x int, y int) int {
	return x + y
}

func swap(x, y string) (string, string) {
	return y, x
}

func main() {
	fmt.Println("add:", add(42, 13))
	a, b := "first", "second"
	a, b = swap(a, b)
	fmt.Println("swap:", a, b)
}
```

```golang-run-result
add: 55
swap: second first

```

Функция может принимать и возвращать сколько угодно значений, но ресивер + аргументы + возвращаемые значения не должны превышать 1 ГБ (maxStackSize = 1 << 30) (см. [исходники](https://go.dev/src/cmd/compile/internal/ssagen/pgen.go)). И конечно же у вас должно быть минимум 1 ГБ оперативки для таких трюков.
##### Именованые возвращаемые значения
```go
package main

import "fmt"

func split(sum int) (int, int) {
	x := sum * 4 / 9
	y := sum - x
	return x, y
}

func main() {
	fmt.Println(split(19))
}

```

```golang-run-result
8 11

```

Полезно использовать если вы хотите более явно показать другим разработчикам что же возвращает ваша функция, но использовать надо осторожно, особенно с `defer` поведение может быть не самым очевидным ;)
#### Управление потоком
##### FOR
Типичный `for` как в С-шных языка
```go
package main

import "fmt"

func main() {
	sum := 0
	for i := 0; i < 10; i++ {
		sum += i
	}
	fmt.Println(sum)
}
```

```golang-run-result
45

```

Чаще используется `for` с `range`:
```go
package main

import "fmt"

func main(){
	nums := []int{1, 2, 5, 7, 22}
	for i := range nums {
		if i%2 == 0 {
			nums[i] *= 2
		}
	}
	fmt.Println("first iteration:", nums)
	
	for i, v := range nums {
		if v >= 10 {
			nums[i] = 0
		}
	}
	fmt.Println("second iteration:", nums)
	
	for _, v := range nums {
		v += 10
	}
	fmt.Println("third iteration:", nums)

	
}
```

```golang-run-result
first iteration: [2 2 10 7 44]
second iteration: [2 2 0 7 0]
third iteration: [2 2 0 7 0]

```


Вместо `while` тоже используется `for`:
```go
package main

import "fmt"

func main() {
	sum := 1
	for sum < 1000 {
		sum += sum
	}
	fmt.Println(sum)
}
```

```golang-run-result
1024

```

БЕСКОНЕЧНЫЙ ЦИКЛ!!
```
for {
	
}
```

##### IF
Базово ничего особенного, вот такой синтаксис:
```go
if x > 0 {
} else { // если надо else
}
```

Часто в коде можно увидеть как после `if` сначала переменной присваивается результат и проверка идет именно по результату, особенно при обработке ошибок так выходит компкатнее:
```go
package main

import "errors"

func main(){
	if err := someFunc(); err != nil {
		panic(err)
	}
}

func someFunc() error {
	return errors.New("big bad error!")
}

```

```golang-run-result
panic: big bad error!

goroutine 1 [running]:
main.main()
	/tmp/sandbox2629892396/src/prog.go:7 +0x47

```

##### SWITCH
Полезно если вы делаете какой-либо маппинг, например конвертировать ошибку к текстовке для фронта:
```go
package main

import (
	"fmt"
	"errors"
)

const (
	messageErrNotFound = "Глупенький пользователь, плохой запрос!"
	messageErrValidationFailed = "Чото забыли заполнить... Попробуйте снова"
	messageErrWierd = "Мы сами не поняли что произошло"
)

var (
	errNotFound = errors.New("not found")
	errValidationFailed = errors.New("validation failed")
)

func main(){
	err := errNotFound
	fmt.Println(errMapping(err))
}

func errMapping(err error) string {
	switch err {
	case errNotFound:
		return messageErrNotFound
	case errValidationFailed:
		return messageErrValidationFailed
	default:
		return messageErrWierd
	}
}
```

```golang-run-result
no

```
BTW `switch` идет сверху вниз: сначала сравнивается с `errNotFound`,   потом с `errValidationFailed`, но `default` всегда идет последним, только если ничего не подошло, даже если он стоит на первом месте

`switch` можно использовать и без условий, тогда просто возьмется за основу `switch true`

##### DEFER
Функция которая написана после стейтмента `defer` обрабатывается сразу перед выходом из родительской функции:
```go
package main

import "fmt"

func main() {
	defer fmt.Println("world")

	fmt.Println("hello")
}
```

```golang-run-result
hello
world

```

Аргументы в `defer` обрабатываются сразу, ноооо.... тут сложный момент
```go
package main

import "fmt"

func main() {
	i := true
	if i {
		defer fmt.Println("hey!")
	}
	i = false
}

```

```golang-run-result
hey!

```

```go
package main

import "fmt"

func main() {
	fmt.Println(deferFuncOne())
	fmt.Println(deferFuncTwo())
}

func deferFuncOne() (num int) {
	num = 5
	defer func(){
		num += 10
	}()
	return
}

func deferFuncTwo() (num int) {
	num = 5
	defer func(){
		num += 10
	}()
	return -30
}

```

```golang-run-result
15
-20

```