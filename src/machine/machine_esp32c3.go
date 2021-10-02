// +build esp32c3

package machine

import (
	"device"
	"device/esp"
	"device/riscv"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

// CPUFrequency returns the current CPU frequency of the chip.
// Currently it is a fixed frequency but it may allow changing in the future.
func CPUFrequency() uint32 {
	return 160e6 // 160MHz
}

const (
	PinOutput PinMode = iota
	PinInput
	PinInputPullup
	PinInputPulldown

	sclk_freq = 80 * 1000000
	max_div   = 0xfff // UART divider integer part only has 12 bits

	UART_DATA_5_BITS   = 0x0
	UART_DATA_6_BITS   = 0x1
	UART_DATA_7_BITS   = 0x2
	UART_DATA_8_BITS   = 0x3
	UART_DATA_BITS_MAX = 0x4

	UART_STOP_BITS_1   = 0x1
	UART_STOP_BITS_1_5 = 0x2
	UART_STOP_BITS_2   = 0x3
	UART_STOP_BITS_MAX = 0x4

	UART_FIFO_LEN             = 128
	UART_EMPTY_THRESH_DEFAULT = 10
	UART_FULL_THRESH_DEFAULT  = 120
)

const (
	UART_RXFIFO_FULL_INT_ENA      = 1 << iota // This is the enable bit for UART_RXFIFO_FULL_INT_ST register. (R/W)
	UART_TXFIFO_EMPTY_INT_ENA                 // This is the enable bit for UART_TXFIFO_EMPTY_INT_ST register. (R/W)
	UART_PARITY_ERR_INT_ENA                   // This is the enable bit for UART_PARITY_ERR_INT_ST register. (R/W)
	UART_FRM_ERR_INT_ENA                      // This is the enable bit for UART_FRM_ERR_INT_ST register. (R/W)
	UART_RXFIFO_OVF_INT_ENA                   // This is the enable bit for UART_RXFIFO_OVF_INT_ST register. (R/W)
	UART_DSR_CHG_INT_ENA                      // This is the enable bit for UART_DSR_CHG_INT_ST register. (R/W)
	UART_CTS_CHG_INT_ENA                      // This is the enable bit for UART_CTS_CHG_INT_ST register. (R/W)
	UART_BRK_DET_INT_ENA                      // This is the enable bit for UART_BRK_DET_INT_ST register. (R/W)
	UART_RXFIFO_TOUT_INT_ENA                  // This is the enable bit for UART_RXFIFO_TOUT_INT_ST register. (R/W)
	UART_SW_XON_INT_ENA                       // This is the enable bit for UART_SW_XON_INT_ST register. (R/W)
	UART_SW_XOFF_INT_ENA                      // This is the enable bit for UART_SW_XOFF_INT_ST register. (R/W)
	UART_GLITCH_DET_INT_ENA                   // This is the enable bit for UART_GLITCH_DET_INT_ST register. (R/W)
	UART_TX_BRK_DONE_INT_ENA                  // This is the enable bit for UART_TX_BRK_DONE_INT_ST register. (R/W)
	UART_TX_BRK_IDLE_DONE_INT_ENA             // This is the enable bit for UART_TX_BRK_IDLE_DONE_INT_ST register. (R/W)
	UART_TX_DONE_INT_ENA                      // This is the enable bit for UART_TX_DONE_INT_ST register. (R/W)
	UART_RS485_PARITY_ERR_INT_ENA             // This is the enable bit for UART_RS485_PARITY_ERR_INT_ST register. (R/W)
	UART_RS485_FRM_ERR_INT_ENA                // This is the enable bit for UART_RS485_PARITY_ERR_INT_ST register. (R/W)
	UART_RS485_CLASH_INT_ENA                  // This is the enable bit for UART_RS485_CLASH_INT_ST register. (R/W)
	UART_AT_CMD_CHAR_DET_INT_ENA              // This is the enable bit for UART_AT_CMD_CHAR_DET_INT_ST register. (R/W)
	UART_WAKEUP_INT_ENA                       // This is the enable bit for UART_WAKEUP_INT_ST register. (R/W)

)

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	if p == NoPin {
		// This simplifies pin configuration in peripherals such as SPI.
		return
	}

	var muxConfig uint32

	// Configure this pin as a GPIO pin.
	const function = 1 // function 1 is GPIO for every pin
	muxConfig |= function << esp.IO_MUX_GPIO_MCU_SEL_Pos

	// Make this pin an input pin (always).
	muxConfig |= esp.IO_MUX_GPIO_FUN_IE

	// Set drive strength: 0 is lowest, 3 is highest.
	muxConfig |= 2 << esp.IO_MUX_GPIO_FUN_DRV_Pos

	// Select pull mode.
	if config.Mode == PinInputPullup {
		muxConfig |= esp.IO_MUX_GPIO_FUN_WPU
	} else if config.Mode == PinInputPulldown {
		muxConfig |= esp.IO_MUX_GPIO_FUN_WPD
	}

	// Configure the pad with the given IO mux configuration.
	p.mux().Set(muxConfig)

	// Set the output signal to the simple GPIO output.
	p.outFunc().Set(0x80)

	switch config.Mode {
	case PinOutput:
		// Set the 'output enable' bit.
		esp.GPIO.ENABLE_W1TS.Set(1 << p)
	case PinInput, PinInputPullup, PinInputPulldown:
		// Clear the 'output enable' bit.
		esp.GPIO.ENABLE_W1TC.Set(1 << p)
	}
}

