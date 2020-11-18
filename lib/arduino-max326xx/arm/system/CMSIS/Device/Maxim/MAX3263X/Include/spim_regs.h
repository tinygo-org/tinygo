/**
 * @file    
 * @brief   Registers, Bit Masks and Bit Positions for the SPIM Peripheral Module.
 */

/* ****************************************************************************
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
 * $Date: 2017-02-16 08:50:20 -0600 (Thu, 16 Feb 2017) $
 * $Revision: 26455 $
 *
 *************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _MXC_SPIM_REGS_H_
#define _MXC_SPIM_REGS_H_

/* **** Includes **** */
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif


///@cond
/*
    If types are not defined elsewhere (CMSIS) define them here
*/
#ifndef __IO
#define __IO volatile
#endif
#ifndef __I
#define __I  volatile const
#endif
#ifndef __O
#define __O  volatile
#endif
#ifndef __R
#define __R  volatile const
#endif
///@endcond


/**
 * @ingroup     spim
 * @defgroup    spim_registers Registers
 * @brief       Registers, Bit Masks and Bit Positions
 * @{
 */

/**
 * Structure type to access the SPIM Peripheral Module Registers
 */
typedef struct {
    __IO uint32_t mstr_cfg;                             /**< <tt>\b 0x0000:</tt> SPIM_MSTR_CFG Register - SPI Master Configuration Register                         */
    __IO uint32_t ss_sr_polarity;                       /**< <tt>\b 0x0004:</tt> SPIM_SS_SR_POLARITY Register - SPI Master Polarity Control for SS and SR Signals   */
    __IO uint32_t gen_ctrl;                             /**< <tt>\b 0x0008:</tt> SPIM_GEN_CTRL Register - SPI Master General Control Register                       */
    __IO uint32_t fifo_ctrl;                            /**< <tt>\b 0x000C:</tt> SPIM_FIFO_CTRL Register - SPI Master FIFO Control Register                         */
    __IO uint32_t spcl_ctrl;                            /**< <tt>\b 0x0010:</tt> SPIM_SPCL_CTRL Register - SPI Master Special Mode Controls                         */
    __IO uint32_t intfl;                                /**< <tt>\b 0x0014:</tt> SPIM_INTFL Register - SPI Master Interrupt Flags                                   */
    __IO uint32_t inten;                                /**< <tt>\b 0x0018:</tt> SPIM_INTEN Register - SPI Master Interrupt Enable/Disable Settings                 */
    __IO uint32_t simple_headers;                       /**< <tt>\b 0x001C:</tt> SPIM_SIMPLE_HEADERS Register - SPI Master Simple Mode Transaction Headers          */
} mxc_spim_regs_t;


/**
 * @ingroup spim_registers
 * @defgroup spim_fifos SPIM TX and RX FIFOs
 * @brief TX and RX FIFO access for reads and writes using 8-bit, 16-bit and 32-bit data types. 
 * @{
 */
/**
 * Structure type for the SPIM Transmit and Receive FIFOs. 
 */
 typedef struct {
    union {                                             /*  0x0000-0x07FC   SPI Master FIFO Write Space for Transaction Setup   */
        __IO uint8_t  trans_8[2048];                    /**< 8-bit access to Transmit FIFO                                      */
        __IO uint16_t trans_16[1024];                   /**< 16-bit access to Transmit FIFO                                     */
        __IO uint32_t trans_32[512];                    /**< 32-bit access to Transmit FIFO                                     */
    };
    union {                                             /*  0x0800-0x0FFC   SPI Master FIFO Read Space for Results Data         */
        __IO uint8_t  rslts_8[2048];                    /**< 8-bit access to Receive FIFO                                       */
        __IO uint16_t rslts_16[1024];                   /**< 16-bit access to Receive FIFO                                      */
        __IO uint32_t rslts_32[512];                    /**< 32-bit access to Receive FIFO                                      */
    };
} mxc_spim_fifo_regs_t;
/**@} end of group spim_fifos */
/**@} end of group spim_registers */


