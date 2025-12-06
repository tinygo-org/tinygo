//go:build (gc.conservative || gc.precise) && scheduler.threads

package runtime

import "unsafe"

//export tinygo_scan_list
func getScanList() **objHeader

func (b gcBlock) stateAtomic() blockState {
	return b.stateFromByte(b.stateByteAtomic())
}

func (b gcBlock) stateByteAtomic() byte {
	return atomicLoad8((*uint8)(unsafe.Add(metadataStart, b/blocksPerStateByte)))
}

func atomicLoad8(ptr *uint8) uint8

func (b gcBlock) mark() bool {
	mask := byte(blockStateMark) << ((b % blocksPerStateByte) * stateBits)
	return mask&^atomicOr8((*uint8)(unsafe.Add(metadataStart, b/blocksPerStateByte)), mask) != 0
}

func atomicOr8(ptr *uint8, mask uint8) uint8
