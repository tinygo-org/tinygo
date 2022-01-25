package builder

import (
	"bytes"
	"debug/elf"
	"debug/macho"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"runtime"
)

// ReadBuildID reads the build ID from the currently running executable.
func ReadBuildID() ([]byte, error) {
	executable, err := os.Executable()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(executable)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	switch runtime.GOOS {
	case "linux", "freebsd":
		// Read the GNU build id section. (Not sure about FreeBSD though...)
		file, err := elf.NewFile(f)
		if err != nil {
			return nil, err
		}
		var gnuID, goID []byte
		for _, section := range file.Sections {
			if section.Type != elf.SHT_NOTE ||
				(section.Name != ".note.gnu.build-id" && section.Name != ".note.go.buildid") {
				continue
			}
			buf := make([]byte, section.Size)
			n, err := section.ReadAt(buf, 0)
			if uint64(n) != section.Size || err != nil {
				return nil, fmt.Errorf("could not read build id: %w", err)
			}
			if section.Name == ".note.gnu.build-id" {
				gnuID = buf
			} else {
				goID = buf
			}
		}
		if gnuID != nil {
			return gnuID, nil
		} else if goID != nil {
			return goID, nil
		}
	case "darwin":
		// Read the LC_UUID load command, which contains the equivalent of a
		// build ID.
		file, err := macho.NewFile(f)
		if err != nil {
			return nil, err
		}
		for _, load := range file.Loads {
			// Unfortunately, the debug/macho package doesn't support the
			// LC_UUID command directly. So we have to read it from
			// macho.LoadBytes.
			load, ok := load.(macho.LoadBytes)
			if !ok {
				continue
			}
			raw := load.Raw()
			command := binary.LittleEndian.Uint32(raw)
			if command != 0x1b {
				// Looking for the LC_UUID load command.
				// LC_UUID is defined here as 0x1b:
				// https://opensource.apple.com/source/xnu/xnu-4570.71.2/EXTERNAL_HEADERS/mach-o/loader.h.auto.html
				continue
			}
			return raw[4:], nil
		}
	default:
		// On other platforms (such as Windows) there isn't such a convenient
		// build ID. Luckily, Go does have an equivalent of the build ID, which
		// is stored as a special symbol named go.buildid. You can read it
		// using `go tool buildid`, but the code below extracts it directly
		// from the binary.
		// Unfortunately, because of stripping with the -w flag, no symbol
		// table might be available. Therefore, we have to scan the binary
		// directly. Luckily the build ID is always at the start of the file.
		// For details, see:
		// https://github.com/golang/go/blob/master/src/cmd/internal/buildid/buildid.go
		fileStart := make([]byte, 4096)
		_, err := io.ReadFull(f, fileStart)
		index := bytes.Index(fileStart, []byte("\xff Go build ID: \""))
		if index < 0 || index > len(fileStart)-103 {
			return nil, fmt.Errorf("could not find build id in %s", err)
		}
		buf := fileStart[index : index+103]
		if bytes.HasPrefix(buf, []byte("\xff Go build ID: \"")) && bytes.HasSuffix(buf, []byte("\"\n \xff")) {
			return buf[len("\xff Go build ID: \"") : len(buf)-1], nil
		}
	}
	return nil, fmt.Errorf("could not find build ID in %s", executable)
}
