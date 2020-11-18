/**
 * @file
 * @brief      MAX3263X device specific definitions for the core, peripherals,
 *             features, memory, and IRQs.
 */
/* *****************************************************************************
 * Copyright (C) 2016 Maxim Integrated Products, Inc., All Rights Reserved.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS
 * OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
 * MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL MAXIM INTEGRATED BE LIABLE FOR ANY CLAIM, DAMAGES
 * OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 * ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 *
 * Except as contained in this notice, the name of Maxim Integrated
 * Products, Inc. shall not be used except as stated in the Maxim Integrated
 * Products, Inc. Branding Policy.
 *
 * The mere transfer of this software does not imply any licenses
 * of trade secrets, proprietary technology, copyrights, patents,
 * trademarks, maskwork rights, or any other form of intellectual
 * property whatsoever. Maxim Integrated Products, Inc. retains all
 * ownership rights.
 *
 *
 * $Date: 2017-06-02 08:54:20 -0500 (Fri, 02 Jun 2017) $
 * $Revision: 28314 $
 *
*************************************************************************** */

/* **** Includes **** */
#include <stdint.h>

/* Define to prevent redundant inclusion */
#ifndef _MAX3263X_H_
#define _MAX3263X_H_

/**
  * @ingroup    syscfg
  * @defgroup   product_name MAX3263X
  * @brief      MAX3263X device specific definitions for the core, peripherals,
  *             features, memory, and IRQs.
  * @details    The <b><em>MAX32630/MAX32631</em></b> is an ARM&reg;
  *             Cortex&reg;-M4F 32-bit microcontroller with a floating point
  *             unit, ideal for the emerging category of wearable medical and
  *             fitness applications. The architecture combines ultra-low power
  *             high-efficiency signal processing functionality with
  *             significantly reduced power consumption and ease of use. The
  *             device features four powerful and flexible power modes. A
  *             peripheral management unit (PMU) enables intelligent peripheral
  *             control with up to six channels to significantly reduce power
  *             consumption. Built-in dynamic clock gating and
  *             firmware-controlled power gating allows the user to optimize
  *             power for the specific application. Multiple SPI, UART and
  *             I&sup2;C serial interfaces, as well as 1-Wire&reg; master and
  *             USB, allow for interconnection to a wide variety of external
  *             sensors. A four-input, 10-bit ADC with selectable references is
  *             available to monitor analog input from external sensors and
  *             meters. The small 100-ball WLP package provides a tiny, 4.37mm x
  *             4.37mm footprint.  The <b><em>MAX32630/MAX32631</em></b> include
  *             a hardware AES engine. The <b><em>MAX32631</em></b> is a secure
  *             version of the <b><em>MAX32630</em></b>. It incorporates a trust
  *             protection unit (TPU) with encryption and advanced security
  *             features. These features include a modular arithmetic
  *             accelerator (MAA) for fast ECDSA, a hardware PRNG entropy
  *             generator, and a secure boot loader. 
  * @{
  */
#ifndef  FALSE
/**
  * @internal False
  */
#define  FALSE      (0) 
#endif

#ifndef  TRUE
/**
 * @internal True 
 */
#define  TRUE       (1)
#endif

/* COMPILER SPECIFIC DEFINES (IAR, ARMCC and GNUC) */
#if defined ( __GNUC__ )
#define __weak __attribute__((weak))                /**< GNUC weak function keyword. */
#elif defined ( __CC_ARM)
#define inline __inline                             /**< inline keyword for Keil compiler. */
#pragma anon_unions
#endif
/**@}*/
/**
 * @ingroup product_name
 * @defgroup nvic_table Nested Interrupt Vector Table (NVIC)
 * Device specific interrupt request NVIC entries.
 * @{
 */
/**
 * \MXIM_Device Nested Interrupt Vector Table (NVIC).    
 * @details
 * NVIC Peripheral Entry numbers and Offsets are shown in the table below.
 * 
 * | Entry   | Offset  | Peripheral                            |
 * |-------: | ------: | :------------------------------------ |
 * | 0x10    | 0x0040  |  CLKMAN                               |
 * | 0x11    | 0x0044  |  PWRMAN                               |
 * | 0x12    | 0x0048  |  Flash Controller                     |
 * | 0x13    | 0x004C  |  RTC Counter match with Compare 0     |
 * | 0x14    | 0x0050  |  RTC Counter match with Compare 1     |
 * | 0x15    | 0x0054  |  RTC Prescaler interval compare match |
 * | 0x16    | 0x0058  |  RTC Overflow                         |
 * | 0x17    | 0x005C  |  Peripheral Management Unit (PMU/DMA) |
 * | 0x18    | 0x0060  |  USB                                  |
 * | 0x19    | 0x0064  |  AES                                  |
 * | 0x1A    | 0x0068  |  MAA                                  |
 * | 0x1B    | 0x006C  |  Watchdog 0 timeout                   |
 * | 0x1C    | 0x0070  |  Watchdog 0 pre-window (fed too early)|
 * | 0x1D    | 0x0074  |  Watchdog 1 timeout                   |
 * | 0x1E    | 0x0078  |  Watchdog 1 pre-window (fed too early)|
 * | 0x1F    | 0x007C  |  GPIO Port 0                          |
 * | 0x20    | 0x0080  |  GPIO Port 1                          |
 * | 0x21    | 0x0084  |  GPIO Port 2                          |
 * | 0x22    | 0x0088  |  GPIO Port 3                          |
 * | 0x23    | 0x008C  |  GPIO Port 4                          |
 * | 0x24    | 0x0090  |  GPIO Port 5                          |
 * | 0x25    | 0x0094  |  GPIO Port 6                          |
 * | 0x26    | 0x0098  |  Timer 0 (32-bit, 16-bit #0)          |
 * | 0x27    | 0x009C  |  Timer 0 (16-bit #1)                  |
 * | 0x28    | 0x00A0  |  Timer 1 (32-bit, 16-bit #0)          |
 * | 0x29    | 0x00A4  |  Timer 1 (16-bit #1)                  |
 * | 0x2A    | 0x00A8  |  Timer 2 (32-bit, 16-bit #0)          |
 * | 0x2B    | 0x00AC  |  Timer 2 (16-bit #1)                  |
 * | 0x2C    | 0x00B0  |  Timer 3 (32-bit, 16-bit #0)          |
 * | 0x2D    | 0x00B4  |  Timer 3 (16-bit #1)                  |
 * | 0x2E    | 0x00B8  |  Timer 4 (32-bit, 16-bit #0)          |
 * | 0x2F    | 0x00BC  |  Timer 4 (16-bit #1)                  |
 * | 0x30    | 0x00C0  |  Timer 5 (32-bit, 16-bit #0)          |
 * | 0x31    | 0x00C4  |  Timer 5 (16-bit #1)                  |
 * | 0x32    | 0x00C8  |  UART 0                               |
 * | 0x33    | 0x00CC  |  UART 1                               |
 * | 0x34    | 0x00D0  |  UART 2                               |
 * | 0x35    | 0x00D4  |  UART 3                               |
 * | 0x36    | 0x00D8  |  Pulse Trains                         |
 * | 0x37    | 0x00DC  |  I2C Master 0                         |
 * | 0x38    | 0x00E0  |  I2C Master 1                         |
 * | 0x39    | 0x00E4  |  I2C Master 2                         |
 * | 0x3A    | 0x00E8  |  I2C Slave                            |
 * | 0x3B    | 0x00EC  |  SPI Master 0                         |
 * | 0x3C    | 0x00F0  |  SPI Master 1                         |
 * | 0x3D    | 0x00F4  |  SPI Master 2                         |
 * | 0x3E    | 0x00F8  |  SPI Bridge                           |
 * | 0x3F    | 0x00FC  |  1-Wire Master                        |
 * | 0x40    | 0x0100  |  ADC                                  |
 * | 0x41    | 0x0104  |  SPI Slave                            |
 * | 0x42    | 0x0108  |  GPIO Port 7                          |
 * | 0x43    | 0x010C  |  GPIO Port 8                          |            
 */

