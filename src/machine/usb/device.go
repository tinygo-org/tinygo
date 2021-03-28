package usb

// Unexported USB device type definitions.
type (
	// deviceClassDriver defines the class driver interface.
	//
	// This interface is used internally to abstract communication between a USB
	// device and the device class (e.g., CDC, HID, MSC, etc.) in which it is
	// configured.
	deviceClassDriver interface {
		init(device *device, config *deviceClassConfig, id uint8) status // Class driver initialization- entry  of the class driver
		deinit() status                                                  // Class driver de-initialization
		event(event deviceClassEventID, param interface{}) status        // Class driver event callback
		send(ep uint8, buffer []uint8, length uint32) status             // Class driver send to endpoint
		receive(ep uint8, buffer []uint8, length uint32) status          // Class driver receive from endpoint
	}

	// deviceClassEventHandler defines the methods required to receive all device-
	// and class-level event notifications on a given USB port.
	//
	// This interface is intended as the primary communication mechanism between
	// a USB device (of any class) and the application layer using that device.
	// Thus, it is meant to be implemented internally by one of the application's
	// USB handlers (e.g., a serial UART driver using the CDC-ACM device class).
	deviceClassEventHandler interface {
		deviceEvent(ev deviceEventID, param interface{}) status
		classEvent(ev uint32, param interface{}) status
	}

	// deviceEndpointController defines the method(s) required
	deviceEndpointController interface {
		controlEndpoint(message deviceEndpointControlMessage, param interface{}) status
	}

	deviceEventFunc           func(ev deviceEventID, param interface{}) status
	deviceRequestCallbackFunc func(setup *deviceSetup, buffer *[]uint8, length *uint32) status

	deviceNotificationID    uint8
	deviceControlID         uint8
	deviceStatusID          uint8
	deviceStateID           uint8
	deviceEndpointStatusID  uint8
	deviceClassEventID      uint8
	deviceEventID           uint8
	deviceClassID           uint8
	deviceControlRWSequence uint8

	deviceDescriptorString *[]uint8

	deviceDescriptorLanguage struct {
		pString []deviceDescriptorString
		ident   uint16
	}

	deviceDescriptor struct {
		pDevice  *[]uint8
		pConfig  *[]uint8
		language []deviceDescriptorLanguage
	}

	deviceEndpointConfig struct {
		maxPacketSize uint16 // Endpoint maximum packet size
		address       uint8  // Endpoint address
		transferType  uint8  // Endpoint transfer type
		zlt           uint8  // ZLT flag
		interval      uint8  // Endpoint interval
	}

	deviceEndpointStatus struct {
		address uint8  // Endpoint address
		status  uint16 // Endpoint status (idle or stalled)
	}

	deviceNotification struct {
		buffer  []uint8              // Transferred buffer
		length  uint32               // Transferred data length
		code    deviceNotificationID // Notification code
		isSetup bool                 // Is in a setup phase
	}

	// deviceSetup contains the setup information for a USB device.
	deviceSetupBitmap uint64
	deviceSetupBuffer [deviceSetupSize]uint8
	deviceSetup       struct {
		bmRequestType uint8  //  8, 1 (bits, bytes)
		bRequest      uint8  //  8, 1
		wValue        uint16 // 16, 2
		wIndex        uint16 // 16, 2
		wLength       uint16 // 16, 2 (= 64 bits, 8 bytes)
	}

	deviceEndpointControlMessage struct {
		buffer  []uint8 // Transferred buffer
		length  uint32  // Transferred data length
		isSetup bool    // Is in a setup phase
	}

	deviceEndpointControlList [2 * configDeviceMaxEndpoints]deviceEndpointControl
	deviceEndpointControl     struct {
		handler deviceEndpointController
		param   interface{} // Parameter for callback function
		isBusy  bool
	}

	// deviceEndpoint contains the information for a USB device endpoint.
	deviceEndpoint struct {
		address       uint8  // Endpoint address
		transferType  uint8  // Endpoint transfer type
		maxPacketSize uint16 // Endpoint maximum packet size
		interval      uint8  // Endpoint interval
	}

	// deviceInterface contains the endpoints and class-specific information for
	// a USB device interface.
	deviceInterface struct {
		alternateSetting uint8            // Alternate setting number
		endpoint         []deviceEndpoint // Endpoints of the interface
		classSpecific    interface{}      // Class specific structure handle
	}

	// deviceClassInterface contains the USB device class details, including
	// all of its device interfaces.
	deviceClassInterface struct {
		classCode       uint8             // Class code of the interface
		subclassCode    uint8             // Subclass code of the interface
		protocolCode    uint8             // Protocol code of the interface
		interfaceNumber uint8             // Interface number
		deviceInterface []deviceInterface // Interface structure list
	}

	// deviceClassInfo contains the USB device class ID and all of its class
	// interfaces.
	deviceClassInfo struct {
		classID       deviceClassID          // Class type
		interfaceList []deviceClassInterface // Interfaces of the class
	}

	// deviceClassConfig contains the configuration for a USB device class.
	deviceClassConfig struct {
		driver deviceClassDriver // USB device class driver interface
		info   deviceClassInfo   // Detailed information of the class
	}

	// deviceClass contains common device class state information.
	deviceClass struct {
		device            *device                 // USB device handle
		config            []deviceClassConfig     // USB device class configuration list
		handler           deviceClassEventHandler // application callback
		setupBuffer       []uint8                 // Setup packet data buffer
		transcationBuffer uint16                  // Get status/configuration, get/set interface, get sync frame
	}

	device struct {
		port            uint8                     // USB port (core index)
		controller      deviceController          // Controller interface
		class           *deviceClass              // USB device class
		endpointControl deviceEndpointControlList // Endpoint callback function structure
		deviceAddress   uint8                     // Current device address
		state           deviceStateID             // Current device state
		isResetting     bool                      // Is doing device reset or not
		hwTick          int64                     // (volatile) Current hw tick (ms)
	}

	// // 			This structure is used to pass the control request information.
	// // 			The structure is used in following two cases.
	// // 			1. Case one, the host wants to send data to the device in the control data stage: @n
	// // 			        a. If a setup packet is received, the structure is used to pass the setup packet data and wants to get the
	// // 			buffer to receive data sent from the host.
	// // 			           The field isSetup is 1.
	// // 			           The length is the requested buffer length.
	// // 			           The buffer is filled by the class or application by using the valid buffer address.
	// // 			           The setup is the setup packet address.
	// // 			        b. If the data received is sent by the host, the structure is used to pass the data buffer address and the
	// // 			data
	// // 			length sent by the host.
	// // 			           In this way, the field isSetup is 0.
	// // 			           The buffer is the address of the data sent from the host.
	// // 			           The length is the received data length.
	// // 			           The setup is the setup packet address. @n
	// // 			2. Case two, the host wants to get data from the device in control data stage: @n
	// // 			           If the setup packet is received, the structure is used to pass the setup packet data and wants to get the
	// // 			data buffer address to send data to the host.
	// // 			           The field isSetup is 1.
	// // 			           The length is the requested data length.
	// // 			           The buffer is filled by the class or application by using the valid buffer address.
	// // 			           The setup is the setup packet address.
	deviceControlRequest struct {
		setup   deviceSetup // Setup data
		buffer  []uint8     // Buffer
		length  uint32      // Buffer length or requested length
		isSetup bool        // Indicates whether a setup packet is received
	}

	// // deviceGetDescriptorCommon contains the result of a control request for:
	// // get descriptor common
	// deviceGetDescriptorCommon struct {
	// 	buffer []uint8 // Buffer
	// 	length uint32  // Buffer length
	// }

	// deviceGetDeviceDescriptor contains the result of a control request for:
	// get device descriptor
	deviceGetDeviceDescriptor struct {
		buffer []uint8 // Buffer
		length uint32  // Buffer length
	}

	// // deviceGetDeviceQualifierDescriptor contains the result of a control
	// // request for: get device qualifier descriptor
	// deviceGetDeviceQualifierDescriptor struct {
	// 	buffer []uint8 // Buffer
	// 	length uint32  // Buffer length
	// }

	// deviceGetConfigurationDescriptor contains the result of a control request
	// for: get configuration descriptor
	deviceGetConfigurationDescriptor struct {
		buffer        []uint8 // Buffer
		length        uint32  // Buffer length
		configuration uint8   // The configuration number
	}

	// // deviceGetBOSDescriptor contains the result of a control request for: get
	// // bos descriptor
	// deviceGetBOSDescriptor struct {
	// 	buffer []uint8 // Buffer
	// 	length uint32  // Buffer length
	// }

	// deviceGetStringDescriptor contains the result of a control request for:
	// get string descriptor
	deviceGetStringDescriptor struct {
		buffer      []uint8 // Buffer
		length      uint32  // Buffer length
		languageID  uint16  // Language ID
		stringIndex uint8   // String index
	}

	// // deviceGetHIDDescriptor contains the result of a control request for: get
	// // HID descriptor
	// deviceGetHIDDescriptor struct {
	// 	buffer          []uint8 // Buffer
	// 	length          uint32  // Buffer length
	// 	interfaceNumber uint8   // The interface number
	// }

	// // deviceGetHIDReportDescriptor contains the result of a control request for:
	// // get HID report descriptor
	// deviceGetHIDReportDescriptor struct {
	// 	buffer          []uint8 // Buffer
	// 	length          uint32  // Buffer length
	// 	interfaceNumber uint8   // The interface number
	// }

	// // deviceGetHIDPhysicalDescriptor contains the result of a control request
	// // for: get HID physical descriptor
	// deviceGetHIDPhysicalDescriptor struct {
	// 	buffer          []uint8 // Buffer
	// 	length          uint32  // Buffer length
	// 	index           uint8   // Physical index
	// 	interfaceNumber uint8   // The interface number
	// }

)

