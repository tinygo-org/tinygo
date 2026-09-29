//go:build esp32c3

package machine

import (
	"device/esp"
	"runtime/interrupt"
)

// GPIO matrix signals (IDF soc/esp32c3/include/soc/gpio_sig_map.h).
const (
	i2sSignalTxBCK = 13
	i2sSignalTxWS  = 14
	i2sSignalTxSD  = 15
	i2sSignalRxSD  = 15
	i2sSignalRxBCK = 16
	i2sSignalRxWS  = 17
)

const cpuInterruptFromI2S = 11

// The C3 has one interrupt register for both directions of a channel.
const (
	gdmaOutEOF   = esp.DMA_INT_RAW_CH0_OUT_EOF
	gdmaInSucEOF = esp.DMA_INT_RAW_CH0_IN_SUC_EOF
)

var i2sGDMARegs = i2sGDMA{
	outConf0:   &esp.DMA.OUT_CONF0_CH0,
	outLink:    &esp.DMA.OUT_LINK_CH0,
	outPeriSel: &esp.DMA.OUT_PERI_SEL_CH0,
	outIntEna:  &esp.DMA.INT_ENA_CH0,
	outIntClr:  &esp.DMA.INT_CLR_CH0,
	outIntSt:   &esp.DMA.INT_ST_CH0,
	outEOFAddr: &esp.DMA.OUT_EOF_DES_ADDR_CH0,
	inConf0:    &esp.DMA.IN_CONF0_CH0,
	inLink:     &esp.DMA.IN_LINK_CH0,
	inPeriSel:  &esp.DMA.IN_PERI_SEL_CH0,
	inIntEna:   &esp.DMA.INT_ENA_CH0,
	inIntClr:   &esp.DMA.INT_CLR_CH0,
	inIntSt:    &esp.DMA.INT_ST_CH0,
	inEOFAddr:  &esp.DMA.IN_SUC_EOF_DES_ADDR_CH0,
}

// i2sEnableClocks enables and resets I2S and enables GDMA. The one I2S block
// uses the I2S1 bits (IDF hal/esp32c3/include/hal/clk_gate_ll.h line 36).
func i2sEnableClocks() {
	esp.SYSTEM.SetPERIP_CLK_EN0_I2S1_CLK_EN(1)
	esp.SYSTEM.SetPERIP_RST_EN0_I2S1_RST(1)
	esp.SYSTEM.SetPERIP_RST_EN0_I2S1_RST(0)
	if esp.SYSTEM.GetPERIP_CLK_EN1_DMA_CLK_EN() == 0 {
		esp.SYSTEM.SetPERIP_CLK_EN1_DMA_CLK_EN(1)
		esp.SYSTEM.SetPERIP_RST_EN1_DMA_RST(1)
		esp.SYSTEM.SetPERIP_RST_EN1_DMA_RST(0)
	}
	esp.DMA.MISC_CONF.SetBits(esp.DMA_MISC_CONF_CLK_EN)
	esp.I2S0.SetTX_CLKM_CONF_CLK_EN(1)
}

func i2sSetTxClock(div, x, y, z, yn1 uint32) {
	i2sSetClockDiv(&esp.I2S0.TX_CLKM_CONF, &esp.I2S0.TX_CLKM_DIV_CONF, 0, div, x, y, z, yn1)
	esp.I2S0.SetTX_CLKM_CONF_TX_CLK_SEL(i2sPLL160MClockSel)
	esp.I2S0.SetTX_CLKM_CONF_TX_CLK_ACTIVE(1)
	esp.I2S0.SetRX_CLKM_CONF_MCLK_SEL(0)
}

func i2sSetRxClock(div, x, y, z, yn1 uint32) {
	i2sSetClockDiv(&esp.I2S0.RX_CLKM_CONF, &esp.I2S0.RX_CLKM_DIV_CONF, 0, div, x, y, z, yn1)
	esp.I2S0.SetRX_CLKM_CONF_RX_CLK_SEL(i2sPLL160MClockSel)
	esp.I2S0.SetRX_CLKM_CONF_RX_CLK_ACTIVE(1)
}

func i2sEnableInterrupt() {
	esp.INTERRUPT_CORE0.SetDMA_CH0_INT_MAP(cpuInterruptFromI2S)
	interrupt.New(cpuInterruptFromI2S, func(interrupt.Interrupt) {
		I2S0.handleInterrupt()
	}).Enable()
}
