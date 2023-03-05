// Hand written file mostly derived from https://problemkaputt.de/gbatek.htm

//go:build gameboyadvance

package gba

import (
	"runtime/volatile"
	"unsafe"
)

// Interrupt numbers.
const (
	IRQ_VBLANK  = 0
	IRQ_HBLANK  = 1
	IRQ_VCOUNT  = 2
	IRQ_TIMER0  = 3
	IRQ_TIMER1  = 4
	IRQ_TIMER2  = 5
	IRQ_TIMER3  = 6
	IRQ_COM     = 7
	IRQ_DMA0    = 8
	IRQ_DMA1    = 9
	IRQ_DMA2    = 10
	IRQ_DMA3    = 11
	IRQ_KEYPAD  = 12
	IRQ_GAMEPAK = 13
)

// Peripherals
var (
	// Display registers
	DISP = (*DISP_Type)(unsafe.Add(unsafe.Pointer(REG_BASE), uintptr(0x0000)))

	// Background control registers
	BGCNT0 = (*BGCNT_Type)(unsafe.Add(unsafe.Pointer(REG_BASE), uintptr(0x0008)))
	BGCNT1 = (*BGCNT_Type)(unsafe.Add(unsafe.Pointer(REG_BASE), uintptr(0x000A)))
	BGCNT2 = (*BGCNT_Type)(unsafe.Add(unsafe.Pointer(REG_BASE), uintptr(0x000C)))
	BGCNT3 = (*BGCNT_Type)(unsafe.Add(unsafe.Pointer(REG_BASE), uintptr(0x000E)))

	BG0 = (*BG_Type)(unsafe.Add(unsafe.Pointer(REG_BASE), uintptr(0x0010)))
	BG1 = (*BG_Type)(unsafe.Add(unsafe.Pointer(REG_BASE), uintptr(0x0014)))
	BG2 = (*BG_Type)(unsafe.Add(unsafe.Pointer(REG_BASE), uintptr(0x0018)))
	BG3 = (*BG_Type)(unsafe.Add(unsafe.Pointer(REG_BASE), uintptr(0x001C)))

	BGA2 = (*BG_AFFINE_Type)(unsafe.Add(unsafe.Pointer(REG_BASE), uintptr(0x0020)))
	BGA3 = (*BG_AFFINE_Type)(unsafe.Add(unsafe.Pointer(REG_BASE), uintptr(0x0030)))

	WIN = (*WIN_Type)(unsafe.Add(unsafe.Pointer(REG_BASE), uintptr(0x0040)))

	GRAPHICS = (*GRAPHICS_Type)(unsafe.Add(unsafe.Pointer(REG_BASE), uintptr(0x004C)))

	// GBA Sound Channel 1 - Tone & Sweep
	SOUND1 = (*SOUND_Type)(unsafe.Add(unsafe.Pointer(REG_BASE), uintptr(0x0060)))

	// GBA Sound Channel 2 - Tone
	SOUND2 = (*SOUND_Type)(unsafe.Add(unsafe.Pointer(REG_BASE), uintptr(0x0068)))

	// TODO: Sound channel 3 and 4

	TM0 = (*TIMER_Type)(unsafe.Add(unsafe.Pointer(REG_BASE), uintptr(0x0100)))
	TM1 = (*TIMER_Type)(unsafe.Add(unsafe.Pointer(REG_BASE), uintptr(0x0104)))
	TM2 = (*TIMER_Type)(unsafe.Add(unsafe.Pointer(REG_BASE), uintptr(0x0108)))
	TM3 = (*TIMER_Type)(unsafe.Add(unsafe.Pointer(REG_BASE), uintptr(0x010C)))

	KEY = (*KEY_Type)(unsafe.Add(unsafe.Pointer(REG_BASE), uintptr(0x0130)))

	INTERRUPT = (*INTERRUPT_Type)(unsafe.Add(unsafe.Pointer(REG_BASE), uintptr(0x0200)))
)

