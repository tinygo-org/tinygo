//go:build linux && !baremetal && !tinygo.wasm && !nintendoswitch

package os

import "syscall"

// Storage for the two by-value POSIX objects posix_spawn takes. musl declares
// them in lib/musl/include/spawn.h as
//
//	typedef struct {
//		int __pad0[2];
//		void *__actions;
//		int __pad[16];
//	} posix_spawn_file_actions_t;
//
//	typedef struct {
//		int __flags;
//		pid_t __pgrp;
//		sigset_t __def, __mask;
//		int __prio, __pol;
//		void *__fn;
//		char __pad[64-sizeof(void *)];
//	} posix_spawnattr_t;
//
// with a sigset_t of 128 bytes. The file-actions object is then 80 bytes on
// LP64 and 76 bytes on a 32-bit target, and the attribute object is 336 bytes.
// The arrays below are larger than that and uint64 for the alignment. Only
// libc looks inside them.
type spawnFileActions [16]uint64

type spawnAttr [48]uint64

// The sigset_t of musl, 128 bytes. The zero value is the empty set.
type sigset [16]uint64

// checkSysProcAttr reports whether the SysProcAttr asks for something that
// posix_spawn cannot do. Linux declares more fields than POSIX, and each of
// them needs Go code to run in the child between the clone and the exec.
func checkSysProcAttr(sys *syscall.SysProcAttr) error {
	if err := checkSysProcAttrCommon(sys); err != nil {
		return err
	}
	switch {
	case sys.Pdeathsig != 0:
		return errUnsupportedSysField("Pdeathsig")
	case sys.Cloneflags != 0:
		return errUnsupportedSysField("Cloneflags")
	case sys.Unshareflags != 0:
		return errUnsupportedSysField("Unshareflags")
	case sys.UidMappings != nil:
		return errUnsupportedSysField("UidMappings")
	case sys.GidMappings != nil:
		return errUnsupportedSysField("GidMappings")
	case sys.GidMappingsEnableSetgroups:
		return errUnsupportedSysField("GidMappingsEnableSetgroups")
	case sys.AmbientCaps != nil:
		return errUnsupportedSysField("AmbientCaps")
	case sys.UseCgroupFD:
		return errUnsupportedSysField("UseCgroupFD")
	case sys.CgroupFD != 0:
		return errUnsupportedSysField("CgroupFD")
	case sys.PidFD != nil:
		return errUnsupportedSysField("PidFD")
	}
	return nil
}
