//go:build nucleog0b1re

package main

// FDCAN Test Suite for STM32 Nucleo-G0B1RE board
// Wiring diagram: https://www.st.com/resource/en/datasheet/stm32g0b1re.pdf
// Waveshare CAN-HAT: https://www.waveshare.com/wiki/CAN-HAT
// Connect CAN-HAT TX to PA12, RX to PA11, GND to GND, 3.3V to VCC
// Provides a menu-driven interface for testing all CAN features

import (
	"machine"
	"runtime/volatile"
	"time"
)

var (
	led        machine.Pin
	canConfig  machine.FDCANConfig
	canRunning bool // Track if CAN bus is currently running

	// For interrupt-driven RX
	rxCount    volatile.Register32
	rxReady    volatile.Register32
	lastRxID   uint32
	lastRxDLC  uint8
	lastRxFD   bool
	lastRxData [64]byte // Extended to 64 bytes for FD support
)

func main() {
	led = machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// Default CAN configuration
	canConfig = machine.FDCANConfig{
		TransferRate:   machine.FDCANTransferRate500kbps,
		TransferRateFD: machine.FDCANTransferRate2000kbps,
		Mode:           machine.FDCANModeNormal,
		Tx:             machine.CAN1_TX_PIN,
		Rx:             machine.CAN1_RX_PIN,
		EnableFD:       true,
	}

	println("")
	print(colorCyan, colorBold)
	println("╔══════════════════════════════════════╗")
	println("║   FDCAN Test Suite for Nucleo-G0B1RE ║")
	println("╚══════════════════════════════════════╝")
	print(colorReset)
	println("")

	for {
		showMainMenu()
	}
}

func showMainMenu() {
	println("")
	// Show bus status in title
	print(colorCyan, colorBold, "═══ Main Menu ═══  ", colorReset)
	print("[CAN: ")
	if canRunning {
		print(colorGreen, "RUNNING", colorReset)
	} else {
		print(colorRed, "STOPPED", colorReset)
	}
	println("]")
	println("")
	printMenuItem("1", "Configure CAN")
	printMenuItem("2", "Bus Control")
	printMenuItem("3", "Basic TX/RX Test")
	printMenuItem("4", "Filter Test")
	printMenuItem("5", "Blocking TX Test")
	printMenuItem("6", "TX Cancellation Test")
	printMenuItem("7", "Interrupt RX Test")
	printMenuItem("8", "Error Diagnostics")
	printMenuItem("9", "Passive Monitor")
	printMenuItem("0", "Show Current Config")
	print(colorYellow, "f", colorReset)
	println(". CAN FD Test (large payloads)")
	println("")
	// Quick actions
	print(colorWhite, "Quick: ", colorReset)
	print(colorGreen, "s", colorReset)
	print("=start ")
	print(colorRed, "x", colorReset)
	print("=stop ")
	print(colorYellow, "r", colorReset)
	println("=restart")
	println("")
	print(colorWhite, "Enter choice: ", colorReset)

	choice := readChar()
	println("")

	switch choice {
	case '1':
		configureMenu()
	case '2':
		busControlMenu()
	case '3':
		basicTxRxTest()
	case '4':
		filterTest()
	case '5':
		blockingTxTest()
	case '6':
		txCancellationTest()
	case '7':
		interruptRxTest()
	case '8':
		errorDiagnosticsTest()
	case '9':
		passiveMonitor()
	case '0':
		showConfig()
	case 's', 'S':
		startCAN()
	case 'x', 'X':
		stopCAN()
	case 'r', 'R':
		restartCAN()
	case 'f', 'F':
		fdCanTest()
	default:
		printError("Invalid choice")
		println("")
	}
}

func configureMenu() {
	printTitle("═══ Configure CAN ═══")
	printMenuItem("1", "Normal Mode")
	printMenuItem("2", "Internal Loopback Mode")
	printMenuItem("3", "External Loopback Mode")
	printMenuItem("4", "Bus Monitoring Mode")
	print(colorYellow, "5", colorReset)
	print(". Toggle FD Mode (current: ", boolStr(canConfig.EnableFD), ")")
	println("")
	printMenuItem("6", "Set Baud Rate")
	println("")
	print(colorWhite, "Enter choice: ", colorReset)

	choice := readChar()
	println("")

	switch choice {
	case '1':
		canConfig.Mode = machine.FDCANModeNormal
		printSuccess("Mode set: ")
		println("Normal")
	case '2':
		canConfig.Mode = machine.FDCANModeInternalLoopback
		printSuccess("Mode set: ")
		println("Internal Loopback")
	case '3':
		canConfig.Mode = machine.FDCANModeExternalLoopback
		printSuccess("Mode set: ")
		println("External Loopback")
	case '4':
		canConfig.Mode = machine.FDCANModeBusMonitoring
		printSuccess("Mode set: ")
		println("Bus Monitoring")
	case '5':
		canConfig.EnableFD = !canConfig.EnableFD
		print("FD Mode: ", boolStr(canConfig.EnableFD))
		println("")
	case '6':
		setBaudRateMenu()
	}
}

