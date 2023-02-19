//go:build raspberrypi

package machine

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	PinOutput PinMode = iota
	PinInput
	PinInputPulldown
	PinInputPullup
)

const GPIO10 Pin = 10
const GPIO17 Pin = 17
const GPIO22 Pin = 22
const GPIO23 Pin = 23
const GPIO27 Pin = 27

const P6 Pin = 6
const P7 Pin = 7
const P8 Pin = 8
const P9 Pin = 9

const LED = GPIO17

const deviceName = "raspberrypi"

var gpiochip0 int

func init() {
	gpiochip0, _ = syscall.Open("/dev/gpiochip0", os.O_RDWR, 0)
}

const GPIO_V2_LINES_MAX = 64
const GPIO_MAX_NAME_SIZE = 32

type gpio_v2_line_request struct {
	offsets           [GPIO_V2_LINES_MAX]uint32
	consumer          [GPIO_MAX_NAME_SIZE]uint8
	config            gpio_v2_line_config
	num_lines         uint32
	event_buffer_size uint32
	/* Pad to fill implicit padding and reserve space for future use. */
	padding [5]uint32
	fd      int32
}

const GPIO_V2_LINE_NUM_ATTRS_MAX = 10

type gpio_v2_line_config struct {
	flags     uint64
	num_attrs uint32
	/* Pad to fill implicit padding and reserve space for future use. */
	padding [5]uint32
	attrs   [GPIO_V2_LINE_NUM_ATTRS_MAX]gpio_v2_line_config_attribute
}

type gpio_v2_line_config_attribute struct {
	attr gpio_v2_line_attribute
	mask uint64
}

type gpio_v2_line_attribute struct {
	id      uint32
	padding uint32
	union   [8]byte
	/* union {
		__aligned_u64 flags;
		__aligned_u64 values;
		__u32 debounce_period_us;
	};*/
}

// gpio_v2_line_attr_id
const GPIO_V2_LINE_ATTR_ID_FLAGS = 1
const GPIO_V2_LINE_ATTR_ID_OUTPUT_VALUES = 2

// gpio_v2_line_flag
const GPIO_V2_LINE_FLAG_USED = 1
const GPIO_V2_LINE_FLAG_ACTIVE_LOW = 1 << 1
const GPIO_V2_LINE_FLAG_INPUT = 1 << 2
const GPIO_V2_LINE_FLAG_OUTPUT = 1 << 3
const GPIO_V2_LINE_FLAG_EDGE_RISING = 1 << 4
const GPIO_V2_LINE_FLAG_EDGE_FALLING = 1 << 5
const GPIO_V2_LINE_FLAG_OPEN_DRAIN = 1 << 6
const GPIO_V2_LINE_FLAG_OPEN_SOURCE = 1 << 7
const GPIO_V2_LINE_FLAG_BIAS_PULL_UP = 1 << 8
const GPIO_V2_LINE_FLAG_BIAS_PULL_DOWN = 1 << 9
const GPIO_V2_LINE_FLAG_BIAS_DISABLED = 1 << 10

const GPIO_V2_GET_LINE_IOCTL = 0xc250b407
const GPIO_V2_LINE_GET_VALUES_IOCTL = 0xc010b40e
const GPIO_V2_LINE_SET_VALUES_IOCTL = 0xc010b40f

type gpio_v2_line_values struct {
	bits uint64
	mask uint64
}

func Ioctl(fd int, req uint, buf unsafe.Pointer) (n int, err error) {
	_, _, err1 := syscall.RawSyscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(req), uintptr(buf))
	if err1 < 0 {
		err = fmt.Errorf("Error %d", err1) // getErrno()
	}
	return
}

var request gpio_v2_line_request
var request_need_apply bool

func init() {
	// set consumer
	copy(request.consumer[:], ([]uint8)("tinygo"))
	request.config.num_attrs = uint32(PinInputPullup + 1)
	// prepare the 4 pinmodes
	request.config.attrs[PinOutput].attr.id = GPIO_V2_LINE_ATTR_ID_FLAGS
	var flags = uint64(GPIO_V2_LINE_FLAG_OUTPUT)
	copy(request.config.attrs[PinOutput].attr.union[:], (*[8]byte)(unsafe.Pointer(&flags))[:]) // uint64 to union

	request.config.attrs[PinInput].attr.id = GPIO_V2_LINE_ATTR_ID_FLAGS
	flags = uint64(GPIO_V2_LINE_FLAG_INPUT)
	copy(request.config.attrs[PinInput].attr.union[:], (*[8]byte)(unsafe.Pointer(&flags))[:]) // uint64 to union

	request.config.attrs[PinInputPulldown].attr.id = GPIO_V2_LINE_ATTR_ID_FLAGS
	flags = uint64(GPIO_V2_LINE_FLAG_INPUT | GPIO_V2_LINE_FLAG_BIAS_PULL_DOWN)
	copy(request.config.attrs[PinInputPulldown].attr.union[:], (*[8]byte)(unsafe.Pointer(&flags))[:]) // uint64 to union

	request.config.attrs[PinInputPullup].attr.id = GPIO_V2_LINE_ATTR_ID_FLAGS
	flags = uint64(GPIO_V2_LINE_FLAG_INPUT | GPIO_V2_LINE_FLAG_BIAS_PULL_UP)
	copy(request.config.attrs[PinInputPullup].attr.union[:], (*[8]byte)(unsafe.Pointer(&flags))[:]) // uint64 to unionå
}

