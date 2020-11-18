/**
 * @file
 * @brief   Registers, Fields, Field Positions, Masks and Values for the SPIX Peripheral Module.
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
 * $Date: 2017-02-16 08:53:14 -0600 (Thu, 16 Feb 2017) $
 * $Revision: 26456 $
 *
 *************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _MXC_SPIX_REGS_H_
#define _MXC_SPIX_REGS_H_

/* **** Includes **** */
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/// @cond
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
/// @endcond

/* **** Definitions **** */

/**
 * @defgroup    spix_registers Registers
 * @ingroup     spix
 * @brief       Registers, Bit Masks and Bit Positions for the SPIX Peripheral Module.
 * @{
 */

/**
 * Structure type to access the SPIX Registers.
 */
  typedef struct {
    __IO uint32_t master_cfg;                           /**<  SPIX_MASTER_CFG Register.                                               */
    __IO uint32_t fetch_ctrl;                           /**<  SPIX_FETCH_CTRL Register.                                               */
    __IO uint32_t mode_ctrl;                            /**<  SPIX_MODE_CTRL Register.                                                */
    __IO uint32_t mode_data;                            /**<  SPIX_MODE_DATA Register.                                                */
    __IO uint32_t sck_fb_ctrl;                          /**<  SPIX_SCK_FB_CTRL Register.                                              */
} mxc_spix_regs_t;

/**
 * @defgroup   SPIX_Register_Offsets Register Offsets
 * @ingroup spix_registers
 * @brief      SPIX Peripheral Register Offsets from the SPIX Base Peripheral Address, #MXC_BASE_SPIX. 
 * @{
 */
#define MXC_R_SPIX_OFFS_MASTER_CFG                          ((uint32_t)0x00000000UL) /**< Offset from #MXC_BASE_SPIX: <b><tt>0x000</tt></b>  */
#define MXC_R_SPIX_OFFS_FETCH_CTRL                          ((uint32_t)0x00000004UL) /**< Offset from #MXC_BASE_SPIX: <b><tt>0x004</tt></b>  */
#define MXC_R_SPIX_OFFS_MODE_CTRL                           ((uint32_t)0x00000008UL) /**< Offset from #MXC_BASE_SPIX: <b><tt>0x008</tt></b>  */
#define MXC_R_SPIX_OFFS_MODE_DATA                           ((uint32_t)0x0000000CUL) /**< Offset from #MXC_BASE_SPIX: <b><tt>0x00C</tt></b>  */
#define MXC_R_SPIX_OFFS_SCK_FB_CTRL                         ((uint32_t)0x00000010UL) /**< Offset from #MXC_BASE_SPIX: <b><tt>0x010</tt></b>  */
/**@} end of SPIX_Register_Offsets */

/**
 * @defgroup SPIX_Master_Cfg_Register SPIX_MASTER_CFG Register Fields
 * @ingroup spix_registers
 * @brief Register Fields and Shifted Field Masks for the SPIX_MASTER_CFG Register.
 * @{
 */
