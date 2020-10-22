# CosmWasm

This is a minor patch on TinyGo to build CosmWasm-compatible
contracts. It adds a new target `cosmwasm`, which can be used in place
of `wasm`, in order to build wasm blob that are compatible the cosmwasm.

In particular, it doesn't require a number of imports like `putc`.

## Building a docker image

We use these custom docker images in `cosmwasm-go` to build the wasm
files. When updating to a newer TinyGo, we need to update our
build container:

```
docker build . -f Dockerfile.wasm -t cosmwasm/tinygo:v0.14.1
```