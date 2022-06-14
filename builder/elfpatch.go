package builder

import (
	"debug/elf"
	"fmt"
	"os"
)

func getElfSectionData(executable string, sectionName string) ([]byte, elf.FileHeader, error) {
	elfFile, err := elf.Open(executable)
	if err != nil {
		return nil, elf.FileHeader{}, err
	}
	defer elfFile.Close()

	section := elfFile.Section(sectionName)
	if section == nil {
		return nil, elf.FileHeader{}, nil
	}

	data, err := section.Data()

	return data, elfFile.FileHeader, err
}

func replaceElfSection(executable string, sectionName string, data []byte) error {
	fp, err := os.OpenFile(executable, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer fp.Close()

	elfFile, err := elf.Open(executable)
	if err != nil {
		return err
	}
	defer elfFile.Close()

	section := elfFile.Section(sectionName)
	if section == nil {
		return fmt.Errorf("could not find %s section", sectionName)
	}

	// Implicitly check for compressed sections
	if section.Size != section.FileSize {
		return fmt.Errorf("expected section %s to have identical size and file size, got %d and %d", sectionName, section.Size, section.FileSize)
	}

	// Only permit complete replacement of section
	if section.Size != uint64(len(data)) {
		return fmt.Errorf("expected section %s to have size %d, was actually %d", sectionName, len(data), section.Size)
	}

	// Write the replacement section data
	_, err = fp.WriteAt(data, int64(section.Offset))
	return err
}