// outFunc returns the FUNCx_OUT_SEL_CFG register used for configuring the
// output function selection.
func (p Pin) outFunc() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(&esp.GPIO.FUNC0_OUT_SEL_CFG)) + uintptr(p)*4)))
}

// inFunc returns the FUNCy_IN_SEL_CFG register used for configuring the input
// function selection.
func inFunc(signal uint32) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(&esp.GPIO.FUNC0_IN_SEL_CFG)) + uintptr(signal)*4)))
}

// mux returns the I/O mux configuration register corresponding to the given
// GPIO pin.
func (p Pin) mux() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(&esp.IO_MUX.GPIO0)) + uintptr(p)*4)))
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p Pin) Set(value bool) {
	if value {
		reg, mask := p.portMaskSet()
		reg.Set(mask)
	} else {
		reg, mask := p.portMaskClear()
		reg.Set(mask)
	}
}

// Return the register and mask to enable a given GPIO pin. This can be used to
// implement bit-banged drivers.
//
// Warning: only use this on an output pin!
func (p Pin) PortMaskSet() (*uint32, uint32) {
	reg, mask := p.portMaskSet()
	return &reg.Reg, mask
}

// Return the register and mask to disable a given GPIO pin. This can be used to
// implement bit-banged drivers.
//
// Warning: only use this on an output pin!
func (p Pin) PortMaskClear() (*uint32, uint32) {
	reg, mask := p.portMaskClear()
	return &reg.Reg, mask
}

func (p Pin) portMaskSet() (*volatile.Register32, uint32) {
	return &esp.GPIO.OUT_W1TS, 1 << p
}

func (p Pin) portMaskClear() (*volatile.Register32, uint32) {
	return &esp.GPIO.OUT_W1TC, 1 << p
}

var DefaultUART = UART0

var (
	UART0  = &_UART0
	_UART0 = UART{Bus: esp.UART0, Buffer: NewRingBuffer()}
	UART1  = &_UART1
	_UART1 = UART{Bus: esp.UART1, Buffer: NewRingBuffer()}
)

type UART struct {
	Bus    *esp.UART_Type
	Buffer *RingBuffer
}

// ISR - Interrupt Status Register

// @driver/uart.c
//
// esp_err_t uart_driver_install(
// 	uart_port_t uart_num,
// 	int rx_buffer_size,
// 	int tx_buffer_size,
// 	int event_queue_size,
// 	QueueHandle_t *uart_queue,
// 	int intr_alloc_flags)
func (uart *UART) Configure(config *UARTConfig) error {

	uart.initUART()

	// enable register synchronization by clearing UART_UPDATE_CTRL
	uart.Bus.ID.ClearBits(esp.UART_ID_REG_UPDATE_Msk)
	// wait for Core Clock to ready for configuration
	for uart.Bus.ID.HasBits(esp.UART_ID_REG_UPDATE_Msk) {
		riscv.Asm("nop")
	}

	uart.configure(config)

	// Finish Config process
	uart.Bus.ID.SetBits(esp.UART_ID_REG_UPDATE)

	uart.enableTransmitter()
	uart.enableReceiver()
	uart.setupInterrupt()

	// device.Asm("ebreak")

	// Start TX/RX
	// - enable TX/RX clock
	uart.Bus.CLK_CONF.SetBits(esp.UART_CLK_CONF_TX_SCLK_EN_Msk | esp.UART_CLK_CONF_RX_SCLK_EN_Msk)

	// device.Asm("ebreak")

	// Setup GPIO Pins
	uart.setupPins(config)

	// device.Asm("ebreak")

	return nil
}