#define MXC_F_SPIX_MASTER_CFG_SPI_MODE_POS                  0                                                                         /**< SPI_MODE Field Position                                  */
#define MXC_F_SPIX_MASTER_CFG_SPI_MODE                      ((uint32_t)(0x00000003UL << MXC_F_SPIX_MASTER_CFG_SPI_MODE_POS))          /**< SPI_MODE Shifted Field Mask                                      */
#define MXC_F_SPIX_MASTER_CFG_SS_ACT_LO_POS                 2                                                                         /**< SS_ACT_LO Field Position                                 */
#define MXC_F_SPIX_MASTER_CFG_SS_ACT_LO                     ((uint32_t)(0x00000001UL << MXC_F_SPIX_MASTER_CFG_SS_ACT_LO_POS))         /**< SS_ACT_LO Shifted Field Mask                                     */
#define MXC_F_SPIX_MASTER_CFG_ALT_TIMING_EN_POS             3                                                                         /**< ALT_TIMING_EN Field Position                             */
#define MXC_F_SPIX_MASTER_CFG_ALT_TIMING_EN                 ((uint32_t)(0x00000001UL << MXC_F_SPIX_MASTER_CFG_ALT_TIMING_EN_POS))     /**< ALT_TIMING_EN Shifted Field Mask                                 */
#define MXC_F_SPIX_MASTER_CFG_SLAVE_SEL_POS                 4                                                                         /**< SLAVE_SEL Field Position                                 */
#define MXC_F_SPIX_MASTER_CFG_SLAVE_SEL                     ((uint32_t)(0x00000007UL << MXC_F_SPIX_MASTER_CFG_SLAVE_SEL_POS))         /**< SLAVE_SEL Shifted Field Mask                                     */
#define MXC_F_SPIX_MASTER_CFG_SCK_LO_CLK_POS                8                                                                         /**< SCK_LO_CLK Field Position                                */
#define MXC_F_SPIX_MASTER_CFG_SCK_LO_CLK                    ((uint32_t)(0x0000000FUL << MXC_F_SPIX_MASTER_CFG_SCK_LO_CLK_POS))        /**< SCK_LO_CLK Shifted Field Mask                                    */
#define MXC_F_SPIX_MASTER_CFG_SCK_HI_CLK_POS                12                                                                        /**< SCK_HI_CLK Field Position                                */
#define MXC_F_SPIX_MASTER_CFG_SCK_HI_CLK                    ((uint32_t)(0x0000000FUL << MXC_F_SPIX_MASTER_CFG_SCK_HI_CLK_POS))        /**< SCK_HI_CLK Shifted Field Mask                                    */
#define MXC_F_SPIX_MASTER_CFG_ACT_DELAY_POS                 16                                                                        /**< ACT_DELAY Field Position                                 */
#define MXC_F_SPIX_MASTER_CFG_ACT_DELAY                     ((uint32_t)(0x00000003UL << MXC_F_SPIX_MASTER_CFG_ACT_DELAY_POS))         /**< ACT_DELAY Shifted Field Mask                                     */
#define MXC_F_SPIX_MASTER_CFG_INACT_DELAY_POS               18                                                                        /**< INACT_DELAY Field Position                               */
#define MXC_F_SPIX_MASTER_CFG_INACT_DELAY                   ((uint32_t)(0x00000003UL << MXC_F_SPIX_MASTER_CFG_INACT_DELAY_POS))       /**< INACT_DELAY Shifted Field Mask                                   */
#define MXC_F_SPIX_MASTER_CFG_ALT_SCK_LO_CLK_POS            20                                                                        /**< ALT_SCK_LO_CLK Field Position                            */
#define MXC_F_SPIX_MASTER_CFG_ALT_SCK_LO_CLK                ((uint32_t)(0x0000000FUL << MXC_F_SPIX_MASTER_CFG_ALT_SCK_LO_CLK_POS))    /**< ALT_SCK_LO_CLK Shifted Field Mask                                */
#define MXC_F_SPIX_MASTER_CFG_ALT_SCK_HI_CLK_POS            24                                                                        /**< ALT_SCK_HI_CLK Field Position                            */
#define MXC_F_SPIX_MASTER_CFG_ALT_SCK_HI_CLK                ((uint32_t)(0x0000000FUL << MXC_F_SPIX_MASTER_CFG_ALT_SCK_HI_CLK_POS))    /**< ALT_SCK_HI_CLK Shifted Field Mask                                */
#define MXC_F_SPIX_MASTER_CFG_SDIO_SAMPLE_POINT_POS         28                                                                        /**< SDIO_SAMPLE_POINT Field Position                         */
#define MXC_F_SPIX_MASTER_CFG_SDIO_SAMPLE_POINT             ((uint32_t)(0x0000000FUL << MXC_F_SPIX_MASTER_CFG_SDIO_SAMPLE_POINT_POS)) /**< SDIO_SAMPLE_POINT Shifted Field Mask                             */
/**@}*/
/**
 * @defgroup SPIX_Fetch_Ctrl_Register SPIX_FETCH_CTRL Register Fields
 * @ingroup spix_registers
 * @brief Register Fields and Shifted Masks for the SPIX_FETCH_CTRL Register.
 * @{
 */
