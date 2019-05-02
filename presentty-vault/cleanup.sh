#!/bin/bash
set -u
set -o pipefail
export plugin_name=${plugin_name:-sirlatrom/docker-secretprovider-plugin-vault:latest}
export remote=${remote:-${plugin_name}}
rm -fr plugin/rootfs/.dockerenv plugin/rootfs/*
docker container rm -f -v vault
docker stack rm demo
docker service rm vault-helper-service
docker secret rm secret-zero
plugin_service=${plugin_name/*\//}
plugin_service=${plugin_service/:*/}
docker service rm ${plugin_service}
while docker plugin inspect ${plugin_name} &> /dev/null
do
    echo "waiting for plugin to be uninstalled"
    sleep 1
done
