#!/bin/bash
set -euo pipefail

docker compose down -v
rm -f docker/compose/files/{keys.json,role_id,secret_id}