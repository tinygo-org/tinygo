package objfile

// This file provides an abstraction over object file writers.

import "debug/elf"

type ObjectFile interface {
	AddSymbol(name, section string, linkage Linkage, data []byte) int
	AddReloc(symbolIndex int, offset uint64, reloc Reloc, symbol string, addend int64)
	Bytes() []byte
}

type Linkage uint8

const (
	LinkageLocal Linkage = iota
	LinkageGlobal
	LinkageODR
)

func (b Linkage) elf() elf.SymBind {
	switch b {
	case LinkageLocal:
		return elf.STB_LOCAL
	case LinkageGlobal:
		return elf.STB_GLOBAL
	case LinkageODR:
		return elf.STB_WEAK
	default:
		panic("unknown symbol binding")
	}
}

type Reloc uint8

const (
	RelocNone Reloc = iota
	RelocADDR
	RelocCALL
	RelocPCREL
	RelocTLS_LE
	RelocPCREL_LDST64
)

func (r Reloc) String() string {
	switch r {
	case RelocADDR:
		return "ADDR"
	case RelocCALL:
		return "CALL"
	case RelocPCREL:
		return "PCREL"
	case RelocTLS_LE:
		return "TLS_LE"
	case RelocPCREL_LDST64:
		return "PCREL_LDST64"
	default:
		return "UNKNOWN"
	}
}
