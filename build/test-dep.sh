#!/bin/sh

export TEST_DEP="$(dep ensure -dry-run)"

if [ $(echo "$TEST_DEP" | wc -l) -gt "1" ]; then
	echo "\"dep ensure -dry-run\" returned changes. Please run \"make dep\" to install changes"
	echo "$TEST_DEP"
    exit 1
fi
