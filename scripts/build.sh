#!/usr/bin/env bash
#
# Copyright 2020 Netflix, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

#
# This script builds the application from source for multiple platforms.
set -e

# This how we want to name the binary output
BINARY="${BINARY_NAME:-"weep"}"
VERSION="${VERSION:-"unknown"}"
VERSION_PRERELEASE="${VERSION_PRERELEASE:-""}"
BUILD_DATE=$(date +%FT%T%z)
EMBEDDED_CONFIG_FILE="${EMBEDDED_CONFIG_FILE:-""}"

# Set build tags
BUILD_TAGS="${BUILD_TAGS:-"weep"}"

# Get the git commit
GIT_COMMIT="$(git rev-parse HEAD)"
GIT_DIRTY="$(test -n "`git status --porcelain`" && echo "+CHANGES" || true)"

rm pkger.go 2&> /dev/null || true

echo "=> Building..."
if [ ! -z "$EMBEDDED_CONFIG_FILE" ]; then
  echo "Bundling config"
  pkger -include "${EMBEDDED_CONFIG_FILE}"
else
  echo "Not bundling config"
fi
go build \
    -ldflags "${LD_FLAGS} \
    -X github.com/netflix/weep/config.EmbeddedConfigFile=${EMBEDDED_CONFIG_FILE} \
    -X github.com/netflix/weep/version.Version=${VERSION} \
    -X github.com/netflix/weep/version.VersionPrerelease=${VERSION_PRERELEASE} \
    -X github.com/netflix/weep/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} \
    -X github.com/netflix/weep/version.BuildDate=${BUILD_DATE}" \
    -trimpath \
    -o ${BINARY}
