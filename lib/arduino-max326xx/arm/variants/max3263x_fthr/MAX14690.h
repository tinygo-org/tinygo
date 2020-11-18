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
 *******************************************************************************
 */

#ifndef _MAX14690_H_
#define _MAX14690_H_

#define MAX14690_NO_ERROR   0
#define MAX14690_ERROR      -1

// 7 bit address
#define MAX14690_I2C_ADDR 0x28

#define MAX14690_LDO_MIN_MV 800
#define MAX14690_LDO_MAX_MV 3600
#define MAX14690_LDO_STEP_MV 100

#define MAX14690_OFF_COMMAND 0xB2

class MAX14690
{
public:

    /**
     * @brief   Register Addresses
     * @details Enumerated MAX14690 register addresses
     */
    enum registers_t {
        REG_CHIP_ID = 0x00,		///< Chip ID
        REG_CHIP_REV = 0x01,	///< Chip Revision
        REG_STATUS_A = 0x02,   ///< Status Register A
        REG_STATUS_B = 0x03,   ///< Status Register B
        REG_STATUS_C = 0x04,   ///< Status Register C
        REG_INT_A = 0x05,   ///< Interrupt Register A
        REG_INT_B = 0x06,   ///< Interrupt Register B
        REG_INT_MASK_A = 0x07,   ///< Interrupt Mask A
        REG_INT_MASK_B = 0x08,   ///< Interrupt Mask B
        REG_I_LIM_CNTL = 0x09,   ///< Input Limit Control
        REG_CHG_CNTL_A = 0x0A,   ///< Charger Control A
        REG_CHG_CNTL_B = 0x0B,   ///< Charger Control B
        REG_CHG_TMR = 0x0C,   ///< Charger Timers
        REG_BUCK1_CFG = 0x0D,   ///< Buck 1 Configuration
        REG_BUCK1_VSET = 0x0E,   ///< Buck 1 Voltage Setting
        REG_BUCK2_CFG = 0x0F,   ///< Buck 2 Configuration
        REG_BUCK2_VSET = 0x10,   ///< Buck 2 Voltage Setting
        REG_RSVD_11 = 0x11,   ///< Reserved 0x11
        REG_LDO1_CFG = 0x12,   ///< LDO 1 Configuration
        REG_LDO1_VSET = 0x13,   ///< LDO 1 Voltage Setting
        REG_LDO2_CFG = 0x14,   ///< LDO 2 Configuration
        REG_LDO2_VSET = 0x15,   ///< LDO 2 Voltage Setting
        REG_LDO3_CFG = 0x16,   ///< LDO 3 Configuration
        REG_LDO3_VSET = 0x17,   ///< LDO 3 Voltage Setting
        REG_THRM_CFG = 0x18,   ///< Thermistor Configuration
        REG_MON_CFG = 0x19,   ///< Monitor Multiplexer Configuration
        REG_BOOT_CFG = 0x1A,   ///< Boot Configuration
        REG_PIN_STATUS = 0x1B,   ///< Pin Status
        REG_BUCK_EXTRA = 0x1C,   ///< Additional Buck Settings
        REG_PWR_CFG = 0x1D,   ///< Power Configuration
        REG_NULL = 0x1E,   ///< Reserved 0x1E
        REG_PWR_OFF = 0x1F,   ///< Power Off Register
    };

    /**
     * @brief   Thermal Status
     * @details Thermal status determined by thermistor
     */
    enum thermStat_t {
        THMSTAT_000,	///< T < T1
        THMSTAT_001,	///< T1 < T < T2
        THMSTAT_010,	///< T2 < T < T3
        THMSTAT_011,	///< T3 < T < T4
        THMSTAT_100,	///< T > T4
        THMSTAT_101,	///< No theremistor detected
        THMSTAT_110,	///< Thermistor Disabled by ThermEn
        THMSTAT_111,	///< CHGIN not present
    };

