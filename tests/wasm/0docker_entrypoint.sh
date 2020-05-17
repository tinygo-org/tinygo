#!/usr/bin/env bash
set -e

# launch headless chrome and our test server, exit immediately if one of them terminates
/headless-shell/headless-shell --no-sandbox --remote-debugging-address=0.0.0.0 --remote-debugging-port=9222 &

cd /wasm
/bin/server &

wait -n