/**
 * Enumeration type of all \MXIM_Device NVIC entries.
 */
typedef enum {
    NonMaskableInt_IRQn    = -14, /**< ARM Core : Non-maskable IRQ  */
    HardFault_IRQn         = -13, /**< ARM Core : Hard Fault IRQ  */
    MemoryManagement_IRQn  = -12, /**< ARM Core : Memory Management IRQ  */
    BusFault_IRQn          = -11, /**< ARM Core : Bus Fault IRQ  */
    UsageFault_IRQn        = -10, /**< ARM Core : Usage Fault IRQ  */
    SVCall_IRQn            = -5,  /**< ARM Core : SVCall IRQ  */
    DebugMonitor_IRQn      = -4,  /**< ARM Core : Debug Monitor IRQ          */
    PendSV_IRQn            = -2,  /**< ARM Core : PendSV IRQ                 */
    SysTick_IRQn           = -1,  /**< ARM Core : SysTick IRQ                */
    CLKMAN_IRQn = 0,              /**< CLKMAN                                */
    PWRMAN_IRQn,                  /**< PWRMAN                                */
    FLC_IRQn,                     /**< Flash Controller                      */
    RTC0_IRQn,                    /**< RTC Counter match with Compare 0      */
    RTC1_IRQn,                    /**< RTC Counter match with Compare 1      */
    RTC2_IRQn,                    /**< RTC Prescaler interval compare match  */
    RTC3_IRQn,                    /**< RTC Overflow                          */
    PMU_IRQn,                     /**< Peripheral Management Unit (PMU/DMA)  */
    USB_IRQn,                     /**< USB                                   */
    AES_IRQn,                     /**< AES                                   */
    MAA_IRQn,                     /**< MAA                                   */
    WDT0_IRQn,                    /**< Watchdog 0 timeout                   */
    WDT0_P_IRQn,                  /**< Watchdog 0 pre-window (fed too early) */
    WDT1_IRQn,                    /**< Watchdog 1 timeout                    */
    WDT1_P_IRQn,                  /**< Watchdog 1 pre-window (fed too early) */
    GPIO_P0_IRQn,                 /**< GPIO Port 0                           */
    GPIO_P1_IRQn,                 /**< GPIO Port 1                           */
    GPIO_P2_IRQn,                 /**< GPIO Port 2                           */
    GPIO_P3_IRQn,                 /**< GPIO Port 3                           */
    GPIO_P4_IRQn,                 /**< GPIO Port 4                           */
    GPIO_P5_IRQn,                 /**< GPIO Port 5                           */
    GPIO_P6_IRQn,                 /**< GPIO Port 6                           */
    TMR0_0_IRQn,                  /**< Timer 0 (32-bit, 16-bit #0)           */
    TMR0_1_IRQn,                  /**< Timer 0 (16-bit #1)                   */
    TMR1_0_IRQn,                  /**< Timer 1 (32-bit, 16-bit #0)           */
    TMR1_1_IRQn,                  /**< Timer 1 (16-bit #1)                   */
    TMR2_0_IRQn,                  /**< Timer 2 (32-bit, 16-bit #0)           */
    TMR2_1_IRQn,                  /**< Timer 2 (16-bit #1)                   */
    TMR3_0_IRQn,                  /**< Timer 3 (32-bit, 16-bit #0)           */
    TMR3_1_IRQn,                  /**< Timer 3 (16-bit #1)                   */
    TMR4_0_IRQn,                  /**< Timer 4 (32-bit, 16-bit #0)           */
    TMR4_1_IRQn,                  /**< Timer 4 (16-bit #1)                   */
    TMR5_0_IRQn,                  /**< Timer 5 (32-bit, 16-bit #0)           */
    TMR5_1_IRQn,                  /**< Timer 5 (16-bit #1)                   */
    UART0_IRQn,                   /**< UART 0                                */
    UART1_IRQn,                   /**< UART 1                                */
    UART2_IRQn,                   /**< UART 2                                */
    UART3_IRQn,                   /**< UART 3                                */
    PT_IRQn,                      /**< Pulse Trains                          */
    I2CM0_IRQn,                   /**< I2C Master 0                          */
    I2CM1_IRQn,                   /**< I2C Master 1                          */
    I2CM2_IRQn,                   /**< I2C Master 2                          */
    I2CS_IRQn,                    /**< I2C Slave                             */
    SPIM0_IRQn,                   /**< SPI Master 0                          */
    SPIM1_IRQn,                   /**< SPI Master 1                          */
    SPIM2_IRQn,                   /**< SPI Master 2                          */
    SPIB_IRQn,                    /**< SPI Bridge                            */
    OWM_IRQn,                     /**< 1-Wire Master                         */
    AFE_IRQn,                     /**< ADC                                   */
    SPIS_IRQn,                    /**< SPI Slave                             */
    GPIO_P7_IRQn,                 /**< GPIO Port 7                           */
    GPIO_P8_IRQn,                 /**< GPIO Port 8                           */
    MXC_IRQ_EXT_COUNT             /**< Total number of non-core IRQ vectors.                                                */
} IRQn_Type;

#define MXC_IRQ_COUNT (MXC_IRQ_EXT_COUNT + 16) /**< Total number of device IRQs inclusive of core and non-core IRQ vectors. */
/**@}end of group nvic_table*/

/* ================================================================================ */
/* ================      Processor and Core Peripheral Section     ================ */
/* ================================================================================ */
/**
 * @ingroup product_name
 * @defgroup Cortex_M4 Cortex-M Configuration
 * @{
 */
/* ----------------------  Configuration of the Cortex-M Processor and Core Peripherals  ---------------------- */
#define __CM4_REV                 0x0100            /**< Cortex-M4 Core Revision                                */
#define __MPU_PRESENT                  1            /**< MPU is present                                         */
#define __NVIC_PRIO_BITS               3            /**< Number of Bits used for IRQ Priority Levels            */
#define __Vendor_SysTickConfig         0            /**< Using standard CMSIS SysTickConfig                     */
#define __FPU_PRESENT                  1            /**< FPU is Present                                         */
/**@} end of ingroup Cortex_M4*/
#include <core_cm4.h>                               /*!< Cortex-M4 processor and core peripherals               */
#include "system_max3263x.h"                        /*!< System Header                                          */


/* ================================================================================ */
/* ==================       Device Specific Memory Section       ================== */
/* ================================================================================ */
/**
 * @ingroup product_name
 * @{
 */
#define MXC_FLASH_MEM_BASE         0x00000000UL   /**< Internal Flash Memory Start Address. */
#define MXC_FLASH_PAGE_SIZE        0x00002000UL   /**< Internal Flash Memory Page Size. */
#define MXC_FLASH_FULL_MEM_SIZE    0x00200000UL   /**< Internal Flash Memory Size. */
#define MXC_SYS_MEM_BASE           0x20000000UL   /**< System Memory Start Address. */
#define MXC_SRAM_FULL_MEM_SIZE     0x00080000UL   /**< Internal SRAM Size. */
#define MXC_EXT_FLASH_MEM_BASE     0x10000000UL   /**< External Flash Memory Start Address, SPIX interface. */

/* ================================================================================ */
/* ================       Device Specific Peripheral Section       ================ */
/* ================================================================================ */


