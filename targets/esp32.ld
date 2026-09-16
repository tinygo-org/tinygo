/* Linker script for the ESP32 (flash XIP)
 *
 * The ESP32 has 520KB of internal SRAM:
 *   - SRAM0 (192KB): 0x40070000-0x4009FFFF — upper 64KB used by ICache when
 *     flash XIP is active. Usable IRAM: 0x40080000 (128KB).
 *   - SRAM1 (128KB): 0x3FFE0000-0x3FFFFFFF — data only.
 *   - SRAM2 (200KB): 0x3FFAE000-0x3FFDFFFF — data only.
 *
 * Flash is memory-mapped via the cache:
 *   - DROM (read-only data): 0x3F400000, up to 4MB (64KB pages via MMU)
 *   - IROM (executable code): 0x400D0000, up to 4MB (64KB pages via MMU)
 * The startup code in IRAM configures the MMU to map flash pages.
 * The TinyGo builder places DROM/IROM content at page-aligned flash offsets.
 */

MEMORY
{
    /* SRAM2 usable DRAM (0x3FFB0000-0x3FFE0000, 192K): reserve the first 8KB
     * (0x3FFAE000-0x3FFB0000) for ROM PHY data (rom_phyFuns at 0x3FFAE0C0),
     * then use the rest for the stack, initialized .data, .bss, and the Go
     * GC heap. The .btdm_reserved section below keeps the first 64K free when
     * the Bluetooth driver is linked in. */
    DRAM  (rw) : ORIGIN = 0x3FFB0000, LENGTH = 192K

    /* Top of SRAM1 (0x3FFF0000-0x40000000, 64K): pool 7 + pool 6. This is the
     * ONLY part of SRAM1 that is safe for us to use:
     *   - 0x3FFE0000-0x3FFE0440 : ROM PRO data (ROM writes it) — avoid
     *   - 0x3FFE0440-0x3FFE8000 : ROM boot-stack area — NOT free during WiFi;
     *                             putting the arena here corrupts DMA and breaks
     *                             association ("auth expired") — avoid
     *   - 0x3FFE8000-0x3FFF0000 : pool 8, "used for MAC dump" by the WiFi MAC —
     *                             corrupts anything placed here — avoid
     *   - 0x3FFF0000-0x40000000 : pool 7/6 — safe (verified: arena + .bss both
     *                             work here). Used for the WiFi arena.
     * The arena (DMA buffers) owns this whole 64K block; .bss stays in SRAM2. */
    DRAM1 (rw) : ORIGIN = 0x3FFF0000, LENGTH = 64K

    IRAM  (x)  : ORIGIN = 0x40080000, LENGTH = 128K /* SRAM0 (usable portion) */

    DROM (r)  : ORIGIN = 0x3F400000, LENGTH = 4M /* Flash data bus (page-aligned) */
    IROM (rx) : ORIGIN = 0x400D0000, LENGTH = 4M /* Flash instruction bus (page-aligned) */
}

ENTRY(call_start_cpu0)

