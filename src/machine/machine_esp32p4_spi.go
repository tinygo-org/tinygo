//go:build esp32p4

package machine

import (
	"device/esp"
	"errors"
	"runtime/volatile"
	"unsafe"
)

const spiClockFreq = 40_000_000 // XTAL_CLK, GPSPI2_CLK_SRC_SEL 0. See ESP-IDF hal/esp32p4/include/hal/spi_ll.h spi_ll_set_clk_source.

// GPIO matrix signals for GPSPI2. See ESP-IDF
// components/soc/esp32p4/include/soc/gpio_sig_map.h.
const (
	SPI2_CK_PAD_OUT_IDX = 53
	SPI2_Q_PAD_IN_IDX   = 54
	SPI2_D_PAD_OUT_IDX  = 55
	SPI2_CS_PAD_OUT_IDX = 62
)

var (
	ErrInvalidSPIBus  = errors.New("machine: SPI bus is invalid")
	ErrInvalidSPIMode = errors.New("machine: SPI mode is invalid")
)

// SPI is a general purpose SPI controller on the ESP32-P4.
type SPI struct {
	Bus *esp.SPI2_Type
}

var (
	// SPI0 and SPI1 are used for flash and PSRAM.
	SPI2 = &SPI{esp.SPI2}
)

// Configure and make the SPI peripheral ready to use.
func (spi *SPI) Configure(config SPIConfig) error {
	if spi.Bus != esp.SPI2 {
		return ErrInvalidSPIBus
	}

	// Clock and reset sequence from ESP-IDF esp_hal_gpspi/esp32p4/include/hal/spi_ll.h.
	esp.HP_SYS_CLKRST.SetSOC_CLK_CTRL1_REG_GPSPI2_SYS_CLK_EN(1)
	esp.HP_SYS_CLKRST.SetSOC_CLK_CTRL2_REG_GPSPI2_APB_CLK_EN(1)
	esp.HP_SYS_CLKRST.SetHP_RST_EN2_REG_RST_EN_SPI2(1)
	esp.HP_SYS_CLKRST.SetHP_RST_EN2_REG_RST_EN_SPI2(0)
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL116_REG_GPSPI2_CLK_SRC_SEL(0) // XTAL
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL116_REG_GPSPI2_HS_CLK_DIV_NUM(0)
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL116_REG_GPSPI2_MST_CLK_DIV_NUM(0)
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL116_REG_GPSPI2_HS_CLK_EN(1)
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL116_REG_GPSPI2_MST_CLK_EN(1)

	spi.Bus.SPI_SLAVE.Set(0)
	spi.Bus.SPI_MISC.Set(0)
	spi.Bus.SPI_USER.Set(0)
	spi.Bus.SPI_USER1.Set(0)
	spi.Bus.SetSPI_USER2_SPI_MST_REMPTY_ERR_END_EN(0)
	spi.Bus.SPI_CTRL.Set(0)
	spi.Bus.SPI_CLK_GATE.Set(0)
	spi.Bus.SPI_DMA_CONF.Set(0)
	spi.Bus.SetSPI_DMA_CONF_SPI_RX_AFIFO_RST(1)
	spi.Bus.SetSPI_DMA_CONF_SPI_BUF_AFIFO_RST(1)
	spi.Bus.SPI_CLOCK.Set(0)

	spi.Bus.SetSPI_CLK_GATE_SPI_CLK_EN(1)
	spi.Bus.SetSPI_CLK_GATE_SPI_MST_CLK_SEL(1)
	spi.Bus.SetSPI_CLK_GATE_SPI_MST_CLK_ACTIVE(1)
	spi.Bus.SetSPI_DMA_CONF_SPI_SLV_TX_SEG_TRANS_CLR_EN(1)
	spi.Bus.SetSPI_DMA_CONF_SPI_SLV_RX_SEG_TRANS_CLR_EN(1)
	spi.Bus.SetSPI_USER_SPI_USR_MOSI(1)
	spi.Bus.SetSPI_USER_SPI_USR_MISO(1)
	spi.Bus.SetSPI_USER_SPI_DOUTDIN(1)

	switch config.Mode {
	case Mode0:
		spi.Bus.SetSPI_MISC_SPI_CK_IDLE_EDGE(0)
		spi.Bus.SetSPI_USER_SPI_CK_OUT_EDGE(0)
	case Mode1:
		spi.Bus.SetSPI_MISC_SPI_CK_IDLE_EDGE(0)
		spi.Bus.SetSPI_USER_SPI_CK_OUT_EDGE(1)
	case Mode2:
		spi.Bus.SetSPI_MISC_SPI_CK_IDLE_EDGE(1)
		spi.Bus.SetSPI_USER_SPI_CK_OUT_EDGE(1)
	case Mode3:
		spi.Bus.SetSPI_MISC_SPI_CK_IDLE_EDGE(1)
		spi.Bus.SetSPI_USER_SPI_CK_OUT_EDGE(0)
	default:
		return ErrInvalidSPIMode
	}

	if config.LSBFirst {
		spi.Bus.SetSPI_CTRL_SPI_WR_BIT_ORDER(1)
		spi.Bus.SetSPI_CTRL_SPI_RD_BIT_ORDER(1)
	}

	spi.Bus.SPI_CLOCK.Set(freqToClockDiv(config.Frequency))

	// A zero CS is GPIO0, so treat it as unset.
	if config.CS == 0 {
		config.CS = NoPin
	}

	if config.SDI != NoPin {
		config.SDI.configure(PinConfig{Mode: PinInput}, SPI2_Q_PAD_IN_IDX)
	}
	config.SDO.configure(PinConfig{Mode: PinOutput}, SPI2_D_PAD_OUT_IDX)
	config.SCK.configure(PinConfig{Mode: PinOutput}, SPI2_CK_PAD_OUT_IDX)
	config.CS.configure(PinConfig{Mode: PinOutput}, SPI2_CS_PAD_OUT_IDX)

	return nil
}

// Transfer writes and reads a single byte.
func (spi *SPI) Transfer(w byte) (byte, error) {
	spi.Bus.SetSPI_MS_DLEN_SPI_MS_DATA_BITLEN(7)
	spi.Bus.SPI_W0.Set(uint32(w))
	spi.start()
	return byte(spi.Bus.SPI_W0.Get()), nil
}

// Tx writes w and reads into r. The shorter slice is padded with zeros or
// its extra input is discarded.
func (spi *SPI) Tx(w, r []byte) error {
	toTransfer := max(len(w), len(r))
	buf := (*[16]volatile.Register32)(unsafe.Pointer(&spi.Bus.SPI_W0))
	for toTransfer > 0 {
		chunkSize := min(toTransfer, 64)
		spiTxFillBuffer(buf, w)
		spi.Bus.SetSPI_MS_DLEN_SPI_MS_DATA_BITLEN(uint32(chunkSize)*8 - 1)
		spi.start()

		rxSize := min(chunkSize, len(r))
		for i := 0; i < rxSize; i++ {
			r[i] = byte(buf[i/4].Get() >> ((i % 4) * 8))
		}

		w = w[min(chunkSize, len(w)):]
		r = r[rxSize:]
		toTransfer -= chunkSize
	}
	return nil
}

func (spi *SPI) start() {
	spi.Bus.SetSPI_CMD_SPI_UPDATE(1)
	for spi.Bus.GetSPI_CMD_SPI_UPDATE() != 0 {
	}
	spi.Bus.SetSPI_CMD_SPI_USR(1)
	for spi.Bus.GetSPI_CMD_SPI_USR() != 0 {
	}
}
