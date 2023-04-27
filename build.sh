#!/bin/bash
set -euo pipefail


docker buildx build \
  --output "type=image,push=false" \
  --platform linux/arm64/v8,linux/amd64 \
  --tag justen-walker/vault-audit:latest .