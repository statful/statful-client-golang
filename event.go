package statful

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
	PublisherId        string      `json:"publisherId"`
	EventType          string      `json:"eventType"`
	Amount             Amount      `json:"amount"`
	VariableAttributes []Attribute `json:"variableAttributes"`
	Timestamp          int         `json:"timestamp"`
}

func NewEvent(eventId string, userId string, extUserId string, gameId string, operatorId string, aggregatorId string, publisherId string, eventType string, amount int, currency string, attributes []Attribute, timestamp int) Event {
	return Event{
		EventId:      eventId,
		UserId:       userId,
		ExtUserId:    extUserId,
		GameId:       gameId,
		OperatorId:   operatorId,
		AggregatorId: aggregatorId,
		PublisherId:  publisherId,
		EventType:    eventType,
		Amount: Amount{
			Value:    amount,
			Currency: currency,
		},
		VariableAttributes: attributes,
		Timestamp:          timestamp,
	}
}

// Add an event to event buffer.
func (c *Client) Event(event Event) {
	c.eventBuffer.Event(event)
}

// Send all events in event buffer.
func (c *Client) FlushEvents() error {
	return c.eventBuffer.Flush()
}