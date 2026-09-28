//go:build !scheduler.cores

package runtime

func waitForSecondaryCoresReady() {}
