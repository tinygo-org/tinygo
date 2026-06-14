//go:build (gc.conservative || gc.precise) && !(amd64 || arm64 || (arm && !baremetal && !tinygo.wasm) || (cortexm && !cortexm.noclz) || mips || mipsle || tinygo.wasm)

package runtime

// hasFastCLZ indicates whether the target CPU has a "Count Leading Zeroes" or
// "Find First Set" instruction. These enable efficient bitmap processing. Most
// common architectures have such an instruction, but there are a few major
// exceptions that we need to deal with:
//   - ARM Cortex M0/M0+ omit the CLZ instruction
//   - AVR has extremely limited bit-manipulation instructions (no CLZ)
//   - RISC-V's CLZ instruction requires the B extension. No supported devices
//     currently implement this extension.
const hasFastCLZ = false
