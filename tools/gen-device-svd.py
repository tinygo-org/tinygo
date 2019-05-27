#!/usr/bin/env python3

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
    if element is None:
        return "None"
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

        interruptEls = periphEl.findall('interrupt')
        for interrupt in interruptEls:
            intrName = getText(interrupt.find('name'))
            intrIndex = int(getText(interrupt.find('value')))
            addInterrupt(interrupts, intrName, intrIndex, description)
            # As a convenience, also use the peripheral name as the interrupt
            # name. Only do that for the nrf for now, as the stm32 .svd files
            # don't always put interrupts in the correct peripheral...
            if len(interruptEls) == 1 and deviceName.startswith('nrf'):
                addInterrupt(interrupts, name, intrIndex, description)

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
            if 'subtypes' in derivedFrom:
                for subtype in derivedFrom['subtypes']:
                    subp = {
                        'name':        name + "_"+subtype['clusterName'],
                        'groupName':   subtype['groupName'],
                        'description': subtype['description'],
                        'baseAddress': baseAddress,
                    }
                    device.peripherals.append(subp)
            continue

        peripheral = {
            'name':        name,
            'groupName':   groupName or name,
            'description': description,
            'baseAddress': baseAddress,
            'registers':   [],
            'subtypes':   [],
        }
        device.peripherals.append(peripheral)
        peripheralDict[name] = peripheral

        if groupName and groupName not in groups:
            groups[groupName] = peripheral

        regsEls = periphEl.findall('registers')
        if regsEls:
            if len(regsEls) != 1:
                raise ValueError('expected just one <registers> in a <peripheral>')
            for register in regsEls[0].findall('register'):
                peripheral['registers'].extend(parseRegister(groupName or name, register, baseAddress))
            for cluster in regsEls[0].findall('cluster'):
                clusterName = getText(cluster.find('name')).replace('[%s]', '')
                clusterDescription = getText(cluster.find('description'))
                clusterPrefix = clusterName + '_'
                clusterOffset = int(getText(cluster.find('addressOffset')), 0)
                if cluster.find('dim') is None:
                    if clusterOffset is 0:
                        # make this a separate peripheral
                        cpRegisters = []
                        for regEl in cluster.findall('register'):
                            cpRegisters.extend(parseRegister(groupName, regEl, baseAddress, clusterName+"_"))
                        cpRegisters.sort(key=lambda r: r['address'])
                        clusterPeripheral = {
                            'name':        name+ "_" +clusterName,
                            'groupName':   groupName+ "_" +clusterName,
                            'description': description+ " - " +clusterName,
                            'clusterName': clusterName,
                            'baseAddress': baseAddress,
                            'registers':   cpRegisters,
                        }
                        device.peripherals.append(clusterPeripheral)
                        peripheral['subtypes'].append(clusterPeripheral)
                        continue
                    dim = None
                    dimIncrement = None
                else:
                    dim = int(getText(cluster.find('dim')))
                    dimIncrement = int(getText(cluster.find('dimIncrement')), 0)
                clusterRegisters = []
                for regEl in cluster.findall('register'):
                    clusterRegisters.extend(parseRegister(groupName or name, regEl, baseAddress + clusterOffset, clusterPrefix))
                clusterRegisters.sort(key=lambda r: r['address'])
                if dimIncrement is None:
                    lastReg = clusterRegisters[-1]
                    lastAddress = lastReg['address']
                    if lastReg['array'] is not None:
                        lastAddress = lastReg['address'] + lastReg['array'] * lastReg['elementsize']
                    firstAddress = clusterRegisters[0]['address']
                    dimIncrement = lastAddress - firstAddress
                peripheral['registers'].append({
                    'name':        clusterName,
                    'address':     baseAddress + clusterOffset,
                    'description': clusterDescription,
                    'registers':   clusterRegisters,
                    'array':       dim,
                    'elementsize': dimIncrement,
                })
        peripheral['registers'].sort(key=lambda r: r['address'])

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

def addInterrupt(interrupts, intrName, intrIndex, description):
    if intrName in interrupts:
        if interrupts[intrName]['index'] != intrIndex:
            raise ValueError('interrupt with the same name has different indexes: %s (%d vs %d)'
                % (intrName, interrupts[intrName]['index'], intrIndex))
        if description not in interrupts[intrName]['description'].split(' // '):
            interrupts[intrName]['description'] += ' // ' + description
    else:
        interrupts[intrName] = {
            'name':        intrName,
            'index':       intrIndex,
            'description': description,
        }