func setBaudRateMenu() {
	printTitle("═══ Set Baud Rate ═══")
	printMenuItem("1", "100 kbps")
	printMenuItem("2", "125 kbps")
	print(colorYellow, "3", colorReset)
	println(". 250 kbps")
	print(colorYellow, "4", colorReset)
	print(". 500 kbps ")
	printSuccess("(default)")
	println("")
	printMenuItem("5", "1000 kbps")
	println("")
	print(colorWhite, "Enter choice: ", colorReset)

	choice := readChar()
	println("")

	switch choice {
	case '1':
		canConfig.TransferRate = machine.FDCANTransferRate100kbps
	case '2':
		canConfig.TransferRate = machine.FDCANTransferRate125kbps
	case '3':
		canConfig.TransferRate = machine.FDCANTransferRate250kbps
	case '4':
		canConfig.TransferRate = machine.FDCANTransferRate500kbps
	case '5':
		canConfig.TransferRate = machine.FDCANTransferRate1000kbps
	}
	printSuccess("Baud rate set to: ")
	printDec(uint32(canConfig.TransferRate))
	println(" bps")
}

func showConfig() {
	printTitle("═══ Current Configuration ═══")
	printInfo("Mode:         ")
	switch canConfig.Mode {
	case machine.FDCANModeNormal:
		println("Normal")
	case machine.FDCANModeInternalLoopback:
		printWarning("Internal Loopback")
		println("")
	case machine.FDCANModeExternalLoopback:
		printWarning("External Loopback")
		println("")
	case machine.FDCANModeBusMonitoring:
		printWarning("Bus Monitoring")
		println("")
	}
	printInfo("Nominal Rate: ")
	printHighlight("")
	printDec(uint32(canConfig.TransferRate))
	println(" bps")
	printInfo("FD Data Rate: ")
	printDec(uint32(canConfig.TransferRateFD))
	println(" bps")
	printInfo("FD Enabled:   ")
	print(boolStr(canConfig.EnableFD))
	println("")
}

func initCAN() error {
	// If already running, stop first to reconfigure
	if canRunning {
		machine.CAN1.Stop()
		canRunning = false
	}
	err := machine.CAN1.Configure(canConfig)
	if err != nil {
		return err
	}
	err = machine.CAN1.Start()
	if err == nil {
		canRunning = true
	}
	return err
}

// ensureCAN starts CAN if not already running, returns error if failed
func ensureCAN() error {
	if canRunning {
		return nil
	}
	return initCAN()
}

// startCAN starts the CAN bus with current configuration
func startCAN() {
	if canRunning {
		printWarning("CAN bus is already running")
		println("")
		return
	}

	printInfo("Starting CAN bus...")
	println("")
	err := initCAN()
	if err != nil {
		printError("Failed to start: ")
		println(err.Error())
		return
	}
	printSuccess("CAN bus started!")
	println("")
}

// stopCAN stops the CAN bus
func stopCAN() {
	if !canRunning {
		printWarning("CAN bus is already stopped")
		println("")
		return
	}

	machine.CAN1.Stop()
	canRunning = false
	printSuccess("CAN bus stopped")
	println("")
}

// restartCAN stops and restarts the CAN bus
func restartCAN() {
	printInfo("Restarting CAN bus...")
	println("")
	if canRunning {
		machine.CAN1.Stop()
		canRunning = false
	}
	err := initCAN()
	if err != nil {
		printError("Failed to restart: ")
		println(err.Error())
		return
	}
	printSuccess("CAN bus restarted!")
	println("")
}

// busControlMenu shows the bus control submenu
func busControlMenu() {
	printTitle("═══ Bus Control ═══")
	print("Status: ")
	if canRunning {
		print(colorGreen, "RUNNING", colorReset)
	} else {
		print(colorRed, "STOPPED", colorReset)
	}
	println("")
	println("")

	printMenuItem("1", "Start CAN Bus")
	printMenuItem("2", "Stop CAN Bus")
	printMenuItem("3", "Restart CAN Bus")
	printMenuItem("4", "Live Bus Status")
	printMenuItem("5", "Back to Main Menu")
	println("")
	print(colorWhite, "Enter choice: ", colorReset)

	choice := readChar()
	println("")

	switch choice {
	case '1':
		startCAN()
	case '2':
		stopCAN()
	case '3':
		restartCAN()
	case '4':
		liveBusStatus()
	case '5':
		// Return to main menu
	default:
		printError("Invalid choice")
		println("")
	}
}

