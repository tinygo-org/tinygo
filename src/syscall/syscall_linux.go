package syscall

func Setuid(code int) (err error)
func Setgid(code int) (err error)
func Setreuid(ruid, euid int) (err error)
func Setregid(rgid, egid int) (err error)
func Setresuid(ruid, euid, suid int) (err error)
func Setresgid(rgid, egid, sgid int) (err error)
