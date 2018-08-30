#!/bin/bash
set -u
set -o pipefail
rm -fr plugin/rootfs/.dockerenv plugin/rootfs/*
docker container rm -f -v vault
docker service rm snitch vault-helper-service
docker secret rm secret-zero secret wrapped_secret generic_vault_token
docker plugin disable ${REGISTRY}/absukl/secrets-plugin:latest
docker plugin rm ${REGISTRY}/absukl/secrets-plugin:latest
docker build --quiet --tag rootfsimage .
id=$(docker create rootfsimage true)
mkdir -p plugin/rootfs/
docker export "$id" | tar -x -C plugin/rootfs/
docker rm -vf "$id"
docker plugin create ${REGISTRY}/absukl/secrets-plugin:latest ${PWD}/plugin
docker container run --detach --name vault --publish 8200:8200 vault server -dev -dev-root-token-id=1234
sleep 1
docker container exec -i --env VAULT_ADDR=http://127.0.0.1:8200 --env VAULT_TOKEN=1234 vault vault policy write snitch -<<EOF
path "secret/data/*" {
  capabilities = ["create","read","update"]
}
path "auth/token/create" {
  capabilities = ["create", "update"]
}
EOF
docker container exec --env VAULT_ADDR=http://127.0.0.1:8200 --env VAULT_TOKEN=1234 vault vault kv put secret/secret value=this_was_not_wrapped
docker container exec --env VAULT_ADDR=http://127.0.0.1:8200 --env VAULT_TOKEN=1234 vault vault kv put secret/wrapped_secret value=this_was_once_wrapped
echo -n '1234' | docker secret create secret-zero -
docker service create --mode global --constraint 'node.role == manager' --name vault-helper-service --secret secret-zero --restart-condition on-failure busybox tail -f /dev/null
docker plugin enable ${REGISTRY}/absukl/secrets-plugin:latest
docker secret create --driver ${REGISTRY}/absukl/secrets-plugin --label "dk.almbrand.docker.plugin.secretprovider.vault.wrap"="false" secret
docker secret create --driver ${REGISTRY}/absukl/secrets-plugin --label "dk.almbrand.docker.plugin.secretprovider.vault.wrap"="true" wrapped_secret
docker secret create --driver ${REGISTRY}/absukl/secrets-plugin --label "dk.almbrand.docker.plugin.secretprovider.vault.type"="vault_token" generic_vault_token
docker node ls --filter role=worker -q | wc -l | grep -q 0 && snitch_role=manager || snitch_role=worker
docker service create --constraint 'node.role == '$snitch_role --detach --name snitch --replicas 2 --env VAULT_ADDR=http://172.17.0.1:8200 --secret secret --secret wrapped_secret --secret generic_vault_token vault sh -c 'echo -n "secret:              "; cat /run/secrets/secret; echo; echo -n "wrapped_secret:      "; cat /run/secrets/wrapped_secret; echo; echo -n "unwrapped_secret:    "; VAULT_TOKEN=$(cat /run/secrets/wrapped_secret) vault unwrap -field=value; echo; echo -n "generic_vault_token: "; cat /run/secrets/generic_vault_token; echo'
docker service logs -f snitch
exit $?