// Unexported enumerated constant values for USB device.
const (
	deviceNotifyBusReset          deviceNotificationID = iota + 0x10 // Reset signal detected
	deviceNotifySuspend                                              // Suspend signal detected
	deviceNotifyResume                                               // Resume signal detected
	deviceNotifyLPMSleep                                             // LPM signal detected
	deviceNotifyLPMResume                                            // Resume signal detected
	deviceNotifyError                                                // Errors happened in bus
	deviceNotifyDetach                                               // Device disconnected from a host
	deviceNotifyAttach                                               // Device connected to a host
	deviceNotifyDCDDetectFinished                                    // Device charger detection finished
)

const (
	deviceControlRun                 deviceControlID = iota // Enable the device functionality
	deviceControlStop                                       // Disable the device functionality
	deviceControlEndpointInit                               // Initialize a specified endpoint
	deviceControlEndpointDeinit                             // De-initialize a specified endpoint
	deviceControlEndpointStall                              // Stall a specified endpoint
	deviceControlEndpointUnstall                            // Un-stall a specified endpoint
	deviceControlGetDeviceStatus                            // Get device status
	deviceControlGetEndpointStatus                          // Get endpoint status
	deviceControlSetDeviceAddress                           // Set device address
	deviceControlGetSynchFrame                              // Get current frame
	deviceControlResume                                     // Drive controller to generate a resume signal in USB bus
	deviceControlSleepResume                                // Drive controller to generate a LPM resume signal in USB bus
	deviceControlSuspend                                    // Drive controller to enter into suspend mode
	deviceControlSleep                                      // Drive controller to enter into sleep mode
	deviceControlSetDefaultStatus                           // Set controller to default status
	deviceControlGetSpeed                                   // Get current speed
	deviceControlGetOTGStatus                               // Get OTG status
	deviceControlSetOTGStatus                               // Set OTG status
	deviceControlSetTestMode                                // Drive xCHI into test mode
	deviceControlGetRemoteWakeUp                            // Get flag of LPM Remote Wake-up Enabled by USB host.
	deviceControlDCDDisable                                 // disable dcd module function.
	deviceControlDCDEnable                                  // enable dcd module function.
	deviceControlPreSetDeviceAddress                        // Pre set device address
	deviceControlUpdateHwTick                               // update hardware tick
)

