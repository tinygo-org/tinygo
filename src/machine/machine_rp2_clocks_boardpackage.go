//go:build (rp2040 || rp2350) && boardpackage

package machine

type pllEntry struct {
	xtalHz uint32
	sysFB  uint32
	sysPD1 uint32
	sysPD2 uint32
}

var supportedXTALs = [...]uint32{12 * MHz}

var pllTable [len(supportedXTALs)]pllEntry

func xtalFreqHz() uint32 {
	return boardXTAL()
}

// Precompute supported crystal PLL divider tables at compile time via interp.
func init() {
	for i, xtal := range supportedXTALs {
		fbSys, _, pd1Sys, pd2Sys, err := calcPLLDivs(
			pllSearch{LockRefDiv: 1},
			uint64(xtal),
			cpuFreq,
			MHz,
		)
		if err != nil {
			panic(err)
		}
		pllTable[i] = pllEntry{
			xtalHz: xtal,
			sysFB:  uint32(fbSys),
			sysPD1: uint32(pd1Sys),
			sysPD2: uint32(pd2Sys),
		}
	}
}

func configurePLLs(xtalHz uint32) {
	for i := range pllTable {
		if pllTable[i].xtalHz == xtalHz {
			pllSys.init(
				1,
				pllTable[i].sysFB,
				pllTable[i].sysPD1,
				pllTable[i].sysPD2,
			)
			// Match in-tree USB PLL dividers until both paths adopt the 750 to 1600 MHz VCO range.
			// RP2040 datasheet section 2.18.2, https://datasheets.raspberrypi.com/rp2040/rp2040-datasheet.pdf
			pllUSB.init(1, 40, 5, 2)
			return
		}
	}
	panic("machine: unsupported crystal frequency")
}
