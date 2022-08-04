package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/device/sam"
)

type canMsg struct {
	ch   byte
	id   uint32
	dlc  byte
	data []byte
}

func main() {
	ch := make(chan canMsg, 10)
	go func() {
		for {
			select {
			case m := <-ch:
				fmt.Printf("%d %03X %X ", m.ch, m.id, m.dlc)
				for _, d := range m.data {
					fmt.Printf("%02X ", d)
				}
				fmt.Printf("\r\n")
			}

		}
	}()

	can1 := machine.CAN1
	can1.Configure(machine.CANConfig{
		TransferRate:   machine.CANTransferRate500kbps,
		TransferRateFD: machine.CANTransferRate1000kbps,
		Rx:             machine.CAN1_RX,
		Tx:             machine.CAN1_TX,
		Standby:        machine.CAN1_STANDBY,
	})
	// RF0NE : Rx FIFO 0 New Message Interrupt Enable
	can1.SetInterrupt(sam.CAN_IE_RF0NE, func(*machine.CAN) {
		rxMsg := machine.CANRxBufferElement{}
		can1.RxRaw(&rxMsg)
		m := canMsg{ch: 1, id: rxMsg.ID, dlc: rxMsg.DLC, data: rxMsg.Data()}
		select {
		case ch <- m:
		}
	})

	can0 := machine.CAN0
	can0.Configure(machine.CANConfig{
		TransferRate:   machine.CANTransferRate500kbps,
		TransferRateFD: machine.CANTransferRate1000kbps,
		Rx:             machine.CAN0_RX,
		Tx:             machine.CAN0_TX,
		Standby:        machine.NoPin,
	})
	// RF0NE : Rx FIFO 0 New Message Interrupt Enable
	can0.SetInterrupt(sam.CAN_IE_RF0NE, func(*machine.CAN) {
		rxMsg := machine.CANRxBufferElement{}
		can0.RxRaw(&rxMsg)
		m := canMsg{ch: 2, id: rxMsg.ID, dlc: rxMsg.DLC, data: rxMsg.Data()}
		select {
		case ch <- m:
		}
	})

	for {
		can0.Tx(0x123, []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF}, false, false)
		time.Sleep(time.Millisecond * 500)
		can1.Tx(0x456, []byte{0xAA, 0xBB, 0xCC}, false, false)
		time.Sleep(time.Millisecond * 1000)
	}
}
