/**
 * @file 
 * @brief      This file contains the function implementations for the Analog to
 *             Digital Converter (ADC) peripheral module.
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
 * $Date: 2017-02-14 18:08:19 -0600 (Tue, 14 Feb 2017) $
 * $Revision: 26419 $
 *
 *************************************************************************** */



/* **** Includes **** */
#include "mxc_config.h"
#include "mxc_assert.h"
#include "mxc_sys.h"
#include "adc.h"

/**
 * @ingroup group_adc
 * @{
 */

/* **** Definitions **** */

/* **** Globals **** */

/* **** Functions **** */

/* ************************************************************************* */
int ADC_Init(void)
{
  int err;

  if ((err = SYS_ADC_Init()) != E_NO_ERROR) {
    return err;
  }

  /* Wipe previous configuration */
  MXC_ADC->intr = 0;
  
  /* Clear all ADC interrupt flags (W1C) */
  MXC_ADC->intr = MXC_ADC->intr;

  /* Enable done interrupt */
  MXC_ADC->intr = MXC_F_ADC_INTR_ADC_DONE_IE;
  
  /* Power up the ADC */
  MXC_ADC->ctrl = (MXC_F_ADC_CTRL_ADC_PU |
		   MXC_F_ADC_CTRL_ADC_CLK_EN |
		   MXC_F_ADC_CTRL_BUF_PU |
		   MXC_F_ADC_CTRL_ADC_REFBUF_PU |
		   MXC_F_ADC_CTRL_ADC_CHGPUMP_PU);
  
  return E_NO_ERROR;
}

/* ************************************************************************* */
void ADC_StartConvert(mxc_adc_chsel_t channel, unsigned int adc_scale, unsigned int bypass)
{
  uint32_t ctrl_tmp;

  /* Clear the ADC done flag */
  ADC_ClearFlags(MXC_F_ADC_INTR_ADC_DONE_IF);
  
  /* Insert channel selection */
  ctrl_tmp = MXC_ADC->ctrl;
  ctrl_tmp &= ~(MXC_F_ADC_CTRL_ADC_CHSEL);
  ctrl_tmp |= ((channel << MXC_F_ADC_CTRL_ADC_CHSEL_POS) & MXC_F_ADC_CTRL_ADC_CHSEL);
  
  /* Clear channel configuration */
  ctrl_tmp &= ~(MXC_F_ADC_CTRL_ADC_REFSCL | MXC_F_ADC_CTRL_ADC_SCALE | MXC_F_ADC_CTRL_BUF_BYPASS);

  /* ADC reference scaling must be set for all channels but two*/
  if ((channel != ADC_CH_VDD18) && (channel != ADC_CH_VDD12)) {
    ctrl_tmp |= MXC_F_ADC_CTRL_ADC_REFSCL;
  }

  /* Finalize user-requested channel configuration */
  if (adc_scale || channel > ADC_CH_3) {
    ctrl_tmp |= MXC_F_ADC_CTRL_ADC_SCALE;
  }
  if (bypass) {
    ctrl_tmp |= MXC_F_ADC_CTRL_BUF_BYPASS;
  }
  
  /* Write this configuration */
  MXC_ADC->ctrl = ctrl_tmp;
  
  /* Start conversion */
  MXC_ADC->ctrl |= MXC_F_ADC_CTRL_CPU_ADC_START;

}

/* ************************************************************************* */
int ADC_GetData(uint16_t *outdata)
{
  /* See if a conversion is in process */
  if (MXC_ADC->status & MXC_F_ADC_STATUS_ADC_ACTIVE) {
    /* Wait for conversion to complete */
    while ((MXC_ADC->intr & MXC_F_ADC_INTR_ADC_DONE_IF) == 0);
  }

  /* Read 32-bit value and truncate to 16-bit for output depending on data align bit*/
  if((MXC_ADC->ctrl & MXC_F_ADC_CTRL_ADC_DATAALIGN) == 0)
      *outdata = (uint16_t)(MXC_ADC->data); /* LSB justified */
  else
      *outdata = (uint16_t)(MXC_ADC->data >> 6); /* MSB justified */

  /* Check for overflow */
  if (MXC_ADC->status & MXC_F_ADC_STATUS_ADC_OVERFLOW) {
    return E_OVERFLOW;
  }
  
  return E_NO_ERROR;
}

/* ************************************************************************* */
int ADC_SetLimit(mxc_adc_limitsel_t unit, mxc_adc_chsel_t channel,
		 unsigned int low_enable, unsigned int low_limit,
		 unsigned int high_enable, unsigned int high_limit)
{
  /* Check args */
  if ((unit >= ADC_LIMIT_MAX) || (channel >= ADC_CH_MAX))
    return E_BAD_PARAM;

  /* set channel using the limit */
  MXC_ADC->limit[unit] = ((channel << MXC_F_ADC_LIMIT0_CH_SEL_POS) & MXC_F_ADC_LIMIT0_CH_SEL);

  /* enable/disable the limit*/
  if (low_enable) {
    MXC_ADC->limit[unit] |= MXC_F_ADC_LIMIT0_CH_LO_LIMIT_EN |
      ((low_limit << MXC_F_ADC_LIMIT0_CH_LO_LIMIT_POS) & MXC_F_ADC_LIMIT0_CH_LO_LIMIT);
  }
  else{
      MXC_ADC->limit[unit] &= ~MXC_F_ADC_LIMIT0_CH_LO_LIMIT_EN;
  }

  if (high_enable) {
    MXC_ADC->limit[unit] |= MXC_F_ADC_LIMIT0_CH_HI_LIMIT_EN |
      ((high_limit << MXC_F_ADC_LIMIT0_CH_HI_LIMIT_POS) & MXC_F_ADC_LIMIT0_CH_HI_LIMIT);
  }
  else{
      MXC_ADC->limit[unit] &= ~MXC_F_ADC_LIMIT0_CH_HI_LIMIT_EN;
  }

  return E_NO_ERROR;
}

/** @} */
