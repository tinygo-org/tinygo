typedef short myint;
int add(int a, int b);
typedef int (*binop_t) (int, int);
int doCallback(int a, int b, binop_t cb);
typedef int * intPointer;
void store(int value, int *ptr);

typedef struct collection {
	short s;
	long  l;
	float f;
} collection_t;

// test globals
extern int global;
extern _Bool globalBool;
extern _Bool globalBool2;
extern float globalFloat;
extern double globalDouble;
extern _Complex float globalComplexFloat;
extern _Complex double globalComplexDouble;
extern _Complex double globalComplexLongDouble;
extern collection_t globalStruct;

// test duplicate definitions
int add(int a, int b);
extern int global;
