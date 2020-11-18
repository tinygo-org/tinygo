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
 *
 * $Date: 2017-01-16 14:05:15 -0600 (Mon, 16 Jan 2017) $
 * $Revision: 25899 $
 *
 ******************************************************************************/

#include <stdlib.h>
#include <string.h>
#include "mxc_config.h"
#include "usb_regs.h"
#include "usb.h"

typedef struct {
    volatile uint32_t buf0_desc;
    volatile uint32_t buf0_address;
    volatile uint32_t buf1_desc;
    volatile uint32_t buf1_address;
} ep_buffer_t;

typedef struct {
    ep_buffer_t out_buffer;
    ep_buffer_t in_buffer;
} ep0_buffer_t;

typedef struct {
    ep0_buffer_t ep0;
    ep_buffer_t ep[MXC_USB_NUM_EP - 1];
} ep_buffer_descriptor_t;

/* static storage for endpoint buffer descriptor table, must be 512 byte aligned for DMA */
#ifdef __IAR_SYSTEMS_ICC__
#pragma data_alignment = 512
#else
__attribute__ ((aligned (512)))
#endif
ep_buffer_descriptor_t ep_buffer_descriptor;

/* storage for active endpoint data request objects */
static usb_req_t *usb_request[MXC_USB_NUM_EP];

static uint8_t ep_size[MXC_USB_NUM_EP];

int usb_init(maxusb_cfg_options_t *options)
{
    int i;
    for (i = 0; i < MXC_USB_NUM_EP; i++) {
        usb_request[i] = NULL;
        ep_size[i] = 0;
    }

    /* set the EP0 size */
    ep_size[0] = MXC_USB_MAX_PACKET;

    /* reset the device */
    MXC_USB->cn = 0;
    MXC_USB->cn = 1;
    MXC_USB->dev_inten = 0;
    MXC_USB->dev_cn = 0;
    MXC_USB->dev_cn = MXC_F_USB_DEV_CN_URST;
    MXC_USB->dev_cn = 0;

    /* set the descriptor location */
    MXC_USB->ep_base = (uint32_t)&ep_buffer_descriptor;

    return 0;
}

int usb_shutdown(void)
{
    MXC_USB->dev_cn = MXC_F_USB_DEV_CN_URST;
    MXC_USB->dev_cn = 0;
    MXC_USB->cn = 0;
    return 0;
}

int usb_connect(void)
{
    /* enable interrupts */
    MXC_USB->dev_inten |= (MXC_F_USB_DEV_INTEN_SETUP | MXC_F_USB_DEV_INTEN_EP_IN | MXC_F_USB_DEV_INTEN_EP_OUT | MXC_F_USB_DEV_INTEN_DMA_ERR);

    /* allow interrupts on ep0 */
    MXC_USB->ep[0] |= MXC_F_USB_EP_INT_EN;

    /* pull-up enable */
    MXC_USB->dev_cn |= (MXC_F_USB_DEV_CN_CONNECT | MXC_F_USB_DEV_CN_FIFO_MODE);

    return 0;
}

int usb_disconnect(void)
{
    MXC_USB->dev_cn &= ~MXC_F_USB_DEV_CN_CONNECT;
    return 0;
}

int usb_config_ep(unsigned int ep, maxusb_ep_type_t type, unsigned int size)
{
    uint32_t ep_ctrl;

    if (ep >= MXC_USB_NUM_EP) {
        return -1;
    }

    if (size > MXC_USB_MAX_PACKET) {
        return -1;
    }

    ep_ctrl = (type << MXC_F_USB_EP_DIR_POS);    /* input 'type' matches USB spec for direction */
    ep_ctrl |= MXC_F_USB_EP_DT;

    if (type == MAXUSB_EP_TYPE_DISABLED) {
        ep_ctrl &= ~MXC_F_USB_EP_INT_EN;
    } else {
        ep_ctrl |= MXC_F_USB_EP_INT_EN;
    }

    ep_size[ep] = size;

    MXC_USB->ep[ep] = ep_ctrl;
    return 0;
}

int usb_is_configured(unsigned int ep)
{
    return (((MXC_USB->ep[ep] & MXC_F_USB_EP_DIR) >> MXC_F_USB_EP_DIR_POS) != MXC_V_USB_EP_DIR_DISABLE);
}

