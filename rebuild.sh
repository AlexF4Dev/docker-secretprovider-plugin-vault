#!/bin/bash
set -u
set -o pipefail
export plugin_name=${plugin_name:-sirlatrom/docker-secretprovider-plugin-vault:latest}
export remote=${remote:-${plugin_name}}
rm -fr plugin/rootfs/.dockerenv plugin/rootfs/*
docker container rm -f -v vault
docker stack rm repro
docker service rm repro_snitch vault-helper-service
docker secret rm secret-zero # multi_a multi_b multi_json multi_json_v1 multi_json_v2 secret wrapped_secret generic_vault_token
# docker plugin disable ${plugin_name}
# docker plugin rm ${plugin_name}
plugin_service=${plugin_name/*\//}
plugin_service=${plugin_service/:*/}
docker service rm ${plugin_service}

# docker build --quiet --tag rootfsimage .
# id=$(docker create rootfsimage true)
# mkdir -p plugin/rootfs/
# docker export "$id" | tar -x -C plugin/rootfs/
# docker rm -vf "$id"
# docker plugin create ${plugin_name} ${PWD}/plugin
# docker plugin disable ${plugin_name}
# docker plugin rm ${plugin_name}

docker container run --detach --name vault --publish 8200:8200 vault server -dev -dev-root-token-id=1234
sleep 1
docker container exec -i --env VAULT_ADDR=http://127.0.0.1:8200 --env VAULT_TOKEN=1234 vault vault policy write repro_snitch -<<EOF
path "secret/*" {
  capabilities = ["create","read","update","list","delete"]
}
path "auth/token/create" {
  capabilities = ["create", "update"]
}
EOF
docker container exec --env VAULT_ADDR=http://127.0.0.1:8200 --env VAULT_TOKEN=1234 vault vault kv put secret/multi a=something b=else
docker container exec --env VAULT_ADDR=http://127.0.0.1:8200 --env VAULT_TOKEN=1234 vault vault kv put secret/multi a=something b=else c=only_in_v2
docker container exec --env VAULT_ADDR=http://127.0.0.1:8200 --env VAULT_TOKEN=1234 vault vault kv put secret/repro_secret value=this_was_not_wrapped
docker container exec --env VAULT_ADDR=http://127.0.0.1:8200 --env VAULT_TOKEN=1234 vault vault kv put secret/repro_wrapped_secret value=this_was_once_wrapped
echo -n '1234' | docker secret create secret-zero -
docker service create --mode global --constraint 'node.role == manager' --name vault-helper-service --secret secret-zero --restart-condition on-failure busybox tail -f /dev/null

set -x
go run installer/plugin_installer.go
sleep 10
set +x

# docker plugin enable ${plugin_name}
docker node ls --filter role=worker -q | wc -l | grep -q 0 && snitch_role=manager || snitch_role=worker
export snitch_role
docker stack deploy --compose-file docker-compose.yml repro
docker-stack-waiter -c docker-compose.yml repro
docker service logs -f repro_snitch
exit $?