/*
   Base addresses and configuration settings for all MAX3263X peripheral modules.
*/


/* *************************************************************************** */
/*                                                     System Manager Settings */

#define MXC_BASE_SYSMAN                  ((uint32_t)0x40000000UL)
#define MXC_SYSMAN                       ((mxc_sysman_regs_t *)MXC_BASE_SYSMAN)



/* *************************************************************************** */
/*                                                        System Clock Manager */

#define MXC_BASE_CLKMAN                  ((uint32_t)0x40000400UL)
#define MXC_CLKMAN                       ((mxc_clkman_regs_t *)MXC_BASE_CLKMAN)



/* *************************************************************************** */
/*                                                        System Power Manager */

#define MXC_BASE_PWRMAN                  ((uint32_t)0x40000800UL)
#define MXC_PWRMAN                       ((mxc_pwrman_regs_t *)MXC_BASE_PWRMAN)



/* *************************************************************************** */
/*                                                             Real Time Clock */

#define MXC_BASE_RTCTMR                  ((uint32_t)0x40000A00UL)
#define MXC_RTCTMR                       ((mxc_rtctmr_regs_t *)MXC_BASE_RTCTMR)
#define MXC_BASE_RTCCFG                  ((uint32_t)0x40000A70UL)
#define MXC_RTCCFG                       ((mxc_rtccfg_regs_t *)MXC_BASE_RTCCFG)

#define MXC_RTCTMR_GET_IRQ(i)            (IRQn_Type)(i == 0 ? RTC0_IRQn :   \
                                          i == 1 ? RTC1_IRQn :              \
                                          i == 2 ? RTC2_IRQn :              \
                                          i == 3 ? RTC3_IRQn : 0)



/* *************************************************************************** */
/*                                                             Power Sequencer */

#define MXC_BASE_PWRSEQ                  ((uint32_t)0x40000A30UL)
#define MXC_PWRSEQ                       ((mxc_pwrseq_regs_t *)MXC_BASE_PWRSEQ)



/* *************************************************************************** */
/*                                                          System I/O Manager */
/**@} end of ingroup product_name*/
/**
 * @ingroup ioman_registers
 * @{
 */
#define MXC_BASE_IOMAN                   ((uint32_t)0x40000C00UL)               /**< Base Peripheral Address for IOMAN */
#define MXC_IOMAN                        ((mxc_ioman_regs_t *)MXC_BASE_IOMAN)   /**< Pointer to the #mxc_ioman_regs_t structure representing the IOMAN Registers. */
/**@}*/

/**
 * @ingroup product_name
 * @{
 */
/* *************************************************************************** */
/*                                                       Shadow Trim Registers */

#define MXC_BASE_TRIM                    ((uint32_t)0x40001000UL)
#define MXC_TRIM                         ((mxc_trim_regs_t *)MXC_BASE_TRIM)



/* *************************************************************************** */
/*                                                            Flash Controller */

#define MXC_BASE_FLC                     ((uint32_t)0x40002000UL)
#define MXC_FLC                          ((mxc_flc_regs_t *)MXC_BASE_FLC)

#define MXC_FLC_PAGE_SIZE_SHIFT          (13)
#define MXC_FLC_PAGE_SIZE                (1 << MXC_FLC_PAGE_SIZE_SHIFT)
#define MXC_FLC_PAGE_ERASE_MSK           ((~(1 << (MXC_FLC_PAGE_SIZE_SHIFT - 1))) >> MXC_FLC_PAGE_SIZE_SHIFT) << MXC_FLC_PAGE_SIZE_SHIFT



/* *************************************************************************** */
/*                                                           Instruction Cache */

#define MXC_BASE_ICC                     ((uint32_t)0x40003000UL)
#define MXC_ICC                          ((mxc_icc_regs_t *)MXC_BASE_ICC)


/**@} end of ingroup product_name*/
/* *************************************************************************** */
/*                                                           SPI XIP Interface */
/**
 * @ingroup spix_registers
 * @{
 */
#define MXC_BASE_SPIX                    ((uint32_t)0x40004000UL)           /**< SPIX Base Peripheral Address. */
#define MXC_SPIX                         ((mxc_spix_regs_t *)MXC_BASE_SPIX) /**< SPIX pointer to the #mxc_spix_regs_t register structure type. */
/**@} end of ingroup spix_registers*/

/**
 * @ingroup product_name
 * @{
 */
/* *************************************************************************** */
/*                                                  Peripheral Management Unit */

#define MXC_CFG_PMU_CHANNELS             (6)

#define MXC_BASE_PMU0                    ((uint32_t)0x40005000UL)
#define MXC_PMU0                         ((mxc_pmu_regs_t *)MXC_BASE_PMU0)
#define MXC_BASE_PMU1                    ((uint32_t)0x40005020UL)
#define MXC_PMU1                         ((mxc_pmu_regs_t *)MXC_BASE_PMU1)
#define MXC_BASE_PMU2                    ((uint32_t)0x40005040UL)
#define MXC_PMU2                         ((mxc_pmu_regs_t *)MXC_BASE_PMU2)
#define MXC_BASE_PMU3                    ((uint32_t)0x40005060UL)
#define MXC_PMU3                         ((mxc_pmu_regs_t *)MXC_BASE_PMU3)
#define MXC_BASE_PMU4                    ((uint32_t)0x40005080UL)
#define MXC_PMU4                         ((mxc_pmu_regs_t *)MXC_BASE_PMU4)
#define MXC_BASE_PMU5                    ((uint32_t)0x400050A0UL)
#define MXC_PMU5                         ((mxc_pmu_regs_t *)MXC_BASE_PMU5)

#define MXC_PMU_GET_BASE(i)              ((i) == 0  ? MXC_BASE_PMU0 :          \
                                          (i) == 1  ? MXC_BASE_PMU1 :          \
                                          (i) == 2  ? MXC_BASE_PMU2 :          \
                                          (i) == 3  ? MXC_BASE_PMU3 :          \
                                          (i) == 4  ? MXC_BASE_PMU4 :          \
                                          (i) == 5  ? MXC_BASE_PMU5 : 0)

#define MXC_PMU_GET_PMU(i)               ((i) == 0 ? MXC_PMU0 :                \
                                          (i) == 1 ? MXC_PMU1 :                \
                                          (i) == 2 ? MXC_PMU2 :                \
                                          (i) == 3 ? MXC_PMU3 :                \
                                          (i) == 4 ? MXC_PMU4 :                \
                                          (i) == 5 ? MXC_PMU5 : 0)

#define MXC_PMU_GET_IDX(p)               ((p) == MXC_PMU0 ? 0 :                \
                                          (p) == MXC_PMU1 ? 1 :                \
                                          (p) == MXC_PMU2 ? 2 :                \
                                          (p) == MXC_PMU3 ? 3 :                \
                                          (p) == MXC_PMU4 ? 4 :                \
                                          (p) == MXC_PMU5 ? 5 : -1)

/* *************************************************************************** */
/*                                                       USB Device Controller */

#define MXC_BASE_USB                     ((uint32_t)0x40100000UL)
#define MXC_USB                          ((mxc_usb_regs_t *)MXC_BASE_USB)

#define MXC_USB_MAX_PACKET               (64)
#define MXC_USB_NUM_EP                   (8)



/* *************************************************************************** */
/*                                                        CRC-16/CRC-32 Engine */

#define MXC_BASE_CRC                     ((uint32_t)0x40006000UL)
#define MXC_CRC                          ((mxc_crc_regs_t *)MXC_BASE_CRC)
#define MXC_BASE_CRC_DATA                ((uint32_t)0x40101000UL)
#define MXC_CRC_DATA                     ((mxc_crc_data_regs_t *)MXC_BASE_CRC_DATA)

