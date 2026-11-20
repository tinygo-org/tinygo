//go:build (rp2040 || rp2350) && boardpackage

package machine

type pllEntry struct {
	xtalHz uint32
	sysFB  uint32
	sysRef uint32
	sysPD1 uint32
	sysPD2 uint32
	usbFB  uint32
	usbRef uint32
	usbPD1 uint32
	usbPD2 uint32
}

var supportedXTALs = [...]uint32{12 * MHz}

var pllTable [len(supportedXTALs)]pllEntry

func xtalFreqHz() uint32 {
	return boardXTAL()
}

// Precompute supported crystal PLL divider tables at compile time via interp.
func init() {
	for i, xtal := range supportedXTALs {
		outOfRange := xtal < 1*MHz || xtal > 15*MHz
		if outOfRange || xtal%MHz != 0 {
			panic("machine: supportedXTAL entry must be whole MHz in 1..15 MHz")
		}
		fbSys, refSys, pd1Sys, pd2Sys, err := calcPLLDivs(
			pllSearch{LockRefDiv: 1},
			uint64(xtal),
			cpuFreq,
			MHz,
		)
		if err != nil {
			panic(err)
		}
		fbUSB, refUSB, pd1USB, pd2USB, err := calcPLLDivs(
			pllSearch{LockRefDiv: 1},
			uint64(xtal),
			48*MHz,
			MHz,
		)
		if err != nil {
			panic(err)
		}
		pllTable[i] = pllEntry{
			xtalHz: xtal,
			sysFB:  uint32(fbSys),
			sysRef: uint32(refSys),
			sysPD1: uint32(pd1Sys),
			sysPD2: uint32(pd2Sys),
			usbFB:  uint32(fbUSB),
			usbRef: uint32(refUSB),
			usbPD1: uint32(pd1USB),
			usbPD2: uint32(pd2USB),
		}
	}
}

func configurePLLs(xtalHz uint32) {
	for i := range pllTable {
		if pllTable[i].xtalHz == xtalHz {
			pllSys.init(
				pllTable[i].sysRef,
				pllTable[i].sysFB,
				pllTable[i].sysPD1,
				pllTable[i].sysPD2,
			)
			pllUSB.init(
				pllTable[i].usbRef,
				pllTable[i].usbFB,
				pllTable[i].usbPD1,
				pllTable[i].usbPD2,
			)
			return
		}
	}
	panic("machine: unsupported crystal frequency")
}
