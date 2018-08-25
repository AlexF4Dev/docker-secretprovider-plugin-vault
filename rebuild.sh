#!/bin/bash
sudo rm -frv plugin/rootfs/.dockerenv plugin/rootfs/*
docker container rm -f -v vault
docker service rm snitch vault-helper-service
docker secret rm secret-zero haha VAULT_TOKEN_SORT_OF
docker plugin disable ${REGISTRY}/absukl/secrets-plugin:latest
docker plugin rm ${REGISTRY}/absukl/secrets-plugin:latest
docker build --tag rootfsimage .
id=$(docker create rootfsimage true)
sudo docker export "$id" | sudo tar -x -C plugin/rootfs/
docker rm -vf "$id"
docker plugin create ${REGISTRY}/absukl/secrets-plugin:latest ${PWD}/plugin
docker container run --detach --name vault --publish 8200:8200 vault server -dev -dev-root-token-id=1234
docker plugin enable ${REGISTRY}/absukl/secrets-plugin:latest
echo -n '1234' | docker secret create secret-zero
docker service create --constraint 'node.role == manager' --name vault-helper-service --secret secret-zero --restart-condition on-failure busybox tail -f /dev/null
docker secret create --driver ${REGISTRY}/absukl/secrets-plugin haha
docker secret create --driver ${REGISTRY}/absukl/secrets-plugin --label "dk.almbrand.docker.plugin.secretprovider.vault.type"="vault_token" VAULT_TOKEN_SORT_OF
docker service create --constraint 'node.role == worker' --detach --name snitch --secret haha --secret VAULT_TOKEN_SORT_OF busybox sh -c 'find /run/secrets -type f | xargs -i -n 1 sh -c "echo {}:; cat {}; echo"; sleep 2'
docker container exec --env VAULT_ADDR=http://127.0.0.1:8200 --env VAULT_TOKEN=1234 vault vault kv put secret/haha value=sosecret
docker service logs -f snitch
exit $?
