#!/usr/bin/python3

import sys
import os
from xml.etree import ElementTree
from glob import glob
from collections import OrderedDict
import re
import argparse

class Device:
    # dummy
    pass

def getText(element):
    return ''.join(element.itertext())

def formatText(text):
    text = re.sub('[ \t\n]+', ' ', text) # Collapse whitespace (like in HTML)
    text = text.replace('\\n ', '\n')
    text = text.strip()
    return text

def readSVD(path, sourceURL):
    # Read ARM SVD files.
    device = Device()
    xml = ElementTree.parse(path)
    root = xml.getroot()
    deviceName = getText(root.find('name'))
    deviceDescription = getText(root.find('description')).strip()
    licenseTexts = root.findall('licenseText')
    if len(licenseTexts) == 0:
        licenseText = None
    elif len(licenseTexts) == 1:
        licenseText = formatText(getText(licenseTexts[0]))
    else:
        raise ValueError('multiple <licenseText> elements')

    device.peripherals = []
    peripheralDict = {}
    groups = {}

    interrupts = OrderedDict()

    for periphEl in root.findall('./peripherals/peripheral'):
        name = getText(periphEl.find('name'))
        descriptionTags = periphEl.findall('description')
        description = ''
        if descriptionTags:
            description = formatText(getText(descriptionTags[0]))
        baseAddress = int(getText(periphEl.find('baseAddress')), 0)
        groupNameTags = periphEl.findall('groupName')
        groupName = None
        if groupNameTags:
            groupName = getText(groupNameTags[0])

        if periphEl.get('derivedFrom') or groupName in groups:
            if periphEl.get('derivedFrom'):
                derivedFromName = periphEl.get('derivedFrom')
                derivedFrom = peripheralDict[derivedFromName]
            else:
                derivedFrom = groups[groupName]
            peripheral = {
                'name':        name,
                'groupName':   derivedFrom['groupName'],
                'description': description or derivedFrom['description'],
                'baseAddress': baseAddress,
            }
            device.peripherals.append(peripheral)
            peripheralDict[name] = peripheral
            continue

        peripheral = {
            'name':        name,
            'groupName':   groupName or name,
            'description': description,
            'baseAddress': baseAddress,
            'registers':   [],
        }
        device.peripherals.append(peripheral)
        peripheralDict[name] = peripheral

        if groupName and groupName not in groups:
            groups[groupName] = peripheral

        for interrupt in periphEl.findall('interrupt'):
            intrName = getText(interrupt.find('name'))
            intrIndex = int(getText(interrupt.find('value')))
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

        regsEls = periphEl.findall('registers')
        if regsEls:
            if len(regsEls) != 1:
                raise ValueError('expected just one <registers> in a <peripheral>')
            for register in regsEls[0].findall('register'):
                peripheral['registers'].append(parseRegister(groupName or name, register, baseAddress))
            for cluster in regsEls[0].findall('cluster'):
                if cluster.find('dim') is not None:
                    continue # TODO
                clusterPrefix = getText(cluster.find('name')) + '_'
                clusterOffset = int(getText(cluster.find('addressOffset')), 0)
                for regEl in cluster.findall('register'):
                    peripheral['registers'].append(parseRegister(groupName or name, regEl, baseAddress + clusterOffset, clusterPrefix))
                else:
                    continue

    device.interrupts = sorted(interrupts.values(), key=lambda v: v['index'])
    licenseBlock = ''
    if licenseText is not None:
        licenseBlock = '//     ' + licenseText.replace('\n', '\n//     ')
        licenseBlock = '\n'.join(map(str.rstrip, licenseBlock.split('\n'))) # strip trailing whitespace
    device.metadata = {
        'file':             os.path.basename(path),
        'descriptorSource': sourceURL,
        'name':             deviceName,
        'nameLower':        deviceName.lower(),
        'description':      deviceDescription,
        'licenseBlock':     licenseBlock,
    }

    return device

