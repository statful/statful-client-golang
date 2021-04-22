package statful

import (
	"bytes"
	"encoding/json"
	"sync"
)

type eventBuffer struct {
	buffer     []Event
	eventCount int
	flushSize  int
	mu         sync.Mutex

	Logger Logger
	Sender Sender
}

func (e *eventBuffer) Event(event Event) {
	e.mu.Lock()

	e.buffer = append(e.buffer, event)
	e.eventCount++

	e.mu.Unlock()
}

func (e *eventBuffer) Send() error {
	e.mu.Lock()

	event, err := json.Marshal(e.buffer)
	if err != nil {
		e.mu.Unlock()
		return err
	}

	err = e.Sender.SendEvent(bytes.NewBuffer(event))
	e.buffer = []Event{}
	e.eventCount = 0

	e.mu.Unlock()

	return err
}
