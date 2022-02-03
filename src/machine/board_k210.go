//go:build maixbit
// +build maixbit

// Chip datasheet: https://s3.cn-north-1.amazonaws.com.cn/dl.kendryte.com/documents/kendryte_datasheet_20181011163248_en.pdf

package machine

// K210 IO pins.
const (
	P00 Pin = 0
	P01 Pin = 1
	P02 Pin = 2
	P03 Pin = 3
	P04 Pin = 4
	P05 Pin = 5
	P06 Pin = 6
	P07 Pin = 7
	P08 Pin = 8
	P09 Pin = 9
	P10 Pin = 10
	P11 Pin = 11
	P12 Pin = 12
	P13 Pin = 13
	P14 Pin = 14
	P15 Pin = 15
	P16 Pin = 16
	P17 Pin = 17
	P18 Pin = 18
	P19 Pin = 19
	P20 Pin = 20
	P21 Pin = 21
	P22 Pin = 22
	P23 Pin = 23
	P24 Pin = 24
	P25 Pin = 25
	P26 Pin = 26
	P27 Pin = 27
	P28 Pin = 28
	P29 Pin = 29
	P30 Pin = 30
	P31 Pin = 31
	P32 Pin = 32
	P33 Pin = 33
	P34 Pin = 34
	P35 Pin = 35
	P36 Pin = 36
	P37 Pin = 37
	P38 Pin = 38
	P39 Pin = 39
	P40 Pin = 40
	P41 Pin = 41
	P42 Pin = 42
	P43 Pin = 43
	P44 Pin = 44
	P45 Pin = 45
	P46 Pin = 46
	P47 Pin = 47
)

type FPIOAFunction uint8

