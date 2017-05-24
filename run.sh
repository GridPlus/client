#!/bin/bash

export GOPATH="$(pwd -P)"
export GOBIN=$GOPATH/bin

cd src && ./src && cd ..
