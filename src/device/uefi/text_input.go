package uefi

import (
	"errors"
	"unicode/utf8"
)

var errNilTextInputProtocol = errors.New("uefi: nil simple text input protocol")

type TextInputSource uint8

const (
	TextInputNone TextInputSource = iota
	TextInputSimpleTextInputEx
	TextInputSimpleTextInput
)

type TextInput struct {
	protoEx *EFI_SIMPLE_TEXT_INPUT_EX_PROTOCOL
	proto   *EFI_SIMPLE_TEXT_INPUT_PROTOCOL
	pending [utf8.UTFMax]byte
	start   int
	end     int
}

func NewTextInput(proto *EFI_SIMPLE_TEXT_INPUT_PROTOCOL) *TextInput {
	return &TextInput{proto: proto}
}

func NewTextInputEx(proto *EFI_SIMPLE_TEXT_INPUT_EX_PROTOCOL) *TextInput {
	return &TextInput{protoEx: proto}
}

func ConsoleInput() (*TextInput, error) {
	r := &TextInput{}
	if protoEx, status := SimpleTextInExProtocol(); status == EFI_SUCCESS {
		r.protoEx = protoEx
	}
	if proto, status := SimpleTextInProtocol(); status == EFI_SUCCESS {
		r.proto = proto
	}
	if r.protoEx == nil && r.proto == nil {
		return nil, ErrNotFound
	}
	return r, nil
}

func (r *TextInput) Read(p []byte) (int, error) {
	if r == nil || (r.protoEx == nil && r.proto == nil) {
		return 0, errNilTextInputProtocol
	}
	if len(p) == 0 {
		return 0, nil
	}

	n := 0
	for n < len(p) {
		if r.start != r.end {
			p[n] = r.pending[r.start]
			r.start++
			n++
			continue
		}

		key, err := r.ReadKey()
		if err != nil {
			if n != 0 {
				return n, nil
			}
			return 0, err
		}
		if key.Key.UnicodeChar == 0 {
			continue
		}

		runeValue := rune(key.Key.UnicodeChar)
		r.end = utf8.EncodeRune(r.pending[:], runeValue)
		r.start = 0
	}

	return n, nil
}

func (r *TextInput) ReadByte() (byte, error) {
	var buf [1]byte
	_, err := r.Read(buf[:])
	return buf[0], err
}

func (r *TextInput) ReadKey() (EFI_KEY_DATA, error) {
	key, _, err := r.ReadKeyWithSource()
	return key, err
}

func (r *TextInput) ReadKeyWithSource() (EFI_KEY_DATA, TextInputSource, error) {
	if r == nil {
		return EFI_KEY_DATA{}, TextInputNone, errNilTextInputProtocol
	}
	if r.protoEx != nil {
		key, status := r.protoEx.GetKey()
		if status == EFI_SUCCESS {
			return key, TextInputSimpleTextInputEx, nil
		}
		if r.proto == nil {
			return key, TextInputSimpleTextInputEx, StatusError(status)
		}
	}
	if r.proto != nil {
		key, status := r.proto.GetKey()
		return EFI_KEY_DATA{Key: key}, TextInputSimpleTextInput, StatusError(status)
	}
	return EFI_KEY_DATA{}, TextInputNone, errNilTextInputProtocol
}

func (s TextInputSource) String() string {
	switch s {
	case TextInputSimpleTextInputEx:
		return "STIEx"
	case TextInputSimpleTextInput:
		return "STIP"
	default:
		return "none"
	}
}

func (r *TextInput) HasTextInputEx() bool {
	return r != nil && r.protoEx != nil
}

func (r *TextInput) HasTextInput() bool {
	return r != nil && r.proto != nil
}
