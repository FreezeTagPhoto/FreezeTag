#!/bin/sh
# arguments passed in:
# $1 is the absolute plugin folder location
# $2 is the shell arguments to "uv venv" (aside from --seed and --managed-python)
# $3 is the absolute location of the freezetag package
# $4 is the requirements path according to the manifest
set -e

export PATH="$HOME/.local/bin:$PATH"
export UV_LINK_MODE=copy

cd "$1"
uv venv --managed-python --seed --allow-existing $2
. .venv/bin/activate
cd "$3"
uv pip install .
cd "$1"
if [ -n "$4" ]; then
    uv pip install -r "$4"
fi