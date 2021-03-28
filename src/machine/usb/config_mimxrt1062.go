// +build mimxrt1062

package usb

const configCPUFrequencyHz = 600000000

// The following constants must be defined for USB 2.0 host/device support.
//
// However, some or all of these constants may be unused in the core driver
// code, depending on which components are allocated by the USB driver.
//
// These values are __PLATFORM-SPECIFIC__, and they serve two primary roles:
//   1. Determine which components to allocate from the USB driver.
//   2. Configure driver attributes NOT defined by the USB 2.0 specification.
//
const (
	// configDeviceCount defines the number of USB device-mode ports available.
	configDeviceCount = configDeviceCDCACMCount

	// configHostCount defines the number of USB host-mode ports available.
	configHostCount = 0

	// configInterruptQueueSize defines the number of interrupts to retain in the
	// queue of runtime processes.
	configInterruptQueueSize = 8

	// configDeviceBufferSize defines the size of the send and receive data
	// endpoint buffers.
	configDeviceBufferSize = 512

	// configDeviceMaxEndpoints defines the maximum number of endpoints supported.
	configDeviceMaxEndpoints = 4

	// configDeviceControllerMaxPacketSize defines the maximum packet size for
	// communication with an endpoint. The maximum per USB 2.0 spec is 64 bytes,
	// although a platform may restrict this to something lower if needed.
	configDeviceControllerMaxPacketSize = 64

	// configDeviceMaxPower defines the maximum power consumption of the USB
	// device when fully-operational, expressed in 2 mA units.
	configDeviceMaxPower = 50 // 100 mA

	// configDeviceSelfPowered defines whether the device is self-powered (1) or
	// not (0).
	configDeviceSelfPowered = 1

	// configDeviceRemoteWakeup defines whether the device supports remote-wakeup
	// (1) or not (0).
	configDeviceRemoteWakeup = 0 // (NOT YET SUPPORTED)

	// configDeviceControllerMaxDTD defines the maximum number of DTD supported.
	configDeviceControllerMaxDTD = 16

	// configDeviceControllerQHAlign defines the memory alignment of QH buffer.
	// Ensure this agrees with the go:align pragma on deviceControllerQHBuffer.
	configDeviceControllerQHAlign = 2048

	// configDeviceControllerDTDAlign defines the memory alignment of DTD buffer.
	// Ensure this agrees with the go:align pragma on deviceControllerDTDBuffer.
	configDeviceControllerDTDAlign = 32
)

// If USB CDC-ACM device support is required, the following constants must be
// defined.
const (
	// configDeviceCDCACMCount defines the number of USB CDC-ACM interfaces
	// to initialize; must be less-than or equal to configDeviceCount.
	configDeviceCDCACMCount = 1

	// configDeviceCDCACMConfigurationCount defines the number of configurations
	// required for each CDC-ACM device port.
	configDeviceCDCACMConfigurationCount = 1

	// configDeviceCDCACMConfigurationIndex defines the default configuration
	// index used to initialize CDC-ACM device class configurations. This is a
	// 1-based index into the list of CDC-ACM configurations (which are 0-based
	// slices, and thus this index is offset by -1). A configuration index of 0
	// represents the invalid configuration index.
	configDeviceCDCACMConfigurationIndex = 1

	// configDeviceCDCACMInterfaceCount defines the number of interfaces required
	// for each CDC-ACM configuration.
	configDeviceCDCACMInterfaceCount = 2

	// configDeviceCDCACMInterruptInPacketSize defines the packet size of CDC-ACM
	// communication interface's interrupt input endpoint.
	configDeviceCDCACMInterruptInPacketSize = configDeviceCDCACMFSInterruptInPacketSize // (full-speed)

	// configDeviceCDCACMBulkInPacketSize defines the packet size of CDC-ACM data
	// interface's bulk input endpoint.
	configDeviceCDCACMBulkInPacketSize = configDeviceCDCACMFSBulkInPacketSize // (full-speed)

	// configDeviceCDCACMBulkOutPacketSize defines the packet size of CDC-ACM data
	// interface's bulk output endpoint.
	configDeviceCDCACMBulkOutPacketSize = configDeviceCDCACMFSBulkOutPacketSize // (full-speed)

	// configDeviceCDCACMInterruptInInterval defines the interval of CDC-ACM
	// communication interface's interrupt input endpoint.
	configDeviceCDCACMInterruptInInterval = configDeviceCDCACMFSInterruptInInterval // (full-speed)
)

