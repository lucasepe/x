package eventbus_test

import (
	"context"
	"fmt"
	"time"

	"github.com/lucasepe/x/eventbus"
)

const eventExample eventbus.EventID = "example.event"

type exampleEvent struct {
	Message string
}

func (e exampleEvent) EventID() eventbus.EventID {
	return eventExample
}

func ExampleBus_PublishSync() {
	bus := eventbus.New()

	bus.Subscribe(eventExample, func(ctx context.Context, event eventbus.Event) error {
		payload := event.(exampleEvent)
		fmt.Println("sync:", payload.Message)
		return nil
	})

	result := bus.PublishSync(context.Background(), exampleEvent{Message: "hello"})
	fmt.Println("delivered:", result.Delivered)

	// Output:
	// sync: hello
	// delivered: 1
}

func ExampleBus_PublishAsync() {
	bus := eventbus.New(eventbus.WithPublishTimeout(time.Second))

	bus.Subscribe(eventExample, func(ctx context.Context, event eventbus.Event) error {
		payload := event.(exampleEvent)
		fmt.Println("async:", payload.Message)
		return nil
	})

	resultCh := bus.PublishAsync(context.Background(), exampleEvent{Message: "hello"})

	// Qui il chiamante può fare altro lavoro prima di leggere l'esito finale.
	result := <-resultCh
	fmt.Println("delivered:", result.Delivered)

	// Output:
	// async: hello
	// delivered: 1
}
