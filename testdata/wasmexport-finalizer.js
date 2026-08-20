const fs = require('fs');

require('../targets/wasm_exec.js');

let finalized = 0;
let backgroundRuns = 0;
const go = new Go();
go.importObject.tester = {
    finalizerRan: () => {
        finalized++;
    },
    backgroundRan: () => {
        backgroundRuns++;
    },
};

WebAssembly.instantiate(fs.readFileSync(process.argv[2]), go.importObject).then(async result => {
    await go.run(result.instance);
    result.instance.exports.registerFinalizers();
    if (backgroundRuns !== 0) {
        throw new Error('wasm export ran an unrelated goroutine before returning');
    }
    for (let i = 0; i < 500 && finalized === 0; i++) {
        await new Promise(resolve => setTimeout(resolve, 1));
    }
    if (finalized === 0) {
        throw new Error('no wasm-export finalizer ran after returning to JavaScript');
    }
}).catch(err => {
    console.error(err);
    process.exit(1);
});
