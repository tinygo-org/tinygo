package builder

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/blakesmith/ar"
)

// makeArchive creates an arcive for static linking from a list of object files
// given as a parameter. It is equivalent to the following command:
//
//     ar -rcs <archivePath> <objs...>
func makeArchive(archivePath string, objs []string) error {
	// Open the archive file.
	arfile, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer arfile.Close()
	arwriter := ar.NewWriter(arfile)
	err = arwriter.WriteGlobalHeader()
	if err != nil {
		return &os.PathError{Op: "write ar header", Path: archivePath, Err: err}
	}

	// Open all object files and read the symbols for the symbol table.
	symbolTable := []struct {
		name      string // symbol name
		fileIndex int    // index into objfiles
	}{}
	objfiles := make([]struct {
		file          *os.File
		archiveOffset int32
	}, len(objs))
	for i, objpath := range objs {
		objfile, err := os.Open(objpath)
		if err != nil {
			return err
		}
		objfiles[i].file = objfile

		// Read the symbols and add them to the symbol table.
		dbg, err := elf.NewFile(objfile)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", objpath, err)
		}
		symbols, err := dbg.Symbols()
		if err != nil {
			return err
		}
		for _, symbol := range symbols {
			bind := elf.ST_BIND(symbol.Info)
			if bind != elf.STB_GLOBAL && bind != elf.STB_WEAK {
				// Don't include local symbols (STB_LOCAL).
				continue
			}
			if elf.ST_TYPE(symbol.Info) != elf.STT_FUNC && elf.ST_TYPE(symbol.Info) != elf.STT_OBJECT {
				// Not a function.
				continue
			}
			// Include in archive.
			symbolTable = append(symbolTable, struct {
				name      string
				fileIndex int
			}{symbol.Name, i})
		}
	}

	// Create the symbol table buffer.
	// For some (sparse) details on the file format:
	// https://en.wikipedia.org/wiki/Ar_(Unix)#System_V_(or_GNU)_variant
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.BigEndian, int32(len(symbolTable)))
	for range symbolTable {
		// This is a placeholder index, it will be updated after all files have
		// been written to the archive (see the end of this function).
		err = binary.Write(buf, binary.BigEndian, int32(0))
		if err != nil {
			return err
		}
	}
	for _, sym := range symbolTable {
		_, err := buf.Write([]byte(sym.name + "\x00"))
		if err != nil {
			return err
		}
	}
	for buf.Len()%2 != 0 {
		// The symbol table must be aligned.
		// This appears to be required by lld.
		buf.WriteByte(0)
	}

	// Write the symbol table.
	err = arwriter.WriteHeader(&ar.Header{
		Name:    "/",
		ModTime: time.Unix(0, 0),
		Uid:     0,
		Gid:     0,
		Mode:    0,
		Size:    int64(buf.Len()),
	})
	if err != nil {
		return err
	}

	// Keep track of the start of the symbol table.
	symbolTableStart, err := arfile.Seek(0, os.SEEK_CUR)
	if err != nil {
		return err
	}

	// Write symbol table contents.
	_, err = arfile.Write(buf.Bytes())
	if err != nil {
		return err
	}

	// Add all object files to the archive.
	for i, objfile := range objfiles {
		// Store the start index, for when we'll update the symbol table with
		// the correct file start indices.
		offset, err := arfile.Seek(0, os.SEEK_CUR)
		if err != nil {
			return err
		}
		if int64(int32(offset)) != offset {
			return errors.New("large archives (4GB+) not supported: " + archivePath)
		}
		objfiles[i].archiveOffset = int32(offset)

		// Write the file header.
		st, err := objfile.file.Stat()
		if err != nil {
			return err
		}
		err = arwriter.WriteHeader(&ar.Header{
			Name:    filepath.Base(objfile.file.Name()),
			ModTime: time.Unix(0, 0),
			Uid:     0,
			Gid:     0,
			Mode:    0644,
			Size:    st.Size(),
		})
		if err != nil {
			return err
		}

		// Copy the file contents into the archive.
		n, err := io.Copy(arwriter, objfile.file)
		if err != nil {
			return err
		}
		if n != st.Size() {
			return errors.New("file modified during ar creation: " + archivePath)
		}

		// File is not needed anymore.
		objfile.file.Close()
	}

	// Create symbol indices.
	indicesBuf := &bytes.Buffer{}
	for _, sym := range symbolTable {
		err = binary.Write(indicesBuf, binary.BigEndian, objfiles[sym.fileIndex].archiveOffset)
		if err != nil {
			return err
		}
	}

	// Overwrite placeholder indices.
	_, err = arfile.WriteAt(indicesBuf.Bytes(), symbolTableStart+4)
	return err
}