// liveBusStatus shows continuous bus status updates
func liveBusStatus() {
	printTitle("═══ Live Bus Status ═══")
	printWarning("Press any key to return")
	println("")
	println("")

	for {
		if hasChar() {
			readChar()
			break
		}

		// Bus state
		print("Bus: ")
		if canRunning {
			print(colorGreen, "RUNNING", colorReset)

			// Show error counters and state
			txErr, rxErr := machine.CAN1.GetErrorCounters()
			state := machine.CAN1.GetBusState()

			printInfo("  TEC=")
			if txErr > 96 {
				print(colorRed)
			} else if txErr > 0 {
				print(colorYellow)
			}
			printDec(uint32(txErr))
			print(colorReset)

			printInfo(" REC=")
			if rxErr > 96 {
				print(colorRed)
			} else if rxErr > 0 {
				print(colorYellow)
			}
			printDec(uint32(rxErr))
			print(colorReset)

			printInfo(" State=")
			stateStr := state.String()
			if stateStr == "ErrorActive" {
				print(colorGreen, stateStr, colorReset)
			} else if stateStr == "ErrorPassive" {
				print(colorYellow, stateStr, colorReset)
			} else {
				print(colorRed, stateStr, colorReset)
			}

			// Show RX activity
			if !machine.CAN1.RxFifoIsEmpty() {
				rxID, _, _, isFD, _, _ := machine.CAN1.Rx8()
				print(colorGreen, "  RX:", colorReset)
				print(colorMagenta, "0x", colorReset)
				printHex16(uint16(rxID))
				if isFD {
					print(colorYellow, " FD", colorReset)
				}
			}
		} else {
			print(colorRed, "STOPPED", colorReset)
		}
		println("")

		led.Set(!led.Get())
		time.Sleep(500 * time.Millisecond)
	}
}

// passiveMonitor monitors the CAN bus without sending anything
func passiveMonitor() {
	printTitle("═══ Passive Monitor ═══")
	println("Watching CAN traffic (RX only)")
	printWarning("Press any key to stop")
	println("")

	// Start CAN if not already running
	wasRunning := canRunning
	if !canRunning {
		printInfo("Starting CAN...")
		println("")
		err := initCAN()
		if err != nil {
			printError("Failed to start: ")
			println(err.Error())
			return
		}
	}
	println("")

	msgCount := uint32(0)

	for {
		if hasChar() {
			readChar()
			break
		}

		// Check for received messages - use Rx64 when FD enabled for full payload
		for !machine.CAN1.RxFifoIsEmpty() {
			var rxID uint32
			var length uint8
			var isFD bool
			var rxData64 [64]byte
			var rxData8 [8]byte

			if canConfig.EnableFD {
				rxID, rxData64, length, isFD, _, _ = machine.CAN1.Rx64()
			} else {
				rxID, rxData8, length, isFD, _, _ = machine.CAN1.Rx8()
			}
			msgCount++

			print(colorWhite, "[", colorReset)
			printDec(msgCount)
			print(colorWhite, "] ", colorReset)
			// Show which Rx method is being used
			if canConfig.EnableFD {
				printSuccess("RX64 ")
			} else {
				printSuccess("RX8 ")
			}
			print(colorMagenta, "0x", colorReset)
			printHex16(uint16(rxID))
			if isFD {
				print(colorYellow, " FD", colorReset)
			}
			print(colorWhite, " [", colorReset)
			printDec(uint32(length))
			print(colorWhite, "] ", colorReset)

			// Print data bytes - handle both array types
			if canConfig.EnableFD {
				if length <= 8 {
					for i := uint8(0); i < length; i++ {
						printHex8(rxData64[i])
						print(" ")
					}
				} else {
					// Show first 4, then ..., then last 4
					for i := uint8(0); i < 4; i++ {
						printHex8(rxData64[i])
						print(" ")
					}
					print(".. ")
					for i := length - 4; i < length; i++ {
						printHex8(rxData64[i])
						print(" ")
					}
				}
			} else {
				for i := uint8(0); i < length; i++ {
					printHex8(rxData8[i])
					print(" ")
				}
			}
			println("")
			led.Set(!led.Get())
		}

		time.Sleep(1 * time.Millisecond)
	}

	// If we started CAN for this monitor, stop it
	if !wasRunning {
		machine.CAN1.Stop()
		canRunning = false
	}

	println("")
	printInfo("Messages received: ")
	print(colorGreen)
	printDec(msgCount)
	print(colorReset)
	println("")
	printWarning("Monitor stopped")
	println("")
}