    /**
     * @brief   Charge Status
     * @details Current operating mode of charger
     */
    enum chgStat_t {
        CHGSTAT_000,	///< Charger off
        CHGSTAT_001,	///< Charging suspended by temperature
        CHGSTAT_010,	///< Pre-charge
        CHGSTAT_011,	///< Fast-charge constant current
        CHGSTAT_100,	///< Fast-charge constant voltage
        CHGSTAT_101,	///< Maintain charge
        CHGSTAT_110,	///< Done
        CHGSTAT_111,	///< Charger fault
    };

    /**
     * @brief   Input Current Limit
     * @details CHGIN input current limit values
     */
    enum iLimCntl_t {
        ILIM_0mA,		///< 0mA
        ILIM_100mA,		///< 100mA
        ILIM_500mA,		///< 500mA
        ILIM_1000mA,		///< 1000mA
    };

    /**
     * @brief   Recharge Threshold
     * @details Battery recharge voltage threshold
     */
    enum batReChg_t {
        BAT_RECHG_70mV,		///< 70mV
        BAT_RECHG_120mV,	///< 120mV
        BAT_RECHG_170mV,	///< 170mV
        BAT_RECHG_220mV,	///< 220mV
    };

    /**
     * @brief   Battery Regulation Voltage
     * @details Battery regulation voltages set point
     */
    enum batReg_t {
        BAT_REG_4050mV,		///< 4.05V
        BAT_REG_4100mV,		///< 4.10V
        BAT_REG_4150mV,		///< 4.15V
        BAT_REG_4200mV,		///< 4.20V
        BAT_REG_4250mV,		///< 4.25V
        BAT_REG_4300mV,		///< 4.30V
        BAT_REG_4350mV,		///< 4.35V
        BAT_REG_RSVD,		///< reserved
    };

    /**
     * @brief   Precharge Voltage
     * @details Battery precharge voltage threshold
     */
    enum vPChg_t {
        VPCHG_2100mV,		///< 2.10V
        VPCHG_2250mV,		///< 2.25V
        VPCHG_2400mV,		///< 2.40V
        VPCHG_2550mV,		///< 2.55V
        VPCHG_2700mV,		///< 2.70V
        VPCHG_2850mV,		///< 2.85V
        VPCHG_3000mV,		///< 3.00V
        VPCHG_3150mV,		///< 3.15V
    };

    /**
     * @brief   Precharge Current
     * @details Battery precharge current value
     */
    enum iPChg_t {
        IPCHG_5,		///< 5% of Ifchg
        IPCHG_10,		///< 10% of Ifchg
        IPCHG_20,		///< 20% of Ifchg
        IPCHG_30,		///< 30% of Ifchg
    };

    /**
     * @brief   Done Current
     * @details Charger done current where charging stops
     */
    enum chgDone_t {
        CHGDONE_5,		///< 5% of Ifchg
        CHGDONE_10,		///< 10% of Ifchg
        CHGDONE_20,		///< 20% of Ifchg
        CHGDONE_30,		///< 30% of Ifchg
    };

    /**
     * @brief   Maintain Charge Timer
     * @details Timeout settings for maintain charge mode
     */
    enum mtChgTmr_t {
        MTCHGTMR_0min,		///< 0 min
        MTCHGTMR_15min,		///< 15 min
        MTCHGTMR_30min,		///< 30 min
        MTCHGTMR_60min,		///< 60 min
    };

    /**
     * @brief   Fast Charge Timer
     * @details Timeout settings for fast charge mode
     */
    enum fChgTmr_t {
        FCHGTMR_75min,		///< 75 min
        FCHGTMR_150min,		///< 150 min
        FCHGTMR_300min,		///< 300 min
        FCHGTMR_600min,		///< 600 min
    };

    /**
     * @brief   Precharge Timer
     * @details Timeout settings for precharge mode
     */
    enum pChgTmr_t {
        PCHGTMR_30min,		///< 30 min
        PCHGTMR_60min,		///< 60 min
        PCHGTMR_120min,		///< 120 min
        PCHGTMR_240min,		///< 240 min
    };

