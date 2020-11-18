/**
 * @file
 * @brief   Modular Arithmetic Accelerator (MAA) API Function Implementations. 
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
 **************************************************************************** */

/* **** Includes **** */
#include <string.h>
#include "mxc_assert.h"
#include "mxc_lock.h"
#include "mxc_sys.h"
#include "maa.h"

/**
 * @ingroup maa
 * @{
 */
  
///@cond
#define maa_is_running() (MXC_MAA->ctrl & MXC_F_MAA_CTRL_START ? 1 : 0)

/**
 * Macro that translates from #mxc_maa_reg_select_t to a pointer for MAA memory operations.
 */
#define UMAA_REGFILE_TO_ADDR(base, x) ((base + (MXC_MAA_HALF_SIZE * x)))

/**
 * Macro that adjusts the pointer so that it is pointing to the last 32-bit word in the maa (sub-)segment
 */
#define UMAA_ADDR_INDEX_LAST_32BIT(x) ((uint8_t *)x + (MXC_MAA_HALF_SIZE-4))
///@endcond


/* ************************************************************************* */
mxc_maa_ret_t MAA_Init(void)
{
  SYS_MAA_Init();

  return MXC_E_MAA_OK;
}

/* ************************************************************************* */
mxc_maa_ret_t MAA_WipeRAM(void)
{
  /* Check for running MAA */
  if (maa_is_running()) {
    return MXC_E_MAA_ERR;
  }
  
  /* Clear register files */
  memset((void *)MXC_MAA_MEM->seg0, 0, sizeof(MXC_MAA_MEM->seg0));
  memset((void *)MXC_MAA_MEM->seg1, 0, sizeof(MXC_MAA_MEM->seg1));
  memset((void *)MXC_MAA_MEM->seg2, 0, sizeof(MXC_MAA_MEM->seg2));
  memset((void *)MXC_MAA_MEM->seg3, 0, sizeof(MXC_MAA_MEM->seg3));
  memset((void *)MXC_MAA_MEM->seg4, 0, sizeof(MXC_MAA_MEM->seg4));
  memset((void *)MXC_MAA_MEM->seg5, 0, sizeof(MXC_MAA_MEM->seg5));
  
  return MXC_E_MAA_OK;
}

/* ************************************************************************* */
mxc_maa_ret_t MAA_Load(mxc_maa_reg_select_t regfile, const uint8_t *data, unsigned int size, mxc_maa_endian_select_t flag)
{
  uint32_t *maaptr;
  uint32_t fill;
  unsigned int zerotmp;
  
  if ((regfile > MXC_E_REG_51) || (size > MXC_MAA_REG_SIZE)) {
    /* Out of range */
    return MXC_E_MAA_ERR;
  }

  if (flag == MXC_MAA_F_MEM_REVERSE) {
    /* This is not currently implemented */
    return MXC_E_MAA_ERR;
  }

  maaptr = (uint32_t *)UMAA_REGFILE_TO_ADDR(MXC_BASE_MAA_MEM, regfile);

  /* 
   * MAA (sub-)segments must be loaded with zero pad to a 64-bit boundary, or the "garbage bits"
   *  will case erroneous results.
   */
  /* Find the ceiling for the closest 64-bit boundary based on the selected MAWS */
  zerotmp = (((MXC_MAA->maws & MXC_F_MAA_MAWS_MODLEN) >> MXC_F_MAA_MAWS_MODLEN_POS) + 63) & 0xfc0;
  /* Convert to bytes */
  zerotmp /= 8;
  
  /* Fill uMAA memory in long word sized chunks */
  while (size > 3) {
    *maaptr++ = (data[3] << 24) + (data[2] << 16) + (data[1] << 8) + data[0];
    data += 4;
    size -= 4;
    zerotmp = (zerotmp > 4) ? (zerotmp - 4) : 0;
  }
  
  /* Remainder */
  if (size) {
    fill = data[0];
    fill |= ((size > 1) ? (data[1] << 8) : 0);
    fill |= ((size > 2) ? (data[2] << 16) : 0);
    *maaptr++ = fill;
    
    /* We just filled 4 bytes in this section */
    zerotmp = (zerotmp > 4) ? (zerotmp - 4) : 0;
  }

  /* Wipe the remaining "garbage bits" */
  while (zerotmp) {
    *maaptr++ = 0;
    zerotmp = (zerotmp > 4) ? (zerotmp - 4) : 0;
  }
  
  return MXC_E_MAA_OK;
}