const (
	deviceStatusTestMode       deviceStatusID = iota + 1 // Test mode
	deviceStatusSpeed                                    // Current speed
	deviceStatusOTG                                      // OTG status
	deviceStatusDevice                                   // Device status
	deviceStatusEndpoint                                 // Endpoint state usb_device_endpoint_status_t
	deviceStatusDeviceState                              // Device state
	deviceStatusAddress                                  // Device address
	deviceStatusSynchFrame                               // Current frame
	deviceStatusBus                                      // Bus status
	deviceStatusBusSuspend                               // Bus suspend
	deviceStatusBusSleep                                 // Bus suspend
	deviceStatusBusResume                                // Bus resume
	deviceStatusRemoteWakeup                             // Remote wakeup state
	deviceStatusBusSleepResume                           // Bus resume
)

const (
	deviceStateConfigured deviceStateID = iota // Device state, Configured
	deviceStateAddress                         // Device state, Address
	deviceStateDefault                         // Device state, Default
	deviceStateAddressing                      // Device state, Address setting
	deviceStateTestMode                        // Device state, Test mode
	deviceStateInit                            // Device state, initializing
)

const (
	deviceEndpointStateIdle    deviceEndpointStatusID = iota // Endpoint state, idle
	deviceEndpointStateStalled                               // Endpoint state, stalled
)