    /**
     * @brief   LDO Enable Mode
     * @details Enumerated enable modes for voltage regulators
     */
    enum ldoMode_t {
        LDO_DISABLED,	///< Disabled, Regulator Mode
        SW_DISABLED,	///< Disabled, Switch Mode
        LDO_ENABLED,	///< Enabled, Regulator Mode
        SW_ENABLED,	///< Enabled, Switch Mode
        LDO_EN_MPC0,	///< Regulator Enabled by MPC pin
        SW_EN_MPC0,	///< Switch Enabled by MPC pin
        LDO_EN_MPC1,	///< Regulator Enabled by MPC pin
        SW_EN_MPC1,	///< Switch Enabled by MPC pin
        LDO_DISABLED_DSC,	///< Regulator Disabled
        SW_DISABLED_DSC,	///< Switch Disabled
        LDO_ENABLED_DSC,	///< Regulator Enabled
        SW_ENABLED_DSC,	///< Switch Enabled
        LDO_EN_MPC0_DSC,	///< Regulator Enabled by MPC pin
        SW_EN_MPC0_DSC,	///< Switch Enabled by MPC pin
        LDO_EN_MPC1_DSC,	///< Regulator Enabled by MPC pin
        SW_EN_MPC1_DSC,	///< Switch Enabled by MPC pin
    };

    /**
     * @brief   Buck Operating Modes
     * @details Enumerated operating modes for buck regulator
     */
    enum buckMd_t {
        BUCK_BURST,		///< Burst Mode Operation
        BUCK_FPWM,		///< Forced PWM Operation
        BUCK_MPC0_FPWM,	///< MPC activated Forced PWM
        BUCK_MPC1_FPWM,	///< MPC activated Forced PWM
    };

    /**
     * @brief   Thermistor Configuration
     * @details Enumerated thermistor operating modes
     */
    enum thrmCfg_t {
        THRM_DISABLED,		///< Thermistor monitoring disabled
        THRM_ENABLED,		///< Basic thermistor monitoring
        THRM_RSVD,	///< reserved, do not use
        THRM_JEITA,	///< JEITA thermistor monitoring
    };

    /**
     * @brief   Monitor Configurations
     * @details Enumerated configuration modes for monitor multiplexer
     */
    enum monCfg_t {
        MON_PULLDOWN = 0x0,	///< Pulled down by 100k Ohm
        MON_BAT = 0x1,		///< BAT Selected
        MON_SYS = 0x2,		///< SYS Selected
        MON_BUCK1 = 0x3,	///< BUCK1 Selected
        MON_BUCK2 = 0x4,	///< BUCK2 Selected
        MON_LDO1 = 0x5,		///< LDO1 Selected
        MON_LDO2 = 0x6,		///< LDO2 Selected
        MON_LDO3 = 0x7,		///< LDO3nSelected
        MON_HI_Z = 0x8,		///< High Impedance
    };

    /**
     * @brief   Monitor Divide Ratio
     * @details Ratio settings for monitor divider
     */
    enum monRatio_t {
        MON_DIV4,		///< 4:1 Monitor Ratio
        MON_DIV3,		///< 3:1 Monitor Ratio
        MON_DIV2,		///< 2:1 Monitor Ratio
        MON_DIV1,		///< 1:1 Monitor Ratio
    };

    /**
     * MAX14690 constructor.
     */
    MAX14690();

    /**
     * MAX14690 destructor.
     */
    ~MAX14690();

    /**
     * @brief   Initialize MAX14690
     * @details Applies settings to MAX14690.
     *  Settings are stored in public variables.
     *  The variables are pre-loaded with the most common configuation.
     *  Assign new values to the public variables before calling init.
     *  This will update all the settings including the LDO voltages
     *  and modes.
     * @returns 0 if no errors, -1 if error.
    */
    int init();

    /**
     * @brief   Set the LDO Voltage
     * @details Sets the voltage for the boost regulator.
     *  The voltage is specified in millivolts.
     *  The MAX14690 cannot update the voltage when enabled.
     *  This function checks the local ldoMode variable and if the
     *  regualtor is enabled it will send the disable command before
     *  sending the new voltage and re-enable the LDO after
     *  the new voltage is written.
     * @param   mV voltage for boost regualtor in millivolts
     * @returns 0 if no errors, -1 if error.
    */
    int ldo2SetVoltage(int mV);

