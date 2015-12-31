package cqrs_test

import (
	"fmt"
	_ "testing"

	"github.com/thebookofeveryone/cqrs"
)

type SomethingDoneEvent struct {
	Value string
}

type Handlers struct {
}

func (h *Handlers) HandleSomethingDoneEvent(e SomethingDoneEvent) {
	fmt.Println("Handled", e.Value)
}

func ExampleHandlers() {
	bus := cqrs.NewBus()
	handlers := Handlers{}
	bus.RegisterHandlers(&handlers)
	bus.Publish(SomethingDoneEvent{"Test"})
	// Output:
	// Handled Test
}

func ExampleGlobalHandlers() {
	bus := cqrs.NewBus()
	handlers := Handlers{}
	bus.RegisterHandlers(&handlers)
	bus.AddGlobalHandler(func(e interface{}) {
		switch e.(type) {
		case SomethingDoneEvent:
			fmt.Printf("[LOG] SomethingDoneEvent: %s\n",
				e.(SomethingDoneEvent).Value)
		}
	})
	bus.Publish(SomethingDoneEvent{"Test"})
	// Output:
	// Handled Test
	// [LOG] SomethingDoneEvent: Test
}
