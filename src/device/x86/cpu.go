//go:build amd64

package x86

const (
	// CPUID_TIME_STAMP_COUNTER
	// EAX  Returns processor base frequency information described by the
	//      type CPUID_PROCESSOR_FREQUENCY_EAX.
	// EBX  Returns maximum frequency information described by the type
	//      CPUID_PROCESSOR_FREQUENCY_EBX.
	// ECX  Returns bus frequency information described by the type
	//      CPUID_PROCESSOR_FREQUENCY_ECX.
	// EDX  Reserved.
	CPUID_TIME_STAMP_COUNTER = 0x15

	// CPUID_PROCESSOR_FREQUENCY
	// EAX  Returns processor base frequency information described by the
	//      type CPUID_PROCESSOR_FREQUENCY_EAX.
	// EBX  Returns maximum frequency information described by the type
	//      CPUID_PROCESSOR_FREQUENCY_EBX.
	// ECX  Returns bus frequency information described by the type
	//      CPUID_PROCESSOR_FREQUENCY_ECX.
	// EDX  Reserved.
	CPUID_PROCESSOR_FREQUENCY = 0x16
)

type CpuExtendedFamily uint16

const (
	// AMD
	CPU_FAMILY_AMD_11H CpuExtendedFamily = 0x11
	// Intel
	CPU_FAMILY_INTEL_CORE    CpuExtendedFamily = 6
	CPU_MODEL_NEHALEM        CpuExtendedFamily = 0x1e
	CPU_MODEL_NEHALEM_EP     CpuExtendedFamily = 0x1a
	CPU_MODEL_NEHALEM_EX     CpuExtendedFamily = 0x2e
	CPU_MODEL_WESTMERE       CpuExtendedFamily = 0x25
	CPU_MODEL_WESTMERE_EP    CpuExtendedFamily = 0x2c
	CPU_MODEL_WESTMERE_EX    CpuExtendedFamily = 0x2f
	CPU_MODEL_SANDYBRIDGE    CpuExtendedFamily = 0x2a
	CPU_MODEL_SANDYBRIDGE_EP CpuExtendedFamily = 0x2d
	CPU_MODEL_IVYBRIDGE_EP   CpuExtendedFamily = 0x3a
	CPU_MODEL_HASWELL_E3     CpuExtendedFamily = 0x3c
	CPU_MODEL_HASWELL_E7     CpuExtendedFamily = 0x3f
	CPU_MODEL_BROADWELL      CpuExtendedFamily = 0x3d
)

func GetExtendedCpuFamily() CpuExtendedFamily {
	var family CpuExtendedFamily
	family = CpuExtendedFamily((stdCpuid1Eax >> 8) & 0x0f)
	family += CpuExtendedFamily((stdCpuid1Eax >> 20) & 0xff)
	return family
}

//export asmPause
func AsmPause()

//export asmReadRdtsc
func AsmReadRdtsc() uint64

//export asmCpuid
func AsmCpuid(index uint32, registerEax *uint32, registerEbx *uint32, registerEcx *uint32, registerEdx *uint32) int

var maxCpuidIndex uint32
var stdVendorName0 uint32
var stdCpuid1Eax uint32
var stdCpuid1Ebx uint32
var stdCpuid1Ecx uint32
var stdCpuid1Edx uint32

func init() {
	AsmCpuid(0, &maxCpuidIndex, &stdVendorName0, nil, nil)
	AsmCpuid(1, &stdCpuid1Eax, &stdCpuid1Ebx, &stdCpuid1Ecx, &stdCpuid1Edx)
}

func GetMaxCpuidIndex() uint32 {
	return maxCpuidIndex
}

func IsIntel() bool {
	return stdVendorName0 == 0x756e6547
}

func IsIntelFamilyCore() bool {
	return IsIntel() && GetExtendedCpuFamily() == CPU_FAMILY_INTEL_CORE
}

func IsAmd() bool {
	return stdVendorName0 == 0x68747541
}

func InternalGetPerformanceCounterFrequency() uint64 {
	if maxCpuidIndex >= CPUID_TIME_STAMP_COUNTER {
		return CpuidCoreClockCalculateTscFrequency()
	}

	return 0
}

func CpuidCoreClockCalculateTscFrequency() uint64 {
	var TscFrequency uint64
	var CoreXtalFrequency uint64
	var RegEax uint32
	var RegEbx uint32
	var RegEcx uint32

	AsmCpuid(CPUID_TIME_STAMP_COUNTER, &RegEax, &RegEbx, &RegEcx, nil)

	// If EAX or EBX returns 0, the XTAL ratio is not enumerated.
	if (RegEax == 0) || (RegEbx == 0) {
		return 0
	}

	// If ECX returns 0, the XTAL frequency is not enumerated.
	// And PcdCpuCoreCrystalClockFrequency defined should base on processor series.
	//
	if RegEcx == 0 {
		// Specifies CPUID Leaf 0x15 Time Stamp Counter and Nominal Core Crystal Clock Frequency.
		// https://github.com/torvalds/linux/blob/master/tools/power/x86/turbostat/turbostat.c
		if IsIntelFamilyCore() {
			switch GetExtendedCpuFamily() {
			case 0x5F: // INTEL_FAM6_ATOM_GOLDMONT_D
				CoreXtalFrequency = 25000000
			case 0x5C: // INTEL_FAM6_ATOM_GOLDMONT
				CoreXtalFrequency = 19200000
			case 0x7A: // INTEL_FAM6_ATOM_GOLDMONT_PLUS
				CoreXtalFrequency = 19200000
			default:
				CoreXtalFrequency = 24000000
			}
		} else {
			return 0
		}
	} else {
		CoreXtalFrequency = uint64(RegEcx)
	}

	// Calculate TSC frequency = (ECX, Core Xtal Frequency) * EBX/EAX
	TscFrequency = ((CoreXtalFrequency * uint64(RegEbx)) + (uint64(RegEax) / 2)) / uint64(RegEax)
	return TscFrequency
}
