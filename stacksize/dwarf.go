package stacksize

// This file implements parsing DWARF call frame information and interpreting
// the CFI bytecode, or enough of it for most practical code.

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"io"
)

// dwarfCIE represents one DWARF Call Frame Information structure.
type dwarfCIE struct {
	bytecode            []byte
	codeAlignmentFactor uint64
}

// parseFrames parses all call frame information from a .debug_frame section and
// provides the passed in symbols map with frame size information.
func parseFrames(f *elf.File, data []byte, symbols map[uint64]*CallNode) error {
	if f.Class != elf.ELFCLASS32 {
		// TODO: ELF64
		return fmt.Errorf("expected ELF32")
	}
	cies := make(map[uint32]*dwarfCIE)

	// Read each entity.
	r := bytes.NewBuffer(data)
	for {
		start := len(data) - r.Len()
		var length uint32
		err := binary.Read(r, binary.LittleEndian, &length)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		var cie uint32
		err = binary.Read(r, binary.LittleEndian, &cie)
		if err != nil {
			return err
		}
		if cie == 0xffffffff {
			// This is a CIE.
			var fields struct {
				Version      uint8
				Augmentation uint8
				AddressSize  uint8
				SegmentSize  uint8
			}
			err = binary.Read(r, binary.LittleEndian, &fields)
			if err != nil {
				return err
			}
			if fields.Version != 4 {
				return fmt.Errorf("unimplemented: .debug_frame version %d", fields.Version)
			}
			if fields.Augmentation != 0 {
				return fmt.Errorf("unimplemented: .debug_frame with augmentation")
			}
			if fields.SegmentSize != 0 {
				return fmt.Errorf("unimplemented: .debug_frame with segment size")
			}
			codeAlignmentFactor, err := readULEB128(r)
			if err != nil {
				return err
			}
			_, err = readSLEB128(r) // data alignment factor
			if err != nil {
				return err
			}
			_, err = readULEB128(r) // return address register
			if err != nil {
				return err
			}
			rest := (start + int(length) + 4) - (len(data) - r.Len())
			bytecode := r.Next(rest)
			cies[uint32(start)] = &dwarfCIE{
				codeAlignmentFactor: codeAlignmentFactor,
				bytecode:            bytecode,
			}
		} else {
			// This is a FDE.
			var fields struct {
				InitialLocation uint32
				AddressRange    uint32
			}
			err = binary.Read(r, binary.LittleEndian, &fields)
			if err != nil {
				return err
			}
			if _, ok := cies[cie]; !ok {
				return fmt.Errorf("could not find CIE 0x%x in .debug_frame section", cie)
			}
			frame := frameInfo{
				cie:    cies[cie],
				start:  uint64(fields.InitialLocation),
				loc:    uint64(fields.InitialLocation),
				length: uint64(fields.AddressRange),
			}
			rest := (start + int(length) + 4) - (len(data) - r.Len())
			bytecode := r.Next(rest)

			if frame.start == 0 {
				// Not sure where these come from but they don't seem to be
				// important.
				continue
			}

			_, err = frame.exec(frame.cie.bytecode)
			if err != nil {
				return err
			}
			entries, err := frame.exec(bytecode)
			if err != nil {
				return err
			}
			var maxFrameSize uint64
			for _, entry := range entries {
				switch f.Machine {
				case elf.EM_ARM:
					if entry.cfaRegister != 13 { // r13 or sp
						// something other than a stack pointer (on ARM)
						return fmt.Errorf("%08x..%08x: unknown CFA register number %d", frame.start, frame.start+frame.length, entry.cfaRegister)
					}
				default:
					return fmt.Errorf("unknown architecture: %s", f.Machine)
				}
				if entry.cfaOffset > maxFrameSize {
					maxFrameSize = entry.cfaOffset
				}
			}
			node := symbols[frame.start]
			if node.Size != frame.length {
				return fmt.Errorf("%s: symtab gives symbol length %d while DWARF gives symbol length %d", node, node.Size, frame.length)
			}
			node.FrameSize = maxFrameSize
			node.FrameSizeType = Bounded
			if debugPrint {
				fmt.Printf("%08x..%08x: frame size %4d %s\n", frame.start, frame.start+frame.length, maxFrameSize, node)
			}
		}
	}
}

// frameInfo contains the state of executing call frame information bytecode.
type frameInfo struct {
	cie         *dwarfCIE
	start       uint64
	loc         uint64
	length      uint64
	cfaRegister uint64
	cfaOffset   uint64
}

// frameInfoLine represents one line in the frame table (.debug_frame) at one
// point in the execution of the bytecode.
type frameInfoLine struct {
	loc         uint64
	cfaRegister uint64
	cfaOffset   uint64
}

func (fi *frameInfo) newLine() frameInfoLine {
	return frameInfoLine{
		loc:         fi.loc,
		cfaRegister: fi.cfaRegister,
		cfaOffset:   fi.cfaOffset,
	}
}

