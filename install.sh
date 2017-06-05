#!/bin/bash
mkdir bin
export GOPATH="$(pwd -P)"
export GOBIN=$GOPATH/bin

echo "Fetching packages..."
cd src && go get && cd ..

cd init && go build -ldflags -s && ./init && cd ..

echo "Services configured."

cd src && go build -ldflags -s && cd ..

echo "Client installed. Fill out your setup_keys.toml file and run with 'run.sh'"