#define MXC_F_SPIX_FETCH_CTRL_CMD_VALUE_POS                 0                                                                         /**< CMD_VALUE Field Position  */
#define MXC_F_SPIX_FETCH_CTRL_CMD_VALUE                     ((uint32_t)(0x000000FFUL << MXC_F_SPIX_FETCH_CTRL_CMD_VALUE_POS))         /**< CMD_VALUE Shifted Field Mask      */
#define MXC_F_SPIX_FETCH_CTRL_CMD_WIDTH_POS                 8                                                                         /**< CMD_WIDTH Field Position  */
#define MXC_F_SPIX_FETCH_CTRL_CMD_WIDTH                     ((uint32_t)(0x00000003UL << MXC_F_SPIX_FETCH_CTRL_CMD_WIDTH_POS))         /**< CMD_WIDTH Shifted Field Mask      */
#define MXC_F_SPIX_FETCH_CTRL_ADDR_WIDTH_POS                10                                                                        /**< ADDR_WIDTH Field Position  */
#define MXC_F_SPIX_FETCH_CTRL_ADDR_WIDTH                    ((uint32_t)(0x00000003UL << MXC_F_SPIX_FETCH_CTRL_ADDR_WIDTH_POS))        /**< ADDR_WIDTH Shifted Field Mask      */
#define MXC_F_SPIX_FETCH_CTRL_DATA_WIDTH_POS                12                                                                        /**< DATA_WIDTH Field Position  */
#define MXC_F_SPIX_FETCH_CTRL_DATA_WIDTH                    ((uint32_t)(0x00000003UL << MXC_F_SPIX_FETCH_CTRL_DATA_WIDTH_POS))        /**< DATA_WIDTH Shifted Field Mask      */
#define MXC_F_SPIX_FETCH_CTRL_FOUR_BYTE_ADDR_POS            16                                                                        /**< FOUR_BYTE_ADDR Field Position  */
#define MXC_F_SPIX_FETCH_CTRL_FOUR_BYTE_ADDR                ((uint32_t)(0x00000001UL << MXC_F_SPIX_FETCH_CTRL_FOUR_BYTE_ADDR_POS))    /**< FOUR_BYTE_ADDRField Mask      */
/**@}*/
/**
 * @defgroup SPIX_Mode_Ctrl_Register SPIX_MODE_CTRL Register Fields
 * @ingroup spix_registers
 * @brief Register Fields and Shifted Masks for the SPIX_MODE_CTRL Register.
 * @{
 */
#define MXC_F_SPIX_MODE_CTRL_MODE_CLOCKS_POS                0                                                                         /**< MODE_CLOCKS Field Position  */
#define MXC_F_SPIX_MODE_CTRL_MODE_CLOCKS                    ((uint32_t)(0x0000000FUL << MXC_F_SPIX_MODE_CTRL_MODE_CLOCKS_POS))        /**< MODE_CLOCKS Shifted Field Mask      */
#define MXC_F_SPIX_MODE_CTRL_NO_CMD_MODE_POS                8                                                                         /**< NO_CMD_MODE Field Position  */
#define MXC_F_SPIX_MODE_CTRL_NO_CMD_MODE                    ((uint32_t)(0x00000001UL << MXC_F_SPIX_MODE_CTRL_NO_CMD_MODE_POS))        /**< NO_CMD_MODE Shifted Field Mask      */
/**@}*/
/**
 * @defgroup SPIX_Mode_Data_Register SPIX_MODE_DATA Register Fields
 * @ingroup spix_registers
 * @brief Register Fields and Shifted Masks for the SPIX_MODE_DATA Register.
 * @{
 */
