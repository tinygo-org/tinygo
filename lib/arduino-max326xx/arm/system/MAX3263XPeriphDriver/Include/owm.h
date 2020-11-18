/**
 * @file
 * @brief      Registers, Bit Masks and Bit Positions for the 1-Wire Master
 *             peripheral module.
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
 * $Date: 2016-03-14 10:08:53 -0500 (Mon, 14 Mar 2016) $
 * $Revision: 21855 $
 *
 **************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _OWM_H_
#define _OWM_H_

/* **** Includes **** */
#include "mxc_config.h"
#include "mxc_sys.h"
#include "owm_regs.h"

#ifdef __cplusplus
extern "C" {
#endif

/**
 * @ingroup periphlibs
 * @defgroup owm 1-Wire Master (OWM)
 * @{
 */

/* **** Definitions **** */

/**
 * Enumeration type for 1-Wire Overdrive Speed Options.
 */
typedef enum {
  OWM_OVERDRIVE_UNUSED = MXC_V_OWM_CTRL_STAT_OD_SPEC_MODE_12US, /**< 12us Overdrive Speed Select. */ 
  OWM_OVERDRIVE_12US = MXC_V_OWM_CTRL_STAT_OD_SPEC_MODE_12US,   /**< 12us Overdrive Speed Select. */
  OWM_OVERDRIVE_10US = MXC_V_OWM_CTRL_STAT_OD_SPEC_MODE_10US    /**< 10us Overdrive Speed Select. */
} owm_overdrive_t;

/**
 * Enumeration type for specifying options for 1-Wire external pullup mode.
 */
typedef enum {
  OWM_EXT_PU_ACT_HIGH = 0,  /**< Pullup pin is active high when enabled.        */
  OWM_EXT_PU_ACT_LOW = 1,   /**< Pullup pin is active low when enabled.         */
  OWM_EXT_PU_UNUSED = 2,    /**< Pullup pin is not used for an external pullup. */
} owm_ext_pu_t;

/**
 * Structure type for 1-Wire Master configuration.
 */
typedef struct {
    uint8_t int_pu_en;              /**< 1 = internal pullup on.   */
    owm_ext_pu_t ext_pu_mode;       /**< See #owm_ext_pu_t.   */
    uint8_t long_line_mode;         /**< 1 = long line mode enable.    */
    owm_overdrive_t overdrive_spec; /**< 0 = timeslot is 12us, 1 = timeslot is 10us.   */
} owm_cfg_t;


#define READ_ROM_COMMAND        0x33      /**< Read ROM Command */
#define MATCH_ROM_COMMAND       0x55      /**< Match ROM Command */
#define SEARCH_ROM_COMMAND      0xF0      /**< Search ROM Command */
#define SKIP_ROM_COMMAND        0xCC      /**< Skip ROM Command */
#define OD_SKIP_ROM_COMMAND     0x3C      /**< Overdrive Skip ROM Command */
#define OD_MATCH_ROM_COMMAND    0x69      /**< Overdrive Match ROM Command */
#define RESUME_COMMAND          0xA5      /**< Resume Command */

/* **** Globals **** */

/* **** Function Prototypes **** */

/**
 * @brief   Initialize and enable OWM module.
 * @param   owm         Pointer to OWM regs.
 * @param   cfg         Pointer to OWM configuration.
 * @param   sys_cfg     Pointer to system configuration object
 *
 * @retval  #E_NO_ERROR if everything is successful
 * @retval  #E_NULL_PTR if parameter is a null pointer
 * @retval  #E_BUSY if IOMAN was not configured correctly
 * @retval  #E_UNINITIALIZED if OWM CLK disabled
 * @retval  #E_NOT_SUPPORTED if 1MHz CLK cannot be created with given system and owm CLK
 * @retval  #E_BAD_PARAM if bad cfg parameter passed in
 */
int OWM_Init(mxc_owm_regs_t *owm, const owm_cfg_t *cfg, const sys_cfg_owm_t *sys_cfg);

/**
 * @brief   Shutdown OWM module.
 * @param   owm         Pointer to OWM regs.
 * @retval  #E_NO_ERROR if everything is successful
 * @retval  #E_BUSY if IOMAN was not released
 */
int OWM_Shutdown(mxc_owm_regs_t *owm);

/**
 * @brief   Send 1-Wire reset pulse. Will block until transaction is complete.
 * @param   owm         Pointer to OWM regs.
 * @retval  (0) = no presence pulse detected, (1) = presence pulse detected
 */
int OWM_Reset(mxc_owm_regs_t *owm);

/**
 * @brief   Send and receive one byte of data. Will block until transaction is complete.
 * @param   owm         Pointer to OWM regs.
 * @param   data        data to send
 * @retval  data read (1 byte)
 */
int OWM_TouchByte(mxc_owm_regs_t *owm, uint8_t data);

/**
 * @brief   Write one byte of data. Will block until transaction is complete.
 * @param   owm         Pointer to OWM regs.
 * @param   data        data to send
 * @retval  #E_NO_ERROR if everything is successful
 * @retval  #E_COMM_ERR if data written != data parameter
 */
int OWM_WriteByte(mxc_owm_regs_t *owm, uint8_t data);

/**
 * @brief   Read one byte of data. Will block until transaction is complete.
 * @param   owm         Pointer to OWM regs.
 * @retval  data read (1 byte)
 */
int OWM_ReadByte(mxc_owm_regs_t *owm);

/**
 * @brief   Send and receive one bit of data. Will block until transaction is complete.
 * @param   owm         Pointer to OWM regs.
 * @param   bit         bit to send
 * @retval  bit read
 */
int OWM_TouchBit(mxc_owm_regs_t *owm, uint8_t bit);

/**
 * @brief   Write one bit of data. Will block until transaction is complete.
 * @param   owm         Pointer to OWM regs.
 * @param   bit         bit to send
 * @retval  #E_NO_ERROR if everything is successful
 * @retval  #E_COMM_ERR if bit written != bit parameter
 */
int OWM_WriteBit(mxc_owm_regs_t *owm, uint8_t bit);

/**
 * @brief   Read one bit of data. Will block until transaction is complete.
 * @param   owm         Pointer to OWM regs.
 * @retval  bit read
 */
int OWM_ReadBit(mxc_owm_regs_t *owm);

/**
 * @brief   Write multiple bytes of data. Will block until transaction is complete.
 * @param   owm     Pointer to OWM regs.
 * @param   data    Pointer to buffer for write data.
 * @param   len     Number of bytes to write.
 *
 * @retval  Number of bytes written if successful
 * @retval  #E_COMM_ERR if line short detected before transaction
 */
int OWM_Write(mxc_owm_regs_t *owm, uint8_t* data, int len);

/**
 * @brief   Read multiple bytes of data. Will block until transaction is complete.
 * @param   owm     Pointer to OWM regs.
 * @param   data    Pointer to buffer for read data.
 * @param   len     Number of bytes to read.
 *
 * @retval Number of bytes read if successful
 * @retval #E_COMM_ERR if line short detected before transaction
 */
int OWM_Read(mxc_owm_regs_t *owm, uint8_t* data, int len);

/**
 * @brief   Starts 1-Wire communication with Read ROM command
 * @note    Only use the Read ROM command with one slave on the bus
 * @param   owm         Pointer to OWM regs.
 * @param   ROMCode     Pointer to buffer for ROM code read
 * @retval  #E_NO_ERROR if everything is successful
 * @retval  #E_COMM_ERR if reset, read or write fails
 */
int OWM_ReadROM(mxc_owm_regs_t *owm, uint8_t* ROMCode);

/**
 * @brief   Starts 1-Wire communication with Match ROM command
 * @param   owm         Pointer to OWM regs.
 * @param   ROMCode     Pointer to buffer with ROM code to match
 * @retval  #E_NO_ERROR if everything is successful
 * @retval  #E_COMM_ERR if reset or write fails
 */
int OWM_MatchROM(mxc_owm_regs_t *owm, uint8_t* ROMCode);

/**
 * @brief   Starts 1-Wire communication with Overdrive Match ROM command
 * @note    After Overdrive Match ROM command is sent, the OWM is set to
 *          overdrive speed. To set back to standard speed use OWM_SetOverdrive.
 * @param   owm         Pointer to OWM regs.
 * @param   ROMCode     Pointer to buffer with ROM code to match
 * @retval  #E_NO_ERROR if everything is successful
 * @retval  #E_COMM_ERR if reset or write fails
 */
int OWM_ODMatchROM(mxc_owm_regs_t *owm, uint8_t* ROMCode);

/**
 * @brief   Starts 1-Wire communication with Skip ROM command
 * @param   owm         Pointer to OWM regs.
 * @retval  #E_NO_ERROR if everything is successful
 * @retval  #E_COMM_ERR if reset or write fails
 */
int OWM_SkipROM(mxc_owm_regs_t *owm);

/**
 * @brief   Starts 1-Wire communication with Overdrive Skip ROM command
 * @note    After Overdrive Skip ROM command is sent, the OWM is set to
 *          overdrive speed. To set back to standard speed use OWM_SetOverdrive
 * @param   owm         Pointer to OWM regs.
 * @retval  #E_NO_ERROR if everything is successful
 * @retval  #E_COMM_ERR if reset or write fails
 */
int OWM_ODSkipROM(mxc_owm_regs_t *owm);

/**
 * @brief   Starts 1-Wire communication with Resume command
 * @param   owm         Pointer to OWM regs.
 * @retval  #E_NO_ERROR if everything is successful
 * @retval  #E_COMM_ERR if reset or write fails
 */
int OWM_Resume(mxc_owm_regs_t *owm);

/**
 * @brief   Starts 1-Wire communication with Search ROM command
 * @param   owm         Pointer to OWM regs.
 * @param   newSearch   (1) = start new search, (0) = continue search for next ROM
 * @param   ROMCode     Pointer to buffer with ROM code found
 * @retval  (1) = ROM found, (0) = no new ROM found, end of search
 */
int OWM_SearchROM(mxc_owm_regs_t *owm, int newSearch, uint8_t* ROMCode);

/**
 * @brief   Clear interrupt flags.
 * @param   owm         Pointer to OWM regs.
 * @param   mask        Mask of interrupts to clear.
 */
__STATIC_INLINE void OWM_ClearFlags(mxc_owm_regs_t *owm, uint32_t mask)
{
    owm->intfl = mask;
}

/**
 * @brief   Get interrupt flags.
 * @param   owm         Pointer to OWM regs.
 * @retval  Mask of active flags.
 */
__STATIC_INLINE unsigned OWM_GetFlags(mxc_owm_regs_t *owm)
{
    return (owm->intfl);
}

/**
 * @brief   Enables/Disables the External pullup
 * @param   owm         Pointer to OWM regs.
 * @param   enable      (1) = enable, (0) = disable
 */
__STATIC_INLINE void OWM_SetExtPullup(mxc_owm_regs_t *owm, int enable)
{
    if(enable)
        owm->cfg |= MXC_F_OWM_CFG_EXT_PULLUP_ENABLE;
    else
        owm->cfg &= ~(MXC_F_OWM_CFG_EXT_PULLUP_ENABLE);
}

/**
 * @brief   Enables/Disables Overdrive speed
 * @param   owm         Pointer to OWM regs.
 * @param   enable      (1) = overdrive, (0) = standard
 */
__STATIC_INLINE void OWM_SetOverdrive(mxc_owm_regs_t *owm, int enable)
{
    if(enable)
        owm->cfg |= MXC_F_OWM_CFG_OVERDRIVE;
    else
        owm->cfg &= ~(MXC_F_OWM_CFG_OVERDRIVE);
}

/**@} end of group owm */
#ifdef __cplusplus
}
#endif

#endif /* _OWM_H_ */
