package audit

import "time"

type Event struct {
	ActorID   string
	Action    string
	Resource  string
	CreatedAt time.Time
}

func NewEvent(actorID string, action string, resource string) Event {
	return Event{
		ActorID:   actorID,
		Action:    action,
		Resource:  resource,
		CreatedAt: time.Now().UTC(),
	}
}
