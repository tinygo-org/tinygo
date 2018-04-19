
#include <unistd.h>
#include <string.h>
#include <stdio.h>
#include "runtime.h"

void __go_printstring(string_t str) {
	for (int i = 0; i < str.len; i++) {
		putchar(str.buf[i]);
	}
}

void __go_printint(intgo_t n) {
	// Print integer in signed big-endian base-10 notation, for humans to
	// read.
	// TODO: don't recurse, but still be compact (and don't divide/mod
	// more than necessary).
	if (n < 0) {
		putchar('-');
		n = -n;
	}
	intgo_t prevdigits = n / 10;
	if (prevdigits != 0) {
		__go_printint(prevdigits);
	}
	putchar((n % 10) + '0');
}

void __go_printbyte(uint8_t c) {
	putchar(c);
}

void __go_printspace() {
	putchar(' ');
}

void __go_printnl() {
	putchar('\n');
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