/*
   Register offsets for module SPIM.
*/
/**
 * @ingroup    spim_registers
 * @defgroup   SPIM_Register_Offsets Register Offsets
 * @brief      SPI Master Register Offsets from the SPIM[n] Base Peripheral Address, where  \c n \c = SPIM Instance Number. 
 * @{
 */
#define MXC_R_SPIM_OFFS_MSTR_CFG                            ((uint32_t)0x00000000UL)    /**< Offset from SPIM[n] Base Address: <tt>\b 0x0000</tt>*/
#define MXC_R_SPIM_OFFS_SS_SR_POLARITY                      ((uint32_t)0x00000004UL)    /**< Offset from SPIM[n] Base Address: <tt>\b 0x0004</tt>*/
#define MXC_R_SPIM_OFFS_GEN_CTRL                            ((uint32_t)0x00000008UL)    /**< Offset from SPIM[n] Base Address: <tt>\b 0x0008</tt>*/
#define MXC_R_SPIM_OFFS_FIFO_CTRL                           ((uint32_t)0x0000000CUL)    /**< Offset from SPIM[n] Base Address: <tt>\b 0x000C</tt>*/
#define MXC_R_SPIM_OFFS_SPCL_CTRL                           ((uint32_t)0x00000010UL)    /**< Offset from SPIM[n] Base Address: <tt>\b 0x0010</tt>*/
#define MXC_R_SPIM_OFFS_INTFL                               ((uint32_t)0x00000014UL)    /**< Offset from SPIM[n] Base Address: <tt>\b 0x0014</tt>*/
#define MXC_R_SPIM_OFFS_INTEN                               ((uint32_t)0x00000018UL)    /**< Offset from SPIM[n] Base Address: <tt>\b 0x0018</tt>*/
#define MXC_R_SPIM_OFFS_SIMPLE_HEADERS                      ((uint32_t)0x0000001CUL)    /**< Offset from SPIM[n] Base Address: <tt>\b 0x001C</tt>*/
/**@} end of group SPIM_Register_Offsets*/
/**
 * @ingroup    spim_registers
 * @defgroup   SPIM_FIFO_Offsets FIFO Offsets
 * @brief      SPI Master FIFO Offsets from the SPIM[n] Base FIFO Address, where  \c n \c = SPIM Instance Number. 
 * @{
 */
#define MXC_R_SPIM_FIFO_OFFS_TRANS                          ((uint32_t)0x00000000UL)    /**< Offset from SPIM[n] Base FIFO Address: <tt>\b 0x0000</tt>*/
#define MXC_R_SPIM_FIFO_OFFS_RSLTS                          ((uint32_t)0x00000800UL)    /**< Offset from SPIM[n] Base FIFO Address: <tt>\b 0x0800</tt>*/
/**@} end of group SPIM_FIFO_Offsets*/

/*
   Field positions and masks for module SPIM.
*/
/**
 * @ingroup  spim_registers
 * @defgroup SPIM_MSTR_CFG_Register SPIM_MSTR_CFG
 * @brief    Field Positions and Bit Masks for the SPIM_MSTR_CFG register
 * @{
 */
