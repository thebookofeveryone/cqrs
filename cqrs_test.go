package cqrs_test

import (
	"fmt"
	_ "testing"

	"github.com/thebookofeveryone/cqrs"
)

type UserCreatedEvent struct {
	Name string
}

type UserNameChangedEvent struct {
	NewName string
}

type Handlers struct {
}

func (h *Handlers) HandleUserCreatedEvent(e UserCreatedEvent) {
	fmt.Println("User created:", e.Name)
}

func ExampleHandlers() {
	bus := cqrs.NewBus()
	handlers := Handlers{}
	bus.RegisterHandlers(&handlers)
	bus.Publish(UserCreatedEvent{"John"})
	// Output:
	// User created: John
}

func ExampleGlobalHandlers() {
	bus := cqrs.NewBus()
	bus.AddGlobalHandler(func(e interface{}) {
		fmt.Printf("[LOG] %T: %v\n", e, e)
	})
	bus.Publish(UserCreatedEvent{"John"})
	bus.Publish(UserNameChangedEvent{"Mark"})
	// Output:
	// [LOG] cqrs_test.UserCreatedEvent: {John}
	// [LOG] cqrs_test.UserNameChangedEvent: {Mark}
}
