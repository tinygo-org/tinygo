package uefi

import "errors"

var errNilTextOutputProtocol = errors.New("uefi: nil simple text output protocol")

type TextOutput struct {
	proto *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL
}

func NewTextOutput(proto *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL) *TextOutput {
	return &TextOutput{proto: proto}
}

func ConsoleOut() *TextOutput {
	return NewTextOutput(ST().ConOut)
}

func StandardError() *TextOutput {
	return NewTextOutput(ST().StdErr)
}

func (w *TextOutput) Write(p []byte) (int, error) {
	if w == nil || w.proto == nil {
		return 0, errNilTextOutputProtocol
	}
	if len(p) == 0 {
		return 0, nil
	}

	buf := StringToCHAR16Z(string(p))
	status := w.proto.OutputString(&buf[0])
	if status != EFI_SUCCESS {
		return 0, StatusError(status)
	}
	return len(p), nil
}

func (w *TextOutput) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}
