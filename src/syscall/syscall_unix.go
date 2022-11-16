package syscall

func Exec(argv0 string, argv []string, envv []string) (err error)
