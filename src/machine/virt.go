//go:build tinygo.riscv32 && virt

// Machine implementation for VirtIO targets.
// At the moment only QEMU RISC-V is supported, but support for ARM for example
// should not be difficult to add with a change to virtioFindDevice.

package machine

import (
	"errors"
	"runtime/volatile"
	"sync"
	"unsafe"
)

const deviceName = "riscv-qemu"

func (p Pin) Set(high bool) {
	// no pins defined
}

var rngLock sync.Mutex
var rngDevice *virtioDevice1
var rngBuf volatile.Register32

var errNoRNG = errors.New("machine: no entropy source found")
var errNoRNGData = errors.New("machine: entropy source didn't return enough data")

// GetRNG returns random numbers from a VirtIO entropy source.
// When running in QEMU, it requires adding the RNG device:
//
//	-device virtio-rng-device
func GetRNG() (uint32, error) {
	rngLock.Lock()

	// Initialize the device on first use.
	if rngDevice == nil {
		// Search for an available RNG.
		rngDevice = virtioFindDevice(virtioDeviceEntropySource)
		if rngDevice == nil {
			rngLock.Unlock()
			return 0, errNoRNG
		}

		// Initialize the device.
		rngDevice.status.Set(0) // reset device
		rngDevice.status.Set(virtioDeviceStatusAcknowledge)
		rngDevice.status.Set(virtioDeviceStatusAcknowledge | virtioDeviceStatusDriver)
		rngDevice.hostFeaturesSel.Set(0)
		rngDevice.status.Set(virtioDeviceStatusAcknowledge | virtioDeviceStatusDriver | virtioDeviceStatusDriverOk)
		rngDevice.guestPageSize.Set(4096)

		// Configure queue, according to section 4.2.4 "Legacy interface".
		// Note: we're skipping checks for queuePFM and queueNumMax.
		rngDevice.queueSel.Set(0)      // use queue 0 (the only queue)
		rngDevice.queueNum.Set(1)      // use a single buffer in the queue
		rngDevice.queueAlign.Set(4096) // default alignment appears to be 4096
		rngDevice.queuePFN.Set(uint32(uintptr(unsafe.Pointer(&rngQueue))) / 4096)

		// Configure the only buffer in the queue (but don't increment
		// rngQueue.available yet).
		rngQueue.buffers[0].address = uint64(uintptr(unsafe.Pointer(&rngBuf)))
		rngQueue.buffers[0].length = uint32(unsafe.Sizeof(rngBuf))
		rngQueue.buffers[0].flags = 2 // 2 means write-only buffer
	}

	// Increment the available ring buffer. This doesn't actually change the
	// buffer index (it's a ring with a single entry), but the number needs to
	// be incremented otherwise the device won't recognize a new buffer.
	index := rngQueue.available.index
	rngQueue.available.index = index + 1
	rngDevice.queueNotify.Set(0) // notify the device of the 'new' (reused) buffer
	for rngQueue.used.index.Get() != index+1 {
		// Busy wait until the RNG buffer is filled.
		// A better way would be to wait for an interrupt, but since this driver
		// implementation is mostly used for testing it's good enough for now.
	}

	// Check that we indeed got 4 bytes back.
	if rngQueue.used.ring[0].length != 4 {
		rngLock.Unlock()
		return 0, errNoRNGData
	}

	// Read the resulting random numbers.
	result := rngBuf.Get()

	rngLock.Unlock()

	return result, nil
}

// Implement a driver for the VirtIO entropy device.
// https://docs.oasis-open.org/virtio/virtio/v1.2/csd01/virtio-v1.2-csd01.html
// http://wiki.osdev.org/Virtio
// http://www.dumais.io/index.php?article=aca38a9a2b065b24dfa1dee728062a12

const (
	virtioDeviceStatusAcknowledge = 1
	virtioDeviceStatusDriver      = 2
	virtioDeviceStatusDriverOk    = 4
	virtioDeviceStatusFeaturesOk  = 8
	virtioDeviceStatusFailed      = 128
)

const (
	virtioDeviceReserved = iota
	virtioDeviceNetworkCard
	virtioDeviceBlockDevice
	virtioDeviceConsole
	virtioDeviceEntropySource
	// there are more device types
)

// VirtIO device version 1
type virtioDevice1 struct {
	magic            volatile.Register32 // always 0x74726976
	version          volatile.Register32
	deviceID         volatile.Register32
	vendorID         volatile.Register32
	hostFeatures     volatile.Register32
	hostFeaturesSel  volatile.Register32
	_                [2]uint32
	guestFeatures    volatile.Register32
	guestFeaturesSel volatile.Register32
	guestPageSize    volatile.Register32
	_                uint32
	queueSel         volatile.Register32
	queueNumMax      volatile.Register32
	queueNum         volatile.Register32
	queueAlign       volatile.Register32
	queuePFN         volatile.Register32
	_                [3]uint32
	queueNotify      volatile.Register32
	_                [3]uint32
	interruptStatus  volatile.Register32
	interruptAck     volatile.Register32
	_                [2]uint32
	status           volatile.Register32
}

// VirtIO queue, with a single buffer.
type virtioQueue struct {
	buffers [1]struct {
		address uint64
		length  uint32
		flags   uint16
		next    uint16
	} // 16 bytes

	available struct {
		flags      uint16
		index      uint16
		ring       [1]uint16
		eventIndex uint16
	} // 8 bytes

	_ [4096 - 16*1 - 8*1]byte // padding (to align on a 4096 byte boundary)

	used struct {
		flags uint16
		index volatile.Register16
		ring  [1]struct {
			index  uint32
			length uint32
		}
		availEvent uint16
	}
}

func virtioFindDevice(deviceID uint32) *virtioDevice1 {
	// On RISC-V, QEMU defines 8 VirtIO devices starting at 0x10001000 and
	// repeating every 0x1000 bytes.
	// The memory map can be seen in the QEMU source code:
	// https://github.com/qemu/qemu/blob/master/hw/riscv/virt.c
	for i := 0; i < 8; i++ {
		dev := (*virtioDevice1)(unsafe.Pointer(uintptr(0x10001000 + i*0x1000)))
		if dev.magic.Get() != 0x74726976 || dev.version.Get() != 1 || dev.deviceID.Get() != deviceID {
			continue
		}
		return dev
	}
	return nil
}

// A VirtIO queue needs to be page-aligned.
//
//go:align 4096
var rngQueue virtioQueue
