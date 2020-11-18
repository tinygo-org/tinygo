/**
 * @file
 * @brief Pulse Train data types, definitions and function prototypes.
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
 * $Date: 2016-10-10 19:27:24 -0500 (Mon, 10 Oct 2016) $
 * $Revision: 24669 $
 *
 ***************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _PT_H_
#define _PT_H_

/* **** Includes **** */
#include "mxc_config.h"
#include "pt_regs.h"
#include "mxc_assert.h"
#include "mxc_sys.h"

#ifdef __cplusplus
extern "C" {
#endif

/**
 * @ingroup periphlibs
 * @defgroup pulsetrain Pulse Train Engine
 * @brief This is the high level API for the pulse train engine.
 * @{
 */

/**
 * Structure type for pulse train mode configuration.
 * @note       Do not use for square wave
 */
typedef struct { 
    uint32_t bps;           /**< pulse train bit rate */
    uint32_t pattern;       /**< Output pattern to shift out, starts at LSB */
    uint8_t ptLength;       /**< Number of bits in pulse train, 0 = 32bits, 1 = non valid , 2 = 2 bits, ... */
    uint16_t loop;          /**< Number of times to repeat the train, 0 = continuous */
    uint16_t loopDelay;     /**< Delay between loops specified in bits Example: loopDelay = 4,  delays time  = time it takes to shift out 4 bits */
} pt_pt_cfg_t;

/**
 * @brief      This function initializes the pulse trains to a known stopped
 *             state and sets the global PT clock scale.
 * @param      clk_scale  Scale the system clock for the global PT clock.
 */
void PT_Init(sys_pt_clk_scale clk_scale);

/**
 * @brief      Configures the pulse train in the specified mode.
 * @details    The parameters in the config structure must be set before calling
 *             this function. This function should be used for configuring pulse
 *             train mode only.
 * @note       The pulse train cannot be running when this function is called.
 *
 * @param      pt      Pulse train to operate on.
 * @param      cfg     Pointer to pulse train configuration.
 * @param      sysCfg  Pointer to pulse train system GPIO configuration.
 *
 * @return     #E_NO_ERROR if everything is successful, @ref MXC_Error_Codes
 *             "error" if unsuccessful.
 */
int PT_PTConfig(mxc_pt_regs_t *pt, pt_pt_cfg_t *cfg, const sys_cfg_pt_t *sysCfg);

/**
 * @brief   Configures the pulse train in the square wave mode.
 * @details This function should be used for configuring square wave mode only.
 * @note    The pulse train cannot be running when this function is called
 *
 * @param   pt      pulse train to operate on
 * @param   freq    square wave output frequency in Hz
 * @param   sysCfg  pointer to pulse train system GPIO configuration
 *
 * @returns #E_NO_ERROR if everything is successful, \ref MXC_Error_Codes "error" if unsuccessful.
 */
int PT_SqrWaveConfig(mxc_pt_regs_t *pt, uint32_t freq, const sys_cfg_pt_t *sysCfg);

/**
 * @brief   Starts the pulse train specified.
 *
 * @param   pt      Pulse train to operate on.
 */
__STATIC_INLINE void PT_Start(mxc_pt_regs_t *pt)
{
    int ptIndex = MXC_PT_GET_IDX(pt);

    MXC_PTG->enable |= (1 << ptIndex);

    //wait for PT to start
    while( (MXC_PTG->enable & (1 << ptIndex)) == 0 );
}

/**
 * @brief   Start multiple pulse train modules together.
 *
 * @param   pts     Set the bits of pulse trains to start
 *                  Bit0-\>pt0, Bit1-\>pt1... etc.
 */
__STATIC_INLINE void PT_StartMulti(uint32_t pts)
{
    MXC_PTG->enable |= pts;

    //wait for PTs to start
    while( (MXC_PTG->enable & pts) != pts );
}

/**
 * @brief   Stops a pulse train.
 *
 * @param   pt      Pulse train to operate on.
 */
__STATIC_INLINE void PT_Stop(mxc_pt_regs_t *pt)
{
    int ptIndex = MXC_PT_GET_IDX(pt);

    MXC_PTG->enable &= ~(1 << ptIndex);
}

/**
 * @brief   Stop multiple pulse trains together
 *
 * @param   pts     Set the bits of pulse trains to stop
 *                  Bit0-\>pt0, Bit1-\>pt1... etc.
 */
__STATIC_INLINE void PT_StopMulti(uint32_t pts)
{
    MXC_PTG->enable &= ~(pts);
}

/**
 * @brief   Determines if the pulse train is running.
 *
 * @param   pt      Pulse train to operate on.
 *
 * @return  0       Pulse train is off.
 * @return  \>0     Pulse train is on.
 */
__STATIC_INLINE uint32_t PT_IsActive(mxc_pt_regs_t *pt)
{
    int ptIndex = MXC_PT_GET_IDX(pt);

    return (!!(MXC_PTG->enable & (1 << ptIndex)));
}

/**
 * @brief      Determines if the pulse trains selected are running
 *
 * @param      pts   Set the bits of pulse trains to check Bit0-\>pt0,
 *                   Bit1-\>pt1... etc.
 *
 * @return     0            All pulse trains are off.
 * @return     \>0          At least one pulse train is on.
 */
__STATIC_INLINE uint32_t PT_IsActiveMulti(uint32_t pts)
{
    return (MXC_PTG->enable & pts);
}

/**
 * @brief   Sets the pattern of the pulse train
 *
 * @param   pt      Pointer to pulse train to operate on
 * @param   pattern Output pattern.
 *
 */
__STATIC_INLINE void PT_SetPattern(mxc_pt_regs_t *pt, uint32_t pattern)
{
    pt->train = pattern;
}

/**
 * @brief      Enable pulse train interrupt.
 *
 * @param      pt    Pointer to pulse train to operate on.
 */
__STATIC_INLINE void PT_EnableINT(mxc_pt_regs_t *pt)
{
    int ptIndex = MXC_PT_GET_IDX(pt);

    MXC_PTG->inten |= (1 << ptIndex);
}

/**
 * @brief      Enable interrupts for the pulse trains selected.
 *
 * @param      pts   Bit mask of which pulse trains to enable. Set the bit
 *                   position of each pulse train to enable it. Bit0-\>pt0,
 *                   Bit1-\>pt1... etc, 1 will enable the interrupt, 0 to leave
 *                   a PT channel in its current state.
 */
__STATIC_INLINE void PT_EnableINTMulti(uint32_t pts)
{
    MXC_PTG->inten |= pts;
}

/**
 * @brief      Disable pulse train interrupt.
 *
 * @param      pt    pulse train to operate on.
 */
__STATIC_INLINE void PT_DisableINT(mxc_pt_regs_t *pt)
{
    int ptIndex = MXC_PT_GET_IDX(pt);

    MXC_PTG->inten &= ~(1 << ptIndex);
}

/**
 * @brief      Disable interrupts for the pulse trains selected.
 *
 * @param      pts   Bit mask of what pulse trains to disable. Set the bit
 *                   position of each pulse train to disable it. Bit0-\>pt0,
 *                   Bit1-\>pt1... etc, 1 will disable the interrupt, 0 to leave
 *                   a PT channel in its current state.
 */
__STATIC_INLINE void PT_DisableINTMulti(uint32_t pts)
{
    MXC_PTG->inten &= ~pts;
}
/**
 * @brief      Gets the pulse trains's interrupt flags.
 *
 * @return     The Pulse Train Interrupt Flags, \ref PT_INTFL_Register Register
 *             for details.
 */
__STATIC_INLINE uint32_t PT_GetFlags(void)
{
    return MXC_PTG->intfl;
}

/**
 * @brief      Clears the pulse train's interrupt flag.
 *
 * @param      mask  bits to clear, see \ref PT_INTFL_Register Register for details. 
 */
__STATIC_INLINE void PT_ClearFlags(uint32_t mask)
{
    MXC_PTG->intfl = mask;
}

/**
 * @brief      Setup and enables a pulse train to restart after another pulse
 *             train has exited its loop. Each pulse train can have up to two
 *             restart triggers.
 *
 * @param      ptToRestart   pulse train to restart after @c ptStop ends.
 * @param      ptStop        pulse train that stops and triggers @p ptToRestart
 *                           to begin.
 * @param      restartIndex  selects which restart trigger to set (0 or 1).
 */
__STATIC_INLINE void PT_SetRestart(mxc_pt_regs_t *ptToRestart, mxc_pt_regs_t *ptStop, uint8_t restartIndex)
{
    int ptStopIndex = MXC_PT_GET_IDX(ptStop);

    MXC_ASSERT(ptStopIndex >= 0);

    if(restartIndex) {
        ptToRestart->restart |= (ptStopIndex << MXC_F_PT_RESTART_PT_Y_SELECT_POS) |
                                MXC_F_PT_RESTART_ON_PT_Y_LOOP_EXIT;
    } else {
        ptToRestart->restart |= (ptStopIndex << MXC_F_PT_RESTART_PT_X_SELECT_POS) |
                                MXC_F_PT_RESTART_ON_PT_X_LOOP_EXIT;
    }
}

/**
 * @brief      Disable the restart for the specified pulse train
 *
 * @param      ptToRestart   pulse train to disable the restart
 * @param      restartIndex  selects which restart trigger to disable (0 or 1)
 */
__STATIC_INLINE void PT_RestartDisable(mxc_pt_regs_t *ptToRestart, uint8_t restartIndex)
{
    if(restartIndex)
        ptToRestart->restart &= ~MXC_F_PT_RESTART_ON_PT_Y_LOOP_EXIT;
    else
        ptToRestart->restart &= ~MXC_F_PT_RESTART_ON_PT_X_LOOP_EXIT;
}

/**
 * @brief      Resynchronize individual pulse trains together. Resync will stop
 *             those resync_pts; others will be still running
 *
 * @param      resyncPts  pulse train modules that need to be re-synced by bit
 *                        number. Bit0-\>pt0, Bit1-\>pt1... etc.
 */
__STATIC_INLINE void PT_Resync(uint32_t resyncPts)
{
    MXC_PTG->resync = resyncPts;
    while(MXC_PTG->resync);
}
/**@} end of group pulsetrains*/

#ifdef __cplusplus
}
#endif

#endif /* _PT_H_ */
