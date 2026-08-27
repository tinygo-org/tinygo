# Smoke tests: check that TinyGo can build a binary for every supported board.

.PHONY: testchdir
testchdir:
	# test 'build' command with{,out} -C argument
	$(TINYGO) build -C tests/testing/chdir chdir.go && rm tests/testing/chdir/chdir
	$(TINYGO) build ./tests/testing/chdir/chdir.go && rm chdir
	# test 'run' command with{,out} -C argument
	EXPECT_DIR=$(PWD)/tests/testing/chdir $(TINYGO) run -C tests/testing/chdir chdir.go
	EXPECT_DIR=$(PWD) $(TINYGO) run ./tests/testing/chdir/chdir.go

.PHONY: smoketest
smoketest: testchdir
	$(TINYGO) version
	$(TINYGO) targets > /dev/null
	# regression test for #2892
	cd tests/testing/recurse && ($(TINYGO) test ./... > recurse.log && cat recurse.log && test $$(wc -l < recurse.log) = 2 && rm recurse.log)
	# compile-only platform-independent examples
	cd tests/text/template/smoke && $(TINYGO) test -c && rm -f smoke.test
	# regression test for #2563
	cd tests/os/smoke && $(TINYGO) test -c -target=pybadge && rm smoke.test
	# test all examples (except pwm)
	$(TINYGO) build -size short -o test.hex -target=pga2350             examples/echo
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/adc
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/blinkm
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/blinky2
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/button
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/button2
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/echo
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/echo2
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=circuitplay-express examples/i2s
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/mcp3008
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/memstats
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=microbit            examples/microbit-blink
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/pininterrupt
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nano-rp2040         examples/rtcinterrupt
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/machinetest
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/systick
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/test
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/time-offset
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=wioterminal         examples/hid-mouse
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=wioterminal         examples/hid-keyboard
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-rp2040      examples/i2c-target
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-rp2040      examples/watchdog
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-rp2040      examples/device-id
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pico2-ice           examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build             -o test.efi -target=uefi-amd64          examples/test
	@$(MD5SUM) test.efi
	# test simulated boards on play.tinygo.org
ifneq ($(WASM), 0)
	GOOS=js GOARCH=wasm $(TINYGO) build -size short -o test.wasm -tags=arduino_uno          examples/blinky1
	@$(MD5SUM) test.wasm
	GOOS=js GOARCH=wasm $(TINYGO) build -size short -o test.wasm -tags=hifive1b             examples/blinky1
	@$(MD5SUM) test.wasm
	GOOS=js GOARCH=wasm $(TINYGO) build -size short -o test.wasm -tags=reelboard            examples/blinky1
	@$(MD5SUM) test.wasm
	GOOS=js GOARCH=wasm $(TINYGO) build -size short -o test.wasm -tags=microbit             examples/microbit-blink
	@$(MD5SUM) test.wasm
	GOOS=js GOARCH=wasm $(TINYGO) build -size short -o test.wasm -tags=circuitplay_express  examples/blinky1
	@$(MD5SUM) test.wasm
	GOOS=js GOARCH=wasm $(TINYGO) build -size short -o test.wasm -tags=circuitplay_bluefruit examples/blinky1
	@$(MD5SUM) test.wasm
	GOOS=js GOARCH=wasm $(TINYGO) build -size short -o test.wasm -tags=mch2022              examples/machinetest
	@$(MD5SUM) test.wasm
	GOOS=js GOARCH=wasm $(TINYGO) build -size short -o test.wasm -tags=gopher_badge         examples/blinky1
	@$(MD5SUM) test.wasm
	GOOS=js GOARCH=wasm $(TINYGO) build -size short -o test.wasm -tags=pico                 examples/blinky1
	@$(MD5SUM) test.wasm
	GOOS=js GOARCH=wasm $(TINYGO) build -size short -o test.wasm -tags=xiao_esp32s3         examples/blinky1
	@$(MD5SUM) test.wasm
