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
 # $Date: 2017-01-17 11:51:57 -0600 (Tue, 17 Jan 2017) $ 
 # $Revision: 25948 $
 #
 ###############################################################################

TARGET=MAX32620

COMPILER=GCC

TARGET_UC:=$(shell echo $(TARGET) | tr a-z A-Z)
TARGET_LC:=$(shell echo $(TARGET) | tr A-Z a-z)

# This is the name of the build output file
PROJECT=$(TARGET_LC)_usbLib

# This is the path to the peripheral drivers
PERIPH_DRIVER_DIR=../$(TARGET_UC)PeriphDriver

BUILD_DIR=./Build

# This is the path to the CMSIS root directory
CMSIS_ROOT=../CMSIS

# Source files for this test (add path to VPATH below)
SRCS = usb_event.c
SRCS += usb.c
SRCS += enumerate.c
SRCS += cdc_acm.c
SRCS += fifo.c

# Where to find source files for this project
VPATH  = src/core/arm
VPATH += src/core/
VPATH += src/devclass
VPATH += src/enumerate

# Where to find header files for this project

IPATH  = include/core/arm
IPATH += include/core
IPATH += include/devclass
IPATH += include/enumerate
IPATH += $(CMSIS_ROOT)/Device/Maxim/$(TARGET)/Include
IPATH += $(CMSIS_ROOT)/Include
IPATH += $(PERIPH_DRIVER_DIR)/Include

# Include the rules for building for this target
include $(CMSIS_ROOT)/Device/Maxim/$(TARGET_UC)/Source/$(COMPILER)/$(TARGET_LC).mk

# Build this as a library
.DEFAULT_GOAL := lib
