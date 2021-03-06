package statful

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"sync"
	"testing"
)

const (
	defaultUuid   = "00000000-0000-0000-0000-000000000000"
	testAmount    = 5000
	testEventType = "event type test"
	testExtUserId = "20"
	testCurrency  = "PT"
	testEpoch     = 1319010256

	expectedJson = `[{"eventId":"00000000-0000-0000-0000-000000000000","userId":"00000000-0000-0000-0000-000000000000","extUserId":"20","gameId":"00000000-0000-0000-0000-000000000000","operatorId":"00000000-0000-0000-0000-000000000000","aggregatorId":"00000000-0000-0000-0000-000000000000","publisherId":"00000000-0000-0000-0000-000000000000","eventType":"event type test","amount":{"value":5000,"currency":"PT"},"variableAttributes":[],"timestamp":1319010256}]`
)

var expectedEvent = Event{
	EventId:      defaultUuid,
	UserId:       defaultUuid,
	ExtUserId:    testExtUserId,
	GameId:       defaultUuid,
	OperatorId:   defaultUuid,
	AggregatorId: defaultUuid,
	PublisherId:  defaultUuid,
	EventType:    testEventType,
	Amount: Amount{
		Value:    testAmount,
		Currency: testCurrency,
	},
	VariableAttributes: []Attribute{},
	Timestamp:          testEpoch,
}

func (c *ChannelSender) SendEvents(data io.Reader) error {
	all, err := ioutil.ReadAll(data)
	if err != nil {
		return err
	}
	c.data <- all
	return nil
}

func TestFlushEventBuffer(t *testing.T) {
	eventData := make(chan []byte, 1)

	testBuffer := eventBuffer{
		buffer:     []Event{},
		dryRun:     false,
		eventCount: 0,
		mu:         sync.Mutex{},

		Sender: &ChannelSender{
			data: eventData,
		},
	}

	length := len(testBuffer.buffer)
	if length != 0 {
		t.Errorf("Expected buffer current length error: Expected: %d, Got: %d ", 0, length)
	}

	if testBuffer.eventCount != 0 {
		t.Errorf("Expected buffer event count error: Expected: %d, Got: %d ", 0, length)
	}

	testBuffer.Event(expectedEvent)
	length = len(testBuffer.buffer)
	if length != 1 {
		t.Errorf("Expected buffer current length error: Expected: %d, Got: %d ", 1, length)
	}

	if testBuffer.eventCount != 1 {
		t.Errorf("Expected buffer event count error: Expected: %d, Got: %d ", 1, length)
	}

	err := testBuffer.Flush()
	if err != nil {
		t.Errorf("Error returned sending event: %s ", err)
	}

	length = len(testBuffer.buffer)
	if length != 0 {
		t.Errorf("Expected buffer current length error: Expected: %d, Got: %d ", 0, length)
	}

	if testBuffer.eventCount != 0 {
		t.Errorf("Expected buffer event count error: Expected: %d, Got: %d ", 0, length)
	}
}

func (e *Event) toJson() (string, error) {
	bytes, err := json.Marshal(e)
	if err != nil {
		return "", err
	}

	jsonString := string(bytes[:])

	return jsonString, nil
}
