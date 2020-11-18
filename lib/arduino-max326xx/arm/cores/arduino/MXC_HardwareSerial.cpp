/*
  MXC_HardwareSerial.cpp - Hardware serial library for Wiring
  copied from HardwareSerial.cpp

  HardwareSerial.cpp - Hardware serial library for Wiring
  Copyright (c) 2006 Nicholas Zambetti.  All right reserved.

  This library is free software; you can redistribute it and/or
  modify it under the terms of the GNU Lesser General Public
  License as published by the Free Software Foundation; either
  version 2.1 of the License, or (at your option) any later version.

  This library is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
  Lesser General Public License for more details.

  You should have received a copy of the GNU Lesser General Public
  License along with this library; if not, write to the Free Software
  Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA

  Modified 23 November 2006 by David A. Mellis
  Modified 28 September 2010 by Mark Sproul
  Modified 14 August 2012 by Alarus
  Modified 3 December 2013 by Matthijs Kooijman
  Modified 2017 by Maxim Integrated for MAX326xx
*/

#include "Arduino.h"
#include "MXC_HardwareSerial.h"
#include "nvic_table.h"

#define UART_ERRORS     (MXC_F_UART_INTEN_RX_FIFO_OVERFLOW |  \
                         MXC_F_UART_INTEN_RX_FRAMING_ERR |    \
                         MXC_F_UART_INTEN_RX_PARITY_ERR)

#define UART_READ_INTS  (MXC_F_UART_INTEN_RX_FIFO_AF |        \
                         MXC_F_UART_INTEN_RX_FIFO_NOT_EMPTY | \
                         MXC_F_UART_INTEN_RX_STALLED |        \
                         UART_ERRORS)

#define UART_WRITE_INTS (MXC_F_UART_INTEN_TX_UNSTALLED |      \
                         MXC_F_UART_INTEN_TX_FIFO_AE)

#define EXTRA_STOP_MASK 0x01
#define PARITY_MASK     0x06
#define BIT_COUNT_MASK  0xF0

MXC_HardwareSerial::MXC_HardwareSerial(uint32_t idx) :
  _uart(MXC_UART_GET_UART(idx)),
  _fifo(MXC_UART_GET_FIFO(idx)),
  _irqn(MXC_UART_GET_IRQ(idx)),
  _rx_buffer_head(0), _rx_buffer_tail(0),
  _tx_buffer_head(0), _tx_buffer_tail(0)
{
}

void MXC_HardwareSerial::begin(unsigned long baud, uint8_t config)
{
  uart_cfg_t uart_cfg = {
    .extra_stop = (uint8_t)((config & EXTRA_STOP_MASK) ? 1 : 0),
    .cts = 0,
    .rts = 0,
    .baud = baud,
    .size = ((config & BIT_COUNT_MASK) == 0x50) ? UART_DATA_SIZE_5_BITS :
            ((config & BIT_COUNT_MASK) == 0x60) ? UART_DATA_SIZE_6_BITS :
            ((config & BIT_COUNT_MASK) == 0x70) ? UART_DATA_SIZE_7_BITS : UART_DATA_SIZE_8_BITS,
    .parity = ((config & PARITY_MASK) == 0x02) ? UART_PARITY_ODD : ((config & PARITY_MASK) == 0x04) ? UART_PARITY_EVEN : UART_PARITY_DISABLE
  };

  sys_cfg_uart_t sys_cfg = {
    .clk_scale = CLKMAN_SCALE_AUTO,
    .io_cfg = IOMAN_UART(MXC_UART_GET_IDX(_uart), IOMAN_MAP_A, IOMAN_MAP_UNUSED, IOMAN_MAP_UNUSED, 1, 0, 0)
  };

  if (UART_Init(_uart, &uart_cfg, &sys_cfg) != E_NO_ERROR) {
    // Fail silently
    return;
  }

  NVIC_EnableIRQ(_irqn);

  // Clear pending interrupts prior to enabling interrupts
  _uart->intfl = _uart->intfl;
  _uart->inten |= UART_READ_INTS;

  // Interrupt when there is one byte left in the TXFIFO
  _uart->tx_fifo_ctrl = ((MXC_UART_FIFO_DEPTH - 1) << MXC_F_UART_TX_FIFO_CTRL_FIFO_AE_LVL_POS);
  _uart->inten |= UART_WRITE_INTS;
}

void MXC_HardwareSerial::end()
{
  UART_Shutdown(_uart);
}

int MXC_HardwareSerial::available(void)
{
  return ((unsigned int)(SERIAL_RX_BUFFER_SIZE + _rx_buffer_head - _rx_buffer_tail)) % SERIAL_RX_BUFFER_SIZE;
}

int MXC_HardwareSerial::peek(void)
{
  if (_rx_buffer_head == _rx_buffer_tail) {
    return -1;
  } else {
    return _rx_buffer[_rx_buffer_tail];
  }
}

