package bus_test

import (
	. "bus"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO lpf test  ping/pong
// TODO lpf benchmark test

func TestRegister(t *testing.T) {
	bus := New(10, 2)
	error := bus.Register(EventHandler)
	require.NoError(t, error)
}

func TestRegisterWithType(t *testing.T) {
	bus := New(10, 2)
	error := bus.Register(EventHandler, Event{})
	require.NoError(t, error)
}

func TestRegisterMultiple(t *testing.T) {
	bus := New(10, 2)
	error := bus.Register(EventHandler)
	require.NoError(t, error)

	error = bus.Register(EventHandler)
	require.NoError(t, error)
}

func TestRegisterInvalidSignature(t *testing.T) {
	bus := New(10, 2)
	var wrong func(int, int)

	error := bus.Register(wrong)
	require.Error(t, error, "Handler must have a single argument")
}

func TestRegisterInvalidEvent(t *testing.T) {
	bus := New(10, 2)
	error := bus.Register(EventHandler, "")
	require.Error(t, error, "Handler argument bus_test.Event is not compatible with type string")
}

func TestHandlerNotFound(t *testing.T) {
	bus := New(10, 2)
	bus.Register(EventHandler)
	error := bus.Publish("")
	require.Error(t, error, "No handler found for Event type 'string'")
}

func TestHandle(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	bus := New(10, 2)
	bus.Register(EventHandler)
	counter := int32(0)
	bus.Publish(Event{"Hallo", &wg, &counter})
	wg.Wait()
	assert.Equal(t, int32(1), counter)
}

func TestDuplicateRegistration(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	bus := New(10, 2)
	bus.Register(EventHandler)
	bus.Register(EventHandler)
	counter := int32(0)
	bus.Publish(Event{"Hallo", &wg, &counter})
	wg.Wait()
	assert.Equal(t, int32(1), counter)
}

func TestHandleMultiplex(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)
	bus := New(10, 2)
	bus.Register(EventHandler)
	bus.Register(EventHandler2)
	counter := int32(0)
	bus.Publish(Event{"Hallo", &wg, &counter})
	wg.Wait()
	assert.Equal(t, int32(2), counter)
}

func TestHandleALot(t *testing.T) {
	iterations := 10000

	bus := New(100, 1000)
	bus.Register(EventHandler)

	var wg sync.WaitGroup
	wg.Add(iterations)

	counter := int32(0)

	for i := 0; i < iterations; i++ {
		go func() {
			bus.Publish(Event{"Hallo", &wg, &counter})
		}()
	}
	wg.Wait()
	assert.Equal(t, int32(iterations), counter)
}

type Event struct {
	Message string
	WG      *sync.WaitGroup
	Counter *int32
}

func EventHandler(event Event) {
	time.Sleep(10 * time.Millisecond)
	atomic.AddInt32(event.Counter, 1)
	event.WG.Done()
}

func EventHandler2(event Event) {
	time.Sleep(10 * time.Millisecond)
	atomic.AddInt32(event.Counter, 1)
	event.WG.Done()
}
