package main

import (
	"device/sam"
	"fmt"
	"machine"
	"time"
)

var (
	can0 = machine.CAN0
	can1 = machine.CAN1
	ch   = make(chan canMsg, 10)
)

type canMsg struct {
	bus byte
	can machine.CANRxBufferElement
}

func main() {
	go print(ch)

	can1.Configure(machine.CANConfig{
		TransferRate:   machine.CANTransferRate500kbps,
		TransferRateFD: machine.CANTransferRate1000kbps,
		Rx:             machine.CAN1_RX,
		Tx:             machine.CAN1_TX,
		Standby:        machine.CAN1_STANDBY,
	})
	// RF0NE : Rx FIFO 0 New Message Interrupt Enable
	can1.SetInterrupt(sam.CAN_IE_RF0NE, func(can *machine.CAN) {
		rxMsg := machine.CANRxBufferElement{}
		can.RxRaw(&rxMsg)
		msg := canMsg{
			bus: 1,
			can: rxMsg,
		}
		select {
		case ch <- msg:
		}
	})

	can0.Configure(machine.CANConfig{
		TransferRate:   machine.CANTransferRate500kbps,
		TransferRateFD: machine.CANTransferRate1000kbps,
		Rx:             machine.CAN0_RX,
		Tx:             machine.CAN0_TX,
		Standby:        machine.NoPin,
	})
	// RF0NE : Rx FIFO 0 New Message Interrupt Enable
	can0.SetInterrupt(sam.CAN_IE_RF0NE, func(can *machine.CAN) {
		rxMsg := machine.CANRxBufferElement{}
		can.RxRaw(&rxMsg)
		msg := canMsg{
			bus: 1,
			can: rxMsg,
		}
		select {
		case ch <- msg:
		}
	})

	for {
		can0.Tx(0x123, []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF}, false, false)
		time.Sleep(time.Millisecond * 500)
		can1.Tx(0x456, []byte{0xAA, 0xBB, 0xCC}, false, false)
		time.Sleep(time.Millisecond * 1000)
	}
}

func print(ch <-chan canMsg) {
	for {
		select {
		case m := <-ch:
			fmt.Printf("%d %03X %X ", m.bus, m.can.ID, m.can.DLC)
			for _, d := range m.can.DB[:m.can.Length()] {
				fmt.Printf("%02X ", d)
			}
			fmt.Printf("\r\n")
		}
	}
}
