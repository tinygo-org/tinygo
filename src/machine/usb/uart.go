package usb

import (
	"errors"
)

var (
	ErrInvalidPort = errors.New("invalid USB port")
)

type (
	UARTConfig struct {
		BaudRate uint32
	}

	// UART represents a virtual serial (UART) device emulation using the USB
	// CDC-ACM device class driver.
	UART struct {
		port  uint8                   // USB port (core index, e.g., 0-1)
		desc  *ConfigDeviceDescriptor // User-provided USB device descriptor information
		class *deviceClass            // USB device class handle
		acm   *deviceCDCACM           // USB CDC-ACM device class handle

		attached  bool
		transact  bool
		config    uint8 // selected configuration index (1-based, 0=invalid)
		speed     uint8 // enumerated bus speed (low, full, high)
		alternate deviceCDCACMAlternateList
	}
)

var (
	uartAbstractState = []uint8{0x00, 0x00}
	uartCountryCode   = []uint8{0x00, 0x00}
)

func (uart *UART) SetPort(port uint8) {
	if port >= ConfigPortCount || port >= configDeviceCount ||
		port >= configDeviceCDCACMCount {
		return // invalid port for USB CDC-ACM device class
	}
	// TODO: if changed, may need to reset port controller and tell host to
	//       re-enumerate devices
	uart.port = port
}

func (uart *UART) SetDeviceDescriptor(desc *ConfigDeviceDescriptor) {
	if nil == desc {
		return // invalid device descriptor information
	}
	// TODO: if changed, may need to reset port controller and tell host to
	//       re-enumerate devices
	uart.desc = desc
}

func (uart *UART) Configure(config UARTConfig) error {

	if uart.port >= ConfigPortCount || uart.port >= configDeviceCount ||
		uart.port >= configDeviceCDCACMCount {
		return ErrInvalidPort
	}

	// use default configuration index (1-based index; 0=invalid)
	uart.config = configDeviceCDCACMConfigurationIndex

	// modify the global basic configuration struct configDeviceCDCACM for our USB
	// port and configuration index.
	//
	// these settings are copied into the real CDC-ACM object, using interface
	// deviceClassDriver, via initialization method (*deviceCDCACM).init().

	// change baud rate from default, if provided
	if config.BaudRate != 0 {
		configDeviceCDCACM[uart.port][uart.config-1].lineCodingBaudRate =
			config.BaudRate
	}

	// verify we have a free USB port and take ownership of it
	port, status := initPort(uart.port, modeDevice)
	if !status.OK() {
		return status
	}

	// apply the CDC-ACM configuration to our port
	uart.acm, uart.class = port.initCDCACM(uart.config, uart)

	// enable USB device mode functionality, which allows a host to enumerate us
	status = port.device.controller.enable(true)
	if !status.OK() {
		return status
	}

	return nil
}

func (uart *UART) deviceEvent(ev deviceEventID, param interface{}) status {

	switch ev {
	case deviceEventBusReset:
		uart.attached = false
		uart.config = 0 // clear selected configuration (1-based index)
		s := uart.acm.device.busSpeed(&uart.speed)
		if s.OK() {
			s = uart.acm.device.setBusSpeed(uart.speed)
		}
		return s

	case deviceEventSetConfiguration:
		if id, ok := param.(uint8); ok {
			if 0 == id {
				uart.attached = false
				uart.config = 0 // clear selected configuration (1-based index)
				return statusSuccess
			}
			// verify we have a config struct at given index
			if nil != uart.class && nil != uart.class.config &&
				int(id) <= len(uart.class.config) &&
				int(id) <= len(configDeviceCDCACM[uart.port]) {
				config := uart.class.config[id-1]             // receiver configuration
				system := configDeviceCDCACM[uart.port][id-1] // platform configuration
				// verify the config is a CDC device and has a defined driver
				if deviceClassCDC == config.info.classID && nil != config.driver {
					// finally, verify the config driver is an ACM device driver
					if _, ok := config.driver.(*deviceCDCACM); ok {
						uart.attached = true
						uart.config = id
						return uart.acm.receive(system.dataBulkOutEndpoint, uart.acm.recvBuffer[:],
							uint32(system.dataBulkOutPacketSize))
					}
				}
			}
		}
		return statusNotSupported // all non-error paths return early

	case deviceEventSetInterface:
		if uart.attached {
			if setting, ok := param.(uint16); ok {
				inf := (setting & 0xFF00) >> 8
				if inf < configDeviceCDCACMInterfaceCount {
					uart.alternate[inf] = setting & 0x00FF
				}
			}
		}

	case deviceEventGetConfiguration:
	case deviceEventGetInterface:
	case deviceEventGetDeviceDescriptor:
	case deviceEventGetConfigurationDescriptor:
	case deviceEventGetStringDescriptor:
	}

	return statusSuccess
}

