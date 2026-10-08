# Releasing

Only maintainers with access to pushing tags are able to perform releases. If
changes have been merged into the master branch but a release has not yet been
scheduled, you can contact one of the maintainers to request and plan the
release.

## Binaries and docker images

In order to make a new release push a semver tag eg. `v0.1.0`. If you want to
publish a prerelase, simply do a prerelease tag eg. `v0.2.0-rc.1`.

This will run [goreleaser](https://goreleaser.com/) according to the
configuration in `.goreleaser.yaml`.

The release workflow reads the Go version from `go.mod`. GoReleaser v2.18.0
requires Go 1.27.0 or newer. For local releases, `make release` installs the
GoReleaser version pinned in `go.mod`.
