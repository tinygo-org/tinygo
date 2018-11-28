'use strict';

const WASM_URL = '../../../wasm.wasm';

var wasm;

function updateResult() {
  wasm.exports.update();
}

function init() {
  document.querySelector('#a').oninput = updateResult;
  document.querySelector('#b').oninput = updateResult;

  const go = new Go();
  WebAssembly.instantiateStreaming(fetch(WASM_URL), go.importObject).then(function(obj) {
    wasm = obj.instance;
    go.run(wasm);
    updateResult();
  })
}

init();
