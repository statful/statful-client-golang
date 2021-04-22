package statful

import (
	"encoding/json"
	"io"
	"net/http"
)

const (
	jsonEncoding = "application/json"
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

func NewEvent(eventId string, userId string, extUserId string, gameId string, operatorId string, aggregatorId string, eventType string, amount int, currency string, attributes []Attribute, timestamp int) Event {
	return Event{
		EventId:      eventId,
		UserId:       userId,
		ExtUserId:    extUserId,
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

func (c *Client) Event(event Event) {
	c.eventBuffer.Event(event)
}

func (c *Client) FlushEvents() error {
	return c.eventBuffer.Send()
}

func (h *HttpSender) SendEvent(data io.Reader) error {
	url := h.Url + h.BasePath

	return h.do(http.MethodPut, url, jsonEncoding, data)
}
