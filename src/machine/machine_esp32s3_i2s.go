//go:build esp32s3

package machine

import (
	"device/esp"
	"runtime/interrupt"
)

// GPIO matrix signals for I2S0 (IDF soc/esp32s3/include/soc/gpio_sig_map.h).
const (
	i2sSignalTxBCK = 22
	i2sSignalTxWS  = 24
	i2sSignalTxSD  = 25
	i2sSignalRxSD  = 25
	i2sSignalRxBCK = 26
	i2sSignalRxWS  = 27
)

// Line 18 is a free level 1 interrupt. espradio uses 12, 13 and 17
// (ESP32-S3 TRM, table "CPU Interrupts").
const cpuInterruptFromI2S = 18

// Interrupt bits from IDF soc/esp32s3/include/soc/gdma_struct.h.
const (
	gdmaOutEOF   = 1 << 1
	gdmaInSucEOF = 1 << 1
)

var i2sGDMARegs = i2sGDMA{
	outConf0:   &esp.DMA.CH[0].OUT_CONF0,
	outLink:    &esp.DMA.CH[0].OUT_LINK,
	outPeriSel: &esp.DMA.CH[0].OUT_PERI_SEL,
	outIntEna:  &esp.DMA.CH[0].OUT_INT_ENA,
	outIntClr:  &esp.DMA.CH[0].OUT_INT_CLR,
	outIntSt:   &esp.DMA.CH[0].OUT_INT_ST,
	outEOFAddr: &esp.DMA.CH[0].OUT_EOF_DES_ADDR,
	inConf0:    &esp.DMA.CH[0].IN_CONF0,
	inLink:     &esp.DMA.CH[0].IN_LINK,
	inPeriSel:  &esp.DMA.CH[0].IN_PERI_SEL,
	inIntEna:   &esp.DMA.CH[0].IN_INT_ENA,
	inIntClr:   &esp.DMA.CH[0].IN_INT_CLR,
	inIntSt:    &esp.DMA.CH[0].IN_INT_ST,
	inEOFAddr:  &esp.DMA.CH[0].IN_SUC_EOF_DES_ADDR,
}

func i2sEnableClocks() {
	esp.SYSTEM.SetPERIP_CLK_EN0_I2S0_CLK_EN(1)
	esp.SYSTEM.SetPERIP_RST_EN0_I2S0_RST(1)
	esp.SYSTEM.SetPERIP_RST_EN0_I2S0_RST(0)
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
	interruptMap(esp.IRQ_DMA_OUT_CH0).Set(cpuInterruptFromI2S)
	interruptMap(esp.IRQ_DMA_IN_CH0).Set(cpuInterruptFromI2S)
	interrupt.New(cpuInterruptFromI2S, func(interrupt.Interrupt) {
		I2S0.handleInterrupt()
	}).Enable()
}
