typedef short myint;
int add(int a, int b);
typedef int (*binop_t) (int, int);
int doCallback(int a, int b, binop_t cb);
typedef int * intPointer;
void store(int value, int *ptr);

typedef struct collection {
	short         s;
	long          l;
	float         f;
	unsigned char c;
} collection_t;

// linked list
typedef struct list_t {
	int           n;
	struct list_t *next;
} list_t;

typedef union joined {
	myint s;
	float f;
	short data[3];
} joined_t;
void unionSetShort(short s);
void unionSetFloat(float f);
void unionSetData(short f0, short f1, short f2);

// test globals
extern int global;
extern _Bool globalBool;
extern _Bool globalBool2;
extern float globalFloat;
extern double globalDouble;
extern _Complex float globalComplexFloat;
extern _Complex double globalComplexDouble;
extern _Complex double globalComplexLongDouble;
extern char globalChar;
extern collection_t globalStruct;
extern int globalStructSize;
extern short globalArray[3];
extern joined_t globalUnion;
extern int globalUnionSize;

// test duplicate definitions
int add(int a, int b);
extern int global;
