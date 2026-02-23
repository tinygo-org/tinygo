package main

import (
	"device/esp"
	"fmt"
	"machine"
	"time"
)

func dump(label string) {
	fmt.Println("==============", label, "==============")

	// RTC_CNTL
	fmt.Printf("RTC_CNTL.ANA_CONF                  = 0x%08X\n", esp.RTC_CNTL.ANA_CONF.Get())

	// SENS: ADC1 / общий SAR
	fmt.Printf("SENS.SAR_POWER_XPD_SAR             = 0x%08X\n", esp.SENS.SAR_POWER_XPD_SAR.Get())
	fmt.Printf("SENS.SAR_READER1_CTRL              = 0x%08X\n", esp.SENS.SAR_READER1_CTRL.Get())
	fmt.Printf("SENS.SAR_READER1_STATUS            = 0x%08X\n", esp.SENS.SAR_READER1_STATUS.Get())
	fmt.Printf("SENS.SAR_MEAS1_CTRL1               = 0x%08X\n", esp.SENS.SAR_MEAS1_CTRL1.Get())
	fmt.Printf("SENS.SAR_MEAS1_CTRL2               = 0x%08X\n", esp.SENS.SAR_MEAS1_CTRL2.Get())
	fmt.Printf("SENS.SAR_MEAS1_MUX                 = 0x%08X\n", esp.SENS.SAR_MEAS1_MUX.Get())
	fmt.Printf("SENS.SAR_ATTEN1                    = 0x%08X\n", esp.SENS.SAR_ATTEN1.Get())
	fmt.Printf("SENS.SAR_AMP_CTRL1                 = 0x%08X\n", esp.SENS.SAR_AMP_CTRL1.Get())
	fmt.Printf("SENS.SAR_AMP_CTRL2                 = 0x%08X\n", esp.SENS.SAR_AMP_CTRL2.Get())
	fmt.Printf("SENS.SAR_AMP_CTRL3                 = 0x%08X\n", esp.SENS.SAR_AMP_CTRL3.Get())

	// SENS: ADC2
	fmt.Printf("SENS.SAR_READER2_CTRL              = 0x%08X\n", esp.SENS.SAR_READER2_CTRL.Get())
	fmt.Printf("SENS.SAR_READER2_STATUS            = 0x%08X\n", esp.SENS.SAR_READER2_STATUS.Get())
	fmt.Printf("SENS.SAR_MEAS2_CTRL1               = 0x%08X\n", esp.SENS.SAR_MEAS2_CTRL1.Get())
	fmt.Printf("SENS.SAR_MEAS2_CTRL2               = 0x%08X\n", esp.SENS.SAR_MEAS2_CTRL2.Get())
	fmt.Printf("SENS.SAR_MEAS2_MUX                 = 0x%08X\n", esp.SENS.SAR_MEAS2_MUX.Get())
	fmt.Printf("SENS.SAR_ATTEN2                    = 0x%08X\n", esp.SENS.SAR_ATTEN2.Get())

	// SENS: прочее по SAR-периферии
	fmt.Printf("SENS.SAR_PERI_CLK_GATE_CONF        = 0x%08X\n", esp.SENS.SAR_PERI_CLK_GATE_CONF.Get())
	fmt.Printf("SENS.SAR_PERI_RESET_CONF           = 0x%08X\n", esp.SENS.SAR_PERI_RESET_CONF.Get())
	fmt.Printf("SENS.SAR_DEBUG_CONF                = 0x%08X\n", esp.SENS.SAR_DEBUG_CONF.Get())

	// APB_SARADC: общий FSM/клок/интерфейсы
	fmt.Printf("APB_SARADC.CTRL                    = 0x%08X\n", esp.APB_SARADC.CTRL.Get())
	fmt.Printf("APB_SARADC.CTRL2                   = 0x%08X\n", esp.APB_SARADC.CTRL2.Get())
	fmt.Printf("APB_SARADC.FSM_WAIT                = 0x%08X\n", esp.APB_SARADC.FSM_WAIT.Get())
	fmt.Printf("APB_SARADC.SAR1_STATUS             = 0x%08X\n", esp.APB_SARADC.SAR1_STATUS.Get())
	fmt.Printf("APB_SARADC.SAR2_STATUS             = 0x%08X\n", esp.APB_SARADC.SAR2_STATUS.Get())
	fmt.Printf("APB_SARADC.FILTER_CTRL0            = 0x%08X\n", esp.APB_SARADC.FILTER_CTRL0.Get())
	fmt.Printf("APB_SARADC.FILTER_CTRL1            = 0x%08X\n", esp.APB_SARADC.FILTER_CTRL1.Get())
	fmt.Printf("APB_SARADC.ARB_CTRL                = 0x%08X\n", esp.APB_SARADC.ARB_CTRL.Get())
	fmt.Printf("APB_SARADC.CLKM_CONF               = 0x%08X\n", esp.APB_SARADC.CLKM_CONF.Get())
	fmt.Printf("APB_SARADC.INT_ENA                 = 0x%08X\n", esp.APB_SARADC.INT_ENA.Get())
	fmt.Printf("APB_SARADC.INT_RAW                 = 0x%08X\n", esp.APB_SARADC.INT_RAW.Get())
	fmt.Printf("APB_SARADC.INT_ST                  = 0x%08X\n", esp.APB_SARADC.INT_ST.Get())
	fmt.Printf("APB_SARADC.INT_CLR                 = 0x%08X\n", esp.APB_SARADC.INT_CLR.Get())
	fmt.Printf("APB_SARADC.DMA_CONF                = 0x%08X\n", esp.APB_SARADC.DMA_CONF.Get())
	fmt.Printf("APB_SARADC.APB_SARADC1_DATA_STATUS = 0x%08X\n", esp.APB_SARADC.APB_SARADC1_DATA_STATUS.Get())
	fmt.Printf("APB_SARADC.APB_SARADC2_DATA_STATUS = 0x%08X\n", esp.APB_SARADC.APB_SARADC2_DATA_STATUS.Get())

	fmt.Println()
}

func main() {
	// Дадим время подключиться к CDC/Serial
	time.Sleep(2 * time.Second)

	// 1) Состояние сразу после загрузки (только Arduino что-то мог настроить)
	dump("BEFORE TinyGo initADC / Get")

	// 2) Наш initADC() и один Get() с GPIO4 (ADC1 ch3)
	fmt.Println("---- TinyGo InitADC + one Get() on GPIO4 ----")
	machine.InitADC()
	s := machine.ADC{Pin: machine.GPIO4}
	_ = s.Configure(machine.ADCConfig{})
	_ = s.Get()

	// Небольшая пауза, чтобы все биты успели установиться
	time.Sleep(100 * time.Millisecond)

	// 3) Состояние после нашей инициализации и одного чтения
	dump("AFTER TinyGo initADC / Get")

	// Дальше просто спим, чтобы логи не улетели
	for {
		time.Sleep(time.Second * 10)
	}
}