func basicTxRxTest() {
	printTitle("═══ Basic TX/RX Test ═══")
	printInfo("Initializing CAN...")
	println("")

	err := initCAN()
	if err != nil {
		printError("Init failed: ")
		println(err.Error())
		return
	}

	printSuccess("CAN initialized!")
	println("")
	println("Sending test messages...")
	printWarning("Press any key to stop")
	println("")
	println("")

	data := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}
	txCount := uint32(0)

	for {
		// Check for keypress to exit
		if hasChar() {
			readChar()
			break
		}

		// Send message
		txCount++
		err := machine.CAN1.Tx(0x123, data, canConfig.EnableFD, false)
		if err != nil {
			print(colorWhite, "[", colorReset)
			printDec(txCount)
			print(colorWhite, "] ", colorReset)
			printError("TX Error: ")
			println(err.Error())
		} else {
			print(colorWhite, "[", colorReset)
			printDec(txCount)
			print(colorWhite, "] ", colorReset)
			printInfo("TX ")
			print(colorMagenta, "0x123", colorReset)
		}

		// Check for received messages
		time.Sleep(10 * time.Millisecond)
		for !machine.CAN1.RxFifoIsEmpty() {
			rxID, rxData, length, isFD, _, _ := machine.CAN1.Rx8()
			printSuccess(" -> RX ")
			print(colorMagenta, "0x", colorReset)
			printHex16(uint16(rxID))
			if isFD {
				print(colorYellow, " FD", colorReset)
			}
			print(colorWhite, " [", colorReset)
			printDec(uint32(length))
			print(colorWhite, "]", colorReset)
			if length > 0 {
				print(" ")
				printHex8(rxData[0])
			}
		}
		println("")

		led.Set(!led.Get())
		time.Sleep(500 * time.Millisecond)
	}

	machine.CAN1.Stop()
	printWarning("Test stopped")
	println("")
}

func filterTest() {
	printTitle("═══ Filter Test ═══")

	// Use loopback for filter testing
	savedMode := canConfig.Mode
	canConfig.Mode = machine.FDCANModeInternalLoopback

	err := initCAN()
	if err != nil {
		printError("Init failed: ")
		println(err.Error())
		canConfig.Mode = savedMode
		return
	}

	// Configure filters
	printInfo("Setting up filters:")
	println("")
	print("  Filter 0: Accept ID ")
	print(colorMagenta, "0x123", colorReset)
	println("")
	machine.CAN1.AcceptID(0, 0x123, false)

	print("  Filter 1: Accept range ")
	print(colorMagenta, "0x200-0x2FF", colorReset)
	println("")
	machine.CAN1.AcceptRange(1, 0x200, 0x2FF, false)

	print("  Filter 2: Accept mask ")
	print(colorMagenta, "0x300/0x7F0", colorReset)
	println(" (0x300-0x30F)")
	machine.CAN1.AcceptMask(2, 0x300, 0x7F0, false)

	print("  Non-matching: ")
	printError("REJECT")
	println("")
	machine.CAN1.RejectNonMatching()

	println("")
	printInfo("Testing filter combinations:")
	println("")
	println("")

	testIDs := []uint32{0x100, 0x123, 0x124, 0x200, 0x250, 0x2FF, 0x300, 0x305, 0x310, 0x400}
	data := []byte{0xAA, 0xBB, 0xCC, 0xDD}

	for _, id := range testIDs {
		err := machine.CAN1.Tx(id, data, false, false)
		if err != nil {
			print("TX ")
			print(colorMagenta, "0x", colorReset)
			printHex16(uint16(id))
			printError(" - TX error")
			println("")
			continue
		}

		time.Sleep(2 * time.Millisecond)

		if !machine.CAN1.RxFifoIsEmpty() {
			rxID, _, _, _, _, _ := machine.CAN1.Rx8()
			print("TX ")
			print(colorMagenta, "0x", colorReset)
			printHex16(uint16(id))
			printSuccess(" -> RX ")
			print(colorMagenta, "0x", colorReset)
			printHex16(uint16(rxID))
			print(" ")
			printSuccess("PASSED")
			println("")
		} else {
			print("TX ")
			print(colorMagenta, "0x", colorReset)
			printHex16(uint16(id))
			print(" -> ")
			printWarning("FILTERED")
			println("")
		}
	}

	machine.CAN1.Stop()
	canConfig.Mode = savedMode
	println("")
	printSuccess("Filter test complete")
	println("")
}

