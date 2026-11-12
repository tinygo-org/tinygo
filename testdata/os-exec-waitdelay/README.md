# WaitDelay with a background child

This manual test uses the Go standard library `os/exec` with TinyGo's `os`.
The shell exits and a background `sleep` keeps the output pipe open for two
seconds. `WaitDelay` is 100 ms. The test returns status 1 if the wait takes
one second or more, or if the error is not `exec.ErrWaitDelay`.

Build and run from the repository root with the process changes in TINYGOROOT.

```sh
tinygo build -p 1 -o /tmp/waitdelay ./testdata/os-exec-waitdelay/main.go
timeout 10 /tmp/waitdelay
timeout 10 /tmp/waitdelay pipe
go build -p 1 -o /tmp/waitdelay-go ./testdata/os-exec-waitdelay/main.go
timeout 10 /tmp/waitdelay-go
timeout 10 /tmp/waitdelay-go pipe
```

`timeout` is an external limit for Linux. The background child exits after
two seconds. The `pipe` mode closes a reader during a read. It closes the
writer two seconds later so that the test can finish if the read stays blocked.

Measured on Linux arm64 with the released TinyGo 0.42.0 compiler, Go 1.27.0,
and a TINYGOROOT copy with PR #5634 and the fd remapping fix.

| Test | TinyGo | Go |
| --- | --- | --- |
| WaitDelay 100 ms | 2.004 s, ErrWaitDelay | 101 ms, ErrWaitDelay |
| Read after reader close | 2.003 s, EOF | 57 us, file already closed |

`src/os/file_anyos.go` calls `syscall.Read` and `syscall.Close` directly.
Closing the reader does not stop the active blocking read in this test.
Go's `Cmd.awaitGoroutines` closes the pipes when the timer expires, then waits
for the copy goroutines. That wait lasts until the background child closes
the pipe. The timer fires, but it cannot enforce the limit.

This remains open. A fix needs interruptible pipe I/O and coordination between
close and active I/O. It must also prevent an old operation from using a reused
descriptor. Changes to spawn file actions or Darwin fcntl do not fix this Linux
failure. PR #5630 addresses lock contention, not this blocked read.

The follow-up must test WaitDelay after normal exit and context cancellation,
blocked pipe reads and writes, prompt close, and descriptor reuse on hosted
Linux and Darwin. A process that keeps its inherited output open must not
keep `Cmd.Wait` blocked after the configured limit.