func (p Pin) Configure(config PinConfig) {
	found := false
	slot := 0
	for ; slot < int(request.num_lines); slot++ {
		if request.offsets[slot] == (uint32)(p) {
			found = true
			break
		}
	}
	if !found {
		slot = (int)(request.num_lines)
		request.offsets[slot] = (uint32)(p)
		request.num_lines++
	}
	// set the bit in the Mode mask
	request.config.attrs[config.Mode].mask |= 1 << slot
	request_need_apply = true
}

func apply_config() {
	Ioctl(gpiochip0, GPIO_V2_GET_LINE_IOCTL, unsafe.Pointer(&request))
}

func (p Pin) Set(value bool) {
	if request_need_apply {
		apply_config()
		request_need_apply = false
	}
	var i = 0
	for ; i < int(request.num_lines); i++ {
		if request.offsets[i] == (uint32)(p) {
			break
		}
	}
	if i == int(request.num_lines) {
		return
	}
	var values gpio_v2_line_values

	values.mask = 1 << i
	if value {
		values.bits = values.mask
	}
	Ioctl(int(request.fd), GPIO_V2_LINE_SET_VALUES_IOCTL, unsafe.Pointer(&values))
}

func (p Pin) Get() bool {
	if request_need_apply {
		apply_config()
		request_need_apply = false
	}
	i := 0
	for ; i < int(request.num_lines); i++ {
		if request.offsets[i] == (uint32)(p) {
			break
		}
	}
	if i == int(request.num_lines) {
		return false
	}
	var values gpio_v2_line_values

	values.mask = 1 << i
	Ioctl(int(request.fd), GPIO_V2_LINE_GET_VALUES_IOCTL, unsafe.Pointer(&values))
	return (values.bits & values.mask) != 0
}

type SPI struct {
	Bus int
	Fd  int
}

var (
	SPI0  = &_SPI0
	_SPI0 = SPI{
		Bus: 0,
	}
)

// SPIConfig is used to store config info for SPI.
type SPIConfig struct {
	Frequency uint32
	// LSB not supported on rp2040.
	LSBFirst bool
	// Mode's two most LSB are CPOL and CPHA. i.e. Mode==2 (0b10) is CPOL=1, CPHA=0
	Mode uint8
	// Number of data bits per transfer. Valid values 4..16. Default and recommended is 8.
	DataBits uint8
	// Serial clock pin
}

func (spi *SPI) Configure(config SPIConfig) error {
	fd, err := syscall.Open("/dev/spidev0.0", syscall.O_RDWR, 0)
	if err != nil {
		return err
	}
	spi.Fd = fd

	var mode uint8
	const SPI_IOC_RD_MODE = 0x80016b01
	const SPI_IOC_WR_MODE = 0x40016b01
	const SPI_IOC_RD_MAX_SPEED_HZ = 0x80046b04
	const SPI_IOC_WR_MAX_SPEED_HZ = 0x80046b04

	// read mode
	if _, err = Ioctl(fd, SPI_IOC_RD_MODE, unsafe.Pointer(&mode)); err != nil {
		return err
	}

	mode &= 0 ^ 3       // mask bits
	mode |= config.Mode // set mode bits
	// set mode
	if _, err = Ioctl(fd, SPI_IOC_WR_MODE, unsafe.Pointer(&mode)); err != nil {
		return err
	}

	speed := uint32(config.Frequency)
	if _, err = Ioctl(fd, SPI_IOC_WR_MAX_SPEED_HZ, unsafe.Pointer(&speed)); err != nil {
		return err
	}
	return nil
}

func (spi SPI) Transfer(w byte) (byte, error) {
	return 0, nil
}

func (spi SPI) Tx(w, r []byte) (err error) {
	switch {
	case w == nil:
		// read only
		_, err = syscall.Read(spi.Fd, r)
	case r == nil:
		// write only
		_, err = syscall.Write(spi.Fd, w)

	case len(w) == 1 && len(r) > 1:
		panic("not impl")
	default:
		panic("not impl")
	}

	return nil
}