SECTIONS
{
    /* Constant global variables, stored in flash (DROM).
     * The builder places this at a page-aligned flash offset.
     */
    .rodata : ALIGN(4)
    {
        _rodata_start = ABSOLUTE(.);
        *(.rodata*)
        . = ALIGN (4);
        _rodata_end = ABSOLUTE(.);
    } >DROM

    /* The Bluetooth controller uses exchange memory at 0x3FFB0000 and BSS at
     * 0x3FFB8000. The addresses are fixed in ROM, so this 64K block must stay
     * free. The espradio Bluetooth driver defines _btdm_mem_needed. Programs
     * that do not use Bluetooth keep the 64K, because --gc-sections removes
     * the symbol with the driver.
     * See ESP-IDF components/bt/controller/esp32/bt.c, btdm_dram_available_region.
     */
    .btdm_reserved (NOLOAD) :
    {
        . = ALIGN(4);
        _btdm_reserved_start = ABSOLUTE(.);
        . += DEFINED(_btdm_mem_needed) ? 0x10000 : 0;
        _btdm_reserved_end = ABSOLUTE(.);
    } >DRAM
    ASSERT(_btdm_reserved_start == 0x3FFB0000, "BT reserved block must start at 0x3FFB0000")

    /* Put the stack at the bottom of DRAM, so that the application will
     * crash on stack overflow instead of silently corrupting memory.
     */
    .stack (NOLOAD) :
    {
        . = ALIGN(16);
        . += _stack_size;
        _stack_top = .;
    } >DRAM

    /* Global variables that are mutable and zero-initialized. Kept in SRAM2
     * (DRAM): the WiFi blob's .bss holds structures the MAC/DMA touches, and
     * SRAM1 has no room for both .bss and an adequately-sized arena in its only
     * safe region (pool 7/6, 64K) — so .bss stays in SRAM2 and the arena owns
     * the SRAM1 block. */
    .bss (NOLOAD) : ALIGN(4)
    {
        . = ALIGN (4);
        _sbss = ABSOLUTE(.);
        *(.bss .bss.*)
        . = ALIGN (4);
        _ebss = ABSOLUTE(.);
    } >DRAM

    /* Reserved zero-initialized storage for WiFi blob global variables.
     * On ESP32-S3 these live at fixed ROM DRAM addresses; on ESP32 the blob
     * doesn't export them so we allocate space here. */
    .wifi_bss (NOLOAD) : ALIGN(4)
    {
        _wifi_bss_start = ABSOLUTE(.);
        /* Function pointers patched by netif.c (6 × 4 bytes) */
        PROVIDE(g_config_func = ABSOLUTE(.));             . += 4;
        PROVIDE(g_net80211_tx_func = ABSOLUTE(.));        . += 4;
        PROVIDE(g_timer_func = ABSOLUTE(.));              . += 4;
        PROVIDE(s_michael_mic_failure_cb = ABSOLUTE(.));  . += 4;
        PROVIDE(g_tx_done_cb_func = ABSOLUTE(.));         . += 4;
        PROVIDE(s_encap_amsdu_func = ABSOLUTE(.));        . += 4;
        /* Pointer-sized globals used by pp/lmac (3 × 4 bytes) */
        PROVIDE(pp_wdev_funcs = ABSOLUTE(.));             . += 4;
        PROVIDE(our_tx_eb = ABSOLUTE(.));                 . += 4;
        PROVIDE(our_wait_eb = ABSOLUTE(.));               . += 4;
        PROVIDE(lmacConfMib_ptr = ABSOLUTE(.));           . += 4;
        . = ALIGN(4);
        _wifi_bss_end = ABSOLUTE(.);
    } >DRAM

    /* RX ring buffer: placed in DRAM1 (ahead of the arena) to free ~9.6 KB
     * of SRAM2 for the Go GC heap. espradio's netif.c tags the RX ring
     * with __attribute__((section(".rxring"))) on ESP32. */
    .rxring (NOLOAD) : ALIGN(16)
    {
        *(.rxring .rxring.*)
    } >DRAM1

    /* WiFi-only BSS: ISR tables, timer slots, and ISR ring buffers moved
     * from SRAM2 to DRAM1 to free ~2.8 KB for the Go GC heap. */
    .wifibss (NOLOAD) : ALIGN(4)
    {
        *(.wifibss .wifibss.*)
    } >DRAM1

    /* WiFi blob arena (DMA buffers): owns the remainder of DRAM1 (SRAM1
     * pool 7/6) after the RX ring. Outside the Go GC heap; DMA-capable;
     * isolated from GC objects. */
    .arena (NOLOAD) : ALIGN(16)
    {
        _arena_start = ABSOLUTE(.);
    } >DRAM1
    _arena_end = ORIGIN(DRAM1) + LENGTH(DRAM1);
    ASSERT((_arena_end - _arena_start) >= 32K, "arena region < 32K")

    /* Mutable global variables, initialized by the ROM bootloader. */
    .data : ALIGN(4)
    {
        . = ALIGN (4);
        _sdata = ABSOLUTE(.);
        *(.data .data.*)
        *(.dram*)
        . = ALIGN (4);
        /* DROM flash physical address, patched by the TinyGo ESP image builder
         * with the flash offset where the .rodata (DROM) segment was placed.
         * The startup code reads this to program the flash cache MMU. */
        _drom_flash_addr = ABSOLUTE(.);
        LONG(0x10000);
        . = ALIGN (4);
        _edata = ABSOLUTE(.);
    } >DRAM

    /* Check that the boot ROM stack (for the APP CPU) does not overlap with the
     * data that is loaded by the boot ROM.
     */
    ASSERT(_edata < 0x3ffe1320, "the .data section overlaps with the stack used by the boot ROM, possibly causing corruption at startup")

    /* IRAM segment: boot code, interrupt vectors, and WiFi/BLE blob code
     * that must run from RAM. */
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

        . = ALIGN(4);
        _iram_end = .;
    } >IRAM

    /* IROM segment: main code executed from flash via MMU cache.
     * The builder places this at a page-aligned flash offset.
     */
    .text : ALIGN(4)
    {
        _text_start = ABSOLUTE(.);
        *(.literal .text)
        *(.literal.* .text.*)
        *(.phyiram*)
        . = ALIGN(4);
        _text_end = ABSOLUTE(.);
    } >IROM

    /DISCARD/ :
    {
        *(.eh_frame)
    }
}

