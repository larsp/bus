package bus

import (
	"fmt"
	"reflect"
	"sync"
)

// EventBus interfaces declares supported methods for registering an handle and publishing to the bus
type EventBus interface {
	// Register an event handler
	Register(function interface{}, forTypes ...interface{}) error

	// Publish will push the passed event to a channel. If that channel is full this method will block.
	// In order to avoid blocking the user could wrap Publish into a goroutine.
	Publish(event interface{}) error
}

type eventBus struct {
	// TODO lpf replace with sync.Map when available in go 1.10?
	// guarded by lock
	handlers map[reflect.Type]map[reflect.Value]bool
	lock     sync.RWMutex
	queue    chan interface{}
}

// New creates a new bus instances. Given your workload you need to specify the following parameters:
// queueSize is used in order to create a channel with given size.
// workers creates as many workers to process events from the channel.
func New(queueSize int, workers int) EventBus {
	bus := &eventBus{
		make(map[reflect.Type]map[reflect.Value]bool),
		sync.RWMutex{},
		make(chan interface{}, queueSize),
	}

	for i := 1; i <= workers; i++ {
		go func(worker int) {
			for event := range bus.queue {
				bus.handle(worker, event)
			}
		}(i)
	}
	return bus
}

func (bus *eventBus) Register(fn interface{}, forTypes ...interface{}) error {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	v := reflect.ValueOf(fn)
	def := v.Type()

	if def.NumIn() != 1 {
		return fmt.Errorf("Handler must have a single argument")
	}

	argument := def.In(0)

	for _, typ := range forTypes {
		t := reflect.TypeOf(typ)
		if !t.ConvertibleTo(argument) {
			return fmt.Errorf("Handler argument %v is not compatible with type %v", argument, t)
		}
		bus.addHandler(t, v)
	}

	if len(forTypes) == 0 {
		bus.addHandler(argument, v)
	}
	return nil
}

func (bus *eventBus) Publish(event interface{}) error {
	bus.lock.RLock()
	defer bus.lock.RUnlock()

	t := reflect.TypeOf(event)

	_, ok := bus.handlers[t]
	if !ok {
		return fmt.Errorf("No handler found for Event type '%s'", t)
	}

	bus.queue <- event

	return nil
}

func (bus *eventBus) handle(id int, event interface{}) {
	bus.lock.RLock()
	defer bus.lock.RUnlock()

	t := reflect.TypeOf(event)
	handlers := bus.handlers[t]

	args := [...]reflect.Value{reflect.ValueOf(event)}
	for fn := range handlers {
		fn.Call(args[:])
	}
}

func (bus *eventBus) addHandler(fnType reflect.Type, fn reflect.Value) {
	handlers, ok := bus.handlers[fnType]
	if !ok {
		handlers = make(map[reflect.Value]bool)
	}

	handlers[fn] = true
	bus.handlers[fnType] = handlers
}
