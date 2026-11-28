package main

var commands = map[string]func() string{
	"status":  func() string { return "ok" },
	"version": func() string { return "1.0" },
	"reset":   func() string { return "resetting" },
}

func run(name string) string {
	if cmd, ok := commands[name]; ok {
		return cmd()
	}
	return "unknown command"
}

func main() {
	println(run("status"))
	println(run("help"))
}
