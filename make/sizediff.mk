# Programs that the binary size workflow compares between dev and a PR.
# The workflow adds this output to the drivers smoke-test output and passes both to tools/sizediff.

SIZEDIFF_OUT = build/sizediff.elf

.PHONY: sizediff
sizediff:
	@mkdir -p build
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=z -target=hifive1b        examples/echo
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=z -target=microbit        examples/serial
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=z -target=wioterminal     examples/pininterrupt
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=z -target=pico            examples/serial
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=2 -target=pico            examples/serial
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=z -target=pico2           examples/serial
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=2 -target=pico2           examples/serial
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=z -target=esp32c3-generic examples/serial
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=2 -target=esp32c3-generic examples/serial
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=z -target=cortex-m-qemu   examples/size/slice
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=2 -target=cortex-m-qemu   examples/size/slice
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=z -target=cortex-m-qemu   examples/size/defer
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=z -target=cortex-m-qemu   examples/size/contains
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=z -target=pico2           examples/size/contains
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=z -target=pico2           examples/size/os
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=2 -target=pico2           examples/size/os
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=z -target=pico2           examples/size/timer
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=2 -target=pico2           examples/size/timer
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=z -target=cortex-m-qemu   examples/size/corpus
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=2 -target=cortex-m-qemu   examples/size/corpus
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=z -target=riscv-qemu      examples/size/corpus
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=2 -target=riscv-qemu      examples/size/corpus
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=z -target=cortex-m-qemu   examples/size/init-recover
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=2 -target=cortex-m-qemu   examples/size/init-recover
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=z -target=pico2           examples/size/init-regexp
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=2 -target=pico2           examples/size/init-regexp
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=z -target=cortex-m-qemu   examples/size/init-crc
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=2 -target=cortex-m-qemu   examples/size/init-crc
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=z -target=cortex-m-qemu   examples/size/init-map
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=2 -target=cortex-m-qemu   examples/size/init-map
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=z -target=cortex-m-qemu   examples/size/init-errors
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=2 -target=cortex-m-qemu   examples/size/init-errors
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=z -target=cortex-m-qemu   examples/size/init-config
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=2 -target=cortex-m-qemu   examples/size/init-config
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=z -target=cortex-m-qemu   examples/size/init-strings
	$(TINYGO) build -size short -o $(SIZEDIFF_OUT) -opt=2 -target=cortex-m-qemu   examples/size/init-strings
