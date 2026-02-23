# ESP32-S3 ADC — сверка с ESP-IDF

Чтобы уточнить порядок регистров и убедиться, что железо работает:

1. **Проверить железо через IDF**
   - Клонировать esp-idf, собрать пример:
     - `idf.py set-target esp32s3`
     - `idf.py -C examples/peripherals/adc/oneshot_read build flash monitor`
   - В примере используется `adc_oneshot_read()`; если там есть осмысленные значения — ADC и пины в порядке.

2. **Где смотреть порядок регистров в IDF**
   - Драйвер: `components/esp_adc/adc_oneshot.c` — вызовы `adc_oneshot_hal_setup`, `adc_oneshot_hal_convert`.
   - HAL (часто в отдельном репозитории **esp-hal**): `adc_oneshot_hal.c`, `adc_ll.c` — работа с APB_SARADC (pattern table, START, INT_RAW, DATA_STATUS).
   - Для S3 oneshot использует APB_SARADC; перед конвертацией вызывается `adc_apb_periph_claim()` и `ANALOG_CLOCK_ENABLE()`.

3. **Полезные документы**
   - [ESP32-S3 Technical Reference Manual](https://www.espressif.com/sites/default/files/documentation/esp32-s3_technical_reference_manual_en.pdf) — раздел SAR ADC, регистры APB_SARADC, формат PATT_TAB и PATT_LEN.

4. **В TinyGo**
   - Инициализация: RTC_CNTL/SENS питание, FSM_WAIT, CLKM_CONF.
   - Один замер: SENS (atten, EN_PAD) + APB (pattern 1 слот, PATT_P_CLEAR, PATT_LEN=1, START, ожидание INT_RAW DONE, чтение DATA_STATUS). Для ADC2 — ARB_CTRL APB_FORCE/GRANT_FORCE.