// Main memory sections
const (
	// External work RAM
	MEM_EWRAM uintptr = 0x02000000

	// Internal work RAM
	MEM_IWRAM uintptr = 0x03000000

	// I/O registers
	MEM_IO uintptr = 0x04000000

	// Palette. Note: no 8bit write !!
	MEM_PAL uintptr = 0x05000000

	// Video RAM. Note: no 8bit write !!
	MEM_VRAM uintptr = 0x06000000

	// Object Attribute Memory (OAM) Note: no 8bit write !!
	MEM_OAM uintptr = 0x07000000

	// ROM. No write at all (duh)
	MEM_ROM uintptr = 0x08000000

	// Static RAM. 8bit write only
	MEM_SRAM uintptr = 0x0E000000
)

// Main section sizes
const (
	EWRAM_SIZE uintptr = 0x40000
	IWRAM_SIZE uintptr = 0x08000
	PAL_SIZE   uintptr = 0x00400
	VRAM_SIZE  uintptr = 0x18000
	OAM_SIZE   uintptr = 0x00400
	SRAM_SIZE  uintptr = 0x10000
)

// Sub section sizes
const (
	// BG palette size
	PAL_BG_SIZE = 0x00200

	// Object palette size
	PAL_OBJ_SIZE = 0x00200

	// Charblock size
	CBB_SIZE = 0x04000

	// Screenblock size
	SBB_SIZE = 0x00800

	// BG VRAM size
	VRAM_BG_SIZE = 0x10000

	// Object VRAM size
	VRAM_OBJ_SIZE = 0x08000

	// Mode 3 buffer size
	M3_SIZE = 0x12C00

	// Mode 4 buffer size
	M4_SIZE = 0x09600

	// Mode 5 buffer size
	M5_SIZE = 0x0A000

	// Bitmap page size
	VRAM_PAGE_SIZE = 0x0A000
)

// Sub sections
var (
	REG_BASE uintptr = MEM_IO

	// Background palette address
	MEM_PAL_BG = MEM_PAL

	// Object palette address
	MEM_PAL_OBJ = MEM_PAL + PAL_BG_SIZE

	// Front page address
	MEM_VRAM_FRONT = MEM_VRAM

	// Back page address
	MEM_VRAM_BACK = MEM_VRAM + VRAM_PAGE_SIZE

	// Object VRAM address
	MEM_VRAM_OBJ = MEM_VRAM + VRAM_BG_SIZE
)

// Display registers
type DISP_Type struct {
	DISPCNT  volatile.Register16
	_        [2]byte
	DISPSTAT volatile.Register16
	VCOUNT   volatile.Register16
}

// Background control registers
type BGCNT_Type struct {
	CNT volatile.Register16
}

// Regular background scroll registers. (write only!)
type BG_Type struct {
	HOFS volatile.Register16
	VOFS volatile.Register16
}

// Affine background parameters. (write only!)
type BG_AFFINE_Type struct {
	PA volatile.Register16
	PB volatile.Register16
	PC volatile.Register16
	PD volatile.Register16
	X  volatile.Register32
	Y  volatile.Register32
}

type WIN_Type struct {
	// win0 right, left (0xLLRR)
	WIN0H volatile.Register16

	// win1 right, left (0xLLRR)
	WIN1H volatile.Register16

	// win0 bottom, top (0xTTBB)
	WIN0V volatile.Register16

	// win1 bottom, top (0xTTBB)
	WIN1V volatile.Register16

	// win0, win1 control
	IN volatile.Register16

	// winOut, winObj control
	OUT volatile.Register16
}

type GRAPHICS_Type struct {
	// Mosaic control
	MOSAIC volatile.Register32

	// Alpha control
	BLDCNT volatile.Register16

	// Fade level
	BLDALPHA volatile.Register16

	// Blend levels
	BLDY volatile.Register16
}

type SOUND_Type struct {
	// Sweep register
	CNT_L volatile.Register16

	// Duty/Len/Envelope
	CNT_H volatile.Register16

	// Frequency/Control
	CNT_X volatile.Register16
}

// TODO: DMA