const (
	deviceEventBusReset                     deviceEventID = iota + 1 // USB bus reset signal detected
	deviceEventSuspend                                               // USB bus suspend signal detected
	deviceEventResume                                                // USB bus resume signal detected. The resume signal is driven by itself or a host
	deviceEventSleeped                                               // USB bus LPM suspend signal detected
	deviceEventLPMResume                                             // USB bus LPM resume signal detected. The resume signal is driven by itself or a host
	deviceEventError                                                 // An error is happened in the bus.
	deviceEventDetach                                                // USB device is disconnected from a host.
	deviceEventAttach                                                // USB device is connected to a host.
	deviceEventSetConfiguration                                      // Set configuration.
	deviceEventSetInterface                                          // Set interface.
	deviceEventGetDeviceDescriptor                                   // Get device descriptor.
	deviceEventGetConfigurationDescriptor                            // Get configuration descriptor.
	deviceEventGetStringDescriptor                                   // Get string descriptor.
	deviceEventGetHIDDescriptor                                      // Get HID descriptor.
	deviceEventGetHIDReportDescriptor                                // Get HID report descriptor.
	deviceEventGetHIDPhysicalDescriptor                              // Get HID physical descriptor.
	deviceEventGetBOSDescriptor                                      // Get configuration descriptor.
	deviceEventGetDeviceQualifierDescriptor                          // Get device qualifier descriptor.
	deviceEventVendorRequest                                         // Vendor request.
	deviceEventSetRemoteWakeup                                       // Enable or disable remote wakeup function.
	deviceEventGetConfiguration                                      // Get current configuration index
	deviceEventGetInterface                                          // Get current interface alternate setting value
	deviceEventSetBHNPEnable                                         // Enable or disable BHNP.
	deviceEventDCDDetectionfinished                                  // The DCD detection finished
)

const (
	deviceClassInvalid deviceClassID = iota
	deviceClassHID
	deviceClassCDC
	deviceClassMSC
	deviceClassAudio
	deviceClassPHDC
	deviceClassVideo
	deviceClassPrinter
	deviceClassDFU
	deviceClassCCID
)

const (
	deviceClassEventInvalid deviceClassEventID = iota
	deviceClassEventClassRequest
	deviceClassEventDeviceReset
	deviceClassEventSetConfiguration
	deviceClassEventSetInterface
	deviceClassEventSetEndpointHalt
	deviceClassEventClearEndpointHalt
)

const (
	deviceControlPipeSetupStage  deviceControlRWSequence = iota // Setup stage
	deviceControlPipeDataStage                                  // Data stage
	deviceControlPipeStatusStage                                // status stage
)

// Sizes of various device structures and buffers.
const (
	deviceSetupSize = 8 // deviceSetup struct (bytes)
)

var (
	deviceClassInstance       [configDeviceCount]deviceClass
	deviceSetupBufferInstance [configDeviceCount]deviceSetupBuffer

	deviceEndpointControlInConfig = deviceEndpointConfig{
		maxPacketSize: configDeviceControllerMaxPacketSize,
		address:       uint8(specEndpointControl | specDescriptorEndpointAddressDirectionIn),
		transferType:  specEndpointControl,
		zlt:           1,
		interval:      0,
	}
	deviceEndpointControlOutConfig = deviceEndpointConfig{
		maxPacketSize: configDeviceControllerMaxPacketSize,
		address:       uint8(specEndpointControl | specDescriptorEndpointAddressDirectionOut),
		transferType:  specEndpointControl,
		zlt:           1,
		interval:      0,
	}
)

func (d *device) init(port uint8) (s status) {

	// initialize device
	d.port = port
	d.controller = d.initController() // obtain a device controller handle
	d.deviceAddress = 0
	d.state = deviceStateDefault
	d.isResetting = false
	d.hwTick = 0
	for i := range d.endpointControl {
		d.endpointControl[i].handler = nil
		d.endpointControl[i].param = nil
		d.endpointControl[i].isBusy = false
	}

	// initialize platform via device controller interface
	return d.controller.init()
}

func (d *device) deinit() status {

	// de=initialize device
	s := d.controller.deinit()
	d.controller = nil
	return s
}

func (d *device) initClass(id uint8, config []deviceClassConfig,
	handler deviceClassEventHandler) *deviceClass {

	// verify a device configuration was provided
	if nil == d || nil == config || len(config) == 0 {
		return nil
	}

	if 0 == id || int(id) > len(config) {
		return nil
	}

	c := &deviceClassInstance[d.port]
	b := &deviceSetupBufferInstance[d.port]

	// initialize device class
	c.device = d
	c.config = config
	c.handler = handler
	c.setupBuffer = b[:]
	c.transcationBuffer = 0

	// add a class reference to the receiver
	d.class = c

	// initialze each of the device class drivers
	for i := range c.config {
		if nil != c.config[i].driver {
			if !c.config[i].driver.init(d, &c.config[i], id).OK() {
				// remove the driver from configuration if it fails initialization
				c.config[i].driver = nil
			}
		}
	}

	return c
}

