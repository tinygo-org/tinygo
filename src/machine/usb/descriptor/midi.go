package descriptor

var interfaceAssociationMIDI = [interfaceAssociationTypeLen]byte{
	interfaceAssociationTypeLen,
	TypeInterfaceAssociation,
	0x02, // EndpointAddress
	0x02, // Attributes
	0x01, // MaxPacketSizeL
	0x01, // MaxPacketSizeH
	0x00, // Interval
	0x00, // Interval
}

var InterfaceAssociationMIDI = InterfaceAssociationType{
	data: interfaceAssociationMIDI[:],
}

var interfaceAudio = [interfaceTypeLen]byte{
	interfaceTypeLen,
	TypeInterface,
	0x02, // InterfaceNumber
	0x00, // AlternateSetting
	0x00, // NumEndpoints
	0x01, // InterfaceClass
	0x01, // InterfaceSubClass
	0x00, // InterfaceProtocol
	0x00, // Interface
}

var InterfaceAudio = InterfaceType{
	data: interfaceAudio[:],
}

var interfaceMIDIStreaming = [interfaceTypeLen]byte{
	interfaceTypeLen,
	TypeInterface,
	0x03, // InterfaceNumber
	0x00, // AlternateSetting
	0x02, // NumEndpoints
	0x01, // InterfaceClass
	0x03, // InterfaceSubClass
	0x00, // InterfaceProtocol
	0x00, // Interface
}

var InterfaceMIDIStreaming = InterfaceType{
	data: interfaceMIDIStreaming[:],
}

const classSpecificAudioTypeLen = 9

var classSpecificAudioInterface = [classSpecificAudioTypeLen]byte{
	classSpecificAudioTypeLen,
	TypeClassSpecific,
	0x01,
	0x00, //
	0x01, //
	0x09,
	0x00,
	0x01,
	0x03,
}

var ClassSpecificAudioInterface = ClassSpecificType{
	data: classSpecificAudioInterface[:],
}

const classSpecificMIDIHeaderLen = 7

var classSpecificMIDIHeader = [classSpecificMIDIHeaderLen]byte{
	classSpecificMIDIHeaderLen,
	TypeClassSpecific,
	0x01,
	0x00, //
	0x01, //
	0x41,
	0x00,
}

var ClassSpecificMIDIHeader = ClassSpecificType{
	data: classSpecificMIDIHeader[:],
}

const classSpecificMIDIInJackLen = 6

var classSpecificMIDIInJack1 = [classSpecificMIDIInJackLen]byte{
	classSpecificMIDIInJackLen,
	TypeClassSpecific,
	0x02, // MIDI In jack
	0x01, // Jack type (embedded)
	0x01, // Jack ID
	0x00, // iJack
}

var ClassSpecificMIDIInJack1 = ClassSpecificType{
	data: classSpecificMIDIInJack1[:],
}

var classSpecificMIDIInJack2 = [classSpecificMIDIInJackLen]byte{
	classSpecificMIDIInJackLen,
	TypeClassSpecific,
	0x02, // MIDI In jack
	0x02, // Jack type (external)
	0x02, // Jack ID
	0x00, // iJack
}

var ClassSpecificMIDIInJack2 = ClassSpecificType{
	data: classSpecificMIDIInJack2[:],
}

const classSpecificMIDIOutJackLen = 9

var classSpecificMIDIOutJack1 = [classSpecificMIDIOutJackLen]byte{
	classSpecificMIDIOutJackLen,
	TypeClassSpecific,
	0x03, // MIDI Out jack
	0x01, // Jack type (embedded)
	0x03, // Jack ID
	0x01, // number of input pins
	0x02, // source ID
	0x01, // source pin
	0x00, // iJack
}

var ClassSpecificMIDIOutJack1 = ClassSpecificType{
	data: classSpecificMIDIOutJack1[:],
}

var classSpecificMIDIOutJack2 = [classSpecificMIDIOutJackLen]byte{
	classSpecificMIDIOutJackLen,
	TypeClassSpecific,
	0x03, // MIDI Out jack
	0x02, // Jack type (external)
	0x04, // Jack ID
	0x01, // number of input pins
	0x01, // source ID
	0x01, // source pin
	0x00, // iJack
}

var ClassSpecificMIDIOutJack2 = ClassSpecificType{
	data: classSpecificMIDIOutJack2[:],
}

const classSpecificMIDIEndpointLen = 5

var classSpecificMIDIOutEndpoint = [classSpecificMIDIEndpointLen]byte{
	classSpecificMIDIEndpointLen,
	TypeClassSpecificEndpoint,
	0x01, // CS endpoint
	0x01, // number MIDI IN jacks
	0x01, // assoc Jack ID
}

var ClassSpecificMIDIOutEndpoint = ClassSpecificType{
	data: classSpecificMIDIOutEndpoint[:],
}

var classSpecificMIDIInEndpoint = [classSpecificMIDIEndpointLen]byte{
	classSpecificMIDIEndpointLen,
	TypeClassSpecificEndpoint,
	0x01, // CS endpoint
	0x01, // number MIDI Out jacks
	0x03, // assoc Jack ID
}

var ClassSpecificMIDIInEndpoint = ClassSpecificType{
	data: classSpecificMIDIInEndpoint[:],
}

const endpointMIDITypeLen = 9

var endpointEP6IN = [endpointMIDITypeLen]byte{
	endpointMIDITypeLen,
	TypeEndpoint,
	0x86, // EndpointAddress
	0x02, // Attributes
	0x40, // MaxPacketSizeL
	0x00, // MaxPacketSizeH
	0x00, // Interval
	0x00, // refresh
	0x00, // sync address
}

var EndpointEP6IN = EndpointType{
	data: endpointEP6IN[:],
}

var endpointEP7OUT = [endpointMIDITypeLen]byte{
	endpointMIDITypeLen,
	TypeEndpoint,
	0x07, // EndpointAddress
	0x02, // Attributes
	0x40, // MaxPacketSizeL
	0x00, // MaxPacketSizeH
	0x00, // Interval
	0x00, // refresh
	0x00, // sync address
}

var EndpointEP7OUT = EndpointType{
	data: endpointEP7OUT[:],
}

var configurationCDCMIDI = [configurationTypeLen]byte{
	configurationTypeLen,
	TypeConfiguration,
	0xaf, 0x00, // adjust length as needed
	0x04, // number of interfaces
	0x01, // configuration value
	0x00, // index to string description
	0xa0, // attributes
	0x32, // maxpower
}

var ConfigurationCDCMIDI = ConfigurationType{
	data: configurationCDCMIDI[:],
}

var CDCMIDI = Descriptor{
	Device: DeviceCDC.Bytes(),
	Configuration: Append([][]byte{
		ConfigurationCDCMIDI.Bytes(),
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
		InterfaceAssociationMIDI.Bytes(),
		InterfaceAudio.Bytes(),
		ClassSpecificAudioInterface.Bytes(),
		InterfaceMIDIStreaming.Bytes(),
		ClassSpecificMIDIHeader.Bytes(),
		ClassSpecificMIDIInJack1.Bytes(),
		ClassSpecificMIDIInJack2.Bytes(),
		ClassSpecificMIDIOutJack1.Bytes(),
		ClassSpecificMIDIOutJack2.Bytes(),
		EndpointEP7OUT.Bytes(),
		ClassSpecificMIDIOutEndpoint.Bytes(),
		EndpointEP6IN.Bytes(),
		ClassSpecificMIDIInEndpoint.Bytes(),
	}),
}
