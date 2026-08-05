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

Existing PY32 target files are retained for compatibility. In particular, their
established aliases and flashing commands are not regenerated. Newly generated
targets intentionally omit flashing commands because a compatible programmer
identifier has not been verified for every Puya part.

## Machine support status

All target definitions can be loaded independently of machine-driver support.
A minimal program currently compiles for 68 of the 87 concrete targets. The
remaining targets are E407, F001C, F002C, F032, F403, F410, L090, T020, T090,
and T092 devices whose raw SVD register layouts do not match assumptions in the
existing generic PY32 GPIO, RCC, or UART implementations. Their definitions are
included deliberately so machine support can be added without revisiting target
and linker metadata.