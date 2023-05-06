package builder

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"

	"github.com/sigurn/crc16"
	"github.com/tinygo-org/tinygo/compileopts"
)

// Structure of the manifest.json file.
type jsonManifest struct {
	Manifest struct {
		Application struct {
			BinaryFile     string        `json:"bin_file"`
			DataFile       string        `json:"dat_file"`
			InitPacketData nrfInitPacket `json:"init_packet_data"`
		} `json:"application"`
		DFUVersion float64 `json:"dfu_version"` // yes, this is a JSON number, not a string
	} `json:"manifest"`
}

// Structure of the init packet.
// Source:
// https://github.com/adafruit/Adafruit_nRF52_Bootloader/blob/master/lib/sdk11/components/libraries/bootloader_dfu/dfu_init.h#L47-L57
type nrfInitPacket struct {
	ApplicationVersion uint32   `json:"application_version"`
	DeviceRevision     uint16   `json:"device_revision"`
	DeviceType         uint16   `json:"device_type"`
	FirmwareCRC16      uint16   `json:"firmware_crc16"`
	SoftDeviceRequired []uint16 `json:"softdevice_req"` // this is actually a variable length array
}

// Create the init packet (the contents of application.dat).
func (p nrfInitPacket) createInitPacket() []byte {
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, p.DeviceType)                      // uint16_t device_type;
	binary.Write(buf, binary.LittleEndian, p.DeviceRevision)                  // uint16_t device_rev;
	binary.Write(buf, binary.LittleEndian, p.ApplicationVersion)              // uint32_t app_version;
	binary.Write(buf, binary.LittleEndian, uint16(len(p.SoftDeviceRequired))) // uint16_t softdevice_len;
	binary.Write(buf, binary.LittleEndian, p.SoftDeviceRequired)              // uint16_t softdevice[1];
	binary.Write(buf, binary.LittleEndian, p.FirmwareCRC16)
	return buf.Bytes()
}

// Make a Nordic DFU firmware image from an ELF file.
func makeDFUFirmwareImage(options *compileopts.Options, infile, outfile string) error {
	// Read ELF file as input and convert it to a binary image file.
	_, data, err := extractROM(infile)
	if err != nil {
		return err
	}

	// Create the zip file in memory.
	// It won't be very large anyway.
	buf := &bytes.Buffer{}
	w := zip.NewWriter(buf)

	// Write the application binary to the zip file.
	binw, err := w.Create("application.bin")
	if err != nil {
		return err
	}
	_, err = binw.Write(data)
	if err != nil {
		return err
	}

	// Create the init packet.
	initPacket := nrfInitPacket{
		ApplicationVersion: 0xffff_ffff, // appears to be unused by the Adafruit bootloader
		DeviceRevision:     0xffff,      // DFU_DEVICE_REVISION_EMPTY
		DeviceType:         0x0052,      // ADAFRUIT_DEVICE_TYPE
		FirmwareCRC16:      crc16.Checksum(data, crc16.MakeTable(crc16.CRC16_CCITT_FALSE)),
		SoftDeviceRequired: []uint16{0xfffe}, // DFU_SOFTDEVICE_ANY
	}

	// Write the init packet to the zip file.
	datw, err := w.Create("application.dat")
	if err != nil {
		return err
	}
	_, err = datw.Write(initPacket.createInitPacket())
	if err != nil {
		return err
	}

	// Create the JSON manifest.
	manifest := &jsonManifest{}
	manifest.Manifest.Application.BinaryFile = "application.bin"
	manifest.Manifest.Application.DataFile = "application.dat"
	manifest.Manifest.Application.InitPacketData = initPacket
	manifest.Manifest.DFUVersion = 0.5

	// Write the JSON manifest to the file.
	jsonw, err := w.Create("manifest.json")
	if err != nil {
		return err
	}
	enc := json.NewEncoder(jsonw)
	enc.SetIndent("", "    ")
	err = enc.Encode(manifest)
	if err != nil {
		return err
	}

	// Finish the zip file.
	err = w.Close()
	if err != nil {
		return err
	}

	return os.WriteFile(outfile, buf.Bytes(), 0o666)
}
