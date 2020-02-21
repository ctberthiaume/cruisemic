#!/bin/bash
# Build cruisemic command-line tool for 64-bit MacOS and Linux

VERSION=$(git describe --long --dirty)

[[ -d cruisemic.darwin-amd64 ]] && rm -rf cruisemic.darwin-amd64
[[ -d cruisemic.linux-amd64 ]] && rm -rf cruisemic.linux-amd64
GOOS=darwin GOARCH=amd64 go build -o cruisemic.${VERSION}.darwin-amd64/cruisemic cmd/cruisemic/main.go || exit 1
GOOS=linux GOARCH=amd64 go build -o cruisemic.${VERSION}.linux-amd64/cruisemic cmd/cruisemic/main.go || exit 1
zip -q -r cruisemic.${VERSION}.darwin-amd64.zip cruisemic.${VERSION}.darwin-amd64 || exit 1
zip -q -r cruisemic.${VERSION}.linux-amd64.zip cruisemic.${VERSION}.linux-amd64 || exit 1
