//go:build esp32p4

// Per-controller I2S hardware mapping for ESP32-P4.
//
// I2S0 and I2S1 have different GPIO-matrix signal IDs, clock-control
// registers, and AHB-DMA peripheral IDs.  The CrowPanel routes audio to I2S1;
// the WaveShare ESP32-P4-NANO routes the ES8311 to I2S0.

package machine

import "device/esp"

// I2SController selects the ESP32-P4 I2S controller instance.
type I2SController uint8

const (
	// I2SControllerDefault keeps the existing I2S1 behavior.
	I2SControllerDefault I2SController = iota
	// I2SController0 selects I2S0 (WaveShare ESP32-P4-NANO / ES8311).
	I2SController0
	// I2SController1 selects I2S1 (CrowPanel ESP32-P4 10.1").
	I2SController1
)

// i2sHw holds the controller-specific register and signal mapping.
type i2sHw struct {
	ctrl I2SController
	name string
	dev  *esp.I2S_Type

	bckOutIdx   uint32
	wsOutIdx    uint32
	sdOutIdx    uint32
	mclkOutIdx  uint32
	sdInIdx     uint32
	rxBckOutIdx uint32
	rxWsOutIdx  uint32

	dmaPeriph uint8
}

func i2sHardware(ctrl I2SController) i2sHw {
	if ctrl == I2SController0 {
		return i2sHw{
			ctrl:        I2SController0,
			name:        "I2S0",
			dev:         esp.I2S0,
			bckOutIdx:   25, // I2S0_O_BCK_PAD_OUT_IDX
			wsOutIdx:    27, // I2S0_O_WS_PAD_OUT_IDX
			sdOutIdx:    28, // I2S0_O_SD_PAD_OUT_IDX
			mclkOutIdx:  26, // I2S0_MCLK_PAD_OUT_IDX
			sdInIdx:     28, // I2S0_I_SD_PAD_IN_IDX
			rxBckOutIdx: 29, // I2S0_I_BCK_PAD_OUT_IDX
			rxWsOutIdx:  30, // I2S0_I_WS_PAD_OUT_IDX
			dmaPeriph:   3,  // SOC_GDMA_TRIG_PERIPH_I2S0
		}
	}
	return i2sHw{
		ctrl:        I2SController1,
		name:        "I2S1",
		dev:         esp.I2S1,
		bckOutIdx:   31, // I2S1_O_BCK_PAD_OUT_IDX
		wsOutIdx:    33, // I2S1_O_WS_PAD_OUT_IDX
		sdOutIdx:    34, // I2S1_O_SD_PAD_OUT_IDX
		mclkOutIdx:  32, // I2S1_MCLK_PAD_OUT_IDX
		sdInIdx:     34, // I2S1_I_SD_PAD_IN_IDX
		rxBckOutIdx: 35, // I2S1_I_BCK_PAD_OUT_IDX
		rxWsOutIdx:  36, // I2S1_I_WS_PAD_OUT_IDX
		dmaPeriph:   4,  // SOC_GDMA_TRIG_PERIPH_I2S1
	}
}

func i2sBool(v bool) uint32 {
	if v {
		return 1
	}
	return 0
}

func (h i2sHw) enableAPB(enable bool) {
	if h.ctrl == I2SController0 {
		esp.HP_SYS_CLKRST.SetSOC_CLK_CTRL2_REG_I2S0_APB_CLK_EN(i2sBool(enable))
	} else {
		esp.HP_SYS_CLKRST.SetSOC_CLK_CTRL2_REG_I2S1_APB_CLK_EN(i2sBool(enable))
	}
}

func (h i2sHw) resetAPB() {
	if h.ctrl == I2SController0 {
		esp.HP_SYS_CLKRST.SetHP_RST_EN2_REG_RST_EN_I2S0_APB(1)
		esp.HP_SYS_CLKRST.SetHP_RST_EN2_REG_RST_EN_I2S0_APB(0)
	} else {
		esp.HP_SYS_CLKRST.SetHP_RST_EN2_REG_RST_EN_I2S1_APB(1)
		esp.HP_SYS_CLKRST.SetHP_RST_EN2_REG_RST_EN_I2S1_APB(0)
	}
}

func (h i2sHw) enableMCLKPad(enable bool) {
	if h.ctrl == I2SController0 {
		esp.LP_AON_CLKRST.SetLP_AONCLKRST_HP_CLK_CTRL_LP_AONCLKRST_HP_PAD_I2S0_MCLK_EN(i2sBool(enable))
	} else {
		esp.LP_AON_CLKRST.SetLP_AONCLKRST_HP_CLK_CTRL_LP_AONCLKRST_HP_PAD_I2S1_MCLK_EN(i2sBool(enable))
	}
}

