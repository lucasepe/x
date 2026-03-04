package eventbus

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	eventSolarEclipse = EventID("solar_eclipse")
	eventMoonEclipse  = EventID("moon_eclipse")
)

type solarEclipseEvent struct {
	duration time.Duration
}

func (e *solarEclipseEvent) EventID() EventID {
	return eventSolarEclipse
}

type moonEclipseEvent struct {
	duration time.Duration
}

func (e *moonEclipseEvent) EventID() EventID {
	return eventMoonEclipse
}

func TestBus_SubscribePublish(t *testing.T) {
	bus := New()
	hadEvent := false
	duration := 100 * time.Second

	bus.Subscribe(eventSolarEclipse, func(ctx context.Context, e Event) error {
		assert.Equal(t, e.EventID(), eventSolarEclipse)
		se := e.(*solarEclipseEvent)
		assert.Equal(t, se.duration, duration)
		hadEvent = true
		return nil
	})
	bus.Subscribe(eventMoonEclipse, func(ctx context.Context, e Event) error {
		t.Fatalf("should never be called")
		return nil
	})
	assert.Equal(t, hadEvent, false)

	result := bus.PublishSync(context.Background(), &solarEclipseEvent{
		duration: duration,
	})
	assert.NoError(t, result.Err)
	assert.Empty(t, result.Errors)
	assert.Equal(t, hadEvent, true)
}

func TestBus_PublishIncompatibleEvent(t *testing.T) {
	bus := New()
	duration := 100 * time.Second
	bus.Subscribe(eventMoonEclipse, func(ctx context.Context, e Event) error {
		t.Fatalf("should never be called")
		return nil
	})

	result := bus.PublishSync(context.Background(), &solarEclipseEvent{
		duration: duration,
	})
	assert.NoError(t, result.Err)
	assert.Zero(t, result.Delivered)
}

func TestBus_SubscribeUnsubscribe(t *testing.T) {
	bus := New()
	hadEvent := false
	duration := 42 * time.Millisecond

	id := bus.Subscribe(eventMoonEclipse, func(ctx context.Context, e Event) error {
		assert.Equal(t, e.EventID(), eventMoonEclipse)
		se := e.(*moonEclipseEvent)
		assert.Equal(t, se.duration, duration)
		hadEvent = true
		return nil
	})
	result := bus.PublishSync(context.Background(), &moonEclipseEvent{
		duration: duration,
	})
	assert.NoError(t, result.Err)
	assert.Equal(t, hadEvent, true)

	hadEvent = false
	bus.Unsubscribe(id)
	result = bus.PublishSync(context.Background(), &moonEclipseEvent{
		duration: duration,
	})
	assert.NoError(t, result.Err)
	assert.Equal(t, hadEvent, false)
}

