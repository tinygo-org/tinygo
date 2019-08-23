'use strict';

const WASM_URL = 'wasm.wasm';

var wasm;

function updateRight() {
    const value = document.getElementById("a").value;
    window.runner(function (value) {
        document.getElementById("b").value = value;
    }, value);
}

function updateLeft() {
    const value = document.getElementById("b").value;
    window.runner(function (value) {
        document.getElementById("a").value = value;
    }, value);
}

function init() {
    document.querySelector('#a').oninput = updateRight;
    document.querySelector('#b').oninput = updateLeft;

    const go = new Go();
    if ('instantiateStreaming' in WebAssembly) {
        WebAssembly.instantiateStreaming(fetch(WASM_URL), go.importObject).then(function (obj) {
            wasm = obj.instance;
            go.run(wasm);
        })
    } else {
        fetch(WASM_URL).then(resp =>
            resp.arrayBuffer()
        ).then(bytes =>
            WebAssembly.instantiate(bytes, go.importObject).then(function (obj) {
                wasm = obj.instance;
                go.run(wasm);
            })
        )
    }
}

init();