/* For the garbage collector. */
_globals_start = _sbss;
_globals_end = _edata;
_heap_start = _edata;
/* Go GC heap = the rest of SRAM2 above .data (~113KB), a single safe contiguous
 * block. The WiFi arena lives in DRAM1 (SRAM1 pool 7/6), so its ~64KB does not
 * consume Go heap and its DMA is isolated from GC objects. The heap never
 * touches SRAM1's ROM/MAC-dump regions (0x3FFE0000-0x3FFF0000, left unused). */
_heap_end = 0x3FFE0000;

_stack_size = 4K;

/* ESP32 ROM-owned PHY DRAM symbols from ESP-IDF esp32.rom.ld.
 * Exposed as linker symbols so firmware code can reference them without
 * hardcoding addresses in C sources. */
PROVIDE ( rom_phyFuns = 0x3ffae0c0 );
PROVIDE ( g_phyFuns_instance = 0x3ffae0c4 );

/* From ESP-IDF:
 * components/esp_rom/esp32/ld/esp32.rom.newlib-funcs.ld
 * This is the subset that is sometimes used by LLVM during codegen, and thus
 * must always be present.
 */
memcpy  = 0x4000c2c8;
memmove = 0x4000c3c0;
memset  = 0x4000c44c;

/* From ESP-IDF:
 * components/esp_rom/esp32/ld/esp32.rom.libgcc.ld
 * These are called from LLVM during codegen. The original license is Apache
 * 2.0, but I believe that a list of function names and addresses can't really
 * be copyrighted.
 */
__absvdi2      = 0x4006387c;
__absvsi2      = 0x40063868;
__adddf3       = 0x40002590;
__addsf3       = 0x400020e8;
__addvdi3      = 0x40002cbc;
__addvsi3      = 0x40002c98;
__ashldi3      = 0x4000c818;
__ashrdi3      = 0x4000c830;
__bswapdi2     = 0x40064b08;
__bswapsi2     = 0x40064ae0;
__clrsbdi2     = 0x40064b7c;
__clrsbsi2     = 0x40064b64;
__clzdi2       = 0x4000ca50;
__clzsi2       = 0x4000c7e8;
__cmpdi2       = 0x40063820;
__ctzdi2       = 0x4000ca64;
__ctzsi2       = 0x4000c7f0;
__divdc3       = 0x400645a4;
__divdf3       = 0x40002954;
__divdi3       = 0x4000ca84;
__divsi3       = 0x4000c7b8;
__eqdf2        = 0x400636a8;
__eqsf2        = 0x40063374;
__extendsfdf2  = 0x40002c34;
__ffsdi2       = 0x4000ca2c;
__ffssi2       = 0x4000c804;
__fixdfdi      = 0x40002ac4;
__fixdfsi      = 0x40002a78;
__fixsfdi      = 0x4000244c;
__fixsfsi      = 0x4000240c;
__fixunsdfsi   = 0x40002b30;
__fixunssfdi   = 0x40002504;
__fixunssfsi   = 0x400024ac;
__floatdidf    = 0x4000c988;
__floatdisf    = 0x4000c8c0;
__floatsidf    = 0x4000c944;
__floatsisf    = 0x4000c870;
__floatundidf  = 0x4000c978;
__floatundisf  = 0x4000c8b0;
__floatunsidf  = 0x4000c938;
__floatunsisf  = 0x4000c864;
__gcc_bcmp     = 0x40064a70;
__gedf2        = 0x40063768;
__gesf2        = 0x4006340c;
__gtdf2        = 0x400636dc;
__gtsf2        = 0x400633a0;
__ledf2        = 0x40063704;
__lesf2        = 0x400633c0;
__lshrdi3      = 0x4000c84c;
__ltdf2        = 0x40063790;
__ltsf2        = 0x4006342c;
__moddi3       = 0x4000cd4c;
__modsi3       = 0x4000c7c0;
__muldc3       = 0x40063c90;
__muldf3       = 0x4006358c;
__muldi3       = 0x4000c9fc;
__mulsf3       = 0x400632c8;
__mulsi3       = 0x4000c7b0;
__mulvdi3      = 0x40002d78;
__mulvsi3      = 0x40002d60;
__nedf2        = 0x400636a8;
__negdf2       = 0x400634a0;
__negdi2       = 0x4000ca14;
__negsf2       = 0x400020c0;
__negvdi2      = 0x40002e98;
__negvsi2      = 0x40002e78;
__nesf2        = 0x40063374;
__nsau_data    = 0x3ff96544;
__paritysi2    = 0x40002f3c;
__popcount_tab = 0x3ff96544;
__popcountdi2  = 0x40002ef8;
__popcountsi2  = 0x40002ed0;
__powidf2      = 0x400638e4;
__subdf3       = 0x400026e4;
__subsf3       = 0x400021d0;
__subvdi3      = 0x40002d20;
__subvsi3      = 0x40002cf8;
__truncdfsf2   = 0x40002b90;
__ucmpdi2      = 0x40063840;
__udiv_w_sdiv  = 0x40064bec;
__udivdi3      = 0x4000cff8;
__udivmoddi4   = 0x40064bf4;
__udivsi3      = 0x4000c7c8;
__umoddi3      = 0x4000d280;
__umodsi3      = 0x4000c7d0;
__umulsidi3    = 0x4000c7d8;
__unorddf2     = 0x400637f4;
__unordsf2     = 0x40063478;

