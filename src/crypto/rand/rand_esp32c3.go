// +build esp32c3

// This implementation of crypto/rand uses on-chip random generator
// to generate random numbers.
//

package rand

import (
	"device/esp"
	"device/riscv"
	"machine"
	"unsafe"
)

const MHZ = 1000000

func init() {
	Reader = &esp32c3RndReader{}

	// When using the random number generator, make sure at least either the SAR ADC,
	// high-speed ADC1, or RTC20M_CLK2 is enabled. Otherwise, pseudo-random numbers will be returned.
	//  • SAR ADC can be enabled by using the DIG ADC controller. For details,
	//    please refer to Chapter 6 On-Chip Sensors and Analog Signal Processing [to be added later].
	//  • High-speed ADC is enabled automatically when the Wi-Fi or Bluetooth modules
	//    is enabled.
	//  • RTC20M_CLK is enabled by setting the RTC_CNTL_DIG_CLK20M_EN bit in
	//    the RTC_CNTL_CLK_CONF_REG register.
	// Note:
	//  1. Note that, when the Wi-Fi module is enabled, the value read from the high-speed
	//     ADC can be saturated in some extreme cases, which lowers the entropy. Thus, it
	//     is advisable to also enable the SAR ADC as the noise source for the random
	//     number generator for such cases.
	//  2. Enabling RTC20M_CLK increases the RNG entropy. However, to ensure maximum entropy,
	//     it’s recommended to always enable an ADC source as well.

	// Enable SAR ADC
	esp.SYSTEM.PERIP_CLK_EN0.SetBits(esp.SYSTEM_PERIP_CLK_EN0_APB_SARADC_CLK_EN)

	// High-speed ADC
	esp.SYSTEM.PERIP_CLK_EN0.SetBits(esp.SYSTEM_PERIP_CLK_EN0_ADC2_ARB_CLK_EN)

	// Enable RTC20M_CLK2
	// Unfortunately, the technical reference document from where the above note is taken
	// has no information on RTC_CNTL_DIG_CLK20M_EN, nither the SVD have such information.
	/*

		https://github.com/espressif/esp-idf/blob/5f38b766a83d18f78167d1d0dd8c8427ea1a36cb/components/hal/esp32c3/include/hal/i2c_ll.h#L823

		// rtc_clk needs to switch on.
		if (src_clk == I2C_SCLK_RTC) {
			SET_PERI_REG_MASK(RTC_CNTL_CLK_CONF_REG, RTC_CNTL_DIG_CLK8M_EN_M);
			esp_rom_delay_us(DELAY_RTC_CLK_SWITCH);
		}

		esp.RTC_CLK_CONF


	*/
	// esp.RTC_CNTL.RTC_CLK_CONF.SetBits( ??? )

}

type esp32c3RndReader struct {
	lastCpuTick     uint32
	minimalCPUTicks uint32
}

func getApbFreqHZ() uint32 {
	v := esp.RTC_CNTL.RTC_STORE5.Get() & 0xffff
	v = v << 12
	v += MHZ / 2
	r := v % MHZ
	return v - r
}

const CSR_CPU_COUNTER riscv.CSR = 0x7e2

func getCpuTickCount() uint32 {
	count := CSR_CPU_COUNTER.Get()
	return uint32(count)
}

func (r *esp32c3RndReader) hw_rand() uint32 {
	currentCPUTick := getCpuTickCount()
	result := esp.APB_CTRL.RND_DATA.Get()
	for (currentCPUTick - r.lastCpuTick) < r.minimalCPUTicks {
		currentCPUTick = getCpuTickCount()
		result ^= esp.APB_CTRL.RND_DATA.Get()
	}
	r.lastCpuTick = currentCPUTick
	return result ^ esp.APB_CTRL.RND_DATA.Get()
}

func (r *esp32c3RndReader) Read(b []byte) (n int, err error) {
	if len(b) != 0 {
		// update minimalCPUTicks in case the APB frequency has changed
		r.minimalCPUTicks = 16 * (machine.CPUFrequency() / getApbFreqHZ())
		for i := 0; i < len(b); {
			nextRandom := r.hw_rand()
			byteArray := (*[4]byte)(unsafe.Pointer(&nextRandom))[:]
			for k := 0; k < 4 && i < len(b); {
				b[i] = byteArray[k]
				k++
				i++
			}
		}
	}
	return len(b), nil
}
