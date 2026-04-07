/* Linker script for the ESP32-S3 (flash XIP)
 *
 * The ESP32-S3 has 512KB of internal SRAM:
 *   - SRAM0 (32KB): 0x40370000-0x40377FFF — used by ICache when flash XIP is
 *     active, so we MUST NOT place code here.
 *   - SRAM1 (416KB): dual-mapped as IRAM 0x40378000-0x403DFFFF and
 *     DRAM 0x3FC88000-0x3FCEFFFF.
 *
 * Flash is memory-mapped via the cache:
 *   - DROM (read-only data): 0x3C000000, up to 32MB
 *   - IROM (executable code): 0x42000000, up to 32MB
 * The MMU uses 64KB pages, so the bottom 16 bits of the virtual address
 * and the flash offset must match. Dummy sections handle this alignment.
 */

MEMORY
{
    DRAM (rw) : ORIGIN = 0x3FC88000, LENGTH = 416K
    IRAM (x)  : ORIGIN = 0x40378000, LENGTH = 416K /* SRAM1 only (SRAM0 used by ICache) */

    DROM (r)  : ORIGIN = 0x3C000000, LENGTH = 32M /* Flash data bus (read-only) */
    IROM (rx) : ORIGIN = 0x42000000, LENGTH = 32M /* Flash instruction bus */
}

/* The entry point. It is set in the image flashed to the chip, so must be
 * defined.
 */
ENTRY(call_start_cpu0)

SECTIONS
{
    /* Dummy section so that .rodata starts right after the image header
     * and DROM segment header in the flash image.
     */
    .rodata_dummy (NOLOAD): ALIGN(4)
    {
        . += 0x18; /* esp_image_header_t at start of flash */
        . += 0x8;  /* DROM segment header (8 bytes) */
    } > DROM

    /* Constant global variables, stored in flash (DROM). */
    .rodata : ALIGN(4)
    {
        *(.rodata*)
        . = ALIGN (4);
        _drom_end = .;
    } >DROM

    /* Put the stack at the bottom of DRAM, so that the application will
     * crash on stack overflow instead of silently corrupting memory.
     */
    .stack (NOLOAD) :
    {
        . = ALIGN(16);
        . += _stack_size;
        _stack_top = .;
    } >DRAM

    /* Global variables that are mutable and zero-initialized. */
    .bss (NOLOAD) : ALIGN(4)
    {
        . = ALIGN (4);
        _sbss = ABSOLUTE(.);
        *(.bss .bss.*)
        . = ALIGN (4);
        _ebss = ABSOLUTE(.);
    } >DRAM

    /* Mutable global variables, initialized by the ROM bootloader. */
    .data : ALIGN(4)
    {
        . = ALIGN (4);
        _sdata = ABSOLUTE(.);
        *(.data .data.*)
        *(.dram*)
        . = ALIGN (4);
        _edata = ABSOLUTE(.);
    } >DRAM

    /* Dummy section to skip past stack+bss+data in IRAM (dual-mapped with DRAM). */
    .iram_dummy (NOLOAD): ALIGN(4)
    {
        . += SIZEOF(.stack);
        . += SIZEOF(.bss);
        . += SIZEOF(.data);
    } > IRAM

    /* IRAM segment: boot code, interrupt vectors, and any code that must
     * run from RAM. Loaded into SRAM by the ROM bootloader.
     */
    .iram : ALIGN(4)
    {
        /* Boot entry point and its literals */
        *(.literal.call_start_cpu0)
        *(.text.call_start_cpu0)

        /* Xtensa exception/interrupt vector table — must be 0x400-aligned */
        . = ALIGN(0x400);
        *(.text.exception_vectors)

        /* Level-1 interrupt handler */
        *(.literal._handle_level1)
        *(.text._handle_level1)

        /* WiFi/BLE blob IRAM sections */
        *(.iram*)
        *(.wifislprxiram*)
        *(.wifiextrairam*)
        *(.wifi0iram*)
        *(.wifislpiram*)
        *(.wifirxiram*)
        *(.wifiorslpiram*)
        *(.iram1*)
        *(.coexiram*)

        /* Init code — reclaimed for heap after startup */
        __init_start = .;
        *(.init)
        __init_end = .;
        . = ALIGN(4);
        _iram_end = .;
    } >IRAM

    /* Dummy section to put the IROM segment at the correct flash offset. */
    .text_dummy (NOLOAD): ALIGN(4)
    {
        . += 0x18;                                                  /* esp_image_header_t */
        . += SIZEOF(.rodata) + ((SIZEOF(.rodata) != 0) ? 0x8 : 0); /* DROM segment (optional) */
        . += SIZEOF(.data)   + ((SIZEOF(.data)   != 0) ? 0x8 : 0); /* DRAM segment (optional) */
        . += SIZEOF(.iram)   + 0x8;                                 /* IRAM segment */
        . += 0x8;                                                   /* IROM segment header */
    } > IROM

    /* IROM segment: main code executed from flash via cache. */
    .text : ALIGN(4)
    {
        *(.literal .text)
        *(.literal.* .text.*)
        _irom_end = .;
    } >IROM

    /DISCARD/ :
    {
        *(.eh_frame)
    }
}

