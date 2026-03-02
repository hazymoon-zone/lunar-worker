#!/bin/sh
set -eu

if ! command -v golangci-lint >/dev/null 2>&1; then
  echo "golangci-lint-staged: golangci-lint is not installed" >&2
  exit 1
fi

files=$(git diff --cached --name-only --diff-filter=ACMR -- '*.go')
if [ -z "${files}" ]; then
  echo "golangci-lint-staged: no staged Go files"
  exit 0
fi

# Lint only packages touched by staged files.
packages=$(printf '%s\n' "$files" | xargs -n1 dirname | sed 's#^#./#' | sort -u)

echo "golangci-lint-staged: linting staged packages"
golangci-lint run $packages
