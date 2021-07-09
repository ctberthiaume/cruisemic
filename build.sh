#!/bin/bash
# Build cruisemic command-line tool for 64-bit MacOS and Linux

VERSION=$(git describe --tags)

[[ -d build ]] || mkdir build
GOOS=darwin GOARCH=amd64 go build -o build/cruisemic.${VERSION}.darwin-amd64 cmd/cruisemic/main.go || exit 1
GOOS=linux GOARCH=amd64 go build -o build/cruisemic.${VERSION}.linux-amd64 cmd/cruisemic/main.go || exit 1
