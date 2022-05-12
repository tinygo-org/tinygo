//go:build baremetal && usb.cdc && (atsamd51 || atsame5x)
// +build baremetal
// +build usb.cdc
// +build atsamd51 atsame5x

package usb

import "unsafe"

//go:inline
func (d *dhw) descriptorTable() uintptr {
	return uintptr(unsafe.Pointer(&descCDC[d.cc.config-1].ed[0]))
}

// endpointDescriptor returns the endpoint descriptor for the given endpoint
// address, encoded as direction D and endpoint number N with the 8-bit mask
// D000NNNN.
//go:inline
func (d *dhw) endpointDescriptor(endpoint uint8) *dhwEPDesc {
	num, dir := unpackEndpoint(endpoint)
	return &descCDC[d.cc.config-1].ed[num][dir]
}

//go:inline
func (d *dhw) controlSetupBuffer() uintptr {
	return uintptr(unsafe.Pointer(&descCDC[d.cc.config-1].sx[0]))
}

//go:inline
func (d *dhw) controlStatusBuffer(data []uint8) uintptr {
	// reference to class configuration data
	c := descCDC[d.cc.config-1]
	for i := range c.cx {
		c.cx[i] = 0 // zero out the control reply buffer
	}
	// copy the given data into control reply buffer
	copy(c.cx[:], data)
	return uintptr(unsafe.Pointer(&c.cx[0]))
}

// =============================================================================
//  [CDC-ACM] Serial UART (Virtual COM Port)
// =============================================================================

func (d *dhw) cdcConfigure() {

	acm := &descCDC[d.cc.config-1]

	acm.setState(descCDCStateConfigured)

	// SAMx51 only supports USB full-speed (FS) operation
	acm.sxSize = descCDCStatusFSPacketSize
	acm.rxSize = descCDCDataRxFSPacketSize
	acm.txSize = descCDCDataTxFSPacketSize

	rq := acm.rxq[:]
	tq := acm.txq[:]

	// Rx gives priority to incoming data, Tx gives priority to outgoing data
	acm.rq.Init(&rq, int(acm.rxSize), QueueFullDiscardFirst)
	acm.tq.Init(&tq, int(acm.txSize), QueueFullDiscardLast)

	d.endpointEnable(txEndpoint(descCDCEndpointStatus),
		false, descCDCConfigAttrStatus)
	d.endpointEnable(rxEndpoint(descCDCEndpointDataRx),
		false, descCDCConfigAttrDataRx)
	d.endpointEnable(txEndpoint(descCDCEndpointDataTx),
		false, descCDCConfigAttrDataTx)

	d.endpointConfigure(txEndpoint(descCDCEndpointStatus),
		nil)
	d.endpointConfigure(rxEndpoint(descCDCEndpointDataRx),
		d.cdcReceiveComplete)
	d.endpointConfigure(txEndpoint(descCDCEndpointDataTx),
		d.cdcTransmitComplete)

	d.cdcReceiveStart(rxEndpoint(descCDCEndpointDataRx))
}

func (d *dhw) cdcSetLineState(state uint16) {
	acm := &descCDC[d.cc.config-1]
	acm.setState(descCDCStateLineState)
	if acm.ls.parse(state) {
		// TBD: respond to changes in line state?
	}
}

func (d *dhw) cdcSetLineCoding(coding []uint8) {
	acm := &descCDC[d.cc.config-1]
	acm.setState(descCDCStateLineCoding)
	if acm.lc.parse(coding) {
		switch acm.lc.baud {
		case 1200:
			if acm.ls.dataTerminalReady {
				// reboot CPU
			}
		}
	}
}

func (d *dhw) cdcReady() bool {
	acm := &descCDC[d.cc.config-1]
	// Ensure we have received SET_CONFIGURATION class request, and then both
	// SET_LINE_STATE and SET_LINE_CODING CDC requests (in that order).
	return d.state() == dcdStateConfigured &&
		acm.st.Get() == uint8(descCDCStateLineCoding)
}

func (d *dhw) cdcReceiveStart(endpoint uint8) {
	acm := &descCDC[d.cc.config-1]
	num := uint16(endpoint) & descEndptAddrNumberMsk

	ready, _ := d.ep[num][descDirRx].scheduleTransfer(
		uintptr(unsafe.Pointer(&acm.rx[0])), acm.rxSize)
	if ready {
		if xfer, ok := d.ep[num][descDirRx].pendingTransfer(); ok {
			// Update the active transfer descriptor on the corresponding endpoint.
			d.ep[num][descDirRx].setActiveTransfer(xfer)
			d.endpointTransfer(endpoint, xfer.data, xfer.size)
		}
	}
}

