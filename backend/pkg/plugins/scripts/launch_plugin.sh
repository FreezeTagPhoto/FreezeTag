#!/bin/sh
# arguments passed in:
# $1 is the absolute path of the plugin folder
# $2 is the name of the main file according to the manifest
# NOTE: this script only works if a .venv exists in the folder already (create_venv should have run)
set -e

export PATH="$HOME/.local/bin:$PATH"
export UV_LINK_MODE=copy

cd "$1"
. .venv/bin/activate
uv run "$2"