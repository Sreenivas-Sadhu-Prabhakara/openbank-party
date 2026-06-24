package party

import (
	"context"
	"sort"
	"sync"
)

// MemRepository is an in-memory Repository used by unit and handler tests and
// for running the service without a database. It is safe for concurrent use.
type MemRepository struct {
	mu    sync.RWMutex
	store map[string]Party
}

// NewMemRepository returns an empty in-memory repository.
func NewMemRepository() *MemRepository {
	return &MemRepository{store: make(map[string]Party)}
}

// NewSeededMemRepository returns an in-memory repository pre-populated with the
// demo PSU parties, so the service is useful with zero infrastructure. The seed
// data matches the Postgres migration and the accounts service demo PSU.
func NewSeededMemRepository() *MemRepository {
	r := NewMemRepository()
	for _, p := range seedParties() {
		r.store[p.PartyID] = p
	}
	return r
}

func (r *MemRepository) GetParty(_ context.Context, partyID string) (*Party, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.store[partyID]
	if !ok {
		return nil, ErrNotFound
	}
	out := p
	return &out, nil
}

func (r *MemRepository) GetPrimaryParty(_ context.Context) (*Party, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.store {
		if p.Primary {
			out := p
			return &out, nil
		}
	}
	return nil, ErrNotFound
}

func (r *MemRepository) ListParties(_ context.Context) ([]Party, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Party, 0, len(r.store))
	for _, p := range r.store {
		out = append(out, p)
	}
	// Sort by id so the listing is deterministic.
	sort.Slice(out, func(i, j int) bool { return out[i].PartyID < out[j].PartyID })
	return out, nil
}

// seedParties is the canonical demo dataset, shared with the migration. The
// primary party (PSU-001) is "Mr Kelvin C Smith", matching the accounts service
// demo PSU.
func seedParties() []Party {
	return []Party{
		{
			PartyID:       "PSU-001",
			PartyType:     PartyTypeIndividual,
			Name:          "Kelvin Smith",
			FullLegalName: "Mr Kelvin C Smith",
			EmailAddress:  "kelvin.smith@example.com",
			Phone:         "+44 7700 900123",
			AddressLine:   "1 Bank Street",
			PostCode:      "EC2R 8AH",
			Country:       "GB",
			Primary:       true,
		},
		{
			PartyID:       "PARTY-002",
			PartyType:     PartyTypeIndividual,
			Name:          "Amy Jones",
			FullLegalName: "Ms Amy R Jones",
			EmailAddress:  "amy.jones@example.com",
			Phone:         "+44 7700 900456",
			AddressLine:   "27 High Holborn",
			PostCode:      "WC1V 6AA",
			Country:       "GB",
			Primary:       false,
		},
	}
}
