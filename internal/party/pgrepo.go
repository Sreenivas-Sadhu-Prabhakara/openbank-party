package party

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgRepository is the Postgres-backed Repository. The party service owns the
// "party" schema; this type touches nothing outside it.
type PgRepository struct {
	pool *pgxpool.Pool
}

// NewPgRepository returns a Postgres repository over the given pool.
func NewPgRepository(pool *pgxpool.Pool) *PgRepository {
	return &PgRepository{pool: pool}
}

const partyColumns = `party_id, party_type, name, full_legal_name, email_address,
	phone, address_line, post_code, country, primary_party`

func (r *PgRepository) GetParty(ctx context.Context, partyID string) (*Party, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+partyColumns+` FROM party.parties WHERE party_id = $1`, partyID)
	p, err := scanParty(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return p, err
}

func (r *PgRepository) GetPrimaryParty(ctx context.Context) (*Party, error) {
	// Order by id so the result is stable if more than one party is ever marked
	// primary; the migration guarantees a single primary party.
	row := r.pool.QueryRow(ctx,
		`SELECT `+partyColumns+` FROM party.parties WHERE primary_party = true ORDER BY party_id LIMIT 1`)
	p, err := scanParty(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return p, err
}

func (r *PgRepository) ListParties(ctx context.Context) ([]Party, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+partyColumns+` FROM party.parties ORDER BY party_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Party
	for rows.Next() {
		p, err := scanParty(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

// scanParty reads a row in partyColumns order into a Party.
func scanParty(row pgx.Row) (*Party, error) {
	var (
		p   Party
		typ string
	)
	if err := row.Scan(
		&p.PartyID, &typ, &p.Name, &p.FullLegalName, &p.EmailAddress,
		&p.Phone, &p.AddressLine, &p.PostCode, &p.Country, &p.Primary,
	); err != nil {
		return nil, err
	}
	p.PartyType = PartyType(typ)
	return &p, nil
}
