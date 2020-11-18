/**
 * @file 
 * @brief      Assertion checks for debugging.
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
 *
 * $Date: 2017-02-16 13:35:22 -0600 (Thu, 16 Feb 2017) $ 
 * $Revision: 26485 $
 *
 *************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _MXC_ASSERT_H_
#define _MXC_ASSERT_H_

/* **** Includes **** */


#ifdef __cplusplus
extern "C" {
#endif

/**
 * @ingroup    syscfg
 * @defgroup   mxc_assertions Assertion Checks for Debugging
 * @brief      Assertion checks for debugging.
 * @{
 */ 
/* **** Definitions **** */
/**
 * @note       To use debug assertions, the symbol @c MXC_ASSERT_ENABLE must be
 *             defined. 
 */
///@cond
#ifdef MXC_ASSERT_ENABLE
/**
 * Macro that checks the expression for true and generates an assertion.
 * @note       To use debug assertions, the symbol @c MXC_ASSERT_ENABLE must be
 *             defined.
 */
#define MXC_ASSERT(expr)                                \
if (!(expr))                                            \
{                                                       \
    mxc_assert(#expr, __FILE__, __LINE__);              \
}
/**
 * Macro that generates an assertion with the message "FAIL".
 * @note       To use debug assertions, the symbol @c MXC_ASSERT_ENABLE must be
 *             defined.
 */
#define MXC_ASSERT_FAIL() mxc_assert("FAIL", __FILE__, __LINE__);
#else
#define MXC_ASSERT(expr)
#define MXC_ASSERT_FAIL()
#endif
///@endcond
/* **** Globals **** */

/* **** Function Prototypes **** */

/**
 * @brief      Assert an error when the given expression fails during debugging.
 * @param      expr  String with the expression that failed the assertion.
 * @param      file  File containing the failed assertion.
 * @param      line  Line number for the failed assertion.
 * @note       This is defined as a weak function and can be overridden at the
 *             application layer to print the debugging information. 
 *             @code 
 *             printf("%s, file: %s, line %d\n", expr, file, line);
 *             @endcode
 * @note       To use debug assertions, the symbol @c MXC_ASSERT_ENABLE must be
 *             defined. 
 */
void mxc_assert(const char *expr, const char *file, int line);

/**@} end of group MXC_Assertions*/

#ifdef __cplusplus
}
#endif

#endif /* _MXC_ASSERT_H_ */
