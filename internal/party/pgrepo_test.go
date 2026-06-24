//go:build integration

package party

import (
	"context"
	"os"
	"testing"

	"github.com/sreeni/openbank-bian/pkg/pg"
	"github.com/sreeni/openbank-bian/pkg/testutil"
)

// newPgRepo spins up a throwaway Postgres, applies migrations (which seed the
// demo parties) and returns a Postgres-backed repository. Migrations are read
// from the module's migrations directory relative to this test package.
func newPgRepo(t *testing.T) *PgRepository {
	t.Helper()
	ctx := context.Background()
	dsn := testutil.PostgresDSN(t)

	pool, err := pg.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := pg.RunMigrations(ctx, pool, os.DirFS("../.."), "migrations", "party"); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return NewPgRepository(pool)
}

func TestPgRepositoryGetPrimaryParty(t *testing.T) {
	ctx := context.Background()
	repo := newPgRepo(t)

	p, err := repo.GetPrimaryParty(ctx)
	if err != nil {
		t.Fatalf("get primary party: %v", err)
	}
	if p.PartyID != "PSU-001" || !p.Primary {
		t.Fatalf("unexpected primary party %+v", p)
	}
	if p.FullLegalName != "Mr Kelvin C Smith" || p.PostCode != "EC2R 8AH" || p.Country != "GB" {
		t.Fatalf("unexpected primary party fields %+v", p)
	}
}

func TestPgRepositoryGetParty(t *testing.T) {
	ctx := context.Background()
	repo := newPgRepo(t)

	p, err := repo.GetParty(ctx, "PARTY-002")
	if err != nil {
		t.Fatalf("get party: %v", err)
	}
	if p.PartyID != "PARTY-002" || p.Primary {
		t.Fatalf("unexpected party %+v", p)
	}
	if p.PartyType != PartyTypeIndividual {
		t.Fatalf("party type = %s", p.PartyType)
	}
}

func TestPgRepositoryListParties(t *testing.T) {
	ctx := context.Background()
	repo := newPgRepo(t)

	parties, err := repo.ListParties(ctx)
	if err != nil {
		t.Fatalf("list parties: %v", err)
	}
	if len(parties) != 2 {
		t.Fatalf("len = %d, want 2", len(parties))
	}
	// ListParties orders by id, so PARTY-002 precedes PSU-001.
	if parties[0].PartyID != "PARTY-002" || parties[1].PartyID != "PSU-001" {
		t.Fatalf("unexpected order %+v", parties)
	}
}

func TestPgRepositoryGetMissing(t *testing.T) {
	repo := newPgRepo(t)
	if _, err := repo.GetParty(context.Background(), "nope"); err != ErrNotFound {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
