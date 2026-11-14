// Host for gcstacks-bridge.go: one Go instance per cycle, mount, halt, then
// check that no Uint8Array is still held by any instance's bridge table.
'use strict'

const fs = require('fs')

require('../targets/wasm_exec.js')

const cycles = 20
const tick = (ms = 5) => new Promise((resolve) => setTimeout(resolve, ms))
globalThis.fetch = () => new Promise((resolve) => {
    setTimeout(() => {
        const body = new ArrayBuffer(136)
        resolve({
            status: 200,
            arrayBuffer: () => new Promise((r) => setTimeout(() => r(body), 1)),
        })
    }, 1)
})

async function main() {
    const instances = []
    for (let i = 0; i < cycles; i++) {
        globalThis.slot = {}
        const go = new Go()
        const {instance} = await WebAssembly.instantiate(fs.readFileSync(process.argv[2]), go.importObject)
        const running = go.run(instance)
        await tick(2)
        slot.mount()
        await tick(15)
        slot.halt()
        await Promise.race([running, tick(1000)])
        // The host is done with this instance: release the bridge slots it
        // holds. A stopped program never runs finalizers, so without this the
        // response buffers stay pinned for as long as the module is kept.
        if (typeof go.releaseBridge === 'function') go.releaseBridge()
        instances.push(go)
        await tick(2)
    }

    // A WeakRef target survives the turn that dereferenced it, so collect in a
    // later turn.
    await tick(10)
    gc()
    await tick(20)
    gc()

    let leaked = 0
    for (const go of instances) {
        for (const value of go._values) {
            if (value instanceof Uint8Array) leaked++
        }
    }
    if (leaked !== 0) {
        console.error(`gcstacks-bridge: ${leaked} Uint8Array values still in _values`)
        process.exit(1)
    }
    console.log('bridge slots released')
}

main().catch((err) => {
    console.error(err)
    process.exit(1)
})
