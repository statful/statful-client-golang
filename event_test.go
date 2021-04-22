package statful

import (
	"fmt"
	"io"
	"io/ioutil"
	"testing"
)

const (
	defaultUuid   = "00000000-0000-0000-0000-000000000000"
	testAmount    = 5000
	testEventType = "event type test"
	testExtUserId = "20"
	testCurrency  = "PT"
	testEpoch     = 1319010256

	expectedJson = `[{"eventId":"00000000-0000-0000-0000-000000000000","userId":"00000000-0000-0000-0000-000000000000","extUserId":"20","gameId":"00000000-0000-0000-0000-000000000000","operatorId":"00000000-0000-0000-0000-000000000000","aggregatorId":"00000000-0000-0000-0000-000000000000","eventType":"event type test","amount":{"value":5000,"currency":"PT"},"variableAttributes":[],"timestamp":1319010256}]`
)

var expectedEvent = Event{
	EventId:      defaultUuid,
	UserId:       defaultUuid,
	ExtUserId:    testExtUserId,
	GameId:       defaultUuid,
	OperatorId:   defaultUuid,
	AggregatorId: defaultUuid,
	EventType:    testEventType,
	Amount: Amount{
		Value:    testAmount,
		Currency: testCurrency,
	},
	VariableAttributes: []Attribute{},
	Timestamp:          testEpoch,
}

func TestNewEvent(t *testing.T) {

	event := NewEvent(defaultUuid, defaultUuid, testExtUserId, defaultUuid, defaultUuid, defaultUuid, testEventType, testAmount, testCurrency, []Attribute{}, testEpoch)

	if event.UserId != expectedEvent.UserId {
		t.Errorf("Different userId returned: \nExpected: %s \nGot: %s ", expectedEvent.UserId, event.UserId)
	}

	if event.GameId != expectedEvent.GameId {
		t.Errorf("Different gameId returned: \nExpected: %s \nGot: %s ", expectedEvent.GameId, event.GameId)
	}

	if event.OperatorId != expectedEvent.OperatorId {
		t.Errorf("Different operatorId returned: \nExpected: %s \nGot: %s ", expectedEvent.OperatorId, event.OperatorId)
	}

	if event.AggregatorId != expectedEvent.AggregatorId {
		t.Errorf("Different aggregatorId returned: \nExpected: %s \nGot: %s ", expectedEvent.AggregatorId, event.AggregatorId)
	}

	if event.EventType != expectedEvent.EventType {
		t.Errorf("Different eventType returned: \nExpected: %s \nGot: %s ", expectedEvent.EventType, event.EventType)
	}

	if event.Amount != expectedEvent.Amount {
		t.Errorf("Different amount returned: \nExpected: %q \nGot: %q ", expectedEvent.Amount, event.Amount)
	}

	if event.Timestamp != expectedEvent.Timestamp {
		t.Errorf("Different timestamp returned: \nExpected: %d \nGot: %d ", expectedEvent.Timestamp, event.Timestamp)
	}

	jsonString, err := event.toJson()

	if err != nil {
		t.Errorf("Error returned while parsing to Json: %s ", err)
	}

	jsonString = fmt.Sprintf("[%s]", jsonString)

	if expectedJson != jsonString {
		t.Errorf("Wrong jsonString returned: \nExpected: %s \nGot: %s ", expectedJson, jsonString)
	}
}

func (c *ChannelSender) SendEvent(data io.Reader) error {
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
		buffer: []Event{},

		Sender: &ChannelSender{
			data: eventData,
		},
	}

	length := len(testBuffer.buffer)
	if length != 0 {
		t.Errorf("Expected buffer current length error: Expected: %d, Got: %d ", 0, length)
	}

	testBuffer.Event(expectedEvent)
	length = len(testBuffer.buffer)
	if length != 1 {
		t.Errorf("Expected buffer current length error: Expected: %d, Got: %d ", 1, length)
	}

	err := testBuffer.Send()
	if err != nil {
		t.Errorf("Error returned sending event: %s ", err)
	}

	length = len(testBuffer.buffer)
	if length != 0 {
		t.Errorf("Expected buffer current length error: Expected: %d, Got: %d ", 0, length)
	}
}
