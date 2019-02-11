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
  fetch(WASM_URL).then(resp =>
    resp.arrayBuffer()
  ).then(bytes =>
    WebAssembly.instantiate(bytes, go.importObject).then(function(obj) {
      wasm = obj.instance;
      go.run(wasm);
      updateResult();
    })
  )
  // the following code is preferred but it doesn't work on Safari at this point.
  // https://bugs.webkit.org/show_bug.cgi?id=173105
  /*WebAssembly.instantiateStreaming(fetch(WASM_URL), go.importObject).then(function(obj) {
    wasm = obj.instance;
    go.run(wasm);
    updateResult();
  })*/
}

init();