def parseBitfields(groupName, regName, fieldsEls, bitfieldPrefix=''):
    fields = []
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
                'name':        '{}_{}{}_{}_Pos'.format(groupName, bitfieldPrefix, regName, fieldName),
                'description': 'Position of %s field.' % fieldName,
                'value':       lsb,
            })
            fields.append({
                'name':        '{}_{}{}_{}_Msk'.format(groupName, bitfieldPrefix, regName, fieldName),
                'description': 'Bit mask of %s field.' % fieldName,
                'value':       (0xffffffff >> (31 - (msb - lsb))) << lsb,
            })
            if lsb == msb: # single bit
                fields.append({
                    'name':        '{}_{}{}_{}'.format(groupName, bitfieldPrefix, regName, fieldName),
                    'description': 'Bit %s.' % fieldName,
                    'value':       1 << lsb,
                })
            for enumEl in fieldEl.findall('enumeratedValues/enumeratedValue'):
                enumName = getText(enumEl.find('name'))
                enumDescription = getText(enumEl.find('description'))
                enumValue = int(getText(enumEl.find('value')), 0)
                fields.append({
                    'name':        '{}_{}{}_{}_{}'.format(groupName, bitfieldPrefix, regName, fieldName, enumName),
                    'description': enumDescription,
                    'value':       enumValue,
                })
    return fields

def parseRegister(groupName, regEl, baseAddress, bitfieldPrefix=''):
    regName = getText(regEl.find('name'))
    regDescription = getText(regEl.find('description'))
    offsetEls = regEl.findall('offset')
    if not offsetEls:
        offsetEls = regEl.findall('addressOffset')
    address = baseAddress + int(getText(offsetEls[0]), 0)
    
    size = 4
    elSizes = regEl.findall('size')
    if elSizes:
        size = int(getText(elSizes[0]), 0) // 8
    
    dimEls = regEl.findall('dim')
    fieldsEls = regEl.findall('fields')

    array = None
    if dimEls:
        array = int(getText(dimEls[0]), 0)
        dimIncrement = int(getText(regEl.find('dimIncrement')), 0)
        if "[%s]" in regName:
            # just a normal array of registers
            regName = regName.replace('[%s]', '')
        elif "%s" in regName:
            # a "spaced array" of registers, special processing required
            # we need to generate a separate register for each "element"
            results = []
            for i in range(array):
                regAddress = address + (i * dimIncrement)
                results.append({
                    'name':        regName.replace('%s', str(i)),
                    'address':     regAddress,
                    'description': regDescription.replace('\n', ' '),
                    'bitfields':   [],
                    'array':       None,
                    'elementsize': size,
                })
            # set first result bitfield
            shortName = regName.replace('_%s', '').replace('%s', '')
            results[0]['bitfields'] = parseBitfields(groupName, shortName, fieldsEls, bitfieldPrefix)
            return results

    return [{
        'name':        regName,
        'address':     address,
        'description': regDescription.replace('\n', ' '),
        'bitfields':   parseBitfields(groupName, regName, fieldsEls, bitfieldPrefix),
        'array':       array,
        'elementsize': size,
    }]

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

import (
	"runtime/volatile"
	"unsafe"
)

// Special types that causes loads/stores to be volatile (necessary for
// memory-mapped registers).
type Register8 struct {{
	Reg uint8
}}

// Get returns the value in the register. It is the volatile equivalent of:
//
//     *r.Reg
//
//go:inline
func (r *Register8) Get() uint8 {{
	return volatile.LoadUint8(&r.Reg)
}}

// Set updates the register value. It is the volatile equivalent of:
//
//     *r.Reg = value
//
//go:inline
func (r *Register8) Set(value uint8) {{
	volatile.StoreUint8(&r.Reg, value)
}}

// SetBits reads the register, sets the given bits, and writes it back. It is
// the volatile equivalent of:
//
//     r.Reg |= value
//
//go:inline
func (r *Register8) SetBits(value uint8) {{
	volatile.StoreUint8(&r.Reg, volatile.LoadUint8(&r.Reg) | value)
}}

