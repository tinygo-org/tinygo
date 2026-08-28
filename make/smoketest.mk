# Smoke tests: check that TinyGo can build a binary for every supported board.

.PHONY: testchdir
testchdir:
	# test 'build' command with{,out} -C argument
	$(TINYGO) build -C tests/testing/chdir chdir.go && rm tests/testing/chdir/chdir
	$(TINYGO) build ./tests/testing/chdir/chdir.go && rm chdir
	# test 'run' command with{,out} -C argument
	EXPECT_DIR=$(PWD)/tests/testing/chdir $(TINYGO) run -C tests/testing/chdir chdir.go
	EXPECT_DIR=$(PWD) $(TINYGO) run ./tests/testing/chdir/chdir.go

SMOKETEST_SUBTARGETS = \
	smoketest-selftest \
	smoketest-examples \
	smoketest-wasm-sim \
	smoketest-nrf \
	smoketest-samd \
	smoketest-nxp \
	smoketest-rp2xxx \
	smoketest-pwm-usb \
	smoketest-stm32 \
	smoketest-avr \
	smoketest-esp \
	smoketest-riscv \
	smoketest-wasm \
	smoketest-flags

.PHONY: smoketest $(SMOKETEST_SUBTARGETS)

# Build a binary for every supported board. Run `make -j smoketest` to
# build the groups in parallel.
smoketest: testchdir $(SMOKETEST_SUBTARGETS)

# Each group writes to its own output name so that a parallel build does
# not let one group overwrite the output of another.
SMOKE_OUT = build/smoke/test

build/smoke:
	@mkdir -p build/smoke

smoketest-selftest: | build/smoke
	$(TINYGO) version
	$(TINYGO) targets > /dev/null
	# regression test for #2892
	cd tests/testing/recurse && ($(TINYGO) test ./... > recurse.log && cat recurse.log && test $$(wc -l < recurse.log) = 2 && rm recurse.log)
	# compile-only platform-independent examples
	cd tests/text/template/smoke && $(TINYGO) test -c && rm -f smoke.test
	# regression test for #2563
	cd tests/os/smoke && $(TINYGO) test -c -target=pybadge && rm smoke.test

smoketest-examples: SMOKE_OUT = build/smoke/examples
smoketest-examples: | build/smoke
	# test all examples (except pwm)
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pga2350             examples/echo
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10040            examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10040            examples/adc
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10040            examples/blinkm
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10040            examples/blinky2
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10040            examples/button
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10040            examples/button2
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10040            examples/echo
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10040            examples/echo2
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=circuitplay-express examples/i2s
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10040            examples/mcp3008
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10040            examples/memstats
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=microbit            examples/microbit-blink
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10040            examples/pininterrupt
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=nano-rp2040         examples/rtcinterrupt
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10040            examples/machinetest
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10040            examples/systick
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10040            examples/test
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10040            examples/time-offset
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=wioterminal         examples/hid-mouse
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=wioterminal         examples/hid-keyboard
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=feather-rp2040      examples/i2c-target
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=feather-rp2040      examples/watchdog
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=feather-rp2040      examples/device-id
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pico2-ice           examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build             -o $(SMOKE_OUT).efi -target=uefi-amd64          examples/test
	@$(MD5SUM) $(SMOKE_OUT).efi

smoketest-wasm-sim: SMOKE_OUT = build/smoke/wasm-sim
smoketest-wasm-sim: | build/smoke
	# test simulated boards on play.tinygo.org
