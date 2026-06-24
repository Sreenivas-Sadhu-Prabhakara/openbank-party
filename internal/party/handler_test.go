package party

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sreeni/openbank-bian/pkg/consentcli"
	"github.com/sreeni/openbank-bian/pkg/obie"
)

// newTestHandler builds a handler over a seeded repo and the given consent
// client.
func newTestHandler(consent ConsentClient) http.Handler {
	return NewHandler(NewService(NewSeededMemRepository(), consent), "http://party.test").Routes()
}

// do issues a request to the handler and returns the recorder. consentID, when
// non-empty, is sent as the x-consent-id header.
func do(t *testing.T, h http.Handler, method, path, consentID string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, nil)
	if consentID != "" {
		r.Header.Set(consentHeader, consentID)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func mustDecode(t *testing.T, w *httptest.ResponseRecorder, dst any) {
	t.Helper()
	if err := json.Unmarshal(w.Body.Bytes(), dst); err != nil {
		t.Fatalf("decode body %q: %v", w.Body.String(), err)
	}
}

func TestHandlerGetPrimaryParty(t *testing.T) {
	h := newTestHandler(fakeConsent{view: authorisedView("ReadParty")})

	w := do(t, h, http.MethodGet, "/party", "consent-1")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", w.Code, w.Body)
	}
	var resp struct {
		Data  partyData  `json:"Data"`
		Links obie.Links `json:"Links"`
	}
	mustDecode(t, w, &resp)
	if resp.Data.Party.PartyID != "PSU-001" {
		t.Fatalf("PartyId = %s", resp.Data.Party.PartyID)
	}
	if resp.Data.Party.FullLegalName != "Mr Kelvin C Smith" {
		t.Fatalf("FullLegalName = %s", resp.Data.Party.FullLegalName)
	}
	if resp.Data.Party.Address == nil || resp.Data.Party.Address.PostCode != "EC2R 8AH" {
		t.Fatalf("unexpected address %+v", resp.Data.Party.Address)
	}
	if resp.Links.Self != "http://party.test/party" {
		t.Fatalf("Self = %s", resp.Links.Self)
	}
}

func TestHandlerGetPrimaryPartyMissingHeaderIsUnauthorized(t *testing.T) {
	h := newTestHandler(fakeConsent{view: authorisedView("ReadParty")})

	w := do(t, h, http.MethodGet, "/party", "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body=%s", w.Code, w.Body)
	}
	var errBody obie.ErrorResponse
	mustDecode(t, w, &errBody)
	if errBody.Code != "Unauthorized" || len(errBody.Errors) == 0 {
		t.Fatalf("unexpected error body %+v", errBody)
	}
	if errBody.Errors[0].ErrorCode != obie.ErrHeaderMissing {
		t.Fatalf("error code = %s", errBody.Errors[0].ErrorCode)
	}
}

func TestHandlerGetPartyKnownAndUnknown(t *testing.T) {
	h := newTestHandler(fakeConsent{view: authorisedView("ReadParty")})

	w := do(t, h, http.MethodGet, "/parties/PARTY-002", "consent-1")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", w.Code, w.Body)
	}
	var resp struct {
		Data partyData `json:"Data"`
	}
	mustDecode(t, w, &resp)
	if resp.Data.Party.PartyID != "PARTY-002" {
		t.Fatalf("PartyId = %s", resp.Data.Party.PartyID)
	}

	w = do(t, h, http.MethodGet, "/parties/UNKNOWN", "consent-1")
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body=%s", w.Code, w.Body)
	}
	var errBody obie.ErrorResponse
	mustDecode(t, w, &errBody)
	if errBody.Code != "Not Found" || len(errBody.Errors) == 0 {
		t.Fatalf("unexpected error body %+v", errBody)
	}
}

func TestHandlerInternalGetPartyNeedsNoConsent(t *testing.T) {
	// A consent client that always 404s proves the internal route never calls it.
	h := newTestHandler(fakeConsent{err: consentcli.ErrNotFound})

	w := do(t, h, http.MethodGet, "/internal/parties/PSU-001", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", w.Code, w.Body)
	}
	var dto internalPartyDTO
	mustDecode(t, w, &dto)
	if dto.PartyID != "PSU-001" || !dto.Primary {
		t.Fatalf("unexpected internal party %+v", dto)
	}

	w = do(t, h, http.MethodGet, "/internal/parties/UNKNOWN", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body=%s", w.Code, w.Body)
	}
}