/* For the garbage collector.
 * _heap_start must be after the DRAM shadow of the IRAM section.
 * IRAM and DRAM share the same physical SRAM1, with addresses offset by
 * (IRAM_ORIGIN - DRAM_ORIGIN) = 0x6F0000. We use _iram_end (the actual
 * end of .iram in IRAM space) converted to DRAM to avoid counting any
 * alignment gaps between .iram_dummy and .iram.
 * Init code at the end of .iram can be reclaimed for heap.
 */
_globals_start = _sbss;
_globals_end = _edata;
_heap_start = _iram_end - 0x6F0000 - (__init_end - __init_start);
/* Stop the heap before the ROM BSS/data region.
 * ESP-IDF defines SOC_ROM_STACK_START = 0x3fceb710 as the top of the ROM boot
 * stack.  Everything from 0x3fceb710 up to SOC_DIRAM_DRAM_HIGH (0x3fcf0000)
 * is ROM-reserved memory used by the WiFi/PHY firmware at runtime:
 *   - 0x3fceb710-0x3fcef3d3 : ROM BSS — internal firmware allocations
 *     (e.g. our_instances_ptr @ 0x3ff1ee58 holds 0x3fcef308)
 *   - 0x3fcef3d4-0x3fcef66c : PHY function table (g_phyFuns, 0x298 bytes)
 *   - 0x3fcef81c-0x3fcef954 : ROM global variables (phy_param_rom … pTxRx)
 *   - 0x3fcefe6d            : g_scan_forever
 * The Go GC stores metadata at [metadataStart, heapEnd) and will corrupt
 * any blob data that overlaps this range. */
_heap_end = 0x3fceb710;

_stack_size = 4K;

/* From ESP-IDF:
 * components/esp_rom/esp32/ld/esp32.rom.newlib-funcs.ld
 * This is the subset that is sometimes used by LLVM during codegen, and thus
 * must always be present.
 */
memset = 0x400011e8;
memcpy = 0x400011f4;
memmove = 0x40001200;
memcmp = 0x4000120c;

/* From ESP-IDF:
 * components/esp_rom/esp32/ld/esp32.rom.libgcc.ld
 * These are called from LLVM during codegen. The original license is Apache
 * 2.0, but I believe that a list of function names and addresses can't really
 * be copyrighted.
 */
__absvdi2 = 0x4000216c;
__absvsi2 = 0x40002178;
__adddf3 = 0x40002184;
__addsf3 = 0x40002190;
__addvdi3 = 0x4000219c;
__addvsi3 = 0x400021a8;
__ashldi3 = 0x400021b4;
__ashrdi3 = 0x400021c0;
__bswapdi2 = 0x400021cc;
__bswapsi2 = 0x400021d8;
__clear_cache = 0x400021e4;
__clrsbdi2 = 0x400021f0;
__clrsbsi2 = 0x400021fc;
__clzdi2 = 0x40002208;
__clzsi2 = 0x40002214;
__cmpdi2 = 0x40002220;
__ctzdi2 = 0x4000222c;
__ctzsi2 = 0x40002238;
__divdc3 = 0x40002244;
__divdf3 = 0x40002250;
__divdi3 = 0x4000225c;
__divsc3 = 0x40002268;
__divsf3 = 0x40002274;
__divsi3 = 0x40002280;
__eqdf2 = 0x4000228c;
__eqsf2 = 0x40002298;
__extendsfdf2 = 0x400022a4;
__ffsdi2 = 0x400022b0;
__ffssi2 = 0x400022bc;
__fixdfdi = 0x400022c8;
__fixdfsi = 0x400022d4;
__fixsfdi = 0x400022e0;
__fixsfsi = 0x400022ec;
__fixunsdfsi = 0x400022f8;
__fixunssfdi = 0x40002304;
__fixunssfsi = 0x40002310;
__floatdidf = 0x4000231c;
__floatdisf = 0x40002328;
__floatsidf = 0x40002334;
__floatsisf = 0x40002340;
__floatundidf = 0x4000234c;
__floatundisf = 0x40002358;
__floatunsidf = 0x40002364;
__floatunsisf = 0x40002370;
__gcc_bcmp = 0x4000237c;
__gedf2 = 0x40002388;
__gesf2 = 0x40002394;
__gtdf2 = 0x400023a0;
__gtsf2 = 0x400023ac;
__ledf2 = 0x400023b8;
__lesf2 = 0x400023c4;
__lshrdi3 = 0x400023d0;
__ltdf2 = 0x400023dc;
__ltsf2 = 0x400023e8;
__moddi3 = 0x400023f4;
__modsi3 = 0x40002400;
__muldc3 = 0x4000240c;
__muldf3 = 0x40002418;
__muldi3 = 0x40002424;
__mulsc3 = 0x40002430;
__mulsf3 = 0x4000243c;
__mulsi3 = 0x40002448;
__mulvdi3 = 0x40002454;
__mulvsi3 = 0x40002460;
__nedf2 = 0x4000246c;
__negdf2 = 0x40002478;
__negdi2 = 0x40002484;
__negsf2 = 0x40002490;
__negvdi2 = 0x4000249c;
__negvsi2 = 0x400024a8;
__nesf2 = 0x400024b4;
__paritysi2 = 0x400024c0;
__popcountdi2 = 0x400024cc;
__popcountsi2 = 0x400024d8;
__powidf2 = 0x400024e4;
__powisf2 = 0x400024f0;
__subdf3 = 0x400024fc;
__subsf3 = 0x40002508;
__subvdi3 = 0x40002514;
__subvsi3 = 0x40002520;
__truncdfsf2 = 0x4000252c;
__ucmpdi2 = 0x40002538;
__udivdi3 = 0x40002544;
__udivmoddi4 = 0x40002550;
__udivsi3 = 0x4000255c;
__udiv_w_sdiv = 0x40002568;
__umoddi3 = 0x40002574;
__umodsi3 = 0x40002580;
__unorddf2 = 0x4000258c;
__unordsf2 = 0x40002598;

