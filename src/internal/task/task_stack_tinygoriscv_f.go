//go:build (scheduler.tasks || scheduler.cores) && tinygo.riscv && tinygo.riscv.f

package task

// numFPREGS is fs0-fs11 and fcsr, saved after the integer registers.
const numFPREGS = 13