/* *************************************************************************** */
/*                                       Pseudo-random number generator (PRNG) */

#define MXC_BASE_PRNG                     ((uint32_t)0x40007000UL)
#define MXC_PRNG                          ((mxc_prng_regs_t *)MXC_BASE_PRNG)

/* *************************************************************************** */
/*                                                    AES Cryptographic Engine */

#define MXC_BASE_AES                     ((uint32_t)0x40007400UL)
#define MXC_AES                          ((mxc_aes_regs_t *)MXC_BASE_AES)
#define MXC_BASE_AES_MEM                 ((uint32_t)0x40102000UL)
#define MXC_AES_MEM                      ((mxc_aes_mem_regs_t *)MXC_BASE_AES_MEM)

/* *************************************************************************** */
/*                                                    MAA Cryptographic Engine */

#define MXC_BASE_MAA                     ((uint32_t)0x40007800UL)
#define MXC_MAA                          ((mxc_maa_regs_t *)MXC_BASE_MAA)
#define MXC_BASE_MAA_MEM                 ((uint32_t)0x40102800UL)
#define MXC_MAA_MEM                      ((mxc_maa_mem_regs_t *)MXC_BASE_MAA_MEM)

/* *************************************************************************** */
/*                                                 Trust Protection Unit (TPU) */

#define MXC_BASE_TPU                     ((uint32_t)0x40007000UL)
#define MXC_TPU                          ((mxc_tpu_regs_t *)MXC_BASE_TPU)
#define MXC_BASE_TPU_TSR                 ((uint32_t)0x40007C00UL)
#define MXC_TPU_TSR                      ((mxc_tpu_tsr_regs_t *)MXC_BASE_TPU_TSR)
/**@} end of ingroup product_name*/
/* *************************************************************************** */
/*                                                             Watchdog Timers */
/**
 * @ingroup wdt_registers
 * @{
 */
#define MXC_CFG_WDT_INSTANCES            (2)                                        /**< Define for the number of timers on the \MXIM_Device */

#define MXC_BASE_WDT0                    ((uint32_t)0x40008000UL)                   /**< Base Peripheral Address for WDT 0 */
#define MXC_WDT0                         ((mxc_wdt_regs_t *)MXC_BASE_WDT0)          /**< Pointer to the #mxc_wdt_regs_t structure representing WDT0 Registers. */
#define MXC_BASE_WDT1                    ((uint32_t)0x40009000UL)                   /**< Base Peripheral Address for WDT 1 */
#define MXC_WDT1                         ((mxc_wdt_regs_t *)MXC_BASE_WDT1)          /**< Pointer to the #mxc_wdt_regs_t structure representing WDT1 Registers. */
/**
 * Macro that returns the WDT[i] IRQ, where i=0 to i < #MXC_CFG_WDT_INSTANCES.
 */
#define MXC_WDT_GET_IRQ(i)               (IRQn_Type)((i) == 0 ? WDT0_IRQn :    \
                                          (i) == 1 ? WDT1_IRQn : 0)

#define MXC_WDT_GET_IRQ_P(i)             (IRQn_Type)((i) == 0 ? WDT0_P_IRQn :  \
                                          (i) == 1 ? WDT1_P_IRQn : 0)
/**
 * Macro to return the base address for a requested Watchdog Timer index number. 
 * @p i WDT instance number.
 * @p returns the base peripheral address for the requested Watchdog Timer instance. 
 */
#define MXC_WDT_GET_BASE(i)              ((i) == 0 ? MXC_BASE_WDT0 :           \
                                          (i) == 1 ? MXC_BASE_WDT1 : 0)
/**
 * Macro to return a pointer to the #mxc_tmr_regs_t object for the requested Watchdog Timer. 
 * @p i Watchdog Timer instance number.  
 * @p returns a pointer to a #mxc_wdt_regs_t for the requested WDT number. 
 */
#define MXC_WDT_GET_WDT(i)               ((i) == 0 ? MXC_WDT0 :                \
                                          (i) == 1 ? MXC_WDT1 : 0)
/**
 * Macro to return the index number for a given #mxc_wdt_regs_t structure.
 * @p p pointer to a #mxc_wdt_regs_t structure. 
 * @p returns a watchdog timer instance number. 
 */
#define MXC_WDT_GET_IDX(i)               ((i) == MXC_WDT0 ? 0:                 \
                                          (i) == MXC_WDT1 ? 1: -1)

/**@} end of ingroup wdt_registers */
/* *************************************************************************** */
/*                                                    Always-On Watchdog Timer */
/**
 * @ingroup wdt2_registers
 * @{
 */
#define MXC_BASE_WDT2                    ((uint32_t)0x40007C60UL)                   /**< Base Peripheral Address for WDT 2 */ 
#define MXC_WDT2                         ((mxc_wdt2_regs_t *)MXC_BASE_WDT2)         /**< Pointer to the #mxc_wdt2_regs_t structure representing the WDT2 hardware registers. */  
/**@} end of ingroup wdt2_registers */


/* *************************************************************************** */
/*                                            General Purpose I/O Ports (GPIO) */
/**
 * @ingroup gpio_registers
 * @{
 */
#define MXC_GPIO_NUM_PORTS               (9)                                        /**< Number of GPIO Ports for the \MXIM_Device. */
#define MXC_GPIO_MAX_PINS_PER_PORT       (8)                                        /**< Number of port pins per port for the \MXIM_Device */

#define MXC_BASE_GPIO                    ((uint32_t)0x4000A000UL)                   /**< GPIO Base Peripheral Offset */
#define MXC_GPIO                         ((mxc_gpio_regs_t *)MXC_BASE_GPIO)         /**< Pointer to the #mxc_gpio_regs_t object representing GPIO Registers. */
/**
 * @def MXC_GPIO_GET_IRQ(i)
 * Macro that returns the IRQn_Type for GPIO port number \a i, where <tt> i=0 to i < #MXC_GPIO_NUM_PORTS </tt>.
 */
#define MXC_GPIO_GET_IRQ(i)              (IRQn_Type)((i) == 0 ? GPIO_P0_IRQn :  \
                                          (i) == 1 ? GPIO_P1_IRQn :             \
                                          (i) == 2 ? GPIO_P2_IRQn :             \
                                          (i) == 3 ? GPIO_P3_IRQn :             \
                                          (i) == 4 ? GPIO_P4_IRQn :             \
                                          (i) == 5 ? GPIO_P5_IRQn :             \
                                          (i) == 6 ? GPIO_P6_IRQn :             \
                                          (i) == 7 ? GPIO_P7_IRQn :             \
                                          (i) == 8 ? GPIO_P8_IRQn : 0)

/**@} end of ingroup gpio_registers */

/* *************************************************************************** */
/*                                                    16/32 bit Timer/Counters */
/**
 * @ingroup tmr_registers
 * @{
 */