func (d *dhw) cdcReceiveComplete(endpoint uint8, size uint32) {
	acm := &descCDC[d.cc.config-1]
	num := uint16(endpoint) & descEndptAddrNumberMsk

	if xfer, ok := d.ep[num][descDirRx].activeTransfer(); ok {
		for ptr := xfer.data; ptr < xfer.data+uintptr(size); ptr++ {
			acm.rq.Enq(*(*uint8)(unsafe.Pointer(ptr)))
		}
	}
	d.ep[num][descDirRx].setActiveTransfer(nil)
	d.cdcReceiveStart(endpoint)
}

func (d *dhw) cdcTransmitStart(endpoint uint8) {
	acm := &descCDC[d.cc.config-1]
	num := uint16(endpoint) & descEndptAddrNumberMsk

	// BULK data endpoints can simply use a single time slot in the schedule, and
	// repeatedly transfer from the same transmit buffer (acm.tx) as soon as the
	// transaction complete callback has been called for a prior transaction.

	// Do not schedule another transfer if one is already active, or if our Tx
	// FIFO is currently empty.
	if d.ep[num][descDirTx].hasActiveTransfer() || acm.tq.Len() == 0 {
		return
	}

	if send, err := acm.tq.Read(acm.tx[:]); err == nil && send > 0 {
		ready, _ := d.ep[num][descDirTx].scheduleTransfer(
			uintptr(unsafe.Pointer(&acm.tx[0])), uint32(send))
		if ready {
			if xfer, ok := d.ep[num][descDirTx].pendingTransfer(); ok {
				d.ep[num][descDirTx].setActiveTransfer(xfer)
				d.endpointTransfer(endpoint, xfer.data, xfer.size)
			}
		}
	}
}

func (d *dhw) cdcTransmitComplete(endpoint uint8, size uint32) {
	acm := &descCDC[d.cc.config-1]
	num := uint16(endpoint) & descEndptAddrNumberMsk
	if size > 0 && size%acm.txSize == 0 {
		// Send ZLP if transfer length is a non-zero multiple of max packet size.
		d.endpointTransfer(endpoint, 0, 0)
	}
	d.ep[num][descDirTx].setActiveTransfer(nil)
	d.cdcTransmitStart(endpoint)
}

// cdcFlush discards all buffered input (Rx) data.
func (d *dhw) cdcFlush() {
	acm := &descCDC[d.cc.config-1]
	acm.rq.Reset(int(acm.rxSize))
}

func (d *dhw) cdcAvailable() int {
	acm := &descCDC[d.cc.config-1]
	return acm.rq.Len()
}

func (d *dhw) cdcPeek() (uint8, bool) {
	acm := &descCDC[d.cc.config-1]
	return acm.rq.Front()
}

func (d *dhw) cdcReadByte() (uint8, bool) {
	acm := &descCDC[d.cc.config-1]
	return acm.rq.Deq()
}

func (d *dhw) cdcRead(data []uint8) (int, error) {
	acm := &descCDC[d.cc.config-1]
	return acm.rq.Read(data)
}

func (d *dhw) cdcWriteByte(c uint8) error {
	_, err := d.cdcWrite([]uint8{c})
	return err
}

func (d *dhw) cdcWrite(data []uint8) (int, error) {

	acm := &descCDC[d.cc.config-1]
	num := uint16(descCDCEndpointDataTx) & descEndptAddrNumberMsk

	var sent int
	var werr error
	for off := 0; off < len(data); off += int(acm.txSize) {

		cnt := len(data[off:])
		if cnt > int(acm.txSize) {
			cnt = int(acm.txSize)
		}

		// Block until we have room in the Tx FIFO. Space will become available once
		// the endpoint transaction complete interrupt is raised for the Tx BULK data
		// endpoint, and then the uartTransmitComplete callback has dequeued data from
		// the Tx FIFO (acm.tq) into the Tx transmit buffer (acm.tx).
		for acm.tq.Rem() < cnt {
		}

		// Add data to Tx FIFO
		add, err := acm.tq.Write(data[off : off+cnt])
		if err != nil {
			werr = err
			break
		}
		sent += add

		if d.ep[num][descDirTx].hasActiveTransfer() {
			// If there is already a transmit in-progress, wait for its callback to
			// detect new data in the FIFO and continue the transfer automatically.
		} else {
			// Otherwise, initiate a new data transfer.
			d.cdcTransmitStart(txEndpoint(descCDCEndpointDataTx))
		}
	}

	return sent, werr
}
