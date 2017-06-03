#!/bin/bash
mkdir bin
export GOPATH="$(pwd -P)"
export GOBIN=$GOPATH/bin

echo "Fetching packages..."
cd src && go get && go build -ldflags -s && cd ..

echo "Client installed. Start it by running 'bash run.sh'"
