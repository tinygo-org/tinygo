/**
 * @file
 * @brief   Analog to Digital Converter function prototypes and data types.
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
 * $Date: 2017-05-04 09:49:55 -0500 (Thu, 04 May 2017) $
 * $Revision: 27759 $
 *
 *************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _ADC_H
#define _ADC_H

/* **** Includes **** */
#include <stdint.h>
/**
 * @ingroup periphlibs
 * @defgroup group_adc ADC
 * @brief 10-Bit Analog to Digital Converter
 * @{
 */
#include "adc_regs.h"

#ifdef __cplusplus
extern "C" {
#endif



/** 
 * @page adc_overview ADC Overview and Usage
 * The ADC supports both synchronous and asynchronous access. 
 * <b>Usage Examples</b>
 * @snippet snippets/adc_snippet.c ADC Synchronous Example
 * 
 * @snippet snippets/adc_snippet.c ADC Asynchronous Example
 * 
 */

/* **** Definitions **** */

/**
 * Enumeration type for ADC Channel Selection. See \ref ADC_CHSEL_values "ADC Channel Select Values" for additional information.
 */
typedef enum {
  ADC_CH_0              = MXC_V_ADC_CTRL_ADC_CHSEL_AIN0,          /**< Channel 0 Select */
  ADC_CH_1              = MXC_V_ADC_CTRL_ADC_CHSEL_AIN1,          /**< Channel 1 Select */
  ADC_CH_2              = MXC_V_ADC_CTRL_ADC_CHSEL_AIN2,          /**< Channel 2 Select */
  ADC_CH_3              = MXC_V_ADC_CTRL_ADC_CHSEL_AIN3,          /**< Channel 3 Select */
  ADC_CH_0_DIV_5        = MXC_V_ADC_CTRL_ADC_CHSEL_AIN0_DIV_5,    /**< Channel 0 divided by 5 */
  ADC_CH_1_DIV_5        = MXC_V_ADC_CTRL_ADC_CHSEL_AIN1_DIV_5,    /**< Channel 1 divided by 5 */
  ADC_CH_VDDB_DIV_4     = MXC_V_ADC_CTRL_ADC_CHSEL_VDDB_DIV_4,    /**< VDDB divided by 4 */
  ADC_CH_VDD18          = MXC_V_ADC_CTRL_ADC_CHSEL_VDD18,         /**< VDD18 input select */
  ADC_CH_VDD12          = MXC_V_ADC_CTRL_ADC_CHSEL_VDD12,         /**< VDD12 input select */
  ADC_CH_VRTC_DIV_2     = MXC_V_ADC_CTRL_ADC_CHSEL_VRTC_DIV_2,    /**< VRTC divided by 2 */
  ADC_CH_TMON           = MXC_V_ADC_CTRL_ADC_CHSEL_TMON,          /**< TMON input select */
#if (MXC_ADC_REV > 0)
  ADC_CH_VDDIO_DIV_4    = MXC_V_ADC_CTRL_ADC_CHSEL_VDDIO_DIV_4,   /**< VDDIO divided by 4 select */
  ADC_CH_VDDIOH_DIV_4   = MXC_V_ADC_CTRL_ADC_CHSEL_VDDIOH_DIV_4,  /**< VDDIOH divided by 4 select */
#endif
  ADC_CH_MAX                                                      /**< Max enum value for channel selection */
} mxc_adc_chsel_t;

/**
 * Enumeration type for the ADC limit register to set 
 */
typedef enum {
  ADC_LIMIT_0 = 0,      /**< ADC Limit Register 0 */
  ADC_LIMIT_1 = 1,      /**< ADC Limit Register 1 */
  ADC_LIMIT_2 = 2,      /**< ADC Limit Register 2 */
  ADC_LIMIT_3 = 3,      /**< ADC Limit Register 3 */
  ADC_LIMIT_MAX         /**< Number of Limit registers */
} mxc_adc_limitsel_t;

/**
 * Mask for all Interrupt Flag Fields
 */
#define ADC_IF_MASK (0xffffffffUL << MXC_F_ADC_INTR_ADC_DONE_IF_POS)

/**
 * Mask for all Interrupt Enable Fields
 */
#define ADC_IE_MASK (0xffffffffUL >> MXC_F_ADC_INTR_ADC_DONE_IF_POS)

/* **** Function Prototypes **** */

/**
 * @brief      Initialize the ADC hardware
 *
 * @return     #E_NO_ERROR if successful
 */
int ADC_Init(void);

/**
 * @brief      Start ADC conversion on the selected channel
 *
 * @param      channel    Channel select from #mxc_adc_chsel_t
 * @param      adc_scale  Enable the ADC input scaling mode if non-zero
 * @param      bypass     Bypass input buffer stage if non-zero
 */
void ADC_StartConvert(mxc_adc_chsel_t channel, unsigned int adc_scale, unsigned int bypass);

/**
 * @brief      Gets the result from the previous ADC conversion
 *
 * @param      outdata      Pointer to store the ADC data conversion 
 *                          result.
 * @return     #E_OVERFLOW   ADC overflow error
 * @return     #E_NO_ERROR   Data returned in outdata parameter
 */
int ADC_GetData(uint16_t *outdata);

/**
 * @brief      Set the data limits for an ADC channel monitor
 *
 * @param      unit         Which data limit unit to configure
 * @param      channel      Channel select from mxc_adc_chsel_t
 * @param      low_enable   Enable the lower limit on this monitor
 * @param      low_limit    Value for lower limit monitor
 * @param      high_enable  Enable the upper limit on this monitor
 * @param      high_limit   Value for upper limit monitor
 *
 * @return     #E_BAD_PARAM  ADC limit or channel greater than supported
 * @return     #E_NO_ERROR   ADC limit set successfully
 */
int ADC_SetLimit(mxc_adc_limitsel_t unit, mxc_adc_chsel_t channel,
                 unsigned int low_enable, unsigned int low_limit,
                 unsigned int high_enable, unsigned int high_limit);

/**
 * @brief      Get interrupt flags
 *
 * @return     ADC Interrupt flags bit mask. See the @ref ADC_INTR_IF_Register
 *             "ADC_INTR Register" for the interrupt flag masks.
 */
__STATIC_INLINE uint32_t ADC_GetFlags()
{
    return (MXC_ADC->intr & ADC_IF_MASK);
}

/**
 * @brief      Clear interrupt flag(s) using the mask parameter. All bits set in
 *             the parameter will be cleared.
 *
 * @param      mask  Interrupt flags to clear. See the @ref ADC_INTR_IF_Register
 *                   "ADC_INTR Register" for the interrupt flag masks.
 */
__STATIC_INLINE void ADC_ClearFlags(uint32_t mask)
{
    MXC_ADC->intr = ((MXC_ADC->intr & ADC_IE_MASK) | mask); // Mask off the Interrupt Enable Bits so they don't get disabled.
}

/**
 * @brief      Get the Status of the ADC
 *
 * @return     ADC status register. See @ref ADC_STATUS_Register "ADC_STATUS
 *             Register" for details.
 */
__STATIC_INLINE uint32_t ADC_GetStatus()
{
    return (MXC_ADC->status);
}

/**
 * @brief      Enables the ADC interrupts specified by the mask parameter
 *
 * @param      mask  ADC interrupts to enable. See @ref ADC_INTR_IE_Register
 *                   "ADC_INTR Register" for the interrupt enable bit masks.
 */
__STATIC_INLINE void ADC_EnableINT(uint32_t mask)
{
    MXC_ADC->intr = ((MXC_ADC->intr & ADC_IE_MASK) | mask);
}

/**
 * @brief      Disable ADC interrupts based on mask
 *
 * @param      mask  ADC interrupts to disable. See @ref ADC_INTR_IE_Register
 *                   "ADC_INTR Register" for the interrupt enable bit masks.
 */
__STATIC_INLINE void ADC_DisableINT(uint32_t mask)
{
    MXC_ADC->intr = ((MXC_ADC->intr & ADC_IE_MASK) & ~mask);
}
/**@} end of group_adc */

#ifdef __cplusplus
}
#endif
 
#endif /* _ADC_H */