func (uart *UART) initUART() {
	// Initialize UARTn
	uart0 := uart.Bus == esp.UART0
	// - enable the clock for UART RAM
	esp.SYSTEM.PERIP_CLK_EN0.SetBits(esp.SYSTEM_PERIP_CLK_EN0_UART_MEM_CLK_EN)
	// - enable APB_CLK for UARTn
	// - clear SYSTEM_UARTn_RST
	if uart0 {
		esp.SYSTEM.PERIP_CLK_EN0.SetBits(esp.SYSTEM_PERIP_CLK_EN0_UART_CLK_EN)
		esp.SYSTEM.PERIP_RST_EN0.ClearBits(esp.SYSTEM_PERIP_RST_EN0_UART_RST)
	} else {
		esp.SYSTEM.PERIP_CLK_EN0.SetBits(esp.SYSTEM_PERIP_CLK_EN0_UART1_CLK_EN)
		esp.SYSTEM.PERIP_RST_EN0.ClearBits(esp.SYSTEM_PERIP_RST_EN0_UART1_RST)
	}
	// - write 1 to UART_RST_CORE
	uart.Bus.CLK_CONF.SetBits(esp.UART_CLK_CONF_RST_CORE)
	// - write 1 to SYSTEM_UARTn_RST
	// - clear SYSTEM_UARTn_RST
	if uart0 {
		esp.SYSTEM.PERIP_RST_EN0.SetBits(esp.SYSTEM_PERIP_RST_EN0_UART_RST)
		esp.SYSTEM.PERIP_RST_EN0.ClearBits(esp.SYSTEM_PERIP_RST_EN0_UART_RST)
	} else {
		esp.SYSTEM.PERIP_RST_EN0.SetBits(esp.SYSTEM_PERIP_RST_EN0_UART1_RST)
		esp.SYSTEM.PERIP_RST_EN0.ClearBits(esp.SYSTEM_PERIP_RST_EN0_UART1_RST)
	}
	// - clear UART_RST_CORE
	uart.Bus.CLK_CONF.ClearBits(esp.UART_CLK_CONF_RST_CORE)
}

func (uart *UART) configure(config *UARTConfig) {
	// Write static registers
	// - disbale TX/RX clock to make sure the UART transmitter or receiver is not at work
	uart.Bus.CLK_CONF.ClearBits(esp.UART_CLK_CONF_TX_SCLK_EN_Msk | esp.UART_CLK_CONF_RX_SCLK_EN_Msk)
	// - Set default clock source (UART_SCLK_APB)
	// UART_SCLK_SEL UART clock source select. 1: APB_CLK; 2: FOSC_CLK; 3: XTAL_CLK.
	uart.Bus.CLK_CONF.SetBits(0x1 << esp.UART_CLK_CONF_SCLK_SEL_Pos)
	// - Set baud rate
	m := config.BaudRate * 0xfff
	sclk_div := (sclk_freq + m - 1) / m
	clk_div := (sclk_freq << 4) / (config.BaudRate * sclk_div)
	// The baud rate configuration register is divided into an integer part and a fractional part.
	uart.Bus.CLKDIV.Set(((clk_div >> 4) << esp.UART_CLKDIV_CLKDIV_Pos) & esp.UART_CLKDIV_CLKDIV_Msk)
	uart.Bus.CLKDIV.Set(((clk_div & 0xf) << esp.UART_CLKDIV_FRAG_Pos) & esp.UART_CLKDIV_FRAG_Msk)
	uart.Bus.CLK_CONF.SetBits((sclk_div - 1) << esp.UART_CLK_CONF_SCLK_DIV_NUM_Pos)
	// - Set UART mode.
	uart.Bus.RS485_CONF.ClearBits(esp.UART_RS485_CONF_RS485_EN_Msk | esp.UART_RS485_CONF_RS485TX_RX_EN_Msk | esp.UART_RS485_CONF_RS485RXBY_TX_EN_Msk)
	uart.Bus.CONF0.ClearBits(esp.UART_CONF0_IRDA_EN)
	// - Disable UART parity
	uart.Bus.CONF0.ClearBits(esp.UART_CONF0_PARITY_EN)
	// - 8-bit world
	uart.Bus.CONF0.SetBits(UART_DATA_8_BITS << esp.UART_CONF0_BIT_NUM_Pos)
	// - 1-bit stop bit
	uart.Bus.CONF0.SetBits(UART_STOP_BITS_1 << esp.UART_CONF0_SW_DTR_Pos)
	// - Set tx idle
	uart.Bus.IDLE_CONF.ClearBits(esp.UART_IDLE_CONF_TX_IDLE_NUM_Msk)
	// - Disable hw-flow control
	uart.Bus.CONF0.ClearBits(esp.UART_CONF0_TX_FLOW_EN)
	uart.Bus.CONF1.ClearBits(esp.UART_CONF1_RX_FLOW_EN)

	// Write other registers
}

