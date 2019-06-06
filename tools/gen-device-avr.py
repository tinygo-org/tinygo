#!/usr/bin/env python3

import sys
import os
from xml.dom import minidom
from glob import glob
from collections import OrderedDict
import re

class Device:
    # dummy
    pass

def getText(element):
    strings = []
    for node in element.childNodes:
        if node.nodeType == node.TEXT_NODE:
            strings.append(node.data)
    return ''.join(strings)

def formatText(text):
    text = re.sub('[ \t\n]+', ' ', text) # Collapse whitespace (like in HTML)
    text = text.replace('\\n ', '\n')
    text = text.strip()
    return text

def readATDF(path):
    # Read Atmel device descriptor files.
    # See: http://packs.download.atmel.com

    device = Device()

    xml = minidom.parse(path)
    device = xml.getElementsByTagName('device')[0]
    deviceName = device.getAttribute('name')
    arch = device.getAttribute('architecture')
    family = device.getAttribute('family')

    memorySizes = {}
    for el in device.getElementsByTagName('address-space'):
        addressSpace = {
            'size': int(el.getAttribute('size'), 0),
            'segments': {},
        }
        memorySizes[el.getAttribute('name')] = addressSpace
        for segmentEl in el.getElementsByTagName('memory-segment'):
            addressSpace['segments'][segmentEl.getAttribute('name')] = int(segmentEl.getAttribute('size'), 0)

    device.interrupts = []
    for el in device.getElementsByTagName('interrupts')[0].getElementsByTagName('interrupt'):
        device.interrupts.append({
            'index':       int(el.getAttribute('index')),
            'name':        el.getAttribute('name'),
            'description': el.getAttribute('caption'),
        })

    allRegisters = {}
    commonRegisters = {}

    device.peripherals = []
    for el in xml.getElementsByTagName('modules')[0].getElementsByTagName('module'):
        peripheral = {
            'name':        el.getAttribute('name'),
            'description': el.getAttribute('caption'),
            'registers':   [],
        }
        device.peripherals.append(peripheral)
        for regElGroup in el.getElementsByTagName('register-group'):
            for regEl in regElGroup.getElementsByTagName('register'):
                size = int(regEl.getAttribute('size'))
                regName = regEl.getAttribute('name')
                regOffset = int(regEl.getAttribute('offset'), 0)
                reg = {
                    'description': regEl.getAttribute('caption'),
                    'bitfields':   [],
                    'array':       None,
                }
                if size == 1:
                    reg['variants'] = [{
                        'name':    regName,
                        'address': regOffset,
                    }]
                elif size == 2:
                    reg['variants'] = [{
                        'name':    regName + 'L',
                        'address': regOffset,
                    }, {
                        'name':    regName + 'H',
                        'address': regOffset + 1,
                    }]
                else:
                    # TODO
                    continue

                for bitfieldEl in regEl.getElementsByTagName('bitfield'):
                    mask = bitfieldEl.getAttribute('mask')
                    if len(mask) == 2:
                        # Two devices (ATtiny102 and ATtiny104) appear to have
                        # an error in the bitfields, leaving out the '0x'
                        # prefix.
                        mask = '0x' + mask
                    reg['bitfields'].append({
                        'name':        regName + '_' + bitfieldEl.getAttribute('name'),
                        'description': bitfieldEl.getAttribute('caption'),
                        'value':       int(mask, 0),
                    })

                if regName in allRegisters:
                    firstReg = allRegisters[regName]
                    if firstReg['register'] in firstReg['peripheral']['registers']:
                        firstReg['peripheral']['registers'].remove(firstReg['register'])
                    if firstReg['address'] != regOffset:
                        continue # TODO
                    commonRegisters = allRegisters[regName]['register']
                    continue
                else:
                    allRegisters[regName] = {'address': regOffset, 'register': reg, 'peripheral': peripheral}

                peripheral['registers'].append(reg)

    ramSize = 0 # for devices with no RAM
    for ramSegmentName in ['IRAM', 'INTERNAL_SRAM', 'SRAM']:
        if ramSegmentName in memorySizes['data']['segments']:
            ramSize = memorySizes['data']['segments'][ramSegmentName]

    device.metadata = {
        'file':             os.path.basename(path),
        'descriptorSource': 'http://packs.download.atmel.com/',
        'name':             deviceName,
        'nameLower':        deviceName.lower(),
        'description':      'Device information for the {}.'.format(deviceName),
        'arch':             arch,
        'family':           family,
        'flashSize':        memorySizes['prog']['size'],
        'ramSize':          ramSize,
        'numInterrupts':    len(device.interrupts),
    }

    return device

