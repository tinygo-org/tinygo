//go:build esp32c6

package machine

import (
	"device/esp"
	"runtime/interrupt"
)

// GPIO matrix signals (IDF soc/esp32c6/include/soc/gpio_sig_map.h).
const (
	i2sSignalTxBCK = 13
	i2sSignalTxWS  = 14
	i2sSignalTxSD  = 15
	i2sSignalRxSD  = 15
	i2sSignalRxBCK = 16
	i2sSignalRxWS  = 17
)

const cpuInterruptFromI2S = 11

// Interrupt bits from IDF soc/esp32c6/include/soc/gdma_struct.h.
const (
	gdmaOutEOF   = 1 << 1
	gdmaInSucEOF = 1 << 1
)

var i2sGDMARegs = i2sGDMA{
	outConf0:   &esp.DMA.CH[0].OUT_CONF0,
	outLink:    &esp.DMA.CH[0].OUT_LINK,
	outPeriSel: &esp.DMA.CH[0].OUT_PERI_SEL,
	outIntEna:  &esp.DMA.OUT_INT_CH[0].ENA,
	outIntClr:  &esp.DMA.OUT_INT_CH[0].CLR,
	outIntSt:   &esp.DMA.OUT_INT_CH[0].ST,
	outEOFAddr: &esp.DMA.CH[0].OUT_EOF_DES_ADDR,
	inConf0:    &esp.DMA.CH[0].IN_CONF0,
	inLink:     &esp.DMA.CH[0].IN_LINK,
	inPeriSel:  &esp.DMA.CH[0].IN_PERI_SEL,
	inIntEna:   &esp.DMA.IN_INT_CH[0].ENA,
	inIntClr:   &esp.DMA.IN_INT_CH[0].CLR,
	inIntSt:    &esp.DMA.IN_INT_CH[0].ST,
	inEOFAddr:  &esp.DMA.CH[0].IN_SUC_EOF_DES_ADDR,
}

func i2sEnableClocks() {
	esp.PCR.SetI2S_CONF_I2S_CLK_EN(1)
	esp.PCR.SetI2S_CONF_I2S_RST_EN(1)
	esp.PCR.SetI2S_CONF_I2S_RST_EN(0)
	if esp.PCR.GetGDMA_CONF_GDMA_CLK_EN() == 0 {
		esp.PCR.SetGDMA_CONF_GDMA_CLK_EN(1)
		esp.PCR.SetGDMA_CONF_GDMA_RST_EN(1)
		esp.PCR.SetGDMA_CONF_GDMA_RST_EN(0)
	}
	esp.DMA.MISC_CONF.SetBits(esp.DMA_MISC_CONF_CLK_EN)
}

// The C6 I2S clocks are in PCR (IDF hal/esp32c6/include/hal/i2s_ll.h).
func i2sSetTxClock(div, x, y, z, yn1 uint32) {
	i2sSetClockDiv(&esp.PCR.I2S_TX_CLKM_CONF, &esp.PCR.I2S_TX_CLKM_DIV_CONF,
		esp.PCR_I2S_TX_CLKM_CONF_I2S_TX_CLKM_DIV_NUM_Pos, div, x, y, z, yn1)
	esp.PCR.SetI2S_TX_CLKM_CONF_I2S_TX_CLKM_SEL(i2sPLL160MClockSel)
	esp.PCR.SetI2S_TX_CLKM_CONF_I2S_TX_CLKM_EN(1)
	esp.PCR.SetI2S_RX_CLKM_CONF_I2S_MCLK_SEL(0)
}

func i2sSetRxClock(div, x, y, z, yn1 uint32) {
	i2sSetClockDiv(&esp.PCR.I2S_RX_CLKM_CONF, &esp.PCR.I2S_RX_CLKM_DIV_CONF,
		esp.PCR_I2S_RX_CLKM_CONF_I2S_RX_CLKM_DIV_NUM_Pos, div, x, y, z, yn1)
	esp.PCR.SetI2S_RX_CLKM_CONF_I2S_RX_CLKM_SEL(i2sPLL160MClockSel)
	esp.PCR.SetI2S_RX_CLKM_CONF_I2S_RX_CLKM_EN(1)
}

func i2sEnableInterrupt() {
	interruptMap(esp.IRQ_DMA_OUT_CH0).Set(cpuInterruptFromI2S)
	interruptMap(esp.IRQ_DMA_IN_CH0).Set(cpuInterruptFromI2S)
	interrupt.New(cpuInterruptFromI2S, func(interrupt.Interrupt) {
		I2S0.handleInterrupt()
	}).Enable()
}
