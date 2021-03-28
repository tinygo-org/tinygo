package usb

// USB specification version number (BCD)
var descriptorUSBSpecification = []uint8{0x00, 0x02} // USB 2.0

// USB CDC-ACM descriptor constants
const (
	// interface indices
	descriptorInterfaceCDCACMComm = 0 // communication/control interface
	descriptorInterfaceCDCACMData = 1 // data interface

	descriptorEndpointCDCACMCommInterruptIn = 1 // interrupt endpoint (input)
	descriptorEndpointCDCACMDataBulkIn      = 2 // data endpoint (input)
	descriptorEndpointCDCACMDataBulkOut     = 3 // data endpoint (output)

	descriptorConfigurationCDCACMHeaderFuncSize = 5
	descriptorConfigurationCDCACMCallManageSize = 5
	descriptorConfigurationCDCACMAbstractSize   = 4
	descriptorConfigurationCDCACMUnionFuncSize  = 5

	descriptorConfigurationCDCACMSize = uint16(
		specDescriptorLengthConfigure + // configuration
			specDescriptorLengthInterface + // communication/control interface
			descriptorConfigurationCDCACMHeaderFuncSize + // CDC header
			descriptorConfigurationCDCACMCallManageSize + // CDC call management
			descriptorConfigurationCDCACMAbstractSize + // CDC abstract control
			descriptorConfigurationCDCACMUnionFuncSize + // CDC union
			specDescriptorLengthEndpoint + // communication/control input endpoint
			specDescriptorLengthInterface + // data interface
			specDescriptorLengthEndpoint + // data input endpoint
			specDescriptorLengthEndpoint) // data output endpoint

	descriptorConfigurationCDCACMAttributes = (specDescriptorConfigureAttributeD7Msk) | // Bit 7: reserved (1)
		(configDeviceSelfPowered << specDescriptorConfigureAttributeSelfPoweredPos) | // Bit 6: self-powered
		(configDeviceRemoteWakeup << specDescriptorConfigureAttributeRemoteWakeupPos) | // Bit 5: remote wakeup
		0 // Bits 0-4: reserved (0)
)

