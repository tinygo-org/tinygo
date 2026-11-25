// TinyGo has one thread and stdio files start unlocked.
// The wasi-libc stdio code still references these symbols.
int __lockfile(void *f) { return 0; }
void __unlockfile(void *f) {}
