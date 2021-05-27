// Hand created file. DO NOT DELETE.
// Type definitions, fields, and constants associated with the FLASH peripheral
// of the STM32H7x7 family of dual-core MCUs.
// These definitions are applicable to both the Cortex-M7 and Cortex-M4 cores.

// +build stm32h7x7

package stm32

import (
	"runtime/volatile"
	"unsafe"
)

var FLASH = (*FLASH_Type)(unsafe.Pointer(uintptr(0x52002000)))

// FLASH
type FLASH_Type struct {
	ACR       volatile.Register32 // FLASH access control register                            Address offset: 0x00
	KEYR1     volatile.Register32 // Flash Key Register for bank1                             Address offset: 0x04
	OPTKEYR   volatile.Register32 // Flash Option Key Register                                Address offset: 0x08
	CR1       volatile.Register32 // Flash Control Register for bank1                         Address offset: 0x0C
	SR1       volatile.Register32 // Flash Status Register for bank1                          Address offset: 0x10
	CCR1      volatile.Register32 // Flash Control Register for bank1                         Address offset: 0x14
	OPTCR     volatile.Register32 // Flash Option Control Register                            Address offset: 0x18
	OPTSR_CUR volatile.Register32 // Flash Option Status Current Register                     Address offset: 0x1C
	OPTSR_PRG volatile.Register32 // Flash Option Status to Program Register                  Address offset: 0x20
	OPTCCR    volatile.Register32 // Flash Option Clear Control Register                      Address offset: 0x24
	PRAR_CUR1 volatile.Register32 // Flash Current Protection Address Register for bank1      Address offset: 0x28
	PRAR_PRG1 volatile.Register32 // Flash Protection Address to Program Register for bank1   Address offset: 0x2C
	SCAR_CUR1 volatile.Register32 // Flash Current Secure Address Register for bank1          Address offset: 0x30
	SCAR_PRG1 volatile.Register32 // Flash Secure Address to Program Register for bank1       Address offset: 0x34
	WPSN_CUR1 volatile.Register32 // Flash Current Write Protection Register on bank1         Address offset: 0x38
	WPSN_PRG1 volatile.Register32 // Flash Write Protection to Program Register on bank1      Address offset: 0x3C
	BOOT7_CUR volatile.Register32 // Flash Current Boot Address for Pelican Core Register     Address offset: 0x40
	BOOT7_PRG volatile.Register32 // Flash Boot Address to Program for Pelican Core Register  Address offset: 0x44
	BOOT4_CUR volatile.Register32 // Flash Current Boot Address for M4 Core Register          Address offset: 0x48
	BOOT4_PRG volatile.Register32 // Flash Boot Address to Program for M4 Core Register       Address offset: 0x4C
	CRCCR1    volatile.Register32 // Flash CRC Control register For Bank1 Register            Address offset: 0x50
	CRCSADD1  volatile.Register32 // Flash CRC Start Address Register for Bank1               Address offset: 0x54
	CRCEADD1  volatile.Register32 // Flash CRC End Address Register for Bank1                 Address offset: 0x58
	CRCDATA   volatile.Register32 // Flash CRC Data Register for Bank1                        Address offset: 0x5C
	ECC_FA1   volatile.Register32 // Flash ECC Fail Address For Bank1 Register                Address offset: 0x60
	_         [160]byte           // Reserved, 0x64 to 0x100
	KEYR2     volatile.Register32 // Flash Key Register for bank2                             Address offset: 0x104
	_         volatile.Register32 // Reserved, 0x108
	CR2       volatile.Register32 // Flash Control Register for bank2                         Address offset: 0x10C
	SR2       volatile.Register32 // Flash Status Register for bank2                          Address offset: 0x110
	CCR2      volatile.Register32 // Flash Status Register for bank2                          Address offset: 0x114
	_         [16]byte            // Reserved, 0x118 to 0x124
	PRAR_CUR2 volatile.Register32 // Flash Current Protection Address Register for bank2      Address offset: 0x128
	PRAR_PRG2 volatile.Register32 // Flash Protection Address to Program Register for bank2   Address offset: 0x12C
	SCAR_CUR2 volatile.Register32 // Flash Current Secure Address Register for bank2          Address offset: 0x130
	SCAR_PRG2 volatile.Register32 // Flash Secure Address Register for bank2                  Address offset: 0x134
	WPSN_CUR2 volatile.Register32 // Flash Current Write Protection Register on bank2         Address offset: 0x138
	WPSN_PRG2 volatile.Register32 // Flash Write Protection to Program Register on bank2      Address offset: 0x13C
	_         [16]byte            // Reserved, 0x140 to 0x14C
	CRCCR2    volatile.Register32 // Flash CRC Control register For Bank2 Register            Address offset: 0x150
	CRCSADD2  volatile.Register32 // Flash CRC Start Address Register for Bank2               Address offset: 0x154
	CRCEADD2  volatile.Register32 // Flash CRC End Address Register for Bank2                 Address offset: 0x158
	CRCDATA2  volatile.Register32 // Flash CRC Data Register for Bank2                        Address offset: 0x15C
	ECC_FA2   volatile.Register32 // Flash ECC Fail Address For Bank2 Register                Address offset: 0x160
}