#define MXC_F_SPIM_MSTR_CFG_SLAVE_SEL_POS                   0                                                                           /**< SLAVE_SEL Position */
#define MXC_F_SPIM_MSTR_CFG_SLAVE_SEL                       ((uint32_t)(0x00000007UL << MXC_F_SPIM_MSTR_CFG_SLAVE_SEL_POS))             /**< SLAVE_SEL Mask */                           
#define MXC_F_SPIM_MSTR_CFG_THREE_WIRE_MODE_POS             3                                                                           /**< THREE_WIRE_MODE Position */
#define MXC_F_SPIM_MSTR_CFG_THREE_WIRE_MODE                 ((uint32_t)(0x00000001UL << MXC_F_SPIM_MSTR_CFG_THREE_WIRE_MODE_POS))       /**< THREE_WIRE_MODE Mask */                           
#define MXC_F_SPIM_MSTR_CFG_SPI_MODE_POS                    4                                                                           /**< SPI_MODE Position */
#define MXC_F_SPIM_MSTR_CFG_SPI_MODE                        ((uint32_t)(0x00000003UL << MXC_F_SPIM_MSTR_CFG_SPI_MODE_POS))              /**< SPI_MODE Mask */                           
#define MXC_F_SPIM_MSTR_CFG_PAGE_SIZE_POS                   6                                                                           /**< PAGE_SIZE Position */
#define MXC_F_SPIM_MSTR_CFG_PAGE_SIZE                       ((uint32_t)(0x00000003UL << MXC_F_SPIM_MSTR_CFG_PAGE_SIZE_POS))             /**< PAGE_SIZE Mask */                           
#define MXC_F_SPIM_MSTR_CFG_SCK_HI_CLK_POS                  8                                                                           /**< SCK_HI_CLK Position */
#define MXC_F_SPIM_MSTR_CFG_SCK_HI_CLK                      ((uint32_t)(0x0000000FUL << MXC_F_SPIM_MSTR_CFG_SCK_HI_CLK_POS))            /**< SCK_HI_CLK Mask */                           
#define MXC_F_SPIM_MSTR_CFG_SCK_LO_CLK_POS                  12                                                                          /**< SCK_LO_CLK  Position */
#define MXC_F_SPIM_MSTR_CFG_SCK_LO_CLK                      ((uint32_t)(0x0000000FUL << MXC_F_SPIM_MSTR_CFG_SCK_LO_CLK_POS))            /**< SCK_LO_CLK Mask */                           
#define MXC_F_SPIM_MSTR_CFG_ACT_DELAY_POS                   16                                                                          /**< ACT_DELAY Position */
#define MXC_F_SPIM_MSTR_CFG_ACT_DELAY                       ((uint32_t)(0x00000003UL << MXC_F_SPIM_MSTR_CFG_ACT_DELAY_POS))             /**< ACT_DELAY Mask */                           
#define MXC_F_SPIM_MSTR_CFG_INACT_DELAY_POS                 18                                                                          /**< INACT_DELAY Position */
#define MXC_F_SPIM_MSTR_CFG_INACT_DELAY                     ((uint32_t)(0x00000003UL << MXC_F_SPIM_MSTR_CFG_INACT_DELAY_POS))           /**< INACT_DELAY Mask */                           
#define MXC_F_SPIM_MSTR_CFG_SDIO_SAMPLE_POINT_POS           20                                                                          /**< SDIO_SAMPLE_POINT Position */
#define MXC_F_SPIM_MSTR_CFG_SDIO_SAMPLE_POINT               ((uint32_t)(0x0000000FUL << MXC_F_SPIM_MSTR_CFG_SDIO_SAMPLE_POINT_POS))     /**< SDIO_SAMPLE_POINT Mask */                           

#define MXC_V_SPIM_MSTR_CFG_PAGE_SIZE_4B                    ((uint32_t)0x00000000UL)                                                    /**< PAGE_SIZE_4B Field Value */
#define MXC_V_SPIM_MSTR_CFG_PAGE_SIZE_8B                    ((uint32_t)0x00000001UL)                                                    /**< PAGE_SIZE_8B Field Value */
#define MXC_V_SPIM_MSTR_CFG_PAGE_SIZE_16B                   ((uint32_t)0x00000002UL)                                                    /**< PAGE_SIZE_16B Field Value */
#define MXC_V_SPIM_MSTR_CFG_PAGE_SIZE_32B                   ((uint32_t)0x00000003UL)                                                    /**< PAGE_SIZE_32B Field Value */

#define MXC_S_SPIM_MSTR_CFG_PAGE_4B                         (MXC_V_SPIM_MSTR_CFG_PAGE_SIZE_4B  << MXC_F_SPIM_MSTR_CFG_PAGE_SIZE_POS)    /**< PAGE_SIZE_4B Shifted Field Value */
#define MXC_S_SPIM_MSTR_CFG_PAGE_8B                         (MXC_V_SPIM_MSTR_CFG_PAGE_SIZE_8B  << MXC_F_SPIM_MSTR_CFG_PAGE_SIZE_POS)    /**< PAGE_SIZE_8B Shifted Field Value */
#define MXC_S_SPIM_MSTR_CFG_PAGE_16B                        (MXC_V_SPIM_MSTR_CFG_PAGE_SIZE_16B << MXC_F_SPIM_MSTR_CFG_PAGE_SIZE_POS)    /**< PAGE_SIZE_16B Shifted Field Value */
#define MXC_S_SPIM_MSTR_CFG_PAGE_32B                        (MXC_V_SPIM_MSTR_CFG_PAGE_SIZE_32B << MXC_F_SPIM_MSTR_CFG_PAGE_SIZE_POS)    /**< PAGE_SIZE_32B Shifted Field Value */
/**@} end of group SPIM_MSTR_CFG*/
/**
 * @ingroup  spim_registers
 * @defgroup SPIM_SS_SR_POLARITY_Register SPIM_SS_SR_POLARITY
 * @brief    Field Positions and Bit Masks for the SPIM_SS_SR_POLARITY register
 * @{
 */