/* From ESP-IDF:
 * components/esp_rom/esp32/ld/esp32.rom.ld
 * ROM functions needed for WiFi/PHY support.
 */
PROVIDE(ets_delay_us = 0x40008534);
PROVIDE(ets_printf = 0x40007d54);
PROVIDE(ets_install_putc1 = 0x40007d18);
PROVIDE(ets_install_uart_printf = 0x40007d28);
PROVIDE(intr_matrix_set = 0x4000681c);
PROVIDE(rtc_get_reset_reason = 0x400081d4);
PROVIDE(phy_get_romfuncs = 0x40004100);
PROVIDE(uart_tx_one_char = 0x40009200);
PROVIDE(uart_rx_one_char = 0x400092d0);
PROVIDE(ets_intr_lock = 0x400067b0);
PROVIDE(ets_intr_unlock = 0x400067c4);
PROVIDE(ets_isr_attach = 0x400067ec);
PROVIDE(ets_isr_mask = 0x400067fc);
PROVIDE(ets_isr_unmask = 0x40006808);
PROVIDE(roundup2 = 0x4000ab7c);

/* From ESP-IDF:
 * components/esp_rom/esp32/ld/esp32.rom.ld
 * Bluetooth controller ROM functions and data. libbtdm_app.a gets most of its
 * undefined symbols from here. The addresses are the same on all ESP32
 * silicon revisions.
 */
