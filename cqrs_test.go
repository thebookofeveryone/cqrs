package cqrs_test

import (
	"log"
	"testing"

	"github.com/thebookofeveryone/cqrs"
)

type SomethingDoneEvent struct {
	Value string
}

type Handlers struct {
}

func (h *Handlers) HandleSomethingDoneEvent(e SomethingDoneEvent) {
	if e.Value != "Test" {
		log.Fatal("Expected e.Value to equal `Test`")
	}
}

func TestHandlers(t *testing.T) {
	bus := cqrs.NewBus()
	handlers := Handlers{}
	bus.RegisterHandlers(&handlers)
	bus.Publish(SomethingDoneEvent{"Test"})
}