/* From ESP-IDF:
 * components/esp_rom/esp32s3/ld/esp32s3.rom.ld
 * These are needed for wifi/BLE support and are available on the Apache 2.0
 * license.
 */

/***************************************
 Group common
 ***************************************/

/* Functions */
rtc_get_reset_reason = 0x4000057c;
analog_super_wdt_reset_happened = 0x40000588;
jtag_cpu_reset_happened = 0x40000594;
rtc_get_wakeup_cause = 0x400005a0;
rtc_select_apb_bridge = 0x400005ac;
rtc_unhold_all_pads = 0x400005b8;
ets_is_print_boot = 0x400005c4;
ets_printf = 0x400005d0;
ets_install_putc1 = 0x400005dc;
ets_install_uart_printf = 0x400005e8;
ets_install_putc2 = 0x400005f4;
PROVIDE( ets_delay_us = 0x40000600 );
ets_get_stack_info = 0x4000060c;
ets_install_lock = 0x40000618;
ets_backup_dma_copy = 0x40000624;
ets_apb_backup_init_lock_func = 0x40000630;
UartRxString = 0x4000063c;
uart_tx_one_char = 0x40000648;
uart_tx_one_char2 = 0x40000654;
uart_rx_one_char = 0x40000660;
uart_rx_one_char_block = 0x4000066c;
uart_rx_readbuff = 0x40000678;
uartAttach = 0x40000684;
uart_tx_flush = 0x40000690;
uart_tx_wait_idle = 0x4000069c;
uart_div_modify = 0x400006a8;
ets_write_char_uart = 0x400006b4;
uart_tx_switch = 0x400006c0;
multofup = 0x400006cc;
software_reset = 0x400006d8;
software_reset_cpu = 0x400006e4;
assist_debug_clock_enable = 0x400006f0;
assist_debug_record_enable = 0x400006fc;
clear_super_wdt_reset_flag = 0x40000708;
disable_default_watchdog = 0x40000714;

/* Interrupt handling */
intr_matrix_set = 0x40001b54;
ets_intr_lock = 0x40001b60;
ets_intr_unlock = 0x40001b6c;
ets_isr_attach = 0x40001b78;
ets_isr_unmask = 0x40001b90;


/***************************************
 Group rom_pp
 ***************************************/

