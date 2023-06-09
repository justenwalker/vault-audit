#!/bin/bash
set -euo pipefail

WAIT_FOR_TIMEOUT=120 # 2 minutes
docker-compose up --detach
echo Waiting for Vault Agent container to be up
curl https://raw.githubusercontent.com/eficode/wait-for/v2.2.3/wait-for | sh -s -- localhost:8200 -t $WAIT_FOR_TIMEOUT -- echo success
docker exec vault-server /bin/sh -c "source /config/init.sh"
docker restart vault-agent
