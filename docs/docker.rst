.. docker:

.. highlight:: none

Using with Docker
=================

A docker container exists for easy access to the ``tinygo`` CLI. For example, to
compile ``wasm.wasm`` for the WebAssembly example, from the root of the
repository::

    docker run --rm -v $(pwd):/src tinygo/tinygo build -o /src/wasm.wasm -target wasm examples/wasm

Note that you cannot run ``tinygo flash`` from inside the docker container,
so it is less useful for microcontroller development.
