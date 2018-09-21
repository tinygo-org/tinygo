#!/usr/bin/python3

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

def readSVD(path):
    # Read ARM SVD files.
    device = Device()
    xml = minidom.parse(path)
    root = xml.getElementsByTagName('device')[0]
    deviceName = getText(root.getElementsByTagName('name')[0])
    deviceDescription = getText(root.getElementsByTagName('description')[0]).strip()
    licenseTexts = root.getElementsByTagName('licenseText')
    if len(licenseTexts) == 0:
        licenseText = None
    elif len(licenseTexts) == 1:
        licenseText = formatText(getText(licenseTexts[0]))
    else:
        raise ValueError('multiple <licenseText> elements')

    device.peripherals = []
    peripheralDict = {}

    interrupts = OrderedDict()

    for periphEl in root.getElementsByTagName('peripherals')[0].getElementsByTagName('peripheral'):
        name = getText(periphEl.getElementsByTagName('name')[0])
        descriptionTags = periphEl.getElementsByTagName('description')
        description = ''
        if descriptionTags:
            description = formatText(getText(descriptionTags[0]))
        baseAddress = int(getText(periphEl.getElementsByTagName('baseAddress')[0]), 0)
        groupNameTags = periphEl.getElementsByTagName('groupName')
        groupName = name
        if groupNameTags:
            groupName = getText(groupNameTags[0])

        if periphEl.hasAttribute('derivedFrom'):
            derivedFromName = periphEl.getAttribute('derivedFrom')
            derivedFrom = peripheralDict[derivedFromName]
            peripheral = {
                'name':        name,
                'groupName':   derivedFrom['groupName'],
                'description': description if description is not None else derivedFrom['description'],
                'baseAddress': baseAddress,
            }
            device.peripherals.append(peripheral)
            continue

        peripheral = {
            'name':        name,
            'groupName':   groupName,
            'description': description,
            'baseAddress': baseAddress,
            'registers':   [],
        }
        device.peripherals.append(peripheral)
        peripheralDict[name] = peripheral

        for interrupt in periphEl.getElementsByTagName('interrupt'):
            intrName = getText(interrupt.getElementsByTagName('name')[0])
            intrIndex = int(getText(interrupt.getElementsByTagName('value')[0]))
            if intrName in interrupts:
                if interrupts[intrName]['index'] != intrIndex:
                    raise ValueError('interrupt with the same name has different indexes: ' + intrName)
                interrupts[intrName]['description'] += ' // ' + description
            else:
                interrupts[intrName] = {
                    'name':        intrName,
                    'index':       intrIndex,
                    'description': description,
                }

        regsEls = periphEl.getElementsByTagName('registers')
        if regsEls:
            for el in regsEls[0].childNodes:
                if el.nodeName == 'register':
                    peripheral['registers'].append(parseSVDRegister(groupName, el, baseAddress))
                elif el.nodeName == 'cluster':
                    if el.getElementsByTagName('dim'):
                        continue # TODO
                    clusterPrefix = getText(el.getElementsByTagName('name')[0]) + '_'
                    clusterOffset = int(getText(el.getElementsByTagName('addressOffset')[0]), 0)
                    for regEl in el.childNodes:
                        if regEl.nodeName == 'register':
                            peripheral['registers'].append(parseSVDRegister(groupName, regEl, baseAddress + clusterOffset, clusterPrefix))
                else:
                    continue

    device.interrupts = sorted(interrupts.values(), key=lambda v: v['index'])
    licenseBlock = ''
    if licenseText is not None:
        licenseBlock = '//     ' + licenseText.replace('\n', '\n//     ')
        licenseBlock = '\n'.join(map(str.rstrip, licenseBlock.split('\n'))) # strip trailing whitespace
    device.metadata = {
        'file':             os.path.basename(path),
        'descriptorSource': 'https://github.com/NordicSemiconductor/nrfx/tree/master/mdk',
        'name':             deviceName,
        'nameLower':        deviceName.lower(),
        'description':      deviceDescription,
        'licenseBlock':     licenseBlock,
    }

    return device