ifneq ($(WASM), 0)
	GOOS=js GOARCH=wasm $(TINYGO) build -size short -o $(SMOKE_OUT).wasm -tags=arduino_uno          examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).wasm
	GOOS=js GOARCH=wasm $(TINYGO) build -size short -o $(SMOKE_OUT).wasm -tags=hifive1b             examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).wasm
	GOOS=js GOARCH=wasm $(TINYGO) build -size short -o $(SMOKE_OUT).wasm -tags=reelboard            examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).wasm
	GOOS=js GOARCH=wasm $(TINYGO) build -size short -o $(SMOKE_OUT).wasm -tags=microbit             examples/microbit-blink
	@$(MD5SUM) $(SMOKE_OUT).wasm
	GOOS=js GOARCH=wasm $(TINYGO) build -size short -o $(SMOKE_OUT).wasm -tags=circuitplay_express  examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).wasm
	GOOS=js GOARCH=wasm $(TINYGO) build -size short -o $(SMOKE_OUT).wasm -tags=circuitplay_bluefruit examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).wasm
	GOOS=js GOARCH=wasm $(TINYGO) build -size short -o $(SMOKE_OUT).wasm -tags=mch2022              examples/machinetest
	@$(MD5SUM) $(SMOKE_OUT).wasm
	GOOS=js GOARCH=wasm $(TINYGO) build -size short -o $(SMOKE_OUT).wasm -tags=gopher_badge         examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).wasm
	GOOS=js GOARCH=wasm $(TINYGO) build -size short -o $(SMOKE_OUT).wasm -tags=pico                 examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).wasm
	GOOS=js GOARCH=wasm $(TINYGO) build -size short -o $(SMOKE_OUT).wasm -tags=xiao_esp32s3         examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).wasm
endif

smoketest-nrf: SMOKE_OUT = build/smoke/nrf
smoketest-nrf: | build/smoke
	# test all targets/boards
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10040-s132v6     examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=microbit            examples/echo
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=microbit-s110v8     examples/echo
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=microbit-v2         examples/microbit-blink
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=microbit-v2-s113v7  examples/microbit-blink
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=microbit-v2-s140v7  examples/microbit-blink
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=nrf52840-mdk        examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=btt-skr-pico        examples/uart
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10031            examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=reelboard           examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=reelboard           examples/blinky2
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10056            examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10056            examples/blinky2
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10059            examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10059            examples/blinky2
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=bluemicro840        examples/blinky2
	@$(MD5SUM) $(SMOKE_OUT).hex

smoketest-samd: SMOKE_OUT = build/smoke/samd
smoketest-samd: | build/smoke
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=itsybitsy-m0        examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=feather-m0          examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=trinket-m0          examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=gemma-m0            examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=circuitplay-express examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=circuitplay-bluefruit examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=circuitplay-express examples/i2s
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=clue-alpha          examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).gba -target=gameboy-advance     examples/gba-display
	@$(MD5SUM) $(SMOKE_OUT).gba
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=grandcentral-m4     examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=itsybitsy-m4        examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=feather-m4          examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=matrixportal-m4     examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pybadge             examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=metro-m4-airlift    examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pyportal            examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=particle-argon      examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=particle-boron      examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=particle-xenon      examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pinetime            examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=x9pro               examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10056-s140v7     examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10059-s140v7     examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=reelboard-s140v7    examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=wioterminal         examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pygamer             examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=xiao                examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=xiao-ble-plus       examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=rak4631             examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=circuitplay-express examples/dac
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pyportal            examples/dac
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=feather-nrf52840  	examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=feather-nrf52840-sense examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=itsybitsy-nrf52840  examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=qtpy                examples/machinetest
	@$(MD5SUM) $(SMOKE_OUT).hex

smoketest-nxp: SMOKE_OUT = build/smoke/nxp
smoketest-nxp: | build/smoke
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=teensy41            examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=teensy40            examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=teensy36            examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=p1am-100            examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=atsame54-xpro       examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=atsame54-xpro       examples/can
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=feather-m4-can      examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=feather-m4-can      examples/caninterrupt
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=arduino-nano33      examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=arduino-mkrwifi1010 examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex

