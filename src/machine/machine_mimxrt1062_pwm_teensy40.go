//go:build teensy40

package machine

// PWM output pins of the Teensy 4.0, from Teensyduino cores/teensy4/pwm.c.
// The QuadTimer pins (10-15, 18, 19) are not supported by this driver.
var pwmPins = []pwmPinInfo{
	{D0, PWM1_1, pwmChannelX, 4},  // [AD_B0_03]
	{D1, PWM1_0, pwmChannelX, 4},  // [AD_B0_02]
	{D2, PWM4_2, pwmChannelA, 1},  // [EMC_04]
	{D3, PWM4_2, pwmChannelB, 1},  // [EMC_05]
	{D4, PWM2_0, pwmChannelA, 1},  // [EMC_06]
	{D5, PWM2_1, pwmChannelA, 1},  // [EMC_08]
	{D6, PWM2_2, pwmChannelA, 2},  // [B0_10]
	{D7, PWM1_3, pwmChannelB, 6},  // [B1_01]
	{D8, PWM1_3, pwmChannelA, 6},  // [B1_00]
	{D9, PWM2_2, pwmChannelB, 2},  // [B0_11]
	{D22, PWM4_0, pwmChannelA, 1}, // [AD_B1_08]
	{D23, PWM4_1, pwmChannelA, 1}, // [AD_B1_09]
	{D24, PWM1_2, pwmChannelX, 4}, // [AD_B0_12]
	{D25, PWM1_3, pwmChannelX, 4}, // [AD_B0_13]
	{D28, PWM3_1, pwmChannelB, 1}, // [EMC_32]
	{D29, PWM3_1, pwmChannelA, 1}, // [EMC_31]
	{D33, PWM2_0, pwmChannelB, 1}, // [EMC_07]
	{D34, PWM1_1, pwmChannelB, 1}, // [SD_B0_03]
	{D35, PWM1_1, pwmChannelA, 1}, // [SD_B0_02]
	{D36, PWM1_0, pwmChannelB, 1}, // [SD_B0_01]
	{D37, PWM1_0, pwmChannelA, 1}, // [SD_B0_00]
	{D38, PWM1_2, pwmChannelB, 1}, // [SD_B0_05]
	{D39, PWM1_2, pwmChannelA, 1}, // [SD_B0_04]
}
