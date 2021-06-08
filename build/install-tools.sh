#!/bin/sh
# installs tools required for makefile to work

echo "-- Installing revive"
GO111MODULE=off go get github.com/mgechev/revive
echo "-- Installing swag"
GO111MODULE=off go get github.com/swaggo/swag
GO_GET_EXIT="$?"
echo "-- Installed revive\n"
exit $GO_GET_EXIT