int usb_stall(unsigned int ep)
{
    usb_req_t *req;

    if (ep == 0) {
        MXC_USB->ep[ep] |= MXC_F_USB_EP_ST_STALL;
    }

    MXC_USB->ep[ep] |= MXC_F_USB_EP_STALL;

    /* clear pending requests */
    req = usb_request[ep];
    usb_request[ep] = NULL;

    if (req) {
        /* complete pending requests with error */
        req->error_code = -1;
        if (req->callback) {
            req->callback(req->cbdata);
        }
    }

    return 0;
}

int usb_unstall(unsigned int ep)
{
    /* clear the data toggle */
    MXC_USB->ep[ep] |= MXC_F_USB_EP_DT;

    MXC_USB->ep[ep] &= ~MXC_F_USB_EP_STALL;

    return 0;
}

int usb_is_stalled(unsigned int ep)
{
    return !!(MXC_USB->ep[ep] & MXC_F_USB_EP_STALL);
}

int usb_reset_ep(unsigned int ep)
{
    usb_req_t *req;

    if (ep >= MXC_USB_NUM_EP) {
        return -1;
    }

    /* clear pending requests */
    req = usb_request[ep];
    usb_request[ep] = NULL;

    /* disable the EP */
    MXC_USB->ep[ep] &= ~MXC_F_USB_EP_DIR;

    /* clear the data toggle */
    MXC_USB->ep[ep] |= MXC_F_USB_EP_DT;

    if (req) {
        /* complete pending requests with error */
        req->error_code = -1;
        if (req->callback) {
            req->callback(req->cbdata);
        }
    }

    return 0;
}

int usb_ackstat(unsigned int ep)
{
    MXC_USB->ep[ep] |= MXC_F_USB_EP_ST_ACK;
    return 0;
}

/* sent packet done handler*/
static void event_in_data(uint32_t irqs)
{
    uint32_t epnum, buffer_bit, data_left;
    usb_req_t *req;
    ep_buffer_t *buf_desc;

    /* Loop for each data endpoint */
    for (epnum = 0; epnum < MXC_USB_NUM_EP; epnum++) {

        buffer_bit = (1 << epnum);
        if ((irqs & buffer_bit) == 0) { /* Not set, next Endpoint */
            continue;
        }

        /* not sure how this could happen, safe anyway */
        if (!usb_request[epnum]) {
            continue;
        }

        req = usb_request[epnum];
        data_left = req->reqlen - req->actlen;

        if (epnum == 0) {
            buf_desc = &ep_buffer_descriptor.ep0.in_buffer;
        } else {
            buf_desc = &ep_buffer_descriptor.ep[epnum-1];
        }

        if (buf_desc->buf0_desc == 0) {
            /* free request first, the callback may re-issue a request */
            usb_request[epnum] = NULL;

            /* must have sent the ZLP, mark done */
            if (req->callback) {
                req->callback(req->cbdata);
            }
            continue;
        }

        if (data_left) {   /* more data to send */
            if (data_left >= ep_size[epnum]) {
                buf_desc->buf0_desc = ep_size[epnum];
            } else {
                buf_desc->buf0_desc = data_left;
            }

            req->actlen += buf_desc->buf0_desc;

            /* update the pointer */
            buf_desc->buf0_address += ep_size[epnum];

            /* start the DMA to send it */
            MXC_USB->in_owner = buffer_bit;
        }
        else {
            /* all done sending data, either send ZLP or done here */
            if ((req->reqlen & (ep_size[epnum]-1)) == 0) {
                /* send ZLP per spec, last packet was full sized and nothing left to send */
                buf_desc->buf0_desc = 0;
                MXC_USB->in_owner = buffer_bit;
            }
            else {
                /* free request */
                usb_request[epnum] = NULL;

                /* set done return value */
                if (req->callback) {
                    req->callback(req->cbdata);
                }
            }
        }
    }
}

