package ili9341

type Rotation uint8

const (

	// register constants based on source:
	// https://github.com/adafruit/Adafruit_ILI9341/blob/master/Adafruit_ILI9341.h

	TFTWIDTH  = 240 ///< ILI9341 max TFT width
	TFTHEIGHT = 320 ///< ILI9341 max TFT height

	NOP     = 0x00 ///< No-op register
	SWRESET = 0x01 ///< Software reset register
	RDDID   = 0x04 ///< Read display identification information
	RDDST   = 0x09 ///< Read Display Status

	SLPIN  = 0x10 ///< Enter Sleep Mode
	SLPOUT = 0x11 ///< Sleep Out
	PTLON  = 0x12 ///< Partial Mode ON
	NORON  = 0x13 ///< Normal Display Mode ON

	RDMODE     = 0x0A ///< Read Display Power Mode
	RDMADCTL   = 0x0B ///< Read Display MADCTL
	RDPIXFMT   = 0x0C ///< Read Display Pixel Format
	RDIMGFMT   = 0x0D ///< Read Display Image Format
	RDSELFDIAG = 0x0F ///< Read Display Self-Diagnostic Result

	INVOFF   = 0x20 ///< Display Inversion OFF
	INVON    = 0x21 ///< Display Inversion ON
	GAMMASET = 0x26 ///< Gamma Set
	DISPOFF  = 0x28 ///< Display OFF
	DISPON   = 0x29 ///< Display ON

	CASET = 0x2A ///< Column Address Set
	PASET = 0x2B ///< Page Address Set
	RAMWR = 0x2C ///< Memory Write
	RAMRD = 0x2E ///< Memory Read

	PTLAR    = 0x30 ///< Partial Area
	VSCRDEF  = 0x33 ///< Vertical Scrolling Definition
	MADCTL   = 0x36 ///< Memory Access Control
	VSCRSADD = 0x37 ///< Vertical Scrolling Start Address
	PIXFMT   = 0x3A ///< COLMOD: Pixel Format Set

	FRMCTR1 = 0xB1 ///< Frame Rate Control (In Normal Mode/Full Colors)
	FRMCTR2 = 0xB2 ///< Frame Rate Control (In Idle Mode/8 colors)
	FRMCTR3 = 0xB3 ///< Frame Rate control (In Partial Mode/Full Colors)
	INVCTR  = 0xB4 ///< Display Inversion Control
	DFUNCTR = 0xB6 ///< Display Function Control

	PWCTR1 = 0xC0 ///< Power Control 1
	PWCTR2 = 0xC1 ///< Power Control 2
	PWCTR3 = 0xC2 ///< Power Control 3
	PWCTR4 = 0xC3 ///< Power Control 4
	PWCTR5 = 0xC4 ///< Power Control 5
	VMCTR1 = 0xC5 ///< VCOM Control 1
	VMCTR2 = 0xC7 ///< VCOM Control 2

	RDID1 = 0xDA ///< Read ID 1
	RDID2 = 0xDB ///< Read ID 2
	RDID3 = 0xDC ///< Read ID 3
	RDID4 = 0xDD ///< Read ID 4

	GMCTRP1 = 0xE0 ///< Positive Gamma Correction
	GMCTRN1 = 0xE1 ///< Negative Gamma Correction
	//PWCTR6     0xFC

	MADCTL_MY  = 0x80 ///< Bottom to top
	MADCTL_MX  = 0x40 ///< Right to left
	MADCTL_MV  = 0x20 ///< Reverse Mode
	MADCTL_ML  = 0x10 ///< LCD refresh Bottom to top
	MADCTL_RGB = 0x00 ///< Red-Green-Blue pixel order
	MADCTL_BGR = 0x08 ///< Blue-Green-Red pixel order
	MADCTL_MH  = 0x04 ///< LCD refresh right to left

)

const (
	Rotation0   Rotation = 0
	Rotation90  Rotation = 1 // 90 degrees clock-wise rotation
	Rotation180 Rotation = 2
	Rotation270 Rotation = 3
)