func (d *device) transfer(address uint8, buffer []uint8, length uint32) status {

	if nil == d.controller {
		return statusInvalidController
	}

	endpoint, direction := unpackEndpoint(address)
	index := (endpoint << 1) | direction

	if d.endpointControl[index].isBusy {
		return statusBusy
	}
	d.endpointControl[index].isBusy = true

	var s status
	if specDescriptorEndpointAddressDirectionIn ==
		address&specDescriptorEndpointAddressDirectionMsk {
		s = d.controller.send(address, buffer, length)
	} else {
		s = d.controller.receive(address, buffer, length)
	}
	if !s.OK() {
		d.endpointControl[index].isBusy = false
	}
	return s
}

func (d *device) send(address uint8, buffer []uint8, length uint32) status {
	return d.transfer((address&specDescriptorEndpointAddressNumberMsk)|
		(specDescriptorEndpointAddressDirectionIn), buffer, length)
}

func (d *device) receive(address uint8, buffer []uint8, length uint32) status {
	return d.transfer((address&specDescriptorEndpointAddressNumberMsk)|
		(specDescriptorEndpointAddressDirectionOut), buffer, length)
}

func (d *device) cancel(address uint8) status {
	if nil == d.controller {
		return statusInvalidController
	}
	return d.controller.cancel(address)
}

func (d *device) control(command deviceControlID, param interface{}) status {
	if nil == d.controller {
		return statusInvalidController
	}
	return d.controller.control(command, param)
}

func (d *device) initControlPipes() status {

	if s := d.initEndpoint(&deviceEndpointControlInConfig, d, d.class); !s.OK() {
		return s
	}

	if s := d.initEndpoint(&deviceEndpointControlOutConfig, d, d.class); !s.OK() {
		_ = d.deinitEndpoint(deviceEndpointControlInConfig.address)
		return s
	}

	return statusSuccess
}

func (d *device) initEndpoint(config *deviceEndpointConfig,
	handler deviceEndpointController, param interface{}) status {

	endpoint, direction := unpackEndpoint(config.address)

	if endpoint >= configDeviceMaxEndpoints {
		return statusInvalidParameter
	}

	d.endpointControl[(endpoint<<1)|direction].handler = handler
	d.endpointControl[(endpoint<<1)|direction].param = param
	d.endpointControl[(endpoint<<1)|direction].isBusy = false

	return d.control(deviceControlEndpointInit, config)
}

func (d *device) deinitEndpoint(address uint8) status {

	s := d.control(deviceControlEndpointDeinit, address)

	endpoint, direction := unpackEndpoint(address)

	if endpoint >= configDeviceMaxEndpoints {
		return statusInvalidParameter
	}

	d.endpointControl[(endpoint<<1)|direction].handler = nil
	d.endpointControl[(endpoint<<1)|direction].param = nil
	d.endpointControl[(endpoint<<1)|direction].isBusy = false

	return s
}

func (d *device) stallEndpoint(address uint8) status {

	endpoint, _ := unpackEndpoint(address)

	if endpoint >= configDeviceMaxEndpoints {
		return statusInvalidParameter
	}
	return d.control(deviceControlEndpointStall, address)
}

func (d *device) unstallEndpoint(address uint8) status {

	endpoint, _ := unpackEndpoint(address)

	if endpoint >= configDeviceMaxEndpoints {
		return statusInvalidParameter
	}
	return d.control(deviceControlEndpointUnstall, address)
}

func (d *device) feedback(setup *deviceSetup, status status, stage deviceControlRWSequence,
	buffer *[]uint8, length *uint32) status {
	direction := uint8(specIn)
	if !status.OK() {
		if (setup.bmRequestType&specRequestTypeTypeMsk) == specRequestTypeTypeStandard &&
			(setup.bmRequestType&specRequestTypeDirMsk) == specRequestTypeDirOut &&
			0 != setup.wLength && deviceControlPipeSetupStage == stage {
			direction = specOut
		}
		return d.stallEndpoint(specEndpointControl |
			(direction << specDescriptorEndpointAddressDirectionPos))
	} else {
		if *length > uint32(setup.wLength) {
			*length = uint32(setup.wLength)
		}
		s := d.send(specEndpointControl, *buffer, *length)
		if s.OK() && (setup.bmRequestType&specRequestTypeDirMsk) == specRequestTypeDirIn {
			s = d.receive(specEndpointControl, nil, 0)
		}
		return s
	}
}

