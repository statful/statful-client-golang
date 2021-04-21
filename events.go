package statful

import (
	"encoding/json"
)

type Amount struct {
	Value    int    `json:"value"`
	Currency string `json:"currency"`
}

type Attribute struct {
	Attribute string `json:"attribute"`
	Value     string `json:"value"`
}

type Event struct {
	EventId            string      `json:"eventId"`
	UserId             string      `json:"userId"`
	ExtUserId          string      `json:"extUserId"`
	GameId             string      `json:"gameId"`
	OperatorId         string      `json:"operatorId"`
	AggregatorId       string      `json:"aggregatorId"`
	EventType          string      `json:"eventType"`
	Amount             Amount      `json:"amount"`
	VariableAttributes []Attribute `json:"variableAttributes"`
	Timestamp          int         `json:"timestamp"`
}

func NewEvent(userId string, gameId string, operatorId string, aggregatorId string, eventType string, amount int, currency string, attributes []Attribute, timestamp int) Event {
	return Event{
		UserId:       userId,
		GameId:       gameId,
		OperatorId:   operatorId,
		AggregatorId: aggregatorId,
		EventType:    eventType,
		Amount: Amount{
			Value:    amount,
			Currency: currency,
		},
		VariableAttributes: attributes,
		Timestamp:          timestamp,
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
