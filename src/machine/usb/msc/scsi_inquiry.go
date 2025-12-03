package msc

import (
	"encoding/binary"
	"machine/usb/msc/csw"
	"machine/usb/msc/scsi"
)

type vpdPage struct {
	PageCode   uint8
	PageLength uint8
	// Page data
	// First four bytes are always Device Type, Page Code, and Page Length (2 bytes) and are omitted here
	Data []byte
}

// These must be sorted in ascending order by PageCode
var vpdPages = []vpdPage{
	{
		// 0xb0 - 5.4.5 Block Limits VPD page (B0h)
		// https://www.seagate.com/files/staticfiles/support/docs/manual/Interface%20manuals/100293068j.pdf
		PageCode:   0xb0,
		PageLength: 0x3c, // 60 bytes
		Data: []byte{
			0x00, 0x00, // WSNZ, MAXIMUM COMPARE AND WRITE LENGTH - Not supported
			0x00, 0x00, // OPTIMAL TRANSFER LENGTH GRANULARITY - Not supported
			0x00, 0x00, 0x00, 0x00, // MAXIMUM TRANSFER LENGTH - Not supported
			0x00, 0x00, 0x00, 0x00, // OPTIMAL TRANSFER LENGTH - Not supported
			0x00, 0x00, 0x00, 0x00, // MAXIMUM PREFETCH LENGTH - Not supported
			0xFF, 0xFF, 0xFF, 0xFF, // MAXIMUM UNMAP LBA COUNT - Maximum count supported
			0x00, 0x00, 0x00, 0x03, // MAXIMUM UNMAP BLOCK DESCRIPTOR COUNT - Max 3 descriptors
			0x00, 0x00, 0x00, 0x00, // OPTIMAL UNMAP GRANULARITY
			0x00, 0x00, 0x00, 0x00, // UNMAP GRANULARITY ALIGNMENT (bit 7 on byte 28 sets UGAVALID)
			// From here on all bytes are zero and can be omitted from the response
			// 0x00, 0x00, 0x00, 0x00, // MAXIMUM WRITE SAME LENGTH - Not supported
			// 0x00, 0x00, 0x00, 0x00, // (8-bytes)
			// 0x00, 0x00, 0x00, 0x00, // MAXIMUM ATOMIC TRANSFER LENGTH - Not supported
			// 0x00, 0x00, 0x00, 0x00, // ATOMIC ALIGNMENT - Not supported
			// 0x00, 0x00, 0x00, 0x00, // ATOMIC TRANSFER LENGTH GRANULARITY - Not supported
			// 0x00, 0x00, 0x00, 0x00, // MAXIMUM ATOMIC TRANSFER LENGTH WITH ATOMIC BOUNDARY - Not supported
			// 0x00, 0x00, 0x00, 0x00, // MAXIMUM ATOMIC BOUNDARY SIZE - Not supported
		},
	},
	{
		// 0xb1 - 5.4.3 Block Device Characteristics VPD page (B1h)
		// https://www.seagate.com/files/staticfiles/support/docs/manual/Interface%20manuals/100293068j.pdf
		PageCode:   0xb1,
		PageLength: 0x3c, // 60 bytes (bytes 9+ are all reserved/zero)
		Data: []byte{
			0x00, 0x01, // Rotation rate (0x0001 - non-rotating medium)
			0x00, // Product type - 0x00: Not indicated, 0x04: MMC/eMMC, 0x05: SD card
			0x00, // WABEREQ/WACEREQ/Form Factor - Not specified
			0x00, // ZBC/BOCS/FUAB/VBULS
			// Reserved (55 bytes)
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00,
		},
	},
	{
		// 0xb2 - 5.4.13 Logical Block Provisioning VPD page (B2h)
		// https://www.seagate.com/files/staticfiles/support/docs/manual/Interface%20manuals/100293068j.pdf
		PageCode:   0xB2,
		PageLength: 0x04,
		Data: []byte{
			0x00, // Logical Block Provisioning Threshold Exponent
			0x80, // 0x80 - LBPU (UNMAP command supported)
			0x00, // Minimum percentage/Provisioning type - Not specified
			0x00, // Threshold percentage - Not supported
		},
	},
}