func (uart *UART) enableTransmitter() {
	uart.Bus.CONF0.SetBits(esp.UART_CONF0_TXFIFO_RST)
	uart.Bus.CONF0.ClearBits(esp.UART_CONF0_TXFIFO_RST)
	// TXINFO empty threshold is when txfifo_empty_int interrupt produced after the amount of data in Tx-FIFO is less than this register value.
	uart.Bus.CONF1.SetBits((UART_EMPTY_THRESH_DEFAULT << esp.UART_CONF1_TXFIFO_EMPTY_THRHD_Pos) & esp.UART_CONF1_TXFIFO_EMPTY_THRHD_Msk)
	// enable UART_TXFIFO_EMPTY_INT interrupt by setting UART_TXFIFO_EMPTY_INT_ENA
	uart.Bus.INT_ENA.SetBits(esp.UART_INT_ENA_TXFIFO_EMPTY_INT_ENA)
}

func (uart *UART) enableReceiver() {
	uart.Bus.CONF0.SetBits(esp.UART_CONF0_RXFIFO_RST)
	uart.Bus.CONF0.ClearBits(esp.UART_CONF0_RXFIFO_RST)
	// configure RXFIFO’s full threshold via UART_RXFIFO_FULL_THRHD
	uart.Bus.CONF1.SetBits((UART_FULL_THRESH_DEFAULT << esp.UART_CONF1_RXFIFO_FULL_THRHD_Pos) & esp.UART_CONF1_RXFIFO_FULL_THRHD_Msk)
	// enable UART_RXFIFO_FULL_INT interrupt by setting UART_RXFIFO_FULL_INT_ENA
	uart.Bus.INT_ENA.SetBits(esp.UART_INT_ENA_RXFIFO_FULL_INT_ENA)

	// detect UART_TXFIFO_FULL_INT and wait until the RXFIFO is full;
	// • read data from RXFIFO via UART_RXFIFO_RD_BYTE, and obtain the number of bytes received in RXFIFO via UART_RXFIFO_CNT.
}

