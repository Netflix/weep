#!/usr/bin/env bash

echo "=> Checking that code complies with gofmt..."

gofmt_files=$(gofmt -l `find . -name '*.go' | grep -v vendor`)
if [[ -n ${gofmt_files} ]]; then
    echo 'Need to run gofmt on the following files:'
    echo "${gofmt_files}"
    echo "Run: \`make fmt\` to reformat code."
    exit 1
fi