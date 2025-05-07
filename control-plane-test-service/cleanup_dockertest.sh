#!/usr/bin/env bash

echo "WARNING: This action will remove all docker containers on your machine!"

read -r -p "Continue? [y/N] " response
case "$response" in
    [yY][eE][sS]|[yY])
        docker stop $(docker ps -aq)
        docker rm $(docker ps -aq)

        docker network rm $(docker network ls -q --filter name=testnet)
        ;;
    *)
        #do_something_else
        ;;
esac
