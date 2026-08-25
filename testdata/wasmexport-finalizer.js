const fs = require('fs');

require('../targets/wasm_exec.js');

let finalized = 0;
let backgroundRuns = 0;
let instance;
const go = new Go();
const sleepTicks = go.importObject.gojs['runtime.sleepTicks'];
let zeroRearms = 0;
go.importObject.gojs['runtime.sleepTicks'] = timeout => {
    if (Number(timeout) === 0) {
        zeroRearms++;
    }
    return sleepTicks(timeout);
};
go.importObject.tester = {
    finalizerRan: () => {
        finalized++;
    },
    backgroundRan: () => {
        backgroundRuns++;
    },
    callNestedExport: () => {
        instance.exports.launchBackground();
    },
};

WebAssembly.instantiate(fs.readFileSync(process.argv[2]), go.importObject).then(async result => {
    instance = result.instance;
    await go.run(instance);

    const topLevelRearms = zeroRearms;
    instance.exports.launchBackground();
    if (backgroundRuns !== 0) {
        throw new Error('wasm export ran a background goroutine before returning');
    }
    if (zeroRearms !== topLevelRearms + 1) {
        throw new Error('top-level wasm export did not schedule exactly one wakeup');
    }
    for (let i = 0; i < 500 && backgroundRuns === 0; i++) {
        await new Promise(resolve => setTimeout(resolve, 1));
    }
    if (backgroundRuns !== 1) {
        throw new Error('wasm scheduler did not resume after an export without finalizers');
    }

    instance.exports.installNestedCallback();
    const nestedRearms = zeroRearms;
    global.nestedExportCallback();
    if (zeroRearms !== nestedRearms) {
        throw new Error('re-entrant wasm export scheduled a redundant wakeup');
    }
    if (backgroundRuns !== 2) {
        throw new Error('outer scheduler did not drain work from a re-entrant wasm export');
    }

    instance.exports.registerFinalizers();
    if (backgroundRuns !== 2) {
        throw new Error('wasm export ran an unrelated goroutine before returning');
    }
    for (let i = 0; i < 500 && (finalized === 0 || backgroundRuns < 3); i++) {
        await new Promise(resolve => setTimeout(resolve, 1));
    }
    if (finalized === 0) {
        throw new Error('no wasm-export finalizer ran after returning to JavaScript');
    }
    if (backgroundRuns !== 3) {
        throw new Error('wasm scheduler did not resume after a finalizer export');
    }
}).catch(err => {
    console.error(err);
    process.exit(1);
});
