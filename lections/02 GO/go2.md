#  **GO**

### Темы:

- Массивы
- Слайсы
- Мапы
- Структуры
- Указатели
- Методы
- Интерфейсы
- any
- Обработка ошибок
- marchal/unmarshal
-  tags

#### Массивы

Массивы имеют четкий тип - он задается типом элементов и размером массива!!
```go
package main

import "fmt"

func main() {
	var a [2]string
	a[0] = "Hello"
	a[1] = "World"
	fmt.Println(a[0], a[1])
	fmt.Println(a)

	primes := [6]int{2, 3, 5, 7, 11, 13}
	fmt.Println(primes)
}
```

```golang-run-result
Hello World
[Hello World]
[2 3 5 7 11 13]

```

```go
package main

import "fmt"

func main() {
	nums := [2]int{1, 2}
	moreNums := [3]int{2, 3, 4}
	fmt.Println(nums, moreNums)
	fmt.Println(nums != moreNums)
}
```

```golang-run-result
Compilation Error:
./prog.go:9:14: invalid operation: nums > moreNums (operator > not defined on array)

```

Честно говоря это нигде в бэкенде не используется за редким-редким исключением
#### Слайсы
Грубо говоря это динамические массивы, но на деле - не совсем, но понимать так легче (пока).


```go
package main

import "fmt"


func main() {
	s1 := make([]int, 1, 10)
	fmt.Printf("BEFORE\n1. len=%d cap=%d\n", len(s1), cap(s1))
	
	s1 = extendSlice(s1)
	fmt.Printf("AFTER\n1. len=%d cap=%d\n", len(s1), cap(s1))
}

func extendSlice(s []int) []int {
	return append(s, make([]int, 500)...)
}

```

```golang-run-result
BEFORE
1. len=1 cap=10
AFTER
2. len=501 cap=512

```

`len` и `cap`
zero-value (nil)
append

#### Мапы

Мапы это те же словари в питоне

Можно проверять наличие элемента в мапе

```go
package main

import "fmt"

func main() {
	//mp := make(map[int]string)
	//mp[1] = "one"
	m := map[int]string{1:"one"}
	
	if v, ok := m[1]; ok { //ok == true
		fmt.Println(v)
	}
	if v, ok := m[2]; ok { //ok == false, v == ""
		fmt.Println(v)
	}
}

```

```golang-run-result
one

```

Ключи мап уникальны - по одному ключу может лежать только одно значение, если положить в тот же ключ другое значение - оно перезапишется.
Не каждый тип может быть ключом мапы - он должен иметь свойство comparable, то есть на нем должны быть определены операции сравнения
```go
package main

import "fmt"

type some struct {
	str string
	i int
	list []string // makes struct not comparable
}

func main(){
	s := some{str: "sdfsd", i: 25}
	m := make(map[some]string)
	m[s] = "wow"
	
	fmt.Println(m[s])
}
```

```golang-run-result
Compilation Error:
./prog.go:13:16: invalid map key type some

```

#### Структуры

Структуры заменяют классы или объекты в Го, классическое определение структуры - "коллекция полей", мы уже немного пользовались ими раннее

```go
package main

import "fmt"

type Vertex struct {
	X int
	Y int
}

func main() {
	v := Vertex{1, 2}
	fmt.Println(v)
	
	v.X = 15 // доступ к полю
	fmt.Println(v)
	fmt.Printf("%T\n", v)
}

```

```golang-run-result
{1 2}
{15 2}
main.Vertex

```

#### Указатель

Указатель это значение адреса ячейки памяти, где хранится переменная 

```go
package main

import "fmt"

func main() {
	i, j := 42, 2701

	p := &i         // присваеваем переменной p адрес памяти                                         // переменной i
	fmt.Println(p)
	fmt.Println(*p) // читаем значение, лежащее по адресу
	
	*p = 21         // устанавливаем новое значение в том же адресе
	fmt.Println(i)  // значение i меняется - мы же меняли ее память

	p = &j         // присваеваем указатель на j
	*p = *p / 37   // значению, лежащему по адресу j присваеваем его же / 37
	fmt.Println(j) // see the new value of j
}
```

```golang-run-result
0xc397d288020
42
21
73

```

`&` - возвращает указатель вместо переменной 
`*` - возвращает значение переменной от ее указателя

#### Методы

Методы можно объявлять на типы, объявленные внутри этого пакета 

```go
package main

import "fmt"

type Person struct {
	Name string
	Age int
}

func main(){
	ivan := Person{Name:"Ivan", Age:25}
	fmt.Println(ivan.makeNewFriend())
	fmt.Println(makeNewFriend(&ivan))
}

func (p *Person) makeNewFriend() *Person {
	return &Person{Name: p.Name + "'s friend", Age: p.Age }
}

func makeNewFriend(p *Person) *Person {
	return &Person{Name: p.Name + "'s friend", Age: p.Age }
}
```

```golang-run-result
&{Ivan's friend 25}
&{Ivan's friend 25}

```

Метод это обычная функция, которая принимает первым аргументом - ресивер. Ресивер - тип на который написан метод (в нашем случае это `*Person`). В других языках к ресиверам можно обращаться через `self, this` и тд

```go
package main

import "fmt"

type Person struct {
	Name string
	Age int
}

func main(){
	ivan := Person{Name:"Ivan", Age:30}
	ivan.makeOlder()
	
	fmt.Println(ivan)
}

func (p *Person) makeOlder() {
	p.Age++
}

```

```golang-run-result
{Ivan 31}

```

