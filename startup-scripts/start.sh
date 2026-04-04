#!/bin/sh
SCRIPT=$(readlink -f "$0")
SCRIPTPATH=$(dirname "$SCRIPT")
if ! command -v docker >/dev/null 2>&1
then
    echo "docker could not be found. Please install docker: https://docs.docker.com/engine/install/" >&2
    exit 1
else
    echo "Found docker."
fi

if ! command -v docker compose >/dev/null 2>&1
then
    echo "docker compose could not be found. Checking for docker-compose (legacy)"
    if ! command -v docker-compose >/dev/null 2>&1
    then
        echo "docker compose could not be found. Please install docker compose: https://docs.docker.com/compose/install/" >&2
        echo "you already have docker, docker compose is a plugin on top of docker" >&2
        exit 1
    else
        echo "Found docker-compose."
        echo "warning: docker-compose is a legacy version of docker compose, and may not work correctly" >&2
        echo "Consider updating: https://docs.docker.com/compose/install/" >&2

        echo "Running the app. If this is the first time, this may take a moment to build all requirements."
        echo "Run build.sh if you need to import new features after pulling a new version of the repo."
        docker-compose -f $SCRIPTPATH/../compose.yaml up -d
    fi
else
    echo "Found docker compose."
    
    echo "Running the app. If this is the first time, this may take a moment to build all requirements."
    echo "Run build.sh if you need to import new features after pulling a new version of the repo."
    docker compose -f $SCRIPTPATH/../compose.yaml up -d
fi