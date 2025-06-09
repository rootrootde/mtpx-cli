#!/bin/bash

set -e

if [ ! -f go.mod ]; then
  echo "Initializing Go module..."
  go mod init mtpx-cli
fi

echo "Getting dependencies..."
go get github.com/ganeshrvel/go-mtpx

echo "Building mtpx-cli..."
go build -o mtpx-cli main_refactored.go

echo "Build complete: ./mtpx-cli"