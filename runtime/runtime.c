
#include <unistd.h>
#include <string.h>
#include <stdio.h>
#include "runtime.h"

void __go_printstring(char *str) {
	write(STDOUT_FILENO, str, strlen(str));
}

void __go_printspace() {
	write(STDOUT_FILENO, " ", 1);
}

void __go_printnl() {
	write(STDOUT_FILENO, "\n", 1);
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