smoketest-rp2xxx: SMOKE_OUT = build/smoke/rp2xxx
smoketest-rp2xxx: | build/smoke
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pico                examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pico -gc=leaking    examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pico-w              examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=nano-33-ble         examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=nano-rp2040         examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=feather-rp2040 		examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=qtpy-rp2040         examples/echo
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=kb2040              examples/echo
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=macropad-rp2040 	examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=badger2040          examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=badger2040-w        examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=tufty2040           examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=thingplus-rp2040    examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=xiao-rp2040         examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=waveshare-rp2040-zero examples/echo
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=challenger-rp2040    examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=trinkey-qt2040      examples/temp
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=gopher-badge      examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=gopher-arcade      examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=ae-rp2040           examples/echo
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=thumby              examples/echo
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pico2               examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pico2-w             examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=tiny2350            examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=badger2350          examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=blinky2350          examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pico-plus2          examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=metro-rp2350        examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=waveshare-rp2040-tiny examples/echo
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=vicharak_shrike-lite examples/echo
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=xiao-rp2350        examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex

smoketest-pwm-usb: SMOKE_OUT = build/smoke/pwm-usb
smoketest-pwm-usb: | build/smoke
	# test pwm
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=itsybitsy-m0        examples/pwm
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=itsybitsy-m4        examples/pwm
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=feather-m4          examples/pwm
	@$(MD5SUM) $(SMOKE_OUT).hex
	# test usb
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=feather-nrf52840    examples/hid-keyboard
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=circuitplay-express examples/hid-keyboard
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=feather-nrf52840    examples/usb-midi
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pico    			examples/usb-storage
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pico2    			examples/usb-storage
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=nrf52840-s140v6-uf2-generic	examples/machinetest
	@$(MD5SUM) $(SMOKE_OUT).hex

smoketest-stm32: SMOKE_OUT = build/smoke/stm32
smoketest-stm32: | build/smoke
ifneq ($(STM32), 0)
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=bluepill            examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=feather-stm32f405   examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=lgt92               examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=nucleo-f103rb       examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=nucleo-f722ze       examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=nucleo-h753zi       examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=nucleo-l031k6       examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=nucleo-l432kc       examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=nucleo-l476rg       examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=nucleo-l552ze       examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=nucleo-wl55jc       examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=stm32f4disco        examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=stm32f4disco        examples/blinky2
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=stm32f4disco-1      examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=stm32f4disco-1      examples/pwm
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=stm32f469disco      examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=lorae5              examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=swan                examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=mksnanov3           examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=stm32l0x1           examples/serial
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=stm32u031           examples/empty
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=arduino-uno-q       examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=arduino-uno-q       examples/serial
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=arduino-uno-q       examples/blinkm
	@$(MD5SUM) $(SMOKE_OUT).hex
endif

smoketest-avr: SMOKE_OUT = build/smoke/avr
smoketest-avr: | build/smoke
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=atmega328pb         examples/blinkm
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=atmega1284p         examples/machinetest
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=arduino-uno         examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=arduino-leonardo    examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=arduino-uno         examples/pwm
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=arduino-uno -scheduler=tasks  examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=arduino-mega1280    examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=arduino-mega1280    examples/pwm
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=arduino-nano        examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=attiny1616          examples/empty
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=digispark           examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=digispark           examples/pwm
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=digispark           examples/mcp3008
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=digispark -gc=leaking examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex

smoketest-esp: SMOKE_OUT = build/smoke/esp
smoketest-esp: | build/smoke
ifneq ($(XTENSA), 0)
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=esp32-generic       examples/machinetest
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=esp32-coreboard-v2  examples/adc
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=esp32c3-generic     examples/machinetest
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=esp32s3-generic     examples/machinetest
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=esp32-mini32      	examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=esp32c3-supermini   examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=esp32c3-supermini   examples/blinkm
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=nodemcu             examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target m5stack-core2       examples/machinetest
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target m5stack             examples/machinetest
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target m5stamp-s3a         examples/machinetest
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target m5stick-c           examples/machinetest
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target m5paper             examples/machinetest
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target mch2022             examples/machinetest
	@$(MD5SUM) $(SMOKE_OUT).bin
	# xiao-esp32c6
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=xiao-esp32c6      	examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=xiao-esp32c6   		examples/blinkm
	@$(MD5SUM) $(SMOKE_OUT).bin
	# xiao-esp32s3
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=xiao-esp32s3   		examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=xiao-esp32s3   		examples/blinkm
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=xiao-esp32s3   		examples/mcp3008
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=xiao-esp32s3   		examples/pwm
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=xiao-esp32s3   		examples/adc
	@$(MD5SUM) $(SMOKE_OUT).bin
	# esp32s3-supermini
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=esp32s3-supermini	    examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=esp32s3-supermini	    examples/blinkm
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=esp32s3-supermini	    examples/mcp3008
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=esp32s3-supermini   	examples/adc
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=esp32s3-box-3       examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).bin
endif