#define MXC_F_SPIX_MODE_DATA_MODE_DATA_BITS_POS             0                                                                         /**< MODE_DATA_BITS Field Position  */
#define MXC_F_SPIX_MODE_DATA_MODE_DATA_BITS                 ((uint32_t)(0x0000FFFFUL << MXC_F_SPIX_MODE_DATA_MODE_DATA_BITS_POS))     /**< MODE_DATA_BITS Shifted Field Mask      */
#define MXC_F_SPIX_MODE_DATA_MODE_DATA_OE_POS               16                                                                         /**< MODE_DATA_OE Field Position  */
#define MXC_F_SPIX_MODE_DATA_MODE_DATA_OE                   ((uint32_t)(0x0000FFFFUL << MXC_F_SPIX_MODE_DATA_MODE_DATA_OE_POS))        /**< MODE_DATA_OE Shifted Field Mask      */
/**@}*/
/**
 * @defgroup SPIX_SCK_Fb_Ctrl_Register SPIX_SCK_FB_CTRL Register Fields
 * @ingroup spix_registers
 * @brief Register Fields and Shifted Masks for the SPIX_SCK_FB_CTRL Register.
 * @{
 */
#define MXC_F_SPIX_SCK_FB_CTRL_ENABLE_SCK_FB_MODE_POS       0                                                                             /**< Field Position  */
#define MXC_F_SPIX_SCK_FB_CTRL_ENABLE_SCK_FB_MODE           ((uint32_t)(0x00000001UL << MXC_F_SPIX_SCK_FB_CTRL_ENABLE_SCK_FB_MODE_POS))   /**< Field Mask      */
#define MXC_F_SPIX_SCK_FB_CTRL_INVERT_SCK_FB_CLK_POS        1                                                                             /**< Field Position  */
#define MXC_F_SPIX_SCK_FB_CTRL_INVERT_SCK_FB_CLK            ((uint32_t)(0x00000001UL << MXC_F_SPIX_SCK_FB_CTRL_INVERT_SCK_FB_CLK_POS))    /**< Field Mask      */

#if(MXC_SPIX_REV == 0)
#define MXC_F_SPIX_SCK_FB_CTRL_IGNORE_CLKS_POS              4                                                                             /**< Field Position  */
#define MXC_F_SPIX_SCK_FB_CTRL_IGNORE_CLKS                  ((uint32_t)(0x0000003FUL << MXC_F_SPIX_SCK_FB_CTRL_IGNORE_CLKS_POS))          /**< Field Mask      */
#define MXC_F_SPIX_SCK_FB_CTRL_IGNORE_CLKS_NO_CMD_POS       12                                                                            /**< Field Position  */
#define MXC_F_SPIX_SCK_FB_CTRL_IGNORE_CLKS_NO_CMD           ((uint32_t)(0x0000003FUL << MXC_F_SPIX_SCK_FB_CTRL_IGNORE_CLKS_NO_CMD_POS))   /**< Field Mask      */
#endif
/**@}*/


/**
 * @defgroup SPIX_Master_Cfg_SCK SCK Sampling Mode Field
 * @ingroup SPIX_Master_Cfg_Register 
 * @brief Field values and shifted field values for setting the SPIX SCK Sampling Mode.
 * @{
 */
#define MXC_V_SPIX_MASTER_CFG_SPI_MODE_SCK_HI_SAMPLE_RISING                     ((uint32_t)(0x00000000UL))  /**< Field value for setting the sampling of the SCK on the rising edge. */
#define MXC_V_SPIX_MASTER_CFG_SPI_MODE_SCK_LO_SAMPLE_FALLING                    ((uint32_t)(0x00000003UL))  /**< Field value for setting the sampling of the SCK on the falling edge. */

