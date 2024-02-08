#!/bin/bash

set -euo pipefail

go build -o embedmd
go test ./...
