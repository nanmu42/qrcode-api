#!/usr/bin/env bash

VERSION=`git describe --tags --dirty`
BUILD=`date +%FT%T%z`

# can not build statically safely
# see https://stackoverflow.com/questions/8140439/why-would-it-be-impossible-to-fully-statically-link-an-application
export LD_LIBRARY_PATH=/usr/local/lib
go build -ldflags "-s -w -X main.Version=$VERSION -X main.BuildDate=$BUILD" -o qrcode-api