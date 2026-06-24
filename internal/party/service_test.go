package party

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/sreeni/openbank-bian/pkg/consentcli"
	"github.com/sreeni/openbank-bian/pkg/httpx"
)

// fakeConsent is a hand-rolled ConsentClient: it returns a canned view (or
// ErrNotFound) so service tests run without the consent service.
type fakeConsent struct {
	view *consentcli.View
	err  error
}

func (f fakeConsent) Get(_ context.Context, _ string) (*consentcli.View, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.view, nil
}

// authorisedView builds a valid account-access view granting the given
// permissions.
func authorisedView(perms ...string) *consentcli.View {
	return &consentcli.View{
		ConsentID:   "consent-1",
		Type:        consentcli.TypeAccountAccess,
		Status:      consentcli.StatusAuthorised,
		Permissions: perms,
	}
}

// newTestService returns a service over a seeded in-memory repo and the given
// consent client.
func newTestService(consent ConsentClient) *Service {
	return NewService(NewSeededMemRepository(), consent)
}

func wantStatus(t *testing.T, err error, status int) {
	t.Helper()
	var apiErr *httpx.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *httpx.APIError, got %v", err)
	}
	if apiErr.Status != status {
		t.Fatalf("status = %d, want %d (%s)", apiErr.Status, status, apiErr.Message)
	}
}

func TestGetPrimaryPartyRequiresConsentHeader(t *testing.T) {
	s := newTestService(fakeConsent{view: authorisedView("ReadParty")})
	// Empty consent id stands in for a missing x-consent-id header.
	_, err := s.GetPrimaryParty(context.Background(), "")
	wantStatus(t, err, http.StatusUnauthorized)
}

func TestGetPrimaryPartyUnknownConsentIsForbidden(t *testing.T) {
	s := newTestService(fakeConsent{err: consentcli.ErrNotFound})
	_, err := s.GetPrimaryParty(context.Background(), "missing")
	wantStatus(t, err, http.StatusForbidden)
}

func TestGetPrimaryPartyWrongTypeIsForbidden(t *testing.T) {
	view := &consentcli.View{
		Type:        consentcli.TypeDomesticPayment,
		Status:      consentcli.StatusAuthorised,
		Permissions: []string{"ReadParty"},
	}
	s := newTestService(fakeConsent{view: view})
	_, err := s.GetPrimaryParty(context.Background(), "consent-1")
	wantStatus(t, err, http.StatusForbidden)
}

func TestGetPrimaryPartyUnauthorisedStatusIsForbidden(t *testing.T) {
	view := &consentcli.View{
		Type:        consentcli.TypeAccountAccess,
		Status:      consentcli.StatusAwaitingAuthorisation,
		Permissions: []string{"ReadParty"},
	}
	s := newTestService(fakeConsent{view: view})
	_, err := s.GetPrimaryParty(context.Background(), "consent-1")
	wantStatus(t, err, http.StatusForbidden)
}

func TestGetPrimaryPartyMissingReadPartyIsForbidden(t *testing.T) {
	s := newTestService(fakeConsent{view: authorisedView("ReadBalances")})
	_, err := s.GetPrimaryParty(context.Background(), "consent-1")
	wantStatus(t, err, http.StatusForbidden)
}

func TestGetPrimaryPartyHappyPath(t *testing.T) {
	s := newTestService(fakeConsent{view: authorisedView("ReadBalances", "ReadParty")})
	p, err := s.GetPrimaryParty(context.Background(), "consent-1")
	if err != nil {
		t.Fatalf("get primary party: %v", err)
	}
	if p.PartyID != "PSU-001" || p.FullLegalName != "Mr Kelvin C Smith" || !p.Primary {
		t.Fatalf("unexpected party %+v", p)
	}
}

func TestGetPartyHappyPathAndUnknown(t *testing.T) {
	ctx := context.Background()
	s := newTestService(fakeConsent{view: authorisedView("ReadParty")})

	p, err := s.GetParty(ctx, "consent-1", "PARTY-002")
	if err != nil {
		t.Fatalf("get party: %v", err)
	}
	if p.PartyID != "PARTY-002" || p.Primary {
		t.Fatalf("unexpected party %+v", p)
	}

	_, err = s.GetParty(ctx, "consent-1", "NOPE")
	wantStatus(t, err, http.StatusNotFound)
}

func TestGetPartyInternalNeedsNoConsentButStill404s(t *testing.T) {
	ctx := context.Background()
	// A consent client that always errors proves the internal path never calls it.
	s := newTestService(fakeConsent{err: consentcli.ErrNotFound})

	p, err := s.GetPartyInternal(ctx, "PSU-001")
	if err != nil {
		t.Fatalf("internal get: %v", err)
	}
	if p.PartyID != "PSU-001" {
		t.Fatalf("unexpected party %+v", p)
	}

	_, err = s.GetPartyInternal(ctx, "NOPE")
	wantStatus(t, err, http.StatusNotFound)
}
