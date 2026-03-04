# eventbus

`eventbus` is a small in-memory event bus for coordinating internal components inside an application.

It is intentionally generic: it does not know anything about HTTP, Kubernetes, persistence, queues, or external brokers. It only provides topic-based event delivery, concurrency-safe subscriptions, cooperative deadlines, and simple failure reporting.

## What it is for

Use this package when you need lightweight communication between internal parts of the same process, for example:

- application services notifying each other about state changes
- background workers reacting to internal lifecycle events
- HTTP handlers emitting domain events consumed by local components
- CLI tools coordinating independent modules
- discovery loops publishing newly observed resources to listeners

This package is a good fit when:

- all producers and consumers live in the same process
- you do not need durable delivery
- you do not need retries, persistence, or distributed fan-out
- you want a simple coordination mechanism with explicit cancellation

This package is not a replacement for Kafka, NATS, RabbitMQ, or a persistent outbox.

## Core concepts

An event implements:

```go
type Event interface {
	EventID() EventID
}
```

`EventID` is the routing topic. The payload is the event value itself.

Subscribers register a handler for a specific topic:

```go
sub := bus.Subscribe(topic, func(ctx context.Context, event eventbus.Event) error {
	// handle event
	return nil
})
defer bus.Unsubscribe(sub)
```

The handler receives a `context.Context` so it can react to cancellation and deadlines. If the handler returns an error, that error is collected in the publish result. If the handler panics, the panic is recovered, converted to an error, and optionally observed through a failure hook.

## Creating a bus

```go
bus := eventbus.New()
```

Optional behaviors can be configured at creation time:

```go
bus := eventbus.New(
	eventbus.WithPublishTimeout(2*time.Second),
	eventbus.WithFailureHook(func(f eventbus.HandlerFailure) {
		log.Printf("subscriber failed: %v", f.Err)
	}),
)
```

### `WithPublishTimeout`

`WithPublishTimeout` applies a maximum wait time to each publish operation.

Important: this timeout does not forcibly stop already-running handlers. It only bounds how long the publisher waits. To make the timeout effective, handlers must cooperate by observing `ctx.Done()` and returning.

### `WithFailureHook`

`WithFailureHook` lets you observe per-subscriber failures:

- returned errors
- recovered panics

This is useful for logging, metrics, debugging, or integration with your own error reporting.

## Publishing events

The package exposes two explicit publishing modes.

### `PublishSync`

```go
result := bus.PublishSync(ctx, myEvent)
```

`PublishSync` waits until:

- all subscribed handlers complete, or
- the publish context is canceled, or
- the configured bus timeout expires

Use `PublishSync` when:

- the caller needs to know whether delivery finished
- the caller wants the collected handler errors immediately
- the next step depends on publish completion
- you want straightforward control flow

This is usually the right default for application code.

### `PublishAsync`

```go
resultCh := bus.PublishAsync(ctx, myEvent)

// do other work

result := <-resultCh
```

`PublishAsync` starts the delivery immediately and returns a channel that will eventually produce the final `PublishResult`.

Use `PublishAsync` when:

- the caller should not block right away
- you want to overlap publish with other work
- you still want a final result later
- you are orchestrating multiple internal activities concurrently

`PublishAsync` is not fire-and-forget. It is deferred-result delivery. If you truly do not care about the result, you may ignore the returned channel, but then you are also discarding timeout and error information.

## Delivery model and guarantees

Each subscribed handler runs in its own goroutine.

That means:

- one slow subscriber does not prevent other subscribers from starting
- completion order is not deterministic
- handlers for the same event may finish in any order

The bus takes a snapshot of current subscribers before publishing. This matters because subscribers may subscribe or unsubscribe while a publish is already running.

Snapshot-based delivery means:

- a subscriber added during a publish does not receive the current event
- a subscriber removed during a publish may still receive the current event if it was already present in the snapshot

This behavior is usually what you want for predictable concurrent delivery.

## `PublishResult`

Publishing returns:

```go
type PublishResult struct {
	Delivered int
	Pending   int
	Errors    []error
	Err       error
}
```

- `Delivered`: handlers that completed before publish finished
- `Pending`: handlers still running when the publish context ended
- `Errors`: per-handler errors from completed handlers
- `Err`: the overall publish error, typically `context.DeadlineExceeded` or `context.Canceled`

`Err` and `Errors` serve different purposes:

- `Err` describes the publish operation itself
- `Errors` describes subscriber failures

## Practical guidance

Prefer `PublishSync` when:

- the publisher needs deterministic control flow
- the event is part of a request/response path
- you want immediate visibility of delivery errors

Prefer `PublishAsync` when:

- publish latency should not block the current code path immediately
- you can collect the result later
- you are coordinating several concurrent operations

Use a timeout when:

- subscribers are expected to respect deadlines
- the publisher must not wait forever

Do not rely on a timeout alone if handlers ignore the context. Timeouts in this package are cooperative by design.