#define MXC_S_SPIX_MASTER_CFG_SPI_MODE_SCK_HI_SAMPLE_RISING                     ((uint32_t)(MXC_V_SPIX_MASTER_CFG_SPI_MODE_SCK_HI_SAMPLE_RISING    << MXC_F_SPIX_MASTER_CFG_SPI_MODE_POS)) /**< SCK sampling on rising edge Field Shifted Value. */
#define MXC_S_SPIX_MASTER_CFG_SPI_MODE_SCK_LO_SAMPLE_FALLING                    ((uint32_t)(MXC_V_SPIX_MASTER_CFG_SPI_MODE_SCK_LO_SAMPLE_FALLING   << MXC_F_SPIX_MASTER_CFG_SPI_MODE_POS)) /**< SCK sampling on falling edge Field Shifted Value. */
/**@}*/
/**
 * @defgroup SPIX_Master_Cfg_SS Slave Select Polarity Field
 * @ingroup SPIX_Master_Cfg_Register 
 * @brief Field values and shifted field values for setting the SPIX Slave Select Active High/Low Field.
 * @{
 */
#define MXC_V_SPIX_MASTER_CFG_SS_ACT_LO_ACTIVE_HIGH                             ((uint32_t)(0x00000000UL))  /**< Slave Select Active High Field selection value.         */
#define MXC_V_SPIX_MASTER_CFG_SS_ACT_LO_ACTIVE_LOW                              ((uint32_t)(0x00000001UL))  /**< Slave Select Active Low Field selection value.         */

#define MXC_S_SPIX_MASTER_CFG_SS_ACT_LO_ACTIVE_HIGH                             ((uint32_t)(MXC_V_SPIX_MASTER_CFG_SS_ACT_LO_ACTIVE_HIGH  << MXC_F_SPIX_MASTER_CFG_SS_ACT_LO_POS)) /**< Slave Select Active High Field Shifted Value.  */
#define MXC_S_SPIX_MASTER_CFG_SS_ACT_LO_ACTIVE_LOW                              ((uint32_t)(MXC_V_SPIX_MASTER_CFG_SS_ACT_LO_ACTIVE_LOW   << MXC_F_SPIX_MASTER_CFG_SS_ACT_LO_POS)) /**< Slave Select Active Low Field Shifted Value.  */
/**@}*/
/**
 * @defgroup SPIX_Master_Cfg_Alt Alternate Timing
 * @ingroup SPIX_Master_Cfg_Register 
 * @brief Field values and shifted field values for setting the SPIX Alternate Timing Field. 
 * @{
 */
#define MXC_V_SPIX_MASTER_CFG_ALT_TIMING_EN_DISABLED                            ((uint32_t)(0x00000000UL))  /**< Alternate Timing Disabled (Default) Field selection value.         */
#define MXC_V_SPIX_MASTER_CFG_ALT_TIMING_EN_ENABLED_AS_NEEDED                   ((uint32_t)(0x00000001UL))  /**< Alternate Timing Enabled As Needed Field selection value.         */

#define MXC_S_SPIX_MASTER_CFG_ALT_TIMING_EN_DISABLED                            ((uint32_t)(MXC_V_SPIX_MASTER_CFG_ALT_TIMING_EN_DISABLED            << MXC_F_SPIX_MASTER_CFG_ALT_TIMING_EN_POS))  /**< Alternate Timing Disabled Field Shifted Value.  */
#define MXC_S_SPIX_MASTER_CFG_ALT_TIMING_EN_ENABLED_AS_NEEDED                   ((uint32_t)(MXC_V_SPIX_MASTER_CFG_ALT_TIMING_EN_ENABLED_AS_NEEDED   << MXC_F_SPIX_MASTER_CFG_ALT_TIMING_EN_POS))  /**< Alternate Timing Enabled As Needed Field Shifted Value.  */
/**@}*/
/**
 * @defgroup SPIX_Master_Cfg_Act Active Delay Settings
 * @ingroup SPIX_Master_Cfg_Register 
 * @brief Field values and shifted field values for setting the SPIX Activity Delay, the number of SPIX clocks between slave selection assert and active SPI clocking. 
 * @{
 */