func (uart *UART) setupInterrupt() {
	// Disable interrupts
	uart.Bus.INT_ENA.ClearBits(0x0fffff)
	// Clear the UART interrupt status
	uart.Bus.INT_CLR.SetBits(0x0fffff)
	uart.Bus.INT_CLR.ClearBits(0x0fffff)

	interrupt.AddHandler(7, &esp.INTERRUPT_CORE0.UART1_INTR_MAP, uart.uartEvent)
	// println("AddHandler done")

	coreInterrupt := riscv.DisableInterrupts()
	// enable all interrupts
	x := uint32(1) << 20
	x--
	x &= ^uint32(UART_TXFIFO_EMPTY_INT_ENA | UART_PARITY_ERR_INT_ENA | UART_TX_BRK_IDLE_DONE_INT_ENA | UART_WAKEUP_INT_ENA)
	uart.Bus.INT_ENA.Set(x)

	riscv.EnableInterrupts(coreInterrupt)

	// device.Asm("ebreak")

	// println("setupInterrupt done")

	// // ETS_UART0_INTR_SOURCE
	// // ETS_UART1_INTR_SOURCE

	// esp_err_t esp_intr_alloc(int source, int flags, intr_handler_t handler, void *arg, intr_handle_t *ret_handle)
	// ret = esp_intr_alloc(uart_periph_signal[uart_num].irq, intr_alloc_flags, fn, arg, handle);

	// intr_alloc.c

	//     uart_intr_config_t uart_intr = {
	//         .intr_enable_mask = UART_INTR_CONFIG_FLAG,
	//         .rxfifo_full_thresh = UART_FULL_THRESH_DEFAULT,
	//         .rx_timeout_thresh = UART_TOUT_THRESH_DEFAULT,
	//         .txfifo_empty_intr_thresh = UART_EMPTY_THRESH_DEFAULT,
	//     };

	// 	   uart_hal_disable_intr_mask(&(uart_context[uart_num].hal), UART_LL_INTR_MASK);
	//     uart_hal_clr_intsts_mask(&(uart_context[uart_num].hal), UART_LL_INTR_MASK);
	//     r = uart_isr_register(uart_num,
	// 			uart_rx_intr_handler_default,
	//  		p_uart_obj[uart_num],
	//  		intr_alloc_flags,
	// 			&p_uart_obj[uart_num]->intr_handle);
	//     if (r != ESP_OK) {
	//         goto err;
	//     }
	//     r = uart_intr_config(uart_num, &uart_intr);
	//     if (r != ESP_OK) {
	//         goto err;
	//     }
}

//go:extern void gpio_matrix_in(uint32_t gpio, uint32_t signal_idx, bool inv);
func gpio_matrix_in(gpio, signla uint32, inv byte)

