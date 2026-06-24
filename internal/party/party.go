// Package party implements the BIAN "Party Reference Data Management" service
// domain and the OBIE Party resource. It owns the reference data describing the
// PSU (Payment Service User) — name, contact details and address — and serves a
// single primary PSU party at GET /party plus individually addressable parties.
// Access to the public OBIE endpoints is gated by an account-access consent
// carrying the ReadParty permission; the consent itself lives in the consent
// service, which this service calls to validate each request.
package party

// PartyType enumerates the OBIE party kinds this service models. The estate only
// deals with personal customers, so a party is either a "Sole" trader or an
// "Individual".
type PartyType string

const (
	PartyTypeSole       PartyType = "Sole"
	PartyTypeIndividual PartyType = "Individual"
)

// Party is the aggregate root: the OBIE Party resource. It is reference data, so
// it has no lifecycle of its own beyond being read. Exactly one party in the
// store is marked Primary — that is the PSU returned by GET /party.
type Party struct {
	PartyID       string
	PartyType     PartyType
	Name          string
	FullLegalName string
	EmailAddress  string
	Phone         string

	// A simple single postal address. OBIE models a richer PostalAddress, but
	// the estate only needs these three lines to identify the PSU.
	AddressLine string
	PostCode    string
	Country     string

	// Primary marks the PSU party served by GET /party. Demo data seeds exactly
	// one primary party.
	Primary bool
}