#define MXC_V_SPIX_MASTER_CFG_ACT_DELAY_OFF                                     ((uint32_t)(0x00000000UL))  /**< Activity Delay Off Field selection value.         */
#define MXC_V_SPIX_MASTER_CFG_ACT_DELAY_FOR_2_MOD_CLK                           ((uint32_t)(0x00000001UL))  /**< 2 Mode Clocks Field selection value.         */
#define MXC_V_SPIX_MASTER_CFG_ACT_DELAY_FOR_4_MOD_CLK                           ((uint32_t)(0x00000002UL))  /**< 4 Mode Clocks Field selection value.         */
#define MXC_V_SPIX_MASTER_CFG_ACT_DELAY_FOR_8_MOD_CLK                           ((uint32_t)(0x00000003UL))  /**< 8 Mode Clocks Field selection value.         */

#define MXC_S_SPIX_MASTER_CFG_ACT_DELAY_OFF                                     ((uint32_t)(MXC_V_SPIX_MASTER_CFG_ACT_DELAY_OFF             << MXC_F_SPIX_MASTER_CFG_ACT_DELAY_POS))  /**< Activity Delay Off Field Shifted Value.  */
#define MXC_S_SPIX_MASTER_CFG_ACT_DELAY_FOR_2_MOD_CLK                           ((uint32_t)(MXC_V_SPIX_MASTER_CFG_ACT_DELAY_FOR_2_MOD_CLK   << MXC_F_SPIX_MASTER_CFG_ACT_DELAY_POS))  /**< 2 Mode Clocks Field Shifted Value.  */
#define MXC_S_SPIX_MASTER_CFG_ACT_DELAY_FOR_4_MOD_CLK                           ((uint32_t)(MXC_V_SPIX_MASTER_CFG_ACT_DELAY_FOR_4_MOD_CLK   << MXC_F_SPIX_MASTER_CFG_ACT_DELAY_POS))  /**< 4 Mode Clocks Field Shifted Value.  */
#define MXC_S_SPIX_MASTER_CFG_ACT_DELAY_FOR_8_MOD_CLK                           ((uint32_t)(MXC_V_SPIX_MASTER_CFG_ACT_DELAY_FOR_8_MOD_CLK   << MXC_F_SPIX_MASTER_CFG_ACT_DELAY_POS))  /**< 8 Mode Clocks Field Shifted Value.  */
/**@}*/
/**
 * @defgroup SPIX_Master_Cfg_Inact Inactive Delay Settings 
 * @ingroup SPIX_Master_Cfg_Register 
 * @brief Field values and shifted field values for setting the SPIX Inactivity Delay, the number of SPIX clocks between the active SPI Clock and the Slave Select Deassertion.
 * @{
 */
#define MXC_V_SPIX_MASTER_CFG_INACT_DELAY_OFF                                   ((uint32_t)(0x00000000UL))  /**< Inactivity Delay Off Field selection value.         */
#define MXC_V_SPIX_MASTER_CFG_INACT_DELAY_FOR_2_MOD_CLK                         ((uint32_t)(0x00000001UL))  /**< 2 Mode Clocks Field selection value.         */
#define MXC_V_SPIX_MASTER_CFG_INACT_DELAY_FOR_4_MOD_CLK                         ((uint32_t)(0x00000002UL))  /**< 4 Mode Clocks Field selection value.         */
#define MXC_V_SPIX_MASTER_CFG_INACT_DELAY_FOR_8_MOD_CLK                         ((uint32_t)(0x00000003UL))  /**< 8 Mode Clocks Field selection value.         */