#define MXC_F_SPIM_SS_SR_POLARITY_SS_POLARITY_POS           0                                                                           /**< SS_POLARITY Position */
#define MXC_F_SPIM_SS_SR_POLARITY_SS_POLARITY               ((uint32_t)(0x000000FFUL << MXC_F_SPIM_SS_SR_POLARITY_SS_POLARITY_POS))     /**< SS_POLARITY Mask */                           
#define MXC_F_SPIM_SS_SR_POLARITY_FC_POLARITY_POS           8                                                                           /**< FC_POLARITY Position */
#define MXC_F_SPIM_SS_SR_POLARITY_FC_POLARITY               ((uint32_t)(0x000000FFUL << MXC_F_SPIM_SS_SR_POLARITY_FC_POLARITY_POS))     /**< FC_POLARITY Mask */                           
/**@} end of group SPIM_SS_SR_POLARITY*/
/**
 * @ingroup  spim_registers
 * @defgroup SPIM_GEN_CTRL_Register SPIM_GEN_CTRL
 * @brief    Field Positions and Bit Masks for the SPIM_GEN_CTRL register
 * @{
 */
#define MXC_F_SPIM_GEN_CTRL_SPI_MSTR_EN_POS                 0                                                                           /**< SPI_MSTR_EN Position */
#define MXC_F_SPIM_GEN_CTRL_SPI_MSTR_EN                     ((uint32_t)(0x00000001UL << MXC_F_SPIM_GEN_CTRL_SPI_MSTR_EN_POS))           /**< SPI_MSTR_EN Mask */                           
#define MXC_F_SPIM_GEN_CTRL_TX_FIFO_EN_POS                  1                                                                           /**< TX_FIFO_EN Position */
#define MXC_F_SPIM_GEN_CTRL_TX_FIFO_EN                      ((uint32_t)(0x00000001UL << MXC_F_SPIM_GEN_CTRL_TX_FIFO_EN_POS))            /**< TX_FIFO_EN Mask */                           
#define MXC_F_SPIM_GEN_CTRL_RX_FIFO_EN_POS                  2                                                                           /**< RX_FIFO_EN Position */
#define MXC_F_SPIM_GEN_CTRL_RX_FIFO_EN                      ((uint32_t)(0x00000001UL << MXC_F_SPIM_GEN_CTRL_RX_FIFO_EN_POS))            /**< RX_FIFO_EN Mask */                           
#define MXC_F_SPIM_GEN_CTRL_BIT_BANG_MODE_POS               3                                                                           /**< BIT_BANG_MODE Position */
#define MXC_F_SPIM_GEN_CTRL_BIT_BANG_MODE                   ((uint32_t)(0x00000001UL << MXC_F_SPIM_GEN_CTRL_BIT_BANG_MODE_POS))         /**< BIT_BANG_MODE Mask */                           
#define MXC_F_SPIM_GEN_CTRL_BB_SS_IN_OUT_POS                4                                                                           /**< BB_SS_IN_OUT Position */
#define MXC_F_SPIM_GEN_CTRL_BB_SS_IN_OUT                    ((uint32_t)(0x00000001UL << MXC_F_SPIM_GEN_CTRL_BB_SS_IN_OUT_POS))          /**< BB_SS_IN_OUT Mask */                           
#define MXC_F_SPIM_GEN_CTRL_BB_SR_IN_POS                    5                                                                           /**< BB_SR_IN Position */
#define MXC_F_SPIM_GEN_CTRL_BB_SR_IN                        ((uint32_t)(0x00000001UL << MXC_F_SPIM_GEN_CTRL_BB_SR_IN_POS))              /**< BB_SR_IN Mask */                           
#define MXC_F_SPIM_GEN_CTRL_BB_SCK_IN_OUT_POS               6                                                                           /**< BB_SCK_IN_OUT Position */
#define MXC_F_SPIM_GEN_CTRL_BB_SCK_IN_OUT                   ((uint32_t)(0x00000001UL << MXC_F_SPIM_GEN_CTRL_BB_SCK_IN_OUT_POS))         /**< BB_SCK_IN_OUT Mask */                           
#define MXC_F_SPIM_GEN_CTRL_BB_SDIO_IN_POS                  8                                                                           /**< BB_SDIO_IN osition */
#define MXC_F_SPIM_GEN_CTRL_BB_SDIO_IN                      ((uint32_t)(0x0000000FUL << MXC_F_SPIM_GEN_CTRL_BB_SDIO_IN_POS))            /**< BB_SDIO_IN Mask */                           
#define MXC_F_SPIM_GEN_CTRL_BB_SDIO_OUT_POS                 12                                                                          /**< BB_SDIO_OUT Position */
#define MXC_F_SPIM_GEN_CTRL_BB_SDIO_OUT                     ((uint32_t)(0x0000000FUL << MXC_F_SPIM_GEN_CTRL_BB_SDIO_OUT_POS))           /**< BB_SDIO_OUT Mask */                           
#define MXC_F_SPIM_GEN_CTRL_BB_SDIO_DR_EN_POS               16                                                                          /**< BB_SDIO_DR_EN Position */
#define MXC_F_SPIM_GEN_CTRL_BB_SDIO_DR_EN                   ((uint32_t)(0x0000000FUL << MXC_F_SPIM_GEN_CTRL_BB_SDIO_DR_EN_POS))         /**< BB_SDIO_DR_EN Mask */                           
#define MXC_F_SPIM_GEN_CTRL_SIMPLE_MODE_POS                 20                                                                          /**< SIMPLE_MODE Position */
#define MXC_F_SPIM_GEN_CTRL_SIMPLE_MODE                     ((uint32_t)(0x00000001UL << MXC_F_SPIM_GEN_CTRL_SIMPLE_MODE_POS))           /**< SIMPLE_MODE Mask */                           
#define MXC_F_SPIM_GEN_CTRL_START_RX_ONLY_POS               21                                                                          /**< START_RX_ONLY Position */
#define MXC_F_SPIM_GEN_CTRL_START_RX_ONLY                   ((uint32_t)(0x00000001UL << MXC_F_SPIM_GEN_CTRL_START_RX_ONLY_POS))         /**< START_RX_ONLY Mask */                           
#define MXC_F_SPIM_GEN_CTRL_DEASSERT_ACT_SS_POS             22                                                                          /**< DEASSERT_ACT_SS Position */
#define MXC_F_SPIM_GEN_CTRL_DEASSERT_ACT_SS                 ((uint32_t)(0x00000001UL << MXC_F_SPIM_GEN_CTRL_DEASSERT_ACT_SS_POS))       /**< DEASSERT_ACT_SS Mask */                           
#define MXC_F_SPIM_GEN_CTRL_ENABLE_SCK_FB_MODE_POS          24                                                                          /**< ENABLE_SCK_FB_MOD Position */
#define MXC_F_SPIM_GEN_CTRL_ENABLE_SCK_FB_MODE              ((uint32_t)(0x00000001UL << MXC_F_SPIM_GEN_CTRL_ENABLE_SCK_FB_MODE_POS))    /**< ENABLE_SCK_FB_MOD Mask */                           
#define MXC_F_SPIM_GEN_CTRL_INVERT_SCK_FB_CLK_POS           25                                                                          /**< INVERT_SCK_FB_CLK Position */
#define MXC_F_SPIM_GEN_CTRL_INVERT_SCK_FB_CLK               ((uint32_t)(0x00000001UL << MXC_F_SPIM_GEN_CTRL_INVERT_SCK_FB_CLK_POS))     /**< INVERT_SCK_FB_CLK Mask */                           
/**@} end of group SPIM_GEN_CTRL*/
/**
 * @ingroup  spim_registers
 * @defgroup SPIM_FIFO_CTRL_Register SPIM_FIFO_CTRL
 * @brief    Field Positions and Bit Masks for the SPIM_FIFO_CTRL register
 * @{
 */    
