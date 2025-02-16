//go:build arduino_unor4

// This contains the pin mappings for the Arduino Uno R4 board.
//
// For more information, see: 
package machine

// GPIO Pins
const (
	D0 Pin = P3_01
	D1 Pin = P3_02
	D2 Pin = P1_04
	D3 Pin = P1_05
	D4 Pin = P1_06
	D5 Pin = P1_07
	D6 Pin = P1_11
	D7 Pin = P1_12
	D8 Pin = P3_04
	D9 Pin = P3_03
	D10 Pin = P1_03
	D11 Pin = P4_11
	D12 Pin = P4_10
	D13 Pin = P1_02
)

// Analog pins
const (
	A0 Pin = P0_14
	A1 Pin = P0_00
	A2 Pin = P0_01
	A3 Pin = P0_02
	A4 Pin = P1_01
	A5 Pin = P1_00
)

const (
	LED = D13
)


// UART1 pins
const (
	UART_TX_PIN Pin = P1_09
	UART_RX_PIN Pin = P1_10
)

// I2C pins
const (
	SDA_PIN Pin = A4
	SCL_PIN Pin = A5
)

// { BSP_IO_PORT_03_PIN_01,    P301   }, /* (0) D0  -------------------------  DIGITAL  */
// { BSP_IO_PORT_03_PIN_02,    P302   }, /* (1) D1  */
// { BSP_IO_PORT_01_PIN_04,    P104   }, /* (2) D2  */
// { BSP_IO_PORT_01_PIN_05,    P105   }, /* (3) D3~ */
// { BSP_IO_PORT_01_PIN_06,    P106   }, /* (4) D4  */
// { BSP_IO_PORT_01_PIN_07,    P107   }, /* (5) D5~ */
// { BSP_IO_PORT_01_PIN_11,    P111   }, /* (6) D6~ */
// { BSP_IO_PORT_01_PIN_12,    P112   }, /* (7) D7  */
// { BSP_IO_PORT_03_PIN_04,    P304   }, /* (8) D8  */
// { BSP_IO_PORT_03_PIN_03,    P303   }, /* (9) D9~  */
// { BSP_IO_PORT_01_PIN_03,    P103   }, /* (10) D10~ */
// { BSP_IO_PORT_04_PIN_11,    P411   }, /* (11) D11~ */
// { BSP_IO_PORT_04_PIN_10,    P410   }, /* (12) D12 */
// { BSP_IO_PORT_01_PIN_02,    P102   }, /* (13) D13 */
// { BSP_IO_PORT_00_PIN_14,    P014   }, /* (14) A0  --------------------------  ANALOG  */
// { BSP_IO_PORT_00_PIN_00,    P000   }, /* (15) A1  */
// { BSP_IO_PORT_00_PIN_01,    P001   }, /* (16) A2  */
// { BSP_IO_PORT_00_PIN_02,    P002   }, /* (17) A3  */
// { BSP_IO_PORT_01_PIN_01,    P101   }, /* (18) A4/SDA  */
// { BSP_IO_PORT_01_PIN_00,    P100   }, /* (19) A5/SCL  */

// { BSP_IO_PORT_05_PIN_00,    P500   }, /* (20) Analog voltage measure pin  */
// { BSP_IO_PORT_04_PIN_08,    P408   }, /* (21) USB switch, drive high for RA4  */

// { BSP_IO_PORT_01_PIN_09,    P109   }, /* (22) D22 ------------------------  TX */
// { BSP_IO_PORT_01_PIN_10,    P110   }, /* (23) D23 ------------------------  RX */
// { BSP_IO_PORT_05_PIN_01,    P501   }, /* (24) D24 ------------------------- TX WIFI */
// { BSP_IO_PORT_05_PIN_02,    P502   }, /* (25) D25 ------------------------- RX WIFI */

// { BSP_IO_PORT_04_PIN_00,    P400   }, /* (26) D26  QWIC SCL */
// { BSP_IO_PORT_04_PIN_01,    P401   }, /* (27) D27  QWIC SDA */

// { BSP_IO_PORT_00_PIN_03,    P003   }, /* (28) D28  */
// { BSP_IO_PORT_00_PIN_04,    P004   }, /* (29) D29  */
// { BSP_IO_PORT_00_PIN_11,    P011   }, /* (30) D30  */
// { BSP_IO_PORT_00_PIN_12,    P012   }, /* (31) D31  */
// { BSP_IO_PORT_00_PIN_13,    P013   }, /* (32) D32  */
// { BSP_IO_PORT_00_PIN_15,    P015   }, /* (33) D33  */
// { BSP_IO_PORT_02_PIN_04,    P204   }, /* (34) D34  */
// { BSP_IO_PORT_02_PIN_05,    P205   }, /* (35) D35  */
// { BSP_IO_PORT_02_PIN_06,    P206   }, /* (36) D36  */
// { BSP_IO_PORT_02_PIN_12,    P212   }, /* (37) D37  */
// { BSP_IO_PORT_02_PIN_13,    P213   }, /* (38) D38  */