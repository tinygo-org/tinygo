// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build baremetal || nintendoswitch || js || wasm_unknown

package syscall

// TODO: generate with runtime/mknacl.sh, allow override with IRT.
const (
	sys_null                 = 1
	sys_nameservice          = 2
	sys_dup                  = 8
	sys_dup2                 = 9
	sys_open                 = 10
	sys_close                = 11
	sys_read                 = 12
	sys_write                = 13
	sys_lseek                = 14
	sys_stat                 = 16
	sys_fstat                = 17
	sys_chmod                = 18
	sys_isatty               = 19
	sys_brk                  = 20
	sys_mmap                 = 21
	sys_munmap               = 22
	sys_getdents             = 23
	sys_mprotect             = 24
	sys_list_mappings        = 25
	sys_exit                 = 30
	sys_getpid               = 31
	sys_sched_yield          = 32
	sys_sysconf              = 33
	sys_gettimeofday         = 40
	sys_clock                = 41
	sys_nanosleep            = 42
	sys_clock_getres         = 43
	sys_clock_gettime        = 44
	sys_mkdir                = 45
	sys_rmdir                = 46
	sys_chdir                = 47
	sys_getcwd               = 48
	sys_unlink               = 49
	sys_imc_makeboundsock    = 60
	sys_imc_accept           = 61
	sys_imc_connect          = 62
	sys_imc_sendmsg          = 63
	sys_imc_recvmsg          = 64
	sys_imc_mem_obj_create   = 65
	sys_imc_socketpair       = 66
	sys_mutex_create         = 70
	sys_mutex_lock           = 71
	sys_mutex_trylock        = 72
	sys_mutex_unlock         = 73
	sys_cond_create          = 74
	sys_cond_wait            = 75
	sys_cond_signal          = 76
	sys_cond_broadcast       = 77
	sys_cond_timed_wait_abs  = 79
	sys_thread_create        = 80
	sys_thread_exit          = 81
	sys_tls_init             = 82
	sys_thread_nice          = 83
	sys_tls_get              = 84
	sys_second_tls_set       = 85
	sys_second_tls_get       = 86
	sys_exception_handler    = 87
	sys_exception_stack      = 88
	sys_exception_clear_flag = 89
	sys_sem_create           = 100
	sys_sem_wait             = 101
	sys_sem_post             = 102
	sys_sem_get_value        = 103
	sys_dyncode_create       = 104
	sys_dyncode_modify       = 105
	sys_dyncode_delete       = 106
	sys_test_infoleak        = 109
	sys_test_crash           = 110
	sys_test_syscall_1       = 111
	sys_test_syscall_2       = 112
	sys_futex_wait_abs       = 120
	sys_futex_wake           = 121
	sys_pread                = 130
	sys_pwrite               = 131
	sys_truncate             = 140
	sys_lstat                = 141
	sys_link                 = 142
	sys_rename               = 143
	sys_symlink              = 144
	sys_access               = 145
	sys_readlink             = 146
	sys_utimes               = 147
	sys_get_random_bytes     = 150
)

// Do the interface allocations only once for common
// Errno values.
var (
	errEAGAIN error = EAGAIN
	errEINVAL error = EINVAL
	errENOENT error = ENOENT
)

// errnoErr returns common boxed Errno values, to prevent
// allocations at runtime.
func errnoErr(e Errno) error {
	switch e {
	case 0:
		return nil
	case EAGAIN:
		return errEAGAIN
	case EINVAL:
		return errEINVAL
	case ENOENT:
		return errENOENT
	}
	return e
}