#define MXC_F_SPIM_FIFO_CTRL_TX_FIFO_AE_LVL_POS             0                                                                           /**< TX_FIFO_AE_LVL Position */
#define MXC_F_SPIM_FIFO_CTRL_TX_FIFO_AE_LVL                 ((uint32_t)(0x0000000FUL << MXC_F_SPIM_FIFO_CTRL_TX_FIFO_AE_LVL_POS))       /**< TX_FIFO_AE_LVL Mask */                           
#define MXC_F_SPIM_FIFO_CTRL_TX_FIFO_USED_POS               8                                                                           /**< TX_FIFO_USED Position */
#define MXC_F_SPIM_FIFO_CTRL_TX_FIFO_USED                   ((uint32_t)(0x0000001FUL << MXC_F_SPIM_FIFO_CTRL_TX_FIFO_USED_POS))         /**< TX_FIFO_USED Mask */                           
#define MXC_F_SPIM_FIFO_CTRL_RX_FIFO_AF_LVL_POS             16                                                                          /**< RX_FIFO_AF_LVL Position */
#define MXC_F_SPIM_FIFO_CTRL_RX_FIFO_AF_LVL                 ((uint32_t)(0x0000001FUL << MXC_F_SPIM_FIFO_CTRL_RX_FIFO_AF_LVL_POS))       /**< RX_FIFO_AF_LVL Mask */                           
#define MXC_F_SPIM_FIFO_CTRL_RX_FIFO_USED_POS               24                                                                          /**< RX_FIFO_USED Position */
#define MXC_F_SPIM_FIFO_CTRL_RX_FIFO_USED                   ((uint32_t)(0x0000003FUL << MXC_F_SPIM_FIFO_CTRL_RX_FIFO_USED_POS))         /**< RX_FIFO_USED Mask */                           
/**@} end of group SPIM_FIFO_CTRL*/
/**
 * @ingroup  spim_registers
 * @defgroup SPIM_SPCL_CTRL_Register SPIM_SPCL_CTRL
 * @brief    Field Positions and Bit Masks for the SPIM_SPCL_CTRL register
 * @{
 */    
