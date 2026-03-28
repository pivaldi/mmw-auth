package events

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	oglpguow "github.com/ovya/ogl/pg/uow"
	"github.com/pivaldi/mmw-auth/internal/domain/user"
	authdef "github.com/pivaldi/mmw-contracts/definitions/auth"
	"github.com/rotisserie/eris"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// OutboxDispatcher writes domain events to the auth.event outbox table.
// It uses the active transaction from context (via ogl UoW) so events are
// written atomically with the domain record.
type OutboxDispatcher struct {
	uow *oglpguow.UnitOfWork
}

// NewOutboxDispatcher creates a new OutboxDispatcher.
func NewOutboxDispatcher(uow *oglpguow.UnitOfWork) *OutboxDispatcher {
	return &OutboxDispatcher{uow: uow}
}

// Dispatch inserts all events into auth.event in a single batch.
func (d *OutboxDispatcher) Dispatch(ctx context.Context, events []user.DomainEvent) error {
	if len(events) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	const query = `INSERT INTO auth.event (event_type, payload, occurred_at) VALUES ($1, $2::jsonb, $3)`

	for _, evt := range events {
		payload, err := marshalEvent(evt)
		if err != nil {
			return eris.Wrapf(err, "marshal event %s", evt.EventType())
		}
		batch.Queue(query, evt.EventType(), string(payload), evt.OccurredAt())
	}

	exec := d.uow.Executor(ctx)
	br := exec.SendBatch(ctx, batch)
	defer br.Close()

	for i := range events {
		if _, err := br.Exec(); err != nil {
			return eris.Wrapf(err, "insert outbox event at index %d", i)
		}
	}

	return nil
}

// marshalEvent serialises a domain event to JSON for the outbox table.
//
// UserDeleted events are serialised as protojson so the payload field names
// match what proto-aware consumers expect (userId/deletedAt in camelCase).
//
// All other event types fall back to encoding/json via their MarshalJSON methods.
// If a new event type gains a proto-aware consumer, add a branch here — the
// fallback will silently produce wrong field names if left unhandled.
func marshalEvent(evt user.DomainEvent) ([]byte, error) {
	if e, ok := evt.(user.UserDeleted); ok {
		pbEvt := &authdef.UserDeletedEvent{
			UserId:    e.AggregateID(),
			DeletedAt: timestamppb.New(e.OccurredAt()),
		}

		payload, err := protojson.Marshal(pbEvt)

		return payload, eris.Wrap(err, "marshal UserDeletedEvent as protojson")
	}

	payload, err := json.Marshal(evt)

	return payload, eris.Wrap(err, "marshal domain event as json")
}
