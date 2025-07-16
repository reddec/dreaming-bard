Thanks for your wish to contribute to the project!

Key moments:

- No CLA: contributing to the project means you agree with the license. It also makes it much harder to change the
  project license later (ie: protects your contribution)
- Avoid huge PRs, but all PRs should be atomic and revertable.
- Most developers usually know how to do things better than others, but please follow "the minimal possible change"
  practice.
- Nothing is dogma: have a better idea? Suggest, but if you have PoC it will be even better.

## Development environment

- latest [Go](https://go.dev/) (see `go.mod` for the exact version)
- [GoReleaser](https://goreleaser.com/)
- [SQLC](https://sqlc.dev/)
- python 3.10+ for code generation
- make, shell, git, git-lfs, sed

Optionally

- [direnv](https://direnv.net/): put your secrets and envs in `.env` file which is ignored by git


## How-to

### Add icon

- Find in https://tabler.io/icons
- Put to `internal/static/img`
- Add ID `icon` to the root element (thanks to iOS)

### Add a new controller

Normally, if controller has collection, then name should be plural.

- From root dir run `./new.py controller <controller name>`
- Add stubs for views
- Generate type hints by `go generate ./...`
- Mount it in `internal/server/server.go`