package descriptor

const (
	interfaceClassMSC = 0x08
	mscSubclassSCSI   = 0x06
	mscProtocolBOT    = 0x50
)

var interfaceAssociationMSC = [interfaceAssociationTypeLen]byte{
	interfaceAssociationTypeLen,
	TypeInterfaceAssociation,
	0x02,              // FirstInterface
	0x01,              // InterfaceCount
	interfaceClassMSC, // FunctionClass
	mscSubclassSCSI,   // FunctionSubClass
	mscProtocolBOT,    // FunctionProtocol
	0x00,              // Function
}

var InterfaceAssociationMSC = InterfaceAssociationType{
	data: interfaceAssociationMSC[:],
}

var interfaceMSC = [interfaceTypeLen]byte{
	interfaceTypeLen,  // Length
	TypeInterface,     // DescriptorType
	0x02,              // InterfaceNumber
	0x00,              // AlternateSetting
	0x02,              // NumEndpoints
	interfaceClassMSC, // InterfaceClass (Mass Storage)
	mscSubclassSCSI,   // InterfaceSubClass (SCSI Transparent)
	mscProtocolBOT,    // InterfaceProtocol (Bulk-Only Transport)
	0x00,              // Interface
}

var InterfaceMSC = InterfaceType{
	data: interfaceMSC[:],
}

var configurationMSC = [configurationTypeLen]byte{
	configurationTypeLen,
	TypeConfiguration,
	0x6a, 0x00, // wTotalLength
	0x03, // number of interfaces (bNumInterfaces)
	0x01, // configuration value (bConfigurationValue)
	0x00, // index to string description (iConfiguration)
	0xa0, // attributes (bmAttributes)
	0x32, // maxpower (100 mA) (bMaxPower)
}

var ConfigurationMSC = ConfigurationType{
	data: configurationMSC[:],
}

// Mass Storage Class
var MSC = Descriptor{
	Device: DeviceCDC.Bytes(),
	Configuration: Append([][]byte{
		ConfigurationMSC.Bytes(),
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
		InterfaceAssociationMSC.Bytes(),
		InterfaceMSC.Bytes(),
		EndpointMSCIN.Bytes(),
		EndpointMSCOUT.Bytes(),
	}),
}
