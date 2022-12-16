package main

import (
	"fmt"
	"machine"
	"time"
)

func main() {
	can1 := machine.CAN1
	can1.Configure(machine.CANConfig{
		TransferRate:   machine.CANTransferRate500kbps,
		TransferRateFD: machine.CANTransferRate1000kbps,
		Rx:             machine.CAN1_RX,
		Tx:             machine.CAN1_TX,
		Standby:        machine.CAN1_STANDBY,
	})

	can0 := machine.CAN0
	can0.Configure(machine.CANConfig{
		TransferRate:   machine.CANTransferRate500kbps,
		TransferRateFD: machine.CANTransferRate1000kbps,
		Rx:             machine.CAN0_RX,
		Tx:             machine.CAN0_TX,
		Standby:        machine.NoPin,
	})

	rxMsg := machine.CANRxBufferElement{}

	for {
		can1.Tx(0x123, []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF}, false, false)
		can1.Tx(0x789, []byte{0x02, 0x24, 0x46, 0x67, 0x89, 0xAB, 0xCD, 0xEF, 0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF}, true, false)
		time.Sleep(time.Millisecond * 1000)

		sz0 := can0.RxFifoSize()
		if sz0 > 0 {
			fmt.Printf("CAN0 %d\r\n", sz0)
			for i := 0; i < sz0; i++ {
				can0.RxRaw(&rxMsg)
				fmt.Printf("-> %08X %X", rxMsg.ID, rxMsg.DLC)
				for j := byte(0); j < rxMsg.Length(); j++ {
					fmt.Printf(" %02X", rxMsg.DB[j])
				}
				fmt.Printf("\r\n")
			}
		}

		sz1 := can1.RxFifoSize()
		if sz1 > 0 {
			fmt.Printf("CAN1 %d\r\n", sz1)
			for i := 0; i < sz1; i++ {
				can1.RxRaw(&rxMsg)
				fmt.Printf("-> %08X %X", rxMsg.ID, rxMsg.DLC)
				for j := byte(0); j < rxMsg.Length(); j++ {
					fmt.Printf(" %02X", rxMsg.DB[j])
				}
				fmt.Printf("\r\n")
			}
		}
	}
}
