// -----------------------------------------------------------------------------
//  file: teensy40.s
//  desc: various startup and configuration data for Teensy 4.0.
// -----------------------------------------------------------------------------
//  References
//    i.MX RT1060 Processor Reference Manual
//       - Section 9.7.1 "Image Vector Table and Boot Data"
//    Teensyduino 1.53 by Paul Stoffregen (PJRC)
//       - cores/teensy4/bootdata.c
//       - cores/teensy4/startup.c
// -----------------------------------------------------------------------------

.section .boot_data
.global  __boot_data
__boot_data:
    .word 0x60000000         // boot start location
    .word _image_size        // flash size
    .word 0                  // plugin flag, use 0 to indicate normal (non-plugin) ROM image

.section .ivt
.global  __ivt
__ivt:
    .word 0x402000D1         // header (version 4.0)
    .word _svectors          // image entry function
    .word 0                  // reserved
    .word 0                  // DCD info (optional, set to 0|NULL if unused)
    .word __boot_data        // boot data struct
    .word __ivt              // self
    .word 0                  // command sequence file (CSF) not provided in image
    .word 0                  // reserved

.section .flash_config
.global  __flash_config
__flash_config:
    // 448 byte common FlexSPI configuration block, 8.6.3.1 page 223 (RT1060 rev 0)
    // MCU_Flashloader_Reference_Manual.pdf, 8.2.1, Table 8-2, page 72-75
    .word 0x42464346         // Tag                                       0x00
    .word 0x56010000         // Version
    .word 0                  // reserved
    .word 0x00020101         // columnAdressWidth,dataSetupTime,dataHoldTime,readSampleClkSrc

    .word 0x00000000         // waitTimeCfgCommands,-,deviceModeCfgEnable
    .word 0                  // deviceModeSeq
    .word 0                  // deviceModeArg
    .word 0x00000000         // -,-,-,configCmdEnable

    .word 0                  // configCmdSeqs                             0x20
    .word 0
    .word 0
    .word 0

    .word 0                  // cfgCmdArgs                                0x30
    .word 0
    .word 0
    .word 0

    .word 0x00000000         // controllerMiscOption                      0x40
    .word 0x00030401         // lutCustomSeqEnable,serialClkFreq,sflashPadType,deviceType
    .word 0                  // reserved
    .word 0                  // reserved

    .word 0x00200000         // sflashA1Size (Teensy 4.0)                 0x50
    //.word 0x00800000         // sflashA1Size (Teensy 4.1)                 0x50

    .word 0                  // sflashA2Size
    .word 0                  // sflashB1Size
    .word 0                  // sflashB2Size

    .word 0                  // csPadSettingOverride                      0x60
    .word 0                  // sclkPadSettingOverride
    .word 0                  // dataPadSettingOverride
    .word 0                  // dqsPadSettingOverride

    .word 0                  // timeoutInMs                               0x70
    .word 0                  // commandInterval
    .word 0                  // dataValidTime
    .word 0x00000000         // busyBitPolarity,busyOffset

    .word 0x0A1804EB         // lookupTable[0]                            0x80
    .word 0x26043206         // lookupTable[1]
    .word 0                  // lookupTable[2]
    .word 0                  // lookupTable[3]

    .word 0x24040405         // lookupTable[4]                            0x90
    .word 0                  // lookupTable[5]
    .word 0                  // lookupTable[6]
    .word 0                  // lookupTable[7]

    .word 0                  // lookupTable[8]                            0xA0
    .word 0                  // lookupTable[9]
    .word 0                  // lookupTable[10]
    .word 0                  // lookupTable[11]

    .word 0x00000406         // lookupTable[12]                           0xB0
    .word 0                  // lookupTable[13]
    .word 0                  // lookupTable[14]
    .word 0                  // lookupTable[15]

    .word 0                  // lookupTable[16]                           0xC0
    .word 0                  // lookupTable[17]
    .word 0                  // lookupTable[18]
    .word 0                  // lookupTable[19]

    .word 0x08180420         // lookupTable[20]                           0xD0
    .word 0                  // lookupTable[21]
    .word 0                  // lookupTable[22]
    .word 0                  // lookupTable[23]

    .word 0                  // lookupTable[24]                           0xE0
    .word 0                  // lookupTable[25]
    .word 0                  // lookupTable[26]
    .word 0                  // lookupTable[27]

    .word 0                  // lookupTable[28]                           0xF0
    .word 0                  // lookupTable[29]
    .word 0                  // lookupTable[30]
    .word 0                  // lookupTable[31]

    .word 0x081804D8         // lookupTable[32]                           0x100
    .word 0                  // lookupTable[33]
    .word 0                  // lookupTable[34]
    .word 0                  // lookupTable[35]

    .word 0x08180402         // lookupTable[36]                           0x110
    .word 0x00002004         // lookupTable[37]
    .word 0                  // lookupTable[38]
    .word 0                  // lookupTable[39]

    .word 0                  // lookupTable[40]                           0x120
    .word 0                  // lookupTable[41]
    .word 0                  // lookupTable[42]
    .word 0                  // lookupTable[43]

    .word 0x00000460         // lookupTable[44]                           0x130
    .word 0                  // lookupTable[45]
    .word 0                  // lookupTable[46]
    .word 0                  // lookupTable[47]

    .word 0                  // lookupTable[48]                           0x140
    .word 0                  // lookupTable[49]
    .word 0                  // lookupTable[50]
    .word 0                  // lookupTable[51]

    .word 0                  // lookupTable[52]                           0x150
    .word 0                  // lookupTable[53]
    .word 0                  // lookupTable[54]
    .word 0                  // lookupTable[55]

    .word 0                  // lookupTable[56]                           0x160
    .word 0                  // lookupTable[57]
    .word 0                  // lookupTable[58]
    .word 0                  // lookupTable[59]

    .word 0                  // lookupTable[60]                           0x170
    .word 0                  // lookupTable[61]
    .word 0                  // lookupTable[62]
    .word 0                  // lookupTable[63]

    .word 0                  // LUT 0: Read                               0x180
    .word 0                  // LUT 1: ReadStatus
    .word 0                  // LUT 3: WriteEnable
    .word 0                  // LUT 5: EraseSector

    .word 0                  // LUT 9: PageProgram                        0x190
    .word 0                  // LUT 11: ChipErase
    .word 0                  // LUT 15: Dummy
    .word 0                  // LUT unused?

    .word 0                  // LUT unused?                               0x1A0
    .word 0                  // LUT unused?
    .word 0                  // LUT unused?
    .word 0                  // LUT unused?

    .word 0                  // reserved                                  0x1B0
    .word 0                  // reserved
    .word 0                  // reserved
    .word 0                  // reserved

    // 64 byte Serial NOR configuration block (8.6.3.2, page 346)

    .word 256                // pageSize                                  0x1C0
    .word 4096               // sectorSize
    .word 1                  // ipCmdSerialClkFreq
    .word 0                  // reserved

    .word 0x00010000         // block size                                0x1D0
    .word 0                  // reserved
    .word 0                  // reserved
    .word 0                  // reserved

    .word 0                  // reserved                                  0x1E0
    .word 0                  // reserved
    .word 0                  // reserved
    .word 0                  // reserved

    .word 0                  // reserved                                  0x1F0
    .word 0                  // reserved
    .word 0                  // reserved
    .word 0                  // reserved
