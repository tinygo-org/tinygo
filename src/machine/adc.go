package machine

// Hardware abstraction layer for the analog-to-digital conversion (ADC)
// peripheral.

// ADCConfig holds ADC configuration parameters. If left unspecified, the zero
// value of each parameter will use the peripheral's default settings.
type ADCConfig struct {
	Reference  uint32 // analog reference voltage (AREF) in millivolts
	Resolution uint32 // number of bits for a single conversion (e.g., 8, 10, 12)
	Samples    uint32 // number of samples for a single conversion (e.g., 4, 8, 16, 32)
	SampleTime uint32 // sample time, in microseconds (µs)
}