func (d *device) controlEndpoint(message deviceEndpointControlMessage, param interface{}) status {

	// verify request and parameters
	if message.length == 0xFFFFFFFF || nil == param {
		return statusInvalidRequest
	}
	class, ok := param.(*deviceClass)
	if !ok {
		return statusInvalidParameter
	}
	// read current setup buffer from given deviceClass in param
	var setup deviceSetup
	setupStatus := setup.parse(class.setupBuffer)
	if !setupStatus.OK() {
		return setupStatus
	}

	// read current device state of receiver
	var state deviceStateID
	s := d.status(deviceStatusDeviceState, &state)
	if !s.OK() {
		return statusInvalidHandle
	}

	// allocate buffer for request responses
	buffer := []uint8{}
	length := uint32(0)

	if message.isSetup {
		// verify message received contains expected setup data
		if nil == message.buffer || deviceSetupSize != message.length {
			return statusInvalidRequest
		}
		// read setup data from given message buffer
		s = setup.parse(message.buffer)
		if !s.OK() {
			return statusInvalidRequest
		}
		// process message as a received setup request
		if (setup.bmRequestType & specRequestTypeTypeMsk) == specRequestTypeTypeStandard {
			// handle standard request
			if int(setup.bRequest) <= specRequestStandardSynchFrame {
				_ = d.requestStandard(&setup, &buffer, &length)
			}
		} else {
			if 0 != setup.wLength &&
				(setup.bmRequestType&specRequestTypeDirMsk) == specRequestTypeDirOut {
				if (setup.bmRequestType & specRequestTypeTypeClass) == specRequestTypeTypeClass {
					req := deviceControlRequest{
						setup:   setup,
						buffer:  nil,
						length:  uint32(setup.wLength),
						isSetup: true,
					}
					d.classEvent(deviceClassEventClassRequest, &req)
					buffer = req.buffer
					length = req.length
				} else if (setup.bmRequestType & specRequestTypeTypeVendor) == specRequestTypeTypeVendor {
					req := deviceControlRequest{
						setup:   setup,
						buffer:  nil,
						length:  uint32(setup.wLength),
						isSetup: true,
					}
					d.event(deviceEventVendorRequest, &req)
					buffer = req.buffer
					length = req.length
				}
				if s.OK() {
					return d.receive(specEndpointControl, buffer, uint32(setup.wLength))
				}
			} else {
				if (setup.bmRequestType & specRequestTypeTypeClass) == specRequestTypeTypeClass {
					req := deviceControlRequest{
						setup:   setup,
						buffer:  nil,
						length:  uint32(setup.wLength),
						isSetup: true,
					}
					d.classEvent(deviceClassEventClassRequest, &req)
					buffer = req.buffer
					length = req.length
				} else if (setup.bmRequestType & specRequestTypeTypeVendor) == specRequestTypeTypeVendor {
					req := deviceControlRequest{
						setup:   setup,
						buffer:  nil,
						length:  uint32(setup.wLength),
						isSetup: true,
					}
					d.event(deviceEventVendorRequest, &req)
					buffer = req.buffer
					length = req.length
				}
			}
		}
		s = d.feedback(&setup, s, deviceControlPipeSetupStage, &buffer, &length)
	} else if deviceStateAddressing == state {
		// handle standard request
		if int(setup.bRequest) <= specRequestStandardSynchFrame {
			_ = d.requestStandard(&setup, &buffer, &length)
		}
	} else if 0 != message.length && 0 != setup.wLength &&
		(setup.bmRequestType&specRequestTypeDirMsk) == specRequestTypeDirOut {
		if setup.bmRequestType&specRequestTypeTypeClass == specRequestTypeTypeClass {
			req := deviceControlRequest{
				setup:   setup,
				buffer:  message.buffer,
				length:  message.length,
				isSetup: false,
			}
			d.classEvent(deviceClassEventClassRequest, &req)
		} else if setup.bmRequestType&specRequestTypeTypeVendor == specRequestTypeTypeVendor {
			req := deviceControlRequest{
				setup:   setup,
				buffer:  message.buffer,
				length:  message.length,
				isSetup: false,
			}
			d.event(deviceEventVendorRequest, &req)
		}
		s = d.feedback(&setup, s, deviceControlPipeDataStage, &buffer, &length)
	}

	return s
}

func (d *device) requestStandard(
	setup *deviceSetup, buffer *[]uint8, length *uint32) status {
	switch setup.bRequest {
	case specRequestStandardGetStatus:
		return d.requestGetStatus(setup, buffer, length)
	case specRequestStandardClearFeature, specRequestStandardSetFeature:
		return d.requestSetClearFeature(setup, buffer, length)
	case specRequestStandardSetAddress:
		return d.requestSetAddress(setup, buffer, length)
	case specRequestStandardGetDescriptor:
		return d.requestGetDescriptor(setup, buffer, length)
	case specRequestStandardGetConfiguration:
		return d.requestGetConfiguration(setup, buffer, length)
	case specRequestStandardSetConfiguration:
		return d.requestSetConfiguration(setup, buffer, length)
	case specRequestStandardGetInterface:
		return d.requestGetInterface(setup, buffer, length)
	case specRequestStandardSetInterface:
		return d.requestSetInterface(setup, buffer, length)
	case specRequestStandardSynchFrame:
		return d.requestSynchFrame(setup, buffer, length)
	default:
		return statusInvalidRequest
	}
}

