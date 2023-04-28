package descriptor

const (
	cdcFunctionalHeader         = 0
	cdcFunctionalCallManagement = 0x1
	cdcFunctionalACM            = 0x2
	cdcFunctionalDirect         = 0x3
	cdcFunctionalRinger         = 0x4
	cdcFunctionalCall           = 0x5
	cdcFunctionalUnion          = 0x6
	cdcFunctionalCountry        = 0x7
	cdcFunctionalOperational    = 0x8
	cdcFunctionalUSB            = 0x9
	cdcFunctionalNetwork        = 0xa
	cdcFunctionalProtocol       = 0xb
	cdcFunctionalExtension      = 0xc
	cdcFunctionalMulti          = 0xd
	cdcFunctionalCAPI           = 0xe
	cdcFunctionalEthernet       = 0xf
	cdcFunctionalATM            = 0x10
)

var classSpecificCDCHeader = [classSpecificTypeLen]byte{
	classSpecificTypeLen,
	TypeClassSpecific,
	cdcFunctionalHeader,
	0x10, //
	0x1,  //
}

var ClassSpecificCDCHeader = ClassSpecificType{
	data: classSpecificCDCHeader[:],
}

var classSpecificCDCCallManagement = [classSpecificTypeLen]byte{
	classSpecificTypeLen,
	TypeClassSpecific,
	cdcFunctionalCallManagement,
	0x0, //
	0x1, //
}

var ClassSpecificCDCCallManagement = ClassSpecificType{
	data: classSpecificCDCCallManagement[:],
}

var classSpecificCDCACM = [classSpecificTypeLen]byte{
	4,
	TypeClassSpecific,
	cdcFunctionalACM,
	0x2, //
}

var ClassSpecificCDCACM = ClassSpecificType{
	data: classSpecificCDCACM[:],
}

var classSpecificCDCUnion = [classSpecificTypeLen]byte{
	classSpecificTypeLen,
	TypeClassSpecific,
	cdcFunctionalUnion,
	0x0, //
	0x1, //
}

var ClassSpecificCDCUnion = ClassSpecificType{
	data: classSpecificCDCUnion[:],
}

var interfaceAssociationCDC = [interfaceAssociationTypeLen]byte{
	interfaceAssociationTypeLen,
	TypeInterfaceAssociation,
	0x00, // FirstInterface
	0x02, // InterfaceCount
	0x02, // FunctionClass
	0x02, // FunctionSubClass
	0x01, // FunctionProtocol
	0x00, // Function
}

var InterfaceAssociationCDC = InterfaceAssociationType{
	data: interfaceAssociationCDC[:],
}

var deviceCDC = [deviceTypeLen]byte{
	deviceTypeLen,
	TypeDevice,
	0x00, 0x02, // USB version
	0xef,       // device class
	0x02,       // device subclass
	0x01,       // protocol
	0x40,       // maxpacketsize
	0x86, 0x28, // vendor id
	0x2d, 0x80, // product id
	0x00, 0x01, // device
	0x01, // manufacturer
	0x02, // product
	0x03, // SerialNumber
	0x01, // NumConfigurations
}

var DeviceCDC = DeviceType{
	data: deviceCDC[:],
}

var configurationCDC = [configurationTypeLen]byte{
	configurationTypeLen,
	TypeConfiguration,
	0x4b, 0x00, // adjust length as needed
	0x02, // number of interfaces
	0x01, // configuration value
	0x00, // index to string description
	0xa0, // attributes
	0x32, // maxpower
}

var ConfigurationCDC = ConfigurationType{
	data: configurationCDC[:],
}

var interfaceCDCControl = [interfaceTypeLen]byte{
	interfaceTypeLen,
	TypeInterface,
	0x00, // InterfaceNumber
	0x00, // AlternateSetting
	0x01, // NumEndpoints
	0x02, // InterfaceClass
	0x02, // InterfaceSubClass
	0x01, // InterfaceProtocol
	0x00, // Interface
}

var InterfaceCDCControl = InterfaceType{
	data: interfaceCDCControl[:],
}

var interfaceCDCData = [interfaceTypeLen]byte{
	interfaceTypeLen,
	TypeInterface,
	0x01, // InterfaceNumber
	0x00, // AlternateSetting
	0x02, // NumEndpoints
	0x0a, // InterfaceClass
	0x00, // InterfaceSubClass
	0x00, // InterfaceProtocol
	0x00, // Interface
}

var InterfaceCDCData = InterfaceType{
	data: interfaceCDCData[:],
}

var CDC = Descriptor{
	Device: DeviceCDC.Bytes(),
	Configuration: Append([][]byte{
		ConfigurationCDC.Bytes(),
		InterfaceAssociationCDC.Bytes(),
		InterfaceCDCControl.Bytes(),
		ClassSpecificCDCHeader.Bytes(),
		ClassSpecificCDCCallManagement.Bytes(),
		ClassSpecificCDCACM.Bytes(),
		ClassSpecificCDCUnion.Bytes(),
		EndpointEP1IN.Bytes(),
		InterfaceCDCData.Bytes(),
		EndpointEP2OUT.Bytes(),
		EndpointEP3IN.Bytes(),
	}),
}
