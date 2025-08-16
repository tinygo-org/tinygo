package uefi

import (
	"errors"
	"unsafe"
)

// Errors Disk IO and Block IO can return
var (
	ErrDiskNotReady   = errors.New("disk not ready")
	ErrNoMediaPresent = errors.New("no media present")
	ErrNegativeOffset = errors.New("negative offset")
	ErrReadOnly       = errors.New("media is read-only")
)

//---------------------------------------------------------------------------
// GUIDs
//---------------------------------------------------------------------------

// EFI_DISK_IO_PROTOCOL_GUID = {CE345171-BA0B-11d2-8E4F-00A0C969723B}
var EFI_DISK_IO_PROTOCOL_GUID = EFI_GUID{
	0xCE345171, 0xBA0B, 0x11D2,
	[8]uint8{0x8e, 0x4f, 0x00, 0xa0, 0xc9, 0x69, 0x72, 0x3b},
}

// EFI_BLOCK_IO_PROTOCOL_GUID = {964E5B21-6459-11d2-8E39-00A0C969723B}
var EFI_BLOCK_IO_PROTOCOL_GUID = EFI_GUID{
	0x964E5B21, 0x6459, 0x11D2,
	[8]uint8{0x8e, 0x39, 0x00, 0xa0, 0xc9, 0x69, 0x72, 0x3b},
}

// Classic subset of MEDIA fields (UEFI 2.x adds more tail fields; we keep the common prefix)
type EFI_BLOCK_IO_MEDIA struct {
	MediaId          uint32
	RemovableMedia   bool
	MediaPresent     bool
	LogicalPartition bool
	ReadOnly         bool
	WriteCaching     bool
	BlockSize        uint32
	IoAlign          uint32
	LastBlock        EFI_LBA
	// Newer fields exist after this point; we intentionally omit for compatibility.
}

// §Block I/O function table order: Revision, Media, Reset, ReadBlocks, WriteBlocks, FlushBlocks
type EFI_BLOCK_IO_PROTOCOL struct {
	Revision    uint64
	Media       *EFI_BLOCK_IO_MEDIA
	reset       uintptr // (this, ExtendedVerification: bool)
	readBlocks  uintptr // (this, MediaId, LBA, BufferSize, Buffer)
	writeBlocks uintptr // (this, MediaId, LBA, BufferSize, Buffer)
	flushBlocks uintptr // (this)
}

// Reset the device (optionally extended verification)
func (p *EFI_BLOCK_IO_PROTOCOL) Reset(extendedVerification bool) EFI_STATUS {
	return UefiCall2(p.reset, uintptr(unsafe.Pointer(p)), convertBool(extendedVerification))
}

// ReadBlocks reads raw LBAs into Buffer (size in bytes).
func (p *EFI_BLOCK_IO_PROTOCOL) ReadBlocks(mediaId uint32, lba EFI_LBA, bufSize UINTN, buffer unsafe.Pointer) EFI_STATUS {
	return UefiCall5(
		p.readBlocks,
		uintptr(unsafe.Pointer(p)),
		uintptr(mediaId),
		uintptr(lba),
		uintptr(bufSize),
		uintptr(buffer),
	)
}


// WriteBlocks writes raw LBAs from Buffer (size in bytes).
func (p *EFI_BLOCK_IO_PROTOCOL) WriteBlocks(mediaId uint32, lba EFI_LBA, bufSize UINTN, buffer unsafe.Pointer) EFI_STATUS {
	return UefiCall5(
		p.writeBlocks,
		uintptr(unsafe.Pointer(p)),
		uintptr(mediaId),
		uintptr(lba),
		uintptr(bufSize),
		uintptr(buffer),
	)
}

// FlushBlocks flushes any device caches.
func (p *EFI_BLOCK_IO_PROTOCOL) FlushBlocks() EFI_STATUS {
	return UefiCall1(p.flushBlocks, uintptr(unsafe.Pointer(p)))
}

//---------------------------------------------------------------------------
// Disk I/O: byte-addressed wrapper over Block I/O
//---------------------------------------------------------------------------

// §Disk I/O function table order: Revision, ReadDisk, WriteDisk
type EFI_DISK_IO_PROTOCOL struct {
	Revision  uint64
	readDisk  uintptr // (this, MediaId, Offset, BufferSize, Buffer)
	writeDisk uintptr // (this, MediaId, Offset, BufferSize, Buffer)
}

// ReadDisk reads BufferSize bytes at byte Offset into Buffer.
func (p *EFI_DISK_IO_PROTOCOL) ReadDisk(mediaId uint32, offset uint64, bufSize UINTN, buffer unsafe.Pointer) EFI_STATUS {
	return UefiCall5(
		p.readDisk,
		uintptr(unsafe.Pointer(p)),
		uintptr(mediaId),
		uintptr(offset),
		uintptr(bufSize),
		uintptr(buffer),
	)
}