// ClearBits reads the register, clears the given bits, and writes it back. It
// is the volatile equivalent of:
//
//     r.Reg &^= value
//
//go:inline
func (r *Register8) ClearBits(value uint8) {{
	volatile.StoreUint8(&r.Reg, volatile.LoadUint8(&r.Reg) &^ value)
}}

// HasBits reads the register and then checks to see if the passed bits are set. It
// is the volatile equivalent of:
//
//     (*r.Reg & value) > 0
//
//go:inline
func (r *Register8) HasBits(value uint8) bool {{
	return (r.Get() & value) > 0
}}

type Register16 struct {{
	Reg uint16
}}

// Get returns the value in the register. It is the volatile equivalent of:
//
//     *r.Reg
//
//go:inline
func (r *Register16) Get() uint16 {{
	return volatile.LoadUint16(&r.Reg)
}}

// Set updates the register value. It is the volatile equivalent of:
//
//     *r.Reg = value
//
//go:inline
func (r *Register16) Set(value uint16) {{
	volatile.StoreUint16(&r.Reg, value)
}}

// SetBits reads the register, sets the given bits, and writes it back. It is
// the volatile equivalent of:
//
//     r.Reg |= value
//
//go:inline
func (r *Register16) SetBits(value uint16) {{
	volatile.StoreUint16(&r.Reg, volatile.LoadUint16(&r.Reg) | value)
}}

// ClearBits reads the register, clears the given bits, and writes it back. It
// is the volatile equivalent of:
//
//     r.Reg &^= value
//
//go:inline
func (r *Register16) ClearBits(value uint16) {{
	volatile.StoreUint16(&r.Reg, volatile.LoadUint16(&r.Reg) &^ value)
}}

// HasBits reads the register and then checks to see if the passed bits are set. It
// is the volatile equivalent of:
//
//     (*r.Reg & value) > 0
//
//go:inline
func (r *Register16) HasBits(value uint16) bool {{
	return (r.Get() & value) > 0
}}

type Register32 struct {{
	Reg uint32
}}

// Get returns the value in the register. It is the volatile equivalent of:
//
//     *r.Reg
//
//go:inline
func (r *Register32) Get() uint32 {{
	return volatile.LoadUint32(&r.Reg)
}}

// Set updates the register value. It is the volatile equivalent of:
//
//     *r.Reg = value
//
//go:inline
func (r *Register32) Set(value uint32) {{
	volatile.StoreUint32(&r.Reg, value)
}}

// SetBits reads the register, sets the given bits, and writes it back. It is
// the volatile equivalent of:
//
//     r.Reg |= value
//
//go:inline
func (r *Register32) SetBits(value uint32) {{
	volatile.StoreUint32(&r.Reg, volatile.LoadUint32(&r.Reg) | value)
}}

// ClearBits reads the register, clears the given bits, and writes it back. It
// is the volatile equivalent of:
//
//     r.Reg &^= value
//
//go:inline
func (r *Register32) ClearBits(value uint32) {{
	volatile.StoreUint32(&r.Reg, volatile.LoadUint32(&r.Reg) &^ value)
}}

