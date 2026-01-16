#!/bin/bash

uv venv --seed
source .venv/bin/activate
cd ../../plugins/freezetag
pip install .
