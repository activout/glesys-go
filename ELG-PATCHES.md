# SDK additions for elg

Based on upstream commit `a25ba1b2ddb53c61a9cb1fc53888c98b9ac0b67e`, following
the official v8.5.0 release. The module remains `github.com/glesys/glesys-go/v8`.
The original MIT licence and public methods remain intact.

Additions:

- `Client.Request`: complete JSON responses through the existing transport,
  validated relative endpoints and scalar GET path parameters.
- `Client.SetHTTPClient`: per-client transport injection; callers configure their
  own redirect policy. The default client refuses redirects.
- `APIError`: HTTP status, provider status/message and Retry-After, usable with
  `errors.As`. Non-JSON HTTP errors retain their status without exposing bodies.
- `EmailAccountPatch`, `EditAccountPatch`, `CreateAccountWithSettings`:
  preserve explicitly supplied zero and empty settings. The new creation method
  omits the undocumented arbitrary password parameter. The old method remains.
- `CostsDetails`: preserve actual pricing field types and decimal precision.
  Read-only production verification found numeric quota amounts that the old
  string field could not decode.
- `DNSRecordPatch` and `ServerPatch`: preserve explicit clearing during edits.

Fixes: every response body is closed, reads are bounded at 16 MiB, malformed or
empty success JSON is rejected, provider failure status is retained even on HTTP
200, and overview filters are escaped. HTTP 200 remains the verified success
status; this patch does not assume all 2xx statuses have interchangeable meaning.

Run `gofmt`, `go vet ./...`, `go test ./...`, and `go test -race ./...`.
Local `httptest` exchanges cover the public additions and transport lifecycle.
No mutation against live resources was used to verify these patches.