PROVIDE(AddBigHexModP256 = 0x40015c98);
PROVIDE(btdm_r_btdm_option_data_p_get = 0x40010004);
PROVIDE(btdm_r_btdm_rom_version_get = 0x40010078);
PROVIDE(btdm_r_import_rf_phy_func_p_get = 0x40054298);
PROVIDE(btdm_r_ip_func_p_get = 0x40019af0);
PROVIDE(btdm_r_modules_func_p_get = 0x4005427c);
PROVIDE(btdm_r_plf_func_p_set = 0x40054288);
PROVIDE(bt_util_buf_env = 0x3ffb8bd4);
PROVIDE(co_default_bdaddr = 0x3ffae704);
PROVIDE(co_null_bdaddr = 0x3ffb80e0);
PROVIDE(_data_start_btdm_rom = 0x4000d4f4);
PROVIDE(_data_start_btdm = 0x3ffae6e0);
PROVIDE(_data_end_btdm = 0x3ffaff10);
PROVIDE(_bss_start_btdm = 0x3ffb8000);
PROVIDE(_bss_end_btdm = 0x3ffbff70);
PROVIDE(dbg_state = 0x3ffb8d5d);
PROVIDE(ecc_env = 0x3ffb8d60);
PROVIDE(em_buf_env = 0x3ffb8d74);
PROVIDE(rwip_coex_cfg = 0x3ff9914c);
PROVIDE(rwip_priority = 0x3ff99159);
PROVIDE(hci_env = 0x3ffb9350);
PROVIDE(rwip_env = 0x3ffb8bcc);
PROVIDE(hci_fc_env = 0x3ffb9340);
PROVIDE(ke_env = 0x3ffb93cc);
PROVIDE(ke_handler_search = 0x4001a430);
PROVIDE(ke_task_env = 0x3ffb81d4);
PROVIDE(lb_default_state_tab_p_get = 0x4001c198);
PROVIDE(lb_env = 0x3ffb9424);
PROVIDE(lb_hci_cmd_handler_tab_p_get = 0x4001c18c);
PROVIDE(lc_default_state_tab_p_get = 0x4002f494);
PROVIDE(lc_env = 0x3ffb94ec);
PROVIDE(lc_hci_cmd_handler_tab_p_get = 0x4002f488);
PROVIDE(lc_state = 0x3ffb9508);
PROVIDE(ld_env = 0x3ffb9510);
PROVIDE(llc_default_state_tab_p_get = 0x40046058);
PROVIDE(llc_env = 0x3ffb96d0);
PROVIDE(llc_hci_cmd_handler_tab_p_get = 0x40042358);
PROVIDE(llcp_pdu_handler_tab_p_get = 0x40043f64);
PROVIDE(lld_evt_deferred_elt_push = 0x400466b4);
PROVIDE(lld_evt_deferred_elt_pop = 0x400466dc);
PROVIDE(lld_evt_slave_time_compute = 0x40046818);
PROVIDE(lld_evt_env = 0x3ffb9704);
PROVIDE(lld_evt_elt_wait_get = 0x400468e4);
PROVIDE(lld_evt_get_next_free_slot = 0x4004692c);
PROVIDE(lld_pdu_llcp_pk_desc_tab = 0x3ff98b68);
PROVIDE(lld_pdu_tx_flush_list = 0x4004a760);
PROVIDE(lld_pdu_pack = 0x4004ab14);
PROVIDE(LLM_AA_CT1 = 0x3ff98d8a);
PROVIDE(LLM_AA_CT2 = 0x3ff98d88);
PROVIDE(llm_default_state_tab_p_get = 0x4004e718);
PROVIDE(llm_hci_cmd_handler_tab_p_get = 0x4004c920);
PROVIDE(llm_le_env = 0x3ffb976c);
PROVIDE(lm_default_state_tab_p_get = 0x40054268);
PROVIDE(lm_env = 0x3ffb9860);
PROVIDE(lm_hci_cmd_handler_tab_p_get = 0x4005425c);
PROVIDE(MultiplyBigHexModP256 = 0x400160b8);
PROVIDE(one_bits = 0x3ff971f8);
PROVIDE(r_btdm_option_data = 0x3ffae6e0);
PROVIDE(r_bt_util_buf_acl_rx_alloc = 0x40010218);
PROVIDE(r_bt_util_buf_acl_rx_free = 0x40010234);
PROVIDE(r_bt_util_buf_acl_tx_alloc = 0x40010268);
PROVIDE(r_bt_util_buf_acl_tx_free = 0x40010280);
PROVIDE(r_bt_util_buf_lmp_tx_alloc = 0x400101d0);
PROVIDE(r_bt_util_buf_lmp_tx_free = 0x400101ec);
PROVIDE(r_bt_util_buf_sync_clear = 0x400103c8);
PROVIDE(r_bt_util_buf_sync_rx_alloc = 0x40010468);
PROVIDE(r_bt_util_buf_sync_rx_free = 0x4001049c);
PROVIDE(r_bt_util_buf_sync_tx_alloc = 0x400103ec);
PROVIDE(r_bt_util_buf_sync_tx_free = 0x40010428);
PROVIDE(lm_n192_mod_mul = 0x40011dc0);
PROVIDE(lm_n192_mod_add = 0x40011e9c);
PROVIDE(lm_n192_mod_sub = 0x40011eec);
PROVIDE(ea_conflict_check = 0x40014e9c);
PROVIDE(ea_prog_timer = 0x40014f88);
PROVIDE(r_ea_offset_req = 0x40015748);
PROVIDE(r_ea_time_get_halfslot_rounded = 0x40015894);
PROVIDE(ecc_point_multiplication_uint8_256 = 0x40016804);
PROVIDE(r_hci_look_for_cmd_desc = 0x40018454);
PROVIDE(r_hci_look_for_dbg_evt_desc = 0x400184c4);
PROVIDE(r_hci_send_2_host = 0x400185bc);
PROVIDE(r_ke_event_schedule = 0x40019cdc);
PROVIDE(r_ke_timer_clear = 0x4001ab90);
PROVIDE(r_ke_timer_set = 0x4001aac0);
PROVIDE(r_lb_init = 0x4001acd4);
PROVIDE(r_lb_reset = 0x4001ad38);
PROVIDE(r_lc_auth_cmp = 0x4001cd54);
PROVIDE(r_lc_init = 0x4001c948);
PROVIDE(r_lc_reset = 0x4001c99c);
PROVIDE(r_lc_sco_host_accept = 0x40022118);
PROVIDE(r_lc_start = 0x4001ca28);
PROVIDE(lc_set_encap_pdu_data_p192 = 0x4002e4c8);
PROVIDE(lc_set_encap_pdu_data_p256 = 0x4002e454);
PROVIDE(lm_get_auth_method = 0x40023420);
PROVIDE(lmp_accepted_ext_handler = 0x40027290);
PROVIDE(lmp_clk_adj_handler = 0x40027468);
PROVIDE(lmp_clk_adj_ack_handler = 0x400274f4);
PROVIDE(lmp_clk_adj_req_handler = 0x4002751c);
PROVIDE(lmp_feats_req_ext_handler = 0x4002ccb0);
PROVIDE(lmp_pkt_type_tbl_req_handler = 0x40027574);
PROVIDE(lmp_esco_link_req_handler = 0x40027610);
PROVIDE(lmp_rmv_esco_link_req_handler = 0x400276e8);
PROVIDE(lmp_ssr_req_handler = 0x4002780c);
PROVIDE(lmp_ssr_res_handler = 0x40027900);
PROVIDE(lmp_pause_enc_req_handler = 0x4002df90);
PROVIDE(lmp_resume_enc_req_handler = 0x4002e084);
PROVIDE(lmp_num_comparison_fail_handler = 0x40027a74);
PROVIDE(lmp_passkey_fail_handler = 0x40027aec);
PROVIDE(lmp_keypress_notif_handler = 0x4002c5c8);
PROVIDE(lmp_pwr_ctrl_req_handler = 0x400263bc);
PROVIDE(lmp_pwr_ctrl_res_handler = 0x40026480);
PROVIDE(lmp_auto_rate_handler = 0x40026548);
PROVIDE(lmp_pref_rate_handler = 0x4002657c);
PROVIDE(lmp_name_req_handler = 0x40025050);
PROVIDE(lmp_name_res_handler = 0x400250bc);
PROVIDE(lmp_accepted_handler = 0x4002e894);
PROVIDE(lmp_clk_off_req_handler = 0x40025a44);
PROVIDE(lmp_clk_off_res_handler = 0x40025ab8);
PROVIDE(lmp_detach_handler = 0x40025b74);
PROVIDE(lmp_tempkey_handler = 0x4002b6b0);
PROVIDE(lmp_temprand_handler = 0x4002b74c);
PROVIDE(lmp_sres_handler = 0x4002b840);
PROVIDE(lmp_aurand_handler = 0x4002bda0);
PROVIDE(lmp_unitkey_handler = 0x4002c13c);
PROVIDE(lmp_inrand_handler = 0x4002c414);
PROVIDE(lmp_oob_fail_handler = 0x40027b84);
PROVIDE(lmp_ping_req_handler = 0x40027c08);
PROVIDE(lmp_ping_res_handler = 0x40027c5c);
PROVIDE(lmp_stop_enc_req_handler = 0x4002de30);
PROVIDE(lmp_incr_pwr_req_handler = 0x4002629c);
PROVIDE(lmp_decr_pwr_req_handler = 0x400262f8);
PROVIDE(lmp_max_pwr_handler = 0x40026354);
PROVIDE(lmp_min_pwr_handler = 0x40026388);
PROVIDE(lmp_ver_req_handler = 0x400265f0);
PROVIDE(lmp_ver_res_handler = 0x40026670);
PROVIDE(lmp_qos_handler = 0x40026790);
PROVIDE(lmp_qos_req_handler = 0x40026844);
PROVIDE(lmp_sco_link_req_handler = 0x40026930);
PROVIDE(lmp_rmv_sco_link_req_handler = 0x40026a10);
PROVIDE(lmp_max_slot_handler = 0x40026a54);
PROVIDE(lmp_max_slot_req_handler = 0x40026aac);
PROVIDE(lmp_timing_accu_req_handler = 0x40026b54);
PROVIDE(lmp_timing_accu_res_handler = 0x40026bcc);
PROVIDE(lmp_setup_cmp_handler = 0x40026c84);
PROVIDE(lmp_feats_res_handler = 0x4002b548);
PROVIDE(lmp_feats_req_handler = 0x4002b620);
PROVIDE(lmp_host_con_req_handler = 0x4002b3d8);
PROVIDE(lmp_use_semi_perm_key_handler = 0x4002b4c4);
PROVIDE(lmp_slot_off_handler = 0x40026cc8);
PROVIDE(lmp_page_mode_req_handler = 0x40026d0c);
PROVIDE(lmp_page_scan_mode_req_handler = 0x40026d4c);
PROVIDE(lmp_supv_to_handler = 0x40026d94);
PROVIDE(lmp_test_activate_handler = 0x40026e7c);
PROVIDE(lmp_test_ctrl_handler = 0x40026ee4);
PROVIDE(lmp_enc_key_size_mask_req_handler = 0x40027038);
PROVIDE(lmp_enc_key_size_mask_res_handler = 0x400270a4);
PROVIDE(lmp_encaps_hdr_handler = 0x40027120);
PROVIDE(lmp_sp_nb_handler = 0x4002acf0);
PROVIDE(lmp_sp_cfm_handler = 0x4002b170);
PROVIDE(lmp_dhkey_chk_handler = 0x4002ab48);
PROVIDE(lmp_io_cap_res_handler = 0x4002c670);
PROVIDE(lmp_io_cap_req_handler = 0x4002c7a4);
PROVIDE(lc_cmd_cmp_bd_addr_send = 0x4002cec4);
PROVIDE(ld_acl_tx_packet_type_select = 0x4002fb40);
PROVIDE(ld_acl_sched = 0x40033268);
PROVIDE(ld_acl_sniff_sched = 0x4003340c);
PROVIDE(ld_acl_sniff_exit = 0x400312b4);
PROVIDE(ld_sco_evt_canceled_cbk = 0x40031e18);
PROVIDE(ld_acl_rx = 0x4003274c);
PROVIDE(ld_acl_tx = 0x4002ffdc);
PROVIDE(ld_acl_rx_sync = 0x4002fbec);
PROVIDE(ld_acl_rx_sync2 = 0x4002fd8c);
PROVIDE(ld_acl_rx_no_sync = 0x4002fe78);
PROVIDE(ld_acl_clk_isr = 0x40030cf8);
PROVIDE(ld_acl_rsw_frm_cbk = 0x40033bb0);
PROVIDE(ld_sco_modify = 0x40031778);
PROVIDE(lm_cmd_cmp_send = 0x40051838);
PROVIDE(ld_sco_frm_cbk = 0x400349dc);
PROVIDE(ld_sco_evt_start_cbk = 0x40031afc);
PROVIDE(ld_sco_evt_stop_cbk = 0x40031d78);
PROVIDE(ld_acl_rsw_evt_start_cbk = 0x40031154);
PROVIDE(ld_acl_sco_rsvd_check = 0x4002fa94);
PROVIDE(ld_acl_sniff_frm_cbk = 0x4003482c);
PROVIDE(ld_inq_end = 0x4003ab48);
PROVIDE(ld_inq_sched = 0x4003aba4);
PROVIDE(ld_inq_frm_cbk = 0x4003ae4c);
PROVIDE(ld_pscan_frm_cbk = 0x4003ebe4);
PROVIDE(r_ld_acl_sniff = 0x4003617c);
PROVIDE(ld_acl_frm_cbk = 0x40034414);
PROVIDE(ld_acl_rsw_end = 0x40032bc0);
PROVIDE(ld_acl_end = 0x40033140);
PROVIDE(ld_acl_resched = 0x40033814);
PROVIDE(ld_acl_test_mode_update = 0x40032050);
PROVIDE(ld_csb_tx_sched = 0x40039c1c);
PROVIDE(ld_csb_tx_evt_start_cbk = 0x40039d0c);
PROVIDE(ld_csb_tx_evt_canceled_cbk = 0x40039ee4);
PROVIDE(r_ld_init = 0x4003c294);
PROVIDE(r_ld_page_start = 0x4003d848);
PROVIDE(r_ld_reset = 0x4003c714);
PROVIDE(r_ld_sched_sniff_add = 0x4003f9c4);
PROVIDE(r_ld_sco_start = 0x40037110);
PROVIDE(r_ld_sscan_start = 0x40040384);
PROVIDE(r_llc_common_cmd_status_send = 0x40044ee0);
PROVIDE(r_llc_init = 0x40044778);
PROVIDE(r_llc_llcp_recv_handler = 0x40044678);
PROVIDE(r_llc_reset = 0x400447b8);
PROVIDE(r_llc_stop = 0x400449ac);
PROVIDE(r_lld_adv_stop = 0x40048ea0);
PROVIDE(r_lld_evt_init = 0x40046b3c);
PROVIDE(r_lld_pdu_tx_flush = 0x4004b414);
PROVIDE(r_lld_scan_stop = 0x40049190);
PROVIDE(r_llm_init = 0x4004c9f8);
PROVIDE(r_lm_init = 0x4004ed34);
PROVIDE(nvds_null_read = 0x400542a0);
PROVIDE(nvds_null_write = 0x400542a8);
PROVIDE(nvds_null_erase = 0x400542b0);
PROVIDE(nvds_read = 0x400542c4);
PROVIDE(nvds_write = 0x400542fc);
PROVIDE(nvds_erase = 0x40054334);
PROVIDE(nvds_init_memory = 0x40054358);
PROVIDE(r_rwip_prevent_sleep_set = 0x40055e64);
PROVIDE(r_rwip_reset = 0x40055ab8);
PROVIDE(r_rwip_schedule = 0x40055b38);
PROVIDE(rwip_rf = 0x3ffbdb28);
PROVIDE(SubtractBigHexMod256 = 0x40015e8c);
PROVIDE(sw_to_hw = 0x3ffb8d40);
PROVIDE(hci_le_rd_rem_used_feats_cmd_handler = 0x400417b4);
PROVIDE(hci_per_inq_mode_cmd_handler = 0x400519b0);
PROVIDE(llcp_length_req_handler = 0x40043808);
PROVIDE(llcp_unknown_rsp_handler = 0x40043ba8);
PROVIDE(llcp_channel_map_req_handler = 0x4004291c);
PROVIDE(llcp_con_up_req_handler = 0x400426f0);
PROVIDE(ets_update_cpu_frequency_rom = 0x40008550);
PROVIDE(hci_tl_env = 0x3ffb8154);
PROVIDE(ld_acl_env = 0x3ffb8258);
PROVIDE(ea_env = 0x3ffb80ec);
PROVIDE(lc_sco_data_path_config = 0x3ffb81f8);
PROVIDE(lc_sco_env = 0x3ffb81fc);
PROVIDE(ld_csb_tx_env = 0x3ffb827c);
PROVIDE(ld_fm_env = 0x3ffb8284);
PROVIDE(ld_inq_env = 0x3ffb82e4);
PROVIDE(ld_page_env = 0x3ffb82f0);
PROVIDE(ld_pscan_env = 0x3ffb8308);
PROVIDE(ld_sched_env = 0x3ffb830c);
PROVIDE(ld_sco_env = 0x3ffb824c);
PROVIDE(LM_Sniff = 0x3ffb8230);
PROVIDE(LM_SniffSubRate = 0x3ffb8214);
PROVIDE(nvds_env = 0x3ffb8364);
PROVIDE(nvds_magic_number = 0x3ff9912a);
PROVIDE(TASK_DESC_LLD = 0x3ff98b58);
PROVIDE(ld_acl_evt_canceled_cbk = 0x40033944);
PROVIDE(ld_acl_rsw_evt_canceled_cbk = 0x40033364);
PROVIDE(ld_acl_evt_stop_cbk = 0x40033870);
PROVIDE(ld_acl_evt_start_cbk = 0x40030ab0);
PROVIDE(ld_acl_sniff_evt_start_cbk = 0x40031360);
PROVIDE(lc_acl_disc_ind_handler = 0x4002f270);
PROVIDE(lc_pca_sscan_start_req_handler = 0x40029b34);
PROVIDE(ld_acl_sniff_trans_sched = 0x40033734);
PROVIDE(lc_pwr_decr_ind_handler = 0x4002859c);
PROVIDE(lc_pwr_incr_ind_handler = 0x400284a8);
PROVIDE(lc_setup_sync_param_check = 0x4002354c);
PROVIDE(lc_lmp_rsp_to_flow_spec_handler = 0x400297f0);
PROVIDE(lc_pca_sscan_clk_ind_handler = 0x4002a38c);
PROVIDE(lc_op_loc_unsniff_req_handler = 0x40028be0);
PROVIDE(lc_op_loc_sniff_req_handler = 0x40028ccc);
PROVIDE(lc_op_loc_switch_req_handler = 0x40028df4);
PROVIDE(lc_op_loc_sync_con_req_handler = 0x40028f6c);
PROVIDE(lm_sync_conf = 0x3ffb8348);
PROVIDE(lm_nb_sync_active = 0x3ffb8346);
PROVIDE(lm_sync_nego = 0x3ffb8345);
PROVIDE(lm_nego_cnt = 0x3ffb8344);
PROVIDE(lm_nego_cntl = 0x3ffb8342);
PROVIDE(lm_nego_max_cnt = 0x3ffb8343);
PROVIDE(lm_nego_pkt_used = 0x3ffb8340);

/* The Bluetooth controller startup code in libbtdm_app.a refers to the start of
 * the BSS. TinyGo calls this symbol _sbss.
 */
PROVIDE(_bss_start = _sbss);