endif
	# test all targets/boards
	$(TINYGO) build -size short -o test.hex -target=pca10040-s132v6     examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=microbit            examples/echo
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=microbit-s110v8     examples/echo
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=microbit-v2         examples/microbit-blink
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=microbit-v2-s113v7  examples/microbit-blink
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=microbit-v2-s140v7  examples/microbit-blink
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nrf52840-mdk        examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=btt-skr-pico        examples/uart
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10031            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=reelboard           examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=reelboard           examples/blinky2
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10056            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10056            examples/blinky2
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10059            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10059            examples/blinky2
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=bluemicro840        examples/blinky2
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=itsybitsy-m0        examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-m0          examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=trinket-m0          examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=gemma-m0            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=circuitplay-express examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=circuitplay-bluefruit examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=circuitplay-express examples/i2s
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=clue-alpha          examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.gba -target=gameboy-advance     examples/gba-display
	@$(MD5SUM) test.gba
	$(TINYGO) build -size short -o test.hex -target=grandcentral-m4     examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=itsybitsy-m4        examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-m4          examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=matrixportal-m4     examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pybadge             examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=metro-m4-airlift    examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pyportal            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=particle-argon      examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=particle-boron      examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=particle-xenon      examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pinetime            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=x9pro               examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10056-s140v7     examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10059-s140v7     examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=reelboard-s140v7    examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=wioterminal         examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pygamer             examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=xiao                examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=xiao-ble-plus       examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=rak4631             examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=circuitplay-express examples/dac
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pyportal            examples/dac
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-nrf52840  	examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-nrf52840-sense examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=itsybitsy-nrf52840  examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=qtpy                examples/machinetest
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=teensy41            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=teensy40            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=teensy36            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=p1am-100            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=atsame54-xpro       examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=atsame54-xpro       examples/can
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-m4-can      examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-m4-can      examples/caninterrupt
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=arduino-nano33      examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=arduino-mkrwifi1010 examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pico                examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pico -gc=leaking    examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pico-w              examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nano-33-ble         examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nano-rp2040         examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-rp2040 		examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=qtpy-rp2040         examples/echo
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=kb2040              examples/echo
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=macropad-rp2040 	examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=badger2040          examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=badger2040-w        examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=tufty2040           examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=thingplus-rp2040    examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=xiao-rp2040         examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=waveshare-rp2040-zero examples/echo
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=challenger-rp2040    examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=trinkey-qt2040      examples/temp
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=gopher-badge      examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=gopher-arcade      examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=ae-rp2040           examples/echo
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=thumby              examples/echo
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pico2               examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pico2-w             examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=tiny2350            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=badger2350          examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=blinky2350          examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pico-plus2          examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=metro-rp2350        examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=waveshare-rp2040-tiny examples/echo
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=vicharak_shrike-lite examples/echo
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=xiao-rp2350        examples/blinky1
	@$(MD5SUM) test.hex
	# test pwm
	$(TINYGO) build -size short -o test.hex -target=itsybitsy-m0        examples/pwm
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=itsybitsy-m4        examples/pwm
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-m4          examples/pwm
	@$(MD5SUM) test.hex
	# test usb
	$(TINYGO) build -size short -o test.hex -target=feather-nrf52840    examples/hid-keyboard
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=circuitplay-express examples/hid-keyboard
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-nrf52840    examples/usb-midi
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pico    			examples/usb-storage
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pico2    			examples/usb-storage
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nrf52840-s140v6-uf2-generic	examples/machinetest
	@$(MD5SUM) test.hex
ifneq ($(STM32), 0)
	$(TINYGO) build -size short -o test.hex -target=bluepill            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-stm32f405   examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=lgt92               examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nucleo-f103rb       examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nucleo-f722ze       examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nucleo-h753zi       examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nucleo-l031k6       examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nucleo-l432kc       examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nucleo-l476rg       examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nucleo-l552ze       examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nucleo-wl55jc       examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=stm32f4disco        examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=stm32f4disco        examples/blinky2
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=stm32f4disco-1      examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=stm32f4disco-1      examples/pwm
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=stm32f469disco      examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=lorae5              examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=swan                examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=mksnanov3           examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=stm32l0x1           examples/serial
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=stm32u031           examples/empty
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=arduino-uno-q       examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=arduino-uno-q       examples/serial
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=arduino-uno-q       examples/blinkm
	@$(MD5SUM) test.hex
endif
	$(TINYGO) build -size short -o test.hex -target=atmega328pb         examples/blinkm
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=atmega1284p         examples/machinetest
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=arduino-uno         examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=arduino-leonardo    examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=arduino-uno         examples/pwm
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=arduino-uno -scheduler=tasks  examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=arduino-mega1280    examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=arduino-mega1280    examples/pwm
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=arduino-nano        examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=attiny1616          examples/empty
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=digispark           examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=digispark           examples/pwm
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=digispark           examples/mcp3008
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=digispark -gc=leaking examples/blinky1
	@$(MD5SUM) test.hex
