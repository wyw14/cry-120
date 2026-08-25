package journal

import (
	"encoding/json"
	"time"

	"github.com/wyw14/cry-120/internal/model"
)

type Event struct {
	ID        model.Identity  `json:"id"`
	Kind      string          `json:"kind"`
	Aggregate string          `json:"aggregate"`
	Revision  model.Revision  `json:"revision"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
}

func NewEvent(kind, aggregate string, revision model.Revision, payload any, now time.Time) (Event, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return Event{}, err
	}
	return Event{ID: model.NewIdentity("event"), Kind: kind, Aggregate: aggregate, Revision: revision, Payload: encoded, CreatedAt: now.UTC()}, nil
}

func (e Event) Decode(target any) error {
	return json.Unmarshal(e.Payload, target)
}

func (e Event) Valid() bool {
	return !e.ID.Empty() && e.Kind != "" && e.Aggregate != "" && !e.CreatedAt.IsZero()
}