func blockingTxTest() {
	printTitle("═══ Blocking TX Test ═══")
	printInfo("Timeout: ")
	println("100ms")
	printWarning("Press any key to stop")
	println("")
	println("")

	err := initCAN()
	if err != nil {
		printError("Init failed: ")
		println(err.Error())
		return
	}

	data := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}
	txCount := uint32(0)
	successCount := uint32(0)
	failCount := uint32(0)

	for {
		if hasChar() {
			readChar()
			break
		}

		// Check for RX
		for !machine.CAN1.RxFifoIsEmpty() {
			rxID, _, _, _, _, _ := machine.CAN1.Rx8()
			printSuccess("RX ")
			print(colorMagenta, "0x", colorReset)
			printHex16(uint16(rxID))
			println("")
		}

		txCount++
		err := machine.CAN1.TxBlocking(0x123, data, canConfig.EnableFD, false, 100)

		if err == nil {
			successCount++
			print(colorWhite, "[", colorReset)
			printDec(txCount)
			print(colorWhite, "] ", colorReset)
			printSuccess("TX OK")
			print(" (")
			print(colorGreen)
			printDec(successCount)
			print(colorReset, "/", colorRed)
			printDec(failCount)
			print(colorReset)
			println(")")
			led.High()
		} else {
			failCount++
			print(colorWhite, "[", colorReset)
			printDec(txCount)
			print(colorWhite, "] ", colorReset)
			printError("TX FAIL: ")
			println(err.Error())
			led.Low()
		}

		time.Sleep(500 * time.Millisecond)
	}

	machine.CAN1.Stop()
	printWarning("Test stopped")
	println("")
}

func txCancellationTest() {
	printTitle("═══ TX Cancellation Test ═══")
	printWarning("Press any key to stop")
	println("")
	println("")

	err := initCAN()
	if err != nil {
		printError("Init failed: ")
		println(err.Error())
		return
	}

	data := []byte{0x11, 0x22, 0x33, 0x44}
	testNum := 0

	for {
		if hasChar() {
			readChar()
			break
		}

		testNum++
		print(colorCyan, "--- Test ", colorReset)
		printDec(uint32(testNum))
		print(colorCyan, " ---", colorReset)
		println("")

		// Queue multiple messages
		printInfo("Queuing 3 messages...")
		println("")
		machine.CAN1.Tx(0x100, data, false, false)
		machine.CAN1.Tx(0x101, data, false, false)
		machine.CAN1.Tx(0x102, data, false, false)

		pending := machine.CAN1.TxPendingCount()
		print("Pending: ")
		print(colorYellow)
		printDec(uint32(pending))
		print(colorReset)
		println("")

		time.Sleep(1 * time.Millisecond)

		if machine.CAN1.TxIsPending() {
			printWarning("Cancelling all...")
			println("")
			cancelled := machine.CAN1.TxCancelAll()
			print("Cancelled: ")
			print(colorRed)
			printDec(uint32(cancelled))
			print(colorReset)
			println("")
		} else {
			printSuccess("All sent successfully")
			println("")
		}

		// Check RX
		for !machine.CAN1.RxFifoIsEmpty() {
			machine.CAN1.Rx8()
		}

		println("")
		time.Sleep(2 * time.Second)
	}

	machine.CAN1.Stop()
	printWarning("Test stopped")
	println("")
}

func interruptRxTest() {
	printTitle("═══ Interrupt RX Test ═══")
	println("Messages trigger interrupt callback")
	printWarning("Press any key to stop")
	println("")
	println("")

	err := initCAN()
	if err != nil {
		printError("Init failed: ")
		println(err.Error())
		return
	}

	rxCount.Set(0)
	rxReady.Set(0)

	// Set up minimal callback - copy full 64 bytes for FD support
	machine.CAN1.SetRxCallback(func(msg *machine.FDCANRxBufferElement) {
		rxCount.Set(rxCount.Get() + 1)
		lastRxID = msg.ID
		lastRxDLC = msg.DLC
		lastRxFD = msg.FDF
		// Copy full 64 bytes to support FD frames
		lastRxData = msg.DB
		rxReady.Set(1)
		led.Set(!led.Get())
	})

	printInfo("Waiting for messages...")
	println("")
	println("")

	for {
		if hasChar() {
			readChar()
			break
		}

		if rxReady.Get() != 0 {
			rxReady.Set(0)
			count := rxCount.Get()
			print(colorWhite, "[", colorReset)
			printDec(count)
			print(colorWhite, "] ", colorReset)
			printSuccess("RX ")
			print(colorMagenta, "0x", colorReset)
			printHex16(uint16(lastRxID))
			if lastRxFD {
				print(colorYellow, " FD", colorReset)
			}
			printInfo(" DLC=")
			printDec(uint32(lastRxDLC))
			printInfo(" DATA=")
			length := machine.FDCANDlcToLength(lastRxDLC, lastRxFD)
			for i := byte(0); i < length && i < 8; i++ {
				printHex8(lastRxData[i])
				print(" ")
			}
			println("")
		}

		time.Sleep(1 * time.Millisecond)
	}

	machine.CAN1.SetRxCallback(nil)
	machine.CAN1.Stop()
	printWarning("Test stopped")
	println("")
}

