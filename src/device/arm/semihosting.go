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
const (
	// Hardware vector reason codes
	SemihostingBranchThroughZero = 20000
	SemihostingUndefinedInstr    = 20001
	SemihostingSoftwareInterrupt = 20002
	SemihostingPrefetchAbort     = 20003
	SemihostingDataAbort         = 20004
	SemihostingAddressException  = 20005
	SemihostingIRQ               = 20006
	SemihostingFIQ               = 20007

	// Software reason codes
	SemihostingBreakPoint          = 20020
	SemihostingWatchPoint          = 20021
	SemihostingStepComplete        = 20022
	SemihostingRunTimeErrorUnknown = 20023
	SemihostingInternalError       = 20024
	SemihostingUserInterruption    = 20025
	SemihostingApplicationExit     = 20026
	SemihostingStackOverflow       = 20027
	SemihostingDivisionByZero      = 20028
	SemihostingOSSpecific          = 20029
)

// Call a semihosting function.
// TODO: implement it here using inline assembly.
//go:linkname SemihostingCall SemihostingCall
func SemihostingCall(num int, arg uintptr) int
