#!/bin/sh
echo "Killing the app."
SCRIPT=$(readlink -f "$0")
SCRIPTPATH=$(dirname "$SCRIPT")

if ! command -v docker compose >/dev/null 2>&1
then
    docker compose -f $SCRIPTPATH/../compose.yaml down
else
    docker-compose -f $SCRIPTPATH/../compose.yaml down
fi