package usb

const (
	// Communication Class
	deviceCDCCommClass = 0x02
	// Data Class
	deviceCDCDataClass = 0x0A

	// Communication Class SubClass Codes
	deviceCDCDirectLineControlModel         = 0x01
	deviceCDCAbstractControlModel           = 0x02
	deviceCDCTelephoneControlModel          = 0x03
	deviceCDCMultiChannelControlModel       = 0x04
	deviceCDCCAPIControlMopdel              = 0x05
	deviceCDCEthernetNetworkingControlModel = 0x06
	deviceCDCATMNetworkingControlModel      = 0x07
	deviceCDCWirelessHandsetControlModel    = 0x08
	deviceCDCDeviceManagement               = 0x09
	deviceCDCMobileDirectLineModel          = 0x0A
	deviceCDCOBEX                           = 0x0B
	deviceCDCEthernetEmulationModel         = 0x0C

	// Communication Class Protocol Codes
	deviceCDCNoClassSpecificProtocol   = 0x00 // also for Data Class Protocol Code
	deviceCDCAT250Protocol             = 0x01
	deviceCDCATPCCA101Protocol         = 0x02
	deviceCDCATPCCA101AnnexO           = 0x03
	deviceCDCATGSM707                  = 0x04
	deviceCDCAT3GPP27007               = 0x05
	deviceCDCATTIACDMA                 = 0x06
	deviceCDCEthernetEmulationProtocol = 0x07
	deviceCDCExternalProtocol          = 0xFE
	deviceCDCVendorSpecific            = 0xFF // also for Data Class Protocol Code

	// Data Class Protocol Codes
	deviceCDCPyhsicalInterfaceProtocol = 0x30
	deviceCDCHDLCProtocol              = 0x31
	deviceCDCTransparentProtocol       = 0x32
	deviceCDCManagementProtocol        = 0x50
	deviceCDCDataLinkQ931Protocol      = 0x51
	deviceCDCDataLinkQ921Protocol      = 0x52
	deviceCDCDataCompressionV42BIS     = 0x90
	deviceCDCEuroISDNProtocol          = 0x91
	deviceCDCRateAdaptionISDNV24       = 0x92
	deviceCDCCAPICommands              = 0x93
	deviceCDCHostBasedDriver           = 0xFD
	deviceCDCUnitFunctional            = 0xFE

	// Descriptor SubType in Communications Class Functional Descriptors
	deviceCDCHeaderFuncDesc             = 0x00
	deviceCDCCallManagementFuncDesc     = 0x01
	deviceCDCAbstractControlFuncDesc    = 0x02
	deviceCDCDirectLineFuncDesc         = 0x03
	deviceCDCTelephoneRingerFuncDesc    = 0x04
	deviceCDCTelephoneReportFuncDesc    = 0x05
	deviceCDCUnionFuncDesc              = 0x06
	deviceCDCCountrySelectFuncDesc      = 0x07
	deviceCDCTelephoneModesFuncDesc     = 0x08
	deviceCDCTerminalFuncDesc           = 0x09
	deviceCDCNetworkChannelFuncDesc     = 0x0A
	deviceCDCProtocolUnitFuncDesc       = 0x0B
	deviceCDCExtensionUnitFuncDesc      = 0x0C
	deviceCDCMultiChannelFuncDesc       = 0x0D
	deviceCDCCAPIControlFuncDesc        = 0x0E
	deviceCDCEthernetNetworkingFuncDesc = 0x0F
	deviceCDCATMNetworkingFuncDesc      = 0x10
	deviceCDCWirelessControlFuncDesc    = 0x11
	deviceCDCMobileDirectLineFuncDesc   = 0x12
	deviceCDCMDLMDetailFuncDesc         = 0x13
	deviceCDCDeviceManagementFuncDesc   = 0x14
	deviceCDCOBEXFuncDesc               = 0x15
	deviceCDCCommandSetFuncDesc         = 0x16
	deviceCDCCommandSetDetailFuncDesc   = 0x17
	deviceCDCTelephoneControlFuncDesc   = 0x18
	deviceCDCOBEXServiceIDFuncDesc      = 0x19

	deviceCDCRequestSendEncapsulatedCommand     = 0x00 // CDC request SEND_ENCAPSULATED_COMMAND
	deviceCDCRequestGetEncapsulatedResponse     = 0x01 // CDC request GET_ENCAPSULATED_RESPONSE
	deviceCDCRequestSetCommFeature              = 0x02 // CDC request SET_COMM_FEATURE
	deviceCDCRequestGetCommFeature              = 0x03 // CDC request GET_COMM_FEATURE
	deviceCDCRequestClearCommFeature            = 0x04 // CDC request CLEAR_COMM_FEATURE
	deviceCDCRequestSetAuxLineState             = 0x10 // CDC request SET_AUX_LINE_STATE
	deviceCDCRequestSetHookState                = 0x11 // CDC request SET_HOOK_STATE
	deviceCDCRequestPulseSetup                  = 0x12 // CDC request PULSE_SETUP
	deviceCDCRequestSendPulse                   = 0x13 // CDC request SEND_PULSE
	deviceCDCRequestSetPulseTime                = 0x14 // CDC request SET_PULSE_TIME
	deviceCDCRequestRingAuxJack                 = 0x15 // CDC request RING_AUX_JACK
	deviceCDCRequestSetLineCoding               = 0x20 // CDC request SET_LINE_CODING
	deviceCDCRequestGetLineCoding               = 0x21 // CDC request GET_LINE_CODING
	deviceCDCRequestSetControlLineState         = 0x22 // CDC request SET_CONTROL_LINE_STATE
	deviceCDCRequestSendBreak                   = 0x23 // CDC request SEND_BREAK
	deviceCDCRequestSetRingerParams             = 0x30 // CDC request SET_RINGER_PARAMS
	deviceCDCRequestGetRingerParams             = 0x31 // CDC request GET_RINGER_PARAMS
	deviceCDCRequestSetOperationParam           = 0x32 // CDC request SET_OPERATION_PARAM
	deviceCDCRequestGetOperationParam           = 0x33 // CDC request GET_OPERATION_PARAM
	deviceCDCRequestSetLineParams               = 0x34 // CDC request SET_LINE_PARAMS
	deviceCDCRequestGetLineParams               = 0x35 // CDC request GET_LINE_PARAMS
	deviceCDCRequestDialDigits                  = 0x36 // CDC request DIAL_DIGITS
	deviceCDCRequestSetUnitParameter            = 0x37 // CDC request SET_UNIT_PARAMETER
	deviceCDCRequestGetUnitParameter            = 0x38 // CDC request GET_UNIT_PARAMETER
	deviceCDCRequestClearUnitParameter          = 0x39 // CDC request CLEAR_UNIT_PARAMETER
	deviceCDCRequestSetEthernetMulticastFilters = 0x40 // CDC request SET_ETHERNET_MULTICAST_FILTERS
	deviceCDCRequestSetEthernetPowPatternFilter = 0x41 // CDC request SET_ETHERNET_POW_PATTER_FILTER
	deviceCDCRequestGetEthernetPowPatternFilter = 0x42 // CDC request GET_ETHERNET_POW_PATTER_FILTER
	deviceCDCRequestSetEthernetPacketFilter     = 0x43 // CDC request SET_ETHERNET_PACKET_FILTER
	deviceCDCRequestGetEthernetStatistic        = 0x44 // CDC request GET_ETHERNET_STATISTIC
	deviceCDCRequestSetATMDataFormat            = 0x50 // CDC request SET_ATM_DATA_FORMAT
	deviceCDCRequestGetATMDeviceStatistics      = 0x51 // CDC request GET_ATM_DEVICE_STATISTICS
	deviceCDCRequestSetATMDefaultVC             = 0x52 // CDC request SET_ATM_DEFAULT_VC
	deviceCDCRequestGetATMVCStatistics          = 0x53 // CDC request GET_ATM_VC_STATISTICS
	deviceCDCRequestMDLMSpecificRequestsMask    = 0x7F // CDC request MDLM_SPECIFIC_REQUESTS_MASK

	deviceCDCNotifyNetworkConnection     = 0x00 // CDC notify NETWORK_CONNECTION
	deviceCDCNotifyResponseAvail         = 0x01 // CDC notify RESPONSE_AVAIL
	deviceCDCNotifyAuxJackHookState      = 0x08 // CDC notify AUX_JACK_HOOK_STATE
	deviceCDCNotifyRingDetect            = 0x09 // CDC notify RING_DETECT
	deviceCDCNotifySerialState           = 0x20 // CDC notify SERIAL_STATE
	deviceCDCNotifyCallStateChange       = 0x28 // CDC notify CALL_STATE_CHANGE
	deviceCDCNotifyLineStateChange       = 0x29 // CDC notify LINE_STATE_CHANGE
	deviceCDCNotifyConnectionSpeedChange = 0x2A // CDC notify CONNECTION_SPEED_CHANGE

	deviceCDCFeatureAbstractState  = 0x01 // CDC feature select ABSTRACT_STATE
	deviceCDCFeatureCountrySetting = 0x02 // CDC feature select COUNTRY_SETTING

	deviceCDCControlSigBitmapCarrierActivation = 0x02 // CDC control signal CARRIER_ACTIVATION
	deviceCDCControlSigBitmapDTEPresence       = 0x01 // CDC control signal DTE_PRESENCE
	deviceCDCUARTStateRxCarrier                = 0x01 // UART state RX_CARRIER
	deviceCDCUARTStateTxCarrier                = 0x02 // UART state TX_CARRIER
	deviceCDCUARTStateBreak                    = 0x04 // UART state BREAK
	deviceCDCUARTStateRingSignal               = 0x08 // UART state RING_SIGNAL
	deviceCDCUARTStateFraming                  = 0x10 // UART state FRAMING
	deviceCDCUARTStateParity                   = 0x20 // UART state PARITY
	deviceCDCUARTStateOverrun                  = 0x40 // UART state OVERRUN

)
