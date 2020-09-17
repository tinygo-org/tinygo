// +build nxp,!mimxrt1062

package runtime

import "device/nxp"

const (
	CFSR_IACCVIOL    = nxp.SystemControl_CFSR_IACCVIOL
	CFSR_DACCVIOL    = nxp.SystemControl_CFSR_DACCVIOL
	CFSR_MUNSTKERR   = nxp.SystemControl_CFSR_MUNSTKERR
	CFSR_MSTKERR     = nxp.SystemControl_CFSR_MSTKERR
	CFSR_MLSPERR     = nxp.SystemControl_CFSR_MLSPERR
	CFSR_IBUSERR     = nxp.SystemControl_CFSR_IBUSERR
	CFSR_PRECISERR   = nxp.SystemControl_CFSR_PRECISERR
	CFSR_IMPRECISERR = nxp.SystemControl_CFSR_IMPRECISERR
	CFSR_UNSTKERR    = nxp.SystemControl_CFSR_UNSTKERR
	CFSR_STKERR      = nxp.SystemControl_CFSR_STKERR
	CFSR_LSPERR      = nxp.SystemControl_CFSR_LSPERR
	CFSR_UNDEFINSTR  = nxp.SystemControl_CFSR_UNDEFINSTR
	CFSR_INVSTATE    = nxp.SystemControl_CFSR_INVSTATE
	CFSR_INVPC       = nxp.SystemControl_CFSR_INVPC
	CFSR_NOCP        = nxp.SystemControl_CFSR_NOCP
	CFSR_UNALIGNED   = nxp.SystemControl_CFSR_UNALIGNED
	CFSR_DIVBYZERO   = nxp.SystemControl_CFSR_DIVBYZERO
	CFSR_MMARVALID   = nxp.SystemControl_CFSR_MMARVALID
	CFSR_BFARVALID   = nxp.SystemControl_CFSR_BFARVALID
)
