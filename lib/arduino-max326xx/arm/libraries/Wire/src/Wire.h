/*
  TwoWire.h - TWI/I2C library for Arduino & Wiring
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

  Modified 2012 by Todd Krein (todd@krein.org) to implement repeated starts

  Modified 2017 by Maxim Integrated for MAX326xx

*/

#ifndef TwoWire_h
#define TwoWire_h

#include <Arduino.h>
#include <inttypes.h>
#include "Stream.h"
#include "i2cm_regs.h"

#define BUFFER_LENGTH 32

// WIRE_HAS_END means Wire has end()
#define WIRE_HAS_END 1

enum TransmitStatus {
    TX_SUCCESS = 0,
    DATA_TOO_LONG,
    NACK_TX_ADDR,
    NACK_TX_DATA,
    OTHER_ERROR
};

class TwoWire : public Stream
{
  private:
    uint8_t rxBuffer[BUFFER_LENGTH];
    uint8_t rxBufferIndex;
    uint8_t rxBufferLength;

    uint8_t txAddress;
    uint8_t txBuffer[BUFFER_LENGTH];
    uint8_t txBufferIndex;
    uint8_t txBufferLength;

    uint8_t transmitting;

    uint8_t targetAddr;
    uint32_t idx;
    mxc_i2cm_regs_t *i2cm;
    mxc_i2cm_fifo_regs_t *fifo;

    uint8_t map_to_arduino_err(int);

  public:
    TwoWire(uint32_t);
    void begin();
    void begin(uint8_t);
    void begin(int);
    void end();
    void setClock(uint32_t);
    void beginTransmission(uint8_t);
    void beginTransmission(int);
    uint8_t endTransmission(void);
    uint8_t endTransmission(uint8_t);
    uint8_t requestFrom(uint8_t, uint8_t);
    uint8_t requestFrom(uint8_t, uint8_t, uint8_t);
    uint8_t requestFrom(int, int);
    uint8_t requestFrom(int, int, int);
    virtual size_t write(uint8_t);
    virtual size_t write(const uint8_t *, size_t);
    virtual int available(void);
    virtual int read(void);
    virtual int peek(void);
    virtual void flush(void);

    inline size_t write(unsigned long n) { return write((uint8_t)n); }
    inline size_t write(long n) { return write((uint8_t)n); }
    inline size_t write(unsigned int n) { return write((uint8_t)n); }
    inline size_t write(int n) { return write((uint8_t)n); }
    using Print::write;
};

extern TwoWire Wire0;
#if (MXC_CFG_I2CM_INSTANCES > 1)
extern TwoWire Wire1;
#endif
#if (MXC_CFG_I2CM_INSTANCES > 2)
extern TwoWire Wire2;
#endif
#if (MXC_CFG_I2CM_INSTANCES > 3)
extern TwoWire Wire3;
#endif
#if (MXC_CFG_I2CM_INSTANCES > 4)
extern TwoWire Wire4;
#endif
#if (MXC_CFG_I2CM_INSTANCES > 5)
extern TwoWire Wire5;
#endif
extern TwoWire &Wire;

#endif // TwoWire_h
