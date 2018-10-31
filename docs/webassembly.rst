.. _webassembly:


WebAssembly
===========

.. highlight:: go

You can call a JavaScript function from Go and call a Go function from WebAssembly::

    package main

    // This calls a JS function from Go.
    func main() {
        println("adding two numbers:", add(2, 3)) // expecting 5
    }

    // This function is imported from JavaScript, as it doesn't define a body.
    // You should define a function named 'main.add' in the WebAssembly 'env'
    // module from JavaScript.
    func add(x, y int)

    // This function is exported to JavaScript, so can be called using
    // exports.add() in JavaScript.
    //go:export multiply
    func multiply(x, y int) int {
        return x * y;
    }


.. highlight:: javascript

Related JavaScript would look something like this::

    // Providing the environment object, used in WebAssembly.instantiateStreaming.
    env: {
        'main.add': function(x, y) {
            return x + y
        }
        // ... other functions
    }

    // Calling the multiply function:
    console.log('multiplied two numbers:', wasm.exports.multiply(5, 3));

A more complete example is provided in the `wasm example
<https://github.com/aykevl/tinygo/tree/master/src/examples/wasm>`_.
