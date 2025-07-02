require('../targets/wasm_exec.js');

function runTests() {
    let testCall = (name, params, expected) => {
        let result = go._inst.exports[name].apply(null, params);
        if (result !== expected) {
            console.error(`${name}(...${params}): expected result ${expected}, got ${result}`);
        }
    }

    // These are the same tests as in TestWasmExport.
    testCall('hello', [], undefined);
    testCall('add', [3, 5], 8);
    testCall('add', [7, 9], 16);
    testCall('add', [6, 1], 7);
    testCall('reentrantCall', [2, 3], 5);
    testCall('reentrantCall', [1, 8], 9);
}

let go = new Go();
go.importObject.tester = {
    callOutside: (a, b) => {
        return go._inst.exports.add(a, b);
    },
    callTestMain: () => {
        runTests();
    },
};
WebAssembly.instantiate(fs.readFileSync(process.argv[2]), go.importObject).then(async (result) => {
    let value = await go.run(result.instance);
    console.log('exit code:', value);
}).catch((err) => {
    console.error(err);
    process.exit(1);
});