func (h i2sHw) setMSTClkTX() {
	if h.ctrl == I2SController0 {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL14_REG_I2S0_MST_CLK_SEL(1)
	} else {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL17_REG_I2S1_MST_CLK_SEL(1)
	}
}

func (h i2sHw) setMSTClkRX() {
	if h.ctrl == I2SController0 {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL14_REG_I2S0_MST_CLK_SEL(0)
	} else {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL17_REG_I2S1_MST_CLK_SEL(0)
	}
}

func (h i2sHw) setTXClockSrcXTAL() {
	if h.ctrl == I2SController0 {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL13_REG_I2S0_TX_CLK_SRC_SEL(0)
	} else {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL15_REG_I2S1_TX_CLK_SRC_SEL(0)
	}
}

func (h i2sHw) setTXClockEnable(enable bool) {
	if h.ctrl == I2SController0 {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL13_REG_I2S0_TX_CLK_EN(i2sBool(enable))
	} else {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL15_REG_I2S1_TX_CLK_EN(i2sBool(enable))
	}
}

func (h i2sHw) setTXDivN(v uint32) {
	if h.ctrl == I2SController0 {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL13_REG_I2S0_TX_DIV_N(v)
	} else {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL16_REG_I2S1_TX_DIV_N(v)
	}
}

func (h i2sHw) setTXDivX(v uint32) {
	if h.ctrl == I2SController0 {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL13_REG_I2S0_TX_DIV_X(v)
	} else {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL16_REG_I2S1_TX_DIV_X(v)
	}
}

func (h i2sHw) setTXDivY(v uint32) {
	if h.ctrl == I2SController0 {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL14_REG_I2S0_TX_DIV_Y(v)
	} else {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL16_REG_I2S1_TX_DIV_Y(v)
	}
}

func (h i2sHw) setTXDivZ(v uint32) {
	if h.ctrl == I2SController0 {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL14_REG_I2S0_TX_DIV_Z(v)
	} else {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL17_REG_I2S1_TX_DIV_Z(v)
	}
}

func (h i2sHw) setTXDivYN1(v uint32) {
	if h.ctrl == I2SController0 {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL14_REG_I2S0_TX_DIV_YN1(v)
	} else {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL17_REG_I2S1_TX_DIV_YN1(v)
	}
}

func (h i2sHw) setRXClockSrcXTAL() {
	if h.ctrl == I2SController0 {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL11_REG_I2S0_RX_CLK_SRC_SEL(0)
	} else {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL14_REG_I2S1_RX_CLK_SRC_SEL(0)
	}
}

func (h i2sHw) setRXClockEnable(enable bool) {
	if h.ctrl == I2SController0 {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL11_REG_I2S0_RX_CLK_EN(i2sBool(enable))
	} else {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL14_REG_I2S1_RX_CLK_EN(i2sBool(enable))
	}
}

func (h i2sHw) setRXDivN(v uint32) {
	if h.ctrl == I2SController0 {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL12_REG_I2S0_RX_DIV_N(v)
	} else {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL14_REG_I2S1_RX_DIV_N(v)
	}
}

func (h i2sHw) setRXDivX(v uint32) {
	if h.ctrl == I2SController0 {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL12_REG_I2S0_RX_DIV_X(v)
	} else {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL15_REG_I2S1_RX_DIV_X(v)
	}
}

func (h i2sHw) setRXDivY(v uint32) {
	if h.ctrl == I2SController0 {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL12_REG_I2S0_RX_DIV_Y(v)
	} else {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL15_REG_I2S1_RX_DIV_Y(v)
	}
}

func (h i2sHw) setRXDivZ(v uint32) {
	if h.ctrl == I2SController0 {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL13_REG_I2S0_RX_DIV_Z(v)
	} else {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL15_REG_I2S1_RX_DIV_Z(v)
	}
}

func (h i2sHw) setRXDivYN1(v uint32) {
	if h.ctrl == I2SController0 {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL13_REG_I2S0_RX_DIV_YN1(v)
	} else {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL15_REG_I2S1_RX_DIV_YN1(v)
	}
}

// clockRegs returns the two/three HP_SYS_CLKRST registers used for logging.
func (h i2sHw) clockRegs() (a, b, c uint32) {
	if h.ctrl == I2SController0 {
		return esp.HP_SYS_CLKRST.PERI_CLK_CTRL13.Get(),
			esp.HP_SYS_CLKRST.PERI_CLK_CTRL14.Get(), 0
	}
	return esp.HP_SYS_CLKRST.PERI_CLK_CTRL15.Get(),
		esp.HP_SYS_CLKRST.PERI_CLK_CTRL16.Get(),
		esp.HP_SYS_CLKRST.PERI_CLK_CTRL17.Get()
}