    /**
     * @brief   Set LDO Enable Mode
     * @details Sets the enable mode for the LDO/SW
     * @param   mode The enable mode for the LDO/SW
     * @returns 0 if no errors, -1 if error.
    */
    int ldo2SetMode(ldoMode_t mode);

    /**
     * @brief   Set the LDO Voltage
     * @details Sets the voltage for the boost regulator.
     *  The voltage is specified in millivolts.
     *  The MAX14690 cannot update the voltage when enabled.
     *  This function checks the local ldoMode variable and if the
     *  regualtor is enabled it will send the disable command before
     *  sending the new voltage and re-enable the LDO after
     *  the new voltage is written.
     * @param   mV voltage for boost regualtor in millivolts
     * @returns 0 if no errors, -1 if error.
    */
    int ldo3SetVoltage(int mV);

    /**
     * @brief   Set LDO Enable Mode
     * @details Sets the enable mode for the LDO/SW
     * @param   mode The enable mode for the LDO/SW
     * @returns 0 if no errors, -1 if error.
    */
    int ldo3SetMode(ldoMode_t mode);

    /**
     * @brief   Configure Mon Pin
     * @details Configures the operating mode of the monitor multiplexer
     * @param   monCfg The configuration mode for the monitor pin
     * @returns 0 if no errors, -1 if error.
    */
    int monSet(monCfg_t monCfg, monRatio_t monRatio);

    /**
     * @brief   Shutdown
     * @details Sends the command to turn off all supplies and put the part
     *  in battery saving shelf mode.
     * @returns 0 if no errors, -1 if error.
    */
    int shutdown();

    /**
     * @brief   Reset settings to default values
     * @details Resets all local variables to the default value.
     *  Note: this only resets the local variables and has no effect
     *  on the part until they are applied by another functions such as
     *  init();
     * @returns 0 if no errors, -1 if error.
    */
	void resetToDefaults();

    /**
     * @brief   Write Register
     * @details Writes the given value to the specified register.
     *  Note, this function provides direct access to the registers
     *  without any awareness or effect on the settings stored in
     *  the public variables.  This is used by the other functions to
     *  set the values inside the MAX14690.  Calling this outside of the
     *  other functions can break the synchronization of the variables
     *  to the state of the MAX14690.
     * @param   reg The register to be written
     * @param   value The data to be written
     * @returns 0 if no errors, -1 if error.
    */
    int writeReg(registers_t reg, char value);

    /**
     * @brief   Read Register
     * @details Reads from the specified register
     * @param   reg The register to be read
     * @param   value Pointer for where to store the data
     * @returns 0 if no errors, -1 if error.
    */
    int readReg(registers_t reg, char *value);