smoketest-riscv: SMOKE_OUT = build/smoke/riscv
smoketest-riscv: | build/smoke
	# esp32c3-supermini
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=esp32c3-supermini	    examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=esp32c3-supermini	    examples/blinkm
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=esp32c3-supermini	    examples/mcp3008
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=esp32c3-supermini   	examples/pwm
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=esp32c3-supermini   	examples/adc
	@$(MD5SUM) $(SMOKE_OUT).bin

	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=esp-c3-32s-kit      examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=qtpy-esp32c3        examples/machinetest
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=m5stamp-c3          examples/machinetest
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=xiao-esp32c3        examples/machinetest
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=esp32-c3-devkit-rust-1 examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=esp32c3-12f         examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).bin

	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=makerfabs-esp32c3spi35 examples/machinetest
	@$(MD5SUM) $(SMOKE_OUT).bin
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=hifive1b            examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=maixbit             examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=tkey                examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=elecrow-rp2040      examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=elecrow-rp2350      examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=hw-651              examples/machinetest
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=hw-651-s110v8       examples/machinetest
	@$(MD5SUM) $(SMOKE_OUT).hex

smoketest-wasm: SMOKE_OUT = build/smoke/wasm
smoketest-wasm: | build/smoke
ifneq ($(WASM), 0)
	$(TINYGO) build -size short -o $(SMOKE_OUT).wasm -target=wasm               examples/wasm/export
	$(TINYGO) build -size short -o $(SMOKE_OUT).wasm -target=wasm               examples/wasm/main
	$(TINYGO) build -size short -o $(SMOKE_OUT).wasm -target=wasm-unknown       examples/hello-wasm-unknown
endif

smoketest-flags: SMOKE_OUT = build/smoke/flags
smoketest-flags: | build/smoke
	# test various compiler flags
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10040 -gc=none -scheduler=none examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10040 -opt=1     examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10040 -serial=none examples/echo
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10040 -serial=rtt examples/echo
	@$(MD5SUM) $(SMOKE_OUT).hex
	$(TINYGO) build             -o $(SMOKE_OUT).nro -target=nintendoswitch      examples/echo2
	@$(MD5SUM) $(SMOKE_OUT).nro
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10040 -opt=0     ./testdata/stdlib.go
	@$(MD5SUM) $(SMOKE_OUT).hex
	GOOS=linux GOARCH=arm $(TINYGO) build -size short -o $(SMOKE_OUT).elf       ./testdata/cgo
	GOOS=linux GOARCH=mips    $(TINYGO) build -size short -o $(SMOKE_OUT).elf   ./testdata/cgo
	GOOS=windows GOARCH=amd64 $(TINYGO) build -size short -o $(SMOKE_OUT).exe   ./testdata/cgo
	GOOS=windows GOARCH=arm64 $(TINYGO) build -size short -o $(SMOKE_OUT).exe   ./testdata/cgo
	GOOS=darwin GOARCH=amd64 $(TINYGO) build  -size short -o $(SMOKE_OUT)       ./testdata/cgo
	GOOS=darwin GOARCH=arm64 $(TINYGO) build  -size short -o $(SMOKE_OUT)       ./testdata/cgo
ifneq ($(OS),Windows_NT)
	# TODO: this does not yet work on Windows. Somehow, unused functions are
	# not garbage collected.
	$(TINYGO) build -o $(SMOKE_OUT).elf -gc=leaking -scheduler=none examples/serial
endif

