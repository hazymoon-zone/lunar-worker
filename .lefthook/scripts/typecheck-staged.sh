#!/bin/sh
set -eu

files=$(git diff --cached --name-only --diff-filter=ACMR -- '*.go')
if [ -z "${files}" ]; then
  echo "typecheck-staged: no staged Go files"
  exit 0
fi

# Compile-check only packages touched by staged files.
packages=$(printf '%s\n' "$files" | xargs -n1 dirname | sed 's#^#./#' | sort -u)

echo "typecheck-staged: checking staged packages"
go test -run=^$ $packages