/* Functions */
PROVIDE(esp_pp_rom_version_get = 0x40005250);
PROVIDE(lmacAdjustTimestamp = 0x40005304);
PROVIDE(lmacDiscardAgedMSDU = 0x40005310);
PROVIDE(lmacEndFrameExchangeSequence = 0x40005328);
PROVIDE(lmacMSDUAged = 0x4000534c);
PROVIDE(lmacProcessAllTxTimeout = 0x40005364);
PROVIDE(lmacProcessCollisions = 0x40005370);
PROVIDE(lmacProcessRxSucData = 0x4000537c);
PROVIDE(lmacProcessTxRts = 0x40005460);
PROVIDE(lmacProcessTxSuccess = 0x4000546c);
PROVIDE(lmacProcessTxTimeout = 0x40005478);
PROVIDE(lmacProcessTxComplete = 0x40005484);
PROVIDE(lmacReachLongLimit = 0x40005388);
PROVIDE(lmacReachShortLimit = 0x40005394);
PROVIDE(lmacRxDone = 0x400053ac);
PROVIDE(lmacSetAcked = 0x400054b4);
PROVIDE(lmacSetTxFrame = 0x400054c0);
PROVIDE(lmacTxDone = 0x400053c4);
PROVIDE(lmacTxFrame = 0x400054d8);
PROVIDE(ppAMPDU2Normal = 0x4000552c);
PROVIDE(ppAssembleAMPDU = 0x40005538);
PROVIDE(ppCalFrameTimes = 0x40005544);
PROVIDE(ppCalSubFrameLength = 0x40005550);
PROVIDE(ppCheckTxAMPDUlength = 0x40005568);
PROVIDE(ppDequeueRxq_Locked = 0x40005574);
PROVIDE(ppDequeueTxQ = 0x40005580);
PROVIDE(ppEmptyDelimiterLength = 0x4000558c);
PROVIDE(ppEnqueueRxq = 0x40005598);
PROVIDE(ppEnqueueTxDone = 0x400055a4);
PROVIDE(ppGetTxQFirstAvail_Locked = 0x400055b0);
PROVIDE(ppGetTxframe = 0x400055bc);
PROVIDE(ppProcessRxPktHdr = 0x400055e0);
PROVIDE(ppRecordBarRRC = 0x400055f8);
PROVIDE(lmacRequestTxopQueue = 0x40005604);
PROVIDE(lmacReleaseTxopQueue = 0x40005610);
PROVIDE(ppRecycleAmpdu = 0x4000561c);
PROVIDE(ppRecycleRxPkt = 0x40005628);
PROVIDE(ppResortTxAMPDU = 0x40005634);
PROVIDE(ppResumeTxAMPDU = 0x40005640);
PROVIDE(ppSearchTxQueue = 0x40005670);
PROVIDE(ppSearchTxframe = 0x4000567c);
PROVIDE(ppSelectNextQueue = 0x40005688);
PROVIDE(ppSubFromAMPDU = 0x40005694);
PROVIDE(ppTxPkt = 0x400056ac);
PROVIDE(ppTxProtoProc = 0x400056b8);
PROVIDE(ppTxqUpdateBitmap = 0x400056c4);
PROVIDE(pp_hdrsize = 0x400056dc);
PROVIDE(pp_post = 0x400056e8);
PROVIDE(pp_process_hmac_waiting_txq = 0x400056f4);
PROVIDE(rcGetAmpduSched = 0x40005700);
PROVIDE(rcUpdateRxDone = 0x4000570c);
PROVIDE(rc_get_trc = 0x40005718);
PROVIDE(rc_get_trc_by_index = 0x40005724);
PROVIDE(rcAmpduLowerRate = 0x40005730);
PROVIDE(rcampduuprate = 0x4000573c);
PROVIDE(rcClearCurAMPDUSched = 0x40005748);
PROVIDE(rcClearCurSched = 0x40005754);
PROVIDE(rcClearCurStat = 0x40005760);
PROVIDE(rcLowerSched = 0x40005778);
PROVIDE(rcSetTxAmpduLimit = 0x40005784);
PROVIDE(rcUpdateAckSnr = 0x4000579c);
PROVIDE(rcUpSched = 0x400057cc);
PROVIDE(rssi_margin = 0x400057d8);
PROVIDE(rx11NRate2AMPDULimit = 0x400057e4);
PROVIDE(TRC_AMPDU_PER_DOWN_THRESHOLD = 0x400057f0);
PROVIDE(TRC_AMPDU_PER_UP_THRESHOLD = 0x400057fc);
PROVIDE(trc_calc_duration = 0x40005808);
PROVIDE(trc_isTxAmpduOperational = 0x40005814);
PROVIDE(trc_onAmpduOp = 0x40005820);
PROVIDE(TRC_PER_IS_GOOD = 0x4000582c);
PROVIDE(trc_SetTxAmpduState = 0x40005838);
PROVIDE(trc_tid_isTxAmpduOperational = 0x40005844);
PROVIDE(trcAmpduSetState = 0x40005850);
PROVIDE(wDevCheckBlockError = 0x4000585c);
PROVIDE(wDev_DiscardFrame = 0x40005874);
PROVIDE(wDev_GetNoiseFloor = 0x40005880);
PROVIDE(wDev_IndicateAmpdu = 0x4000588c);
PROVIDE(wdev_bank_store = 0x400058a4);
PROVIDE(wdev_bank_load = 0x400058b0);
PROVIDE(wdev_mac_reg_load = 0x400058bc);
PROVIDE(wdev_mac_reg_store = 0x400058c8);
PROVIDE(wdev_mac_special_reg_load = 0x400058d4);
PROVIDE(wdev_mac_special_reg_store = 0x400058e0);
PROVIDE(wdev_mac_wakeup = 0x400058ec);
PROVIDE(wdev_mac_sleep = 0x400058f8);
PROVIDE(hal_mac_is_dma_enable = 0x40005904);
PROVIDE(wdevProcessRxSucDataAll = 0x40005928);
PROVIDE(wdev_csi_len_align = 0x40005934);
PROVIDE(ppDequeueTxDone_Locked = 0x40005940);
PROVIDE(config_is_cache_tx_buf_enabled = 0x40005964);
PROVIDE(ppProcessWaitingQueue = 0x4000597c);
PROVIDE(ppDisableQueue = 0x40005988);
PROVIDE(pm_allow_tx = 0x40005994);
PROVIDE(wdev_is_data_in_rxlist = 0x400059a0);
PROVIDE(ppProcTxCallback = 0x400059ac);
/* Data (.data, .bss, .rodata) */
our_instances_ptr = 0x3ff1ee58;
pTxRx = 0x3fcef954;
lmacConfMib_ptr = 0x3fcef950;
our_wait_eb = 0x3fcef94c;
our_tx_eb = 0x3fcef948;
pp_wdev_funcs = 0x3fcef944;
g_osi_funcs_p = 0x3fcef940;
wDevCtrl_ptr = 0x3fcef93c;
g_wdev_last_desc_reset_ptr = 0x3ff1ee54;
wDevMacSleep_ptr = 0x3fcef938;
g_lmac_cnt_ptr = 0x3fcef934;
our_controls_ptr = 0x3ff1ee50;
pp_sig_cnt_ptr = 0x3fcef930;
g_eb_list_desc_ptr = 0x3fcef92c;
s_fragment_ptr = 0x3fcef928;
if_ctrl_ptr = 0x3fcef924;
g_intr_lock_mux = 0x3fcef920;
g_wifi_global_lock = 0x3fcef91c;
s_wifi_queue = 0x3fcef918;
pp_task_hdl = 0x3fcef914;
s_pp_task_create_sem = 0x3fcef910;
s_pp_task_del_sem = 0x3fcef90c;
g_wifi_menuconfig_ptr = 0x3fcef908;
xphyQueue = 0x3fcef904;
ap_no_lr_ptr = 0x3fcef900;
rc11BSchedTbl_ptr = 0x3fcef8fc;
rc11NSchedTbl_ptr = 0x3fcef8f8;
rcLoRaSchedTbl_ptr = 0x3fcef8f4;
BasicOFDMSched_ptr = 0x3fcef8f0;
trc_ctl_ptr = 0x3fcef8ec;
g_pm_cnt_ptr = 0x3fcef8e8;
g_pm_ptr = 0x3fcef8e4;
g_pm_cfg_ptr = 0x3fcef8e0;
g_esp_mesh_quick_funcs_ptr = 0x3fcef8dc;
g_txop_queue_status_ptr = 0x3fcef8d8;
g_mac_sleep_en_ptr = 0x3fcef8d4;
g_mesh_is_root_ptr = 0x3fcef8d0;
g_mesh_topology_ptr = 0x3fcef8cc;
g_mesh_init_ps_type_ptr = 0x3fcef8c8;
g_mesh_is_started_ptr = 0x3fcef8c4;
g_config_func = 0x3fcef8c0;
g_net80211_tx_func = 0x3fcef8bc;
g_timer_func = 0x3fcef8b8;
s_michael_mic_failure_cb = 0x3fcef8b4;
wifi_sta_rx_probe_req = 0x3fcef8b0;
g_tx_done_cb_func = 0x3fcef8ac;
g_per_conn_trc = 0x3fcef860;
s_encap_amsdu_func = 0x3fcef85c;


