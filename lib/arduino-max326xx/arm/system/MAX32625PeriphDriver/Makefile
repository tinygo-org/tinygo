################################################################################
 # Copyright (C) 2016 Maxim Integrated Products, Inc., All Rights Reserved.
 #
 # Permission is hereby granted, free of charge, to any person obtaining a
 # copy of this software and associated documentation files (the "Software"),
 # to deal in the Software without restriction, including without limitation
 # the rights to use, copy, modify, merge, publish, distribute, sublicense,
 # and/or sell copies of the Software, and to permit persons to whom the
 # Software is furnished to do so, subject to the following conditions:
 #
 # The above copyright notice and this permission notice shall be included
 # in all copies or substantial portions of the Software.
 #
 # THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS
 # OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
 # MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 # IN NO EVENT SHALL MAXIM INTEGRATED BE LIABLE FOR ANY CLAIM, DAMAGES
 # OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 # ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 # OTHER DEALINGS IN THE SOFTWARE.
 #
 # Except as contained in this notice, the name of Maxim Integrated
 # Products, Inc. shall not be used except as stated in the Maxim Integrated
 # Products, Inc. Branding Policy.
 #
 # The mere transfer of this software does not imply any licenses
 # of trade secrets, proprietary technology, copyrights, patents,
 # trademarks, maskwork rights, or any other form of intellectual
 # property whatsoever. Maxim Integrated Products, Inc. retains all
 # ownership rights.
 #
 # $Date: 2017-03-15 15:57:56 -0500 (Wed, 15 Mar 2017) $ 
 # $Revision: 27038 $
 #
 ###############################################################################

# This is the name of the build output file
PROJECT=PeriphDriver

TARGET=MAX32625

COMPILER=GCC

BUILD_DIR=./Build

CMSIS_ROOT=../CMSIS

ARDUINO_CORE = ../../cores/arduino

VARIANT_PATH = ../../variants/max32625_mbed

# Source files for this test (add path to VPATH below)
SRCS += adc.c
SRCS += aes.c
SRCS += clkman.c
SRCS += crc.c
SRCS += flc.c
SRCS += gpio.c
SRCS += i2cm.c
SRCS += i2cs.c
SRCS += icc.c
SRCS += ioman.c
SRCS += lp.c
SRCS += maa.c
SRCS += mxc_assert.c
SRCS += mxc_sys.c
SRCS += nvic_table.c
SRCS += owm.c
SRCS += pmu.c
SRCS += prng.c
SRCS += pt.c
SRCS += rtc.c
SRCS += spim.c
SRCS += spis.c
SRCS += spix.c
SRCS += tmr.c
SRCS += tmr_utils.c
SRCS += uart.c
SRCS += wdt.c
SRCS += wdt2.c

# Where to find source files for this project
VPATH = Source

# Where to find header files for this project
IPATH = Include
IPATH += $(ARDUINO_CORE)
IPATH += $(VARIANT_PATH)

# Include the rules for building for this target
TARGET_UC:=$(shell echo $(TARGET) | tr a-z A-Z)
TARGET_LC:=$(shell echo $(TARGET) | tr A-Z a-z)
include $(CMSIS_ROOT)/Device/Maxim/$(TARGET_UC)/Source/$(COMPILER)/$(TARGET_LC).mk

# Build this as a library
.DEFAULT_GOAL := lib
