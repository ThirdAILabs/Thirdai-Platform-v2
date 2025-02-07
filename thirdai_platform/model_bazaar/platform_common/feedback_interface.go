package common

import (
	"encoding/json"
	"fmt"
	"time"
)

type EventData struct {
	Event            map[string]interface{} `json:"event"`
	Timestamp        time.Time              `json:"timestamp"`
	PerformRLHFLater bool                   `json:"perform_rlhf_later"`
}

// Because timestamp entry in the JSON is in a custom format, need to implement a custom UnmarshalJSON method
func (e *EventData) UnmarshalJSON(data []byte) error {
	type Alias EventData // alias to prevent infinite recursion
	aux := &struct {
		Timestamp string `json:"timestamp"` // shadowing the Timestamp field with a string
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	/*
		'Alias' is basically struct similar to EventData, but without the UnmarshalJSON method
		type Alias struct {
			Event            map[string]interface{} `json:"event"`
			Timestamp        string             	`json:"timestamp"`
			PerformRLHFLater bool                   `json:"perform_rlhf_later"`
		}
	*/

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Parsing the timestamp with the custom layout
	layout := "02 January 2006 15:04:05"
	parsedTime, err := time.Parse(layout, aux.Timestamp)
	if err != nil {
		return fmt.Errorf("error parsing time: %w", err)
	}

	e.Timestamp = parsedTime
	return nil
}

func UnmarshalFeedbackEvent(jsonStr string) (EventData, error) {
	var eventData EventData

	// Unmarshal the JSON string into the EventData struct
	err := json.Unmarshal([]byte(jsonStr), &eventData)
	if err != nil {
		return EventData{}, fmt.Errorf("error parsing JSON: %w", err)
	}

	return eventData, nil
}
