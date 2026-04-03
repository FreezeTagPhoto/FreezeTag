#!/bin/sh
echo "Killing the app."
SCRIPT=$(readlink -f "$0")
SCRIPTPATH=$(dirname "$SCRIPT")

docker compose -f $SCRIPTPATH/../compose.yaml down