#define MXC_F_SPIM_SPCL_CTRL_SS_SAMPLE_MODE_POS             0                                                                           /**< SS_SAMPLE_MODE Position */
#define MXC_F_SPIM_SPCL_CTRL_SS_SAMPLE_MODE                 ((uint32_t)(0x00000001UL << MXC_F_SPIM_SPCL_CTRL_SS_SAMPLE_MODE_POS))       /**< SS_SAMPLE_MODE Mask */                           
#define MXC_F_SPIM_SPCL_CTRL_MISO_FC_EN_POS                 1                                                                           /**< MISO_FC_EN Position */
#define MXC_F_SPIM_SPCL_CTRL_MISO_FC_EN                     ((uint32_t)(0x00000001UL << MXC_F_SPIM_SPCL_CTRL_MISO_FC_EN_POS))           /**< MISO_FC_EN Mask */                           
#define MXC_F_SPIM_SPCL_CTRL_SS_SA_SDIO_OUT_POS             4                                                                           /**< SS_SA_SDIO_OUT Position */
#define MXC_F_SPIM_SPCL_CTRL_SS_SA_SDIO_OUT                 ((uint32_t)(0x0000000FUL << MXC_F_SPIM_SPCL_CTRL_SS_SA_SDIO_OUT_POS))       /**< SS_SA_SDIO_OUT Mask */                           
#define MXC_F_SPIM_SPCL_CTRL_SS_SA_SDIO_DR_EN_POS           8                                                                           /**< SS_SA_SDIO_DR_EN Position */
#define MXC_F_SPIM_SPCL_CTRL_SS_SA_SDIO_DR_EN               ((uint32_t)(0x0000000FUL << MXC_F_SPIM_SPCL_CTRL_SS_SA_SDIO_DR_EN_POS))     /**< SS_SA_SDIO_DR_EN Mask */                           

