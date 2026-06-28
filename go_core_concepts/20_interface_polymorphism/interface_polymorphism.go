package main

import (
	"fmt"
	"math"
)

// --- Example 1: Shape polymorphism ---

type Shape interface {
	Area() float64
	Perimeter() float64
	Name() string
}

type Circle struct {
	Radius float64
}

func (c Circle) Area() float64      { return math.Pi * c.Radius * c.Radius }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.Radius }
func (c Circle) Name() string       { return "Circle" }

type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Area() float64      { return r.Width * r.Height }
func (r Rectangle) Perimeter() float64 { return 2 * (r.Width + r.Height) }
func (r Rectangle) Name() string       { return "Rectangle" }

type Triangle struct {
	A, B, C float64 // sides
}

func (t Triangle) Area() float64 {
	s := (t.A + t.B + t.C) / 2
	return math.Sqrt(s * (s - t.A) * (s - t.B) * (s - t.C))
}
func (t Triangle) Perimeter() float64 { return t.A + t.B + t.C }
func (t Triangle) Name() string       { return "Triangle" }

func printShapeInfo(s Shape) {
	fmt.Printf("%-12s area=%.2f  perimeter=%.2f\n", s.Name(), s.Area(), s.Perimeter())
}

func totalArea(shapes []Shape) float64 {
	total := 0.0
	for _, s := range shapes {
		total += s.Area()
	}
	return total
}

// --- Example 2: Notification polymorphism ---

type Notifier interface {
	Send(message string) error
}

type EmailNotifier struct {
	To string
}

func (e EmailNotifier) Send(message string) error {
	fmt.Printf("  [Email → %s]: %s\n", e.To, message)
	return nil
}

type SMSNotifier struct {
	Phone string
}

func (s SMSNotifier) Send(message string) error {
	fmt.Printf("  [SMS → %s]: %s\n", s.Phone, message)
	return nil
}

type SlackNotifier struct {
	Channel string
}

func (s SlackNotifier) Send(message string) error {
	fmt.Printf("  [Slack #%s]: %s\n", s.Channel, message)
	return nil
}

func notifyAll(notifiers []Notifier, message string) {
	for _, n := range notifiers {
		n.Send(message)
	}
}

// --- Example 3: Type assertion and type switch ---

type Animal interface {
	Sound() string
}

type Dog struct{ Name string }
type Cat struct{ Name string }
type Bird struct{ Name string }

func (d Dog) Sound() string  { return "Woof" }
func (c Cat) Sound() string  { return "Meow" }
func (b Bird) Sound() string { return "Tweet" }

func describeAnimal(a Animal) {
	switch v := a.(type) {
	case Dog:
		fmt.Printf("  Dog named %s says %s\n", v.Name, v.Sound())
	case Cat:
		fmt.Printf("  Cat named %s says %s\n", v.Name, v.Sound())
	case Bird:
		fmt.Printf("  Bird named %s says %s\n", v.Name, v.Sound())
	default:
		fmt.Println("  Unknown animal")
	}
}

func main() {
	// Example 1: Shapes
	fmt.Println("=== Example 1: Shape Polymorphism ===")
	shapes := []Shape{
		Circle{Radius: 5},
		Rectangle{Width: 4, Height: 6},
		Triangle{A: 3, B: 4, C: 5},
	}
	for _, s := range shapes {
		printShapeInfo(s)
	}
	fmt.Printf("Total area: %.2f\n", totalArea(shapes))

	fmt.Println()

	// Example 2: Notifiers
	fmt.Println("=== Example 2: Notifier Polymorphism ===")
	notifiers := []Notifier{
		EmailNotifier{To: "user@example.com"},
		SMSNotifier{Phone: "+91-9876543210"},
		SlackNotifier{Channel: "alerts"},
	}
	notifyAll(notifiers, "Server is down!")

	fmt.Println()

	// Example 3: Type switch
	fmt.Println("=== Example 3: Type Switch ===")
	animals := []Animal{
		Dog{Name: "Bruno"},
		Cat{Name: "Whiskers"},
		Bird{Name: "Tweety"},
	}
	for _, a := range animals {
		describeAnimal(a)
	}
}
