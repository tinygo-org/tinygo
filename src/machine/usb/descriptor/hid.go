package descriptor

import (
	"bytes"
	"encoding/binary"
	"errors"
)

var configurationCDCHID = [configurationTypeLen]byte{
	configurationTypeLen,
	TypeConfiguration,
	0x64, 0x00, // adjust length as needed
	0x03, // number of interfaces
	0x01, // configuration value
	0x00, // index to string description
	0xa0, // attributes
	0x32, // maxpower
}

var ConfigurationCDCHID = ConfigurationType{
	data: configurationCDCHID[:],
}

var interfaceHID = [interfaceTypeLen]byte{
	interfaceTypeLen,
	TypeInterface,
	0x02, // InterfaceNumber
	0x00, // AlternateSetting
	0x01, // NumEndpoints
	0x03, // InterfaceClass
	0x00, // InterfaceSubClass
	0x00, // InterfaceProtocol
	0x00, // Interface
}

var InterfaceHID = InterfaceType{
	data: interfaceHID[:],
}

const (
	ClassHIDTypeLen = 9
)

type ClassHIDType struct {
	data []byte
}

func (d ClassHIDType) Bytes() []byte {
	return d.data[:]
}

func (d ClassHIDType) Length(v uint8) {
	d.data[0] = byte(v)
}

func (d ClassHIDType) Type(v uint8) {
	d.data[1] = byte(v)
}

func (d ClassHIDType) HID(v uint16) {
	binary.LittleEndian.PutUint16(d.data[2:4], v)
}

func (d ClassHIDType) CountryCode(v uint8) {
	d.data[4] = byte(v)
}

func (d ClassHIDType) NumDescriptors(v uint8) {
	d.data[5] = byte(v)
}

func (d ClassHIDType) ClassType(v uint8) {
	d.data[6] = byte(v)
}

func (d ClassHIDType) ClassLength(v uint16) {
	binary.LittleEndian.PutUint16(d.data[7:9], v)
}

var errNoClassHIDFound = errors.New("no classHID found")

// FindClassHIDType tries to find the ClassHID class in the descriptor.
func FindClassHIDType(des, class []byte) (ClassHIDType, error) {
	if len(des) < ClassHIDTypeLen || len(class) == 0 {
		return ClassHIDType{}, errNoClassHIDFound
	}

	// search only for ClassHIDType without the ClassLength,
	// in case it has already been set.
	idx := bytes.Index(des, class[:ClassHIDTypeLen-2])
	if idx == -1 {
		return ClassHIDType{}, errNoClassHIDFound
	}

	return ClassHIDType{data: des[idx : idx+ClassHIDTypeLen]}, nil
}

var classHID = [ClassHIDTypeLen]byte{
	ClassHIDTypeLen,
	TypeClassHID,
	0x11, // HID version L
	0x01, // HID version H
	0x00, // CountryCode
	0x01, // NumDescriptors
	0x22, // ClassType
	0x7E, // ClassLength L
	0x00, // ClassLength H
}

var ClassHID = ClassHIDType{
	data: classHID[:],
}

var CDCHID = Descriptor{
	Device: DeviceCDC.Bytes(),
	Configuration: appendSlices([][]byte{
		ConfigurationCDCHID.Bytes(),
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
		InterfaceHID.Bytes(),
		ClassHID.Bytes(),
		EndpointEP4IN.Bytes(),
	}),
	HID: map[uint16][]byte{
		2: []byte{
			0x05, 0x01, // Usage Page (Generic Desktop)
			0x09, 0x06, // Usage (Keyboard)
			0xa1, 0x01, // Collection (Application)
			0x85, 0x02, // Report ID (2)
			0x05, 0x07, // Usage Page (KeyCodes)
			0x19, 0xe0, // Usage Minimum (224)
			0x29, 0xe7, // Usage Maximum (231)
			0x15, 0x00, // Logical Minimum (0)
			0x25, 0x01, // Logical Maximum (1)
			0x75, 0x01, // Report Size (1)
			0x95, 0x08, // Report Count (8)
			0x81, 0x02, // Input (Data, Variable, Absolute), Modifier byte
			0x95, 0x01, // Report Count (1)
			0x75, 0x08, // Report Size (8)
			0x81, 0x03, //
			0x95, 0x06, // Report Count (6)
			0x75, 0x08, // Report Size (8)
			0x15, 0x00, // Logical Minimum (0),
			0x25, 0xFF, //
			0x05, 0x07, // Usage Page (KeyCodes)
			0x19, 0x00, // Usage Minimum (0)
			0x29, 0xFF, // Usage Maximum (255)
			0x81, 0x00, // Input (Data, Array), Key arrays (6 bytes)
			0xc0, // End Collection

			0x05, 0x01, // Usage Page (Generic Desktop)
			0x09, 0x02, // Usage (Mouse)
			0xa1, 0x01, // Collection (Application)
			0x09, 0x01, // Usage (Pointer)
			0xa1, 0x00, // Collection (Physical)
			0x85, 0x01, // Report ID (1)
			0x05, 0x09, // Usage Page (Buttons)
			0x19, 0x01, // Usage Minimum (01)
			0x29, 0x03, // Usage Maximun (03)
			0x15, 0x00, // Logical Minimum (0)
			0x25, 0x01, // Logical Maximum (1)
			0x95, 0x03, // Report Count (3)
			0x75, 0x01, // Report Size (1)
			0x81, 0x02, // Input (Data, Variable, Absolute), ;3 button bits
			0x95, 0x01, // Report Count (1)
			0x75, 0x05, // Report Size (5)
			0x81, 0x03, //
			0x05, 0x01, // Usage Page (Generic Desktop)
			0x09, 0x30, // Usage (X)
			0x09, 0x31, // Usage (Y)
			0x09, 0x38, //
			0x15, 0x81, // Logical Minimum (-127)
			0x25, 0x7f, // Logical Maximum (127)
			0x75, 0x08, // Report Size (8)
			0x95, 0x03, // Report Count (3)
			0x81, 0x06, // Input (Data, Variable, Relative), 2 position bytes (X & Y)
			0xc0, // End Collection
			0xc0, // End Collection

			0x05, 0x0C, // Usage Page (Consumer)
			0x09, 0x01, // Usage (Consumer Control)
			0xA1, 0x01, // Collection (Application)
			0x85, 0x03, // Report ID (3)
			0x15, 0x00, // Logical Minimum (0)
			0x26, 0xFF, 0x1F, // Logical Maximum (8191)
			0x19, 0x00, // Usage Minimum (Unassigned)
			0x2A, 0xFF, 0x1F, // Usage Maximum (0x1FFF)
			0x75, 0x10, // Report Size (16)
			0x95, 0x01, // Report Count (1)
			0x81, 0x00, // Input (Data,Array,Abs,No Wrap,Linear,Preferred State,No Null Position)
			0xC0, // End Collection
		},
	},
}
