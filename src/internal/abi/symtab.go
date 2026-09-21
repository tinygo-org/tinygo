package abi

// PCLnTabMagic is the version at the start of a Go PC-line table.
type PCLnTabMagic uint32

const (
	Go12PCLnTabMagic  PCLnTabMagic = 0xfffffffb
	Go116PCLnTabMagic PCLnTabMagic = 0xfffffffa
	Go118PCLnTabMagic PCLnTabMagic = 0xfffffff0
	Go120PCLnTabMagic PCLnTabMagic = 0xfffffff1

	CurrentPCLnTabMagic = Go120PCLnTabMagic
)