/* received packet */
static void event_out_data(uint32_t irqs)
{
    uint32_t epnum, buffer_bit, reqsize, rxsize;
    usb_req_t *req;
    ep_buffer_t *buf_desc;

    /* Loop for each data endpoint */
    for (epnum = 0; epnum < MXC_USB_NUM_EP; epnum++) {

        buffer_bit = (1 << epnum);
        if ((irqs & buffer_bit) == 0) {
            continue;
        }

        /* this can happen if the callback was called then ZLP received */
        if (!usb_request[epnum]) {
            continue; /* ignored, because callback must have been called */
        }

        if (epnum == 0) {
            buf_desc = &ep_buffer_descriptor.ep0.out_buffer;
        } else {
            buf_desc = &ep_buffer_descriptor.ep[epnum-1];
        }

        req = usb_request[epnum];

        /* what was the last request size? */
        if ((req->reqlen - req->actlen) >= ep_size[epnum]) {
            reqsize = ep_size[epnum];
        } else {
            reqsize = (req->reqlen - req->actlen);
        }

        /* the actual size of data written to buffer will be the lesser of the packet size and the requested size */
        if (reqsize < buf_desc->buf0_desc) {
            rxsize = reqsize;
        } else {
            rxsize = buf_desc->buf0_desc;
        }

        req->actlen += rxsize;

        /* less than a full packet or zero length packet)  */
        if ((req->type == MAXUSB_TYPE_PKT) || (rxsize < ep_size[epnum]) || (rxsize == 0)) {
            /* free request */
            usb_request[epnum] = NULL;

            /* call it done */
            if (req->callback) {
                req->callback(req->cbdata);
            }
        }
        else {
            /* not done yet, push pointers, update descriptor */
            buf_desc->buf0_address += ep_size[epnum];

            /* don't overflow */
            if ((req->reqlen - req->actlen) >= ep_size[epnum]) {
                buf_desc->buf0_desc = ep_size[epnum];
            } else {
                buf_desc->buf0_desc = (req->reqlen - req->actlen);
            }

            /* transfer buffer back to controller */
            MXC_USB->out_owner = buffer_bit;
        }
    }
}

void usb_irq_handler(maxusb_usbio_events_t *evt)
{
    uint32_t in_irqs, out_irqs, irq_flags;

    /* get and clear enabled irqs */
    irq_flags = MXC_USB->dev_intfl & MXC_USB->dev_inten;
    MXC_USB->dev_intfl = irq_flags;

    /* copy all the flags over to the "other" struct */
    evt->dpact  = ((irq_flags & MXC_F_USB_DEV_INTFL_DPACT) >> MXC_F_USB_DEV_INTFL_DPACT_POS);
    evt->rwudn  = ((irq_flags & MXC_F_USB_DEV_INTFL_RWU_DN) >> MXC_F_USB_DEV_INTFL_RWU_DN_POS);
    evt->bact   = ((irq_flags & MXC_F_USB_DEV_INTFL_BACT) >> MXC_F_USB_DEV_INTFL_BACT_POS);
    evt->brst   = ((irq_flags & MXC_F_USB_DEV_INTFL_BRST) >> MXC_F_USB_DEV_INTFL_BRST_POS);
    evt->susp   = ((irq_flags & MXC_F_USB_DEV_INTFL_SUSP) >> MXC_F_USB_DEV_INTFL_SUSP_POS);
    evt->novbus = ((irq_flags & MXC_F_USB_DEV_INTFL_NO_VBUS) >> MXC_F_USB_DEV_INTFL_NO_VBUS_POS);
    evt->vbus   = ((irq_flags & MXC_F_USB_DEV_INTFL_VBUS) >> MXC_F_USB_DEV_INTFL_VBUS_POS);
    evt->brstdn = ((irq_flags & MXC_F_USB_DEV_INTFL_BRST_DN) >> MXC_F_USB_DEV_INTFL_BRST_DN_POS);
    evt->sudav  = ((irq_flags & MXC_F_USB_DEV_INTFL_SETUP) >> MXC_F_USB_DEV_INTFL_SETUP_POS);

    /* do cleanup in cases of bus reset */
    if (irq_flags & MXC_F_USB_DEV_INTFL_BRST) {
        int i;

        /* kill any pending requests */
        for (i = 0; i < MXC_USB_NUM_EP; i++) {
            usb_reset_ep(i);
        }
        /* no need to process events after reset */
        return;
    }

    if (irq_flags & MXC_F_USB_DEV_INTFL_EP_IN) {
        /* get and clear IN irqs */
        in_irqs = MXC_USB->in_int;
        MXC_USB->in_int = in_irqs;
        event_in_data(in_irqs);
    }

    if (irq_flags & MXC_F_USB_DEV_INTFL_EP_OUT) {
        /* get and clear OUT irqs */
        out_irqs = MXC_USB->out_int;
        MXC_USB->out_int = out_irqs;
        event_out_data(out_irqs);
    }
}