/* ************************************************************************* */
mxc_maa_ret_t MAA_Unload(mxc_maa_reg_select_t regfile, uint8_t *data, unsigned int size, mxc_maa_endian_select_t flag)
{
  uint32_t *maaptr;
  uint32_t fill;

  if ((regfile > MXC_E_REG_51) || (size > MXC_MAA_REG_SIZE)) {
    /* Out of range */
    return MXC_E_MAA_ERR;
  }

  if (flag == MXC_MAA_F_MEM_REVERSE) {
    /* This is not currently implemented */
    return MXC_E_MAA_ERR;
  }
  
  maaptr = (uint32_t *)UMAA_REGFILE_TO_ADDR(MXC_BASE_MAA_MEM, regfile);

  /* Unload uMAA memory in long word sized chunks */
  while (size > 3) {
    fill = *maaptr++;
    data[0] = fill & 0xff;
    data[1] = (fill >> 8) & 0xff;
    data[2] = (fill >> 16) & 0xff;
    data[3] = (fill >> 24) & 0xff;
    data += 4;
    size -= 4;
  }

  /* Remainder */
  if (size) {
    fill = *maaptr;
    data[0] = fill & 0xff;
    if (size > 1) {
      data[1] = (fill >> 8) & 0xff;
    }
    if (size > 2) {
      data[2] = (fill >> 16) & 0xff;
    }
  }

  return MXC_E_MAA_OK;
}

/* ************************************************************************* */
mxc_maa_ret_t MAA_Run(mxc_maa_operation_t op,       
          mxc_maa_reg_select_t al, mxc_maa_reg_select_t bl, 
          mxc_maa_reg_select_t rl, mxc_maa_reg_select_t tl)
{
  if (maa_is_running()) {
    /* Attempt to start the MAA while already running */
    return MXC_E_MAA_ERR;
  }

  /* Clear out any previous flags */
  MXC_MAA->ctrl = 0x00000020;

  /* Construct memory segment selections, select operation, and start the uMAA */
  MXC_MAA->ctrl = (((al << MXC_F_MAA_CTRL_SEG_A_POS) & MXC_F_MAA_CTRL_SEG_A) |
       ((bl << MXC_F_MAA_CTRL_SEG_B_POS) & MXC_F_MAA_CTRL_SEG_B) |
       ((rl << MXC_F_MAA_CTRL_SEG_RES_POS) & MXC_F_MAA_CTRL_SEG_RES) |
       ((tl << MXC_F_MAA_CTRL_SEG_TMP_POS) & MXC_F_MAA_CTRL_SEG_TMP) |
       ((op << MXC_F_MAA_CTRL_OPSEL_POS) & MXC_F_MAA_CTRL_OPSEL) |
       MXC_F_MAA_CTRL_START);

  /* Blocking wait for uMAA to complete. */
  while ((MXC_MAA->ctrl & MXC_F_MAA_CTRL_IF_DONE) == 0);

  if (MXC_MAA->ctrl & MXC_F_MAA_CTRL_IF_ERROR) {
    /* MAA signaled error */
    return MXC_E_MAA_ERR;
  }

  return MXC_E_MAA_OK;
}

/* ************************************************************************* */
mxc_maa_ret_t MAA_SetWordSize(unsigned int len)
{
  if ((len > MXC_MAA_REG_SIZE_BITS) || maa_is_running()) {
    return MXC_E_MAA_ERR;
  }

  /* Set bit length for calculation, and disable endian swap */
  MXC_MAA->maws = ((len << MXC_F_MAA_MAWS_MODLEN_POS) & MXC_F_MAA_MAWS_MODLEN);

  return MXC_E_MAA_OK;
}
/**@} end of ingroup maa */
