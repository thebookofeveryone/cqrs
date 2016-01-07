package cqrs_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/thebookofeveryone/cqrs"
	"github.com/thebookofeveryone/cqrs/redis"
)

type UserCreatedEvent struct {
	Name string
}

type UserNameChangedEvent struct {
	NewName string
}

type Handlers struct {
	mock.Mock
	Name string
}

func (h *Handlers) HandleUserCreatedEvent(e UserCreatedEvent) {
	h.Name = e.Name
	h.Called(e)
}

func (h *Handlers) HandleUserNameChangedEvent(e UserNameChangedEvent) {
	h.Name = e.NewName
	h.Called(e)
}

func Eventually(t *testing.T, test func() bool, timeout time.Duration, tick time.Duration) bool {
	timeoutChannel := time.After(timeout)
	ticker := time.NewTicker(tick)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if test() {
				return true
			}
		case <-timeoutChannel:
			assert.Fail(t, fmt.Sprintf("Timeout of %s reached", timeout))
			return false
		}
	}
}

func TestHandlers(t *testing.T) {
	dispatcher := cqrs.NewDispatcher()
	bus := redis.NewMessageBus("commands", redis.ConnectionOptions{
		Host:     "redis:6379",
		Password: "",
		DB:       0,
	}, dispatcher)
	defer bus.Close()
	handlers := Handlers{}
	dispatcher.RegisterHandlers(&handlers)
	handlers.On("HandleUserCreatedEvent", UserCreatedEvent{"John"}).Once()
	bus.PublishMessages([]cqrs.Message{
		cqrs.NewMessage(UserCreatedEvent{"John"})})
	Eventually(t, func() bool {
		return handlers.Name == "John"
	}, 5*time.Second, 50*time.Millisecond)
}

// func ExampleInMemoryBus() {
//
// 	bus := cqrs.NewRedisBus()
// 	defer bus.Close()
//
// 	// bus2 := cqrs.NewRedisBus()
// 	// defer bus2.Close()
// 	//
// 	// bus2.AddGlobalHandler(func(event interface{}) {
// 	// 	fmt.Println("Received:", event.(UserCreatedEvent).Name)
// 	// })
//
// 	handlers := Handlers{}
// 	bus.RegisterHandlers(&handlers)
// 	bus.Publish(UserCreatedEvent{"John"})
//
// 	// Output:
// 	// User created: John
// 	// Received: John
//
// 	time.Sleep(2 * time.Second)
//
// }
//
// func ExampleGlobalHandlers() {
// 	bus := cqrs.NewInMemoryBus()
// 	defer bus.Close()
// 	bus.AddGlobalHandler(func(e interface{}) {
// 		fmt.Printf("[LOG] %T: %v\n", e, e)
// 	})
// 	bus.Publish(UserCreatedEvent{"John"})
// 	bus.Publish(UserNameChangedEvent{"Mark"})
// 	// Output:
// 	// [LOG] cqrs_test.UserCreatedEvent: {John}
// 	// [LOG] cqrs_test.UserNameChangedEvent: {Mark}
// }
//
// type User struct {
// 	Name string
// }
//
// func (u *User) HandleUserCreatedEvent(e UserCreatedEvent) {
// 	u.Name = e.Name
// }
//
// func (u *User) HandleUserNameChanchedEvent(e UserNameChangedEvent) {
// 	u.Name = e.NewName
// }
//
// func TestAggregateRoot(t *testing.T) {
//
// 	user := User{}
// 	root := cqrs.NewAggregateRoot(&user)
//
// 	root.Source(UserCreatedEvent{"John"})
// 	if user.Name != "John" {
// 		t.Fail()
// 	}
//
// 	root.Source(UserNameChangedEvent{"Mark"})
// 	if user.Name != "Mark" {
// 		t.Fail()
// 	}
//
// 	if len(root.Changes) != 2 {
// 		t.Fail()
// 	}
//
// 	root.ClearChanges()
// 	if len(root.Changes) != 0 {
// 		t.Fail()
// 	}
//
// }
//
// func TestAggregateRootFromHistory(t *testing.T) {
//
// 	user := User{}
// 	history := []cqrs.Event{
// 		cqrs.NewEvent(UserCreatedEvent{"John"}),
// 		cqrs.NewEvent(UserNameChangedEvent{"Mark"})}
//
// 	root := cqrs.NewAggregateRootFromHistory(&user, history)
//
// 	if user.Name != "Mark" {
// 		t.Fatal("User name not up to date", user.Name)
// 	}
//
// 	if len(root.Changes) != 0 {
// 		t.Fail()
// 	}
//
// }