/***************************************
 Group rom_net80211
 ***************************************/

/* Functions */
PROVIDE(esp_net80211_rom_version_get = 0x400059b8);
PROVIDE(ampdu_dispatch = 0x400059c4);
PROVIDE(ampdu_dispatch_all = 0x400059d0);
PROVIDE(ampdu_dispatch_as_many_as_possible = 0x400059dc);
PROVIDE(ampdu_dispatch_movement = 0x400059e8);
PROVIDE(ampdu_dispatch_upto = 0x400059f4);
PROVIDE(chm_is_at_home_channel = 0x40005a00);
PROVIDE(cnx_node_is_existing = 0x40005a0c);
PROVIDE(cnx_node_search = 0x40005a18);
PROVIDE(ic_ebuf_recycle_rx = 0x40005a24);
PROVIDE(ic_ebuf_recycle_tx = 0x40005a30);
PROVIDE(ic_reset_rx_ba = 0x40005a3c);
PROVIDE(ieee80211_align_eb = 0x40005a48);
PROVIDE(ieee80211_ampdu_start_age_timer = 0x40005a60);
PROVIDE(ieee80211_is_tx_allowed = 0x40005a78);
PROVIDE(ieee80211_output_pending_eb = 0x40005a84);
PROVIDE(ieee80211_set_tx_desc = 0x40005a9c);
PROVIDE(wifi_get_macaddr = 0x40005ab4);
PROVIDE(wifi_rf_phy_disable = 0x40005ac0);
PROVIDE(wifi_rf_phy_enable = 0x40005acc);
PROVIDE(ic_ebuf_alloc = 0x40005ad8);
PROVIDE(ieee80211_copy_eb_header = 0x40005af0);
PROVIDE(ieee80211_recycle_cache_eb = 0x40005afc);
PROVIDE(ieee80211_search_node = 0x40005b08);
PROVIDE(roundup2 = 0x40005b14);
PROVIDE(ieee80211_crypto_encap = 0x40005b20);
PROVIDE(ieee80211_set_tx_pti = 0x40005b44);
PROVIDE(wifi_is_started = 0x40005b50);
PROVIDE(ieee80211_gettid = 0x40005b5c);
/* Data (.data, .bss, .rodata) */
net80211_funcs = 0x3fcef858;
g_scan = 0x3fcef854;
g_chm = 0x3fcef850;
g_ic_ptr = 0x3fcef84c;
g_hmac_cnt_ptr = 0x3fcef848;
g_tx_cacheq_ptr = 0x3fcef844;
s_netstack_free = 0x3fcef840;
mesh_rxcb = 0x3fcef83c;
sta_rxcb = 0x3fcef838;