def parseSVDRegister(groupName, regEl, baseAddress, namePrefix=''):
    regName = getText(regEl.getElementsByTagName('name')[0])
    regDescription = getText(regEl.getElementsByTagName('description')[0])
    offsetEls = regEl.getElementsByTagName('offset')
    if not offsetEls:
        offsetEls = regEl.getElementsByTagName('addressOffset')
    address = baseAddress + int(getText(offsetEls[0]), 0)

    dimEls = regEl.getElementsByTagName('dim')
    array = None
    if dimEls:
        array = int(getText(dimEls[0]), 0)
        regName = regName.replace('[%s]', '')

    fields = []
    fieldsEls = regEl.getElementsByTagName('fields')
    if fieldsEls:
        for fieldEl in fieldsEls[0].childNodes:
            if fieldEl.nodeName != 'field':
                continue
            fieldName = getText(fieldEl.getElementsByTagName('name')[0])
            descrEls = fieldEl.getElementsByTagName('description')
            lsbTags = fieldEl.getElementsByTagName('lsb')
            if len(lsbTags) == 1:
                lsb = int(getText(lsbTags[0]))
            else:
                lsb = int(getText(fieldEl.getElementsByTagName('bitOffset')[0]))
            msbTags = fieldEl.getElementsByTagName('msb')
            if len(msbTags) == 1:
                msb = int(getText(msbTags[0]))
            else:
                msb = int(getText(fieldEl.getElementsByTagName('bitWidth')[0])) + lsb - 1
            fields.append({
                'name':        '{}_{}{}_{}_Pos'.format(groupName, namePrefix, regName, fieldName),
                'description': 'Position of %s field.' % fieldName,
                'value':       lsb,
            })
            fields.append({
                'name':        '{}_{}{}_{}_Msk'.format(groupName, namePrefix, regName, fieldName),
                'description': 'Bit mask of %s field.' % fieldName,
                'value':       (0xffffffff >> (31 - (msb - lsb))) << lsb,
            })
            for enumEl in fieldEl.getElementsByTagName('enumeratedValue'):
                enumName = getText(enumEl.getElementsByTagName('name')[0])
                enumDescription = getText(enumEl.getElementsByTagName('description')[0])
                enumValue = int(getText(enumEl.getElementsByTagName('value')[0]), 0)
                fields.append({
                    'name':        '{}_{}{}_{}_{}'.format(groupName, namePrefix, regName, fieldName, enumName),
                    'description': enumDescription,
                    'value':       enumValue,
                })

    return {
        'name':    namePrefix + regName,
        'address': address,
        'description': regDescription.replace('\n', ' '),
        'bitfields':   fields,
        'array':       array,
    }

def writeGo(outdir, device):
    # The Go module for this device.
    out = open(outdir + '/' + device.metadata['nameLower'] + '.go', 'w')
    pkgName = os.path.basename(outdir.rstrip('/'))
    out.write('''\
// Automatically generated file. DO NOT EDIT.
// Generated by gen-device.py from {file}, see {descriptorSource}

// +build {pkgName},{nameLower}

// {description}
//
{licenseBlock}
package {pkgName}

import "unsafe"

// Magic type name for the compiler.
type __volatile uint32

// Export this magic type name.
type RegValue = __volatile

// Some information about this device.
const (
	DEVICE     = "{name}"
)
'''.format(pkgName=pkgName, **device.metadata))

    out.write('\n// Interrupt numbers\nconst (\n')
    for intr in device.interrupts:
        out.write('\tIRQ_{name} = {index} // {description}\n'.format(**intr))
    intrMax = max(map(lambda intr: intr['index'], device.interrupts))
    out.write('\tIRQ_max = {} // Highest interrupt number on this device.\n'.format(intrMax))
    out.write(')\n')

    # Define peripheral struct types.
    for peripheral in device.peripherals:
        if 'registers' not in peripheral:
            # This peripheral was derived from another peripheral. No new type
            # needs to be defined for it.
            continue
        out.write('\n// {description}\ntype {groupName}_Type struct {{\n'.format(**peripheral))
        address = peripheral['baseAddress']
        padNumber = 0
        for register in peripheral['registers']:
            if address > register['address']:
                # In Nordic SVD files, these registers are deprecated or
                # duplicates, so can be ignored.
                #print('skip: %s.%s' % (peripheral['name'], register['name']))
                continue

            # insert padding, if needed
            if address < register['address']:
                numSkip = (register['address'] - address) // 4
                if numSkip == 1:
                    out.write('\t_padding{padNumber} __volatile\n'.format(padNumber=padNumber))
                else:
                    out.write('\t_padding{padNumber} [{num}]__volatile\n'.format(padNumber=padNumber, num=numSkip))
                padNumber += 1

            regType = '__volatile'
            if register['array'] is not None:
                regType = '[{}]__volatile'.format(register['array'])
            out.write('\t{name} {regType}\n'.format(**register, regType=regType))

            # next address
            if register['array'] is not None and 1:
                address = register['address'] + 4 * register['array']
            else:
                address = register['address'] + 4
        out.write('}\n')

    # Define actual peripheral pointers.
    out.write('\n// Peripherals.\nvar (\n')
    for peripheral in device.peripherals:
        out.write('\t{name} = (*{groupName}_Type)(unsafe.Pointer(uintptr(0x{baseAddress:x}))) // {description}\n'.format(**peripheral))
    out.write(')\n')

    # Define bitfields.
    for peripheral in device.peripherals:
        if 'registers' not in peripheral:
            # This peripheral was derived from another peripheral. Bitfields are
            # already defined.
            continue
        if not sum(map(lambda r: len(r['bitfields']), peripheral['registers'])): continue
        out.write('\n// Bitfields for {name}: {description}\nconst('.format(**peripheral))
        for register in peripheral['registers']:
            if not register['bitfields']: continue
            out.write('\n\t// {name}'.format(**register))
            if register['description']:
                out.write(': {description}'.format(**register))
            out.write('\n')
            for bitfield in register['bitfields']:
                out.write('\t{name} = 0x{value:x}'.format(**bitfield))
                if bitfield['description']:
                    out.write(' // {description}'.format(**bitfield))
                out.write('\n')
        out.write(')\n')


