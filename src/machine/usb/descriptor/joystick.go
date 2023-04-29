package descriptor

var deviceJoystick = [deviceTypeLen]byte{
	deviceTypeLen,
	TypeDevice,
	0x00, 0x02, // USB version
	0xef,       // device class
	0x02,       // subclass
	0x01,       // protocol
	0x40,       // maxpacketsize
	0x41, 0x23, // vendor id
	0x36, 0x80, // product id
	0x00, 0x01, // device
	0x01, // manufacturer
	0x02, // product
	0x03, // SerialNumber
	0x01, // NumConfigurations
}

var DeviceJoystick = DeviceType{
	data: deviceJoystick[:],
}

var configurationCDCJoystick = [configurationTypeLen]byte{
	configurationTypeLen,
	TypeConfiguration,
	0x6b, 0x00, // adjust length as needed
	0x03, // number of interfaces
	0x01, // configuration value
	0x00, // index to string description
	0xa0, // attributes
	0xfa, // maxpower
}

var ConfigurationCDCJoystick = ConfigurationType{
	data: configurationCDCJoystick[:],
}

var interfaceHIDJoystick = [interfaceTypeLen]byte{
	interfaceTypeLen,
	TypeInterface,
	0x02, // InterfaceNumber
	0x00, // AlternateSetting
	0x02, // NumEndpoints
	0x03, // InterfaceClass
	0x00, // InterfaceSubClass
	0x00, // InterfaceProtocol
	0x00, // Interface
}

var InterfaceHIDJoystick = InterfaceType{
	data: interfaceHIDJoystick[:],
}

var classHIDJoystick = [ClassHIDTypeLen]byte{
	ClassHIDTypeLen,
	TypeClassHID,
	0x11, // HID version L
	0x01, // HID version H
	0x00, // CountryCode
	0x01, // NumDescriptors
	0x22, // ClassType
	0x00, // ClassLength L
	0x00, // ClassLength H
}

var ClassHIDJoystick = ClassHIDType{
	data: classHIDJoystick[:],
}

var JoystickDefaultHIDReport = Append([][]byte{
	HIDUsagePageGenericDesktop,
	HIDUsageDesktopJoystick,
	HIDCollectionApplication,
	HIDReportID(1),
	HIDUsagePageButton,
	HIDUsageMinimum(1),
	HIDUsageMaximum(16),
	HIDLogicalMinimum(0),
	HIDLogicalMaximum(1),
	HIDReportSize(1),
	HIDReportCount(16),
	HIDInputDataVarAbs,
	HIDReportCount(1),
	HIDReportSize(3),
	HIDUnitExponent(-16),
	HIDUnit(0),
	HIDInputDataVarAbs,

	HIDUsagePageGenericDesktop,
	HIDUsageDesktopHatSwitch,
	HIDLogicalMinimum(0),
	HIDLogicalMaximum(7),
	HIDPhysicalMinimum(0),
	HIDPhysicalMaximum(315),
	HIDUnit(0x14),
	HIDReportCount(1),
	HIDReportSize(4),
	HIDInputDataVarAbs,
	HIDUsageDesktopHatSwitch,
	HIDLogicalMinimum(0),
	HIDLogicalMaximum(7),
	HIDPhysicalMinimum(0),
	HIDPhysicalMaximum(315),
	HIDUnit(0x14),
	HIDReportCount(1),
	HIDReportSize(4),
	HIDInputDataVarAbs,
	HIDUsageDesktopPointer,
	HIDLogicalMinimum(-32767),
	HIDLogicalMaximum(32767),
	HIDReportCount(6),
	HIDReportSize(16),
	HIDCollectionPhysical,
	HIDUsageDesktopX,
	HIDUsageDesktopY,
	HIDUsageDesktopZ,
	HIDUsageDesktopRx,
	HIDUsageDesktopRy,
	HIDUsageDesktopRz,
	HIDInputDataVarAbs,
	HIDCollectionEnd,
	HIDCollectionEnd,
})

// CDCJoystick requires that you append the JoystickDescriptor
// to the Configuration before using. This is in order to support
// custom configurations.
var CDCJoystick = Descriptor{
	Device: DeviceJoystick.Bytes(),
	Configuration: Append([][]byte{
		ConfigurationCDCJoystick.Bytes(),
		InterfaceAssociationCDC.Bytes(),
		InterfaceCDCControl.Bytes(),
		ClassSpecificCDCHeader.Bytes(),
		ClassSpecificCDCACM.Bytes(),
		ClassSpecificCDCUnion.Bytes(),
		ClassSpecificCDCCallManagement.Bytes(),
		EndpointEP1IN.Bytes(),
		InterfaceCDCData.Bytes(),
		EndpointEP2OUT.Bytes(),
		EndpointEP3IN.Bytes(),
		InterfaceHIDJoystick.Bytes(),
		ClassHIDJoystick.Bytes(),
		EndpointEP4IN.Bytes(),
		EndpointEP5OUT.Bytes(),
	}),
	HID: map[uint16][]byte{},
}