#define MXC_S_SPIX_MASTER_CFG_INACT_DELAY_OFF                                   ((uint32_t)(MXC_V_SPIX_MASTER_CFG_INACT_DELAY_OFF             << MXC_F_SPIX_MASTER_CFG_INACT_DELAY_POS))  /**< Inactivity Delay Off Shifted Value.    */
#define MXC_S_SPIX_MASTER_CFG_INACT_DELAY_FOR_2_MOD_CLK                         ((uint32_t)(MXC_V_SPIX_MASTER_CFG_INACT_DELAY_FOR_2_MOD_CLK   << MXC_F_SPIX_MASTER_CFG_INACT_DELAY_POS))  /**< 2 Mode Clocks Field Shifted Value.  */
#define MXC_S_SPIX_MASTER_CFG_INACT_DELAY_FOR_4_MOD_CLK                         ((uint32_t)(MXC_V_SPIX_MASTER_CFG_INACT_DELAY_FOR_4_MOD_CLK   << MXC_F_SPIX_MASTER_CFG_INACT_DELAY_POS))  /**< 4 Mode Clocks Field Shifted Value.  */
#define MXC_S_SPIX_MASTER_CFG_INACT_DELAY_FOR_8_MOD_CLK                         ((uint32_t)(MXC_V_SPIX_MASTER_CFG_INACT_DELAY_FOR_8_MOD_CLK   << MXC_F_SPIX_MASTER_CFG_INACT_DELAY_POS))  /**< 8 Mode Clocks Field Shifted Value.  */
/**@}*/
/**
 * @defgroup SPIX_Fetch_ctrl_cmd_width Address Width Values and Shifted Values
 * @ingroup SPIX_Fetch_Ctrl_Register 
 * @brief Field values and shifted field values for selecting the SPIX Command Fetch Width
 * @{
 */
#define MXC_V_SPIX_FETCH_CTRL_CMD_WIDTH_SINGLE                                  ((uint32_t)(0x00000000UL))  /**< x1 command width field value.         */
#define MXC_V_SPIX_FETCH_CTRL_CMD_WIDTH_DUAL_IO                                 ((uint32_t)(0x00000001UL))  /**< x2 Dual command field value.         */
#define MXC_V_SPIX_FETCH_CTRL_CMD_WIDTH_QUAD_IO                                 ((uint32_t)(0x00000002UL))  /**< x4 Quad command field value.         */

#define MXC_S_SPIX_FETCH_CTRL_CMD_WIDTH_SINGLE                                  ((uint32_t)(MXC_V_SPIX_FETCH_CTRL_CMD_WIDTH_SINGLE    << MXC_F_SPIX_FETCH_CTRL_CMD_WIDTH_POS))  /**< x1 command width fetch shifted value.  */
#define MXC_S_SPIX_FETCH_CTRL_CMD_WIDTH_DUAL_IO                                 ((uint32_t)(MXC_V_SPIX_FETCH_CTRL_CMD_WIDTH_DUAL_IO   << MXC_F_SPIX_FETCH_CTRL_CMD_WIDTH_POS))  /**< x2 Dual command width fetch shifted value.  */
#define MXC_S_SPIX_FETCH_CTRL_CMD_WIDTH_QUAD_IO                                 ((uint32_t)(MXC_V_SPIX_FETCH_CTRL_CMD_WIDTH_QUAD_IO   << MXC_F_SPIX_FETCH_CTRL_CMD_WIDTH_POS))  /**< x4 Quad command width fetch shifted value.  */
/**@}*/
/**
 * @defgroup SPIX_Fetch_ctrl_addr_width Address Width Values and Shifted Values
 * @ingroup SPIX_Fetch_Ctrl_Register 
 * @brief Field values and shifted field values for selecting the SPIX Address Fetch Width
 * @{
 */
