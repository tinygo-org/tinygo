package main

/*
void d(void);
void c(void);
void b(void);
void a(void);
*/
import "C"

func main() {
	C.d()
	C.c()
	C.b()
	C.a()
}
