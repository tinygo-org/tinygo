// These are the supported alternate function numberings on the stm32f407
// +build stm32,stm32f407

// Alternate function settings on the stm32f4xx series

package stm32

const (
	// Alternative peripheral pin functions
	// AF0_SYSTEM is defined im the common bitfields package
	AF1_TIM1_2                AltFunc = 1
	AF2_TIM3_4_5                      = 2
	AF3_TIM8_9_10_11                  = 3
	AF4_I2C1_2_3                      = 4
	AF5_SPI1_SPI2                     = 5
	AF6_SPI3                          = 6
	AF7_USART1_2_3                    = 7
	AF8_USART4_5_6                    = 8
	AF9_CAN1_CAN2_TIM12_13_14         = 9
	AF10_OTG_FS_OTG_HS                = 10
	AF11_ETH                          = 11
	AF12_FSMC_SDIO_OTG_HS_1           = 12
	AF13_DCMI                         = 13
	AF14                              = 14
	AF15_EVENTOUT                     = 15
)