// exec executes the given bytecode in the CFI. Most CFI bytecode is actually
// very simple and provides a way to determine the maximum call frame size.
//
// The frame size often changes multiple times in a function, for example the
// frame size may be adjusted in the prologue and epilogue. Each frameInfoLine
// may contain such a change.
func (fi *frameInfo) exec(bytecode []byte) ([]frameInfoLine, error) {
	var entries []frameInfoLine
	r := bytes.NewBuffer(bytecode)
	for {
		op, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				entries = append(entries, fi.newLine())
				return entries, nil
			}
			return nil, err
		}
		// For details on the various opcodes, see:
		// http://dwarfstd.org/doc/DWARF5.pdf (page 239)
		highBits := op >> 6 // high order 2 bits
		lowBits := op & 0x1f
		switch highBits {
		case 1: // DW_CFA_advance_loc
			fi.loc += uint64(lowBits) * fi.cie.codeAlignmentFactor
			entries = append(entries, fi.newLine())
		case 2: // DW_CFA_offset
			// This indicates where a register is saved on the stack in the
			// prologue. We can ignore that for our purposes.
			_, err := readULEB128(r)
			if err != nil {
				return nil, err
			}
		case 3: // DW_CFA_restore
			// Restore a register. Used after an outlined function call.
			// It should be possible to ignore this.
			// TODO: check that this is not the stack pointer.
		case 0:
			switch lowBits {
			case 0: // DW_CFA_nop
				// no operation
			case 0x02: // DW_CFA_advance_loc1
				// Very similar to DW_CFA_advance_loc but allows for a slightly
				// larger range.
				offset, err := r.ReadByte()
				if err != nil {
					return nil, err
				}
				fi.loc += uint64(offset) * fi.cie.codeAlignmentFactor
				entries = append(entries, fi.newLine())
			case 0x03: // DW_CFA_advance_loc2
				var offset uint16
				err := binary.Read(r, binary.LittleEndian, &offset)
				if err != nil {
					return nil, err
				}
				fi.loc += uint64(offset) * fi.cie.codeAlignmentFactor
				entries = append(entries, fi.newLine())
			case 0x04: // DW_CFA_advance_loc4
				var offset uint32
				err := binary.Read(r, binary.LittleEndian, &offset)
				if err != nil {
					return nil, err
				}
				fi.loc += uint64(offset) * fi.cie.codeAlignmentFactor
				entries = append(entries, fi.newLine())
			case 0x05: // DW_CFA_offset_extended
				// Semantics are the same as DW_CFA_offset, but the encoding is
				// different. Ignore it just like DW_CFA_offset.
				_, err := readULEB128(r) // ULEB128 register
				if err != nil {
					return nil, err
				}
				_, err = readULEB128(r) // ULEB128 offset
				if err != nil {
					return nil, err
				}
			case 0x07: // DW_CFA_undefined
				// Marks a single register as undefined. This is used to stop
				// unwinding in tinygo_startTask using:
				//     .cfi_undefined lr
				// Ignore this directive.
				_, err := readULEB128(r)
				if err != nil {
					return nil, err
				}
			case 0x09: // DW_CFA_register
				// Copies a register. Emitted by the machine outliner, for example.
				// It should be possible to ignore this.
				// TODO: check that the stack pointer is not affected.
				_, err := readULEB128(r)
				if err != nil {
					return nil, err
				}
				_, err = readULEB128(r)
				if err != nil {
					return nil, err
				}
			case 0x0c: // DW_CFA_def_cfa
				register, err := readULEB128(r)
				if err != nil {
					return nil, err
				}
				offset, err := readULEB128(r)
				if err != nil {
					return nil, err
				}
				fi.cfaRegister = register
				fi.cfaOffset = offset
			case 0x0e: // DW_CFA_def_cfa_offset
				offset, err := readULEB128(r)
				if err != nil {
					return nil, err
				}
				fi.cfaOffset = offset
			default:
				return nil, fmt.Errorf("could not decode .debug_frame bytecode op 0x%x (for address 0x%x)", op, fi.loc)
			}
		default:
			return nil, fmt.Errorf("could not decode .debug_frame bytecode op 0x%x (for address 0x%x)", op, fi.loc)
		}
	}
}

// Source: https://en.wikipedia.org/wiki/LEB128#Decode_unsigned_integer
func readULEB128(r *bytes.Buffer) (result uint64, err error) {
	// TODO: guard against overflowing 64-bit integers.
	var shift uint8
	for {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		result |= uint64(b&0x7f) << shift
		if b&0x80 == 0 {
			break
		}
		shift += 7
	}
	return
}

// Source: https://en.wikipedia.org/wiki/LEB128#Decode_signed_integer
func readSLEB128(r *bytes.Buffer) (result int64, err error) {
	var shift uint8

	var b byte
	var rawResult uint64
	for {
		b, err = r.ReadByte()
		if err != nil {
			return 0, err
		}
		rawResult |= uint64(b&0x7f) << shift
		shift += 7
		if b&0x80 == 0 {
			break
		}
	}

	// sign bit of byte is second high order bit (0x40)
	if shift < 64 && b&0x40 != 0 {
		// sign extend
		rawResult |= ^uint64(0) << shift
	}
	result = int64(rawResult)

	return
}
