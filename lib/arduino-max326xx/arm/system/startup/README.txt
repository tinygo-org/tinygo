This is a precompile library. If you make changes to any source file(s) listed in Makefile, you must rebuild the library.

Open MINGW for Windows or terminal for Mac/Linux, navigate to this directory and run:

$ make clean
$ target=TARGET_PROCESSOR make
$ rm Build_TARGET_PROCESSOR/*.d Build_TARGET_PROCESSOR/*.o

TARGET_PROCESSOR can be one of:
    max32620
    max32625
    max32630