def writeAsm(outdir, device):
    # The interrupt vector, which is hard to write directly in Go.
    out = open(outdir + '/' + device.metadata['nameLower'] + '.s', 'w')
    out.write('''\
// Automatically generated file. DO NOT EDIT.
// Generated by gen-device.py from {file}, see {descriptorSource}

// {description}
//
{licenseBlock}

.syntax unified

// This is the default handler for interrupts, if triggered but not defined.
.section .text.Default_Handler
.global  Default_Handler
.type    Default_Handler, %function
Default_Handler:
    wfe
    b    Default_Handler

// Avoid the need for repeated .weak and .set instructions.
.macro IRQ handler
.weak  \\handler
.set   \\handler, Default_Handler
.endm

.section .isr_vector
.global  __isr_vector
    // Interrupt vector as defined by Cortex-M, starting with the stack top.
    // On reset, SP is initialized with *0x0 and PC is loaded with *0x4, loading
    // _stack_top and Reset_Handler.
    .long _stack_top
    .long Reset_Handler
    .long NMI_Handler
    .long HardFault_Handler
    .long MemoryManagement_Handler
    .long BusFault_Handler
    .long UsageFault_Handler
    .long 0
    .long 0
    .long 0
    .long 0
    .long SVC_Handler
    .long DebugMon_Handler
    .long 0
    .long PendSV_Handler
    .long SysTick_Handler

    // Extra interrupts for peripherals defined by the hardware vendor.
'''.format(**device.metadata))
    num = 0
    for intr in device.interrupts:
        if intr['index'] < num:
            raise ValueError('interrupt numbers are not sorted or contain a duplicate')
        while intr['index'] > num:
            out.write('    .long 0\n')
            num += 1
        num += 1
        out.write('    .long {name}_IRQHandler\n'.format(**intr))

    out.write('''
    // Define default implementations for interrupts, redirecting to
    // Default_Handler when not implemented.
    IRQ NMI_Handler
    IRQ HardFault_Handler
    IRQ MemoryManagement_Handler
    IRQ BusFault_Handler
    IRQ UsageFault_Handler
    IRQ SVC_Handler
    IRQ DebugMon_Handler
    IRQ PendSV_Handler
    IRQ SysTick_Handler
''')
    for intr in device.interrupts:
        out.write('    IRQ {name}_IRQHandler\n'.format(**intr))

def generate(indir, outdir):
    for filepath in sorted(glob(indir + '/*.svd')):
        print(filepath)
        device = readSVD(filepath)
        writeGo(outdir, device)
        writeAsm(outdir, device)


if __name__ == '__main__':
    indir = sys.argv[1] # directory with register descriptor files (*.svd, *.atdf)
    outdir = sys.argv[2] # output directory
    generate(indir, outdir)