#if (MXC_SPIM_REV == 0)
#define MXC_F_SPIM_SPCL_CTRL_SPECIAL_MODE_3_EN_POS          16                                                                          /**< SPECIAL_MODE_3_EN Position */
#define MXC_F_SPIM_SPCL_CTRL_SPECIAL_MODE_3_EN              ((uint32_t)(0x00000001UL << MXC_F_SPIM_SPCL_CTRL_SPECIAL_MODE_3_EN_POS))    /**< SPECIAL_MODE_3_EN Mask */                           
#else
#define MXC_F_SPIM_SPCL_CTRL_RX_FIFO_MARGIN_POS             12                                                                          /**< RX_FIFO_MARGIN Position */
#define MXC_F_SPIM_SPCL_CTRL_RX_FIFO_MARGIN                 ((uint32_t)(0x00000007UL << MXC_F_SPIM_SPCL_CTRL_RX_FIFO_MARGIN_POS))       /**< RX_FIFO_MARGIN Mask */                           
#define MXC_F_SPIM_SPCL_CTRL_SCK_FB_DELAY_POS               16                                                                          /**< SCK_FB_DELAY Position */
#define MXC_F_SPIM_SPCL_CTRL_SCK_FB_DELAY                   ((uint32_t)(0x0000000FUL << MXC_F_SPIM_SPCL_CTRL_SCK_FB_DELAY_POS))         /**< SCK_FB_DELAY Mask */                           
#define MXC_F_SPIM_SPCL_CTRL_SPARE_RESERVED_POS             20                                                                          /**< SPARE_RESERVED Position */
#define MXC_F_SPIM_SPCL_CTRL_SPARE_RESERVED                 ((uint32_t)(0x00000FFFUL << MXC_F_SPIM_SPCL_CTRL_SPARE_RESERVED_POS))       /**< SPARE_RESERVED Mask */                           
#endif
/**@} end of group SPIM_SPCL_CTRL*/
/**
 * @ingroup  spim_registers
 * @defgroup SPIM_INTFL_Register SPIM_INTFL
 * @brief    Field Positions and Bit Masks for the SPIM_INTFL register
 * @{
 */ 
#define MXC_F_SPIM_INTFL_TX_STALLED_POS                     0                                                                   /**< TX_STALLED Position */
#define MXC_F_SPIM_INTFL_TX_STALLED                         ((uint32_t)(0x00000001UL << MXC_F_SPIM_INTFL_TX_STALLED_POS))       /**< TX_STALLED Mask */                           
#define MXC_F_SPIM_INTFL_RX_STALLED_POS                     1                                                                   /**< RX_STALLED Position */
#define MXC_F_SPIM_INTFL_RX_STALLED                         ((uint32_t)(0x00000001UL << MXC_F_SPIM_INTFL_RX_STALLED_POS))       /**< RX_STALLED Mask */                           
#define MXC_F_SPIM_INTFL_TX_READY_POS                       2                                                                   /**< TX_READY Position */   
#define MXC_F_SPIM_INTFL_TX_READY                           ((uint32_t)(0x00000001UL << MXC_F_SPIM_INTFL_TX_READY_POS))         /**< TX_READY Mask */                           
#define MXC_F_SPIM_INTFL_RX_DONE_POS                        3                                                                   /**< RX_DONE Position */   
#define MXC_F_SPIM_INTFL_RX_DONE                            ((uint32_t)(0x00000001UL << MXC_F_SPIM_INTFL_RX_DONE_POS))          /**< RX_DONE Mask */                           
#define MXC_F_SPIM_INTFL_TX_FIFO_AE_POS                     4                                                                   /**< TX_FIFO_AE Position */
#define MXC_F_SPIM_INTFL_TX_FIFO_AE                         ((uint32_t)(0x00000001UL << MXC_F_SPIM_INTFL_TX_FIFO_AE_POS))       /**< TX_FIFO_AE Mask */                           
#define MXC_F_SPIM_INTFL_RX_FIFO_AF_POS                     5                                                                   /**< RX_FIFO_AF Position */
#define MXC_F_SPIM_INTFL_RX_FIFO_AF                         ((uint32_t)(0x00000001UL << MXC_F_SPIM_INTFL_RX_FIFO_AF_POS))       /**< RX_FIFO_AF Mask */                           
/**@} end of group SPIM_INTFL*/
/**
 * @ingroup  spim_registers
 * @defgroup SPIM_INTEN_Register SPIM_INTEN
 * @brief    Field Positions and Bit Masks for the SPIM_INTEN register
 * @{
 */ 