func (uart *UART) setupPins(config *UARTConfig) error {

	config.RX.Configure(PinConfig{Mode: PinInputPullup})
	// gpio_matrix_in(5, 6, 0)

	// if (rx_io_num >= 0 && !uart_try_set_iomux_pin(uart_num, rx_io_num, SOC_UART_RX_PIN_IDX)) {
	//     gpio_hal_iomux_func_sel(GPIO_PIN_MUX_REG[rx_io_num], PIN_FUNC_GPIO);
	//     gpio_set_pull_mode(rx_io_num, GPIO_PULLUP_ONLY);
	//     gpio_set_direction(rx_io_num, GPIO_MODE_INPUT);
	// #define SOC_UART_RX_PIN_IDX  (1)
	// #define UART_PERIPH_SIGNAL(IDX, PIN) (uart_periph_signal[(IDX)].pins[(PIN)].signal)
	// #define U0RXD_IN_IDX                  6
	//     esp_rom_gpio_connect_in_signal(rx_io_num, UART_PERIPH_SIGNAL(uart_num, SOC_UART_RX_PIN_IDX), 0);
	//     esp_rom_gpio_connect_in_signal(5, 6, 0);
	// }

	// config.TX.Configure(PinConfig{Mode: PinOutput})

	// esp_err_t uart_set_pin(uart_port_t uart_num, int tx_io_num, int rx_io_num, int rts_io_num, int cts_io_num)

	// ESP_RETURN_ON_FALSE((uart_num >= 0), ESP_FAIL, UART_TAG, "uart_num error");
	// ESP_RETURN_ON_FALSE((uart_num < UART_NUM_MAX), ESP_FAIL, UART_TAG, "uart_num error");
	// ESP_RETURN_ON_FALSE((tx_io_num < 0 || (GPIO_IS_VALID_OUTPUT_GPIO(tx_io_num))), ESP_FAIL, UART_TAG, "tx_io_num error");
	// ESP_RETURN_ON_FALSE((rx_io_num < 0 || (GPIO_IS_VALID_GPIO(rx_io_num))), ESP_FAIL, UART_TAG, "rx_io_num error");
	// ESP_RETURN_ON_FALSE((rts_io_num < 0 || (GPIO_IS_VALID_OUTPUT_GPIO(rts_io_num))), ESP_FAIL, UART_TAG, "rts_io_num error");
	// ESP_RETURN_ON_FALSE((cts_io_num < 0 || (GPIO_IS_VALID_GPIO(cts_io_num))), ESP_FAIL, UART_TAG, "cts_io_num error");

	// /* In the following statements, if the io_num is negative, no need to configure anything. */
	// if (tx_io_num >= 0 && !uart_try_set_iomux_pin(uart_num, tx_io_num, SOC_UART_TX_PIN_IDX)) {
	//     gpio_hal_iomux_func_sel(GPIO_PIN_MUX_REG[tx_io_num], PIN_FUNC_GPIO);

	// static inline void gpio_ll_iomux_func_sel(uint32_t pin_name, uint32_t func)
	// {
	// 	if (pin_name == IO_MUX_GPIO18_REG || pin_name == IO_MUX_GPIO19_REG) {
	// 		CLEAR_PERI_REG_MASK(USB_DEVICE_CONF0_REG, USB_DEVICE_USB_PAD_ENABLE);
	// 	}
	// 	PIN_FUNC_SELECT(pin_name, func);
	// }

	//     gpio_set_level(tx_io_num, 1);
	//     esp_rom_gpio_connect_out_signal(tx_io_num, UART_PERIPH_SIGNAL(uart_num, SOC_UART_TX_PIN_IDX), 0, 0);
	// }

	// if (rts_io_num >= 0 && !uart_try_set_iomux_pin(uart_num, rts_io_num, SOC_UART_RTS_PIN_IDX)) {
	//     gpio_hal_iomux_func_sel(GPIO_PIN_MUX_REG[rts_io_num], PIN_FUNC_GPIO);
	//     gpio_set_direction(rts_io_num, GPIO_MODE_OUTPUT);
	//     esp_rom_gpio_connect_out_signal(rts_io_num, UART_PERIPH_SIGNAL(uart_num, SOC_UART_RTS_PIN_IDX), 0, 0);
	// }

	// if (cts_io_num >= 0  && !uart_try_set_iomux_pin(uart_num, cts_io_num, SOC_UART_CTS_PIN_IDX)) {
	//     gpio_hal_iomux_func_sel(GPIO_PIN_MUX_REG[cts_io_num], PIN_FUNC_GPIO);
	//     gpio_set_pull_mode(cts_io_num, GPIO_PULLUP_ONLY);
	//     gpio_set_direction(cts_io_num, GPIO_MODE_INPUT);
	//     esp_rom_gpio_connect_in_signal(cts_io_num, UART_PERIPH_SIGNAL(uart_num, SOC_UART_CTS_PIN_IDX), 0);
	// }

	return nil
}

func (uart *UART) WriteByte(b byte) error {
	for (uart.Bus.STATUS.Get()&esp.UART_STATUS_TXFIFO_CNT_Msk)>>esp.UART_STATUS_TXFIFO_CNT_Pos >= 128 {
		// Read UART_TXFIFO_CNT from the status register, which indicates how
		// many bytes there are in the transmit buffer. Wait until there are
		// less than 128 bytes in this buffer (the default buffer size).
	}

	// // disable UART_TXFIFO_EMPTY_INT interrupt by clearing UART_TXFIFO_EMPTY_INT_ENA.
	// uart.Bus.INT_ENA.ClearBits(eps.UART_INT_ENA_TXFIFO_EMPTY_INT_ENA)
	// // Write data to be sent to UART_RXFIFO_RD_BYTE
	// // clear UART_TXFIFO_EMPTY_INT interrupt by setting UART_TXFIFO_EMPTY_INT_CLR;
	// uart.Bus.INT_CLR.SetBits(eps.UART_INT_CLR_TXFIFO_EMPTY_INT_ENA)
	// // enable UART_TXFIFO_EMPTY_INT interrupt by setting UART_TXFIFO_EMPTY_INT_ENA
	// uart.Bus.INT_ENA.SetBits(eps.UART_INT_ENA_TXFIFO_EMPTY_INT_ENA)

	uart.Bus.FIFO.Set(uint32(b))
	return nil
}

func (uart *UART) uartEvent() {
	device.Asm("ebreak")
	// println("uart event")
}

func (uart *UART) onDataReady() {
	b := uart.Bus.FIFO.Get()
	uart.Buffer.Put(byte(b))
}
