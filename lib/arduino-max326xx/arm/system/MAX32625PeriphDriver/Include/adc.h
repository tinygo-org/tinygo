/*******************************************************************************
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
 * $Date: 2016-03-21 14:44:55 -0500 (Mon, 21 Mar 2016) $
 * $Revision: 22017 $
 *
 ******************************************************************************/

/**
 * @file  adc.h
 * @addtogroup adc ADC
 * @{
 * @brief High-level API for the Analog to Digital Converter (ADC)
 *
 */

#ifndef _ADC_H
#define _ADC_H

#include <stdint.h>
#include "adc_regs.h"

#ifdef __cplusplus
extern "C" {
#endif

/**
 * @brief Channel select
 */

typedef enum {
  ADC_CH_0              = MXC_V_ADC_CTRL_ADC_CHSEL_AIN0,
  ADC_CH_1              = MXC_V_ADC_CTRL_ADC_CHSEL_AIN1,
  ADC_CH_2              = MXC_V_ADC_CTRL_ADC_CHSEL_AIN2,
  ADC_CH_3              = MXC_V_ADC_CTRL_ADC_CHSEL_AIN3,
  ADC_CH_0_DIV_5        = MXC_V_ADC_CTRL_ADC_CHSEL_AIN0_DIV_5,
  ADC_CH_1_DIV_5        = MXC_V_ADC_CTRL_ADC_CHSEL_AIN1_DIV_5,
  ADC_CH_VDDB_DIV_4     = MXC_V_ADC_CTRL_ADC_CHSEL_VDDB_DIV_4,
  ADC_CH_VDD18          = MXC_V_ADC_CTRL_ADC_CHSEL_VDD18,
  ADC_CH_VDD12          = MXC_V_ADC_CTRL_ADC_CHSEL_VDD12,
  ADC_CH_VRTC_DIV_2     = MXC_V_ADC_CTRL_ADC_CHSEL_VRTC_DIV_2,
  ADC_CH_TMON           = MXC_V_ADC_CTRL_ADC_CHSEL_TMON,
#if (MXC_ADC_REV > 0)
  ADC_CH_VDDIO_DIV_4    = MXC_V_ADC_CTRL_ADC_CHSEL_VDDIO_DIV_4,
  ADC_CH_VDDIOH_DIV_4   = MXC_V_ADC_CTRL_ADC_CHSEL_VDDIOH_DIV_4,
#endif
  ADC_CH_MAX
} mxc_adc_chsel_t;

/**
 * @brief Limit monitor select
 */
typedef enum {
  ADC_LIMIT_0 = 0,
  ADC_LIMIT_1,
  ADC_LIMIT_2,
  ADC_LIMIT_3,
  ADC_LIMIT_MAX
} mxc_adc_limitsel_t;

#define ADC_IF_MASK (0xffffffffUL << MXC_F_ADC_INTR_ADC_DONE_IF_POS)
#define ADC_IE_MASK (0xffffffffUL >> MXC_F_ADC_INTR_ADC_DONE_IF_POS)

/**
 * @brief Initialize the ADC hardware
 *
 * @return E_NO_ERROR if successful
 */
int ADC_Init(void);

/**
 * @brief Start ADC conversion on the selected channel
 *
 * @param channel             Channel select from mxc_adc_chsel_t
 * @param adc_scale           Enable the ADC input scaling mode if non-zero
 * @param bypass              Bypass input buffer stage if non-zero
 *
 */
void ADC_StartConvert(mxc_adc_chsel_t channel, unsigned int adc_scale, unsigned int bypass);

/**
 * @brief Get the ADC result from the previous conversion
 *
 * @param outdata             Converted data from ADC 
 *
 * @return E_NO_ERROR if successful
 */
int ADC_GetData(uint16_t *outdata);

/**
 * @brief Set the data limits for a channel monitor
 *
 * @param unit                Which data limit unit to configure
 * @param channel             Channel select from mxc_adc_chsel_t
 * @param low_en              Enable the lower limit on this monitor
 * @param low_limit           Value for lower limit monitor
 * @param high_en             Enable the upper limit on this monitor
 * @param high_limit          Value for upper limit monitor
 * @returns                   E_NO_ERROR if everything is successful
 */
int ADC_SetLimit(mxc_adc_limitsel_t unit, mxc_adc_chsel_t channel,
		 unsigned int low_enable, unsigned int low_limit,
		 unsigned int high_enable, unsigned int high_limit);

/**
 * @brief Get interrupt flags
 *
 * @return Interrupt flag bits per adc_regs.h
 */
__STATIC_INLINE uint32_t ADC_GetFlags()
{
    return (MXC_ADC->intr & ADC_IF_MASK);
}

/**
 * @brief Clear interrupt flag(s)
 *
 * @param mask               Set the bits to clear
 *
 */
__STATIC_INLINE void ADC_ClearFlags(uint32_t mask)
{
    MXC_ADC->intr = ((MXC_ADC->intr & ADC_IE_MASK) | mask);
}

/**
 * @brief Get the Status of the ADC
 *
 * @return Status register per adc_regs.h
 */
__STATIC_INLINE uint32_t ADC_GetStatus()
{
    return (MXC_ADC->status);
}

/**
 * @brief Enable ADC interrupts based on mask
 *
 * @param mask              Set the bits to enable
 *
 */
__STATIC_INLINE void ADC_EnableINT(uint32_t mask)
{
    MXC_ADC->intr = ((MXC_ADC->intr & ADC_IE_MASK) | mask);
}

/**
 * @brief Disable ADC interrupts based on mask
 *
 * @param mask              Set the bits to disable
 *
 */
__STATIC_INLINE void ADC_DisableINT(uint32_t mask)
{
    MXC_ADC->intr = ((MXC_ADC->intr & ADC_IE_MASK) & ~mask);
}


/**
 * @}
 */

#ifdef __cplusplus
}
#endif

#endif
