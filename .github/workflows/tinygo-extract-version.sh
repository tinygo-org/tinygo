#!/bin/sh

# Extract the version string from the source code, to be stored in a variable.
grep 'const version' goenv/version.go | sed 's/^const version = "\(.*\)"$/version=\1/g'