func (d *device) requestGetStatus(
	setup *deviceSetup, buffer *[]uint8, length *uint32) status {
	return statusSuccess
}

func (d *device) requestSetClearFeature(
	setup *deviceSetup, buffer *[]uint8, length *uint32) status {
	return statusSuccess
}

func (d *device) requestSetAddress(
	setup *deviceSetup, buffer *[]uint8, length *uint32) status {
	return statusSuccess
}

func (d *device) requestGetDescriptor(
	setup *deviceSetup, buffer *[]uint8, length *uint32) status {
	return statusSuccess
}

func (d *device) requestGetConfiguration(
	setup *deviceSetup, buffer *[]uint8, length *uint32) status {
	return statusSuccess
}

func (d *device) requestSetConfiguration(
	setup *deviceSetup, buffer *[]uint8, length *uint32) status {
	return statusSuccess
}

func (d *device) requestGetInterface(
	setup *deviceSetup, buffer *[]uint8, length *uint32) status {
	return statusSuccess
}

func (d *device) requestSetInterface(
	setup *deviceSetup, buffer *[]uint8, length *uint32) status {
	return statusSuccess
}

func (d *device) requestSynchFrame(
	setup *deviceSetup, buffer *[]uint8, length *uint32) status {
	return statusSuccess
}

func (d *device) classEvent(ev deviceClassEventID, param interface{}) {
	if nil != d.class && nil != d.class.config {
		// route the event to all class configurations
		for i := range d.class.config {
			_ = d.class.config[i].driver.event(ev, param)
		}
	}
}

func (d *device) event(ev deviceEventID, param interface{}) {

	switch ev {
	case deviceEventBusReset:
		// initialize control pipes
		d.initControlPipes()

		// notify all classes of a bus reset signal
		for i := range d.class.config {
			_ = d.class.config[i].driver.event(deviceClassEventDeviceReset, d.class)
		}
	}
	// notify the application driver of all events
	if nil != d.class.handler {
		_ = d.class.handler.deviceEvent(ev, param)
	}
}

func (d *device) notify(message deviceNotification) {

	switch message.code {
	case deviceNotifyBusReset:
		d.notifyReset(message)

	default:
		endpoint, direction := unpackEndpoint(uint8(message.code))

		if endpoint < configDeviceMaxEndpoints {
			if nil != d.endpointControl[(endpoint<<1)|direction].handler {
				if message.isSetup {
					d.endpointControl[0].isBusy = false
					d.endpointControl[1].isBusy = false
				} else {
					d.endpointControl[(endpoint<<1)|direction].isBusy = false
				}
				// call endpoint callback
				_ = d.endpointControl[(endpoint<<1)|direction].handler.controlEndpoint(
					deviceEndpointControlMessage{
						buffer:  message.buffer,
						length:  message.length,
						isSetup: message.isSetup,
					},
					d.endpointControl[(endpoint<<1)|direction].param,
				)
			}
		}
	}
}

func (d *device) notifyReset(message deviceNotification) {

	d.isResetting = true
	_ = d.control(deviceControlSetDefaultStatus, nil)

	d.state = deviceStateDefault
	d.deviceAddress = 0

	for count := 0; count < 2*configDeviceMaxEndpoints; count++ {
		d.endpointControl[count].handler = nil
		d.endpointControl[count].param = nil
		d.endpointControl[count].isBusy = false
	}

	d.event(deviceEventBusReset, nil)
	d.isResetting = false
}

func (d *device) status(deviceStatus deviceStatusID, param interface{}) status {

	if nil == param {
		return statusInvalidParameter
	}

	switch deviceStatus {
	case deviceStatusSpeed:
		return d.control(deviceControlGetSpeed, param)

	case deviceStatusOTG:
		return d.control(deviceControlGetOTGStatus, param)

	case deviceStatusDeviceState:
		if state, ok := param.(*deviceStateID); ok {
			*state = d.state
			return statusSuccess
		}
		return statusInvalidParameter

	case deviceStatusAddress:
		if address, ok := param.(*uint8); ok {
			*address = d.deviceAddress
			return statusSuccess
		}
		return statusInvalidParameter

	case deviceStatusDevice:
		return d.control(deviceControlGetDeviceStatus, param)

	case deviceStatusEndpoint:
		return d.control(deviceControlGetEndpointStatus, param)

	case deviceStatusSynchFrame:
		return d.control(deviceControlGetSynchFrame, param)

	default:
		return statusInvalidParameter
	}
}

