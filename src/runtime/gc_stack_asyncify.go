//go:build (gc.conservative || gc.custom || gc.precise || gc.boehm) && tinygo.wasm && scheduler.asyncify

package runtime

import "internal/task"

func markTaskStacks() { task.ScanStacks() }
