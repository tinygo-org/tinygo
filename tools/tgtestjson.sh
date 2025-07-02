#!/usr/bin/env bash

# Run tests and convert output to json with go tool test2json.
# This is a workaround for the lack of -json output in tinygo test.
# Some variables must be set in the environment beforehand.
# TODO: let's just add -json support to tinygo test.

TINYGO="${TINYGO:-tinygo}"
PACKAGES="${PACKAGES:-"./tests"}"
TARGET="${TARGET:-wasip2}"
TESTOPTS="${TESTOPTS:-"-x -work"}"

# go clean -testcache
for pkg in $PACKAGES; do
    # Example invocation with test2json in BigGo:
    # go test -test.v=test2json ./$pkg 2>&1 | go tool test2json -p $pkg

    # Uncomment to see resolved commands in output
    # >&2 echo "${TINYGO} test -v -target $TARGET $TESTOPTS $pkg 2>&1 | go tool test2json -p $pkg"
    "${TINYGO}" test -v -target $TARGET $TESTOPTS $pkg 2>&1 | go tool test2json -p $pkg

done
