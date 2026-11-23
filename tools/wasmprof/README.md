# wasmprof

wasmprof converts a [wasmtime](https://wasmtime.dev/) guest profile of a TinyGo
WebAssembly program into a [pprof](https://github.com/google/pprof) profile with
call stacks, file:line information and inlined frames.

```
wasmtime run --profile=guest,prof.json,1ms prog.wasm
go run ./tools/wasmprof -o cpu.pprof prog.wasm prof.json
go tool pprof -top -lines cpu.pprof
```

## Getting a usable profile from wasmtime

1. **Build with debug info.** This is the default. Do not pass `-no-debug`.
   Any `-opt` level works.

2. **Make sure the program returns from its entry point.** wasmtime writes the
   guest profile only when the program returns normally. If the program exits
   through `proc_exit` (any `os.Exit` call) or traps (for example on a panic),
   wasmtime exits without writing the profile.

   - wasip1: TinyGo calls `proc_exit(0)` when `main` returns, so build with
     `-buildmode=wasi-legacy`. Its `_start` returns when `main` returns.

     ```
     tinygo build -target=wasip1 -buildmode=wasi-legacy -o prog.wasm .
     ```

   - wasip3: `wasi:cli/run` returns normally, so no extra flag is needed.

   Do not call `os.Exit` at the end of `main`, and profile programs that do not
   panic.

3. **Run with the guest profiler.** Use a 1ms interval. The default is 10ms,
   which gives few samples for short runs.

   ```
   wasmtime run --profile=guest,prof.json,1ms prog.wasm args...
   ```

   For wasip3, use the flags from `targets/wasip3.json`:

   ```
   wasmtime run --wasm component-model -Sp3 -Wcomponent-model-async \
       -Wcomponent-model-threading --profile=guest,prof.json,1ms prog.wasm
   ```

   wasmtime prints `Profile written to: prof.json` when it worked.

4. **Convert with the same wasmtime binary and the same codegen flags.** The
   profile records native code addresses. wasmprof recompiles the module with
   `wasmtime compile` to map those back to wasm code, so the compiled code has
   to be identical to the profiled run. Use the same wasmtime build, and pass
   every `-O`, `-C`, `-D` and `-W` option given to `wasmtime run`, one per
   `-compile-arg`. Other options (`-S`, `--dir`, `--env`, program arguments)
   do not matter. wasmprof adds `-W epoch-interruption=y` itself, as the guest
   profiler enables it.

   ```
   go run ./tools/wasmprof -o cpu.pprof prog.wasm prof.json

   # wasip3
   go run ./tools/wasmprof \
       -compile-arg=--wasm=component-model \
       -compile-arg=-Wcomponent-model-async \
       -compile-arg=-Wcomponent-model-threading \
       -o cpu.pprof prog.wasm prof.json
   ```

   If the flags or the wasmtime binary differ, wasmprof stops with
   `profiled functions are not where the profile says`. Use `-wasmtime` to
   select a wasmtime binary other than the one on `PATH`.

5. **Look at it.**

   ```
   go tool pprof -top -lines cpu.pprof
   go tool pprof -list 'main\.' cpu.pprof
   go tool pprof -http=: cpu.pprof
   ```

   The JSON profile itself can also be opened in https://profiler.firefox.com/.

## Accuracy

- wasmtime samples only at its epoch checks, which are at function entry and
  at loop headers, and at calls to host functions. Time spent in straight-line
  code is charged to the next check. This gives accurate function and loop
  level attribution, but cannot separate statements inside one loop body.
  A loop header has no line of its own, so wasmprof attributes such samples to
  the first line of the loop body.
- TinyGo always runs `wasm-opt` on wasm output. It inlines without recording
  inline scopes in DWARF, and its asyncify pass (used by the default wasip1
  scheduler) rewrites the scheduler path. Where a scope is missing, wasmprof
  guesses the inlined function from the declarations in the source file and
  leaves the caller's line as 0. The guess can pick the wrong closure or
  generic instance. In practice it affects the frames between `_start` and
  `main` with the asyncify scheduler, and rarely anything else.
- Code without DWARF (adapter modules in components, libraries built without
  debug info) shows function names only.
- `-stats` prints how the leaf position of the samples was resolved.

## How it works

The guest profile is in the Firefox Profiler format. It has full call stacks,
with each frame given as an offset into the native code wasmtime compiled, and
function names from the wasm name section, but no wasm offsets or lines.
wasmprof recompiles the module, reads the native to wasm offset map from
`wasmtime objdump --addrmap`, checks that every profiled function is at the
same address, and resolves the wasm offsets against the DWARF of the core
module they fall in. Components are supported, as wasmtime gives the offsets
relative to the component file.