// HasBits reads the register and then checks to see if the passed bits are set. It
// is the volatile equivalent of:
//
//     (*r.Reg & value) > 0
//
//go:inline
func (r *Register32) HasBits(value uint32) bool {{
	return (r.Get() & value) > 0
}}

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
            if address > register['address'] and 'registers' not in register :
                # In Nordic SVD files, these registers are deprecated or
                # duplicates, so can be ignored.
                #print('skip: %s.%s %s - %s %s' % (peripheral['name'], register['name'], address, register['address'], register['elementsize']))
                continue
            eSize = register['elementsize']
            if eSize == 4:
                regType = 'Register32'
            elif eSize == 2:
                regType = 'Register16'
            elif eSize == 1:
                regType = 'Register8'        
            else:
                eSize = 4
                regType = 'Register32'

            # insert padding, if needed
            if address < register['address']:
                bytesNeeded = register['address'] - address
                if bytesNeeded == 1:
                    out.write('\t_padding{padNumber} {regType}\n'.format(padNumber=padNumber, regType='Register8'))
                elif bytesNeeded == 2:
                    out.write('\t_padding{padNumber} {regType}\n'.format(padNumber=padNumber, regType='Register16'))
                else:
                    numSkip = (register['address'] - address) // eSize
                    if numSkip == 1:
                        out.write('\t_padding{padNumber} {regType}\n'.format(padNumber=padNumber, regType=regType))
                    else:
                        out.write('\t_padding{padNumber} [{num}]{regType}\n'.format(padNumber=padNumber, num=numSkip, regType=regType))
                padNumber += 1
                address = register['address']

            lastCluster = False
            if 'registers' in register:
                # This is a cluster, not a register. Create the cluster type.
                regType = 'struct {\n'
                subaddress = register['address']
                for subregister in register['registers']:
                    if subregister['elementsize'] == 4:
                        subregType = 'Register32'
                    elif subregister['elementsize'] == 2:
                        subregType = 'Register16'
                    else:
                        subregType = 'Register8'

                    if subregister['array']:
                        subregType = '[{}]{}'.format(subregister['array'], subregType)
                    if subaddress != subregister['address']:
                        bytesNeeded = subregister['address'] - subaddress
                        if bytesNeeded == 1:
                            regType += '\t\t_padding{padNumber} {subregType}\n'.format(padNumber=padNumber, subregType='Register8')
                        elif bytesNeeded == 2:
                            regType += '\t\t_padding{padNumber} {subregType}\n'.format(padNumber=padNumber, subregType='Register16')
                        else:
                            numSkip = (subregister['address'] - subaddress)
                            if numSkip < 1:
                                continue
                            elif numSkip == 1:
                                regType += '\t\t_padding{padNumber} {subregType}\n'.format(padNumber=padNumber, subregType='Register8')
                            else:
                                regType += '\t\t_padding{padNumber} [{num}]{subregType}\n'.format(padNumber=padNumber, num=numSkip, subregType='Register8')
                        padNumber += 1
                        subaddress += bytesNeeded
                    if subregister['array'] is not None:
                        subaddress += subregister['elementsize'] * subregister['array']
                    else:
                        subaddress += subregister['elementsize']
                    regType += '\t\t{name} {subregType}\n'.format(name=subregister['name'], subregType=subregType)
                if register['array'] is not None:
                    if subaddress != register['address'] + register['elementsize']:
                        numSkip = ((register['address'] + register['elementsize']) - subaddress) // 4
                        if numSkip <= 1:
                            regType += '\t\t_padding{padNumber} {subregType}\n'.format(padNumber=padNumber, subregType=subregType)
                        else:
                            regType += '\t\t_padding{padNumber} [{num}]{subregType}\n'.format(padNumber=padNumber, num=numSkip, subregType=subregType)
                else:
                    lastCluster = True
                regType += '\t}'
                address = subaddress
            if register['array'] is not None:
                regType = '[{}]{}'.format(register['array'], regType)
            out.write('\t{name} {regType}\n'.format(name=register['name'], regType=regType))

            # next address
            if lastCluster is True:
                lastCluster = False
            elif register['array'] is not None:
                address = register['address'] + register['elementsize'] * register['array']
            else:
                address = register['address'] + register['elementsize']
        out.write('}\n')

    # Define bitfields.
    for peripheral in device.peripherals:
        if 'registers' not in peripheral:
            # This peripheral was derived from another peripheral. Bitfields are
            # already defined.
            continue
        out.write('\n// Bitfields for {name}: {description}\nconst('.format(**peripheral))
        for register in peripheral['registers']:
            if register.get('bitfields'):
                writeGoRegisterBitfields(out, register, register['name'])
            for subregister in register.get('registers', []):
                writeGoRegisterBitfields(out, subregister, register['name'] + '.' + subregister['name'])
        out.write(')\n')

def writeGoRegisterBitfields(out, register, name):
    out.write('\n\t// {}'.format(name))
    if register['description']:
        out.write(': {description}'.format(**register))
    out.write('\n')
    for bitfield in register['bitfields']:
        out.write('\t{name} = 0x{value:x}'.format(**bitfield))
        if bitfield['description']:
            out.write(' // {description}'.format(**bitfield))
        out.write('\n')


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

// Must set the "a" flag on the section:
// https://svnweb.freebsd.org/base/stable/11/sys/arm/arm/locore-v4.S?r1=321049&r2=321048&pathrev=321049
// https://sourceware.org/binutils/docs/as/Section.html#ELF-Version
.section .isr_vector, "a", %progbits
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
        if intr['index'] == num - 1:
            continue
        if intr['index'] < num:
            raise ValueError('interrupt numbers are not sorted')
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
