//go:build tinygo

package syscall

func Seek(fd int, offset int64, whence int) (int64, error) {
	println("syscall.Seek not implemented", fd, offset, whence)
	return 0, ENOSYS
}

func Close(fd int) error {
	return system.Close(fd)
}

func Fsync(fd int) error {
	println("syscall.Fsync not implemented", fd)
	return ENOSYS
}

func Fchmod(fd int, mode uint32) error {
	println("syscall.Fchmod not implemented", fd, mode)
	return ENOSYS
}

func Fchown(fd int, uid, gid int) error {
	println("syscall.Fchown not implemented", fd, uid, gid)
	return ENOSYS
}

func Ftruncate(fd int, length int64) error {
	println("syscall.Ftruncate not implemented", fd, length)
	return ENOSYS
}

func Fstat(fd int, st *Stat_t) error {
	println("syscall.Fstat not implemented", fd, st)
	return ENOSYS
}

func Dup(fd int) (int, error) {
	println("syscall.Dup not implemented", fd)
	return 0, ENOSYS
}

func Fchdir(fd int) error {
	println("syscall.Fchdir not implemented", fd)
	return ENOSYS
}