// USB 2.0 CDC-ACM descriptors
var (
	// descriptorDeviceCDCACM is the default device descriptor for CDC-ACM
	// devices.
	descriptorDeviceCDCACM = []uint8{
		specDescriptorLengthDevice,           // Size of this descriptor in bytes
		specDescriptorTypeDevice,             // DEVICE Descriptor Type
		descriptorUSBSpecification[0],        // USB Specification Release Number in BCD (low)
		descriptorUSBSpecification[1],        // USB Specification Release Number in BCD (high)
		uint8(deviceClassCDC),                // Class code (assigned by the USB-IF).
		0,                                    // Subclass code (assigned by the USB-IF).
		deviceCDCNoClassSpecificProtocol,     // Protocol code (assigned by the USB-IF).
		configDeviceControllerMaxPacketSize,  // Maximum packet size for endpoint zero (8, 16, 32, or 64)
		configDescriptorDeviceCDCACMVID[0],   // Vendor ID (low) (assigned by the USB-IF)
		configDescriptorDeviceCDCACMVID[1],   // Vendor ID (high) (assigned by the USB-IF)
		configDescriptorDeviceCDCACMPID[0],   // Product ID (low) (assigned by the manufacturer)
		configDescriptorDeviceCDCACMPID[1],   // Product ID (high) (assigned by the manufacturer)
		configDescriptorDeviceCDCACMBCD[0],   // Device release number in BCD (low)
		configDescriptorDeviceCDCACMBCD[1],   // Device release number in BCD (high)
		0x01,                                 // Index of string descriptor describing manufacturer
		0x02,                                 // Index of string descriptor describing product
		0x00,                                 // Index of string descriptor describing the device's serial number
		configDeviceCDCACMConfigurationCount, // Number of possible configurations
	}

	// descriptorConfigurationCDCACM is the default configuration descriptor for
	// CDC-ACM devices.
	descriptorConfigurationCDCACM = []uint8{
		specDescriptorLengthConfigure,                 // Size of this descriptor in bytes
		specDescriptorTypeConfigure,                   // Descriptor Type
		uint8(descriptorConfigurationCDCACMSize),      // Total length of data returned for this configuration (low)
		uint8(descriptorConfigurationCDCACMSize >> 8), // Total length of data returned for this configuration (high)
		configDeviceCDCACMInterfaceCount,              // Number of interfaces supported by this configuration
		configDeviceCDCACMConfigurationIndex,          // Value to use to select this configuration
		0,                                             // Index of string descriptor describing this configuration
		descriptorConfigurationCDCACMAttributes,       // Configuration attributes
		configDeviceMaxPower,                          // Max power consumption when fully-operational (2 mA units)

		// Communication/Control Interface Descriptor
		specDescriptorLengthInterface,    // Descriptor length
		specDescriptorTypeInterface,      // Descriptor type
		descriptorInterfaceCDCACMComm,    // Interface index
		0,                                // Alternate setting
		1,                                // Number of endpoints
		deviceCDCCommClass,               // Class code
		deviceCDCAbstractControlModel,    // Subclass code
		deviceCDCNoClassSpecificProtocol, // Protocol code (NOTE: Teensyduino defines this as 1 [AT V.250])
		0,                                // Interface Description String Index

		// CDC Header Functional Descriptor
		descriptorConfigurationCDCACMHeaderFuncSize, // Size of this descriptor in bytes
		specDescriptorTypeCDCInterface,              // Descriptor Type
		deviceCDCHeaderFuncDesc,                     // Descriptor Subtype
		0x10,                                        // USB CDC specification version 1.10 (low)
		0x01,                                        // USB CDC specification version 1.10 (high)

		// CDC Call Management Functional Descriptor
		descriptorConfigurationCDCACMCallManageSize, // Size of this descriptor in bytes
		specDescriptorTypeCDCInterface,              // Descriptor Type
		deviceCDCCallManagementFuncDesc,             // Descriptor Subtype
		0x01,                                        // Capabilities
		1,                                           // Data Interface

		// CDC Abstract Control Management Functional Descriptor
		descriptorConfigurationCDCACMAbstractSize, // Size of this descriptor in bytes
		specDescriptorTypeCDCInterface,            // Descriptor Type
		deviceCDCAbstractControlFuncDesc,          // Descriptor Subtype
		0x06,                                      // Capabilities

		// CDC Union Functional Descriptor
		descriptorConfigurationCDCACMUnionFuncSize, // Size of this descriptor in bytes
		specDescriptorTypeCDCInterface,             // Descriptor Type
		deviceCDCUnionFuncDesc,                     // Descriptor Subtype
		0,                                          // Controlling interface index
		1,                                          // Controlled interface index

		// Communication/Control Notification Endpoint descriptor
		specDescriptorLengthEndpoint, // Size of this descriptor in bytes
		specDescriptorTypeEndpoint,   // Descriptor Type
		descriptorEndpointCDCACMCommInterruptIn | // Endpoint address
			specDescriptorEndpointAddressDirectionIn,
		uint8(specEndpointInterrupt),                        // Attributes
		uint8(configDeviceCDCACMInterruptInPacketSize),      // Max packet size (low)
		uint8(configDeviceCDCACMInterruptInPacketSize >> 8), // Max packet size (high)
		configDeviceCDCACMInterruptInInterval,               // Polling Interval

		// Data Interface Descriptor
		specDescriptorLengthInterface,    // Interface length
		specDescriptorTypeInterface,      // Interface type
		descriptorInterfaceCDCACMData,    // Interface index
		0,                                // Alternate setting
		2,                                // Number of endpoints
		deviceCDCDataClass,               // Class code
		0,                                // Subclass code
		deviceCDCNoClassSpecificProtocol, // Protocol code
		0,                                // Interface Description String Index

		// Data Bulk Input Endpoint descriptor
		specDescriptorLengthEndpoint, // Size of this descriptor in bytes
		specDescriptorTypeEndpoint,   // Descriptor Type
		descriptorEndpointCDCACMDataBulkIn | // Endpoint address
			specDescriptorEndpointAddressDirectionIn,
		uint8(specEndpointBulk),                        // Attributes
		uint8(configDeviceCDCACMBulkInPacketSize),      // Max packet size (low)
		uint8(configDeviceCDCACMBulkInPacketSize >> 8), // Max packet size (high)
		0, // Polling Interval

		// Data Bulk Output Endpoint descriptor
		specDescriptorLengthEndpoint, // Size of this descriptor in bytes
		specDescriptorTypeEndpoint,   // Descriptor Type
		descriptorEndpointCDCACMDataBulkOut | // Endpoint address
			specDescriptorEndpointAddressDirectionOut,
		uint8(specEndpointBulk),                         // Attributes
		uint8(configDeviceCDCACMBulkOutPacketSize),      // Max packet size (low)
		uint8(configDeviceCDCACMBulkOutPacketSize >> 8), // Max packet size (high)
		0, // Polling Interval
	}

	descriptorStringCDCACMLanguage = []uint8{
		2 + 2*1, // Size of this descriptor in bytes
		specDescriptorTypeString,
		0x09, 0x04,
	}
)