#define MXC_CFG_TMR_INSTANCES            (6)  /**< Define for the number of timers on the \MXIM_Device */
#define MXC_BASE_TMR0                    ((uint32_t)0x4000B000UL)           /**< Base Address for Timer 0 */
#define MXC_TMR0                         ((mxc_tmr_regs_t *)MXC_BASE_TMR0)  /**< Pointer to a #mxc_tmr_regs_t structure representing Timer 0 */
#define MXC_BASE_TMR1                    ((uint32_t)0x4000C000UL)           /**< Base Address for Timer 1 */
#define MXC_TMR1                         ((mxc_tmr_regs_t *)MXC_BASE_TMR1)  /**< Pointer to a #mxc_tmr_regs_t structure representing Timer 1 */
#define MXC_BASE_TMR2                    ((uint32_t)0x4000D000UL)           /**< Base Address for Timer 2 */
#define MXC_TMR2                         ((mxc_tmr_regs_t *)MXC_BASE_TMR2)  /**< Pointer to a #mxc_tmr_regs_t structure representing Timer 2 */
#define MXC_BASE_TMR3                    ((uint32_t)0x4000E000UL)           /**< Base Address for Timer 3 */
#define MXC_TMR3                         ((mxc_tmr_regs_t *)MXC_BASE_TMR3)  /**< Pointer to a #mxc_tmr_regs_t structure representing Timer 3 */
#define MXC_BASE_TMR4                    ((uint32_t)0x4000F000UL)           /**< Base Address for Timer 4 */
#define MXC_TMR4                         ((mxc_tmr_regs_t *)MXC_BASE_TMR4)  /**< Pointer to a #mxc_tmr_regs_t structure representing Timer 4 */
#define MXC_BASE_TMR5                    ((uint32_t)0x40010000UL)           /**< Base Address for Timer 5 */
#define MXC_TMR5                         ((mxc_tmr_regs_t *)MXC_BASE_TMR5)  /**< Pointer to a #mxc_tmr_regs_t structure representing Timer 5 */

/**
 * Macro that returns an #IRQn_Type for the requested 32-bit timer interrupt.  
 */
#define MXC_TMR_GET_IRQ_32(i)            (IRQn_Type)((i) == 0 ? TMR0_0_IRQn :     \
                                                     (i) == 1 ? TMR1_0_IRQn :     \
                                                     (i) == 2 ? TMR2_0_IRQn :     \
                                                     (i) == 3 ? TMR3_0_IRQn :     \
                                                     (i) == 4 ? TMR4_0_IRQn :     \
                                                     (i) == 5 ? TMR5_0_IRQn : 0)
/**
 * Macro that returns an IRQn_Type for the requested 16-bit timer interrupt number.  
 */
#define MXC_TMR_GET_IRQ_16(i)            (IRQn_Type)((i) == 0  ? TMR0_0_IRQn :    \
                                                     (i) == 1  ? TMR1_0_IRQn :    \
                                                     (i) == 2  ? TMR2_0_IRQn :    \
                                                     (i) == 3  ? TMR3_0_IRQn :    \
                                                     (i) == 4  ? TMR4_0_IRQn :    \
                                                     (i) == 5  ? TMR5_0_IRQn :    \
                                                     (i) == 6  ? TMR0_1_IRQn :    \
                                                     (i) == 7  ? TMR1_1_IRQn :    \
                                                     (i) == 8  ? TMR2_1_IRQn :    \
                                                     (i) == 9  ? TMR3_1_IRQn :    \
                                                     (i) == 10 ? TMR4_1_IRQn :    \
                                                     (i) == 11 ? TMR5_1_IRQn : 0)
/**
 * Macro to return the base address for a given Timer index number. 
 * @p i Timer instance number.
 * @p returns the base peripheral address for the requested timer instance. 
 */
#define MXC_TMR_GET_BASE(i)              ((i) == 0 ? MXC_BASE_TMR0 :   \
                                          (i) == 1 ? MXC_BASE_TMR1 :   \
                                          (i) == 2 ? MXC_BASE_TMR2 :   \
                                          (i) == 3 ? MXC_BASE_TMR3 :   \
                                          (i) == 4 ? MXC_BASE_TMR4 :   \
                                          (i) == 5 ? MXC_BASE_TMR5 : 0)
/**
 * Macro to return a pointer to the #mxc_tmr_regs_t structure for a given Timer Instance. 
 * @p i Timer instance number.  
 * @p returns a pointer to a #mxc_tmr_regs_t for the requested timer number. 
 */
#define MXC_TMR_GET_TMR(i)               ((i) == 0 ? MXC_TMR0 :        \
                                          (i) == 1 ? MXC_TMR1 :        \
                                          (i) == 2 ? MXC_TMR2 :        \
                                          (i) == 3 ? MXC_TMR3 :        \
                                          (i) == 4 ? MXC_TMR4 :        \
                                          (i) == 5 ? MXC_TMR5 : 0)
/**
 * Macro to return the index number for a given pointer to a #mxc_tmr_regs_t structure.
 * @p p pointer to a #mxc_tmr_regs_t structure. 
 * @p returns a timer instance number. 
 */
#define MXC_TMR_GET_IDX(p)               ((p) == MXC_TMR0 ? 0 :        \
                                          (p) == MXC_TMR1 ? 1 :        \
                                          (p) == MXC_TMR2 ? 2 :        \
                                          (p) == MXC_TMR3 ? 3 :        \
                                          (p) == MXC_TMR4 ? 4 :        \
                                          (p) == MXC_TMR5 ? 5 : -1)

/**@} end of ingroup tmr_registers */

/**
 * @ingroup product_name
 * @{
 */
/* *************************************************************************** */
/*                                                      Pulse Train Generation */
#define MXC_CFG_PT_INSTANCES             (16)

#define MXC_BASE_PTG                     ((uint32_t)0x40011000UL)
#define MXC_PTG                          ((mxc_ptg_regs_t *)MXC_BASE_PTG)
#define MXC_BASE_PT0                     ((uint32_t)0x40011020UL)
#define MXC_PT0                          ((mxc_pt_regs_t *)MXC_BASE_PT0)
#define MXC_BASE_PT1                     ((uint32_t)0x40011040UL)
#define MXC_PT1                          ((mxc_pt_regs_t *)MXC_BASE_PT1)
#define MXC_BASE_PT2                     ((uint32_t)0x40011060UL)
#define MXC_PT2                          ((mxc_pt_regs_t *)MXC_BASE_PT2)
#define MXC_BASE_PT3                     ((uint32_t)0x40011080UL)
#define MXC_PT3                          ((mxc_pt_regs_t *)MXC_BASE_PT3)
#define MXC_BASE_PT4                     ((uint32_t)0x400110A0UL)
#define MXC_PT4                          ((mxc_pt_regs_t *)MXC_BASE_PT4)
#define MXC_BASE_PT5                     ((uint32_t)0x400110C0UL)
#define MXC_PT5                          ((mxc_pt_regs_t *)MXC_BASE_PT5)
#define MXC_BASE_PT6                     ((uint32_t)0x400110E0UL)
#define MXC_PT6                          ((mxc_pt_regs_t *)MXC_BASE_PT6)
#define MXC_BASE_PT7                     ((uint32_t)0x40011100UL)
#define MXC_PT7                          ((mxc_pt_regs_t *)MXC_BASE_PT7)
#define MXC_BASE_PT8                     ((uint32_t)0x40011120UL)
#define MXC_PT8                          ((mxc_pt_regs_t *)MXC_BASE_PT8)
#define MXC_BASE_PT9                     ((uint32_t)0x40011140UL)
#define MXC_PT9                          ((mxc_pt_regs_t *)MXC_BASE_PT9)
#define MXC_BASE_PT10                    ((uint32_t)0x40011160UL)
#define MXC_PT10                         ((mxc_pt_regs_t *)MXC_BASE_PT10)
#define MXC_BASE_PT11                    ((uint32_t)0x40011180UL)
#define MXC_PT11                         ((mxc_pt_regs_t *)MXC_BASE_PT11)
#define MXC_BASE_PT12                    ((uint32_t)0x400111A0UL)
#define MXC_PT12                         ((mxc_pt_regs_t *)MXC_BASE_PT12)
#define MXC_BASE_PT13                    ((uint32_t)0x400111C0UL)
#define MXC_PT13                         ((mxc_pt_regs_t *)MXC_BASE_PT13)
#define MXC_BASE_PT14                    ((uint32_t)0x400111E0UL)
#define MXC_PT14                         ((mxc_pt_regs_t *)MXC_BASE_PT14)
#define MXC_BASE_PT15                    ((uint32_t)0x40011200UL)
#define MXC_PT15                         ((mxc_pt_regs_t *)MXC_BASE_PT15)

