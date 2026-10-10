//go:build teensy41

package runtime

import "device/nxp"

// MPU region size for the 8 MiB QSPI flash of the Teensy 4.1.
const qspiFlashMPUSize = nxp.RGNSZ_8MB