// Every pin on the Kendryte K210 is assigned to an FPIOA function.
// Each pin can be configured with every function below.
const (
	FUNC_JTAG_TCLK           FPIOAFunction = 0   // JTAG Test Clock
	FUNC_JTAG_TDI            FPIOAFunction = 1   // JTAG Test Data In
	FUNC_JTAG_TMS            FPIOAFunction = 2   // JTAG Test Mode Select
	FUNC_JTAG_TDO            FPIOAFunction = 3   // JTAG Test Data Out
	FUNC_SPI0_D0             FPIOAFunction = 4   // SPI0 Data 0
	FUNC_SPI0_D1             FPIOAFunction = 5   // SPI0 Data 1
	FUNC_SPI0_D2             FPIOAFunction = 6   // SPI0 Data 2
	FUNC_SPI0_D3             FPIOAFunction = 7   // SPI0 Data 3
	FUNC_SPI0_D4             FPIOAFunction = 8   // SPI0 Data 4
	FUNC_SPI0_D5             FPIOAFunction = 9   // SPI0 Data 5
	FUNC_SPI0_D6             FPIOAFunction = 10  // SPI0 Data 6
	FUNC_SPI0_D7             FPIOAFunction = 11  // SPI0 Data 7
	FUNC_SPI0_SS0            FPIOAFunction = 12  // SPI0 Chip Select 0
	FUNC_SPI0_SS1            FPIOAFunction = 13  // SPI0 Chip Select 1
	FUNC_SPI0_SS2            FPIOAFunction = 14  // SPI0 Chip Select 2
	FUNC_SPI0_SS3            FPIOAFunction = 15  // SPI0 Chip Select 3
	FUNC_SPI0_ARB            FPIOAFunction = 16  // SPI0 Arbitration
	FUNC_SPI0_SCLK           FPIOAFunction = 17  // SPI0 Serial Clock
	FUNC_UARTHS_RX           FPIOAFunction = 18  // UART High speed Receiver
	FUNC_UARTHS_TX           FPIOAFunction = 19  // UART High speed Transmitter
	FUNC_RESV6               FPIOAFunction = 20  // Reserved function
	FUNC_RESV7               FPIOAFunction = 21  // Reserved function
	FUNC_CLK_SPI1            FPIOAFunction = 22  // Clock SPI1
	FUNC_CLK_I2C1            FPIOAFunction = 23  // Clock I2C1
	FUNC_GPIOHS0             FPIOAFunction = 24  // GPIO High speed 0
	FUNC_GPIOHS1             FPIOAFunction = 25  // GPIO High speed 1
	FUNC_GPIOHS2             FPIOAFunction = 26  // GPIO High speed 2
	FUNC_GPIOHS3             FPIOAFunction = 27  // GPIO High speed 3
	FUNC_GPIOHS4             FPIOAFunction = 28  // GPIO High speed 4
	FUNC_GPIOHS5             FPIOAFunction = 29  // GPIO High speed 5
	FUNC_GPIOHS6             FPIOAFunction = 30  // GPIO High speed 6
	FUNC_GPIOHS7             FPIOAFunction = 31  // GPIO High speed 7
	FUNC_GPIOHS8             FPIOAFunction = 32  // GPIO High speed 8
	FUNC_GPIOHS9             FPIOAFunction = 33  // GPIO High speed 9
	FUNC_GPIOHS10            FPIOAFunction = 34  // GPIO High speed 10
	FUNC_GPIOHS11            FPIOAFunction = 35  // GPIO High speed 11
	FUNC_GPIOHS12            FPIOAFunction = 36  // GPIO High speed 12
	FUNC_GPIOHS13            FPIOAFunction = 37  // GPIO High speed 13
	FUNC_GPIOHS14            FPIOAFunction = 38  // GPIO High speed 14
	FUNC_GPIOHS15            FPIOAFunction = 39  // GPIO High speed 15
	FUNC_GPIOHS16            FPIOAFunction = 40  // GPIO High speed 16
	FUNC_GPIOHS17            FPIOAFunction = 41  // GPIO High speed 17
	FUNC_GPIOHS18            FPIOAFunction = 42  // GPIO High speed 18
	FUNC_GPIOHS19            FPIOAFunction = 43  // GPIO High speed 19
	FUNC_GPIOHS20            FPIOAFunction = 44  // GPIO High speed 20
	FUNC_GPIOHS21            FPIOAFunction = 45  // GPIO High speed 21
	FUNC_GPIOHS22            FPIOAFunction = 46  // GPIO High speed 22
	FUNC_GPIOHS23            FPIOAFunction = 47  // GPIO High speed 23
	FUNC_GPIOHS24            FPIOAFunction = 48  // GPIO High speed 24
	FUNC_GPIOHS25            FPIOAFunction = 49  // GPIO High speed 25
	FUNC_GPIOHS26            FPIOAFunction = 50  // GPIO High speed 26
	FUNC_GPIOHS27            FPIOAFunction = 51  // GPIO High speed 27
	FUNC_GPIOHS28            FPIOAFunction = 52  // GPIO High speed 28
	FUNC_GPIOHS29            FPIOAFunction = 53  // GPIO High speed 29
	FUNC_GPIOHS30            FPIOAFunction = 54  // GPIO High speed 30
	FUNC_GPIOHS31            FPIOAFunction = 55  // GPIO High speed 31
	FUNC_GPIO0               FPIOAFunction = 56  // GPIO pin 0
	FUNC_GPIO1               FPIOAFunction = 57  // GPIO pin 1
	FUNC_GPIO2               FPIOAFunction = 58  // GPIO pin 2
	FUNC_GPIO3               FPIOAFunction = 59  // GPIO pin 3
	FUNC_GPIO4               FPIOAFunction = 60  // GPIO pin 4
	FUNC_GPIO5               FPIOAFunction = 61  // GPIO pin 5
	FUNC_GPIO6               FPIOAFunction = 62  // GPIO pin 6
	FUNC_GPIO7               FPIOAFunction = 63  // GPIO pin 7
	FUNC_UART1_RX            FPIOAFunction = 64  // UART1 Receiver
	FUNC_UART1_TX            FPIOAFunction = 65  // UART1 Transmitter
	FUNC_UART2_RX            FPIOAFunction = 66  // UART2 Receiver
	FUNC_UART2_TX            FPIOAFunction = 67  // UART2 Transmitter
	FUNC_UART3_RX            FPIOAFunction = 68  // UART3 Receiver
	FUNC_UART3_TX            FPIOAFunction = 69  // UART3 Transmitter
	FUNC_SPI1_D0             FPIOAFunction = 70  // SPI1 Data 0
	FUNC_SPI1_D1             FPIOAFunction = 71  // SPI1 Data 1
	FUNC_SPI1_D2             FPIOAFunction = 72  // SPI1 Data 2
	FUNC_SPI1_D3             FPIOAFunction = 73  // SPI1 Data 3
	FUNC_SPI1_D4             FPIOAFunction = 74  // SPI1 Data 4
	FUNC_SPI1_D5             FPIOAFunction = 75  // SPI1 Data 5
	FUNC_SPI1_D6             FPIOAFunction = 76  // SPI1 Data 6
	FUNC_SPI1_D7             FPIOAFunction = 77  // SPI1 Data 7
	FUNC_SPI1_SS0            FPIOAFunction = 78  // SPI1 Chip Select 0
	FUNC_SPI1_SS1            FPIOAFunction = 79  // SPI1 Chip Select 1
	FUNC_SPI1_SS2            FPIOAFunction = 80  // SPI1 Chip Select 2
	FUNC_SPI1_SS3            FPIOAFunction = 81  // SPI1 Chip Select 3
	FUNC_SPI1_ARB            FPIOAFunction = 82  // SPI1 Arbitration
	FUNC_SPI1_SCLK           FPIOAFunction = 83  // SPI1 Serial Clock
	FUNC_SPI_PERIPHERAL_D0   FPIOAFunction = 84  // SPI Peripheral Data 0
	FUNC_SPI_PERIPHERAL_SS   FPIOAFunction = 85  // SPI Peripheral Select
	FUNC_SPI_PERIPHERAL_SCLK FPIOAFunction = 86  // SPI Peripheral Serial Clock
	FUNC_I2S0_MCLK           FPIOAFunction = 87  // I2S0 Main Clock
	FUNC_I2S0_SCLK           FPIOAFunction = 88  // I2S0 Serial Clock(BCLK)
	FUNC_I2S0_WS             FPIOAFunction = 89  // I2S0 Word Select(LRCLK)
	FUNC_I2S0_IN_D0          FPIOAFunction = 90  // I2S0 Serial Data Input 0
	FUNC_I2S0_IN_D1          FPIOAFunction = 91  // I2S0 Serial Data Input 1
	FUNC_I2S0_IN_D2          FPIOAFunction = 92  // I2S0 Serial Data Input 2
	FUNC_I2S0_IN_D3          FPIOAFunction = 93  // I2S0 Serial Data Input 3
	FUNC_I2S0_OUT_D0         FPIOAFunction = 94  // I2S0 Serial Data Output 0
	FUNC_I2S0_OUT_D1         FPIOAFunction = 95  // I2S0 Serial Data Output 1
	FUNC_I2S0_OUT_D2         FPIOAFunction = 96  // I2S0 Serial Data Output 2
	FUNC_I2S0_OUT_D3         FPIOAFunction = 97  // I2S0 Serial Data Output 3
	FUNC_I2S1_MCLK           FPIOAFunction = 98  // I2S1 Main Clock
	FUNC_I2S1_SCLK           FPIOAFunction = 99  // I2S1 Serial Clock(BCLK)
	FUNC_I2S1_WS             FPIOAFunction = 100 // I2S1 Word Select(LRCLK)
	FUNC_I2S1_IN_D0          FPIOAFunction = 101 // I2S1 Serial Data Input 0
	FUNC_I2S1_IN_D1          FPIOAFunction = 102 // I2S1 Serial Data Input 1
	FUNC_I2S1_IN_D2          FPIOAFunction = 103 // I2S1 Serial Data Input 2
	FUNC_I2S1_IN_D3          FPIOAFunction = 104 // I2S1 Serial Data Input 3
	FUNC_I2S1_OUT_D0         FPIOAFunction = 105 // I2S1 Serial Data Output 0
	FUNC_I2S1_OUT_D1         FPIOAFunction = 106 // I2S1 Serial Data Output 1
	FUNC_I2S1_OUT_D2         FPIOAFunction = 107 // I2S1 Serial Data Output 2
	FUNC_I2S1_OUT_D3         FPIOAFunction = 108 // I2S1 Serial Data Output 3
	FUNC_I2S2_MCLK           FPIOAFunction = 109 // I2S2 Main Clock
	FUNC_I2S2_SCLK           FPIOAFunction = 110 // I2S2 Serial Clock(BCLK)
	FUNC_I2S2_WS             FPIOAFunction = 111 // I2S2 Word Select(LRCLK)
	FUNC_I2S2_IN_D0          FPIOAFunction = 112 // I2S2 Serial Data Input 0
	FUNC_I2S2_IN_D1          FPIOAFunction = 113 // I2S2 Serial Data Input 1
	FUNC_I2S2_IN_D2          FPIOAFunction = 114 // I2S2 Serial Data Input 2
	FUNC_I2S2_IN_D3          FPIOAFunction = 115 // I2S2 Serial Data Input 3
	FUNC_I2S2_OUT_D0         FPIOAFunction = 116 // I2S2 Serial Data Output 0
	FUNC_I2S2_OUT_D1         FPIOAFunction = 117 // I2S2 Serial Data Output 1
	FUNC_I2S2_OUT_D2         FPIOAFunction = 118 // I2S2 Serial Data Output 2
	FUNC_I2S2_OUT_D3         FPIOAFunction = 119 // I2S2 Serial Data Output 3
	FUNC_RESV0               FPIOAFunction = 120 // Reserved function
	FUNC_RESV1               FPIOAFunction = 121 // Reserved function
	FUNC_RESV2               FPIOAFunction = 122 // Reserved function
	FUNC_RESV3               FPIOAFunction = 123 // Reserved function
	FUNC_RESV4               FPIOAFunction = 124 // Reserved function
	FUNC_RESV5               FPIOAFunction = 125 // Reserved function
	FUNC_I2C0_SCLK           FPIOAFunction = 126 // I2C0 Serial Clock
	FUNC_I2C0_SDA            FPIOAFunction = 127 // I2C0 Serial Data
	FUNC_I2C1_SCLK           FPIOAFunction = 128 // I2C1 Serial Clock
	FUNC_I2C1_SDA            FPIOAFunction = 129 // I2C1 Serial Data
	FUNC_I2C2_SCLK           FPIOAFunction = 130 // I2C2 Serial Clock
	FUNC_I2C2_SDA            FPIOAFunction = 131 // I2C2 Serial Data
	FUNC_CMOS_XCLK           FPIOAFunction = 132 // DVP System Clock
	FUNC_CMOS_RST            FPIOAFunction = 133 // DVP System Reset
	FUNC_CMOS_PWDN           FPIOAFunction = 134 // DVP Power Down Mode
	FUNC_CMOS_VSYNC          FPIOAFunction = 135 // DVP Vertical Sync
	FUNC_CMOS_HREF           FPIOAFunction = 136 // DVP Horizontal Reference output
	FUNC_CMOS_PCLK           FPIOAFunction = 137 // Pixel Clock
	FUNC_CMOS_D0             FPIOAFunction = 138 // Data Bit 0
	FUNC_CMOS_D1             FPIOAFunction = 139 // Data Bit 1
	FUNC_CMOS_D2             FPIOAFunction = 140 // Data Bit 2
	FUNC_CMOS_D3             FPIOAFunction = 141 // Data Bit 3
	FUNC_CMOS_D4             FPIOAFunction = 142 // Data Bit 4
	FUNC_CMOS_D5             FPIOAFunction = 143 // Data Bit 5
	FUNC_CMOS_D6             FPIOAFunction = 144 // Data Bit 6
	FUNC_CMOS_D7             FPIOAFunction = 145 // Data Bit 7
	FUNC_SCCB_SCLK           FPIOAFunction = 146 // SCCB Serial Clock
	FUNC_SCCB_SDA            FPIOAFunction = 147 // SCCB Serial Data
	FUNC_UART1_CTS           FPIOAFunction = 148 // UART1 Clear To Send
	FUNC_UART1_DSR           FPIOAFunction = 149 // UART1 Data Set Ready
	FUNC_UART1_DCD           FPIOAFunction = 150 // UART1 Data Carrier Detect
	FUNC_UART1_RI            FPIOAFunction = 151 // UART1 Ring Indicator
	FUNC_UART1_SIR_IN        FPIOAFunction = 152 // UART1 Serial Infrared Input
	FUNC_UART1_DTR           FPIOAFunction = 153 // UART1 Data Terminal Ready
	FUNC_UART1_RTS           FPIOAFunction = 154 // UART1 Request To Send
	FUNC_UART1_OUT2          FPIOAFunction = 155 // UART1 User-designated Output 2
	FUNC_UART1_OUT1          FPIOAFunction = 156 // UART1 User-designated Output 1
	FUNC_UART1_SIR_OUT       FPIOAFunction = 157 // UART1 Serial Infrared Output
	FUNC_UART1_BAUD          FPIOAFunction = 158 // UART1 Transmit Clock Output
	FUNC_UART1_RE            FPIOAFunction = 159 // UART1 Receiver Output Enable
	FUNC_UART1_DE            FPIOAFunction = 160 // UART1 Driver Output Enable
	FUNC_UART1_RS485_EN      FPIOAFunction = 161 // UART1 RS485 Enable
	FUNC_UART2_CTS           FPIOAFunction = 162 // UART2 Clear To Send
	FUNC_UART2_DSR           FPIOAFunction = 163 // UART2 Data Set Ready
	FUNC_UART2_DCD           FPIOAFunction = 164 // UART2 Data Carrier Detect
	FUNC_UART2_RI            FPIOAFunction = 165 // UART2 Ring Indicator
	FUNC_UART2_SIR_IN        FPIOAFunction = 166 // UART2 Serial Infrared Input
	FUNC_UART2_DTR           FPIOAFunction = 167 // UART2 Data Terminal Ready
	FUNC_UART2_RTS           FPIOAFunction = 168 // UART2 Request To Send
	FUNC_UART2_OUT2          FPIOAFunction = 169 // UART2 User-designated Output 2
	FUNC_UART2_OUT1          FPIOAFunction = 170 // UART2 User-designated Output 1
	FUNC_UART2_SIR_OUT       FPIOAFunction = 171 // UART2 Serial Infrared Output
	FUNC_UART2_BAUD          FPIOAFunction = 172 // UART2 Transmit Clock Output
	FUNC_UART2_RE            FPIOAFunction = 173 // UART2 Receiver Output Enable
	FUNC_UART2_DE            FPIOAFunction = 174 // UART2 Driver Output Enable
	FUNC_UART2_RS485_EN      FPIOAFunction = 175 // UART2 RS485 Enable
	FUNC_UART3_CTS           FPIOAFunction = 176 // UART3 Clear To Send
	FUNC_UART3_DSR           FPIOAFunction = 177 // UART3 Data Set Ready
	FUNC_UART3_DCD           FPIOAFunction = 178 // UART3 Data Carrier Detect
	FUNC_UART3_RI            FPIOAFunction = 179 // UART3 Ring Indicator
	FUNC_UART3_SIR_IN        FPIOAFunction = 180 // UART3 Serial Infrared Input
	FUNC_UART3_DTR           FPIOAFunction = 181 // UART3 Data Terminal Ready
	FUNC_UART3_RTS           FPIOAFunction = 182 // UART3 Request To Send
	FUNC_UART3_OUT2          FPIOAFunction = 183 // UART3 User-designated Output 2
	FUNC_UART3_OUT1          FPIOAFunction = 184 // UART3 User-designated Output 1
	FUNC_UART3_SIR_OUT       FPIOAFunction = 185 // UART3 Serial Infrared Output
	FUNC_UART3_BAUD          FPIOAFunction = 186 // UART3 Transmit Clock Output
	FUNC_UART3_RE            FPIOAFunction = 187 // UART3 Receiver Output Enable
	FUNC_UART3_DE            FPIOAFunction = 188 // UART3 Driver Output Enable
	FUNC_UART3_RS485_EN      FPIOAFunction = 189 // UART3 RS485 Enable
	FUNC_TIMER0_TOGGLE1      FPIOAFunction = 190 // TIMER0 Toggle Output 1
	FUNC_TIMER0_TOGGLE2      FPIOAFunction = 191 // TIMER0 Toggle Output 2
	FUNC_TIMER0_TOGGLE3      FPIOAFunction = 192 // TIMER0 Toggle Output 3
	FUNC_TIMER0_TOGGLE4      FPIOAFunction = 193 // TIMER0 Toggle Output 4
	FUNC_TIMER1_TOGGLE1      FPIOAFunction = 194 // TIMER1 Toggle Output 1
	FUNC_TIMER1_TOGGLE2      FPIOAFunction = 195 // TIMER1 Toggle Output 2
	FUNC_TIMER1_TOGGLE3      FPIOAFunction = 196 // TIMER1 Toggle Output 3
	FUNC_TIMER1_TOGGLE4      FPIOAFunction = 197 // TIMER1 Toggle Output 4
	FUNC_TIMER2_TOGGLE1      FPIOAFunction = 198 // TIMER2 Toggle Output 1
	FUNC_TIMER2_TOGGLE2      FPIOAFunction = 199 // TIMER2 Toggle Output 2
	FUNC_TIMER2_TOGGLE3      FPIOAFunction = 200 // TIMER2 Toggle Output 3
	FUNC_TIMER2_TOGGLE4      FPIOAFunction = 201 // TIMER2 Toggle Output 4
	FUNC_CLK_SPI2            FPIOAFunction = 202 // Clock SPI2
	FUNC_CLK_I2C2            FPIOAFunction = 203 // Clock I2C2
	FUNC_INTERNAL0           FPIOAFunction = 204 // Internal function signal 0
	FUNC_INTERNAL1           FPIOAFunction = 205 // Internal function signal 1
	FUNC_INTERNAL2           FPIOAFunction = 206 // Internal function signal 2
	FUNC_INTERNAL3           FPIOAFunction = 207 // Internal function signal 3
	FUNC_INTERNAL4           FPIOAFunction = 208 // Internal function signal 4
	FUNC_INTERNAL5           FPIOAFunction = 209 // Internal function signal 5
	FUNC_INTERNAL6           FPIOAFunction = 210 // Internal function signal 6
	FUNC_INTERNAL7           FPIOAFunction = 211 // Internal function signal 7
	FUNC_INTERNAL8           FPIOAFunction = 212 // Internal function signal 8
	FUNC_INTERNAL9           FPIOAFunction = 213 // Internal function signal 9
	FUNC_INTERNAL10          FPIOAFunction = 214 // Internal function signal 10
	FUNC_INTERNAL11          FPIOAFunction = 215 // Internal function signal 11
	FUNC_INTERNAL12          FPIOAFunction = 216 // Internal function signal 12
	FUNC_INTERNAL13          FPIOAFunction = 217 // Internal function signal 13
	FUNC_INTERNAL14          FPIOAFunction = 218 // Internal function signal 14
	FUNC_INTERNAL15          FPIOAFunction = 219 // Internal function signal 15
	FUNC_INTERNAL16          FPIOAFunction = 220 // Internal function signal 16
	FUNC_INTERNAL17          FPIOAFunction = 221 // Internal function signal 17
	FUNC_CONSTANT            FPIOAFunction = 222 // Constant function
	FUNC_INTERNAL18          FPIOAFunction = 223 // Internal function signal 18
	FUNC_DEBUG0              FPIOAFunction = 224 // Debug function 0
	FUNC_DEBUG1              FPIOAFunction = 225 // Debug function 1
	FUNC_DEBUG2              FPIOAFunction = 226 // Debug function 2
	FUNC_DEBUG3              FPIOAFunction = 227 // Debug function 3
	FUNC_DEBUG4              FPIOAFunction = 228 // Debug function 4
	FUNC_DEBUG5              FPIOAFunction = 229 // Debug function 5
	FUNC_DEBUG6              FPIOAFunction = 230 // Debug function 6
	FUNC_DEBUG7              FPIOAFunction = 231 // Debug function 7
	FUNC_DEBUG8              FPIOAFunction = 232 // Debug function 8
	FUNC_DEBUG9              FPIOAFunction = 233 // Debug function 9
	FUNC_DEBUG10             FPIOAFunction = 234 // Debug function 10
	FUNC_DEBUG11             FPIOAFunction = 235 // Debug function 11
	FUNC_DEBUG12             FPIOAFunction = 236 // Debug function 12
	FUNC_DEBUG13             FPIOAFunction = 237 // Debug function 13
	FUNC_DEBUG14             FPIOAFunction = 238 // Debug function 14
	FUNC_DEBUG15             FPIOAFunction = 239 // Debug function 15
	FUNC_DEBUG16             FPIOAFunction = 240 // Debug function 16
	FUNC_DEBUG17             FPIOAFunction = 241 // Debug function 17
	FUNC_DEBUG18             FPIOAFunction = 242 // Debug function 18
	FUNC_DEBUG19             FPIOAFunction = 243 // Debug function 19
	FUNC_DEBUG20             FPIOAFunction = 244 // Debug function 20
	FUNC_DEBUG21             FPIOAFunction = 245 // Debug function 21
	FUNC_DEBUG22             FPIOAFunction = 246 // Debug function 22
	FUNC_DEBUG23             FPIOAFunction = 247 // Debug function 23
	FUNC_DEBUG24             FPIOAFunction = 248 // Debug function 24
	FUNC_DEBUG25             FPIOAFunction = 249 // Debug function 25
	FUNC_DEBUG26             FPIOAFunction = 250 // Debug function 26
	FUNC_DEBUG27             FPIOAFunction = 251 // Debug function 27
	FUNC_DEBUG28             FPIOAFunction = 252 // Debug function 28
	FUNC_DEBUG29             FPIOAFunction = 253 // Debug function 29
	FUNC_DEBUG30             FPIOAFunction = 254 // Debug function 30
	FUNC_DEBUG31             FPIOAFunction = 255 // Debug function 31
)

