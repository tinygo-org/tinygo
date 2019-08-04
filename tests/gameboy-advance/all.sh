#!/bin/bash

set -e

# Always run from the script directory
cd "$(dirname "${BASH_SOURCE[0]}")"

# Flag variables.
RUN='no$gba'                     # emulator to run (must be in $PATH)
NAME=""                          # if set, will only run named subdir
IMAGE="tinygo/tinygo_gba:latest" # container to run

while getopts "dn:c:" opt; do
  case "$opt" in
    d)
        RUN="${RUN}-debug"
        ;;
    n)
        NAME="$OPTARG"
        ;;
    c)
        IMAGE="$OPTARG"
        ;;
    *)
        echo "Usage: $0 [-d] [-n <name>] [-c <container>]"
        echo "  -d          Run inside debugger (appends -debug to emulator)"
        echo "  -n <name>   Only run directory <name>"
        echo "  -c <image>  Run in container, e.g. -c tinygo/tinygo_gba:${USERNAME}"
        exit 1
        ;;
  esac
done
shift $((OPTIND-1))

for DIR in *; do
    [[ -d "$DIR" ]] || continue
    [[ -n "$NAME" && "$(basename "$DIR")" != "$NAME" ]] && continue

    echo "Building ${DIR}..."

    # The / works around mingw's shell munging, which confuses docker.
    docker run --rm \
      -v "/$(realpath "$DIR")":/gba \
      -v "/$(realpath ../../src)":/tinygo/src \
      "$IMAGE" clean build || continue
    
    for CART in "$DIR"/*.gba; do
      echo "Running ${CART}..."

      # On Windows, the .gba needs to be in no$gba's slot directory.
      if [[ "$(go env GOOS)" == "windows" ]]; then
        BIN="$(which "$RUN")"
        DIR="$(dirname "$BIN")/slot"
        mkdir -p "$DIR"
        cp "$CART" "$DIR"
        CART="$(basename $CART)"
      fi

      # Wait for any previous emulators to exit.
      wait

      # Launch the emulator and continue to building the next cartridge.
      "$RUN" "${CART}" || true &
    done
done

# Wait for any remaining emulators.
wait
