//go:build amd64

package x86

const (
	CPUIDTimeStampCounter   = 0x15
	CPUIDProcessorFrequency = 0x16
)

type CPUExtendedFamily uint16

const (
	CPUFamilyIntelCore CPUExtendedFamily = 6
)

//export asmPause
func AsmPause()

//export asmReadRdtsc
func AsmReadRdtsc() uint64

//export asmCpuid
func AsmCpuid(index uint32, registerEax *uint32, registerEbx *uint32, registerEcx *uint32) int

var maxCpuidIndex uint32
var stdVendorName0 uint32
var stdCpuid1Eax uint32

func init() {
	AsmCpuid(0, &maxCpuidIndex, &stdVendorName0, nil)
	AsmCpuid(1, &stdCpuid1Eax, nil, nil)
}

func getExtendedCPUFamily() CPUExtendedFamily {
	family := CPUExtendedFamily((stdCpuid1Eax >> 8) & 0x0f)
	family += CPUExtendedFamily((stdCpuid1Eax >> 20) & 0xff)
	return family
}

func isIntel() bool {
	return stdVendorName0 == 0x756e6547
}

func isIntelFamilyCore() bool {
	return isIntel() && getExtendedCPUFamily() == CPUFamilyIntelCore
}

func InternalGetPerformanceCounterFrequency() uint64 {
	if maxCpuidIndex >= CPUIDTimeStampCounter {
		return cpuidCoreClockCalculateTSCFrequency()
	}
	return 0
}

func cpuidCoreClockCalculateTSCFrequency() uint64 {
	var eax uint32
	var ebx uint32
	var ecx uint32

	AsmCpuid(CPUIDTimeStampCounter, &eax, &ebx, &ecx)
	if eax == 0 || ebx == 0 {
		return 0
	}

	coreCrystalFrequency := uint64(ecx)
	if coreCrystalFrequency == 0 {
		if !isIntelFamilyCore() {
			return 0
		}
		coreCrystalFrequency = 24000000
	}

	return ((coreCrystalFrequency * uint64(ebx)) + (uint64(eax) / 2)) / uint64(eax)
}
