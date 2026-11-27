# wasmprof

wasmprof converts a [wasmtime](https://wasmtime.dev/) guest profile of a TinyGo
WebAssembly program into a [pprof](https://github.com/google/pprof) profile with
call stacks, file and line information, and inlined frames.

```
tinygo build -target=wasip1 -o prog.wasm .
wasmtime run --profile=guest,prof.json,1ms prog.wasm
go run ./tools/wasmprof -o cpu.pprof prog.wasm prof.json
go tool pprof -top -lines cpu.pprof
```

## Getting a usable profile

Build with debug info, which is the default. Any `-opt` level works.

wasmtime only writes the profile when the program returns from its entry
point. A program that calls `os.Exit` or panics gets no profile. Older TinyGo
versions exit through `proc_exit` even when `main` returns. With those, build
with `-buildmode=wasi-legacy`.

Use a 1ms sampling interval. The default of 10ms gives few samples. For wasip3,
also pass the flags from `targets/wasip3.json`.

```
wasmtime run --wasm component-model -Sp3 -Wcomponent-model-async \
    -Wcomponent-model-threading --profile=guest,prof.json,1ms prog.wasm
```

The profile records addresses in the native code wasmtime compiled. wasmprof
recompiles the module with `wasmtime compile` to map them back, so use the same
wasmtime binary and pass every `-O`, `-C`, `-D` and `-W` option of the run with
`-compile-arg`, one each. Other options do not matter. Use `-wasmtime` to pick
a wasmtime binary other than the one on `PATH`.

```
go run ./tools/wasmprof \
    -compile-arg=--wasm=component-model \
    -compile-arg=-Wcomponent-model-async \
    -compile-arg=-Wcomponent-model-threading \
    -o cpu.pprof prog.wasm prof.json
```

wasmprof checks that every profiled function is at the same address in the
recompiled module and stops if not.

Some hosts write profiles with offsets into the wasm binary instead. wasmprof
detects these and skips the recompile, so no wasmtime binary or flags are
needed.

## Accuracy

wasmtime samples at function entry, at loop headers and at calls to host
functions. Time in straight line code is charged to the next of these, so a
profile is accurate per function and per loop but not per statement within a
loop body. Samples at a loop header are attributed to the first line of the
loop body.

TinyGo runs `wasm-opt`, which inlines without recording inline scopes. Where a
scope is missing, wasmprof guesses the inlined function from the declarations
in the source file. The guess can pick the wrong closure or generic instance.
With the asyncify scheduler this affects the frames between `_start` and `main`.

Code without DWARF shows function names only. `-stats` prints how the samples
were resolved.