func errorDiagnosticsTest() {
	printTitle("═══ Error Diagnostics ═══")
	printWarning("Press any key to stop")
	println("")
	println("")

	err := initCAN()
	if err != nil {
		printError("Init failed: ")
		println(err.Error())
		return
	}

	for {
		if hasChar() {
			readChar()
			break
		}

		txErr, rxErr := machine.CAN1.GetErrorCounters()
		state := machine.CAN1.GetBusState()
		lastErr := machine.CAN1.GetLastError()

		printInfo("TEC=")
		if txErr > 0 {
			print(colorYellow)
		}
		printDec(uint32(txErr))
		print(colorReset)
		printInfo(" REC=")
		if rxErr > 0 {
			print(colorYellow)
		}
		printDec(uint32(rxErr))
		print(colorReset)
		printInfo(" State=")
		print(state.String())
		printInfo(" LastErr=")
		println(lastErr.String())

		if machine.CAN1.IsBusOff() {
			print(colorRed, colorBold)
			println("  !! BUS OFF !!")
			print(colorReset)
		}
		if machine.CAN1.IsErrorPassive() {
			print(colorRed)
			println("  !! ERROR PASSIVE !!")
			print(colorReset)
		}
		if machine.CAN1.IsErrorWarning() {
			print(colorYellow)
			println("  !! ERROR WARNING !!")
			print(colorReset)
		}

		led.Set(!led.Get())
		time.Sleep(1 * time.Second)
	}

	machine.CAN1.Stop()
	printWarning("Test stopped")
	println("")
}

func loopbackTest() {
	printTitle("═══ Loopback Test ═══")
	printInfo("Using internal loopback mode")
	println("")
	printWarning("Press any key to stop")
	println("")
	println("")

	savedMode := canConfig.Mode
	canConfig.Mode = machine.FDCANModeInternalLoopback

	err := initCAN()
	if err != nil {
		printError("Init failed: ")
		println(err.Error())
		canConfig.Mode = savedMode
		return
	}

	data := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0xCA, 0xFE, 0xBA, 0xBE}
	txCount := uint32(0)
	localRxCount := uint32(0)

	for {
		if hasChar() {
			readChar()
			break
		}

		txCount++
		data[0] = byte(txCount)

		err := machine.CAN1.Tx(0x7FF, data, canConfig.EnableFD, false)
		if err != nil {
			printError("TX Error: ")
			println(err.Error())
			continue
		}

		time.Sleep(5 * time.Millisecond)

		if !machine.CAN1.RxFifoIsEmpty() {
			rxID, rxData, length, isFD, _, _ := machine.CAN1.Rx8()
			localRxCount++

			print(colorWhite, "[", colorReset)
			printDec(txCount)
			print(colorWhite, "] ", colorReset)
			printInfo("TX->RX ")
			print(colorMagenta, "0x", colorReset)
			printHex16(uint16(rxID))
			if isFD {
				print(colorYellow, " FD", colorReset)
			}
			print(colorWhite, " [", colorReset)
			printDec(uint32(length))
			print(colorWhite, "] ", colorReset)
			for i := uint8(0); i < length && i < 4; i++ {
				printHex8(rxData[i])
				print(" ")
			}

			if rxData[0] == byte(txCount) {
				printSuccess("OK")
				println("")
			} else {
				printError("MISMATCH!")
				println("")
			}
		} else {
			print(colorWhite, "[", colorReset)
			printDec(txCount)
			print(colorWhite, "] ", colorReset)
			printError("TX - NO RX!")
			println("")
		}

		led.Set(!led.Get())
		time.Sleep(500 * time.Millisecond)
	}

	machine.CAN1.Stop()
	canConfig.Mode = savedMode

	println("")
	printInfo("TX: ")
	print(colorGreen)
	printDec(txCount)
	print(colorReset)
	printInfo(" RX: ")
	print(colorGreen)
	printDec(localRxCount)
	print(colorReset)
	println("")
	printWarning("Test stopped")
	println("")
}