# A representative board for each processor architecture. This answers the
# question "can TinyGo build a binary for each architecture" at a fraction of
# the cost of the full smoke test, which runs separately on Linux.
.PHONY: smoketest-quick
smoketest-quick: SMOKE_OUT = build/smoke/quick
smoketest-quick: testchdir | build/smoke
	$(TINYGO) version
	$(TINYGO) targets > /dev/null
	# regression test for #2892
	cd tests/testing/recurse && ($(TINYGO) test ./... > recurse.log && cat recurse.log && test $$(wc -l < recurse.log) = 2 && rm recurse.log)
	# nrf51, Cortex-M0
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=microbit            examples/microbit-blink
	@$(MD5SUM) $(SMOKE_OUT).hex
	# nrf52, Cortex-M4
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10040            examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	# nrf52840 with SoftDevice, which uses a different memory layout
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pca10056-s140v7     examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	# samd21, Cortex-M0+
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=circuitplay-express examples/i2s
	@$(MD5SUM) $(SMOKE_OUT).hex
	# samd51, Cortex-M4 with hardware floating point
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=feather-m4          examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	# same5x, which adds CAN
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=atsame54-xpro       examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	# rp2040, dual Cortex-M0+ with a second stage bootloader
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pico                examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	# rp2350, Cortex-M33
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=pico2               examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	# i.MX RT1062, Cortex-M7
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=teensy41            examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	# ARM7TDMI, the only target that is not Thumb-2
	$(TINYGO) build -size short -o $(SMOKE_OUT).gba -target=gameboy-advance     examples/gba-display
	@$(MD5SUM) $(SMOKE_OUT).gba
	# AVR, ATmega328p
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=arduino-uno         examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	# AVR, ATtiny85, which has a smaller instruction set than the ATmega
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=digispark           examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	# RISC-V 32 bit, esp32c3. Not behind the XTENSA flag, so the compatibility
	# test also builds it.
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=esp32c3-supermini   examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).bin
	# RISC-V 32 bit, SiFive E31
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=hifive1b            examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	# RISC-V 64 bit, the only 64 bit baremetal target
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=maixbit             examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	# x86-64 with PE/COFF output
	$(TINYGO) build -size short -o $(SMOKE_OUT).efi -target=uefi-amd64          examples/test
	@$(MD5SUM) $(SMOKE_OUT).efi
	# aarch64
	$(TINYGO) build             -o $(SMOKE_OUT).nro -target=nintendoswitch      examples/echo2
	@$(MD5SUM) $(SMOKE_OUT).nro
	# cross compilation with cgo
	GOOS=linux GOARCH=arm $(TINYGO) build -size short -o $(SMOKE_OUT).elf       ./testdata/cgo
ifneq ($(STM32), 0)
	# STM32F1, Cortex-M3
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=bluepill            examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
	# STM32F4, Cortex-M4 with hardware floating point
	$(TINYGO) build -size short -o $(SMOKE_OUT).hex -target=stm32f4disco        examples/blinky1
	@$(MD5SUM) $(SMOKE_OUT).hex
endif
ifneq ($(XTENSA), 0)
	# Xtensa LX6
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=esp32-coreboard-v2  examples/adc
	@$(MD5SUM) $(SMOKE_OUT).bin
	# Xtensa LX7
	$(TINYGO) build -size short -o $(SMOKE_OUT).bin -target=esp32s3-generic     examples/machinetest
	@$(MD5SUM) $(SMOKE_OUT).bin
endif
ifneq ($(WASM), 0)
	# wasm with the JavaScript host bindings
	$(TINYGO) build -size short -o $(SMOKE_OUT).wasm -target=wasm              examples/wasm/main
	# wasm without a host, so without any imports
	$(TINYGO) build -size short -o $(SMOKE_OUT).wasm -target=wasm-unknown      examples/hello-wasm-unknown
endif
ifneq ($(OS),Windows_NT)
	# TODO: this does not yet work on Windows. Somehow, unused functions are
	# not garbage collected.
	$(TINYGO) build -o $(SMOKE_OUT).elf -gc=leaking -scheduler=none examples/serial
endif
