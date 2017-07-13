bus
============

[![Build Status](https://api.travis-ci.org/larsp/bus.svg?branch=develop)](https://travis-ci.org/larsp/bus) [![GoDoc](https://godoc.org/github.com/larsp/bus?status.svg)](https://godoc.org/github.com/larsp/bus)

Simple worker based publish/subscribe style communication between Go components.

## Example:

```go
package main

import (
	"fmt"
	"sync"

	. "github.com/larsp/bus"
)

type SomeEvent struct {
	Message string
	WG      *sync.WaitGroup
}

func SomeEventHandler(event SomeEvent) {
	fmt.Println(event.Message)
	event.WG.Done()
}

func main() {
	bus := New(10, 2) // channel can buffer 10 events and 2 workers consume from that channel
	var wg sync.WaitGroup
	wg.Add(1)

	bus.Register(SomeEventHandler)
	bus.Publish(SomeEvent{"Hallo!", &wg})
	wg.Wait() // Waiting for the handler to process the event
}
```

For documentation, check [godoc](http://godoc.org/github.com/larsp/bus).