// If USB CDC-ACM device support is required, the following arrays must be
// initialized with each device's CDC-ACM configuration.
var (
	// configDeviceCDCACM defines the configuration specific to CDC-ACM devices
	// including the properties of endpoints required by the device class driver,
	// as well as the serial UART line coding properties.
	configDeviceCDCACM = [configDeviceCDCACMCount][configDeviceCDCACMConfigurationCount]deviceCDCACMConfig{
		{{ // USB CDC-ACM [0]
			interfaceSpeed: specSpeedFull, // USB full-speed (12 Mbit/s)
			// Serial line configuration
			lineCodingSize:       7,      // Size of line-coding message
			lineCodingBaudRate:   115200, // Data terminal rate
			lineCodingCharFormat: 0,      // Character format
			lineCodingParityType: 0,      // Parity type
			lineCodingDataBits:   8,      // Data word size
			// Communication/control interface
			commInterfaceIndex:        descriptorInterfaceCDCACMComm,           // communication/control interface index
			commInterruptInEndpoint:   descriptorEndpointCDCACMCommInterruptIn, // interrupt input endpoint index (address)
			commInterruptInPacketSize: configDeviceCDCACMInterruptInPacketSize,
			commInterruptInInterval:   configDeviceCDCACMInterruptInInterval,
			// Data interface
			dataInterfaceIndex:    descriptorInterfaceCDCACMData,      // data interface index
			dataBulkInEndpoint:    descriptorEndpointCDCACMDataBulkIn, // bulk input endpoint index (address)
			dataBulkInPacketSize:  configDeviceCDCACMBulkInPacketSize,
			dataBulkOutEndpoint:   descriptorEndpointCDCACMDataBulkOut, // bulk output endpoint index (address)
			dataBulkOutPacketSize: configDeviceCDCACMBulkOutPacketSize,
		}},
	}

	// configDeviceCDCACMDescriptor defines the generic USB 2.0 descriptors used
	// to describe each CDC-ACM device enabled.
	//
	// Note that the actual descriptor byte slices need not (should not) be
	// defined here. Instead, a reusable global slice should be declared as a
	// standalone type and not a member of a composite type; this allows both
	// reusability and better control over the memory layout/alignment for USB
	// PHYs that may require constraints of this sort. The fields of struct
	// deviceDescriptor are therefore all pointers to byte slices so that they
	// can be initialized with references to the aforementioned global arrays.
	configDeviceCDCACMDescriptor = [configDeviceCDCACMCount]deviceDescriptor{
		{ // USB CDC-ACM [0]
			pDevice: &descriptorDeviceCDCACM,
			pConfig: &descriptorConfigurationCDCACM,
			language: []deviceDescriptorLanguage{
				{
					pString: []deviceDescriptorString{
						&descriptorStringCDCACMLanguage,
						&configDescriptorStringCDCACMManufacturer,
						&configDescriptorStringCDCACMProduct,
					},
					ident: 0x0409,
				},
			},
		},
	}

	// default device descriptor info (little-endian byte order).
	configDescriptorDeviceCDCACMVID = []uint8{0xC9, 0x1F} // USB Vendor ID (0x1FC9)
	configDescriptorDeviceCDCACMPID = []uint8{0x94, 0x00} // USB Product ID (0x0094)
	configDescriptorDeviceCDCACMBCD = []uint8{0x01, 0x01} // USB Device Version (0x0101)

	configDescriptorStringCDCACMManufacturer = []uint8{
		2 + 2*18, specDescriptorTypeString,
		'N', 0,
		'X', 0,
		'P', 0,
		' ', 0,
		'S', 0,
		'e', 0,
		'm', 0,
		'i', 0,
		'c', 0,
		'o', 0,
		'n', 0,
		'd', 0,
		'u', 0,
		'c', 0,
		't', 0,
		'o', 0,
		'r', 0,
		's', 0,
	}

	configDescriptorStringCDCACMProduct = []uint8{
		2 + 2*20, specDescriptorTypeString,
		'T', 0,
		'i', 0,
		'n', 0,
		'y', 0,
		'G', 0,
		'o', 0,
		' ', 0,
		'U', 0,
		'S', 0,
		'B', 0,
		' ', 0,
		'(', 0,
		'C', 0,
		'D', 0,
		'C', 0,
		'-', 0,
		'A', 0,
		'C', 0,
		'M', 0,
		')', 0,
	}
)

// The following additional constants are not required by the USB driver but are
// used by the usb package on this platform.
const (

	// configInterruptPriority defines the priority number for USB interrupts.
	configInterruptPriority = 3

	// configDeviceControllerMaxPrimeAttempts defines the maximum number of
	// attempts to prime an endpoint for transfer. If attempts exceeds this
	// value, then the endpoint status has been reset.
	configDeviceControllerMaxPrimeAttempts = 10000000

	// USB CDC-ACM high-speed (480 Mbit/s) packet size
	configDeviceCDCACMHSInterruptInPacketSize = 16
	configDeviceCDCACMHSInterruptInInterval   = 7 // 2^(7-1)/8 = 8ms
	configDeviceCDCACMHSBulkInPacketSize      = 512
	configDeviceCDCACMHSBulkOutPacketSize     = 512

	// USB CDC-ACM full-speed (12 Mbit/s) packet size
	configDeviceCDCACMFSInterruptInPacketSize = 16
	configDeviceCDCACMFSInterruptInInterval   = 8 // 2^(8-1)/8 = 16ms
	configDeviceCDCACMFSBulkInPacketSize      = 64
	configDeviceCDCACMFSBulkOutPacketSize     = 64
)
