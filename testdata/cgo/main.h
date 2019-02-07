typedef short myint;
int add(int a, int b);
typedef int (*binop_t) (int, int);
int doCallback(int a, int b, binop_t cb);
typedef int * intPointer;
extern int global;
void store(int value, int *ptr);

// test duplicate definitions
int add(int a, int b);
extern int global;
