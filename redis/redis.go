package redis

import (
	"encoding/json"
	"reflect"

	"github.com/thebookofeveryone/cqrs"
	"gopkg.in/redis.v3"
)

type ConnectionOptions struct {
	Host     string
	Password string
	DB       int64
}

type MessageBus struct {
	channel string
	client  *redis.Client
}

func NewMessageBus(channel string, options ConnectionOptions) *MessageBus {
	bus := MessageBus{channel, redis.NewClient(&redis.Options{
		Addr:     options.Host,
		Password: options.Password,
		DB:       options.DB,
	})}
	return &bus
}

func (b *MessageBus) Close() {
	b.client.Close()
}

func (b *MessageBus) SetChannel(channel string) {
	b.channel = channel
}

func (b *MessageBus) GetChannel() string {
	return b.channel
}

func (b *MessageBus) PublishMessages(messages []cqrs.Message) error {
	for _, message := range messages {
		result, _ := json.Marshal(message)
		err := b.client.Publish(b.channel, string(result)).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *MessageBus) DispatchTo(dispatcher *cqrs.Dispatcher) (chan bool, error) {
	pubsub, err := b.client.Subscribe(b.channel)
	if err != nil {
		return nil, err
	}
	close := make(chan bool)
	messages := make(chan *redis.Message)
	go func() {
		for {
			msg, err := pubsub.ReceiveMessage()
			if err != nil {
				continue
			}
			messages <- msg
		}
	}()
	go func() {
		defer pubsub.Close()
		for {
			select {
			case msg, _ := <-messages:
				var message cqrs.RawMessage
				json.Unmarshal([]byte(msg.Payload), &message)
				messageType := dispatcher.MessageTypes[message.Type]
				messageValue := reflect.New(messageType)
				err = json.Unmarshal(message.Payload, messageValue.Interface())
				dispatcher.Dispatch(messageValue.Elem().Interface())
			case <-close:
				return
			}
		}
	}()
	return close, nil
}
