//go:build (scheduler.tasks || scheduler.cores) && tinygo.riscv && !tinygo.riscv.f

package task

// numFPREGS is zero on targets without saved FP registers.
const numFPREGS = 0
