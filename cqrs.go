package cqrs

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/satori/go.uuid"
)

func GenerateUUID() string {
	return uuid.NewV4().String()
}

type Event struct {
	Version   string
	Timestamp int64
	Type      string
	Payload   interface{}
}

type Command struct {
	Timestamp int64
	Type      string
	Payload   interface{}
}

type AggregateRoot struct {
	Changes  map[string]Event
	handlers map[reflect.Type]func(interface{})
}

func NewAggregateRoot(source interface{}) AggregateRoot {
	a := AggregateRoot{
		make(map[string]Event),
		make(map[reflect.Type]func(interface{}))}
	a.RegisterHandlers(source)
	return a
}

func (a *AggregateRoot) ClearChanges() {
	a.Changes = make(map[string]Event)
}

func (a *AggregateRoot) Source(originalEvent interface{}) error {
	eventType := reflect.TypeOf(originalEvent)
	event := Event{
		GenerateUUID(),
		time.Now().Unix(),
		eventType.Name(),
		originalEvent}
	if val, ok := a.handlers[eventType]; ok {
		val(originalEvent)
	}
	a.Changes[event.Version] = event
	serialised, _ := json.Marshal(event)
	fmt.Println(string(serialised))
	return nil
}

func (a *AggregateRoot) RegisterHandlers(source interface{}) {

	productType := reflect.TypeOf(source)
	numMethods := productType.NumMethod()

	for i := 0; i < numMethods; i++ {

		method := productType.Method(i)

		if strings.HasPrefix(method.Name, "Handle") {
			eventType := method.Type.In(1)
			handler := func(event interface{}) {
				eventValue := reflect.ValueOf(event)
				method.Func.Call([]reflect.Value{
					reflect.ValueOf(source),
					eventValue})
			}
			a.handlers[eventType] = handler
		}

	}

}

type Handlers map[reflect.Type][]func(interface{})

type Bus struct {
	handlers Handlers
}

func NewBus() Bus {
	return Bus{make(Handlers)}
}

func (b Bus) Publish(message interface{}) error {
	eventType := reflect.TypeOf(message)
	fmt.Println("Publishing...", eventType)
	if val, ok := b.handlers[eventType]; ok {
		for _, handler := range val {
			handler(message)
		}
	}
	return nil
}

func (b *Bus) addHandler(eventType reflect.Type, handler func(interface{})) {
	_, ok := b.handlers[eventType]
	if !ok {
		b.handlers[eventType] = make([]func(interface{}), 0)
	}
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *Bus) RegisterHandler(handler func(interface{})) {
	eventType := reflect.TypeOf(handler).In(1)
	b.addHandler(eventType, handler)
}

func (b *Bus) RegisterHandlers(source interface{}) {

	productType := reflect.TypeOf(source)
	numMethods := productType.NumMethod()

	for i := 0; i < numMethods; i++ {

		method := productType.Method(i)

		if strings.HasPrefix(method.Name, "Handle") {
			eventType := method.Type.In(1)
			handler := func(event interface{}) {
				eventValue := reflect.ValueOf(event)
				method.Func.Call([]reflect.Value{
					reflect.ValueOf(source),
					eventValue})
			}
			b.addHandler(eventType, handler)
		}

	}

}
