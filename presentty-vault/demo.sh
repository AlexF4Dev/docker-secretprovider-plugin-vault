#!/bin/bash
set -u
set -o pipefail
export plugin_name=${plugin_name:-sirlatrom/docker-secretprovider-plugin-vault:latest}
export remote=${remote:-${plugin_name}}
./cleanup.sh &>/dev/null

export VAULT_ADDR=http://127.0.0.1:8200 VAULT_TOKEN=1234
docker container run --detach --name vault --publish 8200:8200 vault server -dev -dev-root-token-id=$VAULT_TOKEN
sleep 1
docker container exec -i --env VAULT_ADDR=http://127.0.0.1:8200 --env VAULT_TOKEN vault vault policy write demo_snitch -<<EOF
path "secret/*" {
  capabilities = ["create","read","update","list","delete"]
}
path "auth/token/create" {
  capabilities = ["create", "update"]
}
EOF
docker container exec --env VAULT_ADDR --env VAULT_TOKEN vault vault kv put secret/multi a=something b=else
docker container exec --env VAULT_ADDR --env VAULT_TOKEN vault vault kv put secret/multi a=something b=else c=only_in_v2
docker container exec --env VAULT_ADDR --env VAULT_TOKEN vault vault kv put secret/demo_secret value=this_was_not_wrapped
docker container exec --env VAULT_ADDR --env VAULT_TOKEN vault vault kv put secret/demo_wrapped_secret value=this_was_once_wrapped
echo -n '1234' | docker secret create secret-zero -
docker service create --mode global --constraint 'node.role == manager' --name vault-helper-service --secret secret-zero --restart-condition on-failure busybox tail -f /dev/null

installer
until [[ "$(docker plugin inspect ${plugin_name} --format '{{.Enabled}}' 2>/dev/null)" == "true" ]]
do
    echo "waiting for plugin to be installed"
    sleep 1
done

docker node ls --filter role=worker -q | wc -l | grep -q 0 && snitch_role=manager || snitch_role=worker
export snitch_role
docker stack deploy --compose-file docker-compose.yml demo
docker service logs -f demo_snitch
exit $?
