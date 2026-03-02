#!/bin/sh
set -eu

files=$(git diff --cached --name-only --diff-filter=ACMR -- '*.go')
if [ -z "${files}" ]; then
  echo "gofmt-staged: no staged Go files"
  exit 0
fi

# Format only staged Go files.
printf '%s\n' "$files" | xargs gofmt -w

# Block commit if formatter changed files, so developer can review and stage manually.
if ! printf '%s\n' "$files" | xargs git diff --quiet --; then
  echo "Commit blocked: gofmt changed files."
  exit 1
fi

echo "gofmt-staged: no formatting changes"