int MXC_HardwareSerial::read(void)
{
  if (_rx_buffer_head == _rx_buffer_tail) {
    return -1;
  } else {
    unsigned char c = _rx_buffer[_rx_buffer_tail];
    _rx_buffer_tail = (rx_buffer_index_t)(_rx_buffer_tail + 1) % SERIAL_RX_BUFFER_SIZE;
    return c;
  }
}

int MXC_HardwareSerial::availableForWrite(void)
{
  // Leading tail
  if (_tx_buffer_tail > _tx_buffer_head) {
      return _tx_buffer_tail - _tx_buffer_head - 1;
  }
  // Leading Head
  if (_tx_buffer_tail < _tx_buffer_head) {
      return SERIAL_TX_BUFFER_SIZE - _tx_buffer_head + 1;
  }

  if (_tx_buffer_tail == _tx_buffer_head) {
    return SERIAL_TX_BUFFER_SIZE - 1;
  }
}

void MXC_HardwareSerial::flush(void)
{
  _rx_buffer_head = _rx_buffer_tail;
}

size_t MXC_HardwareSerial::write(uint8_t data)
{
  _uart->inten &= ~UART_WRITE_INTS;

  // Avoid software buffer if possible
  if (_tx_buffer_head == _tx_buffer_tail) {
    if (UART_NumWriteAvail(_uart)) {

      _fifo->tx = data;

      // Enable almost empty interrupt
      _uart->inten |= UART_WRITE_INTS;

      return 1;
    }
  }

  tx_buffer_index_t i = (_tx_buffer_head + 1) % SERIAL_TX_BUFFER_SIZE;

  if (i == _tx_buffer_tail) {
    _uart->inten |= UART_WRITE_INTS;
    while (i == _tx_buffer_tail);
    _uart->inten &= ~UART_WRITE_INTS;
  }
  _tx_buffer[_tx_buffer_head] = data;
  _tx_buffer_head = i;

  // Enable interrupts
  _uart->inten |= UART_WRITE_INTS;

  return 1;
}

void MXC_HardwareSerial::_handler(void)
{
  uint32_t flags;

  flags = _uart->intfl;
  _uart->intfl = flags;

  if (flags & UART_READ_INTS) {
    _rx_handler();
  }

  if (flags & UART_WRITE_INTS) {
    _tx_handler();
  }
}

void MXC_HardwareSerial::_rx_handler(void)
{
  // Disable receive interrupts while unloading FIFO
  _uart->inten &= ~UART_READ_INTS;

  int avail = UART_NumReadAvail(_uart);

  while (avail--) {
    rx_buffer_index_t i = (unsigned int)(_rx_buffer_head + 1) % SERIAL_RX_BUFFER_SIZE;
    char c = _fifo->rx;
    if (i != _rx_buffer_tail) {
      _rx_buffer[_rx_buffer_head] = c;
      _rx_buffer_head = i;
    }
  }

  _uart->inten |= UART_READ_INTS;
}

void MXC_HardwareSerial::_tx_handler(void)
{
  int avail;
  tx_buffer_index_t t;

  _uart->inten &= ~UART_WRITE_INTS;

  avail = UART_NumWriteAvail(_uart);

  while (avail && (_tx_buffer_tail != _tx_buffer_head)) {
#if (MXC_UART_REV == 0)
    _uart->intfl = MXC_F_UART_INTFL_TX_DONE;
#endif
    _fifo->tx = _tx_buffer[_tx_buffer_tail];

    _tx_buffer_tail = (_tx_buffer_tail + 1) % SERIAL_TX_BUFFER_SIZE;

    avail--;
  }

  _uart->inten |= UART_WRITE_INTS;
}

MXC_HardwareSerial Serial0(0);
MXC_HardwareSerial Serial1(1);
MXC_HardwareSerial Serial2(2);
#ifndef USBCON
HardwareSerial &Serial = CONCAT(Serial, DEFAULT_SERIAL_PORT);
#endif

extern "C" void UART0_IRQHandler(void) { Serial0._handler(); }
extern "C" void UART1_IRQHandler(void) { Serial1._handler(); }
extern "C" void UART2_IRQHandler(void) { Serial2._handler(); }

void serialEvent0(void) __attribute__((weak));
void serialEvent1(void) __attribute__((weak));
void serialEvent2(void) __attribute__((weak));

void serialEventRun(void)
{
  if (Serial0.available()) {
      if (Serial == Serial0) serialEvent();
      else serialEvent0();
  }
  if (Serial1.available()) {
      if (Serial == Serial1) serialEvent();
      else serialEvent1();
  }
  if (Serial2.available()) {
      if (Serial == Serial2) serialEvent();
      else serialEvent2();
  }
}
