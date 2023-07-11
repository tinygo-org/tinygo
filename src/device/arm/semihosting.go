package arm

// Semihosting commands.
// http://infocenter.arm.com/help/index.jsp?topic=/com.arm.doc.dui0471c/Bgbjhiea.html
const (
	// Regular semihosting calls
	SemihostingClock      = 0x10
	SemihostingClose      = 0x02
	SemihostingElapsed    = 0x30
	SemihostingErrno      = 0x13
	SemihostingFileLen    = 0x0C
	SemihostingGetCmdline = 0x15
	SemihostingHeapInfo   = 0x16
	SemihostingIsError    = 0x08
	SemihostingIsTTY      = 0x09
	SemihostingOpen       = 0x01
	SemihostingRead       = 0x06
	SemihostingReadByte   = 0x07
	SemihostingRemove     = 0x0E
	SemihostingRename     = 0x0F
	SemihostingSeek       = 0x0A
	SemihostingSystem     = 0x12
	SemihostingTickFreq   = 0x31
	SemihostingTime       = 0x11
	SemihostingTmpName    = 0x0D
	SemihostingWrite      = 0x05
	SemihostingWrite0     = 0x04
	SemihostingWriteByte  = 0x03

	// Angel semihosting calls
	SemihostingEnterSVC        = 0x17
	SemihostingReportException = 0x18
)

// Special codes for the Angel Semihosting interface.
// https://www.keil.com/support/man/docs/armcc/armcc_pge1358787050566.htm
const (
	// Hardware vector reason codes
	SemihostingBranchThroughZero = 0x20000
	SemihostingUndefinedInstr    = 0x20001
	SemihostingSoftwareInterrupt = 0x20002
	SemihostingPrefetchAbort     = 0x20003
	SemihostingDataAbort         = 0x20004
	SemihostingAddressException  = 0x20005
	SemihostingIRQ               = 0x20006
	SemihostingFIQ               = 0x20007

	// Software reason codes
	SemihostingBreakPoint          = 0x20020
	SemihostingWatchPoint          = 0x20021
	SemihostingStepComplete        = 0x20022
	SemihostingRunTimeErrorUnknown = 0x20023
	SemihostingInternalError       = 0x20024
	SemihostingUserInterruption    = 0x20025
	SemihostingApplicationExit     = 0x20026
	SemihostingStackOverflow       = 0x20027
	SemihostingDivisionByZero      = 0x20028
	SemihostingOSSpecific          = 0x20029
)

// Call a semihosting function.
// TODO: implement it here using inline assembly.
//
//go:linkname SemihostingCall SemihostingCall
func SemihostingCall(num int, arg uintptr) int