#define MXC_PT_GET_BASE(i)               ((i) == 0  ? MXC_BASE_PT0  :           \
                                          (i) == 1  ? MXC_BASE_PT1  :           \
                                          (i) == 2  ? MXC_BASE_PT2  :           \
                                          (i) == 3  ? MXC_BASE_PT3  :           \
                                          (i) == 4  ? MXC_BASE_PT4  :           \
                                          (i) == 5  ? MXC_BASE_PT5  :           \
                                          (i) == 6  ? MXC_BASE_PT6  :           \
                                          (i) == 7  ? MXC_BASE_PT7  :           \
                                          (i) == 8  ? MXC_BASE_PT8  :           \
                                          (i) == 9  ? MXC_BASE_PT9  :           \
                                          (i) == 10 ? MXC_BASE_PT10 :           \
                                          (i) == 11 ? MXC_BASE_PT11 :           \
                                          (i) == 12 ? MXC_BASE_PT12 :           \
                                          (i) == 13 ? MXC_BASE_PT13 :           \
                                          (i) == 14 ? MXC_BASE_PT14 :           \
                                          (i) == 15 ? MXC_BASE_PT15 : 0)

#define MXC_PT_GET_PT(i)                 ((i) == 0  ? MXC_PT0  :                \
                                          (i) == 1  ? MXC_PT1  :                \
                                          (i) == 2  ? MXC_PT2  :                \
                                          (i) == 3  ? MXC_PT3  :                \
                                          (i) == 4  ? MXC_PT4  :                \
                                          (i) == 5  ? MXC_PT5  :                \
                                          (i) == 6  ? MXC_PT6  :                \
                                          (i) == 7  ? MXC_PT7  :                \
                                          (i) == 8  ? MXC_PT8  :                \
                                          (i) == 9  ? MXC_PT9  :                \
                                          (i) == 10 ? MXC_PT10 :                \
                                          (i) == 11 ? MXC_PT11 :                \
                                          (i) == 12 ? MXC_PT12 :                \
                                          (i) == 13 ? MXC_PT13 :                \
                                          (i) == 14 ? MXC_PT14 :                \
                                          (i) == 15 ? MXC_PT15 : 0)

#define MXC_PT_GET_IDX(p)                ((p) == MXC_PT0  ? 0  :                \
                                          (p) == MXC_PT1  ? 1  :                \
                                          (p) == MXC_PT2  ? 2  :                \
                                          (p) == MXC_PT3  ? 3  :                \
                                          (p) == MXC_PT4  ? 4  :                \
                                          (p) == MXC_PT5  ? 5  :                \
                                          (p) == MXC_PT6  ? 6  :                \
                                          (p) == MXC_PT7  ? 7  :                \
                                          (p) == MXC_PT8  ? 8  :                \
                                          (p) == MXC_PT9  ? 9  :                \
                                          (p) == MXC_PT10 ? 10 :                \
                                          (p) == MXC_PT11 ? 11 :                \
                                          (p) == MXC_PT12 ? 12 :                \
                                          (p) == MXC_PT13 ? 13 :                \
                                          (p) == MXC_PT14 ? 14 :                \
                                          (p) == MXC_PT15 ? 15 : -1)


/** @} end of ingroup product_name */

/* *************************************************************************** */
/*                                                UART / Serial Port Interface */

/**
 * @ingroup uart_registers
 * @{
 */
#define MXC_CFG_UART_INSTANCES           (4)                                                  /**< Number of UART Peripherals */
#define MXC_UART_FIFO_DEPTH              (32)                                                 /**< UART FIFO Size */

#define MXC_BASE_UART0                   ((uint32_t)0x40012000UL)                             /**< UART0 Base Address */
#define MXC_UART0                        ((mxc_uart_regs_t *)MXC_BASE_UART0)                  /**< Pointer to the UART0 register structure */
#define MXC_BASE_UART1                   ((uint32_t)0x40013000UL)                             /**< UART0 Base Address */
#define MXC_UART1                        ((mxc_uart_regs_t *)MXC_BASE_UART1)                  /**< Pointer to the UART0 register structure */
#define MXC_BASE_UART2                   ((uint32_t)0x40014000UL)                             /**< UART2 Base Address */
#define MXC_UART2                        ((mxc_uart_regs_t *)MXC_BASE_UART2)                  /**< Pointer to the UART0 register structure */
#define MXC_BASE_UART3                   ((uint32_t)0x40015000UL)                             /**< UART3 Base Address */
#define MXC_UART3                        ((mxc_uart_regs_t *)MXC_BASE_UART3)                  /**< Pointer to the UART0 register structure */
#define MXC_BASE_UART0_FIFO              ((uint32_t)0x40103000UL)                             /**< UART0 FIFO Base Address */
#define MXC_UART0_FIFO                   ((mxc_uart_fifo_regs_t *)MXC_BASE_UART0_FIFO)        /**< Pointer to the UART0 FIFO register structure */
#define MXC_BASE_UART1_FIFO              ((uint32_t)0x40104000UL)                             /**< UART1 FIFO Base Address */
#define MXC_UART1_FIFO                   ((mxc_uart_fifo_regs_t *)MXC_BASE_UART1_FIFO)        /**< Pointer to the UART1 FIFO register structure */
#define MXC_BASE_UART2_FIFO              ((uint32_t)0x40105000UL)                             /**< UART2 FIFO Base Address */
#define MXC_UART2_FIFO                   ((mxc_uart_fifo_regs_t *)MXC_BASE_UART2_FIFO)        /**< Pointer to the UART2 FIFO register structure */
#define MXC_BASE_UART3_FIFO              ((uint32_t)0x40106000UL)                             /**< UART3 FIFO Base Address */
#define MXC_UART3_FIFO                   ((mxc_uart_fifo_regs_t *)MXC_BASE_UART3_FIFO)        /**< Pointer to the UART3 FIFO register structure */

/**
 * @def MXC_UART_GET_IRQ(i)
 * Returns the Interrupt Vector of (IRQn_Type) for the \a i UART port number requested.
 *
 */
#define MXC_UART_GET_IRQ(i)              (IRQn_Type)((i) == 0 ? UART0_IRQn :  \
                                          (i) == 1 ? UART1_IRQn :             \
                                          (i) == 2 ? UART2_IRQn :             \
                                          (i) == 3 ? UART3_IRQn : 0)
/**
 * @def MXC_UART_GET_BASE(i)
 * Returns the base peripheral address for the \a i UART port number requested.
 *
 */
#define MXC_UART_GET_BASE(i)             ((i) == 0 ? MXC_BASE_UART0 :         \
                                          (i) == 1 ? MXC_BASE_UART1 :         \
                                          (i) == 2 ? MXC_BASE_UART2 :         \
                                          (i) == 3 ? MXC_BASE_UART3 : 0)

/**
 * @def MXC_UART_GET_UART(i)
 * Returns a pointer to the @see mxc_uart_regs_t "UART register structure" for the \a i UART port number requested.
 *
 */
#define MXC_UART_GET_UART(i)             ((i) == 0 ? MXC_UART0 :              \
                                          (i) == 1 ? MXC_UART1 :              \
                                          (i) == 2 ? MXC_UART2 :              \
                                          (i) == 3 ? MXC_UART3 : 0)