func (d *device) setStatus(deviceStatus deviceStatusID, param interface{}) status {

	switch deviceStatus {
	case deviceStatusOTG:
		return d.control(deviceControlSetOTGStatus, param)

	case deviceStatusDeviceState:
		if state, ok := param.(deviceStateID); ok {
			d.state = state
			return statusSuccess
		}
		return statusInvalidParameter

	case deviceStatusAddress:
		if d.state != deviceStateAddressing {
			if address, ok := param.(uint8); ok {
				d.deviceAddress = address
				d.state = deviceStateAddressing
				return d.control(deviceControlPreSetDeviceAddress, d.deviceAddress)
			}
			return statusInvalidParameter
		}
		return d.control(deviceControlSetDeviceAddress, d.deviceAddress)

	case deviceStatusBusResume:
		return d.control(deviceControlResume, param)

	case deviceStatusBusSleepResume:
		return d.control(deviceControlSleepResume, param)

	case deviceStatusBusSuspend:
		return d.control(deviceControlSuspend, param)

	case deviceStatusBusSleep:
		return d.control(deviceControlSleep, param)

	default:
		return statusInvalidParameter
	}
}

func (d *device) busSpeed(speed *uint8) status {
	return d.status(deviceStatusSpeed, speed)
}

func (d *device) setBusSpeed(speed uint8) status {
	// TODO
	return statusSuccess
}

func (d *device) deviceDescriptor(desc *deviceGetDeviceDescriptor) status {
	if int(d.port) < len(configDeviceCDCACMDescriptor) {
		if nil != configDeviceCDCACMDescriptor[d.port].pDevice {
			desc.buffer = *configDeviceCDCACMDescriptor[d.port].pDevice
			desc.length = specDescriptorLengthDevice
			return statusSuccess
		}
	}
	return statusInvalidHandle
}

func (d *device) configurationDescriptor(desc *deviceGetConfigurationDescriptor) status {
	if int(d.port) < len(configDeviceCDCACMDescriptor) {
		if nil != configDeviceCDCACMDescriptor[d.port].pConfig {
			desc.buffer = *configDeviceCDCACMDescriptor[d.port].pConfig
			desc.length = uint32(descriptorConfigurationCDCACMSize)
			return statusSuccess
		}
	}
	return statusInvalidHandle
}

func (d *device) stringDescriptor(desc *deviceGetStringDescriptor) status {
	if 0 == desc.stringIndex {
		desc.buffer = descriptorStringCDCACMLanguage
		desc.length = uint32(len(descriptorStringCDCACMLanguage))
		return statusSuccess
	}
	if int(d.port) < len(configDeviceCDCACMDescriptor) {
		for _, lang := range configDeviceCDCACMDescriptor[d.port].language {
			if desc.languageID == lang.ident {
				if int(desc.stringIndex) < len(lang.pString) {
					desc.buffer = *lang.pString[desc.stringIndex]
					desc.length = uint32(len(desc.buffer))
				}
			}
		}
	}
	return statusInvalidRequest
}

func (s deviceSetup) pack() deviceSetupBitmap {
	return deviceSetupBitmap(
		((uint64(s.bmRequestType) & 0xFF) << 0) | // uint8  // 8 (bits)
			((uint64(s.bRequest) & 0xFF) << 8) | // uint8  // 8
			((uint64(s.wValue) & 0xFFFF) << 16) | // uint16 // 16
			((uint64(s.wIndex) & 0xFFFF) << 32) | // uint16 // 16
			((uint64(s.wLength) & 0xFFFF) << 48)) // uint16 // 16 (= 64 bits)
}

func (s deviceSetup) bytes() deviceSetupBuffer {
	return [deviceSetupSize]uint8{
		s.bmRequestType,
		s.bRequest,
		// for 16-bit words: low-byte is at index N, high-byte is at N+1
		uint8(s.wValue), uint8(s.wValue >> 8),
		uint8(s.wIndex), uint8(s.wIndex >> 8),
		uint8(s.wLength), uint8(s.wLength >> 8),
	}
}

func (s deviceSetupBitmap) bytes() deviceSetupBuffer {
	return [deviceSetupSize]uint8{
		uint8(s >> 0), uint8(s >> 8), uint8(s >> 16), uint8(s >> 24),
		uint8(s >> 32), uint8(s >> 40), uint8(s >> 48), uint8(s >> 56),
	}
}

func (s *deviceSetup) parse(buffer []uint8) status {
	if nil == buffer || len(buffer) < deviceSetupSize {
		return statusInvalidParameter
	}
	s.bmRequestType = buffer[0]
	s.bRequest = buffer[1]
	s.wValue = (uint16(buffer[3]) << 8) | uint16(buffer[2])
	s.wIndex = (uint16(buffer[5]) << 8) | uint16(buffer[4])
	s.wLength = (uint16(buffer[7]) << 8) | uint16(buffer[6])
	return statusSuccess
}
