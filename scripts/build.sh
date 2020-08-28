#!/usr/bin/env bash
#
# This script builds the application from source for multiple platforms.
set -e

# This how we want to name the binary output
BINARY="${BINARY_NAME:-"weep"}"
VERSION="${VERSION:-"unknown"}"
VERSION_PRERELEASE="${VERSION_PRERELEASE:-""}"
BUILD_DATE=$(date +%FT%T%z)

# Set build tags
BUILD_TAGS="${BUILD_TAGS:-"weep"}"

# Get the git commit
GIT_COMMIT="$(git rev-parse HEAD)"
GIT_DIRTY="$(test -n "`git status --porcelain`" && echo "+CHANGES" || true)"

echo "=> Building..."
go build \
    -ldflags "${LD_FLAGS} \
    -X github.com/netflix/weep/version.Version=${VERSION} \
    -X github.com/netflix/weep/version.VersionPrerelease=${VERSION_PRERELEASE} \
    -X github.com/netflix/weep/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} \
    -X github.com/netflix/weep/version.BuildDate=${BUILD_DATE}" \
    -trimpath \
    -o ${BINARY}
