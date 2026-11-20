//go:build (rp2040 || rp2350) && !boardpackage

package machine

func xtalFreqHz() uint32 {
	return xoscFreq * MHz
}

var pllsysFB, pllsysPD1, pllsysPD2 uint32

// Compute clock dividers at compile time via interp.
func init() {
	fb, _, pd1, pd2, err := calcPLLDivs(pllSearch{LockRefDiv: 1}, uint64(xtalFreqHz()), cpuFreq, MHz)
	if err != nil {
		panic(err)
	}
	pllsysFB, pllsysPD1, pllsysPD2 = uint32(fb), uint32(pd1), uint32(pd2)
}

func configurePLLs(xtalHz uint32) {
	pllSys.init(1, pllsysFB, pllsysPD1, pllsysPD2)
	pllUSB.init(1, 40, 5, 2)
}