def parseRegister(groupName, regEl, baseAddress, namePrefix=''):
    regName = getText(regEl.find('name'))
    regDescription = getText(regEl.find('description'))
    offsetEls = regEl.findall('offset')
    if not offsetEls:
        offsetEls = regEl.findall('addressOffset')
    address = baseAddress + int(getText(offsetEls[0]), 0)

    dimEls = regEl.findall('dim')
    array = None
    if dimEls:
        array = int(getText(dimEls[0]), 0)
        regName = regName.replace('[%s]', '')

    fields = []
    fieldsEls = regEl.findall('fields')
    if fieldsEls:
        for fieldEl in fieldsEls[0].findall('field'):
            fieldName = getText(fieldEl.find('name'))
            descrEls = fieldEl.findall('description')
            lsbTags = fieldEl.findall('lsb')
            if len(lsbTags) == 1:
                lsb = int(getText(lsbTags[0]))
            else:
                lsb = int(getText(fieldEl.find('bitOffset')))
            msbTags = fieldEl.findall('msb')
            if len(msbTags) == 1:
                msb = int(getText(msbTags[0]))
            else:
                msb = int(getText(fieldEl.find('bitWidth'))) + lsb - 1
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
            if lsb == msb: # single bit
                fields.append({
                    'name':        '{}_{}{}_{}'.format(groupName, namePrefix, regName, fieldName),
                    'description': 'Bit %s.' % fieldName,
                    'value':       1 << lsb,
                })
            for enumEl in fieldEl.findall('enumeratedValues/enumeratedValue'):
                enumName = getText(enumEl.find('name'))
                enumDescription = getText(enumEl.find('description'))
                enumValue = int(getText(enumEl.find('value')), 0)
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
// Generated by gen-device-svd.py from {file}, see {descriptorSource}

// +build {pkgName},{nameLower}

// {description}
//
{licenseBlock}
package {pkgName}

import "unsafe"

// Special type that causes loads/stores to be volatile (necessary for
// memory-mapped registers).
//go:volatile
type RegValue uint32

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

    # Define actual peripheral pointers.
    out.write('\n// Peripherals.\nvar (\n')
    for peripheral in device.peripherals:
        out.write('\t{name} = (*{groupName}_Type)(unsafe.Pointer(uintptr(0x{baseAddress:x}))) // {description}\n'.format(**peripheral))
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
                    out.write('\t_padding{padNumber} RegValue\n'.format(padNumber=padNumber))
                else:
                    out.write('\t_padding{padNumber} [{num}]RegValue\n'.format(padNumber=padNumber, num=numSkip))
                padNumber += 1

            regType = 'RegValue'
            if register['array'] is not None:
                regType = '[{}]RegValue'.format(register['array'])
            out.write('\t{name} {regType}\n'.format(**register, regType=regType))

            # next address
            if register['array'] is not None:
                address = register['address'] + 4 * register['array']
            else:
                address = register['address'] + 4
        out.write('}\n')

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
// Generated by gen-device-svd.py from {file}, see {descriptorSource}

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

def generate(indir, outdir, sourceURL):
    if not os.path.isdir(indir):
        print('cannot find input directory:', indir, file=sys.stderr)
        sys.exit(1)
    if not os.path.isdir(outdir):
        os.mkdir(outdir)
    infiles = glob(indir + '/*.svd')
    if not infiles:
        print('no .svd files found:', indir, file=sys.stderr)
        sys.exit(1)
    for filepath in sorted(infiles):
        print(filepath)
        device = readSVD(filepath, sourceURL)
        writeGo(outdir, device)
        writeAsm(outdir, device)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Generate Go register descriptors and interrupt vectors from .svd files')
    parser.add_argument('indir', metavar='indir', type=str,
                        help='input directory containing .svd files')
    parser.add_argument('outdir', metavar='outdir', type=str,
                        help='output directory')
    parser.add_argument('--source', metavar='source', type=str,
                        help='output directory',
                        default='<unknown>')
    args = parser.parse_args()
    generate(args.indir, args.outdir, args.source)
