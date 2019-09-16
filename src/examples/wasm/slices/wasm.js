'use strict';

const WASM_URL = 'wasm.wasm';

var wasm;

function update() {
    const value = document.getElementById("a").value;
    document.getElementById("b").innerHTML = JSON.stringify(window.splitter(value));
}

function init() {
    document.querySelector('#a').oninput = update;

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
