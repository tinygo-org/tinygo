// +build esp32c3

// This implementation of crypto/rand uses on-chip random generator
// to generate random numbers.
//

package rand

import (
	"device/esp"
	"unsafe"
)

func init() {
	Reader = &reader{}

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

	// Enable RTC20M_CLK2
	// Unfortunately, the technical reference document from where the above note is taken
	// has no information on RTC_CNTL_DIG_CLK20M_EN, nither the SVD have such information.
	// esp.RTC_CNTL.RTC_CLK_CONF.SetBits( ??? )

}

type reader struct {
}

func (r *reader) Read(b []byte) (n int, err error) {
	if len(b) != 0 {
		for i := 0; i < len(b); {
			r := esp.APB_CTRL.RND_DATA.Get()
			a := (*[4]byte)(unsafe.Pointer(&r))[:]
			for k := 0; k < 4 && i < len(b); {
				b[i] = a[k]
				k++
				i++
			}
		}
	}
	return len(b), nil
}
