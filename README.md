# openbank-party

The **Party** microservice — the BIAN *Party Reference Data Management* service domain, exposing the OBIE **Party** resource (the PSU and related parties).

The public endpoints are consent-protected: the caller supplies `x-consent-id`, and the service requires an authorised `account-access` consent that includes the `ReadParty` permission.

## Endpoints

| Method | Path | Consent |
|---|---|---|
| GET | `/party` | Authorised account-access consent + `ReadParty` |
| GET | `/parties/{partyId}` | Authorised account-access consent + `ReadParty` |
| GET | `/internal/parties/{partyId}` | _(none — service-to-service)_ |
| GET | `/health` | — |

Missing `x-consent-id` → 401; wrong type / not `Authorised` / missing `ReadParty` → 403; unknown party → 404.

## Configuration

| Env | Default | Notes |
|---|---|---|
| `ADDR` | `:8085` | Listen address |
| `BASE_URL` | `http://localhost:8085` | Used for `Links.Self` |
| `DATABASE_URL` | _(unset)_ | Postgres DSN; **unset → in-memory store** (seeded demo data) |
| `CONSENT_URL` | `http://localhost:8081` | Consent service base URL |

Demo data: primary PSU `PSU-001` (Kelvin Smith) plus `PARTY-002`.

## Run

```bash
go run .                              # in-memory + demo data
docker build -t openbank/party . && docker run -p 8085:8085 openbank/party
curl localhost:8085/party -H "x-consent-id: <consentId>"
```

## Test

```bash
go test ./...                       # unit + handler tests (fake consent client, no Docker)
go test -tags=integration ./...     # Postgres repo tests via testcontainers (needs Docker)
```

## Layout notes

- `internal/party/` — domain, `Repository` port (in-memory + Postgres), service (consent enforcement), OBIE handlers.
- `migrations/` — SQL owned by this service, applied on startup when `DATABASE_URL` is set.
- `pkg/` — vendored shared OBIE library, wired via `replace ... => ./pkg`.
