#!/bin/bash

set -euo pipefail

prettier --check .
golangci-lint run ./...

GOPLS_OUT=$(git ls-files '*.go' | xargs gopls check -severity hint)
if [[ -n $GOPLS_OUT ]]; then
  echo "$GOPLS_OUT"
  exit 1
fi