func TestBus_SubscribeMultiple(t *testing.T) {
	moonEventCount := 0
	moonEclipseDuration := 16 * time.Second
	onMoonEclipse := func(ctx context.Context, e Event) error {
		assert.Equal(t, e.EventID(), eventMoonEclipse)
		moonEventCount++
		se := e.(*moonEclipseEvent)
		assert.Equal(t, se.duration, moonEclipseDuration)
		return nil
	}

	solarEventCount := 0
	solarEclipseDuration := 77 * time.Millisecond
	onSolarEclipse := func(ctx context.Context, e Event) error {
		assert.Equal(t, e.EventID(), eventSolarEclipse)
		solarEventCount++
		se := e.(*solarEclipseEvent)
		assert.Equal(t, se.duration, solarEclipseDuration)
		return nil
	}

	bus := New()
	publishMoon := func() {
		bus.PublishSync(context.Background(), &moonEclipseEvent{
			duration: moonEclipseDuration,
		})
	}
	publishSolar := func() {
		bus.PublishSync(context.Background(), &solarEclipseEvent{
			duration: solarEclipseDuration,
		})
	}

	id1 := bus.Subscribe(eventMoonEclipse, onMoonEclipse)
	id2 := bus.Subscribe(eventSolarEclipse, onSolarEclipse)
	id3 := bus.Subscribe(eventMoonEclipse, onMoonEclipse)

	publishMoon()
	assert.Equal(t, moonEventCount, 2)
	assert.Equal(t, solarEventCount, 0)

	publishSolar()
	assert.Equal(t, moonEventCount, 2)
	assert.Equal(t, solarEventCount, 1)

	bus.Unsubscribe(id1)
	bus.Unsubscribe(id2)
	publishMoon()
	publishSolar()
	assert.Equal(t, moonEventCount, 3)
	assert.Equal(t, solarEventCount, 1)

	bus.Unsubscribe(id3)
	publishMoon()
	publishSolar()
	assert.Equal(t, moonEventCount, 3)
	assert.Equal(t, solarEventCount, 1)
}
func TestBus_PublishRecursive(t *testing.T) {
	moonEventCount := 0

	bus := New()
	publishMoon := func(duration time.Duration) {
		bus.PublishSync(context.Background(), &moonEclipseEvent{
			duration: duration,
		})
	}

	onMoonEclipse := func(ctx context.Context, e Event) error {
		assert.Equal(t, e.EventID(), eventMoonEclipse)
		moonEventCount++
		se := e.(*moonEclipseEvent)

		if se.duration < 16*time.Second {
			publishMoon(2 * se.duration)
		}
		return nil
	}

	bus.Subscribe(eventMoonEclipse, onMoonEclipse)
	publishMoon(1 * time.Second)

	assert.Equal(t, moonEventCount, 5)
}

func TestBus_SubscribeDuringPublishUsesSnapshot(t *testing.T) {
	bus := New()
	var calls atomic.Int32

	bus.Subscribe(eventSolarEclipse, func(ctx context.Context, e Event) error {
		calls.Add(1)
		bus.Subscribe(eventSolarEclipse, func(ctx context.Context, e Event) error {
			calls.Add(1)
			return nil
		})
		return nil
	})

	bus.PublishSync(context.Background(), &solarEclipseEvent{duration: time.Second})
	assert.Equal(t, int32(1), calls.Load())

	bus.PublishSync(context.Background(), &solarEclipseEvent{duration: time.Second})
	assert.Equal(t, int32(3), calls.Load())
}

func TestBus_PublishTimeoutReturns(t *testing.T) {
	release := make(chan struct{})
	bus := New(WithPublishTimeout(20 * time.Millisecond))

	bus.Subscribe(eventMoonEclipse, func(ctx context.Context, e Event) error {
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	started := time.Now()
	result := bus.PublishSync(context.Background(), &moonEclipseEvent{duration: time.Second})
	elapsed := time.Since(started)

	assert.Less(t, elapsed, 100*time.Millisecond)
	assert.ErrorIs(t, result.Err, context.DeadlineExceeded)
	assert.Len(t, result.Errors, 0)
	assert.Equal(t, 1, result.Pending)

	close(release)
}

func TestBus_PublishAsyncReturnsReport(t *testing.T) {
	bus := New()

	bus.Subscribe(eventSolarEclipse, func(ctx context.Context, e Event) error {
		return errors.New("subscriber failed")
	})

	result := <-bus.PublishAsync(context.Background(), &solarEclipseEvent{duration: time.Second})

	assert.NoError(t, result.Err)
	assert.Equal(t, 1, result.Delivered)
	assert.Len(t, result.Errors, 1)
	assert.EqualError(t, result.Errors[0], "subscriber failed")
}

func TestBus_FailureHookReceivesPanic(t *testing.T) {
	var calls atomic.Int32
	var failure HandlerFailure

	bus := New(WithFailureHook(func(f HandlerFailure) {
		calls.Add(1)
		failure = f
	}))

	bus.Subscribe(eventSolarEclipse, func(ctx context.Context, e Event) error {
		panic("boom")
	})

	result := bus.PublishSync(context.Background(), &solarEclipseEvent{duration: time.Second})

	assert.NoError(t, result.Err)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, int32(1), calls.Load())
	assert.Equal(t, eventSolarEclipse, failure.Subscription.eventID)
	assert.Equal(t, uint64(0), failure.Subscription.id)
	assert.Equal(t, "boom", failure.Panic)
	assert.Error(t, failure.Err)
}
