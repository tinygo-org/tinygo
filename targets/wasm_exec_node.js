// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// This file has been modified for use by the TinyGo compiler.
// Polyfill required to maintain compatibility with NodeJS 18
"use strict";

if (process.argv.length < 3) {
    console.error("usage: go_js_wasm_exec [wasm binary] [arguments]");
    process.exit(1);
}

globalThis.require = require;
globalThis.fs = require("fs");
globalThis.TextEncoder = require("util").TextEncoder;
globalThis.TextDecoder = require("util").TextDecoder;

globalThis.performance = {
    now() {
        const [sec, nsec] = process.hrtime();
        return sec * 1000 + nsec / 1000000;
    },
};

const crypto = require("crypto");
globalThis.crypto = {
    getRandomValues(b) {
        crypto.randomFillSync(b);
    },
};

require("./wasm_exec");

const go = new Go();
WebAssembly.instantiate(fs.readFileSync(process.argv[2]), go.importObject).then(async (result) => {
    let exitCode = await go.run(result.instance);
    process.exit(exitCode);
}).catch((err) => {
    console.error(err);
    process.exit(1);
});
