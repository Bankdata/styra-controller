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
requires Go 1.27.0 or newer. `make release` is intended only for the release
pipeline, which installs GoReleaser at the version pinned in `go.mod` and
configures Docker Buildx through `docker/setup-buildx-action`.

Docker releases use GoReleaser's `dockers_v2` pipeline and require a Docker
Buildx builder supporting `linux/amd64` and `linux/arm64`. The Dockerfile copies
the pre-built binary from the matching platform directory in GoReleaser's
build context; local `make docker-build` and `make podman-build` override the
binary path with the `BINARY` build argument.

Releases publish multi-platform images tagged `latest`, the major version,
the major/minor version, and the full version. The full-version `-amd64` and
`-arm64` tags remain available for architecture-specific consumers.

With `dockers_v2`, image building and pushing happen together during the publish
phase. `goreleaser build` and `goreleaser release --skip=publish` do not build
images.
