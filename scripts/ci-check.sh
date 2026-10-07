#!/usr/bin/env bash
set -euo pipefail

echo "==> Running go test ./..."
go test ./...

echo "==> Running go build ./..."
go build ./...

echo "==> All CI checks passed successfully!"
