# Interface & Polymorphism in Go

## What is an Interface?

An interface defines a **set of methods** that a type must implement. Any type that implements those methods automatically satisfies the interface — no explicit declaration needed.

```go
type Shape interface {
    Area() float64
    Perimeter() float64
    Name() string
}
```

Any type with these 3 methods IS a Shape — Circle, Rectangle, Triangle all qualify.

---

## What is Polymorphism?

Polymorphism means **one function works on many types**.

```go
func printShapeInfo(s Shape) {
    fmt.Printf("%s area=%.2f\n", s.Name(), s.Area())
}

// Same function, different types:
printShapeInfo(Circle{Radius: 5})
printShapeInfo(Rectangle{Width: 4, Height: 6})
printShapeInfo(Triangle{A: 3, B: 4, C: 5})
```

`printShapeInfo` doesn't know or care what the concrete type is — it just calls the interface methods.

---

## How Go Interfaces Work (Implicit)

```go
// Java (explicit):
class Circle implements Shape { ... }

// Go (implicit):
type Circle struct { Radius float64 }
func (c Circle) Area() float64 { ... }      // just implement the methods
func (c Circle) Perimeter() float64 { ... } // Go figures out the rest
func (c Circle) Name() string { ... }
// Circle automatically satisfies Shape interface
```

No `implements` keyword needed. If the methods match → the interface is satisfied.

---

## Example 1: Shape Polymorphism

```go
shapes := []Shape{
    Circle{Radius: 5},
    Rectangle{Width: 4, Height: 6},
    Triangle{A: 3, B: 4, C: 5},
}

for _, s := range shapes {
    printShapeInfo(s)  // works for all 3 types
}
```

**Output:**
```
Circle       area=78.54  perimeter=31.42
Rectangle    area=24.00  perimeter=20.00
Triangle     area=6.00   perimeter=12.00
```

---

## Example 2: Notifier Polymorphism

```go
type Notifier interface {
    Send(message string) error
}
```

Three different types, one interface:

```go
notifiers := []Notifier{
    EmailNotifier{To: "user@example.com"},
    SMSNotifier{Phone: "+91-9876543210"},
    SlackNotifier{Channel: "alerts"},
}

notifyAll(notifiers, "Server is down!")
// sends via Email, SMS, AND Slack — same function call
```

Adding a new notifier (e.g. PushNotification) requires zero changes to `notifyAll` — just implement `Send()`.

---

## Example 3: Type Switch

When you need to know the **concrete type** behind an interface:

```go
func describeAnimal(a Animal) {
    switch v := a.(type) {
    case Dog:
        fmt.Println("Dog named", v.Name)
    case Cat:
        fmt.Println("Cat named", v.Name)
    case Bird:
        fmt.Println("Bird named", v.Name)
    }
}
```

`a.(type)` extracts the concrete type. `v` is then the concrete value (e.g. a `Dog`) giving access to Dog-specific fields.

---

## Type Assertion vs Type Switch

```go
// Type assertion (single type check):
dog, ok := a.(Dog)
if ok {
    fmt.Println(dog.Name)
}

// Type switch (multiple types):
switch v := a.(type) {
case Dog:  ...
case Cat:  ...
}
```

Use type assertion when you expect one specific type. Use type switch when handling multiple possibilities.

---

## Interface Composition

Interfaces can embed other interfaces:

```go
type Reader interface {
    Read() string
}

type Writer interface {
    Write(s string)
}

type ReadWriter interface {
    Reader   // embeds Reader
    Writer   // embeds Writer
}
```

A type must implement both `Read()` and `Write()` to satisfy `ReadWriter`.

---

## Key Rules

| Rule | Reason |
|------|--------|
| Interface satisfied implicitly | No `implements` keyword — cleaner decoupling |
| Program to interfaces, not types | Makes code extensible without modification |
| Small interfaces are better | `io.Reader` (1 method) > a 10-method interface |
| Empty interface `any` accepts everything | Use sparingly — loses type safety |

---

## Real World Uses

| Use Case | Interface |
|----------|-----------|
| File, Network, Buffer | `io.Reader` / `io.Writer` |
| HTTP handlers | `http.Handler` |
| Sorting any slice | `sort.Interface` |
| Custom error types | `error` interface |
| Plugin systems | Define interface, swap implementations |
