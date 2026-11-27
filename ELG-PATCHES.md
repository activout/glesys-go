# SDK additions for elg

Based on upstream `master` (commit `dae169a`, after v8.6.0). The module remains `github.com/glesys/glesys-go/v8`.
The original MIT licence and public methods remain intact.

Proposed upstream (reworked after review):

- glesys/glesys-go#145: every response body is closed, reads are bounded at 16 MiB,
  and malformed or empty JSON is rejected when the caller passes no result.
- glesys/glesys-go#146: `Client.SetHTTPClient` and `Login.SetHTTPClient`; the
  default client refuses redirects and `nil` restores it.

Fork-only additions:

- `APIError`: HTTP status, provider status/message and Retry-After, usable with
  `errors.As`. Non-JSON HTTP errors retain their status without exposing bodies.
- `EmailAccountPatch`, `EditAccountPatch`, `CreateAccountWithSettings`:
  preserve explicitly supplied zero and empty settings. The new creation method
  omits the undocumented arbitrary password parameter. The old method remains.
- `CostsDetails`: preserve actual pricing field types and decimal precision.
  Read-only production verification found numeric quota amounts that the old
  string field could not decode.
- `DNSRecordPatch` and `ServerPatch`: preserve explicit clearing during edits.

Not carried (declined upstream in #144 and #147): `Client.Request`/`ValidateEndpoint`
(elg makes raw API calls itself) and overview filter escaping (the API does not
unescape filters; elg rejects filters that would alter the request path).

Run `gofmt`, `go vet ./...`, `go test ./...`, and `go test -race ./...`.
