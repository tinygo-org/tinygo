#!/bin/sh
# Print the tag of the prebuilt LLVM Docker image for this source tree.
# The tag changes when the LLVM revision or the build recipe changes.
# Keep the file list in agreement with the COPY lines in Dockerfile.llvm.
set -e
cd "$(dirname "$0")/.."
rev=$(cut -c1-12 llvm-version.txt)
hash=$(cat llvm-version.txt Dockerfile.llvm GNUmakefile make/*.mk \
        | git hash-object --stdin | cut -c1-12)
echo "$rev-$hash"