/**
 * @def MXC_UART_GET_IDX(p)
 * Returns the port number (index), of the UART base address, \a p parameter.
 *
 */
#define MXC_UART_GET_IDX(p)              ((p) == MXC_UART0 ? 0 :              \
                                          (p) == MXC_UART1 ? 1 :              \
                                          (p) == MXC_UART2 ? 2 :              \
                                          (p) == MXC_UART3 ? 3 : -1)

/**
 * @def MXC_UART_GET_BASE_FIFO(i)
 * Returns the base FIFO address of the UART port number \a i parameter. 
 *
 */
#define MXC_UART_GET_BASE_FIFO(i)        ((i) == 0 ? MXC_BASE_UART0_FIFO :    \
                                          (i) == 1 ? MXC_BASE_UART1_FIFO :    \
                                          (i) == 2 ? MXC_BASE_UART2_FIFO :    \
                                          (i) == 3 ? MXC_BASE_UART3_FIFO : 0)
/**
 * @def MXC_UART_GET_FIFO(i)
 * Returns a pointer to the @ref mxc_uart_fifo_regs_t "UART FIFO register structure" for the UART port number \a i parameter. 
 *
 */
#define MXC_UART_GET_FIFO(i)             ((i) == 0 ? MXC_UART0_FIFO :         \
                                          (i) == 1 ? MXC_UART1_FIFO :         \
                                          (i) == 2 ? MXC_UART2_FIFO :         \
                                          (i) == 3 ? MXC_UART3_FIFO : 0)
/**@} */

/** 
 * @ingroup product_name 
 * @{ 
 */
/* *************************************************************************** */
/*                                                        I2C Master Interface */

#define MXC_CFG_I2CM_INSTANCES           (3)
#define MXC_I2CM_FIFO_DEPTH              (8)

#define MXC_BASE_I2CM0                   ((uint32_t)0x40016000UL)
#define MXC_I2CM0                        ((mxc_i2cm_regs_t *)MXC_BASE_I2CM0)
#define MXC_BASE_I2CM1                   ((uint32_t)0x40017000UL)
#define MXC_I2CM1                        ((mxc_i2cm_regs_t *)MXC_BASE_I2CM1)
#define MXC_BASE_I2CM2                   ((uint32_t)0x40018000UL)
#define MXC_I2CM2                        ((mxc_i2cm_regs_t *)MXC_BASE_I2CM2)
#define MXC_BASE_I2CM0_FIFO              ((uint32_t)0x40107000UL)
#define MXC_I2CM0_FIFO                   ((mxc_i2cm_fifo_regs_t *)MXC_BASE_I2CM0_FIFO)
#define MXC_BASE_I2CM1_FIFO              ((uint32_t)0x40108000UL)
#define MXC_I2CM1_FIFO                   ((mxc_i2cm_fifo_regs_t *)MXC_BASE_I2CM1_FIFO)
#define MXC_BASE_I2CM2_FIFO              ((uint32_t)0x40109000UL)
#define MXC_I2CM2_FIFO                   ((mxc_i2cm_fifo_regs_t *)MXC_BASE_I2CM2_FIFO)

#define MXC_I2CM_GET_IRQ(i)              (IRQn_Type)((i) == 0 ? I2CM0_IRQn :   \
                                          (i) == 1 ? I2CM1_IRQn :              \
                                          (i) == 2 ? I2CM2_IRQn : 0)

#define MXC_I2CM_GET_BASE(i)             ((i) == 0 ? MXC_BASE_I2CM0 :          \
                                          (i) == 1 ? MXC_BASE_I2CM1 :          \
                                          (i) == 2 ? MXC_BASE_I2CM2 : 0)

#define MXC_I2CM_GET_I2CM(i)             ((i) == 0 ? MXC_I2CM0 :               \
                                          (i) == 1 ? MXC_I2CM1 :               \
                                          (i) == 2 ? MXC_I2CM2 : 0)

#define MXC_I2CM_GET_IDX(p)              ((p) == MXC_I2CM0 ? 0 :               \
                                          (p) == MXC_I2CM1 ? 1 :               \
                                          (p) == MXC_I2CM2 ? 2 : -1)

#define MXC_I2CM_GET_BASE_FIFO(i)        ((i) == 0 ? MXC_BASE_I2CM0_FIFO :     \
                                          (i) == 1 ? MXC_BASE_I2CM1_FIFO :     \
                                          (i) == 2 ? MXC_BASE_I2CM2_FIFO : 0)

#define MXC_I2CM_GET_FIFO(i)             ((i) == 0 ? MXC_I2CM0_FIFO :          \
                                          (i) == 1 ? MXC_I2CM1_FIFO :          \
                                          (i) == 2 ? MXC_I2CM2_FIFO : 0)



/* *************************************************************************** */
/*                                          I2C Slave Interface (Mailbox type) */

#define MXC_CFG_I2CS_INSTANCES           (1)
#define MXC_CFG_I2CS_BUFFER_SIZE         (32)

#define MXC_BASE_I2CS                    ((uint32_t)0x40019000UL)
#define MXC_I2CS                         ((mxc_i2cs_regs_t *)MXC_BASE_I2CS)

#define MXC_I2CS_GET_IRQ(i)              (IRQn_Type)((i) == 0 ? I2CS_IRQn : 0)

#define MXC_I2CS_GET_BASE(i)             ((i) == 0 ? MXC_BASE_I2CS : 0)

#define MXC_I2CS_GET_I2CS(i)             ((i) == 0 ? MXC_I2CS : 0)

#define MXC_I2CS_GET_IDX(p)              ((p) == MXC_I2CS ? 0 : -1)

/* *************************************************************************** */
/*                                                        SPI Master Interface */

#define MXC_CFG_SPIM_INSTANCES           (3)
#define MXC_CFG_SPIM_FIFO_DEPTH          (16)

#define MXC_BASE_SPIM0                   ((uint32_t)0x4001A000UL)
#define MXC_SPIM0                        ((mxc_spim_regs_t *)MXC_BASE_SPIM0)
#define MXC_BASE_SPIM1                   ((uint32_t)0x4001B000UL)
#define MXC_SPIM1                        ((mxc_spim_regs_t *)MXC_BASE_SPIM1)
#define MXC_BASE_SPIM2                   ((uint32_t)0x4001C000UL)
#define MXC_SPIM2                        ((mxc_spim_regs_t *)MXC_BASE_SPIM2)
#define MXC_BASE_SPIM0_FIFO              ((uint32_t)0x4010A000UL)
#define MXC_SPIM0_FIFO                   ((mxc_spim_fifo_regs_t *)MXC_BASE_SPIM0_FIFO)
#define MXC_BASE_SPIM1_FIFO              ((uint32_t)0x4010B000UL)
#define MXC_SPIM1_FIFO                   ((mxc_spim_fifo_regs_t *)MXC_BASE_SPIM1_FIFO)
#define MXC_BASE_SPIM2_FIFO              ((uint32_t)0x4010C000UL)
#define MXC_SPIM2_FIFO                   ((mxc_spim_fifo_regs_t *)MXC_BASE_SPIM2_FIFO)

#define MXC_SPIM_GET_IRQ(i)              (IRQn_Type)((i) == 0 ? SPIM0_IRQn :  \
                                          (i) == 1 ? SPIM1_IRQn :             \
                                          (i) == 2 ? SPIM2_IRQn : 0)

#define MXC_SPIM_GET_BASE(i)             ((i) == 0 ? MXC_BASE_SPIM0 :         \
                                          (i) == 1 ? MXC_BASE_SPIM1 :         \
                                          (i) == 2 ? MXC_BASE_SPIM2 : 0)