func continuousTxTest() {
	printTitle("═══ Continuous TX Test ═══")
	println("Sending as fast as possible")
	printWarning("Press any key to stop")
	println("")
	println("")

	err := initCAN()
	if err != nil {
		printError("Init failed: ")
		println(err.Error())
		return
	}

	data := []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77}
	txCount := uint32(0)
	errCount := uint32(0)
	startTime := time.Now()

	for {
		if hasChar() {
			readChar()
			break
		}

		// Drain RX to avoid overflow
		for !machine.CAN1.RxFifoIsEmpty() {
			machine.CAN1.Rx8()
		}

		err := machine.CAN1.Tx(0x555, data, canConfig.EnableFD, false)
		if err == nil {
			txCount++
			data[0]++
		} else {
			errCount++
			time.Sleep(1 * time.Millisecond)
		}

		// Print stats every 1000 messages
		if txCount%1000 == 0 && txCount > 0 {
			elapsed := time.Since(startTime).Milliseconds()
			rate := uint32(0)
			if elapsed > 0 {
				rate = (txCount * 1000) / uint32(elapsed)
			}
			printInfo("TX: ")
			print(colorGreen)
			printDec(txCount)
			print(colorReset)
			printInfo(" Err: ")
			if errCount > 0 {
				print(colorRed)
			}
			printDec(errCount)
			print(colorReset)
			printInfo(" Rate: ")
			print(colorMagenta)
			printDec(rate)
			print(colorReset)
			println(" msg/s")
			led.Set(!led.Get())
		}
	}

	machine.CAN1.Stop()
	println("")
	printInfo("Total TX: ")
	print(colorGreen)
	printDec(txCount)
	print(colorReset)
	printInfo(" Errors: ")
	if errCount > 0 {
		print(colorRed)
	} else {
		print(colorGreen)
	}
	printDec(errCount)
	print(colorReset)
	println("")
	printWarning("Test stopped")
	println("")
}

// fdCanTest tests CAN FD with large payloads (up to 64 bytes)
// Uses Rx64() for zero-allocation receive of full FD frames
func fdCanTest() {
	printTitle("═══ CAN FD Test (Large Payloads) ═══")

	// Check if FD mode is enabled
	if !canConfig.EnableFD {
		printWarning("FD mode is disabled. Enabling it for this test...")
		println("")
		canConfig.EnableFD = true
	}

	printInfo("Nominal Rate: ")
	printDec(uint32(canConfig.TransferRate))
	println(" bps")
	printInfo("FD Data Rate: ")
	printDec(uint32(canConfig.TransferRateFD))
	println(" bps")
	println("")

	// Use loopback for testing
	savedMode := canConfig.Mode
	canConfig.Mode = machine.FDCANModeInternalLoopback

	err := initCAN()
	if err != nil {
		printError("Init failed: ")
		println(err.Error())
		canConfig.Mode = savedMode
		return
	}

	printSuccess("CAN FD initialized in loopback mode")
	println("")
	printWarning("Press any key to stop")
	println("")
	println("")

	// CAN FD valid DLC values: 0-8, 12, 16, 20, 24, 32, 48, 64
	testSizes := []uint8{8, 12, 16, 20, 24, 32, 48, 64}
	testNum := uint32(0)

	for {
		if hasChar() {
			readChar()
			break
		}

		for _, size := range testSizes {
			if hasChar() {
				break
			}

			testNum++

			// Create test data pattern
			var txData [64]byte
			for i := uint8(0); i < size; i++ {
				txData[i] = byte(testNum + uint32(i))
			}

			// Send FD frame with specified size
			err := machine.CAN1.Tx(0x1FD, txData[:size], true, false)
			if err != nil {
				print(colorWhite, "[", colorReset)
				printDec(testNum)
				print(colorWhite, "] ", colorReset)
				printError("TX Error (")
				printDec(uint32(size))
				print("B): ")
				println(err.Error())
				continue
			}

			time.Sleep(5 * time.Millisecond)

			// Receive using Rx64 for full FD support
			if !machine.CAN1.RxFifoIsEmpty() {
				rxID, rxData, length, isFD, _, rxErr := machine.CAN1.Rx64()
				if rxErr != nil {
					printError("RX Error: ")
					println(rxErr.Error())
					continue
				}

				print(colorWhite, "[", colorReset)
				printDec(testNum)
				print(colorWhite, "] ", colorReset)

				// Verify
				if !isFD {
					printError("NOT FD! ")
				}
				if length != size {
					printError("LEN MISMATCH! ")
				}

				// Check data integrity
				match := true
				for i := uint8(0); i < length; i++ {
					if rxData[i] != txData[i] {
						match = false
						break
					}
				}

				print(colorMagenta, "0x", colorReset)
				printHex16(uint16(rxID))
				print(colorYellow, " FD", colorReset)
				print(colorWhite, " [", colorReset)
				printDec(uint32(length))
				print(colorWhite, "B] ", colorReset)

				// Show first and last few bytes for large payloads
				if length <= 8 {
					for i := uint8(0); i < length; i++ {
						printHex8(rxData[i])
						print(" ")
					}
				} else {
					// Show first 4 bytes
					for i := uint8(0); i < 4; i++ {
						printHex8(rxData[i])
						print(" ")
					}
					print(".. ")
					// Show last 4 bytes
					for i := length - 4; i < length; i++ {
						printHex8(rxData[i])
						print(" ")
					}
				}

				if match {
					printSuccess("OK")
				} else {
					printError("DATA MISMATCH!")
				}
				println("")
			} else {
				print(colorWhite, "[", colorReset)
				printDec(testNum)
				print(colorWhite, "] ", colorReset)
				printError("TX ")
				printDec(uint32(size))
				print("B - NO RX!")
				println("")
			}

			led.Set(!led.Get())
			time.Sleep(300 * time.Millisecond)
		}

		println("")
		printInfo("Cycle complete. Repeating...")
		println("")
		println("")
		time.Sleep(1 * time.Second)
	}

	machine.CAN1.Stop()
	canConfig.Mode = savedMode

	println("")
	printInfo("Tests run: ")
	print(colorGreen)
	printDec(testNum)
	print(colorReset)
	println("")
	printWarning("FD test stopped")
	println("")
}