// TIMER
type TIMER_Type struct {
	DATA volatile.Register16
	CNT  volatile.Register16
}

// TODO: serial

// Keypad registers
type KEY_Type struct {
	INPUT volatile.Register16
	CNT   volatile.Register16
}

// TODO: Joybus communication

// Interrupt / System registers
type INTERRUPT_Type struct {
	IE      volatile.Register16
	IF      volatile.Register16
	WAITCNT volatile.Register16
	IME     volatile.Register16
	PAUSE   volatile.Register16
}

// Constants for DISP: display
const (
	// BGMODE: background mode.
	// Position of BGMODE field.
	DISPCNT_BGMODE_Pos = 0x0
	// Bit mask of BGMODE field.
	DISPCNT_BGMODE_Msk = 0x4
	// BG Mode 0.
	DISPCNT_BGMODE_0 = 0x0
	// BG Mode 1.
	DISPCNT_BGMODE_1 = 0x1
	// BG Mode 2.
	DISPCNT_BGMODE_2 = 0x2
	// BG Mode 3.
	DISPCNT_BGMODE_3 = 0x3
	// BG Mode 4.
	DISPCNT_BGMODE_4 = 0x4

	// FRAMESELECT: frame select (mode 4 and 5 only).
	DISPCNT_FRAMESELECT_Pos    = 0x4
	DISPCNT_FRAMESELECT_FRAME0 = 0x0
	DISPCNT_FRAMESELECT_FRAME1 = 0x1

	// HBLANKINTERVAL: 1=Allow access to OAM during H-Blank
	DISPCNT_HBLANKINTERVAL_Pos     = 0x5
	DISPCNT_HBLANKINTERVAL_NOALLOW = 0x0
	DISPCNT_HBLANKINTERVAL_ALLOW   = 0x1

	// OBJCHARVRAM: (0=Two dimensional, 1=One dimensional)
	DISPCNT_OBJCHARVRAM_Pos = 0x6
	DISPCNT_OBJCHARVRAM_2D  = 0x0
	DISPCNT_OBJCHARVRAM_1D  = 0x1

	// FORCEDBLANK: (1=Allow FAST access to VRAM,Palette,OAM)
	DISPCNT_FORCEDBLANK_Pos     = 0x7
	DISPCNT_FORCEDBLANK_NOALLOW = 0x0
	DISPCNT_FORCEDBLANK_ALLOW   = 0x1

	// Screen Display BG0
	DISPCNT_SCREENDISPLAY_BG0_Pos     = 0x8
	DISPCNT_SCREENDISPLAY_BG0_ENABLE  = 0x1
	DISPCNT_SCREENDISPLAY_BG0_DISABLE = 0x0

	// Screen Display BG1
	DISPCNT_SCREENDISPLAY_BG1_Pos     = 0x9
	DISPCNT_SCREENDISPLAY_BG1_ENABLE  = 0x1
	DISPCNT_SCREENDISPLAY_BG1_DISABLE = 0x0

	// Screen Display BG2
	DISPCNT_SCREENDISPLAY_BG2_Pos     = 0xA
	DISPCNT_SCREENDISPLAY_BG2_ENABLE  = 0x1
	DISPCNT_SCREENDISPLAY_BG2_DISABLE = 0x0

	// Screen Display BG3
	DISPCNT_SCREENDISPLAY_BG3_Pos     = 0xB
	DISPCNT_SCREENDISPLAY_BG3_ENABLE  = 0x1
	DISPCNT_SCREENDISPLAY_BG3_DISABLE = 0x0

	// Screen Display OBJ
	DISPCNT_SCREENDISPLAY_OBJ_Pos     = 0xC
	DISPCNT_SCREENDISPLAY_OBJ_ENABLE  = 0x1
	DISPCNT_SCREENDISPLAY_OBJ_DISABLE = 0x0

	// Window 0 Display Flag (0=Off, 1=On)
	DISPCNT_WINDOW0_DISPLAY_Pos     = 0xD
	DISPCNT_WINDOW0_DISPLAY_ENABLE  = 0x1
	DISPCNT_WINDOW0_DISPLAY_DISABLE = 0x0

	// Window 1 Display Flag (0=Off, 1=On)
	DISPCNT_WINDOW1_DISPLAY_Pos     = 0xE
	DISPCNT_WINDOW1_DISPLAY_ENABLE  = 0x1
	DISPCNT_WINDOW1_DISPLAY_DISABLE = 0x0

	// OBJ Window Display Flag
	DISPCNT_WINDOWOBJ_DISPLAY_Pos     = 0xF
	DISPCNT_WINDOWOBJ_DISPLAY_ENABLE  = 0x1
	DISPCNT_WINDOWOBJ_DISPLAY_DISABLE = 0x0

	// DISPSTAT: display status.
	// V-blank
	DISPSTAT_VBLANK_Pos     = 0x0
	DISPSTAT_VBLANK_ENABLE  = 0x1
	DISPSTAT_VBLANK_DISABLE = 0x0

	// H-blank
	DISPSTAT_HBLANK_Pos     = 0x1
	DISPSTAT_HBLANK_ENABLE  = 0x1
	DISPSTAT_HBLANK_DISABLE = 0x0

	// V-counter match
	DISPSTAT_VCOUNTER_Pos     = 0x2
	DISPSTAT_VCOUNTER_MATCH   = 0x1
	DISPSTAT_VCOUNTER_NOMATCH = 0x0

	// V-blank IRQ
	DISPSTAT_VBLANK_IRQ_Pos     = 0x3
	DISPSTAT_VBLANK_IRQ_ENABLE  = 0x1
	DISPSTAT_VBLANK_IRQ_DISABLE = 0x0

	// H-blank IRQ
	DISPSTAT_HBLANK_IRQ_Pos     = 0x4
	DISPSTAT_HBLANK_IRQ_ENABLE  = 0x1
	DISPSTAT_HBLANK_IRQ_DISABLE = 0x0

	// V-counter IRQ
	DISPSTAT_VCOUNTER_IRQ_Pos     = 0x5
	DISPSTAT_VCOUNTER_IRQ_ENABLE  = 0x1
	DISPSTAT_VCOUNTER_IRQ_DISABLE = 0x0

	// V-count setting
	DISPSTAT_VCOUNT_SETTING_Pos = 0x8
)

