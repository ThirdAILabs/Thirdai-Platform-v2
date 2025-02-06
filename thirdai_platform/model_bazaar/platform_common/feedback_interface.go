package common

import (
	"encoding/json"
	"fmt"
)

type FeedbackEventInterface interface {
	GetAction() string
}

// Upvote event struct
type UpvoteEvent struct {
	Action         string   `json:"action"`
	ChunkIDs       []int    `json:"chunk_ids,omitempty"`
	Queries        []string `json:"queries,omitempty"`
	ReferenceTexts []string `json:"reference_texts,omitempty"`
}

func (e UpvoteEvent) GetAction() string {
	return e.Action
}

// Associate event struct
type AssociateEvent struct {
	Action  string   `json:"action"`
	Sources []string `json:"sources,omitempty"`
	Targets []string `json:"targets,omitempty"`
}

func (e AssociateEvent) GetAction() string {
	return e.Action
}

type EventData struct {
	Event            FeedbackEventInterface `json:"-"`
	Timestamp        string                 `json:"timestamp"`
	PerformRLHFLater bool                   `json:"perform_rlhf_later"`
}

func UnmarshalFeedbackEvent(jsonStr string) (EventData, error) {
	var base struct {
		Event struct {
			Action string `json:"action"`
		} `json:"event"`
		Timestamp        string `json:"timestamp"`
		PerformRLHFLater bool   `json:"perform_rlhf_later"`
	}

	err := json.Unmarshal([]byte(jsonStr), &base)
	if err != nil {
		return EventData{}, fmt.Errorf("error parsing base JSON: %w", err)
	}

	eventData := EventData{
		Timestamp:        base.Timestamp,
		PerformRLHFLater: base.PerformRLHFLater,
	}

	// event data based on action
	switch base.Event.Action {
	case "upvote":
		var wrapper struct {
			Event UpvoteEvent `json:"event"`
		}
		if err := json.Unmarshal([]byte(jsonStr), &wrapper); err != nil {
			return EventData{}, fmt.Errorf("error parsing upvote event: %w", err)
		}
		eventData.Event = wrapper.Event

	case "associate":
		var wrapper struct {
			Event AssociateEvent `json:"event"`
		}
		if err := json.Unmarshal([]byte(jsonStr), &wrapper); err != nil {
			return EventData{}, fmt.Errorf("error parsing associate event: %w", err)
		}
		eventData.Event = wrapper.Event
	}

	return eventData, nil
}
