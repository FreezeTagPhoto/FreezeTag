#!/bin/bash

FILES=$(gofmt -l $(find . -type f -name '*.go' -not -path "./.go/*"))
LINES=$(echo "$FILES" | wc -l)

if ((LINES == 0)); then
    echo "All Go files are correctly formatted."
else
    echo "Some files are not formatted correctly:"
    echo "$FILES"
    echo "run 'gofmt -d [file]' to see the problem."
    exit 1
fi