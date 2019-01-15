// +build sam,atsamd21g18a

package runtime

import (
	"device/arm"
	"device/sam"
	"machine"
	"unsafe"
)

type timeUnit int64

//go:export Reset_Handler
func main() {
	preinit()
	initAll()
	mainWrapper()
	abort()
}

func init() {
	initClocks()

	// Clock for PORTS
	sam.PM.APBBMASK |= sam.PM_APBBMASK_PORT_

	// Clock SERCOM for Serial
	sam.PM.APBCMASK |= sam.PM_APBCMASK_SERCOM0_ |
		sam.PM_APBCMASK_SERCOM1_ |
		sam.PM_APBCMASK_SERCOM2_ |
		sam.PM_APBCMASK_SERCOM3_ |
		sam.PM_APBCMASK_SERCOM4_ |
		sam.PM_APBCMASK_SERCOM5_

	// Clock TC/TCC for Pulse and Analog
	sam.PM.APBCMASK |= sam.PM_APBCMASK_TCC0_ |
		sam.PM_APBCMASK_TCC1_ |
		sam.PM_APBCMASK_TCC2_ |
		sam.PM_APBCMASK_TC3_ |
		sam.PM_APBCMASK_TC4_ |
		sam.PM_APBCMASK_TC5_

	//machine.UART0.Configure(machine.UARTConfig{})
}

func putchar(c byte) {
	//machine.UART0.WriteByte(c)
}

