#!/usr/bin/env bash
#
# This script builds the application from source for multiple platforms.
set -e

# This how we want to name the binary output
BINARY="${BINARY_NAME:-"weep"}"
VERSION="${VERSION:-"unknown"}"
VERSION_PRERELEASE="${VERSION_PRERELEASE:-""}"
BUILD_DATE=$(date +%FT%T%z)
MTLS_CONFIG_FILE="${MTLS_CONFIG_FILE:-""}"

# Set build tags
BUILD_TAGS="${BUILD_TAGS:-"weep"}"

# Get the git commit
GIT_COMMIT="$(git rev-parse HEAD)"
GIT_DIRTY="$(test -n "`git status --porcelain`" && echo "+CHANGES" || true)"

rm pkger.go 2&> /dev/null || true

echo "=> Building..."
if [[ ! -z $MTLS_CONFIG_FILE ]]; then
  echo "Bundling mTLS config"
  pkger -include "${MTLS_CONFIG_FILE}"
else
  echo "Not bundling mTLS config"
fi
go build \
    -ldflags "${LD_FLAGS} \
    -X github.com/netflix/weep/mtls.EmbeddedConfigFile=${MTLS_CONFIG_FILE} \
    -X github.com/netflix/weep/version.Version=${VERSION} \
    -X github.com/netflix/weep/version.VersionPrerelease=${VERSION_PRERELEASE} \
    -X github.com/netflix/weep/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} \
    -X github.com/netflix/weep/version.BuildDate=${BUILD_DATE}" \
    -trimpath \
    -o ${BINARY}