// These are the default FPIOA values for each function.
// (source: https://github.com/kendryte/kendryte-standalone-sdk/blob/develop/lib/drivers/fpioa.c#L69)
var fpioaFuncDefaults [256]uint32 = [256]uint32{
	0x00900000, 0x00900001, 0x00900002, 0x00001f03, 0x00b03f04, 0x00b03f05, 0x00b03f06, 0x00b03f07, 0x00b03f08,
	0x00b03f09, 0x00b03f0a, 0x00b03f0b, 0x00001f0c, 0x00001f0d, 0x00001f0e, 0x00001f0f, 0x03900010, 0x00001f11,
	0x00900012, 0x00001f13, 0x00900014, 0x00900015, 0x00001f16, 0x00001f17, 0x00901f18, 0x00901f19, 0x00901f1a,
	0x00901f1b, 0x00901f1c, 0x00901f1d, 0x00901f1e, 0x00901f1f, 0x00901f20, 0x00901f21, 0x00901f22, 0x00901f23,
	0x00901f24, 0x00901f25, 0x00901f26, 0x00901f27, 0x00901f28, 0x00901f29, 0x00901f2a, 0x00901f2b, 0x00901f2c,
	0x00901f2d, 0x00901f2e, 0x00901f2f, 0x00901f30, 0x00901f31, 0x00901f32, 0x00901f33, 0x00901f34, 0x00901f35,
	0x00901f36, 0x00901f37, 0x00901f38, 0x00901f39, 0x00901f3a, 0x00901f3b, 0x00901f3c, 0x00901f3d, 0x00901f3e,
	0x00901f3f, 0x00900040, 0x00001f41, 0x00900042, 0x00001f43, 0x00900044, 0x00001f45, 0x00b03f46, 0x00b03f47,
	0x00b03f48, 0x00b03f49, 0x00b03f4a, 0x00b03f4b, 0x00b03f4c, 0x00b03f4d, 0x00001f4e, 0x00001f4f, 0x00001f50,
	0x00001f51, 0x03900052, 0x00001f53, 0x00b03f54, 0x00900055, 0x00900056, 0x00001f57, 0x00001f58, 0x00001f59,
	0x0090005a, 0x0090005b, 0x0090005c, 0x0090005d, 0x00001f5e, 0x00001f5f, 0x00001f60, 0x00001f61, 0x00001f62,
	0x00001f63, 0x00001f64, 0x00900065, 0x00900066, 0x00900067, 0x00900068, 0x00001f69, 0x00001f6a, 0x00001f6b,
	0x00001f6c, 0x00001f6d, 0x00001f6e, 0x00001f6f, 0x00900070, 0x00900071, 0x00900072, 0x00900073, 0x00001f74,
	0x00001f75, 0x00001f76, 0x00001f77, 0x00000078, 0x00000079, 0x0000007a, 0x0000007b, 0x0000007c, 0x0000007d,
	0x0099107e, 0x0099107f, 0x00991080, 0x00991081, 0x00991082, 0x00991083, 0x00001f84, 0x00001f85, 0x00001f86,
	0x00900087, 0x00900088, 0x00900089, 0x0090008a, 0x0090008b, 0x0090008c, 0x0090008d, 0x0090008e, 0x0090008f,
	0x00900090, 0x00900091, 0x00993092, 0x00993093, 0x00900094, 0x00900095, 0x00900096, 0x00900097, 0x00900098,
	0x00001f99, 0x00001f9a, 0x00001f9b, 0x00001f9c, 0x00001f9d, 0x00001f9e, 0x00001f9f, 0x00001fa0, 0x00001fa1,
	0x009000a2, 0x009000a3, 0x009000a4, 0x009000a5, 0x009000a6, 0x00001fa7, 0x00001fa8, 0x00001fa9, 0x00001faa,
	0x00001fab, 0x00001fac, 0x00001fad, 0x00001fae, 0x00001faf, 0x009000b0, 0x009000b1, 0x009000b2, 0x009000b3,
	0x009000b4, 0x00001fb5, 0x00001fb6, 0x00001fb7, 0x00001fb8, 0x00001fb9, 0x00001fba, 0x00001fbb, 0x00001fbc,
	0x00001fbd, 0x00001fbe, 0x00001fbf, 0x00001fc0, 0x00001fc1, 0x00001fc2, 0x00001fc3, 0x00001fc4, 0x00001fc5,
	0x00001fc6, 0x00001fc7, 0x00001fc8, 0x00001fc9, 0x00001fca, 0x00001fcb, 0x00001fcc, 0x00001fcd, 0x00001fce,
	0x00001fcf, 0x00001fd0, 0x00001fd1, 0x00001fd2, 0x00001fd3, 0x00001fd4, 0x009000d5, 0x009000d6, 0x009000d7,
	0x009000d8, 0x009100d9, 0x00991fda, 0x009000db, 0x009000dc, 0x009000dd, 0x000000de, 0x009000df, 0x00001fe0,
	0x00001fe1, 0x00001fe2, 0x00001fe3, 0x00001fe4, 0x00001fe5, 0x00001fe6, 0x00001fe7, 0x00001fe8, 0x00001fe9,
	0x00001fea, 0x00001feb, 0x00001fec, 0x00001fed, 0x00001fee, 0x00001fef, 0x00001ff0, 0x00001ff1, 0x00001ff2,
	0x00001ff3, 0x00001ff4, 0x00001ff5, 0x00001ff6, 0x00001ff7, 0x00001ff8, 0x00001ff9, 0x00001ffa, 0x00001ffb,
	0x00001ffc, 0x00001ffd, 0x00001ffe, 0x00001fff,
}
