#!/usr/bin/env python

from __future__ import print_function

import sys
import os
from xml.etree import ElementTree
from glob import glob
from collections import OrderedDict
import re
import argparse

validName = re.compile('^[a-zA-Z0-9_]+$')


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

# Replace characters that are not allowed in a symbol name with a '_'. This is
# useful to be able to process SVD files with errors.
def cleanName(text):
    if not validName.match(text):
        return ''.join(list(map(lambda c: c if validName.match(c) else '_', text)))
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
            # Some group names (for example the STM32H7A3x) have an invalid
            # group name. Replace invalid characters with '_'.
            groupName = cleanName(getText(groupNameTags[0]))

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
                if cluster.find('dimIndex') is not None:
                    clusterName = clusterName.replace('%s', '')
                clusterDescription = getText(cluster.find('description'))
                clusterPrefix = clusterName + '_'
                clusterOffset = int(getText(cluster.find('addressOffset')), 0)
                if cluster.find('dim') is None:
                    if clusterOffset == 0:
                        # make this a separate peripheral
                        cpRegisters = []
                        for regEl in cluster.findall('register'):
                            cpRegisters.extend(parseRegister(groupName, regEl, baseAddress, clusterName+"_"))
                        # handle sub-clusters of registers
                        for subClusterEl in cluster.findall('cluster'):
                            subclusterName = getText(subClusterEl.find('name')).replace('[%s]', '')
                            subclusterDescription = getText(subClusterEl.find('description'))
                            subclusterPrefix = subclusterName + '_'
                            subclusterOffset = int(getText(subClusterEl.find('addressOffset')), 0)
                            subdim = int(getText(subClusterEl.find('dim')))
                            subdimIncrement = int(getText(subClusterEl.find('dimIncrement')), 16)

                            if subdim > 1:
                                subcpRegisters = []
                                subregSize = 0
                                for regEl in subClusterEl.findall('register'):
                                    subregSize += int(getText(regEl.find('size')))
                                    subcpRegisters.extend(parseRegister(groupName, regEl, baseAddress + subclusterOffset, subclusterPrefix))
                                cpRegisters.append({
                                    'name':        subclusterName,
                                    'address':     baseAddress + subclusterOffset,
                                    'description': subclusterDescription,
                                    'registers':   subcpRegisters,
                                    'array':       subdim,
                                    'elementsize': subdimIncrement,
                                })
                            else:
                                for regEl in subClusterEl.findall('register'):
                                    cpRegisters.extend(parseRegister(getText(regEl.find('name')), regEl, baseAddress + subclusterOffset, subclusterPrefix))

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
                    if dim == 1:
                        dimIncrement = None
                    else:
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
            # Note: some SVD files like the one for STM32H7x7 contain mistakes.
            # Instead of throwing an error, simply log it.
            print ('interrupt with the same name has different indexes: %s (%d vs %d)'
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
            # Some bitfields (like the STM32H7x7) contain invalid bitfield
            # names like 'CNT[31]'. Replace invalid characters with '_' when
            # needed.
            fieldName = cleanName(getText(fieldEl.find('name')))
            if not fieldName[0].isupper() and not fieldName[0].isdigit():
                fieldName = fieldName.upper()
            if len(fieldEl.findall('lsb')) == 1 and len(fieldEl.findall('msb')) == 1:
                # try to use lsb/msb tags
                lsb = int(getText(fieldEl.findall('lsb')[0]))
                msb = int(getText(fieldEl.findall('msb')[0]))
            elif len(fieldEl.findall('bitOffset')) > 0 and len(fieldEl.findall('bitWidth')) > 0:
                # try to use bitOffset/bitWidth tags
                lsb = int(getText(fieldEl.find('bitOffset')))
                msb = int(getText(fieldEl.find('bitWidth'))) + lsb - 1
            elif len(fieldEl.findall('bitRange')) > 0:
                # try use bitRange
                bitRangeTags = fieldEl.findall('bitRange')
                lsb = int(getText(bitRangeTags[0]).split(":")[1][:-1])
                msb = int(getText(bitRangeTags[0]).split(":")[0][1:])
            else:
                # this is an error. what to do?
                print("unable to find lsb/msb in field:", fieldName)

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
                if not enumName[0].isupper() and not enumName[0].isdigit():
                    enumName = enumName.upper()
                enumDescription = getText(enumEl.find('description')).replace('\n', ' ')
                enumValue = int(getText(enumEl.find('value')), 0)
                fields.append({
                    'name':        '{}_{}{}_{}_{}'.format(groupName, bitfieldPrefix, regName, fieldName, enumName),
                    'description': enumDescription,
                    'value':       enumValue,
                })
    return fields

class Register:
    def __init__(self, element, baseAddress):
        self.element = element
        self.baseAddress = baseAddress

    def name(self):
        return getText(self.element.find('name')).replace('[%s]', '')

    def description(self):
        return getText(self.element.find('description')).replace('\n', ' ')

    def address(self):
        offsetEls = self.element.findall('offset')
        if not offsetEls:
            offsetEls = self.element.findall('addressOffset')
        return self.baseAddress + int(getText(offsetEls[0]), 0)

    def dim(self):
        dimEls = self.element.findall('dim')
        if len(dimEls) == 0:
            return None
        elif len(dimEls) == 1:
            return int(getText(dimEls[0]), 0)
        else:
            raise ValueError('expected at most one <dim> element in %s register' % self.name())

    def size(self):
        size = 4
        elSizes = self.element.findall('size')
        if elSizes:
            size = int(getText(elSizes[0]), 0) // 8
        return size


def parseRegister(groupName, regEl, baseAddress, bitfieldPrefix=''):
    reg = Register(regEl, baseAddress)

    fieldsEls = regEl.findall('fields')

    if reg.dim() is not None:
        dimIncrement = int(getText(regEl.find('dimIncrement')), 0)
        if "%s" in reg.name():
            # a "spaced array" of registers, special processing required
            # we need to generate a separate register for each "element"
            results = []
            for i in range(reg.dim()):
                regAddress = reg.address() + (i * dimIncrement)
                results.append({
                    'name':        reg.name().replace('%s', str(i)),
                    'address':     regAddress,
                    'description': reg.description(),
                    'bitfields':   [],
                    'array':       None,
                    'elementsize': reg.size(),
                })
            # set first result bitfield
            shortName = reg.name().replace('_%s', '').replace('%s', '').upper()
            results[0]['bitfields'] = parseBitfields(groupName, shortName, fieldsEls, bitfieldPrefix)
            return results
    regName = reg.name()
    if not regName[0].isupper() and not regName[0].isdigit():
        regName = regName.upper()
    return [{
        'name':        regName,
        'address':     reg.address(),
        'description': reg.description(),
        'bitfields':   parseBitfields(groupName, regName, fieldsEls, bitfieldPrefix),
        'array':       reg.dim(),
        'elementsize': reg.size(),
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
        for register in peripheral['registers']:
            if address > register['address'] and 'registers' not in register :
                # In Nordic SVD files, these registers are deprecated or
                # duplicates, so can be ignored.
                #print('skip: %s.%s %s - %s %s' % (peripheral['name'], register['name'], address, register['address'], register['elementsize']))
                continue
            eSize = register['elementsize']
            if eSize == 4:
                regType = 'volatile.Register32'
            elif eSize == 2:
                regType = 'volatile.Register16'
            elif eSize == 1:
                regType = 'volatile.Register8'
            else:
                eSize = 4
                regType = 'volatile.Register32'

            # insert padding, if needed
            if address < register['address']:
                bytesNeeded = register['address'] - address
                if bytesNeeded == 1:
                    out.write('\t_ {regType}\n'.format(regType='volatile.Register8'))
                elif bytesNeeded == 2:
                    out.write('\t_ {regType}\n'.format(regType='volatile.Register16'))
                elif bytesNeeded == 3:
                    out.write('\t_ [3]{regType}\n'.format(regType='volatile.Register8'))
                else:
                    numSkip = (register['address'] - address) // eSize
                    if numSkip == 1:
                        out.write('\t_ {regType}\n'.format(regType=regType))
                    else:
                        out.write('\t_ [{num}]{regType}\n'.format(num=numSkip, regType=regType))
                address = register['address']

            lastCluster = False
            if 'registers' in register:
                # This is a cluster, not a register. Create the cluster type.
                regType = 'struct {\n'
                subaddress = register['address']
                for subregister in register['registers']:
                    if subregister['elementsize'] == 4:
                        subregType = 'volatile.Register32'
                    elif subregister['elementsize'] == 2:
                        subregType = 'volatile.Register16'
                    elif subregister['elementsize'] == 1:
                        subregType = 'volatile.Register8'

                    if subregister['array']:
                        subregType = '[{}]{}'.format(subregister['array'], subregType)
                    if subaddress != subregister['address']:
                        bytesNeeded = subregister['address'] - subaddress
                        if bytesNeeded == 1:
                            regType += '\t\t_ {subregType}\n'.format(subregType='volatile.Register8')
                        elif bytesNeeded == 2:
                            regType += '\t\t_ {subregType}\n'.format(subregType='volatile.Register16')
                        else:
                            numSkip = (subregister['address'] - subaddress)
                            if numSkip < 1:
                                continue
                            elif numSkip == 1:
                                regType += '\t\t_ {subregType}\n'.format(subregType='volatile.Register8')
                            else:
                                regType += '\t\t_ [{num}]{subregType}\n'.format(num=numSkip, subregType='volatile.Register8')
                        subaddress += bytesNeeded
                    if subregister['array'] is not None:
                        subregSize = subregister['array'] * subregister['elementsize']
                    else:
                        subregSize = subregister['elementsize']
                    subaddress += subregSize
                    regType += '\t\t{name} {subregType}\n'.format(name=subregister['name'], subregType=subregType)
                if register['array'] is not None:
                    if subaddress != register['address'] + register['elementsize']:
                        numSkip = ((register['address'] + register['elementsize']) - subaddress) // subregSize
                        if numSkip <= 1:
                            regType += '\t\t_ {subregType}\n'.format(subregType=subregType)
                        else:
                            regType += '\t\t_ [{num}]{subregType}\n'.format(num=numSkip, subregType=subregType)
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