#define MXC_V_SPIX_FETCH_CTRL_ADDR_WIDTH_SINGLE                                 ((uint32_t)(0x00000000UL))  /**< x1 addr width field value.         */
#define MXC_V_SPIX_FETCH_CTRL_ADDR_WIDTH_DUAL_IO                                ((uint32_t)(0x00000001UL))  /**< x2 Dual addr  field value.         */
#define MXC_V_SPIX_FETCH_CTRL_ADDR_WIDTH_QUAD_IO                                ((uint32_t)(0x00000002UL))  /**< x4 Quad addr  field value.         */

#define MXC_S_SPIX_FETCH_CTRL_ADDR_WIDTH_SINGLE                                 ((uint32_t)(MXC_V_SPIX_FETCH_CTRL_ADDR_WIDTH_SINGLE    << MXC_F_SPIX_FETCH_CTRL_ADDR_WIDTH_POS))  /**< x1 addr width fetch shifted value.  */
#define MXC_S_SPIX_FETCH_CTRL_ADDR_WIDTH_DUAL_IO                                ((uint32_t)(MXC_V_SPIX_FETCH_CTRL_ADDR_WIDTH_DUAL_IO   << MXC_F_SPIX_FETCH_CTRL_ADDR_WIDTH_POS))  /**< x2 Dual addr width fetch shifted value.  */
#define MXC_S_SPIX_FETCH_CTRL_ADDR_WIDTH_QUAD_IO                                ((uint32_t)(MXC_V_SPIX_FETCH_CTRL_ADDR_WIDTH_QUAD_IO   << MXC_F_SPIX_FETCH_CTRL_ADDR_WIDTH_POS))  /**< x4 Quad addr width fetch shifted value.  */
/**@}*/
/**
 * @defgroup SPIX_Fetch_ctrl_data_width Data Width Values and Shifted Values
 * @ingroup SPIX_Fetch_Ctrl_Register 
 * @brief Field values and shifted field values for selecting the SPIX Data Fetch Width
 * @{
 */
#define MXC_V_SPIX_FETCH_CTRL_DATA_WIDTH_SINGLE                                 ((uint32_t)(0x00000000UL))  /**< Value to select x1 data width fetch for SPIX Field selection value.         */
#define MXC_V_SPIX_FETCH_CTRL_DATA_WIDTH_DUAL_IO                                ((uint32_t)(0x00000001UL))  /**< Value to select x2 Dual Mode data width fetch for SPIX Field selection value.         */
#define MXC_V_SPIX_FETCH_CTRL_DATA_WIDTH_QUAD_IO                                ((uint32_t)(0x00000002UL))  /**< Value to select x4 Quad Mode data width fetch for SPIX Field selection value.         */

#define MXC_S_SPIX_FETCH_CTRL_DATA_WIDTH_SINGLE                                 ((uint32_t)(MXC_V_SPIX_FETCH_CTRL_DATA_WIDTH_SINGLE    << MXC_F_SPIX_FETCH_CTRL_DATA_WIDTH_POS))  /**< x1 data width fetch shifted value.  */
#define MXC_S_SPIX_FETCH_CTRL_DATA_WIDTH_DUAL_IO                                ((uint32_t)(MXC_V_SPIX_FETCH_CTRL_DATA_WIDTH_DUAL_IO   << MXC_F_SPIX_FETCH_CTRL_DATA_WIDTH_POS))  /**< x2 Dual data width fetch shifted value.  */
#define MXC_S_SPIX_FETCH_CTRL_DATA_WIDTH_QUAD_IO                                ((uint32_t)(MXC_V_SPIX_FETCH_CTRL_DATA_WIDTH_QUAD_IO   << MXC_F_SPIX_FETCH_CTRL_DATA_WIDTH_POS))  /**< x4 Quad data width fetch shifted value.  */
/**@}*/
/**@} end of defgroup spix_registers */

#ifdef __cplusplus
}
#endif

#endif   /* _MXC_SPIX_REGS_H_ */