// Constants for TIMER
const (
	// PRESCALER: Prescaler Selection (0=F/1, 1=F/64, 2=F/256, 3=F/1024)
	// Position of PRESCALER field.
	TIMERCNT_PRESCALER_Pos = 0x0
	// Bit mask of PRESCALER field.
	TIMERCNT_PRESCALER_Msk = 0x2
	// 0=F/1
	TIMERCNT_PRESCALER_1 = 0x0
	// 1=F/64
	TIMERCNT_PRESCALER_64 = 0x1
	// 2=F/256
	TIMERCNT_PRESCALER_256 = 0x2
	// F/1024
	TIMERCNT_PRESCALER_1024 = 0x3

	// COUNTUP: Count-up Timing   (0=Normal, 1=See below)  ;Not used in TM0CNT_H
	// Position of COUNTUP_TIMING field.
	TIMERCNT_COUNTUP_TIMING_Pos     = 0x2
	TIMERCNT_COUNTUP_TIMING_NORMAL  = 0x0
	TIMERCNT_COUNTUP_TIMING_ENABLED = 0x1

	TIMERCNT_TIMER_IRQ_ENABLED_Pos = 0x06
	TIMERCNT_TIMER_IRQ_ENABLED     = 0x01
	TIMERCNT_TIMER_IRQ_DISABLED    = 0x00

	TIMERCNT_TIMER_STARTSTOP_Pos = 0x07
	TIMERCNT_TIMER_START         = 0x1
	TIMERCNT_TIMER_STOP          = 0x0
)