func initClocks() {
	// ?
	// Turn on the digital interface clock
	sam.PM.APBAMASK |= sam.PM_APBAMASK_GCLK_

	// Set 1 Flash Wait State for 48MHz, cf tables 20.9 and 35.27 in SAMD21 Datasheet */
	sam.NVMCTRL.CTRLB |= (sam.NVMCTRL_CTRLB_RWS_HALF << sam.NVMCTRL_CTRLB_RWS_Pos)

	// Modify PRESCaler value of OSC8M to have 8MHz
	sam.SYSCTRL.OSC8M |= (sam.SYSCTRL_OSC8M_PRESC_0 << sam.SYSCTRL_OSC8M_PRESC_Pos)
	sam.SYSCTRL.OSC8M &^= (1 << sam.SYSCTRL_OSC8M_ONDEMAND_Pos)

	/* ----------------------------------------------------------------------------------------------
	 * 1) Enable OSC32K clock (Internal 32.768Hz oscillator)
	 */
	// u32_t fuse = *(u32_t *)FUSES_OSC32K_CAL_ADDR;
	// u32_t calib = (fuse & FUSES_OSC32K_CAL_Msk) >> FUSES_OSC32K_CAL_Pos;
	//
	// SYSCTRL->OSC32K.reg = SYSCTRL_OSC32K_CALIB(calib) |
	// 		      SYSCTRL_OSC32K_STARTUP(0x6u) |
	// 		      SYSCTRL_OSC32K_EN32K | SYSCTRL_OSC32K_ENABLE;
	//
	// from samd21g18a.h and nvmctrl.h:
	//
	// #define NVMCTRL_OTP4 0x00806020
	//
	// #define SYSCTRL_FUSES_OSC32K_CAL_ADDR (NVMCTRL_OTP4 + 4)
	// #define SYSCTRL_FUSES_OSC32K_CAL_Pos 6            /**< \brief (NVMCTRL_OTP4) OSC32K Calibration */
	// #define SYSCTRL_FUSES_OSC32K_CAL_Msk (0x7Fu << SYSCTRL_FUSES_OSC32K_CAL_Pos)
	// #define SYSCTRL_FUSES_OSC32K_CAL(value) ((SYSCTRL_FUSES_OSC32K_CAL_Msk & ((value) << SYSCTRL_FUSES_OSC32K_CAL_Pos)))
	fuse := *(*uint32)(unsafe.Pointer(uintptr(0x00806020) + 4))
	fuseMask := uint32(0x7f << 6)
	calib := (fuse & fuseMask) >> 6

	sam.SYSCTRL.OSC32K = sam.RegValue((sam.SYSCTRL_OSC32K_CALIB_Msk & (calib << sam.SYSCTRL_OSC32K_CALIB_Pos)) |
		(sam.SYSCTRL_OSC32K_STARTUP_Msk & (0x6 << sam.SYSCTRL_OSC32K_STARTUP_Pos)) |
		sam.SYSCTRL_OSC32K_EN32K |
		sam.SYSCTRL_OSC32K_ENABLE)

	// Wait for oscillator stabilization
	for (sam.SYSCTRL.PCLKSR & sam.SYSCTRL_PCLKSR_OSC32KRDY) == 0 {
	}

	// ??
	/* Software reset the module to ensure it is re-initialized correctly */
	/* Note: Due to synchronization, there is a delay from writing CTRL.SWRST until the reset is complete.
	 * CTRL.SWRST and STATUS.SYNCBUSY will both be cleared when the reset is complete, as described in chapter 13.8.1
	 */
	sam.GCLK.CTRL = sam.GCLK_CTRL_SWRST
	// Wait for reset to complete
	for (sam.GCLK.CTRL&sam.GCLK_CTRL_SWRST) > 0 && (sam.GCLK.STATUS&sam.GCLK_STATUS_SYNCBUSY) > 0 {
	}

	// dfll_init
	/* No prescaler */
	sam.GCLK.GENDIV = sam.RegValue((1 << sam.GCLK_GENDIV_ID_Pos) |
		(0 << sam.GCLK_GENDIV_DIV_Pos))
	// Wait for synchronization
	for (sam.GCLK.STATUS & sam.GCLK_STATUS_SYNCBUSY) > 0 {
	}

	/* Route XOSC32K to GCLK1 */
	// sam.GCLK.GENCTRL =
	//     GCLK_GENCTRL_ID(1) | GCLK_GENCTRL_SRC_XOSC32K | GCLK_GENCTRL_GENEN;
	sam.GCLK.GENCTRL = sam.RegValue((1 << sam.GCLK_GENCTRL_ID_Pos) |
		(sam.GCLK_GENCTRL_SRC_XOSC32K << sam.GCLK_GENCTRL_SRC_Pos) |
		sam.GCLK_GENCTRL_GENEN)
	// Wait for synchronization
	for (sam.GCLK.STATUS & sam.GCLK_STATUS_SYNCBUSY) > 0 {
	}

	// Route GCLK1 to multiplexer 1
	//GCLK->CLKCTRL.reg =
	//GCLK_CLKCTRL_ID(0) | GCLK_CLKCTRL_GEN_GCLK1 | GCLK_CLKCTRL_CLKEN;
	sam.GCLK.CLKCTRL = sam.RegValue16((0 << sam.GCLK_CLKCTRL_ID_Pos) |
		(sam.GCLK_CLKCTRL_GEN_GCLK1 << sam.GCLK_CLKCTRL_GEN_Pos) |
		sam.GCLK_CLKCTRL_CLKEN)
	// Wait for synchronization
	for (sam.GCLK.STATUS & sam.GCLK_STATUS_SYNCBUSY) > 0 {
	}

	// enable dfll
	sam.SYSCTRL.DFLLCTRL = sam.SYSCTRL_DFLLCTRL_ENABLE
	// Wait for ready
	for (sam.SYSCTRL.PCLKSR & sam.SYSCTRL_PCLKSR_DFLLRDY) == 0 {
	}

	// u32_t mul = (SOC_ATMEL_SAM0_MCK_FREQ_HZ +
	// 	SOC_ATMEL_SAM0_GCLK1_FREQ_HZ / 2) /
	// 	SOC_ATMEL_SAM0_GCLK1_FREQ_HZ;
	//
	// SYSCTRL->DFLLMUL.reg = SYSCTRL_DFLLMUL_CSTEP(31) |
	// 			  SYSCTRL_DFLLMUL_FSTEP(511) |
	// 			  SYSCTRL_DFLLMUL_MUL(mul);
	mul := ((machine.CPU_FREQUENCY + machine.SOC_ATMEL_SAM0_GCLK1_FREQ_HZ) / 2) /
		machine.SOC_ATMEL_SAM0_GCLK1_FREQ_HZ
	sam.SYSCTRL.DFLLMUL = sam.RegValue((31 << sam.SYSCTRL_DFLLMUL_CSTEP_Pos) |
		(511 << sam.SYSCTRL_DFLLMUL_FSTEP_Pos) |
		(mul << sam.SYSCTRL_DFLLMUL_MUL_Pos))
	// Wait for synchronization
	for (sam.SYSCTRL.PCLKSR & sam.SYSCTRL_PCLKSR_DFLLRDY) == 0 {
	}

	// SYSCTRL->DFLLCTRL.reg |= SYSCTRL_DFLLCTRL_MODE |
	// 			 SYSCTRL_DFLLCTRL_WAITLOCK |
	// 			 SYSCTRL_DFLLCTRL_QLDIS;
	sam.SYSCTRL.DFLLCTRL |= sam.SYSCTRL_DFLLCTRL_MODE |
		sam.SYSCTRL_DFLLCTRL_WAITLOCK |
		sam.SYSCTRL_DFLLCTRL_QLDIS
	// Wait for ready
	for (sam.SYSCTRL.PCLKSR & sam.SYSCTRL_PCLKSR_DFLLRDY) == 0 {
	}

	/* Enable the DFLL */
	sam.SYSCTRL.DFLLCTRL |= sam.SYSCTRL_DFLLCTRL_ENABLE

	// while (!SYSCTRL->PCLKSR.bit.DFLLLCKC || !SYSCTRL->PCLKSR.bit.DFLLLCKF) {
	// }
	for (sam.SYSCTRL.PCLKSR&sam.SYSCTRL_INTENCLR_DFLLLCKC) == 0 || (sam.SYSCTRL.PCLKSR&sam.SYSCTRL_INTENCLR_DFLLLCKF) == 0 {
	}
	// Wait for ready
	for (sam.SYSCTRL.PCLKSR & sam.SYSCTRL_PCLKSR_DFLLRDY) == 0 {
	}
	// end dfll init

	// gclks_init
	/* DFLL/1 -> GCLK0 */
	// GCLK->GENDIV.reg = GCLK_GENDIV_ID(0) | GCLK_GENDIV_DIV(0);
	sam.GCLK.GENDIV = sam.RegValue((0 << sam.GCLK_GENDIV_ID_Pos) |
		(0 << sam.GCLK_GENDIV_DIV_Pos))
	// Wait for synchronization
	for (sam.GCLK.STATUS & sam.GCLK_STATUS_SYNCBUSY) > 0 {
	}
	// GCLK->GENCTRL.reg = GCLK_GENCTRL_ID(0) | GCLK_GENCTRL_SRC_DFLL48M |
	// 		    GCLK_GENCTRL_IDC | GCLK_GENCTRL_GENEN;
	sam.GCLK.GENCTRL = sam.RegValue((1 << sam.GCLK_GENCTRL_ID_Pos) |
		(sam.GCLK_GENCTRL_SRC_DFLL48M << sam.GCLK_GENCTRL_SRC_Pos) |
		sam.GCLK_GENCTRL_IDC |
		sam.GCLK_GENCTRL_GENEN)
	// Wait for synchronization
	for (sam.GCLK.STATUS & sam.GCLK_STATUS_SYNCBUSY) > 0 {
	}

	/* OSC8M/1 -> GCLK3 */
	// GCLK->GENCTRL.reg =
	//     GCLK_GENCTRL_ID(3) | GCLK_GENCTRL_SRC_OSC8M | GCLK_GENCTRL_GENEN;
	sam.GCLK.GENCTRL = sam.RegValue((3 << sam.GCLK_GENCTRL_ID_Pos) |
		(sam.GCLK_GENCTRL_SRC_OSC8M << sam.GCLK_GENCTRL_SRC_Pos) |
		sam.GCLK_GENCTRL_GENEN)
	// Wait for synchronization
	for (sam.GCLK.STATUS & sam.GCLK_STATUS_SYNCBUSY) > 0 {
	}

	/* OSCULP32K/32 -> GCLK2 */
	//GCLK->GENDIV.reg = GCLK_GENDIV_ID(2) | GCLK_GENDIV_DIV(32 - 1);
	sam.GCLK.GENDIV = sam.RegValue((2 << sam.GCLK_GENDIV_ID_Pos) |
		(31 << sam.GCLK_GENDIV_DIV_Pos))
	// Wait for synchronization
	for (sam.GCLK.STATUS & sam.GCLK_STATUS_SYNCBUSY) > 0 {
	}

	// GCLK->GENCTRL.reg =
	//     GCLK_GENCTRL_ID(2) | GCLK_GENCTRL_SRC_OSC32K | GCLK_GENCTRL_GENEN;
	sam.GCLK.GENCTRL = sam.RegValue((2 << sam.GCLK_GENCTRL_ID_Pos) |
		(sam.GCLK_GENCTRL_SRC_OSC32K << sam.GCLK_GENCTRL_SRC_Pos) |
		sam.GCLK_GENCTRL_GENEN)
	// Wait for synchronization
	for (sam.GCLK.STATUS & sam.GCLK_STATUS_SYNCBUSY) > 0 {
	}

	// end gclks_init

	/* Set the CPU, APBA, B, and C dividers */
	sam.PM.CPUSEL = sam.PM_CPUSEL_CPUDIV_DIV1
	sam.PM.APBASEL = sam.PM_APBASEL_APBADIV_DIV1
	sam.PM.APBBSEL = sam.PM_APBBSEL_APBBDIV_DIV1
	sam.PM.APBCSEL = sam.PM_APBCSEL_APBCDIV_DIV1
}

const tickMicros = 1

var (
	timestamp        timeUnit // microseconds since boottime
	timerLastCounter uint64
)

//go:volatile
type isrFlag bool

var timerWakeup isrFlag

// sleepTicks should sleep for specific number of microseconds.
func sleepTicks(d timeUnit) {
	// TODO: use a real timer here
	for i := 0; i < int(d/535); i++ {
		arm.Asm("")
	}
}

// number of ticks (microseconds) since start.
func ticks() timeUnit {
	return 0 // TODO:
}
