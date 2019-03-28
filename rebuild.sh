#!/bin/bash
set -u
set -o pipefail
export plugin_name=${plugin_name:-sirlatrom/docker-secretprovider-plugin-vault}
export remote=${remote:-${plugin_name}}
rm -fr plugin/rootfs/.dockerenv plugin/rootfs/*
docker container rm -f -v vault
docker service rm snitch vault-helper-service
docker secret rm secret-zero multi_a multi_b multi_json multi_json_v1 multi_json_v2 secret wrapped_secret generic_vault_token
# docker plugin disable ${plugin_name}
# docker plugin rm ${plugin_name}
docker service rm ${plugin_name/*\//}

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
docker container exec -i --env VAULT_ADDR=http://127.0.0.1:8200 --env VAULT_TOKEN=1234 vault vault policy write snitch -<<EOF
path "secret/*" {
  capabilities = ["create","read","update","list","delete"]
}
path "auth/token/create" {
  capabilities = ["create", "update"]
}
EOF
docker container exec --env VAULT_ADDR=http://127.0.0.1:8200 --env VAULT_TOKEN=1234 vault vault kv put secret/multi a=something b=else
docker container exec --env VAULT_ADDR=http://127.0.0.1:8200 --env VAULT_TOKEN=1234 vault vault kv put secret/multi a=something b=else c=only_in_v2
docker container exec --env VAULT_ADDR=http://127.0.0.1:8200 --env VAULT_TOKEN=1234 vault vault kv put secret/secret value=this_was_not_wrapped
docker container exec --env VAULT_ADDR=http://127.0.0.1:8200 --env VAULT_TOKEN=1234 vault vault kv put secret/wrapped_secret value=this_was_once_wrapped
echo -n '1234' | docker secret create secret-zero -
docker service create --mode global --constraint 'node.role == manager' --name vault-helper-service --secret secret-zero --restart-condition on-failure busybox tail -f /dev/null

go run installer/plugin_installer.go
sleep 5

# docker plugin enable ${plugin_name}
docker secret create --driver ${plugin_name} --label "dk.almbrand.docker.plugin.secretprovider.vault.path"="secret/data/multi" --label "dk.almbrand.docker.plugin.secretprovider.vault.field"="a" multi_a
docker secret create --driver ${plugin_name} --label "dk.almbrand.docker.plugin.secretprovider.vault.path"="secret/data/multi" --label "dk.almbrand.docker.plugin.secretprovider.vault.field"="b" multi_b
docker secret create --driver ${plugin_name} --label "dk.almbrand.docker.plugin.secretprovider.vault.json"="true" --label "dk.almbrand.docker.plugin.secretprovider.vault.path"="secret/data/multi" multi_json
docker secret create --driver ${plugin_name} --label "dk.almbrand.docker.plugin.secretprovider.vault.json"="true" --label "dk.almbrand.docker.plugin.secretprovider.vault.path"="secret/data/multi?version=1" multi_json_v1
docker secret create --driver ${plugin_name} --label "dk.almbrand.docker.plugin.secretprovider.vault.json"="true" --label "dk.almbrand.docker.plugin.secretprovider.vault.path"="secret/data/multi?version=2" multi_json_v2
docker secret create --driver ${plugin_name} secret
docker secret create --driver ${plugin_name} --label "dk.almbrand.docker.plugin.secretprovider.vault.wrap"="true" wrapped_secret
docker secret create --driver ${plugin_name} --label "dk.almbrand.docker.plugin.secretprovider.vault.type"="vault_token" generic_vault_token
docker node ls --filter role=worker -q | wc -l | grep -q 0 && snitch_role=manager || snitch_role=worker
docker service create --constraint 'node.role == '$snitch_role --detach --name snitch --replicas 2 --env VAULT_ADDR=http://172.17.0.1:8200 --secret multi_a --secret multi_b --secret multi_json --secret multi_json_v1 --secret multi_json_v2 --secret secret --secret wrapped_secret --secret generic_vault_token vault sh -c '
echo -n "secret:              "
cat /run/secrets/secret
echo
echo -n "wrapped_secret:      "
cat /run/secrets/wrapped_secret
echo
echo -n "unwrapped_secret:    "
VAULT_TOKEN=$(cat /run/secrets/wrapped_secret) vault unwrap -field=value
echo
echo -n "generic_vault_token: "
cat /run/secrets/generic_vault_token
echo
echo -n "multi_a:             "
cat /run/secrets/multi_a
echo
echo -n "multi_b:             "
cat /run/secrets/multi_b
echo
echo -n "multi_json:          "
cat /run/secrets/multi_json
echo
echo -n "multi_json_v1:       "
cat /run/secrets/multi_json_v1
echo
echo -n "multi_json_v2:       "
cat /run/secrets/multi_json_v2
echo'
docker service logs -f snitch
exit $?
