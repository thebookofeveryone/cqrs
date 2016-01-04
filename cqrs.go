package cqrs

import (
	_ "fmt"
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

func NewAggregateRootFromHistory(source interface{}, history []Event) AggregateRoot {
	a := NewAggregateRoot(source)
	for _, event := range history {
		a.dispatchEvent(event.Payload)
	}
	return a
}

func (a *AggregateRoot) ClearChanges() {
	a.Changes = make(map[string]Event)
}

func NewEvent(originalEvent interface{}) Event {
	eventType := reflect.TypeOf(originalEvent)
	return Event{
		GenerateUUID(),
		time.Now().Unix(),
		eventType.Name(),
		originalEvent}
}

func (a *AggregateRoot) Source(originalEvent interface{}) error {
	event := NewEvent(originalEvent)
	a.dispatchEvent(originalEvent)
	a.Changes[event.Version] = event
	return nil
}

func (a *AggregateRoot) dispatchEvent(event interface{}) error {
	if val, ok := a.handlers[reflect.TypeOf(event)]; ok {
		val(event)
	}
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
type GlobalHandler func(interface{})
type GlobalHandlers []GlobalHandler

type Broker interface {
	Broadcast(int, interface{}) error
	Subscribe(func(interface{})) (int, error)
	Unsubscribe(int) error
}

type InMemoryBroker struct {
	subscribers []func(interface{})
}

func (b *InMemoryBroker) Broadcast(from int, message interface{}) error {
	for i, subscriber := range b.subscribers {
		if i == from {
			continue
		}
		subscriber(message)
	}
	return nil
}

func (b *InMemoryBroker) Subscribe(handler func(interface{})) (int, error) {
	b.subscribers = append(b.subscribers, handler)
	return len(b.subscribers) - 1, nil
}

func (b *InMemoryBroker) Unsubscribe(id int) error {
	b.subscribers = append(b.subscribers[:id], b.subscribers[id+1:]...)
	return nil
}

var inMemoryBroker = &InMemoryBroker{make([]func(interface{}), 0)}

type Bus struct {
	broker         Broker
	handlers       Handlers
	globalHandlers GlobalHandlers
	subscriberId   int
}

func NewInMemoryBus() *Bus {
	bus := Bus{inMemoryBroker, make(Handlers), make(GlobalHandlers, 0), 0}
	bus.subscriberId, _ = bus.broker.Subscribe(func(message interface{}) {
		bus.Dispatch(message)
	})
	return &bus
}

func NewRedisBus() Bus {
	return Bus{inMemoryBroker, make(Handlers), make(GlobalHandlers, 0), 0}
}

func (b Bus) Publish(message interface{}) error {
	b.Dispatch(message)
	return b.broker.Broadcast(b.subscriberId, message)
}

func (b Bus) Close() error {
	b.broker.Unsubscribe(b.subscriberId)
	return nil
}

func (b Bus) Dispatch(message interface{}) error {
	eventType := reflect.TypeOf(message)
	if val, ok := b.handlers[eventType]; ok {
		for _, handler := range val {
			handler(message)
		}
	}
	for _, handler := range b.globalHandlers {
		handler(message)
	}
	return nil
}

func (b *Bus) AddGlobalHandler(handler GlobalHandler) {
	b.globalHandlers = append(b.globalHandlers, handler)
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