int usb_irq_enable(maxusb_event_t event)
{
    uint32_t event_bit = 0;

    if (event >= MAXUSB_NUM_EVENTS) {
        return -1;
    }

    /* Note: the enum value is the same as the bit number */
    event_bit =  1 << event;
    MXC_USB->dev_inten |= event_bit;

    return 0;
}

int usb_irq_disable(maxusb_event_t event)
{
    uint32_t event_bit = 0;

    if (event >= MAXUSB_NUM_EVENTS) {
        return -1;
    }

    /* Note: the enum value is the same as the bit number */
    event_bit =  1 << event;
    MXC_USB->dev_inten &= ~event_bit;

    return 0;
}

int usb_irq_clear(maxusb_event_t event)
{
    uint32_t event_bit = 0;

    if (event >= MAXUSB_NUM_EVENTS) {
        return -1;
    }

    /* Note: the enum value is the same as the bit number */
    event_bit =  1 << event;
    MXC_USB->dev_intfl = event_bit;

    return 0;
}

int usb_get_setup(usb_setup_pkt *sud)
{
    memcpy(sud, (void*)&MXC_USB->setup0, 8); /* setup packet is fixed at 8 bytes */
    return 0;
}

int usb_set_func_addr(unsigned int addr)
{
    /* Hardware does this for us. Just return success so caller will ACK the setup packet */
    return 0;
}

usb_req_t *usb_get_request(unsigned int ep)
{
    return usb_request[ep];
}

int usb_write_endpoint(usb_req_t *req)
{
    unsigned int ep  = req->ep;
    uint8_t *data    = req->data;
    unsigned int len = req->reqlen;
    ep_buffer_t *buf_desc;
    uint32_t buffer_bit = (1 << ep);

    if (ep >= MXC_USB_NUM_EP) {
        return -1;
    }

    /* data buffer must be 32-bit aligned */
    if ((unsigned int)data & 0x3) {
        return -1;
    }

    /* EP must be enabled (configured) */
    if (((MXC_USB->ep[ep] & MXC_F_USB_EP_DIR) >> MXC_F_USB_EP_DIR_POS) == MXC_V_USB_EP_DIR_DISABLE) {
        return -1;
    }

    /* if pending request; error */
    if (usb_request[ep] || (MXC_USB->in_owner & buffer_bit)) {
        return -1;
    }

    /* assign req object */
    usb_request[ep] = req;

    /* clear errors */
    req->error_code = 0;

    if (ep == 0) {
        buf_desc = &ep_buffer_descriptor.ep0.in_buffer;
    } else {
        buf_desc = &ep_buffer_descriptor.ep[ep-1];
    }

    if (len > ep_size[ep]) {
        buf_desc->buf0_desc = ep_size[ep];
        usb_request[ep]->actlen = ep_size[ep];
    }
    else {
        buf_desc->buf0_desc = len;
        usb_request[ep]->actlen = len;
    }

    /* set pointer, force single buffered */
    buf_desc->buf0_address = (uint32_t)data;

    /* start the DMA */
    MXC_USB->in_owner = buffer_bit;

    return 0;
}

int usb_read_endpoint(usb_req_t *req)
{
    unsigned int ep  = req->ep;
    ep_buffer_t *buf_desc;
    uint32_t buffer_bit = (1 << ep);

    if (ep >= MXC_USB_NUM_EP) {
        return -1;
    }

    /* data buffer must be 32-bit aligned */
    if ((unsigned int)req->data & 0x3) {
        return -1;
    }

    /* EP must be enabled (configured) and not stalled */
    if (!usb_is_configured(ep) || usb_is_stalled(ep)) {
        return -1;
    }

    if (ep == 0) {
        buf_desc = &ep_buffer_descriptor.ep0.out_buffer;
    } else {
        buf_desc = &ep_buffer_descriptor.ep[ep-1];
    }

    /* if pending request; error */
    if (usb_request[ep] || (MXC_USB->out_owner & buffer_bit)) {
        return -1;
    }

    /* assign the req object */
    usb_request[ep] = req;

    /* clear errors */
    req->error_code = 0;

    /* reset length */
    req->actlen = 0;

    if (req->reqlen < ep_size[ep]) {
        buf_desc->buf0_desc = req->reqlen;
    } else {
        buf_desc->buf0_desc = ep_size[ep];
    }
    buf_desc->buf0_address = (uint32_t)req->data;

    MXC_USB->out_owner = buffer_bit;

    return 0;
}
