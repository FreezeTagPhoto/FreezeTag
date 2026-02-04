#!/bin/sh
# arguments passed in:
# $1 is the absolute path of the plugin folder
# $2 is the name of the main file according to the manifest
# NOTE: this script only works if a .venv exists in the folder already (create_venv should have run)
set -e

cd "$1"
source .venv/bin/activate
uv run "$2"