// Constants for KEY
const (
	// KEYINPUT
	KEYINPUT_PRESSED           = 0x0
	KEYINPUT_RELEASED          = 0x1
	KEYINPUT_BUTTON_A_Pos      = 0x0
	KEYINPUT_BUTTON_B_Pos      = 0x1
	KEYINPUT_BUTTON_SELECT_Pos = 0x2
	KEYINPUT_BUTTON_START_Pos  = 0x3
	KEYINPUT_BUTTON_RIGHT_Pos  = 0x4
	KEYINPUT_BUTTON_LEFT_Pos   = 0x5
	KEYINPUT_BUTTON_UP_Pos     = 0x6
	KEYINPUT_BUTTON_DOWN_Pos   = 0x7
	KEYINPUT_BUTTON_R_Pos      = 0x8
	KEYINPUT_BUTTON_L_Pos      = 0x9

	// KEYCNT
	KEYCNT_IGNORE            = 0x0
	KEYCNT_SELECT            = 0x1
	KEYCNT_BUTTON_A_Pos      = 0x0
	KEYCNT_BUTTON_B_Pos      = 0x1
	KEYCNT_BUTTON_SELECT_Pos = 0x2
	KEYCNT_BUTTON_START_Pos  = 0x3
	KEYCNT_BUTTON_RIGHT_Pos  = 0x4
	KEYCNT_BUTTON_LEFT_Pos   = 0x5
	KEYCNT_BUTTON_UP_Pos     = 0x6
	KEYCNT_BUTTON_DOWN_Pos   = 0x7
	KEYCNT_BUTTON_R_Pos      = 0x8
	KEYCNT_BUTTON_L_Pos      = 0x9
)

// Constants for INTERRUPT
const (
	// IE
	INTERRUPT_IE_ENABLED             = 0x1
	INTERRUPT_IE_DISABLED            = 0x0
	INTERRUPT_IE_VBLANK_Pos          = 0x0
	INTERRUPT_IE_HBLANK_Pos          = 0x1
	INTERRUPT_IE_VCOUNTER_MATCH_Pos  = 0x2
	INTERRUPT_IE_TIMER0_OVERFLOW_Pos = 0x3
	INTERRUPT_IE_TIMER1_OVERFLOW_Pos = 0x4
	INTERRUPT_IE_TIMER2_OVERFLOW_Pos = 0x5
	INTERRUPT_IE_TIMER3_OVERFLOW_Pos = 0x6
	INTERRUPT_IE_SERIAL_Pos          = 0x7
	INTERRUPT_IE_DMA0_Pos            = 0x8
	INTERRUPT_IE_DMA1_Pos            = 0x9
	INTERRUPT_IE_DMA2_Pos            = 0xA
	INTERRUPT_IE_DMA3_Pos            = 0xB
	INTERRUPT_IE_KEYPAD_Pos          = 0xC
	INTERRUPT_IE_GAMPAK_Pos          = 0xD

	// IF
	INTERRUPT_IF_ENABLED             = 0x1
	INTERRUPT_IF_DISABLED            = 0x0
	INTERRUPT_IF_VBLANK_Pos          = 0x0
	INTERRUPT_IF_HBLANK_Pos          = 0x1
	INTERRUPT_IF_VCOUNTER_MATCH_Pos  = 0x2
	INTERRUPT_IF_TIMER0_OVERFLOW_Pos = 0x3
	INTERRUPT_IF_TIMER1_OVERFLOW_Pos = 0x4
	INTERRUPT_IF_TIMER2_OVERFLOW_Pos = 0x5
	INTERRUPT_IF_TIMER3_OVERFLOW_Pos = 0x6
	INTERRUPT_IF_SERIAL_Pos          = 0x7
	INTERRUPT_IF_DMA0_Pos            = 0x8
	INTERRUPT_IF_DMA1_Pos            = 0x9
	INTERRUPT_IF_DMA2_Pos            = 0xA
	INTERRUPT_IF_DMA3_Pos            = 0xB
	INTERRUPT_IF_KEYPAD_Pos          = 0xC
	INTERRUPT_IF_GAMPAK_Pos          = 0xD
)
