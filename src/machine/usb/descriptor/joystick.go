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

var JoystickDefaultHIDReport = []byte{
	0x05, 0x01, // Usage page
	0x09, 0x04, // Joystick
	0xa1, 0x01, // COLLECTION (Application)
	0x85, 0x01, // REPORT_ID (1)
	0x05, 0x09, // USAGE_PAGE (Button)
	0x19, 0x01, // USAGE_MINIMUM (Button 1)
	0x29, 0x10, // USAGE_MAXIMUM (Button 16)
	0x15, 0x00, // LOGICAL_MINIMUM (0)
	0x25, 0x01, // LOGICAL_MAXIMUM (1)
	0x75, 0x01, // REPORT_SIZE (1)
	0x95, 0x10, // REPORT_COUNT (16)
	0x55, 0x00, // Unit Exponent (-16)
	0x65, 0x00, // Unit (0x00)
	0x81, 0x02, // INPUT (Data/Var/Abs)
	0x05, 0x01, // USAGE_PAGE (Generic Desktop Controls)
	0x09, 0x39, // USAGE(Hat Switch)
	0x15, 0x00, // LOGICAL_MINIMUM (0)
	0x25, 0x07, // LOGICAL_MAXIMUM (7)
	0x35, 0x00, // PHYSICAL_MINIMUM (0)
	0x46, 0x3b, 0x01, // PHYSICAL_MAXIMUM(315)
	0x65, 0x14, // UNIT (Eng Rot:Angular Pos)
	0x75, 0x04, // REPORT_SIZE (4)
	0x95, 0x01, // REPORT_COUNT (1)
	0x81, 0x02, // INPUT (Data/Var/Abs)
	0x09, 0x39, // USAGE(Hat Switch)
	0x15, 0x00, // LOGICAL_MINIMUM (0)
	0x25, 0x07, // LOGICAL_MAXIMUM (7)
	0x35, 0x00, // PHYSICAL_MINIMUM (0)
	0x46, 0x3b, 0x01, // PHYSICAL_MAXIMUM(315)
	0x65, 0x14, // UNIT (Eng Rot:Angular Pos)
	0x75, 0x04, // REPORT_SIZE (4)
	0x95, 0x01, // REPORT_COUNT (1)
	0x81, 0x02, // INPUT (Data/Var/Abs)
	0x09, 0x01, // USAGE (Pointer)
	0x16, 0x01, 0x80, // LOGICAL_MINIMUM (-32767)
	0x26, 0xff, 0x7f, // LOGICAL_MAXIMUM (32767)
	0x75, 0x10, // REPORT_SIZE (16bits)
	0x95, 0x06, // REPORT_COUNT (6)
	0xa1, 0x00, // COLLECTION (Physical)
	0x09, 0x30, // USAGE(X)
	0x09, 0x31, // USAGE(Y)
	0x09, 0x32, // USAGE(Z)
	0x09, 0x33, // USAGE(RX)
	0x09, 0x34, // USAGE(RY)
	0x09, 0x35, // USAGE(RZ)
	0x81, 0x02, // INPUT (Data/Var/Abs)
	0xc0, // END_COLLECTION
	0xc0, // END_COLLECTION
}

// CDCJoystick requires that you append the JoystickDescriptor
// to the Configuration before using. This is in order to support
// custom configurations.
var CDCJoystick = Descriptor{
	Device: DeviceJoystick.Bytes(),
	Configuration: appendSlices([][]byte{
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
