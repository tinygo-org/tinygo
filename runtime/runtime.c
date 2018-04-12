
#include <unistd.h>
#include <string.h>
#include <stdio.h>
#include "runtime.h"

#define print(buf, len) write(STDOUT_FILENO, buf, len)

void __go_printstring(string_t *str) {
	write(STDOUT_FILENO, str->buf, str->len);
}

void __go_printint(intgo_t n) {
	// Print integer in signed big-endian base-10 notation, for humans to
	// read.
	// TODO: don't recurse, but still be compact (and don't divide/mod
	// more than necessary).
	if (n < 0) {
		print("-", 1);
		n = -n;
	}
	intgo_t prevdigits = n / 10;
	if (prevdigits != 0) {
		__go_printint(prevdigits);
	}
	char buf[1] = {(n % 10) + '0'};
	print(buf, 1);
}

void __go_printspace() {
	print(" ", 1);
}

void __go_printnl() {
	print("\n", 1);
}

void go_main() __asm__("main.main");

void __go_runtime_main() {
	go_main();
}

__attribute__((weak))
int main() {
	__go_runtime_main();

	return 0;
}
