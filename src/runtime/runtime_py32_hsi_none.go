//go:build py32 && py32_no_hsi_fs

package runtime

// The M4 SVDs do not expose the M0+ ICSCR.HSI_FS selector. Preserve their
// vendor-defined 8 MHz reset clock instead of writing an unverified encoding.
func configureHSI() {
}
