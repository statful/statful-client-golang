package statful

import (
	"testing"
)

const (
	defaultUuid   = "00000000-0000-0000-0000-000000000000"
	testAmount    = 5000
	testEventType = "event type test"
	testCurrency  = "PT"
	testEpoch     = 1319010256
)

const expectedJson = `{"eventId":"","userId":"00000000-0000-0000-0000-000000000000","extUserId":"","gameId":"00000000-0000-0000-0000-000000000000","operatorId":"00000000-0000-0000-0000-000000000000","aggregatorId":"00000000-0000-0000-0000-000000000000","eventType":"event type test","amount":{"value":5000,"currency":"PT"},"variableAttributes":[],"timestamp":1319010256}`

func TestNewEvent(t *testing.T) {

	event := NewEvent(defaultUuid, defaultUuid, defaultUuid, defaultUuid, testEventType, testAmount, testCurrency, []Attribute{}, testEpoch)

	expectedEvent := Event{
		UserId:       defaultUuid,
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
		t.Errorf("Error returning while parsing to Json: %s ", err)
	}

	if expectedJson != jsonString {
		t.Errorf("Wrong jsonString returned: \nExpected: %s \nGot: %s ", expectedJson, jsonString)
	}
}
