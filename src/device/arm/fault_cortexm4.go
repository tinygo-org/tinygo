// +build stm32,stm32f103xx sam,atsamd51 nrf,nrf52 nrf,nrf52840 stm32,stm32f407 nxp,mk66f18

package arm

// GetFaultStatus reads the System Control Block Configurable Fault Status
// Register and returns it as a FaultStatus.
func GetFaultStatus() FaultStatus { return FaultStatus(SCB.CFSR.Get()) }

// Address returns the MemManage Fault Address Register if the fault status
// indicates the address is valid.
//
// "If a MemManage fault occurs and is escalated to a HardFault because of
// priority, the HardFault handler must set this bit to 0. This prevents
// problems on return to a stacked active MemManage fault handler whose MMAR
// value has been overwritten."
func (fs MemFaultStatus) Address() (uintptr, bool) {
	if fs&SCB_CFSR_MMARVALID == 0 {
		return 0, false
	} else {
		return uintptr(SCB.MMAR.Get()), true
	}
}

// Address returns the BusFault Address Register if the fault status
// indicates the address is valid.
//
// "The processor sets this bit to 1 after a BusFault where the address is
// known. Other faults can set this bit to 0, such as a MemManage fault
// occurring later.
//
// If a BusFault occurs and is escalated to a hard fault because of priority,
// the hard fault handler must set this bit to 0. This prevents problems if
// returning to a stacked active BusFault handler whose BFAR value has been
// overwritten."
func (fs BusFaultStatus) Address() (uintptr, bool) {
	if fs&SCB_CFSR_BFARVALID == 0 {
		return 0, false
	} else {
		return uintptr(SCB.BFAR.Get()), true
	}
}
