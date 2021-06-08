#!/bin/sh

GOPRIVATE="github.com/tktip"
GOSUMDB=off
GO111MODULE=on go mod verify
GO111MODULE=on go mod tidy
GO111MODULE=on go mod vendor
