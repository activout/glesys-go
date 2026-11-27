# Agent guide: activout/glesys-go

This is a fork of [glesys/glesys-go](https://github.com/glesys/glesys-go) used by
[activout/elg](https://github.com/activout/elg). The module path stays
`github.com/glesys/glesys-go/v8`; elg points at this fork with a `replace` directive.
Our additions are listed in `ELG-PATCHES.md`.

## Branches

- `elg-support` is the default branch and the one elg consumes: upstream plus our patch.
- `master` mirrors `upstream/master` exactly. Never commit to it; fast-forward it only.
- `elg/*` branches are the same patches split into small PRs for upstream. Keep them
  rebased on `upstream/master` if they are still open there.

## Syncing with upstream

Remotes: `origin` = activout/glesys-go, `upstream` = glesys/glesys-go.
Work in a worktree on a new branch from fresh `origin/elg-support`, never in the main checkout.

```sh
git fetch --all
git worktree add ../glesys-go-upstream-<ver> -b elg/upstream-<ver> origin/elg-support
cd ../glesys-go-upstream-<ver>
git merge upstream/master -m "Merge upstream <ver>"   # merge, not rebase: elg pins our commits
```

1. Resolve conflicts keeping upstream behaviour and our patches (`patch.go`, `request.go`,
   the `APIError`/HTTP client changes in `client.go`). Patches should stay additive; if
   upstream adds an equivalent, drop ours and note it in `ELG-PATCHES.md`.
2. Update the "Based on upstream release" line in `ELG-PATCHES.md`.
3. Run `gofmt -l .`, `go vet ./...`, `go test ./...`, `go test -race ./...`.
4. Open a PR to `elg-support` and merge it (merge commit, so upstream history is kept).
5. Fast-forward `master`: `git push origin upstream/master:master`.

## Consuming it in elg

elg's `go.mod` has `replace github.com/glesys/glesys-go/v8 => github.com/activout/glesys-go/v8 <pseudo-version>`.
After the PR merges, in the elg repo run:

```sh
go get github.com/glesys/glesys-go/v8@<elg-support merge commit sha>   # then re-point the replace:
go mod edit -replace github.com/glesys/glesys-go/v8=github.com/activout/glesys-go/v8@<sha>
go mod tidy && go vet ./... && go test ./...
```

Check `docs/sdk-coverage.md` in elg for new upstream endpoints worth exposing (v8.6.0 added blockstorage).
Then release elg as described in its AGENTS.md.
