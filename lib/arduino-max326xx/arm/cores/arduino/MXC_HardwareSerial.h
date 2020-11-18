/* ******************************************************************************
 * Copyright (C) 2017 Maxim Integrated Products, Inc., All Rights Reserved.
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
 **************************************************************************** */

#ifndef MXC_HardwareSerial_h
#define MXC_HardwareSerial_h

#include <inttypes.h>
#include "uart_regs.h"
#include "uart.h"
#include "HardwareSerial.h"

// Define constants and variables for buffering incoming serial data.  We're
// using a ring buffer (I think), in which head is the index of the location
// to which to write the next incoming character and tail is the index of the
// location from which to read.
// NOTE: a "power of 2" buffer size is reccomended to dramatically
//       optimize all the modulo operations for ring buffers.
// WARNING: When buffer sizes are increased to > 256, the buffer index
// variables are automatically increased in size, but the extra
// atomicity guards needed for that are not implemented. This will
// often work, but occasionally a race condition can occur that makes
// Serial behave erratically. See https://github.com/arduino/Arduino/issues/2405

#if !defined(SERIAL_TX_BUFFER_SIZE)
#if ((RAMEND - RAMSTART) < 1023)
#define SERIAL_TX_BUFFER_SIZE 16
#else
#define SERIAL_TX_BUFFER_SIZE 64
#endif
#endif
#if !defined(SERIAL_RX_BUFFER_SIZE)
#if ((RAMEND - RAMSTART) < 1023)
#define SERIAL_RX_BUFFER_SIZE 16
#else
#define SERIAL_RX_BUFFER_SIZE 64
#endif
#endif
#if (SERIAL_TX_BUFFER_SIZE>256)
typedef uint16_t tx_buffer_index_t;
#else
typedef uint8_t tx_buffer_index_t;
#endif
#if  (SERIAL_RX_BUFFER_SIZE>256)
typedef uint16_t rx_buffer_index_t;
#else
typedef uint8_t rx_buffer_index_t;
#endif

/*
Define config for Serial.begin(baud, config);
Encoding is: bbbbxpps
where, bbbb = number of data bits; 5, 6, 7, 8
         pp = parity; 0-none, 1-odd, 2-even
          s = stop bits; 0-one, 1-two
          x = unused
*/

#define SERIAL_5N1 0x50
#define SERIAL_6N1 0x60
#define SERIAL_7N1 0x70
#define SERIAL_8N1 0x80
#define SERIAL_5N2 0x51
#define SERIAL_6N2 0x61
#define SERIAL_7N2 0x71
#define SERIAL_8N2 0x81
#define SERIAL_5E1 0x54
#define SERIAL_6E1 0x64
#define SERIAL_7E1 0x74
#define SERIAL_8E1 0x84
#define SERIAL_5E2 0x55
#define SERIAL_6E2 0x65
#define SERIAL_7E2 0x75
#define SERIAL_8E2 0x85
#define SERIAL_5O1 0x52
#define SERIAL_6O1 0x62
#define SERIAL_7O1 0x72
#define SERIAL_8O1 0x82
#define SERIAL_5O2 0x53
#define SERIAL_6O2 0x63
#define SERIAL_7O2 0x73
#define SERIAL_8O2 0x83

class MXC_HardwareSerial : public HardwareSerial
{
  protected:
    mxc_uart_regs_t *_uart;
    mxc_uart_fifo_regs_t *_fifo;
    IRQn_Type _irqn;

    volatile rx_buffer_index_t _rx_buffer_head;
    volatile rx_buffer_index_t _rx_buffer_tail;
    volatile tx_buffer_index_t _tx_buffer_head;
    volatile tx_buffer_index_t _tx_buffer_tail;

    unsigned char _rx_buffer[SERIAL_RX_BUFFER_SIZE];
    unsigned char _tx_buffer[SERIAL_TX_BUFFER_SIZE];

  public:
    inline MXC_HardwareSerial(uint32_t port);
    void begin(unsigned long baud) { begin(baud, SERIAL_8N1); }
    void begin(unsigned long, uint8_t);
    void end();
    int available(void);
    int peek(void);
    int read(void);
    int availableForWrite(void);
    void flush(void);
    size_t write(uint8_t n);
    using Print::write; // pull in write(str) and write(buf, size) from Print
    // Interrupt handlers - Not intended to be called externally
    inline void _handler(void);

  private:
    inline void _rx_handler(void);
    inline void _tx_handler(void);
};

extern MXC_HardwareSerial Serial0;
extern MXC_HardwareSerial Serial1;
extern MXC_HardwareSerial Serial2;
extern void serialEventRun(void);

#endif // MXC_HardwareSerial_h