func (uart *UART) classEvent(ev uint32, param interface{}) status {

	if nil == uart.acm {
		return statusInvalidHandle
	}

	switch deviceCDCACMEventID(ev) {
	case deviceCDCACMEventSendResponse:
		if ctl, ok := param.(deviceEndpointControlMessage); ok {

			// verify receiver configuration
			if int(uart.port) >= len(configDeviceCDCACM) || uart.config == 0 ||
				int(uart.config) > len(configDeviceCDCACM[uart.port]) {
				return statusNotSupported
			}
			// get platform configuration
			system := configDeviceCDCACM[uart.port][uart.config-1]

			if ctl.length > 0 && 0 != ctl.length%uint32(system.dataBulkInPacketSize) {
				// If the last packet is the size of endpoint, then send an additional
				// zero-ended packet to notify the host it may flush output.
				return uart.acm.send(system.dataBulkInEndpoint, nil, 0)
			}

			if uart.attached && uart.transact &&
				(nil != ctl.buffer || (nil == ctl.buffer && 0 == ctl.length)) {
				// send complete, schedule buffer for next receive event
				return uart.acm.receive(system.dataBulkOutEndpoint, uart.acm.recvBuffer[:],
					uint32(system.dataBulkOutPacketSize))
			}
		}
		return statusError // all non-error paths return early

	case deviceCDCACMEventRecvResponse:
		if ctl, ok := param.(deviceEndpointControlMessage); ok {
			if uart.attached && uart.transact {
				uart.acm.recvSize = ctl.length
				if 0 == ctl.length {

					// verify receiver configuration
					if int(uart.port) >= len(configDeviceCDCACM) || uart.config == 0 ||
						int(uart.config) > len(configDeviceCDCACM[uart.port]) {
						return statusNotSupported
					}
					// get platform configuration
					system := configDeviceCDCACM[uart.port][uart.config-1]

					// schedule buffer for next receive event
					return uart.acm.receive(system.dataBulkOutEndpoint, uart.acm.recvBuffer[:],
						uint32(system.dataBulkOutPacketSize))
				}
			}
		}
		return statusError // all non-error paths return early

	case deviceCDCACMEventSerialStateNotify:
		uart.acm.hasSentState = false
		return statusSuccess

	case deviceCDCACMEventSendEncapsulatedCommand:
		return statusNotImplemented

	case deviceCDCACMEventGetEncapsulatedResponse:
		return statusNotImplemented

	case deviceCDCACMEventSetCommFeature:
		if req, ok := param.(deviceCDCACMRequestParam); ok {
			switch req.setupValue {
			case deviceCDCFeatureAbstractState:
				if req.isSetup {
					*(req.buffer) = uartAbstractState
				} else {
					*(req.length) = 0
				}
				return statusSuccess

			case deviceCDCFeatureCountrySetting:
				if req.isSetup {
					*(req.buffer) = uartCountryCode
				} else {
					*(req.length) = 0
				}
				return statusSuccess
			}
			return statusInvalidParameter // unrecognized request code
		}
		return statusError // all non-error paths return early

	case deviceCDCACMEventGetCommFeature:
		if req, ok := param.(deviceCDCACMRequestParam); ok {
			switch req.setupValue {
			case deviceCDCFeatureAbstractState:
				*(req.buffer) = uartAbstractState
				*(req.length) = uint32(len(uartAbstractState))
				return statusSuccess

			case deviceCDCFeatureCountrySetting:
				*(req.buffer) = uartCountryCode
				*(req.length) = uint32(len(uartCountryCode))
				return statusSuccess
			}
			return statusInvalidParameter // unrecognized request code
		}
		return statusError // all non-error paths return early

	case deviceCDCACMEventClearCommFeature:
		return statusNotImplemented

	case deviceCDCACMEventGetLineCoding:
		if req, ok := param.(deviceCDCACMRequestParam); ok {
			// verify receiver configuration
			if int(uart.port) >= len(deviceCDCACMLineCoding) || uart.config == 0 ||
				int(uart.config) > len(deviceCDCACMLineCoding[uart.port]) {
				return statusNotSupported
			}
			lineCoding := deviceCDCACMLineCoding[uart.port][uart.config-1]
			*(req.buffer) = lineCoding
			*(req.length) = uint32(len(lineCoding))
			return statusSuccess
		}
		return statusError // all non-error paths return early

	case deviceCDCACMEventSetLineCoding:
		if req, ok := param.(deviceCDCACMRequestParam); ok {
			// verify receiver configuration
			if int(uart.port) >= len(deviceCDCACMLineCoding) || uart.config == 0 ||
				int(uart.config) > len(deviceCDCACMLineCoding[uart.port]) {
				return statusNotSupported
			}
			lineCoding := deviceCDCACMLineCoding[uart.port][uart.config-1]
			if req.isSetup {
				*(req.buffer) = lineCoding
			} else {
				*(req.length) = uint32(len(lineCoding))
			}
			return statusSuccess
		}
		return statusError // all non-error paths return early

	case deviceCDCACMEventSetControlLineState:
		if req, ok := param.(deviceCDCACMRequestParam); ok {
			// verify receiver configuration
			if int(uart.port) >= len(configDeviceCDCACM) || uart.config == 0 ||
				int(uart.config) > len(configDeviceCDCACM[uart.port]) {
				return statusNotSupported
			}
			// get platform configuration
			system := configDeviceCDCACM[uart.port][uart.config-1]

			uart.acm.info.dteStatus = uint8(req.setupValue)

			// activate/deactivate Tx carrier
			if 0 != uart.acm.info.dteStatus&deviceCDCControlSigBitmapCarrierActivation {
				uart.acm.info.uartState |= deviceCDCUARTStateTxCarrier
			} else {
				uart.acm.info.uartState &^= deviceCDCUARTStateTxCarrier
			}

			// activate/deactivate DTE and carrier
			if 0 != uart.acm.info.dteStatus&deviceCDCControlSigBitmapDTEPresence {
				uart.acm.info.uartState |= deviceCDCUARTStateRxCarrier
				uart.acm.info.dtePresent = true // serial device is now open on host
			} else {
				uart.acm.info.uartState &^= deviceCDCUARTStateRxCarrier
				uart.acm.info.dtePresent = false // serial device is now closed on host
			}

			uart.acm.info.serialState[0] = deviceCDCACMRequestNotify           // bmRequestType
			uart.acm.info.serialState[1] = deviceCDCNotifySerialState          // bNotification
			uart.acm.info.serialState[2] = 0                                   // wValue (lo)
			uart.acm.info.serialState[3] = 0                                   // wValue (hi)
			uart.acm.info.serialState[4] = uint8(req.interfaceIndex)           // wIndex (lo)
			uart.acm.info.serialState[5] = uint8(req.interfaceIndex >> 8)      // wIndex (hi)
			uart.acm.info.serialState[6] = deviceCDCACMInfoUARTBitmapSize      // wLength (lo)
			uart.acm.info.serialState[7] = 0                                   // wLength (hi)
			uart.acm.info.serialState[8] = uint8(uart.acm.info.uartState)      // UART bitmap (lo)
			uart.acm.info.serialState[9] = uint8(uart.acm.info.uartState >> 8) // UART bitmap (hi)

			s := statusSuccess
			if !uart.acm.hasSentState {
				s = uart.acm.send(system.commInterruptInEndpoint,
					uart.acm.info.serialState[:], deviceCDCACMSerialStateSize)
				uart.acm.hasSentState = true
			}

			// update status
			if 0 != uart.acm.info.dteStatus&deviceCDCControlSigBitmapCarrierActivation {
				// carrier activated
			}
			if 0 != uart.acm.info.dteStatus&deviceCDCControlSigBitmapDTEPresence {
				// DTE activated
				if uart.attached {
					uart.transact = true
				}
			} else {
				// DTE deactivated
				if uart.attached {
					uart.transact = false
				}
			}

			return s
		}
		return statusError // all non-error paths return early

	case deviceCDCACMEventSendBreak:
		return statusNotImplemented

	default:
		return statusInvalidParameter
	}
}
