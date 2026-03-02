#!/bin/sh
set -eu

files=$(find . -type f -name '*.go' -not -path './.gocache/*' -not -path './.git/*' | sort)
if [ -z "${files}" ]; then
  echo "gofmt-check: no Go files found"
  exit 0
fi

unformatted=$(printf '%s\n' "$files" | xargs gofmt -l)
if [ -n "${unformatted}" ]; then
  echo "gofmt-check: found unformatted files:"
  printf '%s\n' "$unformatted"
  exit 1
fi

echo "gofmt-check: all Go files are formatted"