#define MXC_F_SPIM_INTEN_TX_STALLED_POS                     0                                                                /**< TX_STALLED Position */
#define MXC_F_SPIM_INTEN_TX_STALLED                         ((uint32_t)(0x00000001UL << MXC_F_SPIM_INTEN_TX_STALLED_POS))    /**< TX_STALLED Mask */                           
#define MXC_F_SPIM_INTEN_RX_STALLED_POS                     1                                                                /**< RX_STALLED Position */
#define MXC_F_SPIM_INTEN_RX_STALLED                         ((uint32_t)(0x00000001UL << MXC_F_SPIM_INTEN_RX_STALLED_POS))    /**< RX_STALLED Mask */                           
#define MXC_F_SPIM_INTEN_TX_READY_POS                       2                                                                /**< TX_READY Position */   
#define MXC_F_SPIM_INTEN_TX_READY                           ((uint32_t)(0x00000001UL << MXC_F_SPIM_INTEN_TX_READY_POS))      /**< TX_READY Mask */                           
#define MXC_F_SPIM_INTEN_RX_DONE_POS                        3                                                                /**< RX_DONE Position */   
#define MXC_F_SPIM_INTEN_RX_DONE                            ((uint32_t)(0x00000001UL << MXC_F_SPIM_INTEN_RX_DONE_POS))       /**< RX_DONE Mask */                           
#define MXC_F_SPIM_INTEN_TX_FIFO_AE_POS                     4                                                                /**< TX_FIFO_AE Position */
#define MXC_F_SPIM_INTEN_TX_FIFO_AE                         ((uint32_t)(0x00000001UL << MXC_F_SPIM_INTEN_TX_FIFO_AE_POS))    /**< TX_FIFO_AE Mask */                           
#define MXC_F_SPIM_INTEN_RX_FIFO_AF_POS                     5                                                                /**< RX_FIFO_AF Position */
#define MXC_F_SPIM_INTEN_RX_FIFO_AF                         ((uint32_t)(0x00000001UL << MXC_F_SPIM_INTEN_RX_FIFO_AF_POS))    /**< RX_FIFO_AF Mask */                           
/**@} end of group SPIM_INTEN*/
/**
 * @ingroup  spim_registers
 * @defgroup SPIM_SIMPLE_HEADERS_Register SPIM_SIMPLE_HEADERS
 * @brief    Field Positions and Bit Masks for the SPIM_SIMPLE_HEADERS register
 * @{
 */ 
#define MXC_F_SPIM_SIMPLE_HEADERS_TX_BIDIR_HEADER_POS       0                                                                           /**< TX_BIDIR_HEADER Position */
#define MXC_F_SPIM_SIMPLE_HEADERS_TX_BIDIR_HEADER           ((uint32_t)(0x00003FFFUL << MXC_F_SPIM_SIMPLE_HEADERS_TX_BIDIR_HEADER_POS)) /**< TX_BIDIR_HEADER Mask */                           
#define MXC_F_SPIM_SIMPLE_HEADERS_RX_ONLY_HEADER_POS        16                                                                          /**< RX_ONLY_HEADER Position */
#define MXC_F_SPIM_SIMPLE_HEADERS_RX_ONLY_HEADER            ((uint32_t)(0x00003FFFUL << MXC_F_SPIM_SIMPLE_HEADERS_RX_ONLY_HEADER_POS))  /**< RX_ONLY_HEADER Mask */                           
/**@} end of group SPIM_SIMPLE_HEADERS*/


#ifdef __cplusplus
}
#endif

#endif   /* _MXC_SPIM_REGS_H_ */

