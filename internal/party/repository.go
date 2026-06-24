package party

import (
	"context"
	"errors"
)

// ErrNotFound is returned by a Repository when no party matches the id.
var ErrNotFound = errors.New("party not found")

// Repository is the persistence port for parties. Both the in-memory and the
// Postgres implementations satisfy it, and the service layer depends only on
// this interface — so the same business-logic tests run against either store.
type Repository interface {
	// GetParty returns the party with the given id, or ErrNotFound.
	GetParty(ctx context.Context, partyID string) (*Party, error)
	// GetPrimaryParty returns the primary PSU party served by GET /party, or
	// ErrNotFound if the store has no primary party.
	GetPrimaryParty(ctx context.Context) (*Party, error)
	// ListParties returns every party held by the service.
	ListParties(ctx context.Context) ([]Party, error)
}