ifneq ($(XTENSA), 0)
	$(TINYGO) build -size short -o test.bin -target=esp32-generic       examples/machinetest
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=esp32-coreboard-v2  examples/adc
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=esp32c3-generic     examples/machinetest
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=esp32s3-generic     examples/machinetest
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=esp32-mini32      	examples/blinky1
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=esp32c3-supermini   examples/blinky1
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=esp32c3-supermini   examples/blinkm
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=nodemcu             examples/blinky1
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target m5stack-core2       examples/machinetest
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target m5stack             examples/machinetest
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target m5stamp-s3a         examples/machinetest
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target m5stick-c           examples/machinetest
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target m5paper             examples/machinetest
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target mch2022             examples/machinetest
	@$(MD5SUM) test.bin
	# xiao-esp32c6
	$(TINYGO) build -size short -o test.bin -target=xiao-esp32c6      	examples/blinky1
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=xiao-esp32c6   		examples/blinkm
	@$(MD5SUM) test.bin
	# xiao-esp32s3
	$(TINYGO) build -size short -o test.bin -target=xiao-esp32s3   		examples/blinky1
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=xiao-esp32s3   		examples/blinkm
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=xiao-esp32s3   		examples/mcp3008
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=xiao-esp32s3   		examples/pwm
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=xiao-esp32s3   		examples/adc
	@$(MD5SUM) test.bin
	# esp32s3-supermini
	$(TINYGO) build -size short -o test.bin -target=esp32s3-supermini	    examples/blinky1
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=esp32s3-supermini	    examples/blinkm
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=esp32s3-supermini	    examples/mcp3008
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=esp32s3-supermini   	examples/adc
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=esp32s3-box-3       examples/blinky1
	@$(MD5SUM) test.bin
endif
    # esp32c3-supermini
	$(TINYGO) build -size short -o test.bin -target=esp32c3-supermini	    examples/blinky1
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=esp32c3-supermini	    examples/blinkm
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=esp32c3-supermini	    examples/mcp3008
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=esp32c3-supermini   	examples/pwm
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=esp32c3-supermini   	examples/adc
	@$(MD5SUM) test.bin

	$(TINYGO) build -size short -o test.bin -target=esp-c3-32s-kit      examples/blinky1
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=qtpy-esp32c3        examples/machinetest
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=m5stamp-c3          examples/machinetest
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=xiao-esp32c3        examples/machinetest
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=esp32-c3-devkit-rust-1 examples/blinky1
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=esp32c3-12f         examples/blinky1
	@$(MD5SUM) test.bin

	$(TINYGO) build -size short -o test.bin -target=makerfabs-esp32c3spi35 examples/machinetest
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.hex -target=hifive1b            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=maixbit             examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=tkey                examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=elecrow-rp2040      examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=elecrow-rp2350      examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=hw-651              examples/machinetest
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=hw-651-s110v8       examples/machinetest
	@$(MD5SUM) test.hex
ifneq ($(WASM), 0)
	$(TINYGO) build -size short -o wasm.wasm -target=wasm               examples/wasm/export
	$(TINYGO) build -size short -o wasm.wasm -target=wasm               examples/wasm/main
	$(TINYGO) build -size short -o wasm.wasm -target=wasm-unknown       examples/hello-wasm-unknown
endif
	# test various compiler flags
	$(TINYGO) build -size short -o test.hex -target=pca10040 -gc=none -scheduler=none examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040 -opt=1     examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040 -serial=none examples/echo
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040 -serial=rtt examples/echo
	@$(MD5SUM) test.hex
	$(TINYGO) build             -o test.nro -target=nintendoswitch      examples/echo2
	@$(MD5SUM) test.nro
	$(TINYGO) build -size short -o test.hex -target=pca10040 -opt=0     ./testdata/stdlib.go
	@$(MD5SUM) test.hex
	GOOS=linux GOARCH=arm $(TINYGO) build -size short -o test.elf       ./testdata/cgo
	GOOS=linux GOARCH=mips    $(TINYGO) build -size short -o test.elf   ./testdata/cgo
	GOOS=windows GOARCH=amd64 $(TINYGO) build -size short -o test.exe   ./testdata/cgo
	GOOS=windows GOARCH=arm64 $(TINYGO) build -size short -o test.exe   ./testdata/cgo
	GOOS=darwin GOARCH=amd64 $(TINYGO) build  -size short -o test       ./testdata/cgo
	GOOS=darwin GOARCH=arm64 $(TINYGO) build  -size short -o test       ./testdata/cgo
ifneq ($(OS),Windows_NT)
	# TODO: this does not yet work on Windows. Somehow, unused functions are
	# not garbage collected.
	$(TINYGO) build -o test.elf -gc=leaking -scheduler=none examples/serial
endif
