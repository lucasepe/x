package eventbus_test

import (
	"context"
	"fmt"

	"github.com/lucasepe/x/eventbus"
)

func ExampleWithFailureHook() {
	bus := eventbus.New(
		eventbus.WithFailureHook(func(f eventbus.HandlerFailure) {
			fmt.Printf("failure: %v\n", f.Err)
		}),
	)

	bus.Subscribe(eventExample, func(ctx context.Context, event eventbus.Event) error {
		return fmt.Errorf("subscriber error")
	})

	result := bus.PublishSync(context.Background(), exampleEvent{Message: "hello"})
	fmt.Println("errors:", len(result.Errors))

	// Output:
	// failure: subscriber error
	// errors: 1
}