    /// Thermal Status Change Interrupt Enable: default 0 - Disabled, 1 - Enabled
    bool intEnThermStatus;
    /// Charger Status Change Interrupt Enable: default 0 - Disabled, 1 - Enabled
    bool intEnChgStatus;
    /// Input Limit Interrupt Enable: default 0 - Disabled, 1 - Enabled
    bool intEnILim;
    /// USB Over Voltage Interrupt Enable: default 0 - Disabled, 1 - Enabled
    bool intEnUSBOVP;
    /// USB OK Interrupt Enable: default 0 - Disabled, 1 - Enabled
    bool intEnUSBOK;
    /// Charger Thermal Shutdown Interrupt Enable: default 0 - Disabled, 1 - Enabled
    bool intEnChgThmSD;
    /// Thermal Regulation Interrupt Enable: default 0 - Disabled, 1 - Enabled
    bool intEnThermReg;
    /// Charger Timeout Interrupt Enable: default 0 - Disabled, 1 - Enabled
    bool intEnChgTimeOut;
    /// Buck1 Thermal Error Interrupt Enable: default 0 - Disabled, 1 - Enabled
    bool intEnThermBuck1;
    /// Buck2 Thermal Error Interrupt Enable: default 0 - Disabled, 1 - Enabled
    bool intEnThermBuck2;
    /// LDO1 Thermal Error Interrupt Enable: default 0 - Disabled, 1 - Enabled
    bool intEnThermLDO1;
    /// LDO2 Thermal Error Interrupt Enable: default 0 - Disabled, 1 - Enabled
    bool intEnThermLDO2;
    /// LDO3 Thermal Error Interrupt Enable: default 0 - Disabled, 1 - Enabled
    bool intEnThermLDO3;
    /// CHGIN Input Current Limit Setting: default 500mA
    iLimCntl_t iLimCntl;
    /// Charger Auto Stop: default 1 - move to Charge Done when charging complete, 0 - remain in Maintain Charge mode
    bool chgAutoStp;
    /// Charger Auto Restart: default 1 - restart charging when Vbat drops below threshold, 0 - stay in Charge Done until charging disabled
    bool chgAutoReSta;
    /// Charger Battery Recharge Threshold: default -120mV
    batReChg_t batReChg;
    /// Charger Battery Regulation Voltage: default 4.20V
    batReg_t batReg;
    /// Charger Enable: default 1 - Enabled, 0 - Disabled
    bool chgEn;
    /// Charger Precharge Voltage Threshold: default 3.00V
    vPChg_t vPChg;
    /// Charger Precharge Current Setting: default 10% of fast charge current
    iPChg_t iPChg;
    /// Charger Done Threshold, stop charging when charge current drops to this level: default 10% of fast charge current
    chgDone_t chgDone;
    /// Maintain Charge Timer, time to wait after reaching done current before disabling charger: default 0 min
    mtChgTmr_t mtChgTmr;
    /// Fast Charge Timer, timeout for fast charge duration: default 300 min
    fChgTmr_t fChgTmr;
    /// Precharge Timer, timeout for precharge duration: default 60 min
    pChgTmr_t pChgTmr;
    /// Buck 1 Mode Select: default Burst mode
    buckMd_t buck1Md;
    /// Buck 1 Inductor Select: default 0 - 2.2uH, 1 - 4.7uH
    bool buck1Ind;
    /// Buck 2 Mode Select: default BUCK_BURST
    buckMd_t buck2Md;
    /// Buck 2 Inductor Select: default 0 - 2.2uH, 1 - 4.7uH
    bool buck2Ind;
    /// LDO 2 Mode Select: default LDO_DISABLED
    ldoMode_t ldo2Mode;
    /// LDO 2 Voltage in millivolts: default 3200
    int ldo2Millivolts;
    /// LDO 3 Mode Select: default LDO_DISABLED
    ldoMode_t ldo3Mode;
    /// LDO 3 Voltage in millivolts: default 3000
    int ldo3Millivolts;
    /// Thermistor Configuration: default THRM_ENABLED
    thrmCfg_t thrmCfg;
    /// Monitor Multiplexer Divider Ratio Select: default MON_DIV4
    monRatio_t monRatio;
    /// Monitor Multiplexer Configuration: default MON_PULLDOWN
    monCfg_t monCfg;
    /// Buck 2 Active Discharge: default 0 - discharge only during hard reset, 1 - discharge when regulator is disabled
    bool buck2ActDsc;
    /// Buck 2 Force FET Scaling: default 0 - full FET for better active efficiency, 1 - reduced FET for lower quiescent current
    bool buck2FFET;
    /// Buck 1 Active Discharge: default 0 - discharge only during hard reset, 1 - discharge when regulator is disabled
    bool buck1ActDsc;
    /// Buck 1 Force FET Scaling: default 0 - full FET for better active efficiency, 1 - reduced FET for lower quiescent current
    bool buck1FFET;
    /// PFN pin resistor enable: default 1 - internal pullup/pulldown enabled, 0 - internal pullup/pulldown disabled
    bool pfnResEna;
    /// Stay On Handshake: default 1 - remain on, 0 - turn off if not set within 5s of power-on
    bool stayOn;

private:
    // Internal Resources
    /**
     * @brief   Converts mV to register bits
     * @details Converts integer representing the desired voltage
     *  in millivolts to the coresponding 8 bit register value.
     *  This will check to ensure the voltage is within the allowed
     *  range and return an error (-1) if it is out of range.
     * @param   mV voltage for LDO regulator in millivolts
     * @returns 8 bit register value if no errors, -1 if error.
    */
    int mv2bits(int mV);
};

#endif /* _MAX14690_H_ */