/***************************************
 Group rom_coexist
 ***************************************/

/* Functions */
PROVIDE(esp_coex_rom_version_get = 0x40005b68);
PROVIDE(coex_bt_release = 0x40005b74);
PROVIDE(coex_bt_request = 0x40005b80);
PROVIDE(coex_core_ble_conn_dyn_prio_get = 0x40005b8c);
PROVIDE(coex_core_pti_get = 0x40005ba4);
PROVIDE(coex_core_release = 0x40005bb0);
PROVIDE(coex_core_request = 0x40005bbc);
PROVIDE(coex_core_status_get = 0x40005bc8);
PROVIDE(coex_event_duration_get = 0x40005be0);
PROVIDE(coex_hw_timer_disable = 0x40005bec);
PROVIDE(coex_hw_timer_enable = 0x40005bf8);
PROVIDE(coex_hw_timer_set = 0x40005c04);
PROVIDE(coex_schm_interval_set = 0x40005c10);
PROVIDE(coex_schm_lock = 0x40005c1c);
PROVIDE(coex_schm_unlock = 0x40005c28);
PROVIDE(coex_wifi_release = 0x40005c40);
PROVIDE(esp_coex_ble_conn_dynamic_prio_get = 0x40005c4c);
/* Data (.data, .bss, .rodata) */
coex_env_ptr = 0x3fcef834;
coex_pti_tab_ptr = 0x3fcef830;
coex_schm_env_ptr = 0x3fcef82c;
coexist_funcs = 0x3fcef828;
g_coa_funcs_p = 0x3fcef824;
g_coex_param_ptr = 0x3fcef820;


/***************************************
 Group rom_phy
 ***************************************/

