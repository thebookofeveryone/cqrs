package cqrs

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/satori/go.uuid"
)

func GenerateUUID() string {
	return uuid.NewV4().String()
}

type Message struct {
	Id        string
	Timestamp int64
	Type      string
	Payload   interface{}
}

type RawMessage struct {
	Id        string
	Timestamp int64
	Type      string
	Payload   json.RawMessage
}

type AggregateRoot struct {
	Changes    map[string]Message
	dispatcher Dispatcher
}

type Stringer interface {
	String() string
}

func NewAggregateRoot(source interface{}) *AggregateRoot {
	a := AggregateRoot{
		make(map[string]Message),
		Dispatcher{}}
	a.dispatcher.RegisterHandlers(source)
	return &a
}

func NewAggregateRootFromHistory(source interface{}, history []Message) *AggregateRoot {
	a := NewAggregateRoot(source)
	for _, event := range history {
		a.dispatcher.Dispatch(event.Payload)
	}
	return a
}

func (a *AggregateRoot) ClearChanges() {
	a.Changes = make(map[string]Message)
}

func NewMessage(originalEvent interface{}) Message {
	eventType := reflect.TypeOf(originalEvent)
	message := Message{
		GenerateUUID(),
		time.Now().Unix(),
		eventType.Name(),
		originalEvent}
	return message
}

func (a *AggregateRoot) Source(originalEvent interface{}) error {
	event := NewMessage(originalEvent)
	a.dispatcher.Dispatch(originalEvent)
	a.Changes[event.Id] = event
	return nil
}