func (m *msc) scsiCmdInquiry(cmd scsi.Cmd) {
	evpd := cmd.Data[1] & 0x01
	pageCode := cmd.Data[2]

	// PAGE CODE (byte 2) can't be set if the EVPD bit is not set
	if evpd == 0 {
		if pageCode == 0 {
			// Standard INQUIRY command
			m.scsiStdInquiry(cmd)
		} else {
			// 3.6.1 INQUIRY command introduction
			// https://www.seagate.com/files/staticfiles/support/docs/manual/Interface%20manuals/100293068j.pdf
			m.sendScsiError(csw.StatusFailed, scsi.SenseIllegalRequest, scsi.SenseCodeInvalidFieldInCDB)
			return
		}
	} else {
		m.scsiEvpdInquiry(cmd, pageCode)
	}
}

func (m *msc) scsiEvpdInquiry(cmd scsi.Cmd, pageCode uint8) {
	var pageLength int
	switch pageCode {
	case 0x00:
		// 5.4.18 Supported Vital Product Data pages (00h)
		// https://www.seagate.com/files/staticfiles/support/docs/manual/Interface%20manuals/100293068j.pdf

		pageLength = len(vpdPages) + 1 // Number of pages + 1 for 0x00 (excluded from vpdPages[])
		m.resetBuffer(pageLength + 4)  // n+4 supported VPD pages
		// bytes 4+ - Supported VPD pages in ascending order
		for i := 0; i < len(vpdPages); i++ {
			m.buf[4+i] = vpdPages[i].PageCode
		}
	default:
		found := false
		for i := range vpdPages {
			if vpdPages[i].PageCode == pageCode {
				// Our advertised page length is "for entertainment use only". Some pages have dozens of
				// reserved (zero) bytes at the end that don't actually need to be sent. If we omit them
				// from our response they are (correctly) presumed to be zero bytes by the host
				pageLength = int(vpdPages[i].PageLength)
				// We actually just send the length of the bytes we have plus the same four byte header,
				// but declare the length of the response according to the spec as appropriate
				m.resetBuffer(len(vpdPages[i].Data) + 4)
				copy(m.buf[4:], vpdPages[i].Data)
				found = true
				break
			}
		}
		if !found {
			// VPD page not found, send error
			m.sendScsiError(csw.StatusFailed, scsi.SenseIllegalRequest, scsi.SenseCodeInvalidFieldInCDB)
			return
		}
	}

	// byte 0 - Peripheral Qualifier/Peripheral Device Type (0x00 for direct access block device)
	m.buf[1] = pageCode
	binary.BigEndian.PutUint16(m.buf[2:4], uint16(pageLength))

	// Set total bytes to the length of our response
	m.queuedBytes = uint32(len(m.buf))
	m.transferBytes = uint32(len(m.buf))
}

func (m *msc) scsiStdInquiry(cmd scsi.Cmd) {
	m.resetBuffer(scsi.InquiryRespLen)
	m.queuedBytes = scsi.InquiryRespLen
	m.transferBytes = scsi.InquiryRespLen

	// byte 0 - Device Type (0x00 for direct access block device)
	// byte 1 - Removable media bit
	m.buf[1] = 0x80
	// byte 2 - Version 0x00 - We claim conformance to no standard
	// byte 3 - Response data format
	m.buf[3] = 2
	// byte 4 - Additional length (number of bytes after this one)
	m.buf[4] = scsi.InquiryRespLen - 5
	// byte 5 - Not used
	// byte 6 - Not used
	// byte 7 - Not used
	// bytes 8-15 - Vendor ID
	copy(m.buf[8:16], m.vendorID[:])
	// bytes 16-31 - Product ID
	copy(m.buf[16:32], m.productID[:])
	// bytes 32-35 - Product revision level
	copy(m.buf[32:36], m.productRev[:])
}
