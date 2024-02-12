//go:build nrf52840 && ((nrf52840_generic && !nrf52840_lfxtal_false) || nrf52840_lfxtal_true)

package machine

const HasLowFrequencyCrystal = true
