
#include "runtime.h"

void go_main() __asm__("main.main");

void __go_runtime_main() {
	go_main();
}

__attribute__((weak))
int main() {
	__go_runtime_main();

	return 0;
}
