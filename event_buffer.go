package statful

import (
	"bytes"
	"encoding/json"
	"sync"
)

type eventBuffer struct {
	buffer     []Event
	dryRun     bool
	eventCount int
	flushSize  int
	mu         sync.Mutex

	Logger Logger
	Sender Sender
}

func (e *eventBuffer) Event(event Event) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.buffer = append(e.buffer, event)
	e.eventCount++

}

func (e *eventBuffer) Flush() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	events := e.drainBuffers()

	return e.flushBuffers(events)
}

func (e *eventBuffer) flushBuffers(buffer []Event) error {
	if len(buffer) > 0 {
		if e.dryRun {
			for _, event := range buffer {
				e.Logger.Println("Dry event: ", event)
			}
		} else {
			events, err := json.Marshal(buffer)
			if err != nil {
				return err
			}
			err = e.Sender.SendEvents(bytes.NewBuffer(events))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *eventBuffer) drainBuffers() []Event {
	var events []Event

	if e.eventCount > 0 {
		events = e.buffer
		e.buffer = []Event{}
		e.eventCount = 0
	}

	return events
}
