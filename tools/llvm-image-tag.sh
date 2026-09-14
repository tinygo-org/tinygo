#!/bin/sh
# Print the tag of the prebuilt LLVM Docker image for this source tree.
# The hash holds only the files that change the LLVM build. Dockerfile replaces
# the other make files with COPY ., so a stale copy in the image has no effect.
set -e
cd "$(dirname "$0")/.."
rev=$(cut -c1-12 llvm-version.txt)
hash=$(cat llvm-version.txt Dockerfile.llvm GNUmakefile \
        make/config.mk make/llvm.mk \
        | git hash-object --stdin | cut -c1-12)
echo "$rev-$hash"
