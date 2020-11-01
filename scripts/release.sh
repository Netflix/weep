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

if ! command -v svu &> /dev/null
then
    echo "Please install svu:"
    echo "go get github.com/caarlos0/svu"
    exit
fi

if [ "$1" = "minor" ]; then
  NEXT_VERSION=$(svu minor)
else
  NEXT_VERSION=$(svu patch)
fi

if [ ! "$(git branch --show-current)" = "master" ]; then
  echo "Not on default branch, exiting"
  exit 1
else
  echo "Creating and pushing tag for $NEXT_VERSION"
fi

git pull origin master
git tag -am "$NEXT_VERSION" "$NEXT_VERSION"
git push origin "$NEXT_VERSION"