def writeGo(outdir, device):
    # The Go module for this device.
    out = open(outdir + '/' + device.metadata['nameLower'] + '.go', 'w')
    pkgName = os.path.basename(outdir.rstrip('/'))
    out.write('''\
// Automatically generated file. DO NOT EDIT.
// Generated by gen-device-avr.py from {file}, see {descriptorSource}

// +build {pkgName},{nameLower}

// {description}
package {pkgName}

import (
	"runtime/volatile"
	"unsafe"
)

// Some information about this device.
const (
	DEVICE     = "{name}"
	ARCH       = "{arch}"
	FAMILY     = "{family}"
)
'''.format(pkgName=pkgName, **device.metadata))

    out.write('\n// Interrupts\nconst (\n')
    for intr in device.interrupts:
        out.write('\tIRQ_{name} = {index} // {description}\n'.format(**intr))
    intrMax = max(map(lambda intr: intr['index'], device.interrupts))
    out.write('\tIRQ_max = {} // Highest interrupt number on this device.\n'.format(intrMax))
    out.write(')\n')

    out.write('\n// Peripherals.\nvar (')
    first = True
    for peripheral in device.peripherals:
        out.write('\n\t// {description}\n'.format(**peripheral))
        for register in peripheral['registers']:
            for variant in register['variants']:
                out.write('\t{name} = (*volatile.Register8)(unsafe.Pointer(uintptr(0x{address:x})))\n'.format(**variant))
    out.write(')\n')

    for peripheral in device.peripherals:
        if not sum(map(lambda r: len(r['bitfields']), peripheral['registers'])): continue
        out.write('\n// Bitfields for {name}: {description}\nconst('.format(**peripheral))
        for register in peripheral['registers']:
            if not register['bitfields']: continue
            for variant in register['variants']:
                out.write('\n\t// {name}'.format(**variant))
                if register['description']:
                    out.write(': {description}'.format(**register))
                out.write('\n')
            for bitfield in register['bitfields']:
                name = bitfield['name']
                value = bitfield['value']
                if '{:08b}'.format(value).count('1') == 1:
                    out.write('\t{name} = 0x{value:x}'.format(**bitfield))
                    if bitfield['description']:
                        out.write(' // {description}'.format(**bitfield))
                    out.write('\n')
                else:
                    n = 0
                    for i in range(8):
                        if (value >> i) & 1 == 0: continue
                        out.write('\t{}{} = 0x{:x}'.format(name, n, 1 << i))
                        if bitfield['description']:
                            out.write(' // {description}'.format(**bitfield))
                        n += 1
                        out.write('\n')
        out.write(')\n')

def writeAsm(outdir, device):
    # The interrupt vector, which is hard to write directly in Go.
    out = open(outdir + '/' + device.metadata['nameLower'] + '.s', 'w')
    out.write('''\
; Automatically generated file. DO NOT EDIT.
; Generated by gen-device-avr.py from {file}, see {descriptorSource}

; This is the default handler for interrupts, if triggered but not defined.
; Sleep inside so that an accidentally triggered interrupt won't drain the
; battery of a battery-powered device.
.section .text.__vector_default
.global  __vector_default
__vector_default:
    sleep
    rjmp __vector_default

; Avoid the need for repeated .weak and .set instructions.
.macro IRQ handler
    .weak  \\handler
    .set   \\handler, __vector_default
.endm

; The interrupt vector of this device. Must be placed at address 0 by the linker.
.section .vectors
.global  __vectors
'''.format(**device.metadata))
    num = 0
    for intr in device.interrupts:
        jmp = 'jmp'
        if device.metadata['flashSize'] <= 8 * 1024:
            # When a device has 8kB or less flash, rjmp (2 bytes) must be used
            # instead of jmp (4 bytes).
            # https://www.avrfreaks.net/forum/rjmp-versus-jmp
            jmp = 'rjmp'
        if intr['index'] < num:
            # Some devices have duplicate interrupts, probably for historical
            # reasons.
            continue
        while intr['index'] > num:
            out.write('    {jmp} __vector_default\n'.format(jmp=jmp))
            num += 1
        num += 1
        out.write('    {jmp} __vector_{name}\n'.format(jmp=jmp, **intr))

    out.write('''
    ; Define default implementations for interrupts, redirecting to
    ; __vector_default when not implemented.
''')
    for intr in device.interrupts:
        out.write('    IRQ __vector_{name}\n'.format(**intr))

def writeLD(outdir, device):
    # Variables for the linker script.
    out = open(outdir + '/' + device.metadata['nameLower'] + '.ld', 'w')
    out.write('''\
/* Automatically generated file. DO NOT EDIT. */
/* Generated by gen-device-avr.py from {file}, see {descriptorSource} */

__flash_size = 0x{flashSize:x};
__ram_size   = 0x{ramSize:x};
__num_isrs   = {numInterrupts};
'''.format(**device.metadata))
    out.close()


def generate(indir, outdir):
    for filepath in sorted(glob(indir + '/*.atdf')):
        print(filepath)
        device = readATDF(filepath)
        writeGo(outdir, device)
        writeAsm(outdir, device)
        writeLD(outdir, device)


if __name__ == '__main__':
    indir = sys.argv[1] # directory with register descriptor files (*.atdf)
    outdir = sys.argv[2] # output directory
    generate(indir, outdir)
