package party

import (
	"context"
	"errors"

	"github.com/sreeni/openbank-bian/pkg/consentcli"
	"github.com/sreeni/openbank-bian/pkg/httpx"
	"github.com/sreeni/openbank-bian/pkg/obie"
)

// permissionReadParty is the OBIE Permissions enum value an account-access
// consent must grant before the PSU's party data may be read.
const permissionReadParty = "ReadParty"

// ConsentClient is the slice of the consent service this domain depends on:
// fetching the internal consent view to authorise a request. *consentcli.Client
// satisfies it, and tests supply a fake.
type ConsentClient interface {
	Get(ctx context.Context, id string) (*consentcli.View, error)
}

// Service holds the party business logic. The repository provides reference
// data; the consent client gates access to it.
type Service struct {
	repo    Repository
	consent ConsentClient
}

// NewService wires a Service to its repository and consent client.
func NewService(repo Repository, consent ConsentClient) *Service {
	return &Service{repo: repo, consent: consent}
}

// GetPrimaryParty returns the PSU party served by GET /party. It is
// consent-gated: consentID is the value of the x-consent-id header and must
// reference an Authorised account-access consent granting ReadParty.
func (s *Service) GetPrimaryParty(ctx context.Context, consentID string) (*Party, error) {
	if err := s.authorise(ctx, consentID); err != nil {
		return nil, err
	}
	p, err := s.repo.GetPrimaryParty(ctx)
	if err != nil {
		return nil, s.mapNotFound(err)
	}
	return p, nil
}

// GetParty returns an individually addressed party. It is consent-gated in the
// same way as GetPrimaryParty.
func (s *Service) GetParty(ctx context.Context, consentID, partyID string) (*Party, error) {
	if err := s.authorise(ctx, consentID); err != nil {
		return nil, err
	}
	p, err := s.repo.GetParty(ctx, partyID)
	if err != nil {
		return nil, s.mapNotFound(err)
	}
	return p, nil
}

// GetPartyInternal returns a party for service-to-service callers. It performs
// NO consent check — callers on the internal network are already trusted — but
// still maps an unknown id to a 404.
func (s *Service) GetPartyInternal(ctx context.Context, partyID string) (*Party, error) {
	p, err := s.repo.GetParty(ctx, partyID)
	if err != nil {
		return nil, s.mapNotFound(err)
	}
	return p, nil
}

// authorise enforces the OBIE consent rules for the public party endpoints:
// the x-consent-id header must be present and must reference an Authorised
// account-access consent that grants the ReadParty permission.
func (s *Service) authorise(ctx context.Context, consentID string) error {
	if consentID == "" {
		return httpx.Unauthorized("Missing x-consent-id header",
			httpx.Detail(obie.ErrHeaderMissing, "x-consent-id header is required", ""))
	}

	view, err := s.consent.Get(ctx, consentID)
	if err != nil {
		if errors.Is(err, consentcli.ErrNotFound) {
			return httpx.Forbidden("Consent is not valid for this request",
				httpx.Detail(obie.ErrResourceInvalid, "no such consent", ""))
		}
		return httpx.Internal("could not validate consent")
	}

	if view.Type != consentcli.TypeAccountAccess {
		return httpx.Forbidden("Consent is not valid for this request",
			httpx.Detail(obie.ErrResourceInvalid, "consent is not an account-access consent", ""))
	}
	if view.Status != consentcli.StatusAuthorised {
		return httpx.Forbidden("Consent is not valid for this request",
			httpx.Detail(obie.ErrResourceInvalid, "consent is not Authorised", ""))
	}
	if !view.HasPermission(permissionReadParty) {
		return httpx.Forbidden("Consent does not grant ReadParty",
			httpx.Detail(obie.ErrResourceInvalid, "ReadParty permission is required", ""))
	}
	return nil
}

func (s *Service) mapNotFound(err error) error {
	if errors.Is(err, ErrNotFound) {
		return httpx.NotFound("Party not found",
			httpx.Detail(obie.ErrResourceNotFound, "no such party", ""))
	}
	return httpx.Internal("could not load party")
}
