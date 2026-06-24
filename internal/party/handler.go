package party

import (
	"net/http"

	"github.com/sreeni/openbank-bian/pkg/httpx"
	"github.com/sreeni/openbank-bian/pkg/obie"
)

// consentHeader is the request header carrying the account-access consent id
// that authorises the public party endpoints.
const consentHeader = "x-consent-id"

// Handler exposes the party service over HTTP using OBIE request/response
// shapes. baseURL is used to build absolute Self links.
type Handler struct {
	svc     *Service
	baseURL string
}

// NewHandler constructs the HTTP handler.
func NewHandler(svc *Service, baseURL string) *Handler {
	return &Handler{svc: svc, baseURL: baseURL}
}

// Routes registers every party route on a ServeMux and returns it.
func (h *Handler) Routes() *http.ServeMux {
	mux := http.NewServeMux()

	// Public OBIE Party endpoints (consent-gated).
	mux.HandleFunc("GET /party", h.getPrimaryParty)
	mux.HandleFunc("GET /parties/{partyId}", h.getParty)

	// Internal API used by other services (no consent required).
	mux.HandleFunc("GET /internal/parties/{partyId}", h.internalGetParty)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	return mux
}

func (h *Handler) self(r *http.Request) string { return h.baseURL + r.URL.Path }

// getPrimaryParty handles GET /party, returning the single primary PSU party.
func (h *Handler) getPrimaryParty(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.GetPrimaryParty(r.Context(), r.Header.Get(consentHeader))
	if err != nil {
		httpx.RespondError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK,
		obie.NewResponse(h.self(r), partyData{Party: toDTO(p)}))
}

// getParty handles GET /parties/{partyId}, returning a single Party.
func (h *Handler) getParty(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.GetParty(r.Context(), r.Header.Get(consentHeader), r.PathValue("partyId"))
	if err != nil {
		httpx.RespondError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK,
		obie.NewResponse(h.self(r), partyData{Party: toDTO(p)}))
}

// internalGetParty handles GET /internal/parties/{partyId} for service-to-service
// callers, returning the flat internal projection without any consent check.
func (h *Handler) internalGetParty(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.GetPartyInternal(r.Context(), r.PathValue("partyId"))
	if err != nil {
		httpx.RespondError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toInternalDTO(p))
}