// ANSI Color codes
const (
	colorReset   = "\x1b[0m"
	colorBold    = "\x1b[1m"
	colorRed     = "\x1b[31m"
	colorGreen   = "\x1b[32m"
	colorYellow  = "\x1b[33m"
	colorBlue    = "\x1b[34m"
	colorMagenta = "\x1b[35m"
	colorCyan    = "\x1b[36m"
	colorWhite   = "\x1b[37m"
)

// Color helper functions
func printColor(color, text string) {
	print(color, text, colorReset)
}

func printlnColor(color, text string) {
	println(color, text, colorReset)
}

func printSuccess(text string) {
	print(colorGreen, text, colorReset)
}

func printError(text string) {
	print(colorRed, text, colorReset)
}

func printWarning(text string) {
	print(colorYellow, text, colorReset)
}

func printInfo(text string) {
	print(colorCyan, text, colorReset)
}

func printHighlight(text string) {
	print(colorMagenta, colorBold, text, colorReset)
}

func printTitle(text string) {
	println(colorCyan, colorBold, text, colorReset)
}

func printMenuItem(num string, text string) {
	print(colorYellow, num, colorReset)
	print(". ")
	println(text)
}

// Helper functions

func boolStr(b bool) string {
	if b {
		return colorGreen + "ON" + colorReset
	}
	return colorRed + "OFF" + colorReset
}

func readChar() byte {
	for {
		if machine.Serial.Buffered() > 0 {
			c, _ := machine.Serial.ReadByte()
			return c
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func hasChar() bool {
	return machine.Serial.Buffered() > 0
}

func printDec(v uint32) {
	if v == 0 {
		print("0")
		return
	}
	var buf [10]byte
	i := 0
	for v > 0 {
		buf[i] = byte('0' + v%10)
		v /= 10
		i++
	}
	for i > 0 {
		i--
		print(string(buf[i]))
	}
}

func printHex8(v uint8) {
	digits := "0123456789ABCDEF"
	print(string(digits[(v>>4)&0xF]))
	print(string(digits[v&0xF]))
}

func printHex16(v uint16) {
	digits := "0123456789ABCDEF"
	print(string(digits[(v>>12)&0xF]))
	print(string(digits[(v>>8)&0xF]))
	print(string(digits[(v>>4)&0xF]))
	print(string(digits[v&0xF]))
}

func blinkError(led machine.Pin) {
	for {
		led.High()
		time.Sleep(100 * time.Millisecond)
		led.Low()
		time.Sleep(100 * time.Millisecond)
	}
}