С методами Го сам подстраивается и на самом деле, в примере выше, вызывает `(&ivan).makeOlder()`, то есть сам понимает что метод определен только на указателе на тип Person и вызывать функцию нужно от него
```go
package main

import "fmt"

type Person struct {
	Name string
	Age int
}

func main(){
	ivan := Person{Name:"Ivan", Age:30}
	makeOlder(&ivan)
	
	fmt.Println(ivan)
}

func makeOlder(p *Person) {
	p.Age++
}
```

```golang-run-result
{Ivan 31}

```

В других же случаях нужно явно указывать передаешь ты в функцию указатель или значение, иначе код не скомпилируется
Указатели помогают избежать копирования больших структур!

#### Интерфейсы

Интерфейс - это тип, который определяется набором методов
Типы - реализуют интерфейсы, если реализуют эти методы (сигнатуры должны строго совпадать!)

```go
package main

import "fmt"

type Person struct {
	Name string
	Age int
	PhoneNumber *int
}

/*type Stringer interface{
	String() string
}*/

func (p *Person) String() string {
	return fmt.Sprintf("%s is %d years old, phone number = %d", p.Name, p.Age,       *p.PhoneNumber)
}

func main(){
	p := &Person{Name:"Peter", Age:26, PhoneNumber: new(8999111111)}
	fmt.Printf("person: %s\n", p)
}
```

```golang-run-result
person: Peter is 26 years old, phone number = 8999111111

```

Функция может просить на вход интерфейс, тогда любой тип, реализующий его, может быть ей передан:
```go
package main

import "fmt"

type redisCache struct {
	conn string          //какой-нибудь коннект к редису
}

type memoryCache struct {
	m map[string]string
}

type Cache interface{
	Get() string
}

func (rc *redisCache) Get() string {
	// махинации с редисом
	return rc.conn
}

func (mc *memoryCache) Get() string {
	return mc.m["cache"]
}

func MergeCache(cache1, cache2 Cache) string {
	return cache1.Get() + " + " + cache2.Get()
}

func main(){
	redis := &redisCache{conn: "redis cached data"}
	memory := &memoryCache{m: map[string]string{"cache":"memory cached data"}}
	
	v := MergeCache(memory, redis)
	
	fmt.Println(v)
}
```

```golang-run-result
memory cached data + redis cached data

```

#### any
any - то же что и `interface{}` (пустой интерфейс) - может содержать любой тип

```go
package main

import "fmt"

func main() {
	var i interface{}
	describe(i)

	i = 42
	describe(i)

	i = "hello"
	describe(i)
}

func describe(i interface{}) {
	fmt.Printf("(%v, %T)\n", i, i)
}
```

```golang-run-result
(<nil>, <nil>)
(42, int)
(hello, string)

```

Используется совсем не часто, просто потому что с `any` крайне неудобно работать, потому что придется использовать type assertion (приведение типов):

```go
package main

import "fmt"

func main(){
	process(122)
	process("hehe")
}

func process(v any) { 
	switch val := v.(type) {
		case string: fmt.Println("str:", val) 
		case int: fmt.Println("num:", val) 
	} 
}
```

```golang-run-result
num: 122
str: hehe

```

#### Ошибки

**ОБРАБАТЫВАТЬ ОШИБКИ НУЖНО АБСОЛЮТНО ВСЕГДА** 
99% случаев - просто прокидываем их выше по стеку вызовов, добавляя к ним смыслового контекста

```go
package main

import (
	"fmt"
	"errors"
	"strconv"
)

var errNotFound = errors.New("not found")

func main() {
	s, err := getStudents([]int{1, 2, 3})
	if err != nil { // err == nil
		panic(fmt.Errorf("getStudents: %w", err))
	}
	
	fmt.Println(s)
	
	s, err = getStudents([]int{0, -2})
	if err != nil { // true, err != nil
		panic(fmt.Errorf("getStudents: %w", err))
	}
	
	fmt.Println(s)
}


func getStudents(ids []int) ([]string, error) {
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		name, err := getName(id)
		if err != nil {
			return nil, fmt.Errorf("getName: %w", err)
		}
		names = append(names, name)
	}
	
	return names, nil
}

func getName(id int) (string, error){
	if id < 0 {
		return "", errNotFound
	}
	
	return "name" + strconv.Itoa(id), nil
}
```

```golang-run-result
[name1 name2 name3]
panic: getStudents: getName: not found

goroutine 1 [running]:
main.main()
	/tmp/sandbox2173673646/src/prog.go:21 +0x147

```

Ошибки в го имеют отдельный тип - `error` (интерфейс)

```
type error interface {
    Error() string
}
```

Так как это интерфейс, можно делать кастомные ошибки и кидать их в любые функции, которые обрабатывают ошибки, потому что error это интерфейсик. Так редко делают на практике, обычно это скрыто от глаз бэкендеров внутри библиотек которыми мы пользуемся

#### marshal / unmarshal и тэги

Это функции которые сериализуют / десериализуют JSONы (запаковывают / распаковывают)

```go
package main

import (
	"fmt"
	"encoding/json"
)

const msg = `
{
	"team_name":"amazing digital misis",
	"place":1
}
`
type team struct {
	Name  string //`json:"team_name"` // - тэг
	Place int    //`json:"place"`
}

func main(){
	var team1 team
	
	err := json.Unmarshal([]byte(msg), &team1)
	if err != nil {
		panic(err)
	}
	
	fmt.Println(team1)
	
	team1.Place = 2
	
	team2, err := json.Marshal(&team1)
	if err != nil {
		panic(err)
	}
	
	fmt.Println(string(team2))
}

```

```golang-run-result
{ 1}
{"Name":"","Place":2}

```

