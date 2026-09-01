#!/usr/bin/env bash
set -e
DIR="$(cd "$(dirname "$0")" && pwd)"
exec go run "$DIR/cmd/seed4me" "$@"
