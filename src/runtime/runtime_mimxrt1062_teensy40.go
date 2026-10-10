//go:build teensy40

package runtime

import "device/nxp"

// MPU region size for the 2 MiB QSPI flash of the Teensy 4.0.
const qspiFlashMPUSize = nxp.RGNSZ_2MB
