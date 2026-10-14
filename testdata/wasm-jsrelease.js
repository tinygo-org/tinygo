const fs = require('fs');

require('../targets/wasm_exec.js');

const go = new Go();

// Conservative stack scanning can keep a few values alive.
const allowed = 100;

global.report = created => {
    const live = go._values.filter(v => v instanceof Uint8Array).length;
    if (live > allowed) {
        console.error(`${live} of ${created} JavaScript values are still referenced`);
        process.exit(1);
    }
    process.exit(0);
};

WebAssembly.instantiate(fs.readFileSync(process.argv[2]), go.importObject).then(result => {
    go.run(result.instance);
}).catch(err => {
    console.error(err);
    process.exit(1);
});