const (
	FLASH_SECTOR_TOTAL              = 8                     // 8 sectors
	FLASH_SIZE                      = 0x200000              // 2 MB
	FLASH_BANK_SIZE                 = FLASH_SIZE >> 1       // 1 MB
	FLASH_SECTOR_SIZE               = 0x00020000            // 128 KB
	FLASH_LATENCY_DEFAULT           = FLASH_ACR_LATENCY_7WS // FLASH Seven Latency cycles
	FLASH_NB_32BITWORD_IN_FLASHWORD = 8                     // 256 bits
	DUAL_BANK                       = 0x1                   // Dual-bank Flash

	FLASH_ACR_LATENCY_Pos    = 0
	FLASH_ACR_LATENCY_Msk    = 0xF << FLASH_ACR_LATENCY_Pos // 0x0000000F
	FLASH_ACR_LATENCY        = FLASH_ACR_LATENCY_Msk        // Read Latency
	FLASH_ACR_LATENCY_0WS    = 0x00000000
	FLASH_ACR_LATENCY_1WS    = 0x00000001
	FLASH_ACR_LATENCY_2WS    = 0x00000002
	FLASH_ACR_LATENCY_3WS    = 0x00000003
	FLASH_ACR_LATENCY_4WS    = 0x00000004
	FLASH_ACR_LATENCY_5WS    = 0x00000005
	FLASH_ACR_LATENCY_6WS    = 0x00000006
	FLASH_ACR_LATENCY_7WS    = 0x00000007
	FLASH_ACR_LATENCY_8WS    = 0x00000008
	FLASH_ACR_LATENCY_9WS    = 0x00000009
	FLASH_ACR_LATENCY_10WS   = 0x0000000A
	FLASH_ACR_LATENCY_11WS   = 0x0000000B
	FLASH_ACR_LATENCY_12WS   = 0x0000000C
	FLASH_ACR_LATENCY_13WS   = 0x0000000D
	FLASH_ACR_LATENCY_14WS   = 0x0000000E
	FLASH_ACR_LATENCY_15WS   = 0x0000000F
	FLASH_ACR_WRHIGHFREQ_Pos = 4
	FLASH_ACR_WRHIGHFREQ_Msk = 0x3 << FLASH_ACR_WRHIGHFREQ_Pos // 0x00000030
	FLASH_ACR_WRHIGHFREQ     = FLASH_ACR_WRHIGHFREQ_Msk        // Flash signal delay
	FLASH_ACR_WRHIGHFREQ_0   = 0x1 << FLASH_ACR_WRHIGHFREQ_Pos // 0x00000010
	FLASH_ACR_WRHIGHFREQ_1   = 0x2 << FLASH_ACR_WRHIGHFREQ_Pos // 0x00000020

	FLASH_CR_LOCK_Pos       = 0
	FLASH_CR_LOCK_Msk       = 0x1 << FLASH_CR_LOCK_Pos // 0x00000001
	FLASH_CR_LOCK           = FLASH_CR_LOCK_Msk        // Configuration lock bit
	FLASH_CR_PG_Pos         = 1
	FLASH_CR_PG_Msk         = 0x1 << FLASH_CR_PG_Pos // 0x00000002
	FLASH_CR_PG             = FLASH_CR_PG_Msk        // Internal buffer control bit
	FLASH_CR_SER_Pos        = 2
	FLASH_CR_SER_Msk        = 0x1 << FLASH_CR_SER_Pos // 0x00000004
	FLASH_CR_SER            = FLASH_CR_SER_Msk        // Sector erase request
	FLASH_CR_BER_Pos        = 3
	FLASH_CR_BER_Msk        = 0x1 << FLASH_CR_BER_Pos // 0x00000008
	FLASH_CR_BER            = FLASH_CR_BER_Msk        // Bank erase request
	FLASH_CR_PSIZE_Pos      = 4
	FLASH_CR_PSIZE_Msk      = 0x3 << FLASH_CR_PSIZE_Pos // 0x00000030
	FLASH_CR_PSIZE          = FLASH_CR_PSIZE_Msk        // Program size
	FLASH_CR_PSIZE_0        = 0x1 << FLASH_CR_PSIZE_Pos // 0x00000010
	FLASH_CR_PSIZE_1        = 0x2 << FLASH_CR_PSIZE_Pos // 0x00000020
	FLASH_CR_FW_Pos         = 6
	FLASH_CR_FW_Msk         = 0x1 << FLASH_CR_FW_Pos // 0x00000040
	FLASH_CR_FW             = FLASH_CR_FW_Msk        // Write forcing control bit
	FLASH_CR_START_Pos      = 7
	FLASH_CR_START_Msk      = 0x1 << FLASH_CR_START_Pos // 0x00000080
	FLASH_CR_START          = FLASH_CR_START_Msk        // Erase start control bit
	FLASH_CR_SNB_Pos        = 8
	FLASH_CR_SNB_Msk        = 0x7 << FLASH_CR_SNB_Pos // 0x00000700
	FLASH_CR_SNB            = FLASH_CR_SNB_Msk        // Sector erase selection number
	FLASH_CR_SNB_0          = 0x1 << FLASH_CR_SNB_Pos // 0x00000100
	FLASH_CR_SNB_1          = 0x2 << FLASH_CR_SNB_Pos // 0x00000200
	FLASH_CR_SNB_2          = 0x4 << FLASH_CR_SNB_Pos // 0x00000400
	FLASH_CR_CRC_EN_Pos     = 15
	FLASH_CR_CRC_EN_Msk     = 0x1 << FLASH_CR_CRC_EN_Pos // 0x00008000
	FLASH_CR_CRC_EN         = FLASH_CR_CRC_EN_Msk        // CRC control bit
	FLASH_CR_EOPIE_Pos      = 16
	FLASH_CR_EOPIE_Msk      = 0x1 << FLASH_CR_EOPIE_Pos // 0x00010000
	FLASH_CR_EOPIE          = FLASH_CR_EOPIE_Msk        // End-of-program interrupt control bit
	FLASH_CR_WRPERRIE_Pos   = 17
	FLASH_CR_WRPERRIE_Msk   = 0x1 << FLASH_CR_WRPERRIE_Pos // 0x00020000
	FLASH_CR_WRPERRIE       = FLASH_CR_WRPERRIE_Msk        // Write protection error interrupt enable bit
	FLASH_CR_PGSERRIE_Pos   = 18
	FLASH_CR_PGSERRIE_Msk   = 0x1 << FLASH_CR_PGSERRIE_Pos // 0x00040000
	FLASH_CR_PGSERRIE       = FLASH_CR_PGSERRIE_Msk        // Programming sequence error interrupt enable bit
	FLASH_CR_STRBERRIE_Pos  = 19
	FLASH_CR_STRBERRIE_Msk  = 0x1 << FLASH_CR_STRBERRIE_Pos // 0x00080000
	FLASH_CR_STRBERRIE      = FLASH_CR_STRBERRIE_Msk        // Strobe error interrupt enable bit
	FLASH_CR_INCERRIE_Pos   = 21
	FLASH_CR_INCERRIE_Msk   = 0x1 << FLASH_CR_INCERRIE_Pos // 0x00200000
	FLASH_CR_INCERRIE       = FLASH_CR_INCERRIE_Msk        // Inconsistency error interrupt enable bit
	FLASH_CR_OPERRIE_Pos    = 22
	FLASH_CR_OPERRIE_Msk    = 0x1 << FLASH_CR_OPERRIE_Pos // 0x00400000
	FLASH_CR_OPERRIE        = FLASH_CR_OPERRIE_Msk        // Write/erase error interrupt enable bit
	FLASH_CR_RDPERRIE_Pos   = 23
	FLASH_CR_RDPERRIE_Msk   = 0x1 << FLASH_CR_RDPERRIE_Pos // 0x00800000
	FLASH_CR_RDPERRIE       = FLASH_CR_RDPERRIE_Msk        // Read protection error interrupt enable bit
	FLASH_CR_RDSERRIE_Pos   = 24
	FLASH_CR_RDSERRIE_Msk   = 0x1 << FLASH_CR_RDSERRIE_Pos // 0x01000000
	FLASH_CR_RDSERRIE       = FLASH_CR_RDSERRIE_Msk        // Secure error interrupt enable bit
	FLASH_CR_SNECCERRIE_Pos = 25
	FLASH_CR_SNECCERRIE_Msk = 0x1 << FLASH_CR_SNECCERRIE_Pos // 0x02000000
	FLASH_CR_SNECCERRIE     = FLASH_CR_SNECCERRIE_Msk        // ECC single correction error interrupt enable bit
	FLASH_CR_DBECCERRIE_Pos = 26
	FLASH_CR_DBECCERRIE_Msk = 0x1 << FLASH_CR_DBECCERRIE_Pos // 0x04000000
	FLASH_CR_DBECCERRIE     = FLASH_CR_DBECCERRIE_Msk        // ECC double detection error interrupt enable bit
	FLASH_CR_CRCENDIE_Pos   = 27
	FLASH_CR_CRCENDIE_Msk   = 0x1 << FLASH_CR_CRCENDIE_Pos // 0x08000000
	FLASH_CR_CRCENDIE       = FLASH_CR_CRCENDIE_Msk        // CRC end of calculation interrupt enable bit
	FLASH_CR_CRCRDERRIE_Pos = 28
	FLASH_CR_CRCRDERRIE_Msk = 0x1 << FLASH_CR_CRCRDERRIE_Pos // 0x10000000
	FLASH_CR_CRCRDERRIE     = FLASH_CR_CRCRDERRIE_Msk        // CRC read error interrupt enable bit

	FLASH_SR_BSY_Pos      = 0
	FLASH_SR_BSY_Msk      = 0x1 << FLASH_SR_BSY_Pos // 0x00000001
	FLASH_SR_BSY          = FLASH_SR_BSY_Msk        // Busy flag
	FLASH_SR_WBNE_Pos     = 1
	FLASH_SR_WBNE_Msk     = 0x1 << FLASH_SR_WBNE_Pos // 0x00000002
	FLASH_SR_WBNE         = FLASH_SR_WBNE_Msk        // Write buffer not empty flag
	FLASH_SR_QW_Pos       = 2
	FLASH_SR_QW_Msk       = 0x1 << FLASH_SR_QW_Pos // 0x00000004
	FLASH_SR_QW           = FLASH_SR_QW_Msk        // Wait queue flag
	FLASH_SR_CRC_BUSY_Pos = 3
	FLASH_SR_CRC_BUSY_Msk = 0x1 << FLASH_SR_CRC_BUSY_Pos // 0x00000008
	FLASH_SR_CRC_BUSY     = FLASH_SR_CRC_BUSY_Msk        // CRC busy flag
	FLASH_SR_EOP_Pos      = 16
	FLASH_SR_EOP_Msk      = 0x1 << FLASH_SR_EOP_Pos // 0x00010000
	FLASH_SR_EOP          = FLASH_SR_EOP_Msk        // End-of-program flag
	FLASH_SR_WRPERR_Pos   = 17
	FLASH_SR_WRPERR_Msk   = 0x1 << FLASH_SR_WRPERR_Pos // 0x00020000
	FLASH_SR_WRPERR       = FLASH_SR_WRPERR_Msk        // Write protection error flag
	FLASH_SR_PGSERR_Pos   = 18
	FLASH_SR_PGSERR_Msk   = 0x1 << FLASH_SR_PGSERR_Pos // 0x00040000
	FLASH_SR_PGSERR       = FLASH_SR_PGSERR_Msk        // Programming sequence error flag
	FLASH_SR_STRBERR_Pos  = 19
	FLASH_SR_STRBERR_Msk  = 0x1 << FLASH_SR_STRBERR_Pos // 0x00080000
	FLASH_SR_STRBERR      = FLASH_SR_STRBERR_Msk        // Strobe error flag
	FLASH_SR_INCERR_Pos   = 21
	FLASH_SR_INCERR_Msk   = 0x1 << FLASH_SR_INCERR_Pos // 0x00200000
	FLASH_SR_INCERR       = FLASH_SR_INCERR_Msk        // Inconsistency error flag
	FLASH_SR_OPERR_Pos    = 22
	FLASH_SR_OPERR_Msk    = 0x1 << FLASH_SR_OPERR_Pos // 0x00400000
	FLASH_SR_OPERR        = FLASH_SR_OPERR_Msk        // Write/erase error flag
	FLASH_SR_RDPERR_Pos   = 23
	FLASH_SR_RDPERR_Msk   = 0x1 << FLASH_SR_RDPERR_Pos // 0x00800000
	FLASH_SR_RDPERR       = FLASH_SR_RDPERR_Msk        // Read protection error flag
	FLASH_SR_RDSERR_Pos   = 24
	FLASH_SR_RDSERR_Msk   = 0x1 << FLASH_SR_RDSERR_Pos // 0x01000000
	FLASH_SR_RDSERR       = FLASH_SR_RDSERR_Msk        // Secure error flag
	FLASH_SR_SNECCERR_Pos = 25
	FLASH_SR_SNECCERR_Msk = 0x1 << FLASH_SR_SNECCERR_Pos // 0x02000000
	FLASH_SR_SNECCERR     = FLASH_SR_SNECCERR_Msk        // Single correction error flag
	FLASH_SR_DBECCERR_Pos = 26
	FLASH_SR_DBECCERR_Msk = 0x1 << FLASH_SR_DBECCERR_Pos // 0x04000000
	FLASH_SR_DBECCERR     = FLASH_SR_DBECCERR_Msk        // ECC double detection error flag
	FLASH_SR_CRCEND_Pos   = 27
	FLASH_SR_CRCEND_Msk   = 0x1 << FLASH_SR_CRCEND_Pos // 0x08000000
	FLASH_SR_CRCEND       = FLASH_SR_CRCEND_Msk        // CRC end of calculation flag
	FLASH_SR_CRCRDERR_Pos = 28
	FLASH_SR_CRCRDERR_Msk = 0x1 << FLASH_SR_CRCRDERR_Pos // 0x10000000
	FLASH_SR_CRCRDERR     = FLASH_SR_CRCRDERR_Msk        // CRC read error flag

	FLASH_CCR_CLR_EOP_Pos      = 16
	FLASH_CCR_CLR_EOP_Msk      = 0x1 << FLASH_CCR_CLR_EOP_Pos // 0x00010000
	FLASH_CCR_CLR_EOP          = FLASH_CCR_CLR_EOP_Msk        // EOP flag clear bit
	FLASH_CCR_CLR_WRPERR_Pos   = 17
	FLASH_CCR_CLR_WRPERR_Msk   = 0x1 << FLASH_CCR_CLR_WRPERR_Pos // 0x00020000
	FLASH_CCR_CLR_WRPERR       = FLASH_CCR_CLR_WRPERR_Msk        // WRPERR flag clear bit
	FLASH_CCR_CLR_PGSERR_Pos   = 18
	FLASH_CCR_CLR_PGSERR_Msk   = 0x1 << FLASH_CCR_CLR_PGSERR_Pos // 0x00040000
	FLASH_CCR_CLR_PGSERR       = FLASH_CCR_CLR_PGSERR_Msk        // PGSERR flag clear bit
	FLASH_CCR_CLR_STRBERR_Pos  = 19
	FLASH_CCR_CLR_STRBERR_Msk  = 0x1 << FLASH_CCR_CLR_STRBERR_Pos // 0x00080000
	FLASH_CCR_CLR_STRBERR      = FLASH_CCR_CLR_STRBERR_Msk        // STRBERR flag clear bit
	FLASH_CCR_CLR_INCERR_Pos   = 21
	FLASH_CCR_CLR_INCERR_Msk   = 0x1 << FLASH_CCR_CLR_INCERR_Pos // 0x00200000
	FLASH_CCR_CLR_INCERR       = FLASH_CCR_CLR_INCERR_Msk        // INCERR flag clear bit
	FLASH_CCR_CLR_OPERR_Pos    = 22
	FLASH_CCR_CLR_OPERR_Msk    = 0x1 << FLASH_CCR_CLR_OPERR_Pos // 0x00400000
	FLASH_CCR_CLR_OPERR        = FLASH_CCR_CLR_OPERR_Msk        // OPERR flag clear bit
	FLASH_CCR_CLR_RDPERR_Pos   = 23
	FLASH_CCR_CLR_RDPERR_Msk   = 0x1 << FLASH_CCR_CLR_RDPERR_Pos // 0x00800000
	FLASH_CCR_CLR_RDPERR       = FLASH_CCR_CLR_RDPERR_Msk        // RDPERR flag clear bit
	FLASH_CCR_CLR_RDSERR_Pos   = 24
	FLASH_CCR_CLR_RDSERR_Msk   = 0x1 << FLASH_CCR_CLR_RDSERR_Pos // 0x01000000
	FLASH_CCR_CLR_RDSERR       = FLASH_CCR_CLR_RDSERR_Msk        // RDSERR flag clear bit
	FLASH_CCR_CLR_SNECCERR_Pos = 25
	FLASH_CCR_CLR_SNECCERR_Msk = 0x1 << FLASH_CCR_CLR_SNECCERR_Pos // 0x02000000
	FLASH_CCR_CLR_SNECCERR     = FLASH_CCR_CLR_SNECCERR_Msk        // SNECCERR flag clear bit
	FLASH_CCR_CLR_DBECCERR_Pos = 26
	FLASH_CCR_CLR_DBECCERR_Msk = 0x1 << FLASH_CCR_CLR_DBECCERR_Pos // 0x04000000
	FLASH_CCR_CLR_DBECCERR     = FLASH_CCR_CLR_DBECCERR_Msk        // DBECCERR flag clear bit
	FLASH_CCR_CLR_CRCEND_Pos   = 27
	FLASH_CCR_CLR_CRCEND_Msk   = 0x1 << FLASH_CCR_CLR_CRCEND_Pos // 0x08000000
	FLASH_CCR_CLR_CRCEND       = FLASH_CCR_CLR_CRCEND_Msk        // CRCEND flag clear bit
	FLASH_CCR_CLR_CRCRDERR_Pos = 28
	FLASH_CCR_CLR_CRCRDERR_Msk = 0x1 << FLASH_CCR_CLR_CRCRDERR_Pos // 0x10000000
	FLASH_CCR_CLR_CRCRDERR     = FLASH_CCR_CLR_CRCRDERR_Msk        // CRCRDERR flag clear bit

	FLASH_OPTCR_OPTLOCK_Pos        = 0
	FLASH_OPTCR_OPTLOCK_Msk        = 0x1 << FLASH_OPTCR_OPTLOCK_Pos // 0x00000001
	FLASH_OPTCR_OPTLOCK            = FLASH_OPTCR_OPTLOCK_Msk        // FLASH_OPTCR lock option configuration bit
	FLASH_OPTCR_OPTSTART_Pos       = 1
	FLASH_OPTCR_OPTSTART_Msk       = 0x1 << FLASH_OPTCR_OPTSTART_Pos // 0x00000002
	FLASH_OPTCR_OPTSTART           = FLASH_OPTCR_OPTSTART_Msk        // Option byte start change option configuration bit
	FLASH_OPTCR_MER_Pos            = 4
	FLASH_OPTCR_MER_Msk            = 0x1 << FLASH_OPTCR_MER_Pos // 0x00000010
	FLASH_OPTCR_MER                = FLASH_OPTCR_MER_Msk        // Mass erase request
	FLASH_OPTCR_OPTCHANGEERRIE_Pos = 30
	FLASH_OPTCR_OPTCHANGEERRIE_Msk = 0x1 << FLASH_OPTCR_OPTCHANGEERRIE_Pos // 0x40000000
	FLASH_OPTCR_OPTCHANGEERRIE     = FLASH_OPTCR_OPTCHANGEERRIE_Msk        // Option byte change error interrupt enable bit
	FLASH_OPTCR_SWAP_BANK_Pos      = 31
	FLASH_OPTCR_SWAP_BANK_Msk      = 0x1 << FLASH_OPTCR_SWAP_BANK_Pos // 0x80000000
	FLASH_OPTCR_SWAP_BANK          = FLASH_OPTCR_SWAP_BANK_Msk        // Bank swapping option configuration bit

	FLASH_OPTSR_OPT_BUSY_Pos      = 0
	FLASH_OPTSR_OPT_BUSY_Msk      = 0x1 << FLASH_OPTSR_OPT_BUSY_Pos // 0x00000001
	FLASH_OPTSR_OPT_BUSY          = FLASH_OPTSR_OPT_BUSY_Msk        // Option byte change ongoing flag
	FLASH_OPTSR_BOR_LEV_Pos       = 2
	FLASH_OPTSR_BOR_LEV_Msk       = 0x3 << FLASH_OPTSR_BOR_LEV_Pos // 0x0000000C
	FLASH_OPTSR_BOR_LEV           = FLASH_OPTSR_BOR_LEV_Msk        // Brownout level option status bit
	FLASH_OPTSR_BOR_LEV_0         = 0x1 << FLASH_OPTSR_BOR_LEV_Pos // 0x00000004
	FLASH_OPTSR_BOR_LEV_1         = 0x2 << FLASH_OPTSR_BOR_LEV_Pos // 0x00000008
	FLASH_OPTSR_IWDG1_SW_Pos      = 4
	FLASH_OPTSR_IWDG1_SW_Msk      = 0x1 << FLASH_OPTSR_IWDG1_SW_Pos // 0x00000010
	FLASH_OPTSR_IWDG1_SW          = FLASH_OPTSR_IWDG1_SW_Msk        // IWDG1 control mode option status bit
	FLASH_OPTSR_IWDG2_SW_Pos      = 5
	FLASH_OPTSR_IWDG2_SW_Msk      = 0x1 << FLASH_OPTSR_IWDG2_SW_Pos // 0x00000020
	FLASH_OPTSR_IWDG2_SW          = FLASH_OPTSR_IWDG2_SW_Msk        // IWDG2 control mode option status bit
	FLASH_OPTSR_NRST_STOP_D1_Pos  = 6
	FLASH_OPTSR_NRST_STOP_D1_Msk  = 0x1 << FLASH_OPTSR_NRST_STOP_D1_Pos // 0x00000040
	FLASH_OPTSR_NRST_STOP_D1      = FLASH_OPTSR_NRST_STOP_D1_Msk        // D1 domain DStop entry reset option status bit
	FLASH_OPTSR_NRST_STBY_D1_Pos  = 7
	FLASH_OPTSR_NRST_STBY_D1_Msk  = 0x1 << FLASH_OPTSR_NRST_STBY_D1_Pos // 0x00000080
	FLASH_OPTSR_NRST_STBY_D1      = FLASH_OPTSR_NRST_STBY_D1_Msk        // D1 domain DStandby entry reset option status bit
	FLASH_OPTSR_RDP_Pos           = 8
	FLASH_OPTSR_RDP_Msk           = 0xFF << FLASH_OPTSR_RDP_Pos // 0x0000FF00
	FLASH_OPTSR_RDP               = FLASH_OPTSR_RDP_Msk         // Readout protection level option status byte
	FLASH_OPTSR_FZ_IWDG_STOP_Pos  = 17
	FLASH_OPTSR_FZ_IWDG_STOP_Msk  = 0x1 << FLASH_OPTSR_FZ_IWDG_STOP_Pos // 0x00020000
	FLASH_OPTSR_FZ_IWDG_STOP      = FLASH_OPTSR_FZ_IWDG_STOP_Msk        // IWDG Stop mode freeze option status bit
	FLASH_OPTSR_FZ_IWDG_SDBY_Pos  = 18
	FLASH_OPTSR_FZ_IWDG_SDBY_Msk  = 0x1 << FLASH_OPTSR_FZ_IWDG_SDBY_Pos // 0x00040000
	FLASH_OPTSR_FZ_IWDG_SDBY      = FLASH_OPTSR_FZ_IWDG_SDBY_Msk        // IWDG Standby mode freeze option status bit
	FLASH_OPTSR_ST_RAM_SIZE_Pos   = 19
	FLASH_OPTSR_ST_RAM_SIZE_Msk   = 0x3 << FLASH_OPTSR_ST_RAM_SIZE_Pos // 0x00180000
	FLASH_OPTSR_ST_RAM_SIZE       = FLASH_OPTSR_ST_RAM_SIZE_Msk        // ST RAM size option status
	FLASH_OPTSR_ST_RAM_SIZE_0     = 0x1 << FLASH_OPTSR_ST_RAM_SIZE_Pos // 0x00080000
	FLASH_OPTSR_ST_RAM_SIZE_1     = 0x2 << FLASH_OPTSR_ST_RAM_SIZE_Pos // 0x00100000
	FLASH_OPTSR_SECURITY_Pos      = 21
	FLASH_OPTSR_SECURITY_Msk      = 0x1 << FLASH_OPTSR_SECURITY_Pos // 0x00200000
	FLASH_OPTSR_SECURITY          = FLASH_OPTSR_SECURITY_Msk        // Security enable option status bit
	FLASH_OPTSR_BCM4_Pos          = 22
	FLASH_OPTSR_BCM4_Msk          = 0x1 << FLASH_OPTSR_BCM4_Pos // 0x00400000
	FLASH_OPTSR_BCM4              = FLASH_OPTSR_BCM4_Msk        // Arm Cortex-M4 boot option status bit
	FLASH_OPTSR_BCM7_Pos          = 23
	FLASH_OPTSR_BCM7_Msk          = 0x1 << FLASH_OPTSR_BCM7_Pos // 0x00800000
	FLASH_OPTSR_BCM7              = FLASH_OPTSR_BCM7_Msk        // Arm Cortex-M7 boot option status bit
	FLASH_OPTSR_NRST_STOP_D2_Pos  = 24
	FLASH_OPTSR_NRST_STOP_D2_Msk  = 0x1 << FLASH_OPTSR_NRST_STOP_D2_Pos // 0x01000000
	FLASH_OPTSR_NRST_STOP_D2      = FLASH_OPTSR_NRST_STOP_D2_Msk        // D2 domain DStop entry reset option status bit
	FLASH_OPTSR_NRST_STBY_D2_Pos  = 25
	FLASH_OPTSR_NRST_STBY_D2_Msk  = 0x1 << FLASH_OPTSR_NRST_STBY_D2_Pos // 0x02000000
	FLASH_OPTSR_NRST_STBY_D2      = FLASH_OPTSR_NRST_STBY_D2_Msk        // D2 domain DStandby entry reset option status bit
	FLASH_OPTSR_IO_HSLV_Pos       = 29
	FLASH_OPTSR_IO_HSLV_Msk       = 0x1 << FLASH_OPTSR_IO_HSLV_Pos // 0x20000000
	FLASH_OPTSR_IO_HSLV           = FLASH_OPTSR_IO_HSLV_Msk        // I/O high-speed at low-voltage status bit
	FLASH_OPTSR_OPTCHANGEERR_Pos  = 30
	FLASH_OPTSR_OPTCHANGEERR_Msk  = 0x1 << FLASH_OPTSR_OPTCHANGEERR_Pos // 0x40000000
	FLASH_OPTSR_OPTCHANGEERR      = FLASH_OPTSR_OPTCHANGEERR_Msk        // Option byte change error flag
	FLASH_OPTSR_SWAP_BANK_OPT_Pos = 31
	FLASH_OPTSR_SWAP_BANK_OPT_Msk = 0x1 << FLASH_OPTSR_SWAP_BANK_OPT_Pos // 0x80000000
	FLASH_OPTSR_SWAP_BANK_OPT     = FLASH_OPTSR_SWAP_BANK_OPT_Msk        // Bank swapping option status bit

	FLASH_OPTCCR_CLR_OPTCHANGEERR_Pos = 30
	FLASH_OPTCCR_CLR_OPTCHANGEERR_Msk = 0x1 << FLASH_OPTCCR_CLR_OPTCHANGEERR_Pos // 0x40000000
	FLASH_OPTCCR_CLR_OPTCHANGEERR     = FLASH_OPTCCR_CLR_OPTCHANGEERR_Msk        // OPTCHANGEERR reset bit

	FLASH_PRAR_PROT_AREA_START_Pos = 0
	FLASH_PRAR_PROT_AREA_START_Msk = 0xFFF << FLASH_PRAR_PROT_AREA_START_Pos // 0x00000FFF
	FLASH_PRAR_PROT_AREA_START     = FLASH_PRAR_PROT_AREA_START_Msk          // PCROP area start status bits
	FLASH_PRAR_PROT_AREA_END_Pos   = 16
	FLASH_PRAR_PROT_AREA_END_Msk   = 0xFFF << FLASH_PRAR_PROT_AREA_END_Pos // 0x0FFF0000
	FLASH_PRAR_PROT_AREA_END       = FLASH_PRAR_PROT_AREA_END_Msk          // PCROP area end status bits
	FLASH_PRAR_DMEP_Pos            = 31
	FLASH_PRAR_DMEP_Msk            = 0x1 << FLASH_PRAR_DMEP_Pos // 0x80000000
	FLASH_PRAR_DMEP                = FLASH_PRAR_DMEP_Msk        // PCROP protected erase enable option status bit

	FLASH_SCAR_SEC_AREA_START_Pos = 0
	FLASH_SCAR_SEC_AREA_START_Msk = 0xFFF << FLASH_SCAR_SEC_AREA_START_Pos // 0x00000FFF
	FLASH_SCAR_SEC_AREA_START     = FLASH_SCAR_SEC_AREA_START_Msk          // Secure-only area start status bits
	FLASH_SCAR_SEC_AREA_END_Pos   = 16
	FLASH_SCAR_SEC_AREA_END_Msk   = 0xFFF << FLASH_SCAR_SEC_AREA_END_Pos // 0x0FFF0000
	FLASH_SCAR_SEC_AREA_END       = FLASH_SCAR_SEC_AREA_END_Msk          // Secure-only area end status bits
	FLASH_SCAR_DMES_Pos           = 31
	FLASH_SCAR_DMES_Msk           = 0x1 << FLASH_SCAR_DMES_Pos // 0x80000000
	FLASH_SCAR_DMES               = FLASH_SCAR_DMES_Msk        // Secure access protected erase enable option status bit

	FLASH_WPSN_WRPSN_Pos = 0
	FLASH_WPSN_WRPSN_Msk = 0xFF << FLASH_WPSN_WRPSN_Pos // 0x000000FF
	FLASH_WPSN_WRPSN     = FLASH_WPSN_WRPSN_Msk         // Sector write protection option status byte

	FLASH_BOOT7_BCM7_ADD0_Pos = 0
	FLASH_BOOT7_BCM7_ADD0_Msk = 0xFFFF << FLASH_BOOT7_BCM7_ADD0_Pos // 0x0000FFFF
	FLASH_BOOT7_BCM7_ADD0     = FLASH_BOOT7_BCM7_ADD0_Msk           // Arm Cortex-M7 boot address 0
	FLASH_BOOT7_BCM7_ADD1_Pos = 16
	FLASH_BOOT7_BCM7_ADD1_Msk = 0xFFFF << FLASH_BOOT7_BCM7_ADD1_Pos // 0xFFFF0000
	FLASH_BOOT7_BCM7_ADD1     = FLASH_BOOT7_BCM7_ADD1_Msk           // Arm Cortex-M7 boot address 1

	FLASH_BOOT4_BCM4_ADD0_Pos = 0
	FLASH_BOOT4_BCM4_ADD0_Msk = 0xFFFF << FLASH_BOOT4_BCM4_ADD0_Pos // 0x0000FFFF
	FLASH_BOOT4_BCM4_ADD0     = FLASH_BOOT4_BCM4_ADD0_Msk           // Arm Cortex-M4 boot address 0
	FLASH_BOOT4_BCM4_ADD1_Pos = 16
	FLASH_BOOT4_BCM4_ADD1_Msk = 0xFFFF << FLASH_BOOT4_BCM4_ADD1_Pos // 0xFFFF0000
	FLASH_BOOT4_BCM4_ADD1     = FLASH_BOOT4_BCM4_ADD1_Msk           // Arm Cortex-M4 boot address 1

	FLASH_CRCCR_CRC_SECT_Pos    = 0
	FLASH_CRCCR_CRC_SECT_Msk    = 0x7 << FLASH_CRCCR_CRC_SECT_Pos // 0x00000007
	FLASH_CRCCR_CRC_SECT        = FLASH_CRCCR_CRC_SECT_Msk        // CRC sector number
	FLASH_CRCCR_CRC_BY_SECT_Pos = 8
	FLASH_CRCCR_CRC_BY_SECT_Msk = 0x1 << FLASH_CRCCR_CRC_BY_SECT_Pos // 0x00000100
	FLASH_CRCCR_CRC_BY_SECT     = FLASH_CRCCR_CRC_BY_SECT_Msk        // CRC sector mode select bit
	FLASH_CRCCR_ADD_SECT_Pos    = 9
	FLASH_CRCCR_ADD_SECT_Msk    = 0x1 << FLASH_CRCCR_ADD_SECT_Pos // 0x00000200
	FLASH_CRCCR_ADD_SECT        = FLASH_CRCCR_ADD_SECT_Msk        // CRC sector select bit
	FLASH_CRCCR_CLEAN_SECT_Pos  = 10
	FLASH_CRCCR_CLEAN_SECT_Msk  = 0x1 << FLASH_CRCCR_CLEAN_SECT_Pos // 0x00000400
	FLASH_CRCCR_CLEAN_SECT      = FLASH_CRCCR_CLEAN_SECT_Msk        // CRC sector list clear bit
	FLASH_CRCCR_START_CRC_Pos   = 16
	FLASH_CRCCR_START_CRC_Msk   = 0x1 << FLASH_CRCCR_START_CRC_Pos // 0x00010000
	FLASH_CRCCR_START_CRC       = FLASH_CRCCR_START_CRC_Msk        // CRC start bit
	FLASH_CRCCR_CLEAN_CRC_Pos   = 17
	FLASH_CRCCR_CLEAN_CRC_Msk   = 0x1 << FLASH_CRCCR_CLEAN_CRC_Pos // 0x00020000
	FLASH_CRCCR_CLEAN_CRC       = FLASH_CRCCR_CLEAN_CRC_Msk        // CRC clear bit
	FLASH_CRCCR_CRC_BURST_Pos   = 20
	FLASH_CRCCR_CRC_BURST_Msk   = 0x3 << FLASH_CRCCR_CRC_BURST_Pos // 0x00300000
	FLASH_CRCCR_CRC_BURST       = FLASH_CRCCR_CRC_BURST_Msk        // CRC burst size
	FLASH_CRCCR_CRC_BURST_0     = 0x1 << FLASH_CRCCR_CRC_BURST_Pos // 0x00100000
	FLASH_CRCCR_CRC_BURST_1     = 0x2 << FLASH_CRCCR_CRC_BURST_Pos // 0x00200000
	FLASH_CRCCR_ALL_BANK_Pos    = 22
	FLASH_CRCCR_ALL_BANK_Msk    = 0x1 << FLASH_CRCCR_ALL_BANK_Pos // 0x00400000
	FLASH_CRCCR_ALL_BANK        = FLASH_CRCCR_ALL_BANK_Msk        // CRC select bit

	FLASH_CRCSADD_CRC_START_ADDR_Pos = 0
	FLASH_CRCSADD_CRC_START_ADDR_Msk = 0xFFFFFFFF << FLASH_CRCSADD_CRC_START_ADDR_Pos // 0xFFFFFFFF
	FLASH_CRCSADD_CRC_START_ADDR     = FLASH_CRCSADD_CRC_START_ADDR_Msk               // CRC start address

	FLASH_CRCEADD_CRC_END_ADDR_Pos = 0
	FLASH_CRCEADD_CRC_END_ADDR_Msk = 0xFFFFFFFF << FLASH_CRCEADD_CRC_END_ADDR_Pos // 0xFFFFFFFF
	FLASH_CRCEADD_CRC_END_ADDR     = FLASH_CRCEADD_CRC_END_ADDR_Msk               // CRC end address

	FLASH_CRCDATA_CRC_DATA_Pos = 0
	FLASH_CRCDATA_CRC_DATA_Msk = 0xFFFFFFFF << FLASH_CRCDATA_CRC_DATA_Pos // 0xFFFFFFFF
	FLASH_CRCDATA_CRC_DATA     = FLASH_CRCDATA_CRC_DATA_Msk               // CRC result

	FLASH_ECC_FA_FAIL_ECC_ADDR_Pos = 0
	FLASH_ECC_FA_FAIL_ECC_ADDR_Msk = 0x7FFF << FLASH_ECC_FA_FAIL_ECC_ADDR_Pos // 0x00007FFF
	FLASH_ECC_FA_FAIL_ECC_ADDR     = FLASH_ECC_FA_FAIL_ECC_ADDR_Msk           // ECC error address
)