/* Functions */
PROVIDE(phy_get_romfuncs = 0x40005c58);
PROVIDE(rom_abs_temp = 0x40005c64);
PROVIDE(rom_bb_bss_cbw40_dig = 0x40005c70);
PROVIDE(rom_bb_wdg_test_en = 0x40005c7c);
PROVIDE(rom_bb_wdt_get_status = 0x40005c88);
PROVIDE(rom_bb_wdt_int_enable = 0x40005c94);
PROVIDE(rom_bb_wdt_rst_enable = 0x40005ca0);
PROVIDE(rom_bb_wdt_timeout_clear = 0x40005cac);
PROVIDE(rom_cbw2040_cfg = 0x40005cb8);
PROVIDE(rom_check_noise_floor = 0x40005cc4);
PROVIDE(rom_chip_i2c_readReg = 0x40005cd0);
PROVIDE(rom_chip_i2c_writeReg = 0x40005cdc);
PROVIDE(rom_dc_iq_est = 0x40005ce8);
PROVIDE(rom_disable_agc = 0x40005cf4);
PROVIDE(rom_en_pwdet = 0x40005d00);
PROVIDE(rom_enable_agc = 0x40005d0c);
PROVIDE(rom_get_bbgain_db = 0x40005d18);
PROVIDE(rom_get_data_sat = 0x40005d24);
PROVIDE(rom_get_i2c_read_mask = 0x40005d30);
PROVIDE(rom_get_pwctrl_correct = 0x40005d3c);
PROVIDE(rom_i2c_readReg = 0x40005d48);
PROVIDE(rom_i2c_readReg_Mask = 0x40005d54);
PROVIDE(rom_i2c_writeReg = 0x40005d60);
PROVIDE(rom_i2c_writeReg_Mask = 0x40005d6c);
PROVIDE(rom_index_to_txbbgain = 0x40005d78);
PROVIDE(rom_iq_est_disable = 0x40005d84);
PROVIDE(rom_iq_est_enable = 0x40005d90);
PROVIDE(rom_linear_to_db = 0x40005d9c);
PROVIDE(rom_loopback_mode_en = 0x40005da8);
PROVIDE(rom_mhz2ieee = 0x40005db4);
PROVIDE(rom_noise_floor_auto_set = 0x40005dc0);
PROVIDE(rom_pbus_debugmode = 0x40005dcc);
PROVIDE(rom_pbus_force_mode = 0x40005dd8);
PROVIDE(rom_pbus_force_test = 0x40005de4);
PROVIDE(rom_pbus_rd = 0x40005df0);
PROVIDE(rom_pbus_rd_addr = 0x40005dfc);
PROVIDE(rom_pbus_rd_shift = 0x40005e08);
PROVIDE(rom_pbus_set_dco = 0x40005e14);
PROVIDE(rom_pbus_set_rxgain = 0x40005e20);
PROVIDE(rom_pbus_workmode = 0x40005e2c);
PROVIDE(rom_pbus_xpd_rx_off = 0x40005e38);
PROVIDE(rom_pbus_xpd_rx_on = 0x40005e44);
PROVIDE(rom_pbus_xpd_tx_off = 0x40005e50);
PROVIDE(rom_pbus_xpd_tx_on = 0x40005e5c);
PROVIDE(rom_phy_byte_to_word = 0x40005e68);
PROVIDE(rom_phy_disable_cca = 0x40005e74);
PROVIDE(rom_phy_enable_cca = 0x40005e80);
PROVIDE(rom_phy_get_noisefloor = 0x40005e8c);
PROVIDE(rom_phy_get_rx_freq = 0x40005e98);
PROVIDE(rom_phy_set_bbfreq_init = 0x40005ea4);
PROVIDE(rom_pow_usr = 0x40005eb0);
PROVIDE(rom_pwdet_sar2_init = 0x40005ebc);
PROVIDE(rom_read_hw_noisefloor = 0x40005ec8);
PROVIDE(rom_read_sar_dout = 0x40005ed4);
PROVIDE(rom_set_cal_rxdc = 0x40005ee0);
PROVIDE(rom_set_chan_cal_interp = 0x40005eec);
PROVIDE(rom_set_loopback_gain = 0x40005ef8);
PROVIDE(rom_set_noise_floor = 0x40005f04);
PROVIDE(rom_set_rxclk_en = 0x40005f10);
PROVIDE(rom_set_tx_dig_gain = 0x40005f1c);
PROVIDE(rom_set_txcap_reg = 0x40005f28);
PROVIDE(rom_set_txclk_en = 0x40005f34);
PROVIDE(rom_spur_cal = 0x40005f40);
PROVIDE(rom_spur_reg_write_one_tone = 0x40005f4c);
PROVIDE(rom_target_power_add_backoff = 0x40005f58);
PROVIDE(rom_tx_pwctrl_bg_init = 0x40005f64);
PROVIDE(rom_txbbgain_to_index = 0x40005f70);
PROVIDE(rom_wifi_11g_rate_chg = 0x40005f7c);
PROVIDE(rom_write_gain_mem = 0x40005f88);
PROVIDE(chip728_phyrom_version = 0x40005f94);
PROVIDE(rom_disable_wifi_agc = 0x40005fa0);
PROVIDE(rom_enable_wifi_agc = 0x40005fac);
PROVIDE(rom_bt_index_to_bb = 0x40005fb8);
PROVIDE(rom_bt_bb_to_index = 0x40005fc4);
PROVIDE(rom_spur_coef_cfg = 0x40005fd0);
PROVIDE(rom_bb_bss_cbw40 = 0x40005fdc);
PROVIDE(rom_set_cca = 0x40005fe8);
PROVIDE(rom_tx_paon_set = 0x40005ff4);
PROVIDE(rom_i2cmst_reg_init = 0x40006000);
PROVIDE(rom_iq_corr_enable = 0x4000600c);
PROVIDE(rom_fe_reg_init = 0x40006018);
PROVIDE(rom_agc_reg_init = 0x40006024);
PROVIDE(rom_bb_reg_init = 0x40006030);
PROVIDE(rom_mac_enable_bb = 0x4000603c);
PROVIDE(rom_bb_wdg_cfg = 0x40006048);
PROVIDE(rom_force_txon = 0x40006054);
PROVIDE(rom_fe_txrx_reset = 0x40006060);
PROVIDE(rom_set_rx_comp = 0x4000606c);
PROVIDE(rom_set_pbus_reg = 0x40006078);
PROVIDE(rom_write_chan_freq = 0x40006084);
PROVIDE(rom_phy_xpd_rf = 0x40006090);
PROVIDE(rom_set_xpd_sar = 0x4000609c);
PROVIDE(rom_get_target_power_offset = 0x400060a8);
PROVIDE(rom_write_txrate_power_offset = 0x400060b4);
PROVIDE(rom_get_rate_fcc_index = 0x400060c0);
PROVIDE(rom_get_rate_target_power = 0x400060cc);
PROVIDE(rom_pkdet_vol_start = 0x400060d8);
PROVIDE(rom_read_sar2_code = 0x400060e4);
PROVIDE(rom_get_sar2_vol = 0x400060f0);
PROVIDE(rom_get_pll_vol = 0x400060fc);
PROVIDE(rom_get_phy_target_power = 0x40006108);
PROVIDE(rom_temp_to_power = 0x40006114);
PROVIDE(rom_phy_track_pll_cap = 0x40006120);
PROVIDE(rom_phy_pwdet_always_en = 0x4000612c);
PROVIDE(rom_phy_pwdet_onetime_en = 0x40006138);
PROVIDE(rom_get_i2c_mst0_mask = 0x40006144);
PROVIDE(rom_get_i2c_hostid = 0x40006150);
PROVIDE(rom_enter_critical_phy = 0x4000615c);
PROVIDE(rom_exit_critical_phy = 0x40006168);
PROVIDE(rom_chip_i2c_readReg_org = 0x40006174);
PROVIDE(rom_i2c_paral_set_mst0 = 0x40006180);
PROVIDE(rom_i2c_paral_set_read = 0x4000618c);
PROVIDE(rom_i2c_paral_read = 0x40006198);
PROVIDE(rom_i2c_paral_write = 0x400061a4);
PROVIDE(rom_i2c_paral_write_num = 0x400061b0);
PROVIDE(rom_i2c_paral_write_mask = 0x400061bc);
PROVIDE(rom_bb_bss_cbw40_ana = 0x400061c8);
PROVIDE(rom_chan_to_freq = 0x400061d4);
PROVIDE(rom_open_i2c_xpd = 0x400061e0);
PROVIDE(rom_dac_rate_set = 0x400061ec);
PROVIDE(rom_tsens_read_init = 0x400061f8);
PROVIDE(rom_tsens_code_read = 0x40006204);
PROVIDE(rom_tsens_index_to_dac = 0x40006210);
PROVIDE(rom_tsens_index_to_offset = 0x4000621c);
PROVIDE(rom_tsens_dac_cal = 0x40006228);
PROVIDE(rom_code_to_temp = 0x40006234);
PROVIDE(rom_write_pll_cap_mem = 0x40006240);
PROVIDE(rom_pll_correct_dcap = 0x4000624c);
PROVIDE(rom_phy_en_hw_set_freq = 0x40006258);
PROVIDE(rom_phy_dis_hw_set_freq = 0x40006264);
PROVIDE(rom_pll_vol_cal = 0x40006270);
PROVIDE(rom_wrtie_pll_cap = 0x4000627c);
PROVIDE(rom_set_tx_gain_mem = 0x40006288);
PROVIDE(rom_bt_tx_dig_gain = 0x40006294);
PROVIDE(rom_bt_get_tx_gain = 0x400062a0);
PROVIDE(rom_get_chan_target_power = 0x400062ac);
PROVIDE(rom_get_tx_gain_value = 0x400062b8);
PROVIDE(rom_wifi_tx_dig_gain = 0x400062c4);
PROVIDE(rom_wifi_get_tx_gain = 0x400062d0);
PROVIDE(rom_fe_i2c_reg_renew = 0x400062dc);
PROVIDE(rom_wifi_agc_sat_gain = 0x400062e8);
PROVIDE(rom_i2c_master_reset = 0x400062f4);
PROVIDE(rom_bt_filter_reg = 0x40006300);
PROVIDE(rom_phy_bbpll_cal = 0x4000630c);
PROVIDE(rom_i2c_sar2_init_code = 0x40006318);
PROVIDE(rom_phy_param_addr = 0x40006324);
PROVIDE(rom_phy_reg_init = 0x40006330);
PROVIDE(rom_set_chan_reg = 0x4000633c);
PROVIDE(rom_phy_wakeup_init = 0x40006348);
PROVIDE(rom_phy_i2c_init1 = 0x40006354);
PROVIDE(rom_tsens_temp_read = 0x40006360);
PROVIDE(rom_bt_track_pll_cap = 0x4000636c);
PROVIDE(rom_wifi_track_pll_cap = 0x40006378);
PROVIDE(rom_wifi_set_tx_gain = 0x40006384);
PROVIDE(rom_txpwr_cal_track = 0x40006390);
PROVIDE(rom_tx_pwctrl_background = 0x4000639c);
PROVIDE(rom_bt_set_tx_gain = 0x400063a8);
PROVIDE(rom_noise_check_loop = 0x400063b4);
PROVIDE(rom_phy_close_rf = 0x400063c0);
PROVIDE(rom_phy_xpd_tsens = 0x400063cc);
PROVIDE(rom_phy_freq_mem_backup = 0x400063d8);
PROVIDE(rom_phy_ant_init = 0x400063e4);
PROVIDE(rom_bt_track_tx_power = 0x400063f0);
PROVIDE(rom_wifi_track_tx_power = 0x400063fc);
PROVIDE(rom_phy_dig_reg_backup = 0x40006408);
PROVIDE(chip728_phyrom_version_num = 0x40006414);
PROVIDE(rom_mac_tx_chan_offset = 0x40006420);
PROVIDE(rom_rx_gain_force = 0x4000642c);
/* Data (.data, .bss, .rodata) */
phy_param_rom = 0x3fcef81c;

/* Other ROM symbols */
g_scan_forever = 0x3fcefe6d;
