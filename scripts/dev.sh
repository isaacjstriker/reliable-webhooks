#!/usr/bin/env bash
set -euo pipefail

docker compose up -d
sleep 3
make migrate-up
go run ./cmd/api