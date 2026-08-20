# PY32 target generator

This command generates TinyGo target JSON and linker scripts for Puya PY32
microcontrollers. Run it from the repository root:

```sh
make gen-target-py32
```

The normalized [device table](devices.csv) was derived from the `<device>` and
`<variant>` leaves in the official Puya CMSIS-Pack PDSC files. Memory values are
the inherited default `IROM1` and `IRAM1` regions. The three PY32E407 parts have
multiple adjacent default RAM regions; their linker definitions combine those
regions into one contiguous RAM range.

The table contains 87 concrete devices associated with 40 SVD families. The
pack also supplies `PY32F001xx.svd` without a corresponding PDSC device, so the
generator creates its family target but cannot create a concrete memory target.
Together these inputs produce family targets for all 41 available SVDs.

Cortex-M0+ families inherit `py32`; PY32E407, PY32F403, and PY32F410 families
inherit `py32-m4`, which uses TinyGo's standard soft-float Cortex-M4 ABI. Targets
whose SVD has no `GPIO.AFRH` register receive the `no_gpio_afrh` build tag.

Puya's SVDs are inconsistent about GPIO `groupName`: some describe identical
ports as `GPIOA_Type`, `GPIOB_Type`, and so on, while others use one
`GPIO_Type`. The `py32-svd` updater validates the register structures and
patches the published SVDs to use the common `GPIO` group. This is a correction
to the vendor metadata, not a target capability, so neither TinyGo's generic
SVD generator nor the target build tags need PY32-specific handling for it.

Real register-layout differences are selected by generated capability tags.
These cover the `OSPEEDR`/`OSPDDER` GPIO spelling, RCC GPIO and UART clock
registers, `USART` versus `UART` blocks, split USART receive/transmit data
registers, and HSI selector availability. Keep these classifications in the
target generator rather than adding long family expressions to machine files.

Existing PY32 target files are retained for compatibility. In particular, their
established aliases and flashing commands are not regenerated. Newly generated
targets intentionally omit flashing commands because a compatible programmer
identifier has not been verified for every Puya part.

## Machine support status

A minimal program compiles for all 87 concrete targets. GPIO, RCC, runtime
clock setup, and the default serial block are selected according to each SVD's
register layout. This is compile-time coverage, not hardware validation for
every device. In particular, PY32F410 SVD interrupt metadata stops before the
serial interrupts, so its default USART is configured without receive
interrupts rather than assigning an unverified IRQ number. The M4 SVDs also do
not expose the M0+ `ICSCR.HSI_FS` selector, so the runtime preserves their reset
clock configuration and initially reports the vendor-defined 8 MHz reset
frequency. Applications that establish a different clock tree must update
`machine.CPUFrequencyHz`, then reconfigure SysTick and frequency-dependent
peripherals.