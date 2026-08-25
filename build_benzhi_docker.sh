#!/usr/bin/env bash
set -euo pipefail
docker build -f benzhi.Dockerfile -t task245-qecorr:local .
