#!/bin/bash
mkdir bin
export GOPATH="$(pwd -P)"
export GOBIN=$GOPATH/bin

echo "Fetching packages..."
cd src && go get && cd ..

cd init && go build -ldflags -s && ./init "$1" && cd ..

echo "Keys created."

cd src && go build -ldflags -s && cd ..

echo "Client installed. Start it by running 'bash run.sh'"