#define MXC_SPIM_GET_SPIM(i)             ((i) == 0 ? MXC_SPIM0 :              \
                                          (i) == 1 ? MXC_SPIM1 :              \
                                          (i) == 2 ? MXC_SPIM2 : 0)

#define MXC_SPIM_GET_IDX(p)              ((p) == MXC_SPIM0 ? 0 :              \
                                          (p) == MXC_SPIM1 ? 1 :              \
                                          (p) == MXC_SPIM2 ? 2 : -1)

#define MXC_SPIM_GET_BASE_FIFO(i)        ((i) == 0 ? MXC_BASE_SPIM0_FIFO :    \
                                          (i) == 1 ? MXC_BASE_SPIM1_FIFO :    \
                                          (i) == 2 ? MXC_BASE_SPIM2_FIFO : 0)

#define MXC_SPIM_GET_SPIM_FIFO(i)        ((i) == 0 ? MXC_SPIM0_FIFO :         \
                                          (i) == 1 ? MXC_SPIM1_FIFO :         \
                                          (i) == 2 ? MXC_SPIM2_FIFO : 0)


/*******************************************************************************/
/*                                                         SPI Slave Interface */
#define MXC_CFG_SPIS_INSTANCES           (1)
#define MXC_CFG_SPIS_FIFO_DEPTH          (32)

#define MXC_BASE_SPIS                    ((uint32_t)0x40020000UL)
#define MXC_SPIS                         ((mxc_spis_regs_t *)MXC_BASE_SPIS)
#define MXC_BASE_SPIS_FIFO               ((uint32_t)0x4010E000UL)
#define MXC_SPIS_FIFO                    ((mxc_spis_fifo_regs_t *)MXC_BASE_SPIS_FIFO)

#define MXC_SPIS_GET_IRQ(i)              (IRQn_Type)((i) == 0 ? SPIS_IRQn : 0)

#define MXC_SPIS_GET_BASE(i)             ((i) == 0 ? MXC_BASE_SPIS : 0)

#define MXC_SPIS_GET_SPIS(i)             ((i) == 0 ? MXC_SPIS : 0)

#define MXC_SPIS_GET_IDX(p)              ((p) == MXC_SPIS ? 0 : -1)

#define MXC_SPIS_GET_BASE_FIFO(i)        ((i) == 0 ? MXC_BASE_SPIS_FIFO : 0)

#define MXC_SPIS_GET_SPIS_FIFO(i)        ((i) == 0 ? MXC_SPIS_FIFO :0)


/* *************************************************************************** */
/*                                                     1-Wire Master Interface */

#define MXC_CFG_OWM_INSTANCES            (1)

#define MXC_BASE_OWM                     ((uint32_t)0x4001E000UL)
#define MXC_OWM                          ((mxc_owm_regs_t *)MXC_BASE_OWM)

#define MXC_OWM_GET_IRQ(i)               (IRQn_Type)((i) == 0 ? OWM_IRQn : 0)

#define MXC_OWM_GET_BASE(i)              ((i) == 0 ? MXC_BASE_OWM : 0)

#define MXC_OWM_GET_OWM(i)               ((i) == 0 ? MXC_OWM : 0)

#define MXC_OWM_GET_IDX(p)               ((p) == MXC_OWM ? 0 : -1)


/* *************************************************************************** */
/*                                                                   ADC / AFE */

#define MXC_CFG_ADC_FIFO_DEPTH           (32)

#define MXC_BASE_ADC                     ((uint32_t)0x4001F000UL)
#define MXC_ADC                          ((mxc_adc_regs_t *)MXC_BASE_ADC)



/* *************************************************************************** */
/*                                                      SPIB AHB-to-SPI Bridge */

#define MXC_BASE_SPIB                    ((uint32_t)0x4000D000UL)
#define MXC_SPIB                         ((mxc_spib_regs_t *)MXC_BASE_SPIB)



/* *************************************************************************** */
/*                                                         SPI Slave Interface */

#define MXC_BASE_SPIS                    ((uint32_t)0x40020000UL)
#define MXC_SPIS                         ((mxc_spis_regs_t *)MXC_BASE_SPIS)
#define MXC_BASE_SPIS_FIFO               ((uint32_t)0x4010E000UL)
#define MXC_SPIS_FIFO                    ((mxc_spis_fifo_regs_t *)MXC_BASE_SPIS_FIFO)



/* *************************************************************************** */
/*                                                                Bit Shifting */

#define MXC_F_BIT_0        (1 << 0)
#define MXC_F_BIT_1        (1 << 1)
#define MXC_F_BIT_2        (1 << 2)
#define MXC_F_BIT_3        (1 << 3)
#define MXC_F_BIT_4        (1 << 4)
#define MXC_F_BIT_5        (1 << 5)
#define MXC_F_BIT_6        (1 << 6)
#define MXC_F_BIT_7        (1 << 7)
#define MXC_F_BIT_8        (1 << 8)
#define MXC_F_BIT_9        (1 << 9)
#define MXC_F_BIT_10       (1 << 10)
#define MXC_F_BIT_11       (1 << 11)
#define MXC_F_BIT_12       (1 << 12)
#define MXC_F_BIT_13       (1 << 13)
#define MXC_F_BIT_14       (1 << 14)
#define MXC_F_BIT_15       (1 << 15)
#define MXC_F_BIT_16       (1 << 16)
#define MXC_F_BIT_17       (1 << 17)
#define MXC_F_BIT_18       (1 << 18)
#define MXC_F_BIT_19       (1 << 19)
#define MXC_F_BIT_20       (1 << 20)
#define MXC_F_BIT_21       (1 << 21)
#define MXC_F_BIT_22       (1 << 22)
#define MXC_F_BIT_23       (1 << 23)
#define MXC_F_BIT_24       (1 << 24)
#define MXC_F_BIT_25       (1 << 25)
#define MXC_F_BIT_26       (1 << 26)
#define MXC_F_BIT_27       (1 << 27)
#define MXC_F_BIT_28       (1 << 28)
#define MXC_F_BIT_29       (1 << 29)
#define MXC_F_BIT_30       (1 << 30)
#define MXC_F_BIT_31       (1 << 31)


/* *************************************************************************** */

#define BITBAND(reg, bit)      ((0xf0000000 & (uint32_t)(reg)) + 0x2000000 + (((uint32_t)(reg) & 0x0fffffff) << 5) + ((bit) << 2))
#define MXC_CLRBIT(reg, bit)   (*(volatile uint32_t *)BITBAND(reg, bit) = 0)
#define MXC_SETBIT(reg, bit)   (*(volatile uint32_t *)BITBAND(reg, bit) = 1)
#define MXC_GETBIT(reg, bit)   (*(volatile uint32_t *)BITBAND(reg, bit))


/* *************************************************************************** */

/* SCB CPACR Register Definitions */
/* Note: Added by Maxim Integrated, as these are missing from CMSIS/Core/Include/core_cm4.h */
#define SCB_CPACR_CP10_Pos      20                              /*!< SCB CPACR: Coprocessor 10 Position */
#define SCB_CPACR_CP10_Msk      (0x3UL << SCB_CPACR_CP10_Pos)   /*!< SCB CPACR: Coprocessor 10 Mask */
#define SCB_CPACR_CP11_Pos      22                              /*!< SCB CPACR: Coprocessor 11 Position */
#define SCB_CPACR_CP11_Msk      (0x3UL << SCB_CPACR_CP11_Pos)   /*!< SCB CPACR: Coprocessor 11 Mask */
/**@} end of ingroup product_name */
#endif  /* _MAX3263X_H_ */