// WriteDisk writes BufferSize bytes at byte Offset from Buffer.
func (p *EFI_DISK_IO_PROTOCOL) WriteDisk(mediaId uint32, offset uint64, bufSize UINTN, buffer unsafe.Pointer) EFI_STATUS {
	return UefiCall5(
		p.writeDisk,
		uintptr(unsafe.Pointer(p)),
		uintptr(mediaId),
		uintptr(offset),
		uintptr(bufSize),
		uintptr(buffer),
	)
}

//---------------------------------------------------------------------------
// Helpers: locate Disk I/O + Block I/O, and a tiny io.ReaderAt/io.WriterAt adapter
//---------------------------------------------------------------------------

// Disk bundles Disk I/O with its backing Block I/O (for MediaId, size, alignment).
type Disk struct {
	DiskIO  *EFI_DISK_IO_PROTOCOL
	BlockIO *EFI_BLOCK_IO_PROTOCOL
}

// EnumerateDisks returns Disk adapters for every handle that exposes both Disk I/O and Block I/O.
func EnumerateDisks() ([]*Disk, error) {
	var (
		handleCount  UINTN
		handleBuffer *EFI_HANDLE
	)
	st := BS().LocateHandleBuffer(ByProtocol, &EFI_DISK_IO_PROTOCOL_GUID, nil, &handleCount, &handleBuffer)
	if st == EFI_NOT_FOUND {
		return nil, nil
	}
	if st != EFI_SUCCESS {
		return nil, StatusError(st)
	}
	handles := unsafe.Slice((*EFI_HANDLE)(unsafe.Pointer(handleBuffer)), int(handleCount))

	disks := make([]*Disk, 0, len(handles))
	for _, h := range handles {
		var dio *EFI_DISK_IO_PROTOCOL
		if st = BS().HandleProtocol(h, &EFI_DISK_IO_PROTOCOL_GUID, unsafe.Pointer(&dio)); st != EFI_SUCCESS {
			continue
		}
		var bio *EFI_BLOCK_IO_PROTOCOL
		if st = BS().HandleProtocol(h, &EFI_BLOCK_IO_PROTOCOL_GUID, unsafe.Pointer(&bio)); st != EFI_SUCCESS {
			// Some platforms expose DiskIO without BlockIO on the same handle; try parent inference if you care.
			continue
		}
		disks = append(disks, &Disk{DiskIO: dio, BlockIO: bio})
	}
	return disks, nil
}

// Size returns total bytes addressable on this disk (LastBlock is inclusive).
func (d *Disk) Size() int64 {
	if d == nil || d.BlockIO == nil || d.BlockIO.Media == nil {
		return 0
	}
	blk := int64(d.BlockIO.Media.BlockSize)
	// LastBlock is the highest LBA (inclusive), so count = LastBlock+1
	return int64(d.BlockIO.Media.LastBlock+1) * blk
}

// SectorSize returns the logical block size in bytes.
func (d *Disk) SectorSize() int {
	if d == nil || d.BlockIO == nil || d.BlockIO.Media == nil {
		return 0
	}
	return int(d.BlockIO.Media.BlockSize)
}

// ReadAt satisfies io.ReaderAt over Disk I/O (byte-addressable).
func (d *Disk) ReadAt(p []byte, off int64) (int, error) {
	if d == nil || d.DiskIO == nil || d.BlockIO == nil || d.BlockIO.Media == nil {
		return 0, ErrDiskNotReady
	}
	if !d.BlockIO.Media.MediaPresent {
		return 0, ErrNoMediaPresent
	}
	if off < 0 {
		return 0, ErrNegativeOffset
	}
	if len(p) == 0 {
		return 0, nil
	}
	sz := UINTN(len(p))
	st := d.DiskIO.ReadDisk(d.BlockIO.Media.MediaId, uint64(off), sz, unsafe.Pointer(&p[0]))
	if st == EFI_SUCCESS {
		return int(sz), nil
	}
	return 0, StatusError(st)
}

// WriteAt satisfies io.WriterAt over Disk I/O (byte-addressable).
func (d *Disk) WriteAt(p []byte, off int64) (int, error) {
	if d == nil || d.DiskIO == nil || d.BlockIO == nil || d.BlockIO.Media == nil {
		return 0, ErrDiskNotReady
	}
	if d.BlockIO.Media.ReadOnly {
		return 0, ErrReadOnly
	}
	if off < 0 {
		return 0, ErrNegativeOffset
	}
	if len(p) == 0 {
		return 0, nil
	}
	sz := UINTN(len(p))
	st := d.DiskIO.WriteDisk(d.BlockIO.Media.MediaId, uint64(off), sz, unsafe.Pointer(&p[0]))
	if st == EFI_SUCCESS {
		return int(sz), nil
	}
	return 0, StatusError(st)
}

// Flush flushes device caches via Block I/O (if supported).
func (d *Disk) Flush() error {
	if d == nil || d.BlockIO == nil {
		return ErrDiskNotReady
	}
	st := d.BlockIO.FlushBlocks()
	if st == EFI_SUCCESS {
		return nil
	}
	return StatusError(st)
}
