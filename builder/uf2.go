package builder

// This file converts firmware files from BIN to UF2 format before flashing.
//
// For more information about the UF2 firmware file format, please see:
// https://github.com/Microsoft/uf2
//
//

import (
	"math"
	"os"

	"github.com/soypat/tinyboot/build/uf2"
)

// convertELFFileToUF2File converts an ELF file to a UF2 file.
func convertELFFileToUF2File(infile, outfile string, uf2FamilyID string) error {
	uf2Formatter := uf2.Formatter{ChunkSize: 256}
	if uf2FamilyID != "" {
		err := uf2Formatter.SetFamilyID(uf2FamilyID)
		if err != nil {
			return err
		}
	}
	start, ROM, err := extractROM(infile)
	if err != nil {
		return err
	} else if start > math.MaxUint32 {
		return objcopyError{Op: "ELF start ROM address overflows uint32"}
	}

	// Write raw ROM contents in UF2 format to outfile.
	expectBlocks := len(ROM)/int(uf2Formatter.ChunkSize) + 1
	uf2data := make([]byte, expectBlocks*uf2.BlockSize)
	uf2data, _, err = uf2Formatter.AppendTo(uf2data[:0], ROM, uint32(start))
	if err != nil {
		return objcopyError{Op: "writing UF2 blocks", Err: err}
	}
	return os.WriteFile(outfile, uf2data, 0644)
}
