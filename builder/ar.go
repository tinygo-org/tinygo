package builder

import (
	"bytes"
	"debug/elf"
	"debug/pe"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	wasm "github.com/aykevl/go-wasm"
	"github.com/blakesmith/ar"
)

// makeArchive creates an archive for static linking from a list of object files
// given as a parameter. It is equivalent to the following command:
//
//	ar -rcs <archivePath> <objs...>
func makeArchive(arfile *os.File, objs []string) error {
	// Open the archive file.
	arwriter := ar.NewWriter(arfile)
	err := arwriter.WriteGlobalHeader()
	if err != nil {
		return &os.PathError{Op: "write ar header", Path: arfile.Name(), Err: err}
	}

	// Open all object files and read the symbols for the symbol table.
	symbolTable := []struct {
		name      string // symbol name
		fileIndex int    // index into objfiles
	}{}
	archiveOffsets := make([]int32, len(objs))
	for i, objpath := range objs {
		objfile, err := os.Open(objpath)
		if err != nil {
			return err
		}

		// Read the symbols and add them to the symbol table.
		if dbg, err := elf.NewFile(objfile); err == nil {
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
		} else if dbg, err := pe.NewFile(objfile); err == nil {
			for _, symbol := range dbg.Symbols {
				if symbol.StorageClass != 2 {
					continue
				}
				if symbol.SectionNumber == 0 {
					continue
				}
				symbolTable = append(symbolTable, struct {
					name      string
					fileIndex int
				}{symbol.Name, i})
			}
		} else if dbg, err := wasm.Parse(objfile); err == nil {
			for _, s := range dbg.Sections {
				switch section := s.(type) {
				case *wasm.SectionLinking:
					for _, symbol := range section.Symbols {
						if symbol.Flags&wasm.LinkingSymbolFlagUndefined != 0 {
							// Don't list undefined functions.
							continue
						}
						if symbol.Flags&wasm.LinkingSymbolFlagBindingLocal != 0 {
							// Don't include local symbols.
							continue
						}
						if symbol.Kind != wasm.LinkingSymbolKindFunction && symbol.Kind != wasm.LinkingSymbolKindData {
							// Link functions and data symbols.
							// Some data symbols need to be included, such as
							// __log_data.
							continue
						}
						// Include in the archive.
						symbolTable = append(symbolTable, struct {
							name      string
							fileIndex int
						}{symbol.Name, i})
					}
				}
			}
		} else {
			return fmt.Errorf("failed to open file %s as WASM, ELF or PE/COFF: %w", objpath, err)
		}

		// Close file, to avoid issues with too many open files (especially on
		// MacOS X).
		objfile.Close()
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
	symbolTableStart, err := arfile.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	// Write symbol table contents.
	_, err = arfile.Write(buf.Bytes())
	if err != nil {
		return err
	}

	// Add all object files to the archive.
	var copyBuf bytes.Buffer
	for i, objpath := range objs {
		objfile, err := os.Open(objpath)
		if err != nil {
			return err
		}
		defer objfile.Close()

		// Store the start index, for when we'll update the symbol table with
		// the correct file start indices.
		offset, err := arfile.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}
		if int64(int32(offset)) != offset {
			return errors.New("large archives (4GB+) not supported: " + arfile.Name())
		}
		archiveOffsets[i] = int32(offset)

		// Write the file header.
		st, err := objfile.Stat()
		if err != nil {
			return err
		}
		err = arwriter.WriteHeader(&ar.Header{
			Name:    filepath.Base(objfile.Name()),
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
		// First load all contents into a buffer, then write it all in one go to
		// the archive file. This is a bit complicated, but is necessary because
		// io.Copy can't deal with files that are of an odd size.
		copyBuf.Reset()
		n, err := io.Copy(&copyBuf, objfile)
		if err != nil {
			return fmt.Errorf("could not copy object file into ar file: %w", err)
		}
		if n != st.Size() {
			return errors.New("file modified during ar creation: " + arfile.Name())
		}
		_, err = arwriter.Write(copyBuf.Bytes())
		if err != nil {
			return fmt.Errorf("could not copy object file into ar file: %w", err)
		}

		// File is not needed anymore.
		objfile.Close()
	}

	// Create symbol indices.
	indicesBuf := &bytes.Buffer{}
	for _, sym := range symbolTable {
		err = binary.Write(indicesBuf, binary.BigEndian, archiveOffsets[sym.fileIndex])
		if err != nil {
			return err
		}
	}

	// Overwrite placeholder indices.
	_, err = arfile.WriteAt(indicesBuf.Bytes(), symbolTableStart+4)
	return err
}
