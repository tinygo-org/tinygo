; ModuleID = 'wrapper.go'
source_filename = "wrapper.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime.typecodeID = type { ptr, i32, ptr, ptr, i32 }
%runtime.structField = type { ptr, ptr, ptr, i1 }
%runtime.interfaceMethodInfo = type { ptr, i32 }
%runtime._interface = type { i32, ptr }
%main.Cmd = type { %runtime._interface, { ptr, i32, i32 }, { ptr, i32, i32 } }
%runtime._string = type { ptr, i32 }
%"internal/task.stackState" = type { i32, i32 }
%"internal/task.state" = type { i32, ptr, %"internal/task.stackState", i1 }
%"internal/task.Queue" = type { ptr, ptr }
%sync.Once = type { i1, %sync.Mutex }
%sync.Mutex = type { i1, %"internal/task.Stack" }
%"internal/task.Stack" = type { ptr }
%main.closeOnce = type { ptr, %sync.Once, %runtime._interface }

@"main$string" = internal unnamed_addr constant [23 x i8] c"exec: Stdin already set", align 1
@"reflect/types.type:pointer:named:os.File" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:os.File", i32 0, ptr @"*os.File$methodset", ptr null, i32 0 }
@"reflect/types.type:named:os.File" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{#file:pointer:named:os.file}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:os.File", i32 0 }
@"reflect/types.type:struct:{#file:pointer:named:os.file}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields", i32 0, ptr null, ptr @"reflect/types.type:pointer:struct:{#file:pointer:named:os.file}", i32 0 }
@"reflect/types.structFields" = private unnamed_addr global [1 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:pointer:named:os.file", ptr @"reflect/types.structFieldName.3", ptr null, i1 true }]
@"reflect/types.type:pointer:named:os.file" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:os.file", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.type:named:os.file" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{handle:named:os.FileHandle,name:basic:string}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:os.file", i32 0 }
@"reflect/types.type:struct:{handle:named:os.FileHandle,name:basic:string}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.1", i32 0, ptr null, ptr @"reflect/types.type:pointer:struct:{handle:named:os.FileHandle,name:basic:string}", i32 0 }
@"reflect/types.structFields.1" = private unnamed_addr global [2 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:named:os.FileHandle", ptr @"reflect/types.structFieldName", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:basic:string", ptr @"reflect/types.structFieldName.2", ptr null, i1 false }]
@"reflect/types.type:named:os.FileHandle" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:interface:{Close:func:{}{named:error},Read:func:{slice:basic:uint8}{basic:int,named:error},ReadAt:func:{slice:basic:uint8,basic:int64}{basic:int,named:error},Seek:func:{basic:int64,basic:int}{basic:int64,named:error},Write:func:{slice:basic:uint8}{basic:int,named:error}}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:os.FileHandle", i32 ptrtoint (ptr @"interface:{Close:func:{}{named:error},Read:func:{slice:basic:uint8}{basic:int,named:error},ReadAt:func:{slice:basic:uint8,basic:int64}{basic:int,named:error},Seek:func:{basic:int64,basic:int}{basic:int64,named:error},Write:func:{slice:basic:uint8}{basic:int,named:error}}.$typeassert" to i32) }
@"reflect/types.type:interface:{Close:func:{}{named:error},Read:func:{slice:basic:uint8}{basic:int,named:error},ReadAt:func:{slice:basic:uint8,basic:int64}{basic:int,named:error},Seek:func:{basic:int64,basic:int}{basic:int64,named:error},Write:func:{slice:basic:uint8}{basic:int,named:error}}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.interface:interface{Close() (err error); Read(b []byte) (n int, err error); ReadAt(b []byte, offset int64) (n int, err error); Seek(offset int64, whence int) (newoffset int64, err error); Write(b []byte) (n int, err error)}$interface", i32 0, ptr null, ptr @"reflect/types.type:pointer:interface:{Close:func:{}{named:error},Read:func:{slice:basic:uint8}{basic:int,named:error},ReadAt:func:{slice:basic:uint8,basic:int64}{basic:int,named:error},Seek:func:{basic:int64,basic:int}{basic:int64,named:error},Write:func:{slice:basic:uint8}{basic:int,named:error}}", i32 ptrtoint (ptr @"interface:{Close:func:{}{named:error},Read:func:{slice:basic:uint8}{basic:int,named:error},ReadAt:func:{slice:basic:uint8,basic:int64}{basic:int,named:error},Seek:func:{basic:int64,basic:int}{basic:int64,named:error},Write:func:{slice:basic:uint8}{basic:int,named:error}}.$typeassert" to i32) }
@"reflect/methods.Close() error" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Read([]uint8) (int, error)" = linkonce_odr constant i8 0, align 1
@"reflect/methods.ReadAt([]uint8, int64) (int, error)" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Seek(int64, int) (int64, error)" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Write([]uint8) (int, error)" = linkonce_odr constant i8 0, align 1
@"reflect/types.interface:interface{Close() (err error); Read(b []byte) (n int, err error); ReadAt(b []byte, offset int64) (n int, err error); Seek(offset int64, whence int) (newoffset int64, err error); Write(b []byte) (n int, err error)}$interface" = linkonce_odr constant [5 x ptr] [ptr @"reflect/methods.Close() error", ptr @"reflect/methods.Read([]uint8) (int, error)", ptr @"reflect/methods.ReadAt([]uint8, int64) (int, error)", ptr @"reflect/methods.Seek(int64, int) (int64, error)", ptr @"reflect/methods.Write([]uint8) (int, error)"]
@"reflect/types.type:pointer:interface:{Close:func:{}{named:error},Read:func:{slice:basic:uint8}{basic:int,named:error},ReadAt:func:{slice:basic:uint8,basic:int64}{basic:int,named:error},Seek:func:{basic:int64,basic:int}{basic:int64,named:error},Write:func:{slice:basic:uint8}{basic:int,named:error}}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:interface:{Close:func:{}{named:error},Read:func:{slice:basic:uint8}{basic:int,named:error},ReadAt:func:{slice:basic:uint8,basic:int64}{basic:int,named:error},Seek:func:{basic:int64,basic:int}{basic:int64,named:error},Write:func:{slice:basic:uint8}{basic:int,named:error}}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.type:pointer:named:os.FileHandle" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:os.FileHandle", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.structFieldName" = private unnamed_addr global [6 x i8] c"handle"
@"reflect/types.type:basic:string" = linkonce_odr constant %runtime.typecodeID { ptr null, i32 0, ptr null, ptr @"reflect/types.type:pointer:basic:string", i32 0 }
@"reflect/types.type:pointer:basic:string" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:basic:string", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.structFieldName.2" = private unnamed_addr global [4 x i8] c"name"
@"reflect/types.type:pointer:struct:{handle:named:os.FileHandle,name:basic:string}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{handle:named:os.FileHandle,name:basic:string}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.structFieldName.3" = private unnamed_addr global [4 x i8] c"file"
@"reflect/types.type:pointer:struct:{#file:pointer:named:os.file}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{#file:pointer:named:os.file}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/methods.Fd() uintptr" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Name() string" = linkonce_odr constant i8 0, align 1
@"reflect/methods.ReadDir(int) ([]io/fs.DirEntry, error)" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Readdir(int) ([]io/fs.FileInfo, error)" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Readdirnames(int) ([]string, error)" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Stat() (io/fs.FileInfo, error)" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Sync() error" = linkonce_odr constant i8 0, align 1
@"reflect/methods.SyscallConn() (syscall.RawConn, error)" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Truncate(int64) error" = linkonce_odr constant i8 0, align 1
@"reflect/methods.WriteAt([]uint8, int64) (int, error)" = linkonce_odr constant i8 0, align 1
@"reflect/methods.WriteString(string) (int, error)" = linkonce_odr constant i8 0, align 1
@"os.$methods.readdir(int, os.readdirMode) ([]string, []io/fs.DirEntry, []io/fs.FileInfo, error)" = linkonce_odr constant i8 0, align 1
@"*os.File$methodset" = linkonce_odr constant [17 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.Close() error", i32 ptrtoint (ptr @"(*os.File).Close" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Fd() uintptr", i32 ptrtoint (ptr @"(*os.File).Fd" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Name() string", i32 ptrtoint (ptr @"(*os.File).Name" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Read([]uint8) (int, error)", i32 ptrtoint (ptr @"(*os.File).Read" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadAt([]uint8, int64) (int, error)", i32 ptrtoint (ptr @"(*os.File).ReadAt" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadDir(int) ([]io/fs.DirEntry, error)", i32 ptrtoint (ptr @"(*os.File).ReadDir" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Readdir(int) ([]io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(*os.File).Readdir" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Readdirnames(int) ([]string, error)", i32 ptrtoint (ptr @"(*os.File).Readdirnames" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Seek(int64, int) (int64, error)", i32 ptrtoint (ptr @"(*os.File).Seek" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Stat() (io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(*os.File).Stat" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Sync() error", i32 ptrtoint (ptr @"(*os.File).Sync" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.SyscallConn() (syscall.RawConn, error)", i32 ptrtoint (ptr @"(*os.File).SyscallConn" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Truncate(int64) error", i32 ptrtoint (ptr @"(*os.File).Truncate" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Write([]uint8) (int, error)", i32 ptrtoint (ptr @"(*os.File).Write" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.WriteAt([]uint8, int64) (int, error)", i32 ptrtoint (ptr @"(*os.File).WriteAt" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.WriteString(string) (int, error)", i32 ptrtoint (ptr @"(*os.File).WriteString" to i32) }, %runtime.interfaceMethodInfo { ptr @"os.$methods.readdir(int, os.readdirMode) ([]string, []io/fs.DirEntry, []io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(*os.File).readdir" to i32) }]
@"reflect/types.type:pointer:named:main.closeOnce" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:main.closeOnce", i32 0, ptr @"*main.closeOnce$methodset", ptr null, i32 0 }
@"reflect/types.type:named:main.closeOnce" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{#File:pointer:named:os.File,once:named:sync.Once,err:named:error}", i32 0, ptr @"main.closeOnce$methodset", ptr @"reflect/types.type:pointer:named:main.closeOnce", i32 0 }
@"reflect/types.type:struct:{#File:pointer:named:os.File,once:named:sync.Once,err:named:error}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.4", i32 0, ptr @"struct{*os.File; once sync.Once; err error}$methodset", ptr @"reflect/types.type:pointer:struct:{#File:pointer:named:os.File,once:named:sync.Once,err:named:error}", i32 0 }
@"reflect/types.structFields.4" = private unnamed_addr global [3 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:pointer:named:os.File", ptr @"reflect/types.structFieldName.5", ptr null, i1 true }, %runtime.structField { ptr @"reflect/types.type:named:sync.Once", ptr @"reflect/types.structFieldName.31", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:named:error", ptr @"reflect/types.structFieldName.32", ptr null, i1 false }]
@"reflect/types.structFieldName.5" = private unnamed_addr global [4 x i8] c"File"
@"reflect/types.type:named:sync.Once" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{done:basic:bool,m:named:sync.Mutex}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:sync.Once", i32 0 }
@"reflect/types.type:struct:{done:basic:bool,m:named:sync.Mutex}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.6", i32 0, ptr null, ptr @"reflect/types.type:pointer:struct:{done:basic:bool,m:named:sync.Mutex}", i32 0 }
@"reflect/types.structFields.6" = private unnamed_addr global [2 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:basic:bool", ptr @"reflect/types.structFieldName.7", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:named:sync.Mutex", ptr @"reflect/types.structFieldName.30", ptr null, i1 false }]
@"reflect/types.type:basic:bool" = linkonce_odr constant %runtime.typecodeID { ptr null, i32 0, ptr null, ptr @"reflect/types.type:pointer:basic:bool", i32 0 }
@"reflect/types.type:pointer:basic:bool" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:basic:bool", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.structFieldName.7" = private unnamed_addr global [4 x i8] c"done"
@"reflect/types.type:named:sync.Mutex" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{locked:basic:bool,blocked:named:internal/task.Stack}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:sync.Mutex", i32 0 }
@"reflect/types.type:struct:{locked:basic:bool,blocked:named:internal/task.Stack}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.8", i32 0, ptr null, ptr @"reflect/types.type:pointer:struct:{locked:basic:bool,blocked:named:internal/task.Stack}", i32 0 }
@"reflect/types.structFields.8" = private unnamed_addr global [2 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:basic:bool", ptr @"reflect/types.structFieldName.9", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:named:internal/task.Stack", ptr @"reflect/types.structFieldName.29", ptr null, i1 false }]
@"reflect/types.structFieldName.9" = private unnamed_addr global [6 x i8] c"locked"
@"reflect/types.type:named:internal/task.Stack" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{top:pointer:named:internal/task.Task}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:internal/task.Stack", i32 0 }
@"reflect/types.type:struct:{top:pointer:named:internal/task.Task}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.10", i32 0, ptr null, ptr @"reflect/types.type:pointer:struct:{top:pointer:named:internal/task.Task}", i32 0 }
@"reflect/types.structFields.10" = private unnamed_addr global [1 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:pointer:named:internal/task.Task", ptr @"reflect/types.structFieldName.28", ptr null, i1 false }]
@"reflect/types.type:pointer:named:internal/task.Task" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:internal/task.Task", i32 0, ptr @"*internal/task.Task$methodset", ptr null, i32 0 }
@"reflect/types.type:named:internal/task.Task" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{Next:pointer:named:internal/task.Task,Ptr:basic:unsafe.Pointer,Data:basic:uint64,gcData:named:internal/task.gcData,state:named:internal/task.state,DeferFrame:basic:unsafe.Pointer}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:internal/task.Task", i32 0 }
@"reflect/types.type:struct:{Next:pointer:named:internal/task.Task,Ptr:basic:unsafe.Pointer,Data:basic:uint64,gcData:named:internal/task.gcData,state:named:internal/task.state,DeferFrame:basic:unsafe.Pointer}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.11", i32 0, ptr null, ptr @"reflect/types.type:pointer:struct:{Next:pointer:named:internal/task.Task,Ptr:basic:unsafe.Pointer,Data:basic:uint64,gcData:named:internal/task.gcData,state:named:internal/task.state,DeferFrame:basic:unsafe.Pointer}", i32 0 }
@"reflect/types.structFields.11" = private unnamed_addr global [6 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:pointer:named:internal/task.Task", ptr @"reflect/types.structFieldName.12", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:basic:unsafe.Pointer", ptr @"reflect/types.structFieldName.13", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:basic:uint64", ptr @"reflect/types.structFieldName.14", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:named:internal/task.gcData", ptr @"reflect/types.structFieldName.17", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:named:internal/task.state", ptr @"reflect/types.structFieldName.26", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:basic:unsafe.Pointer", ptr @"reflect/types.structFieldName.27", ptr null, i1 false }]
@"reflect/types.structFieldName.12" = private unnamed_addr global [4 x i8] c"Next"
@"reflect/types.type:basic:unsafe.Pointer" = linkonce_odr constant %runtime.typecodeID { ptr null, i32 0, ptr null, ptr @"reflect/types.type:pointer:basic:unsafe.Pointer", i32 0 }
@"reflect/types.type:pointer:basic:unsafe.Pointer" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:basic:unsafe.Pointer", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.structFieldName.13" = private unnamed_addr global [3 x i8] c"Ptr"
@"reflect/types.type:basic:uint64" = linkonce_odr constant %runtime.typecodeID { ptr null, i32 0, ptr null, ptr @"reflect/types.type:pointer:basic:uint64", i32 0 }
@"reflect/types.type:pointer:basic:uint64" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:basic:uint64", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.structFieldName.14" = private unnamed_addr global [4 x i8] c"Data"
@"reflect/types.type:named:internal/task.gcData" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{stackChain:basic:unsafe.Pointer}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:internal/task.gcData", i32 0 }
@"reflect/types.type:struct:{stackChain:basic:unsafe.Pointer}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.15", i32 0, ptr null, ptr @"reflect/types.type:pointer:struct:{stackChain:basic:unsafe.Pointer}", i32 0 }
@"reflect/types.structFields.15" = private unnamed_addr global [1 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:basic:unsafe.Pointer", ptr @"reflect/types.structFieldName.16", ptr null, i1 false }]
@"reflect/types.structFieldName.16" = private unnamed_addr global [10 x i8] c"stackChain"
@"reflect/types.type:pointer:struct:{stackChain:basic:unsafe.Pointer}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{stackChain:basic:unsafe.Pointer}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.type:pointer:named:internal/task.gcData" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:internal/task.gcData", i32 0, ptr @"*internal/task.gcData$methodset", ptr null, i32 0 }
@"internal/task.$methods.swap()" = linkonce_odr constant i8 0, align 1
@"*internal/task.gcData$methodset" = linkonce_odr constant [1 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"internal/task.$methods.swap()", i32 ptrtoint (ptr @"(*internal/task.gcData).swap" to i32) }]
@"reflect/types.structFieldName.17" = private unnamed_addr global [6 x i8] c"gcData"
@"reflect/types.type:named:internal/task.state" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{entry:basic:uintptr,args:basic:unsafe.Pointer,#stackState:named:internal/task.stackState,launched:basic:bool}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:internal/task.state", i32 0 }
@"reflect/types.type:struct:{entry:basic:uintptr,args:basic:unsafe.Pointer,#stackState:named:internal/task.stackState,launched:basic:bool}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.18", i32 0, ptr null, ptr @"reflect/types.type:pointer:struct:{entry:basic:uintptr,args:basic:unsafe.Pointer,#stackState:named:internal/task.stackState,launched:basic:bool}", i32 0 }
@"reflect/types.structFields.18" = private unnamed_addr global [4 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:basic:uintptr", ptr @"reflect/types.structFieldName.19", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:basic:unsafe.Pointer", ptr @"reflect/types.structFieldName.20", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:named:internal/task.stackState", ptr @"reflect/types.structFieldName.24", ptr null, i1 true }, %runtime.structField { ptr @"reflect/types.type:basic:bool", ptr @"reflect/types.structFieldName.25", ptr null, i1 false }]
@"reflect/types.type:basic:uintptr" = linkonce_odr constant %runtime.typecodeID { ptr null, i32 0, ptr null, ptr @"reflect/types.type:pointer:basic:uintptr", i32 0 }
@"reflect/types.type:pointer:basic:uintptr" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:basic:uintptr", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.structFieldName.19" = private unnamed_addr global [5 x i8] c"entry"
@"reflect/types.structFieldName.20" = private unnamed_addr global [4 x i8] c"args"
@"reflect/types.type:named:internal/task.stackState" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{asyncifysp:basic:uintptr,csp:basic:uintptr}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:internal/task.stackState", i32 0 }
@"reflect/types.type:struct:{asyncifysp:basic:uintptr,csp:basic:uintptr}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.21", i32 0, ptr null, ptr @"reflect/types.type:pointer:struct:{asyncifysp:basic:uintptr,csp:basic:uintptr}", i32 0 }
@"reflect/types.structFields.21" = private unnamed_addr global [2 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:basic:uintptr", ptr @"reflect/types.structFieldName.22", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:basic:uintptr", ptr @"reflect/types.structFieldName.23", ptr null, i1 false }]
@"reflect/types.structFieldName.22" = private unnamed_addr global [10 x i8] c"asyncifysp"
@"reflect/types.structFieldName.23" = private unnamed_addr global [3 x i8] c"csp"
@"reflect/types.type:pointer:struct:{asyncifysp:basic:uintptr,csp:basic:uintptr}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{asyncifysp:basic:uintptr,csp:basic:uintptr}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.type:pointer:named:internal/task.stackState" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:internal/task.stackState", i32 0, ptr @"*internal/task.stackState$methodset", ptr null, i32 0 }
@"internal/task.$methods.unwind()" = linkonce_odr constant i8 0, align 1
@"*internal/task.stackState$methodset" = linkonce_odr constant [1 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"internal/task.$methods.unwind()", i32 ptrtoint (ptr @tinygo_unwind to i32) }]
@"reflect/types.structFieldName.24" = private unnamed_addr global [10 x i8] c"stackState"
@"reflect/types.structFieldName.25" = private unnamed_addr global [8 x i8] c"launched"
@"reflect/types.type:pointer:struct:{entry:basic:uintptr,args:basic:unsafe.Pointer,#stackState:named:internal/task.stackState,launched:basic:bool}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{entry:basic:uintptr,args:basic:unsafe.Pointer,#stackState:named:internal/task.stackState,launched:basic:bool}", i32 0, ptr @"*struct{entry uintptr; args unsafe.Pointer; internal/task.stackState; launched bool}$methodset", ptr null, i32 0 }
@"*struct{entry uintptr; args unsafe.Pointer; internal/task.stackState; launched bool}$methodset" = linkonce_odr constant [1 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"internal/task.$methods.unwind()", i32 ptrtoint (ptr @"(*struct{entry uintptr; args unsafe.Pointer; internal/task.stackState; launched bool}).unwind" to i32) }]
@"reflect/types.type:pointer:named:internal/task.state" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:internal/task.state", i32 0, ptr @"*internal/task.state$methodset", ptr null, i32 0 }
@"internal/task.$methods.initialize(uintptr, unsafe.Pointer, uintptr)" = linkonce_odr constant i8 0, align 1
@"internal/task.$methods.launch()" = linkonce_odr constant i8 0, align 1
@"internal/task.$methods.rewind()" = linkonce_odr constant i8 0, align 1
@"*internal/task.state$methodset" = linkonce_odr constant [4 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"internal/task.$methods.initialize(uintptr, unsafe.Pointer, uintptr)", i32 ptrtoint (ptr @"(*internal/task.state).initialize" to i32) }, %runtime.interfaceMethodInfo { ptr @"internal/task.$methods.launch()", i32 ptrtoint (ptr @tinygo_launch to i32) }, %runtime.interfaceMethodInfo { ptr @"internal/task.$methods.rewind()", i32 ptrtoint (ptr @tinygo_rewind to i32) }, %runtime.interfaceMethodInfo { ptr @"internal/task.$methods.unwind()", i32 ptrtoint (ptr @"(*internal/task.state).unwind" to i32) }]
@"reflect/types.structFieldName.26" = private unnamed_addr global [5 x i8] c"state"
@"reflect/types.structFieldName.27" = private unnamed_addr global [10 x i8] c"DeferFrame"
@"reflect/types.type:pointer:struct:{Next:pointer:named:internal/task.Task,Ptr:basic:unsafe.Pointer,Data:basic:uint64,gcData:named:internal/task.gcData,state:named:internal/task.state,DeferFrame:basic:unsafe.Pointer}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{Next:pointer:named:internal/task.Task,Ptr:basic:unsafe.Pointer,Data:basic:uint64,gcData:named:internal/task.gcData,state:named:internal/task.state,DeferFrame:basic:unsafe.Pointer}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/methods.Resume()" = linkonce_odr constant i8 0, align 1
@"internal/task.$methods.tail() *internal/task.Task" = linkonce_odr constant i8 0, align 1
@"*internal/task.Task$methodset" = linkonce_odr constant [2 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.Resume()", i32 ptrtoint (ptr @"(*internal/task.Task).Resume" to i32) }, %runtime.interfaceMethodInfo { ptr @"internal/task.$methods.tail() *internal/task.Task", i32 ptrtoint (ptr @"(*internal/task.Task).tail" to i32) }]
@"reflect/types.structFieldName.28" = private unnamed_addr global [3 x i8] c"top"
@"reflect/types.type:pointer:struct:{top:pointer:named:internal/task.Task}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{top:pointer:named:internal/task.Task}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.type:pointer:named:internal/task.Stack" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:internal/task.Stack", i32 0, ptr @"*internal/task.Stack$methodset", ptr null, i32 0 }
@"reflect/methods.Pop() *internal/task.Task" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Push(*internal/task.Task)" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Queue() internal/task.Queue" = linkonce_odr constant i8 0, align 1
@"*internal/task.Stack$methodset" = linkonce_odr constant [3 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.Pop() *internal/task.Task", i32 ptrtoint (ptr @"(*internal/task.Stack).Pop" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Push(*internal/task.Task)", i32 ptrtoint (ptr @"(*internal/task.Stack).Push" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Queue() internal/task.Queue", i32 ptrtoint (ptr @"(*internal/task.Stack).Queue" to i32) }]
@"reflect/types.structFieldName.29" = private unnamed_addr global [7 x i8] c"blocked"
@"reflect/types.type:pointer:struct:{locked:basic:bool,blocked:named:internal/task.Stack}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{locked:basic:bool,blocked:named:internal/task.Stack}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.type:pointer:named:sync.Mutex" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:sync.Mutex", i32 0, ptr @"*sync.Mutex$methodset", ptr null, i32 0 }
@"reflect/methods.Lock()" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Unlock()" = linkonce_odr constant i8 0, align 1
@"*sync.Mutex$methodset" = linkonce_odr constant [2 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.Lock()", i32 ptrtoint (ptr @"(*sync.Mutex).Lock" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Unlock()", i32 ptrtoint (ptr @"(*sync.Mutex).Unlock" to i32) }]
@"reflect/types.structFieldName.30" = private unnamed_addr global [1 x i8] c"m"
@"reflect/types.type:pointer:struct:{done:basic:bool,m:named:sync.Mutex}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{done:basic:bool,m:named:sync.Mutex}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.type:pointer:named:sync.Once" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:sync.Once", i32 0, ptr @"*sync.Once$methodset", ptr null, i32 0 }
@"reflect/methods.Do(func())" = linkonce_odr constant i8 0, align 1
@"*sync.Once$methodset" = linkonce_odr constant [1 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.Do(func())", i32 ptrtoint (ptr @"(*sync.Once).Do" to i32) }]
@"reflect/types.structFieldName.31" = private unnamed_addr global [4 x i8] c"once"
@"reflect/types.type:named:error" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:interface:{Error:func:{}{basic:string}}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:error", i32 ptrtoint (ptr @"interface:{Error:func:{}{basic:string}}.$typeassert" to i32) }
@"reflect/types.type:interface:{Error:func:{}{basic:string}}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.interface:interface{Error() string}$interface", i32 0, ptr null, ptr @"reflect/types.type:pointer:interface:{Error:func:{}{basic:string}}", i32 ptrtoint (ptr @"interface:{Error:func:{}{basic:string}}.$typeassert" to i32) }
@"reflect/methods.Error() string" = linkonce_odr constant i8 0, align 1
@"reflect/types.interface:interface{Error() string}$interface" = linkonce_odr constant [1 x ptr] [ptr @"reflect/methods.Error() string"]
@"reflect/types.type:pointer:interface:{Error:func:{}{basic:string}}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:interface:{Error:func:{}{basic:string}}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.type:pointer:named:error" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:error", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.structFieldName.32" = private unnamed_addr global [3 x i8] c"err"
@"struct{*os.File; once sync.Once; err error}$methodset" = linkonce_odr constant [17 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.Close() error", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).Close$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Fd() uintptr", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).Fd$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Name() string", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).Name$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Read([]uint8) (int, error)", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).Read$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadAt([]uint8, int64) (int, error)", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).ReadAt$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadDir(int) ([]io/fs.DirEntry, error)", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).ReadDir$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Readdir(int) ([]io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).Readdir$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Readdirnames(int) ([]string, error)", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).Readdirnames$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Seek(int64, int) (int64, error)", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).Seek$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Stat() (io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).Stat$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Sync() error", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).Sync$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.SyscallConn() (syscall.RawConn, error)", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).SyscallConn$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Truncate(int64) error", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).Truncate$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Write([]uint8) (int, error)", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).Write$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.WriteAt([]uint8, int64) (int, error)", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).WriteAt$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.WriteString(string) (int, error)", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).WriteString$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"os.$methods.readdir(int, os.readdirMode) ([]string, []io/fs.DirEntry, []io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).readdir$invoke" to i32) }]
@"reflect/types.type:pointer:struct:{#File:pointer:named:os.File,once:named:sync.Once,err:named:error}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{#File:pointer:named:os.File,once:named:sync.Once,err:named:error}", i32 0, ptr @"*struct{*os.File; once sync.Once; err error}$methodset", ptr null, i32 0 }
@"*struct{*os.File; once sync.Once; err error}$methodset" = linkonce_odr constant [17 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.Close() error", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).Close" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Fd() uintptr", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).Fd" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Name() string", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).Name" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Read([]uint8) (int, error)", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).Read" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadAt([]uint8, int64) (int, error)", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).ReadAt" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadDir(int) ([]io/fs.DirEntry, error)", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).ReadDir" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Readdir(int) ([]io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).Readdir" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Readdirnames(int) ([]string, error)", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).Readdirnames" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Seek(int64, int) (int64, error)", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).Seek" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Stat() (io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).Stat" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Sync() error", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).Sync" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.SyscallConn() (syscall.RawConn, error)", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).SyscallConn" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Truncate(int64) error", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).Truncate" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Write([]uint8) (int, error)", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).Write" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.WriteAt([]uint8, int64) (int, error)", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).WriteAt" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.WriteString(string) (int, error)", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).WriteString" to i32) }, %runtime.interfaceMethodInfo { ptr @"os.$methods.readdir(int, os.readdirMode) ([]string, []io/fs.DirEntry, []io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).readdir" to i32) }]
@"main.closeOnce$methodset" = linkonce_odr constant [16 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.Fd() uintptr", i32 ptrtoint (ptr @"(main.closeOnce).Fd$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Name() string", i32 ptrtoint (ptr @"(main.closeOnce).Name$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Read([]uint8) (int, error)", i32 ptrtoint (ptr @"(main.closeOnce).Read$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadAt([]uint8, int64) (int, error)", i32 ptrtoint (ptr @"(main.closeOnce).ReadAt$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadDir(int) ([]io/fs.DirEntry, error)", i32 ptrtoint (ptr @"(main.closeOnce).ReadDir$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Readdir(int) ([]io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(main.closeOnce).Readdir$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Readdirnames(int) ([]string, error)", i32 ptrtoint (ptr @"(main.closeOnce).Readdirnames$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Seek(int64, int) (int64, error)", i32 ptrtoint (ptr @"(main.closeOnce).Seek$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Stat() (io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(main.closeOnce).Stat$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Sync() error", i32 ptrtoint (ptr @"(main.closeOnce).Sync$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.SyscallConn() (syscall.RawConn, error)", i32 ptrtoint (ptr @"(main.closeOnce).SyscallConn$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Truncate(int64) error", i32 ptrtoint (ptr @"(main.closeOnce).Truncate$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Write([]uint8) (int, error)", i32 ptrtoint (ptr @"(main.closeOnce).Write$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.WriteAt([]uint8, int64) (int, error)", i32 ptrtoint (ptr @"(main.closeOnce).WriteAt$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.WriteString(string) (int, error)", i32 ptrtoint (ptr @"(main.closeOnce).WriteString$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"os.$methods.readdir(int, os.readdirMode) ([]string, []io/fs.DirEntry, []io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(main.closeOnce).readdir$invoke" to i32) }]
@"main.$methods.close()" = linkonce_odr constant i8 0, align 1
@"*main.closeOnce$methodset" = linkonce_odr constant [18 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.Close() error", i32 ptrtoint (ptr @"(*main.closeOnce).Close" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Fd() uintptr", i32 ptrtoint (ptr @"(*main.closeOnce).Fd" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Name() string", i32 ptrtoint (ptr @"(*main.closeOnce).Name" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Read([]uint8) (int, error)", i32 ptrtoint (ptr @"(*main.closeOnce).Read" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadAt([]uint8, int64) (int, error)", i32 ptrtoint (ptr @"(*main.closeOnce).ReadAt" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadDir(int) ([]io/fs.DirEntry, error)", i32 ptrtoint (ptr @"(*main.closeOnce).ReadDir" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Readdir(int) ([]io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(*main.closeOnce).Readdir" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Readdirnames(int) ([]string, error)", i32 ptrtoint (ptr @"(*main.closeOnce).Readdirnames" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Seek(int64, int) (int64, error)", i32 ptrtoint (ptr @"(*main.closeOnce).Seek" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Stat() (io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(*main.closeOnce).Stat" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Sync() error", i32 ptrtoint (ptr @"(*main.closeOnce).Sync" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.SyscallConn() (syscall.RawConn, error)", i32 ptrtoint (ptr @"(*main.closeOnce).SyscallConn" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Truncate(int64) error", i32 ptrtoint (ptr @"(*main.closeOnce).Truncate" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Write([]uint8) (int, error)", i32 ptrtoint (ptr @"(*main.closeOnce).Write" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.WriteAt([]uint8, int64) (int, error)", i32 ptrtoint (ptr @"(*main.closeOnce).WriteAt" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.WriteString(string) (int, error)", i32 ptrtoint (ptr @"(*main.closeOnce).WriteString" to i32) }, %runtime.interfaceMethodInfo { ptr @"main.$methods.close()", i32 ptrtoint (ptr @"(*main.closeOnce).close" to i32) }, %runtime.interfaceMethodInfo { ptr @"os.$methods.readdir(int, os.readdirMode) ([]string, []io/fs.DirEntry, []io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(*main.closeOnce).readdir" to i32) }]

declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #0

declare void @runtime.trackPointer(ptr nocapture readonly, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden { %runtime._interface, %runtime._interface } @"(*main.Cmd).StdinPipe"(ptr dereferenceable_or_null(32) %c, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %c, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %.unpack = load i32, ptr %c, align 4
  %.elt34 = getelementptr inbounds %runtime._interface, ptr %c, i32 0, i32 1
  %.unpack35 = load ptr, ptr %.elt34, align 4
  call void @runtime.trackPointer(ptr %.unpack35, ptr undef) #8
  %.not = icmp eq i32 %.unpack, 0
  br i1 %.not, label %if.done, label %if.then

if.then:                                          ; preds = %deref.next
  %1 = call %runtime._interface @errors.New(ptr nonnull @"main$string", i32 23, ptr undef) #8
  %2 = extractvalue %runtime._interface %1, 1
  call void @runtime.trackPointer(ptr %2, ptr undef) #8
  %3 = insertvalue { %runtime._interface, %runtime._interface } zeroinitializer, %runtime._interface %1, 1
  ret { %runtime._interface, %runtime._interface } %3

if.done:                                          ; preds = %deref.next
  %4 = call { ptr, ptr, %runtime._interface } @os.Pipe(ptr undef) #8
  %5 = extractvalue { ptr, ptr, %runtime._interface } %4, 0
  call void @runtime.trackPointer(ptr %5, ptr undef) #8
  %6 = extractvalue { ptr, ptr, %runtime._interface } %4, 1
  call void @runtime.trackPointer(ptr %6, ptr undef) #8
  %7 = extractvalue { ptr, ptr, %runtime._interface } %4, 2
  %8 = extractvalue %runtime._interface %7, 1
  call void @runtime.trackPointer(ptr %8, ptr undef) #8
  %9 = extractvalue { ptr, ptr, %runtime._interface } %4, 0
  %10 = extractvalue { ptr, ptr, %runtime._interface } %4, 1
  %11 = extractvalue { ptr, ptr, %runtime._interface } %4, 2
  %12 = extractvalue %runtime._interface %11, 0
  %.not36 = icmp eq i32 %12, 0
  br i1 %.not36, label %if.done2, label %if.then1

if.then1:                                         ; preds = %if.done
  %13 = insertvalue { %runtime._interface, %runtime._interface } zeroinitializer, %runtime._interface %11, 1
  ret { %runtime._interface, %runtime._interface } %13

if.done2:                                         ; preds = %if.done
  br i1 false, label %gep.throw3, label %gep.next4

gep.next4:                                        ; preds = %if.done2
  call void @runtime.trackPointer(ptr %9, ptr undef) #8
  br i1 false, label %store.throw, label %store.next

store.next:                                       ; preds = %gep.next4
  store i32 ptrtoint (ptr @"reflect/types.type:pointer:named:os.File" to i32), ptr %c, align 4
  %c.repack37 = getelementptr inbounds %runtime._interface, ptr %c, i32 0, i32 1
  store ptr %9, ptr %c.repack37, align 4
  br i1 false, label %gep.throw5, label %gep.next6

gep.next6:                                        ; preds = %store.next
  %14 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 1
  br i1 false, label %gep.throw7, label %gep.next8

gep.next8:                                        ; preds = %gep.next6
  br i1 false, label %deref.throw9, label %deref.next10

deref.next10:                                     ; preds = %gep.next8
  %15 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 1
  %.unpack39 = load ptr, ptr %15, align 4
  %.elt40 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 1, i32 1
  %.unpack41 = load i32, ptr %.elt40, align 4
  %.elt42 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 1, i32 2
  %.unpack43 = load i32, ptr %.elt42, align 4
  call void @runtime.trackPointer(ptr %.unpack39, ptr undef) #8
  %varargs = call ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 133 to ptr), ptr undef) #8
  call void @runtime.trackPointer(ptr nonnull %varargs, ptr undef) #8
  call void @runtime.trackPointer(ptr %9, ptr undef) #8
  store i32 ptrtoint (ptr @"reflect/types.type:pointer:named:os.File" to i32), ptr %varargs, align 4
  %varargs.repack44 = getelementptr inbounds %runtime._interface, ptr %varargs, i32 0, i32 1
  store ptr %9, ptr %varargs.repack44, align 4
  %append.new = call { ptr, i32, i32 } @runtime.sliceAppend(ptr %.unpack39, ptr nonnull %varargs, i32 %.unpack41, i32 %.unpack43, i32 1, i32 8, ptr undef) #8
  %append.newPtr = extractvalue { ptr, i32, i32 } %append.new, 0
  call void @runtime.trackPointer(ptr %append.newPtr, ptr undef) #8
  br i1 false, label %store.throw11, label %store.next12

store.next12:                                     ; preds = %deref.next10
  %append.newLen = extractvalue { ptr, i32, i32 } %append.new, 1
  %append.newCap = extractvalue { ptr, i32, i32 } %append.new, 2
  store ptr %append.newPtr, ptr %14, align 4
  %.repack46 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 1, i32 1
  store i32 %append.newLen, ptr %.repack46, align 4
  %.repack48 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 1, i32 2
  store i32 %append.newCap, ptr %.repack48, align 4
  %complit = call ptr @runtime.alloc(i32 24, ptr nonnull inttoptr (i32 2637 to ptr), ptr undef) #8
  call void @runtime.trackPointer(ptr nonnull %complit, ptr undef) #8
  br i1 false, label %store.throw13, label %store.next14

store.next14:                                     ; preds = %store.next12
  store ptr %10, ptr %complit, align 4
  br i1 false, label %gep.throw15, label %gep.next16

gep.next16:                                       ; preds = %store.next14
  br i1 false, label %gep.throw17, label %gep.next18

gep.next18:                                       ; preds = %gep.next16
  br i1 false, label %deref.throw19, label %deref.next20

deref.next20:                                     ; preds = %gep.next18
  %16 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 2
  %.unpack50 = load ptr, ptr %16, align 4
  %.elt51 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 2, i32 1
  %.unpack52 = load i32, ptr %.elt51, align 4
  %.elt53 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 2, i32 2
  %.unpack54 = load i32, ptr %.elt53, align 4
  call void @runtime.trackPointer(ptr %.unpack50, ptr undef) #8
  %varargs21 = call ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 133 to ptr), ptr undef) #8
  call void @runtime.trackPointer(ptr nonnull %varargs21, ptr undef) #8
  call void @runtime.trackPointer(ptr nonnull %complit, ptr undef) #8
  store i32 ptrtoint (ptr @"reflect/types.type:pointer:named:main.closeOnce" to i32), ptr %varargs21, align 4
  %varargs21.repack55 = getelementptr inbounds %runtime._interface, ptr %varargs21, i32 0, i32 1
  store ptr %complit, ptr %varargs21.repack55, align 4
  %append.new28 = call { ptr, i32, i32 } @runtime.sliceAppend(ptr %.unpack50, ptr nonnull %varargs21, i32 %.unpack52, i32 %.unpack54, i32 1, i32 8, ptr undef) #8
  %append.newPtr29 = extractvalue { ptr, i32, i32 } %append.new28, 0
  call void @runtime.trackPointer(ptr %append.newPtr29, ptr undef) #8
  br i1 false, label %store.throw32, label %store.next33

store.next33:                                     ; preds = %deref.next20
  %17 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 2
  %append.newLen30 = extractvalue { ptr, i32, i32 } %append.new28, 1
  %append.newCap31 = extractvalue { ptr, i32, i32 } %append.new28, 2
  store ptr %append.newPtr29, ptr %17, align 4
  %.repack57 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 2, i32 1
  store i32 %append.newLen30, ptr %.repack57, align 4
  %.repack59 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 2, i32 2
  store i32 %append.newCap31, ptr %.repack59, align 4
  %18 = insertvalue %runtime._interface { i32 ptrtoint (ptr @"reflect/types.type:pointer:named:main.closeOnce" to i32), ptr undef }, ptr %complit, 1
  call void @runtime.trackPointer(ptr nonnull %complit, ptr undef) #8
  %19 = insertvalue { %runtime._interface, %runtime._interface } zeroinitializer, %runtime._interface %18, 0
  %20 = insertvalue { %runtime._interface, %runtime._interface } %19, %runtime._interface zeroinitializer, 1
  ret { %runtime._interface, %runtime._interface } %20

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable

gep.throw3:                                       ; preds = %if.done2
  unreachable

store.throw:                                      ; preds = %gep.next4
  unreachable

gep.throw5:                                       ; preds = %store.next
  unreachable

gep.throw7:                                       ; preds = %gep.next6
  unreachable

deref.throw9:                                     ; preds = %gep.next8
  unreachable

store.throw11:                                    ; preds = %deref.next10
  unreachable

store.throw13:                                    ; preds = %store.next12
  unreachable

gep.throw15:                                      ; preds = %store.next14
  unreachable

gep.throw17:                                      ; preds = %gep.next16
  unreachable

deref.throw19:                                    ; preds = %gep.next18
  unreachable

store.throw32:                                    ; preds = %deref.next20
  unreachable
}

declare void @runtime.nilPanic(ptr) #0

declare %runtime._interface @errors.New(ptr, i32, ptr) #0

declare { ptr, ptr, %runtime._interface } @os.Pipe(ptr) #0

declare i1 @"interface:{Close:func:{}{named:error},Read:func:{slice:basic:uint8}{basic:int,named:error},ReadAt:func:{slice:basic:uint8,basic:int64}{basic:int,named:error},Seek:func:{basic:int64,basic:int}{basic:int64,named:error},Write:func:{slice:basic:uint8}{basic:int,named:error}}.$typeassert"(i32) #2

declare %runtime._interface @"(*os.File).Close"(ptr dereferenceable_or_null(4), ptr) #0

declare i32 @"(*os.File).Fd"(ptr dereferenceable_or_null(4), ptr) #0

declare %runtime._string @"(*os.File).Name"(ptr dereferenceable_or_null(4), ptr) #0

declare { i32, %runtime._interface } @"(*os.File).Read"(ptr dereferenceable_or_null(4), ptr, i32, i32, ptr) #0

declare { i32, %runtime._interface } @"(*os.File).ReadAt"(ptr dereferenceable_or_null(4), ptr, i32, i32, i64, ptr) #0

declare { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).ReadDir"(ptr dereferenceable_or_null(4), i32, ptr) #0

declare { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).Readdir"(ptr dereferenceable_or_null(4), i32, ptr) #0

declare { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).Readdirnames"(ptr dereferenceable_or_null(4), i32, ptr) #0

declare { i64, %runtime._interface } @"(*os.File).Seek"(ptr dereferenceable_or_null(4), i64, i32, ptr) #0

declare { %runtime._interface, %runtime._interface } @"(*os.File).Stat"(ptr dereferenceable_or_null(4), ptr) #0

declare %runtime._interface @"(*os.File).Sync"(ptr dereferenceable_or_null(4), ptr) #0

declare { %runtime._interface, %runtime._interface } @"(*os.File).SyscallConn"(ptr dereferenceable_or_null(4), ptr) #0

declare %runtime._interface @"(*os.File).Truncate"(ptr dereferenceable_or_null(4), i64, ptr) #0

declare { i32, %runtime._interface } @"(*os.File).Write"(ptr dereferenceable_or_null(4), ptr, i32, i32, ptr) #0

declare { i32, %runtime._interface } @"(*os.File).WriteAt"(ptr dereferenceable_or_null(4), ptr, i32, i32, i64, ptr) #0

declare { i32, %runtime._interface } @"(*os.File).WriteString"(ptr dereferenceable_or_null(4), ptr, i32, ptr) #0

declare { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(*os.File).readdir"(ptr dereferenceable_or_null(4), i32, i32, ptr) #0

declare { ptr, i32, i32 } @runtime.sliceAppend(ptr, ptr nocapture readonly, i32, i32, i32, i32, ptr) #0

declare void @"(*internal/task.gcData).swap"(ptr dereferenceable_or_null(4), ptr) #0

declare void @tinygo_unwind(ptr nocapture dereferenceable_or_null(8)) #3

; Function Attrs: nounwind
define linkonce_odr hidden void @"(*struct{entry uintptr; args unsafe.Pointer; internal/task.stackState; launched bool}).unwind"(ptr dereferenceable_or_null(20) %arg0, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %arg0, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  %1 = getelementptr inbounds { i32, ptr, %"internal/task.stackState", i1 }, ptr %arg0, i32 0, i32 2
  call void @tinygo_unwind(ptr nonnull %1) #8
  ret void

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable
}

declare void @"(*internal/task.state).initialize"(ptr dereferenceable_or_null(20), i32, ptr, i32, ptr) #0

declare void @tinygo_launch(ptr nocapture dereferenceable_or_null(20)) #4

declare void @tinygo_rewind(ptr nocapture dereferenceable_or_null(20)) #5

; Function Attrs: nounwind
define linkonce_odr hidden void @"(*internal/task.state).unwind"(ptr dereferenceable_or_null(20) %arg0, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %arg0, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  %1 = getelementptr inbounds %"internal/task.state", ptr %arg0, i32 0, i32 2
  call void @tinygo_unwind(ptr nonnull %1) #8
  ret void

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable
}

declare void @"(*internal/task.Task).Resume"(ptr dereferenceable_or_null(48), ptr) #0

declare ptr @"(*internal/task.Task).tail"(ptr dereferenceable_or_null(48), ptr) #0

declare ptr @"(*internal/task.Stack).Pop"(ptr dereferenceable_or_null(4), ptr) #0

declare void @"(*internal/task.Stack).Push"(ptr dereferenceable_or_null(4), ptr dereferenceable_or_null(48), ptr) #0

declare %"internal/task.Queue" @"(*internal/task.Stack).Queue"(ptr dereferenceable_or_null(4), ptr) #0

declare void @"(*sync.Mutex).Lock"(ptr dereferenceable_or_null(8), ptr) #0

declare void @"(*sync.Mutex).Unlock"(ptr dereferenceable_or_null(8), ptr) #0

declare void @"(*sync.Once).Do"(ptr dereferenceable_or_null(12), ptr, ptr, ptr) #0

declare i1 @"interface:{Error:func:{}{basic:string}}.$typeassert"(i32) #6

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(struct{*os.File; once sync.Once; err error}).Close"({ ptr, %sync.Once, %runtime._interface } %f, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call %runtime._interface @"(*os.File).Close"(ptr %0, ptr undef) #8
  %2 = extractvalue %runtime._interface %1, 1
  call void @runtime.trackPointer(ptr %2, ptr undef) #8
  ret %runtime._interface %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr %runtime._interface @"(struct{*os.File; once sync.Once; err error}).Close$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %2 = insertvalue { ptr, %sync.Once, %runtime._interface } undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %3 = insertvalue { ptr, %sync.Once, %runtime._interface } %2, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %4 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %4, ptr %.unpack4.unpack6, 1
  %5 = insertvalue { ptr, %sync.Once, %runtime._interface } %3, %runtime._interface %.unpack47, 2
  %ret = call %runtime._interface @"(struct{*os.File; once sync.Once; err error}).Close"({ ptr, %sync.Once, %runtime._interface } %5, ptr %1)
  ret %runtime._interface %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden i32 @"(struct{*os.File; once sync.Once; err error}).Fd"({ ptr, %sync.Once, %runtime._interface } %f, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call i32 @"(*os.File).Fd"(ptr %0, ptr undef) #8
  ret i32 %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr i32 @"(struct{*os.File; once sync.Once; err error}).Fd$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %2 = insertvalue { ptr, %sync.Once, %runtime._interface } undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %3 = insertvalue { ptr, %sync.Once, %runtime._interface } %2, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %4 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %4, ptr %.unpack4.unpack6, 1
  %5 = insertvalue { ptr, %sync.Once, %runtime._interface } %3, %runtime._interface %.unpack47, 2
  %ret = call i32 @"(struct{*os.File; once sync.Once; err error}).Fd"({ ptr, %sync.Once, %runtime._interface } %5, ptr %1)
  ret i32 %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._string @"(struct{*os.File; once sync.Once; err error}).Name"({ ptr, %sync.Once, %runtime._interface } %f, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call %runtime._string @"(*os.File).Name"(ptr %0, ptr undef) #8
  %2 = extractvalue %runtime._string %1, 0
  call void @runtime.trackPointer(ptr %2, ptr undef) #8
  ret %runtime._string %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr %runtime._string @"(struct{*os.File; once sync.Once; err error}).Name$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %2 = insertvalue { ptr, %sync.Once, %runtime._interface } undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %3 = insertvalue { ptr, %sync.Once, %runtime._interface } %2, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %4 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %4, ptr %.unpack4.unpack6, 1
  %5 = insertvalue { ptr, %sync.Once, %runtime._interface } %3, %runtime._interface %.unpack47, 2
  %ret = call %runtime._string @"(struct{*os.File; once sync.Once; err error}).Name"({ ptr, %sync.Once, %runtime._interface } %5, ptr %1)
  ret %runtime._string %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Read"({ ptr, %sync.Once, %runtime._interface } %f, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call { i32, %runtime._interface } @"(*os.File).Read"(ptr %0, ptr %b.data, i32 %b.len, i32 %b.cap, ptr undef) #8
  %2 = extractvalue { i32, %runtime._interface } %1, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  ret { i32, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Read$invoke"(ptr %0, ptr %1, i32 %2, i32 %3, ptr %4) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %5 = insertvalue { ptr, %sync.Once, %runtime._interface } undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %6 = insertvalue { ptr, %sync.Once, %runtime._interface } %5, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %7 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %7, ptr %.unpack4.unpack6, 1
  %8 = insertvalue { ptr, %sync.Once, %runtime._interface } %6, %runtime._interface %.unpack47, 2
  %ret = call { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Read"({ ptr, %sync.Once, %runtime._interface } %8, ptr %1, i32 %2, i32 %3, ptr %4)
  ret { i32, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).ReadAt"({ ptr, %sync.Once, %runtime._interface } %f, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %offset, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call { i32, %runtime._interface } @"(*os.File).ReadAt"(ptr %0, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %offset, ptr undef) #8
  %2 = extractvalue { i32, %runtime._interface } %1, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  ret { i32, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).ReadAt$invoke"(ptr %0, ptr %1, i32 %2, i32 %3, i64 %4, ptr %5) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %6 = insertvalue { ptr, %sync.Once, %runtime._interface } undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %7 = insertvalue { ptr, %sync.Once, %runtime._interface } %6, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %8 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %8, ptr %.unpack4.unpack6, 1
  %9 = insertvalue { ptr, %sync.Once, %runtime._interface } %7, %runtime._interface %.unpack47, 2
  %ret = call { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).ReadAt"({ ptr, %sync.Once, %runtime._interface } %9, ptr %1, i32 %2, i32 %3, i64 %4, ptr %5)
  ret { i32, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).ReadDir"({ ptr, %sync.Once, %runtime._interface } %f, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).ReadDir"(ptr %0, i32 %n, ptr undef) #8
  %2 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %1, 0
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  %4 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %1, 1
  %5 = extractvalue %runtime._interface %4, 1
  call void @runtime.trackPointer(ptr %5, ptr undef) #8
  ret { { ptr, i32, i32 }, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { { ptr, i32, i32 }, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).ReadDir$invoke"(ptr %0, i32 %1, ptr %2) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %3 = insertvalue { ptr, %sync.Once, %runtime._interface } undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %4 = insertvalue { ptr, %sync.Once, %runtime._interface } %3, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %5 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %5, ptr %.unpack4.unpack6, 1
  %6 = insertvalue { ptr, %sync.Once, %runtime._interface } %4, %runtime._interface %.unpack47, 2
  %ret = call { { ptr, i32, i32 }, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).ReadDir"({ ptr, %sync.Once, %runtime._interface } %6, i32 %1, ptr %2)
  ret { { ptr, i32, i32 }, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Readdir"({ ptr, %sync.Once, %runtime._interface } %f, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).Readdir"(ptr %0, i32 %n, ptr undef) #8
  %2 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %1, 0
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  %4 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %1, 1
  %5 = extractvalue %runtime._interface %4, 1
  call void @runtime.trackPointer(ptr %5, ptr undef) #8
  ret { { ptr, i32, i32 }, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { { ptr, i32, i32 }, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Readdir$invoke"(ptr %0, i32 %1, ptr %2) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %3 = insertvalue { ptr, %sync.Once, %runtime._interface } undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %4 = insertvalue { ptr, %sync.Once, %runtime._interface } %3, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %5 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %5, ptr %.unpack4.unpack6, 1
  %6 = insertvalue { ptr, %sync.Once, %runtime._interface } %4, %runtime._interface %.unpack47, 2
  %ret = call { { ptr, i32, i32 }, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Readdir"({ ptr, %sync.Once, %runtime._interface } %6, i32 %1, ptr %2)
  ret { { ptr, i32, i32 }, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Readdirnames"({ ptr, %sync.Once, %runtime._interface } %f, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).Readdirnames"(ptr %0, i32 %n, ptr undef) #8
  %2 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %1, 0
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  %4 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %1, 1
  %5 = extractvalue %runtime._interface %4, 1
  call void @runtime.trackPointer(ptr %5, ptr undef) #8
  ret { { ptr, i32, i32 }, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { { ptr, i32, i32 }, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Readdirnames$invoke"(ptr %0, i32 %1, ptr %2) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %3 = insertvalue { ptr, %sync.Once, %runtime._interface } undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %4 = insertvalue { ptr, %sync.Once, %runtime._interface } %3, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %5 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %5, ptr %.unpack4.unpack6, 1
  %6 = insertvalue { ptr, %sync.Once, %runtime._interface } %4, %runtime._interface %.unpack47, 2
  %ret = call { { ptr, i32, i32 }, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Readdirnames"({ ptr, %sync.Once, %runtime._interface } %6, i32 %1, ptr %2)
  ret { { ptr, i32, i32 }, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { i64, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Seek"({ ptr, %sync.Once, %runtime._interface } %f, i64 %offset, i32 %whence, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call { i64, %runtime._interface } @"(*os.File).Seek"(ptr %0, i64 %offset, i32 %whence, ptr undef) #8
  %2 = extractvalue { i64, %runtime._interface } %1, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  ret { i64, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { i64, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Seek$invoke"(ptr %0, i64 %1, i32 %2, ptr %3) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %4 = insertvalue { ptr, %sync.Once, %runtime._interface } undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %5 = insertvalue { ptr, %sync.Once, %runtime._interface } %4, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %6 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %6, ptr %.unpack4.unpack6, 1
  %7 = insertvalue { ptr, %sync.Once, %runtime._interface } %5, %runtime._interface %.unpack47, 2
  %ret = call { i64, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Seek"({ ptr, %sync.Once, %runtime._interface } %7, i64 %1, i32 %2, ptr %3)
  ret { i64, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { %runtime._interface, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Stat"({ ptr, %sync.Once, %runtime._interface } %f, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call { %runtime._interface, %runtime._interface } @"(*os.File).Stat"(ptr %0, ptr undef) #8
  %2 = extractvalue { %runtime._interface, %runtime._interface } %1, 0
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  %4 = extractvalue { %runtime._interface, %runtime._interface } %1, 1
  %5 = extractvalue %runtime._interface %4, 1
  call void @runtime.trackPointer(ptr %5, ptr undef) #8
  ret { %runtime._interface, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { %runtime._interface, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Stat$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %2 = insertvalue { ptr, %sync.Once, %runtime._interface } undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %3 = insertvalue { ptr, %sync.Once, %runtime._interface } %2, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %4 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %4, ptr %.unpack4.unpack6, 1
  %5 = insertvalue { ptr, %sync.Once, %runtime._interface } %3, %runtime._interface %.unpack47, 2
  %ret = call { %runtime._interface, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Stat"({ ptr, %sync.Once, %runtime._interface } %5, ptr %1)
  ret { %runtime._interface, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(struct{*os.File; once sync.Once; err error}).Sync"({ ptr, %sync.Once, %runtime._interface } %f, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call %runtime._interface @"(*os.File).Sync"(ptr %0, ptr undef) #8
  %2 = extractvalue %runtime._interface %1, 1
  call void @runtime.trackPointer(ptr %2, ptr undef) #8
  ret %runtime._interface %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr %runtime._interface @"(struct{*os.File; once sync.Once; err error}).Sync$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %2 = insertvalue { ptr, %sync.Once, %runtime._interface } undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %3 = insertvalue { ptr, %sync.Once, %runtime._interface } %2, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %4 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %4, ptr %.unpack4.unpack6, 1
  %5 = insertvalue { ptr, %sync.Once, %runtime._interface } %3, %runtime._interface %.unpack47, 2
  %ret = call %runtime._interface @"(struct{*os.File; once sync.Once; err error}).Sync"({ ptr, %sync.Once, %runtime._interface } %5, ptr %1)
  ret %runtime._interface %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { %runtime._interface, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).SyscallConn"({ ptr, %sync.Once, %runtime._interface } %f, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call { %runtime._interface, %runtime._interface } @"(*os.File).SyscallConn"(ptr %0, ptr undef) #8
  %2 = extractvalue { %runtime._interface, %runtime._interface } %1, 0
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  %4 = extractvalue { %runtime._interface, %runtime._interface } %1, 1
  %5 = extractvalue %runtime._interface %4, 1
  call void @runtime.trackPointer(ptr %5, ptr undef) #8
  ret { %runtime._interface, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { %runtime._interface, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).SyscallConn$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %2 = insertvalue { ptr, %sync.Once, %runtime._interface } undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %3 = insertvalue { ptr, %sync.Once, %runtime._interface } %2, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %4 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %4, ptr %.unpack4.unpack6, 1
  %5 = insertvalue { ptr, %sync.Once, %runtime._interface } %3, %runtime._interface %.unpack47, 2
  %ret = call { %runtime._interface, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).SyscallConn"({ ptr, %sync.Once, %runtime._interface } %5, ptr %1)
  ret { %runtime._interface, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(struct{*os.File; once sync.Once; err error}).Truncate"({ ptr, %sync.Once, %runtime._interface } %f, i64 %size, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call %runtime._interface @"(*os.File).Truncate"(ptr %0, i64 %size, ptr undef) #8
  %2 = extractvalue %runtime._interface %1, 1
  call void @runtime.trackPointer(ptr %2, ptr undef) #8
  ret %runtime._interface %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr %runtime._interface @"(struct{*os.File; once sync.Once; err error}).Truncate$invoke"(ptr %0, i64 %1, ptr %2) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %3 = insertvalue { ptr, %sync.Once, %runtime._interface } undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %4 = insertvalue { ptr, %sync.Once, %runtime._interface } %3, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %5 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %5, ptr %.unpack4.unpack6, 1
  %6 = insertvalue { ptr, %sync.Once, %runtime._interface } %4, %runtime._interface %.unpack47, 2
  %ret = call %runtime._interface @"(struct{*os.File; once sync.Once; err error}).Truncate"({ ptr, %sync.Once, %runtime._interface } %6, i64 %1, ptr %2)
  ret %runtime._interface %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Write"({ ptr, %sync.Once, %runtime._interface } %f, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call { i32, %runtime._interface } @"(*os.File).Write"(ptr %0, ptr %b.data, i32 %b.len, i32 %b.cap, ptr undef) #8
  %2 = extractvalue { i32, %runtime._interface } %1, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  ret { i32, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Write$invoke"(ptr %0, ptr %1, i32 %2, i32 %3, ptr %4) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %5 = insertvalue { ptr, %sync.Once, %runtime._interface } undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %6 = insertvalue { ptr, %sync.Once, %runtime._interface } %5, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %7 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %7, ptr %.unpack4.unpack6, 1
  %8 = insertvalue { ptr, %sync.Once, %runtime._interface } %6, %runtime._interface %.unpack47, 2
  %ret = call { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Write"({ ptr, %sync.Once, %runtime._interface } %8, ptr %1, i32 %2, i32 %3, ptr %4)
  ret { i32, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).WriteAt"({ ptr, %sync.Once, %runtime._interface } %f, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %off, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call { i32, %runtime._interface } @"(*os.File).WriteAt"(ptr %0, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %off, ptr undef) #8
  %2 = extractvalue { i32, %runtime._interface } %1, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  ret { i32, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).WriteAt$invoke"(ptr %0, ptr %1, i32 %2, i32 %3, i64 %4, ptr %5) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %6 = insertvalue { ptr, %sync.Once, %runtime._interface } undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %7 = insertvalue { ptr, %sync.Once, %runtime._interface } %6, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %8 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %8, ptr %.unpack4.unpack6, 1
  %9 = insertvalue { ptr, %sync.Once, %runtime._interface } %7, %runtime._interface %.unpack47, 2
  %ret = call { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).WriteAt"({ ptr, %sync.Once, %runtime._interface } %9, ptr %1, i32 %2, i32 %3, i64 %4, ptr %5)
  ret { i32, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).WriteString"({ ptr, %sync.Once, %runtime._interface } %f, ptr %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call { i32, %runtime._interface } @"(*os.File).WriteString"(ptr %0, ptr %s.data, i32 %s.len, ptr undef) #8
  %2 = extractvalue { i32, %runtime._interface } %1, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  ret { i32, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).WriteString$invoke"(ptr %0, ptr %1, i32 %2, ptr %3) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %4 = insertvalue { ptr, %sync.Once, %runtime._interface } undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %5 = insertvalue { ptr, %sync.Once, %runtime._interface } %4, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %6 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %6, ptr %.unpack4.unpack6, 1
  %7 = insertvalue { ptr, %sync.Once, %runtime._interface } %5, %runtime._interface %.unpack47, 2
  %ret = call { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).WriteString"({ ptr, %sync.Once, %runtime._interface } %7, ptr %1, i32 %2, ptr %3)
  ret { i32, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).readdir"({ ptr, %sync.Once, %runtime._interface } %f, i32 %n, i32 %mode, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue { ptr, %sync.Once, %runtime._interface } %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(*os.File).readdir"(ptr %0, i32 %n, i32 %mode, ptr undef) #8
  %2 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %1, 0
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  %4 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %1, 1
  %5 = extractvalue { ptr, i32, i32 } %4, 0
  call void @runtime.trackPointer(ptr %5, ptr undef) #8
  %6 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %1, 2
  %7 = extractvalue { ptr, i32, i32 } %6, 0
  call void @runtime.trackPointer(ptr %7, ptr undef) #8
  %8 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %1, 3
  %9 = extractvalue %runtime._interface %8, 1
  call void @runtime.trackPointer(ptr %9, ptr undef) #8
  %10 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %1, 0
  %11 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %1, 1
  %12 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %1, 2
  %13 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %1, 3
  %14 = insertvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } zeroinitializer, { ptr, i32, i32 } %10, 0
  %15 = insertvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %14, { ptr, i32, i32 } %11, 1
  %16 = insertvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %15, { ptr, i32, i32 } %12, 2
  %17 = insertvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %16, %runtime._interface %13, 3
  ret { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %17

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).readdir$invoke"(ptr %0, i32 %1, i32 %2, ptr %3) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %4 = insertvalue { ptr, %sync.Once, %runtime._interface } undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %5 = insertvalue { ptr, %sync.Once, %runtime._interface } %4, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %6 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %6, ptr %.unpack4.unpack6, 1
  %7 = insertvalue { ptr, %sync.Once, %runtime._interface } %5, %runtime._interface %.unpack47, 2
  %ret = call { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).readdir"({ ptr, %sync.Once, %runtime._interface } %7, i32 %1, i32 %2, ptr %3)
  ret { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(*struct{*os.File; once sync.Once; err error}).Close"(ptr dereferenceable_or_null(24) %f, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call %runtime._interface @"(*os.File).Close"(ptr %1, ptr undef) #8
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  ret %runtime._interface %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden i32 @"(*struct{*os.File; once sync.Once; err error}).Fd"(ptr dereferenceable_or_null(24) %f, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call i32 @"(*os.File).Fd"(ptr %1, ptr undef) #8
  ret i32 %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._string @"(*struct{*os.File; once sync.Once; err error}).Name"(ptr dereferenceable_or_null(24) %f, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call %runtime._string @"(*os.File).Name"(ptr %1, ptr undef) #8
  %3 = extractvalue %runtime._string %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  ret %runtime._string %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(*struct{*os.File; once sync.Once; err error}).Read"(ptr dereferenceable_or_null(24) %f, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call { i32, %runtime._interface } @"(*os.File).Read"(ptr %1, ptr %b.data, i32 %b.len, i32 %b.cap, ptr undef) #8
  %3 = extractvalue { i32, %runtime._interface } %2, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  ret { i32, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(*struct{*os.File; once sync.Once; err error}).ReadAt"(ptr dereferenceable_or_null(24) %f, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %offset, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call { i32, %runtime._interface } @"(*os.File).ReadAt"(ptr %1, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %offset, ptr undef) #8
  %3 = extractvalue { i32, %runtime._interface } %2, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  ret { i32, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, %runtime._interface } @"(*struct{*os.File; once sync.Once; err error}).ReadDir"(ptr dereferenceable_or_null(24) %f, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).ReadDir"(ptr %1, i32 %n, ptr undef) #8
  %3 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %2, 0
  %4 = extractvalue { ptr, i32, i32 } %3, 0
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  %5 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %2, 1
  %6 = extractvalue %runtime._interface %5, 1
  call void @runtime.trackPointer(ptr %6, ptr undef) #8
  ret { { ptr, i32, i32 }, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, %runtime._interface } @"(*struct{*os.File; once sync.Once; err error}).Readdir"(ptr dereferenceable_or_null(24) %f, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).Readdir"(ptr %1, i32 %n, ptr undef) #8
  %3 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %2, 0
  %4 = extractvalue { ptr, i32, i32 } %3, 0
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  %5 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %2, 1
  %6 = extractvalue %runtime._interface %5, 1
  call void @runtime.trackPointer(ptr %6, ptr undef) #8
  ret { { ptr, i32, i32 }, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, %runtime._interface } @"(*struct{*os.File; once sync.Once; err error}).Readdirnames"(ptr dereferenceable_or_null(24) %f, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).Readdirnames"(ptr %1, i32 %n, ptr undef) #8
  %3 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %2, 0
  %4 = extractvalue { ptr, i32, i32 } %3, 0
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  %5 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %2, 1
  %6 = extractvalue %runtime._interface %5, 1
  call void @runtime.trackPointer(ptr %6, ptr undef) #8
  ret { { ptr, i32, i32 }, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { i64, %runtime._interface } @"(*struct{*os.File; once sync.Once; err error}).Seek"(ptr dereferenceable_or_null(24) %f, i64 %offset, i32 %whence, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call { i64, %runtime._interface } @"(*os.File).Seek"(ptr %1, i64 %offset, i32 %whence, ptr undef) #8
  %3 = extractvalue { i64, %runtime._interface } %2, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  ret { i64, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { %runtime._interface, %runtime._interface } @"(*struct{*os.File; once sync.Once; err error}).Stat"(ptr dereferenceable_or_null(24) %f, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call { %runtime._interface, %runtime._interface } @"(*os.File).Stat"(ptr %1, ptr undef) #8
  %3 = extractvalue { %runtime._interface, %runtime._interface } %2, 0
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  %5 = extractvalue { %runtime._interface, %runtime._interface } %2, 1
  %6 = extractvalue %runtime._interface %5, 1
  call void @runtime.trackPointer(ptr %6, ptr undef) #8
  ret { %runtime._interface, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(*struct{*os.File; once sync.Once; err error}).Sync"(ptr dereferenceable_or_null(24) %f, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call %runtime._interface @"(*os.File).Sync"(ptr %1, ptr undef) #8
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  ret %runtime._interface %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { %runtime._interface, %runtime._interface } @"(*struct{*os.File; once sync.Once; err error}).SyscallConn"(ptr dereferenceable_or_null(24) %f, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call { %runtime._interface, %runtime._interface } @"(*os.File).SyscallConn"(ptr %1, ptr undef) #8
  %3 = extractvalue { %runtime._interface, %runtime._interface } %2, 0
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  %5 = extractvalue { %runtime._interface, %runtime._interface } %2, 1
  %6 = extractvalue %runtime._interface %5, 1
  call void @runtime.trackPointer(ptr %6, ptr undef) #8
  ret { %runtime._interface, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(*struct{*os.File; once sync.Once; err error}).Truncate"(ptr dereferenceable_or_null(24) %f, i64 %size, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call %runtime._interface @"(*os.File).Truncate"(ptr %1, i64 %size, ptr undef) #8
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  ret %runtime._interface %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(*struct{*os.File; once sync.Once; err error}).Write"(ptr dereferenceable_or_null(24) %f, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call { i32, %runtime._interface } @"(*os.File).Write"(ptr %1, ptr %b.data, i32 %b.len, i32 %b.cap, ptr undef) #8
  %3 = extractvalue { i32, %runtime._interface } %2, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  ret { i32, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(*struct{*os.File; once sync.Once; err error}).WriteAt"(ptr dereferenceable_or_null(24) %f, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %off, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call { i32, %runtime._interface } @"(*os.File).WriteAt"(ptr %1, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %off, ptr undef) #8
  %3 = extractvalue { i32, %runtime._interface } %2, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  ret { i32, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(*struct{*os.File; once sync.Once; err error}).WriteString"(ptr dereferenceable_or_null(24) %f, ptr %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call { i32, %runtime._interface } @"(*os.File).WriteString"(ptr %1, ptr %s.data, i32 %s.len, ptr undef) #8
  %3 = extractvalue { i32, %runtime._interface } %2, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  ret { i32, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(*struct{*os.File; once sync.Once; err error}).readdir"(ptr dereferenceable_or_null(24) %f, i32 %n, i32 %mode, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(*os.File).readdir"(ptr %1, i32 %n, i32 %mode, ptr undef) #8
  %3 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %2, 0
  %4 = extractvalue { ptr, i32, i32 } %3, 0
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  %5 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %2, 1
  %6 = extractvalue { ptr, i32, i32 } %5, 0
  call void @runtime.trackPointer(ptr %6, ptr undef) #8
  %7 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %2, 2
  %8 = extractvalue { ptr, i32, i32 } %7, 0
  call void @runtime.trackPointer(ptr %8, ptr undef) #8
  %9 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %2, 3
  %10 = extractvalue %runtime._interface %9, 1
  call void @runtime.trackPointer(ptr %10, ptr undef) #8
  %11 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %2, 0
  %12 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %2, 1
  %13 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %2, 2
  %14 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %2, 3
  %15 = insertvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } zeroinitializer, { ptr, i32, i32 } %11, 0
  %16 = insertvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %15, { ptr, i32, i32 } %12, 1
  %17 = insertvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %16, { ptr, i32, i32 } %13, 2
  %18 = insertvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %17, %runtime._interface %14, 3
  ret { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %18

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden i32 @"(main.closeOnce).Fd"(%main.closeOnce %f, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue %main.closeOnce %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue %main.closeOnce %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue %main.closeOnce %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call i32 @"(*os.File).Fd"(ptr %0, ptr undef) #8
  ret i32 %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr i32 @"(main.closeOnce).Fd$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %2 = insertvalue %main.closeOnce undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %3 = insertvalue %main.closeOnce %2, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %4 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %4, ptr %.unpack4.unpack6, 1
  %5 = insertvalue %main.closeOnce %3, %runtime._interface %.unpack47, 2
  %ret = call i32 @"(main.closeOnce).Fd"(%main.closeOnce %5, ptr %1)
  ret i32 %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._string @"(main.closeOnce).Name"(%main.closeOnce %f, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue %main.closeOnce %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue %main.closeOnce %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue %main.closeOnce %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call %runtime._string @"(*os.File).Name"(ptr %0, ptr undef) #8
  %2 = extractvalue %runtime._string %1, 0
  call void @runtime.trackPointer(ptr %2, ptr undef) #8
  ret %runtime._string %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr %runtime._string @"(main.closeOnce).Name$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %2 = insertvalue %main.closeOnce undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %3 = insertvalue %main.closeOnce %2, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %4 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %4, ptr %.unpack4.unpack6, 1
  %5 = insertvalue %main.closeOnce %3, %runtime._interface %.unpack47, 2
  %ret = call %runtime._string @"(main.closeOnce).Name"(%main.closeOnce %5, ptr %1)
  ret %runtime._string %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(main.closeOnce).Read"(%main.closeOnce %f, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue %main.closeOnce %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue %main.closeOnce %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue %main.closeOnce %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call { i32, %runtime._interface } @"(*os.File).Read"(ptr %0, ptr %b.data, i32 %b.len, i32 %b.cap, ptr undef) #8
  %2 = extractvalue { i32, %runtime._interface } %1, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  ret { i32, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { i32, %runtime._interface } @"(main.closeOnce).Read$invoke"(ptr %0, ptr %1, i32 %2, i32 %3, ptr %4) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %5 = insertvalue %main.closeOnce undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %6 = insertvalue %main.closeOnce %5, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %7 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %7, ptr %.unpack4.unpack6, 1
  %8 = insertvalue %main.closeOnce %6, %runtime._interface %.unpack47, 2
  %ret = call { i32, %runtime._interface } @"(main.closeOnce).Read"(%main.closeOnce %8, ptr %1, i32 %2, i32 %3, ptr %4)
  ret { i32, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(main.closeOnce).ReadAt"(%main.closeOnce %f, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %offset, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue %main.closeOnce %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue %main.closeOnce %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue %main.closeOnce %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call { i32, %runtime._interface } @"(*os.File).ReadAt"(ptr %0, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %offset, ptr undef) #8
  %2 = extractvalue { i32, %runtime._interface } %1, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  ret { i32, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { i32, %runtime._interface } @"(main.closeOnce).ReadAt$invoke"(ptr %0, ptr %1, i32 %2, i32 %3, i64 %4, ptr %5) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %6 = insertvalue %main.closeOnce undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %7 = insertvalue %main.closeOnce %6, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %8 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %8, ptr %.unpack4.unpack6, 1
  %9 = insertvalue %main.closeOnce %7, %runtime._interface %.unpack47, 2
  %ret = call { i32, %runtime._interface } @"(main.closeOnce).ReadAt"(%main.closeOnce %9, ptr %1, i32 %2, i32 %3, i64 %4, ptr %5)
  ret { i32, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, %runtime._interface } @"(main.closeOnce).ReadDir"(%main.closeOnce %f, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue %main.closeOnce %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue %main.closeOnce %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue %main.closeOnce %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).ReadDir"(ptr %0, i32 %n, ptr undef) #8
  %2 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %1, 0
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  %4 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %1, 1
  %5 = extractvalue %runtime._interface %4, 1
  call void @runtime.trackPointer(ptr %5, ptr undef) #8
  ret { { ptr, i32, i32 }, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { { ptr, i32, i32 }, %runtime._interface } @"(main.closeOnce).ReadDir$invoke"(ptr %0, i32 %1, ptr %2) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %3 = insertvalue %main.closeOnce undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %4 = insertvalue %main.closeOnce %3, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %5 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %5, ptr %.unpack4.unpack6, 1
  %6 = insertvalue %main.closeOnce %4, %runtime._interface %.unpack47, 2
  %ret = call { { ptr, i32, i32 }, %runtime._interface } @"(main.closeOnce).ReadDir"(%main.closeOnce %6, i32 %1, ptr %2)
  ret { { ptr, i32, i32 }, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, %runtime._interface } @"(main.closeOnce).Readdir"(%main.closeOnce %f, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue %main.closeOnce %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue %main.closeOnce %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue %main.closeOnce %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).Readdir"(ptr %0, i32 %n, ptr undef) #8
  %2 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %1, 0
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  %4 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %1, 1
  %5 = extractvalue %runtime._interface %4, 1
  call void @runtime.trackPointer(ptr %5, ptr undef) #8
  ret { { ptr, i32, i32 }, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { { ptr, i32, i32 }, %runtime._interface } @"(main.closeOnce).Readdir$invoke"(ptr %0, i32 %1, ptr %2) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %3 = insertvalue %main.closeOnce undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %4 = insertvalue %main.closeOnce %3, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %5 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %5, ptr %.unpack4.unpack6, 1
  %6 = insertvalue %main.closeOnce %4, %runtime._interface %.unpack47, 2
  %ret = call { { ptr, i32, i32 }, %runtime._interface } @"(main.closeOnce).Readdir"(%main.closeOnce %6, i32 %1, ptr %2)
  ret { { ptr, i32, i32 }, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, %runtime._interface } @"(main.closeOnce).Readdirnames"(%main.closeOnce %f, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue %main.closeOnce %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue %main.closeOnce %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue %main.closeOnce %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).Readdirnames"(ptr %0, i32 %n, ptr undef) #8
  %2 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %1, 0
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  %4 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %1, 1
  %5 = extractvalue %runtime._interface %4, 1
  call void @runtime.trackPointer(ptr %5, ptr undef) #8
  ret { { ptr, i32, i32 }, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { { ptr, i32, i32 }, %runtime._interface } @"(main.closeOnce).Readdirnames$invoke"(ptr %0, i32 %1, ptr %2) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %3 = insertvalue %main.closeOnce undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %4 = insertvalue %main.closeOnce %3, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %5 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %5, ptr %.unpack4.unpack6, 1
  %6 = insertvalue %main.closeOnce %4, %runtime._interface %.unpack47, 2
  %ret = call { { ptr, i32, i32 }, %runtime._interface } @"(main.closeOnce).Readdirnames"(%main.closeOnce %6, i32 %1, ptr %2)
  ret { { ptr, i32, i32 }, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { i64, %runtime._interface } @"(main.closeOnce).Seek"(%main.closeOnce %f, i64 %offset, i32 %whence, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue %main.closeOnce %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue %main.closeOnce %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue %main.closeOnce %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call { i64, %runtime._interface } @"(*os.File).Seek"(ptr %0, i64 %offset, i32 %whence, ptr undef) #8
  %2 = extractvalue { i64, %runtime._interface } %1, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  ret { i64, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { i64, %runtime._interface } @"(main.closeOnce).Seek$invoke"(ptr %0, i64 %1, i32 %2, ptr %3) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %4 = insertvalue %main.closeOnce undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %5 = insertvalue %main.closeOnce %4, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %6 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %6, ptr %.unpack4.unpack6, 1
  %7 = insertvalue %main.closeOnce %5, %runtime._interface %.unpack47, 2
  %ret = call { i64, %runtime._interface } @"(main.closeOnce).Seek"(%main.closeOnce %7, i64 %1, i32 %2, ptr %3)
  ret { i64, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { %runtime._interface, %runtime._interface } @"(main.closeOnce).Stat"(%main.closeOnce %f, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue %main.closeOnce %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue %main.closeOnce %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue %main.closeOnce %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call { %runtime._interface, %runtime._interface } @"(*os.File).Stat"(ptr %0, ptr undef) #8
  %2 = extractvalue { %runtime._interface, %runtime._interface } %1, 0
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  %4 = extractvalue { %runtime._interface, %runtime._interface } %1, 1
  %5 = extractvalue %runtime._interface %4, 1
  call void @runtime.trackPointer(ptr %5, ptr undef) #8
  ret { %runtime._interface, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { %runtime._interface, %runtime._interface } @"(main.closeOnce).Stat$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %2 = insertvalue %main.closeOnce undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %3 = insertvalue %main.closeOnce %2, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %4 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %4, ptr %.unpack4.unpack6, 1
  %5 = insertvalue %main.closeOnce %3, %runtime._interface %.unpack47, 2
  %ret = call { %runtime._interface, %runtime._interface } @"(main.closeOnce).Stat"(%main.closeOnce %5, ptr %1)
  ret { %runtime._interface, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(main.closeOnce).Sync"(%main.closeOnce %f, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue %main.closeOnce %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue %main.closeOnce %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue %main.closeOnce %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call %runtime._interface @"(*os.File).Sync"(ptr %0, ptr undef) #8
  %2 = extractvalue %runtime._interface %1, 1
  call void @runtime.trackPointer(ptr %2, ptr undef) #8
  ret %runtime._interface %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr %runtime._interface @"(main.closeOnce).Sync$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %2 = insertvalue %main.closeOnce undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %3 = insertvalue %main.closeOnce %2, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %4 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %4, ptr %.unpack4.unpack6, 1
  %5 = insertvalue %main.closeOnce %3, %runtime._interface %.unpack47, 2
  %ret = call %runtime._interface @"(main.closeOnce).Sync"(%main.closeOnce %5, ptr %1)
  ret %runtime._interface %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { %runtime._interface, %runtime._interface } @"(main.closeOnce).SyscallConn"(%main.closeOnce %f, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue %main.closeOnce %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue %main.closeOnce %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue %main.closeOnce %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call { %runtime._interface, %runtime._interface } @"(*os.File).SyscallConn"(ptr %0, ptr undef) #8
  %2 = extractvalue { %runtime._interface, %runtime._interface } %1, 0
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  %4 = extractvalue { %runtime._interface, %runtime._interface } %1, 1
  %5 = extractvalue %runtime._interface %4, 1
  call void @runtime.trackPointer(ptr %5, ptr undef) #8
  ret { %runtime._interface, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { %runtime._interface, %runtime._interface } @"(main.closeOnce).SyscallConn$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %2 = insertvalue %main.closeOnce undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %3 = insertvalue %main.closeOnce %2, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %4 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %4, ptr %.unpack4.unpack6, 1
  %5 = insertvalue %main.closeOnce %3, %runtime._interface %.unpack47, 2
  %ret = call { %runtime._interface, %runtime._interface } @"(main.closeOnce).SyscallConn"(%main.closeOnce %5, ptr %1)
  ret { %runtime._interface, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(main.closeOnce).Truncate"(%main.closeOnce %f, i64 %size, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue %main.closeOnce %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue %main.closeOnce %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue %main.closeOnce %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call %runtime._interface @"(*os.File).Truncate"(ptr %0, i64 %size, ptr undef) #8
  %2 = extractvalue %runtime._interface %1, 1
  call void @runtime.trackPointer(ptr %2, ptr undef) #8
  ret %runtime._interface %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr %runtime._interface @"(main.closeOnce).Truncate$invoke"(ptr %0, i64 %1, ptr %2) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %3 = insertvalue %main.closeOnce undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %4 = insertvalue %main.closeOnce %3, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %5 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %5, ptr %.unpack4.unpack6, 1
  %6 = insertvalue %main.closeOnce %4, %runtime._interface %.unpack47, 2
  %ret = call %runtime._interface @"(main.closeOnce).Truncate"(%main.closeOnce %6, i64 %1, ptr %2)
  ret %runtime._interface %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(main.closeOnce).Write"(%main.closeOnce %f, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue %main.closeOnce %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue %main.closeOnce %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue %main.closeOnce %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call { i32, %runtime._interface } @"(*os.File).Write"(ptr %0, ptr %b.data, i32 %b.len, i32 %b.cap, ptr undef) #8
  %2 = extractvalue { i32, %runtime._interface } %1, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  ret { i32, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { i32, %runtime._interface } @"(main.closeOnce).Write$invoke"(ptr %0, ptr %1, i32 %2, i32 %3, ptr %4) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %5 = insertvalue %main.closeOnce undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %6 = insertvalue %main.closeOnce %5, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %7 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %7, ptr %.unpack4.unpack6, 1
  %8 = insertvalue %main.closeOnce %6, %runtime._interface %.unpack47, 2
  %ret = call { i32, %runtime._interface } @"(main.closeOnce).Write"(%main.closeOnce %8, ptr %1, i32 %2, i32 %3, ptr %4)
  ret { i32, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(main.closeOnce).WriteAt"(%main.closeOnce %f, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %off, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue %main.closeOnce %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue %main.closeOnce %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue %main.closeOnce %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call { i32, %runtime._interface } @"(*os.File).WriteAt"(ptr %0, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %off, ptr undef) #8
  %2 = extractvalue { i32, %runtime._interface } %1, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  ret { i32, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { i32, %runtime._interface } @"(main.closeOnce).WriteAt$invoke"(ptr %0, ptr %1, i32 %2, i32 %3, i64 %4, ptr %5) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %6 = insertvalue %main.closeOnce undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %7 = insertvalue %main.closeOnce %6, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %8 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %8, ptr %.unpack4.unpack6, 1
  %9 = insertvalue %main.closeOnce %7, %runtime._interface %.unpack47, 2
  %ret = call { i32, %runtime._interface } @"(main.closeOnce).WriteAt"(%main.closeOnce %9, ptr %1, i32 %2, i32 %3, i64 %4, ptr %5)
  ret { i32, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(main.closeOnce).WriteString"(%main.closeOnce %f, ptr %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue %main.closeOnce %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue %main.closeOnce %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue %main.closeOnce %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call { i32, %runtime._interface } @"(*os.File).WriteString"(ptr %0, ptr %s.data, i32 %s.len, ptr undef) #8
  %2 = extractvalue { i32, %runtime._interface } %1, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  ret { i32, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { i32, %runtime._interface } @"(main.closeOnce).WriteString$invoke"(ptr %0, ptr %1, i32 %2, ptr %3) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %4 = insertvalue %main.closeOnce undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %5 = insertvalue %main.closeOnce %4, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %6 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %6, ptr %.unpack4.unpack6, 1
  %7 = insertvalue %main.closeOnce %5, %runtime._interface %.unpack47, 2
  %ret = call { i32, %runtime._interface } @"(main.closeOnce).WriteString"(%main.closeOnce %7, ptr %1, i32 %2, ptr %3)
  ret { i32, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(main.closeOnce).readdir"(%main.closeOnce %f, i32 %n, i32 %mode, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #8
  %f.elt = extractvalue %main.closeOnce %f, 0
  store ptr %f.elt, ptr %f1, align 8
  %f1.repack5 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  %f.elt6 = extractvalue %main.closeOnce %f, 1
  store %sync.Once %f.elt6, ptr %f1.repack5, align 4
  %f1.repack7 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  %f.elt8 = extractvalue %main.closeOnce %f, 2
  %f.elt8.elt = extractvalue %runtime._interface %f.elt8, 0
  store i32 %f.elt8.elt, ptr %f1.repack7, align 8
  %f1.repack7.repack9 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  %f.elt8.elt10 = extractvalue %runtime._interface %f.elt8, 1
  store ptr %f.elt8.elt10, ptr %f1.repack7.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #8
  %1 = call { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(*os.File).readdir"(ptr %0, i32 %n, i32 %mode, ptr undef) #8
  %2 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %1, 0
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  %4 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %1, 1
  %5 = extractvalue { ptr, i32, i32 } %4, 0
  call void @runtime.trackPointer(ptr %5, ptr undef) #8
  %6 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %1, 2
  %7 = extractvalue { ptr, i32, i32 } %6, 0
  call void @runtime.trackPointer(ptr %7, ptr undef) #8
  %8 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %1, 3
  %9 = extractvalue %runtime._interface %8, 1
  call void @runtime.trackPointer(ptr %9, ptr undef) #8
  %10 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %1, 0
  %11 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %1, 1
  %12 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %1, 2
  %13 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %1, 3
  %14 = insertvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } zeroinitializer, { ptr, i32, i32 } %10, 0
  %15 = insertvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %14, { ptr, i32, i32 } %11, 1
  %16 = insertvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %15, { ptr, i32, i32 } %12, 2
  %17 = insertvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %16, %runtime._interface %13, 3
  ret { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %17

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(main.closeOnce).readdir$invoke"(ptr %0, i32 %1, i32 %2, ptr %3) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %4 = insertvalue %main.closeOnce undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 1
  %.unpack2 = load %sync.Once, ptr %.elt1, align 4
  %5 = insertvalue %main.closeOnce %4, %sync.Once %.unpack2, 1
  %.elt3 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2
  %.unpack4.unpack = load i32, ptr %.elt3, align 4
  %6 = insertvalue %runtime._interface undef, i32 %.unpack4.unpack, 0
  %.unpack4.elt5 = getelementptr inbounds %main.closeOnce, ptr %0, i32 0, i32 2, i32 1
  %.unpack4.unpack6 = load ptr, ptr %.unpack4.elt5, align 4
  %.unpack47 = insertvalue %runtime._interface %6, ptr %.unpack4.unpack6, 1
  %7 = insertvalue %main.closeOnce %5, %runtime._interface %.unpack47, 2
  %ret = call { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(main.closeOnce).readdir"(%main.closeOnce %7, i32 %1, i32 %2, ptr %3)
  ret { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %ret
}

; Function Attrs: nounwind
define hidden %runtime._interface @"(*main.closeOnce).Close"(ptr dereferenceable_or_null(24) %c, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %c, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  %1 = getelementptr inbounds %main.closeOnce, ptr %c, i32 0, i32 1
  call void @runtime.trackPointer(ptr nonnull %c, ptr undef) #8
  call void @runtime.trackPointer(ptr nonnull @"(*main.closeOnce).close$bound", ptr undef) #8
  call void @"(*sync.Once).Do"(ptr nonnull %1, ptr nonnull %c, ptr nonnull @"(*main.closeOnce).close$bound", ptr undef) #8
  br i1 false, label %gep.throw1, label %gep.next2

gep.next2:                                        ; preds = %gep.next
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next2
  %2 = getelementptr inbounds %main.closeOnce, ptr %c, i32 0, i32 2
  %.unpack = load i32, ptr %2, align 4
  %3 = insertvalue %runtime._interface undef, i32 %.unpack, 0
  %.elt3 = getelementptr inbounds %main.closeOnce, ptr %c, i32 0, i32 2, i32 1
  %.unpack4 = load ptr, ptr %.elt3, align 4
  %4 = insertvalue %runtime._interface %3, ptr %.unpack4, 1
  call void @runtime.trackPointer(ptr %.unpack4, ptr undef) #8
  ret %runtime._interface %4

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

gep.throw1:                                       ; preds = %gep.next
  unreachable

deref.throw:                                      ; preds = %gep.next2
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden i32 @"(*main.closeOnce).Fd"(ptr dereferenceable_or_null(24) %f, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call i32 @"(*os.File).Fd"(ptr %1, ptr undef) #8
  ret i32 %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._string @"(*main.closeOnce).Name"(ptr dereferenceable_or_null(24) %f, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call %runtime._string @"(*os.File).Name"(ptr %1, ptr undef) #8
  %3 = extractvalue %runtime._string %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  ret %runtime._string %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(*main.closeOnce).Read"(ptr dereferenceable_or_null(24) %f, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call { i32, %runtime._interface } @"(*os.File).Read"(ptr %1, ptr %b.data, i32 %b.len, i32 %b.cap, ptr undef) #8
  %3 = extractvalue { i32, %runtime._interface } %2, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  ret { i32, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(*main.closeOnce).ReadAt"(ptr dereferenceable_or_null(24) %f, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %offset, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call { i32, %runtime._interface } @"(*os.File).ReadAt"(ptr %1, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %offset, ptr undef) #8
  %3 = extractvalue { i32, %runtime._interface } %2, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  ret { i32, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, %runtime._interface } @"(*main.closeOnce).ReadDir"(ptr dereferenceable_or_null(24) %f, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).ReadDir"(ptr %1, i32 %n, ptr undef) #8
  %3 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %2, 0
  %4 = extractvalue { ptr, i32, i32 } %3, 0
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  %5 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %2, 1
  %6 = extractvalue %runtime._interface %5, 1
  call void @runtime.trackPointer(ptr %6, ptr undef) #8
  ret { { ptr, i32, i32 }, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, %runtime._interface } @"(*main.closeOnce).Readdir"(ptr dereferenceable_or_null(24) %f, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).Readdir"(ptr %1, i32 %n, ptr undef) #8
  %3 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %2, 0
  %4 = extractvalue { ptr, i32, i32 } %3, 0
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  %5 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %2, 1
  %6 = extractvalue %runtime._interface %5, 1
  call void @runtime.trackPointer(ptr %6, ptr undef) #8
  ret { { ptr, i32, i32 }, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, %runtime._interface } @"(*main.closeOnce).Readdirnames"(ptr dereferenceable_or_null(24) %f, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).Readdirnames"(ptr %1, i32 %n, ptr undef) #8
  %3 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %2, 0
  %4 = extractvalue { ptr, i32, i32 } %3, 0
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  %5 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %2, 1
  %6 = extractvalue %runtime._interface %5, 1
  call void @runtime.trackPointer(ptr %6, ptr undef) #8
  ret { { ptr, i32, i32 }, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { i64, %runtime._interface } @"(*main.closeOnce).Seek"(ptr dereferenceable_or_null(24) %f, i64 %offset, i32 %whence, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call { i64, %runtime._interface } @"(*os.File).Seek"(ptr %1, i64 %offset, i32 %whence, ptr undef) #8
  %3 = extractvalue { i64, %runtime._interface } %2, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  ret { i64, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { %runtime._interface, %runtime._interface } @"(*main.closeOnce).Stat"(ptr dereferenceable_or_null(24) %f, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call { %runtime._interface, %runtime._interface } @"(*os.File).Stat"(ptr %1, ptr undef) #8
  %3 = extractvalue { %runtime._interface, %runtime._interface } %2, 0
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  %5 = extractvalue { %runtime._interface, %runtime._interface } %2, 1
  %6 = extractvalue %runtime._interface %5, 1
  call void @runtime.trackPointer(ptr %6, ptr undef) #8
  ret { %runtime._interface, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(*main.closeOnce).Sync"(ptr dereferenceable_or_null(24) %f, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call %runtime._interface @"(*os.File).Sync"(ptr %1, ptr undef) #8
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  ret %runtime._interface %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { %runtime._interface, %runtime._interface } @"(*main.closeOnce).SyscallConn"(ptr dereferenceable_or_null(24) %f, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call { %runtime._interface, %runtime._interface } @"(*os.File).SyscallConn"(ptr %1, ptr undef) #8
  %3 = extractvalue { %runtime._interface, %runtime._interface } %2, 0
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  %5 = extractvalue { %runtime._interface, %runtime._interface } %2, 1
  %6 = extractvalue %runtime._interface %5, 1
  call void @runtime.trackPointer(ptr %6, ptr undef) #8
  ret { %runtime._interface, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(*main.closeOnce).Truncate"(ptr dereferenceable_or_null(24) %f, i64 %size, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call %runtime._interface @"(*os.File).Truncate"(ptr %1, i64 %size, ptr undef) #8
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  ret %runtime._interface %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(*main.closeOnce).Write"(ptr dereferenceable_or_null(24) %f, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call { i32, %runtime._interface } @"(*os.File).Write"(ptr %1, ptr %b.data, i32 %b.len, i32 %b.cap, ptr undef) #8
  %3 = extractvalue { i32, %runtime._interface } %2, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  ret { i32, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(*main.closeOnce).WriteAt"(ptr dereferenceable_or_null(24) %f, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %off, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call { i32, %runtime._interface } @"(*os.File).WriteAt"(ptr %1, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %off, ptr undef) #8
  %3 = extractvalue { i32, %runtime._interface } %2, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  ret { i32, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(*main.closeOnce).WriteString"(ptr dereferenceable_or_null(24) %f, ptr %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call { i32, %runtime._interface } @"(*os.File).WriteString"(ptr %1, ptr %s.data, i32 %s.len, ptr undef) #8
  %3 = extractvalue { i32, %runtime._interface } %2, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  ret { i32, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define hidden void @"(*main.closeOnce).close"(ptr dereferenceable_or_null(24) %c, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %c, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %gep.throw1, label %gep.next2

gep.next2:                                        ; preds = %gep.next
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next2
  %1 = load ptr, ptr %c, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call %runtime._interface @"(*os.File).Close"(ptr %1, ptr undef) #8
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #8
  br i1 false, label %store.throw, label %store.next

store.next:                                       ; preds = %deref.next
  %4 = getelementptr inbounds %main.closeOnce, ptr %c, i32 0, i32 2
  %.elt = extractvalue %runtime._interface %2, 0
  store i32 %.elt, ptr %4, align 4
  %.repack3 = getelementptr inbounds %main.closeOnce, ptr %c, i32 0, i32 2, i32 1
  %.elt4 = extractvalue %runtime._interface %2, 1
  store ptr %.elt4, ptr %.repack3, align 4
  ret void

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

gep.throw1:                                       ; preds = %gep.next
  unreachable

deref.throw:                                      ; preds = %gep.next2
  unreachable

store.throw:                                      ; preds = %deref.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(*main.closeOnce).readdir"(ptr dereferenceable_or_null(24) %f, i32 %n, i32 %mode, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #8
  %2 = call { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(*os.File).readdir"(ptr %1, i32 %n, i32 %mode, ptr undef) #8
  %3 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %2, 0
  %4 = extractvalue { ptr, i32, i32 } %3, 0
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  %5 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %2, 1
  %6 = extractvalue { ptr, i32, i32 } %5, 0
  call void @runtime.trackPointer(ptr %6, ptr undef) #8
  %7 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %2, 2
  %8 = extractvalue { ptr, i32, i32 } %7, 0
  call void @runtime.trackPointer(ptr %8, ptr undef) #8
  %9 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %2, 3
  %10 = extractvalue %runtime._interface %9, 1
  call void @runtime.trackPointer(ptr %10, ptr undef) #8
  %11 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %2, 0
  %12 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %2, 1
  %13 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %2, 2
  %14 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %2, 3
  %15 = insertvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } zeroinitializer, { ptr, i32, i32 } %11, 0
  %16 = insertvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %15, { ptr, i32, i32 } %12, 1
  %17 = insertvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %16, { ptr, i32, i32 } %13, 2
  %18 = insertvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %17, %runtime._interface %14, 3
  ret { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %18

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #8
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden void @"(*main.closeOnce).close$bound"(ptr %context) unnamed_addr #1 {
entry:
  call void @"(*main.closeOnce).close"(ptr %context, ptr undef)
  ret void
}

; Function Attrs: nounwind
define hidden void @main.main(ptr %context) unnamed_addr #1 {
entry:
  %new = call ptr @runtime.alloc(i32 32, ptr nonnull inttoptr (i32 2449 to ptr), ptr undef) #8
  call void @runtime.trackPointer(ptr nonnull %new, ptr undef) #8
  %0 = call { %runtime._interface, %runtime._interface } @"(*main.Cmd).StdinPipe"(ptr nonnull %new, ptr undef)
  %1 = extractvalue { %runtime._interface, %runtime._interface } %0, 0
  %2 = extractvalue %runtime._interface %1, 1
  call void @runtime.trackPointer(ptr %2, ptr undef) #8
  %3 = extractvalue { %runtime._interface, %runtime._interface } %0, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #8
  %5 = extractvalue { %runtime._interface, %runtime._interface } %0, 0
  %6 = extractvalue { %runtime._interface, %runtime._interface } %0, 1
  %7 = extractvalue %runtime._interface %6, 0
  %.not = icmp eq i32 %7, 0
  br i1 %.not, label %if.done, label %if.then

if.then:                                          ; preds = %entry
  call void @os.Exit(i32 1, ptr undef) #8
  br label %if.done

if.done:                                          ; preds = %if.then, %entry
  %invoke.func.typecode = extractvalue %runtime._interface %5, 0
  %invoke.func.value = extractvalue %runtime._interface %5, 1
  %8 = call %runtime._interface @"interface:{Close:func:{}{named:error},Write:func:{slice:basic:uint8}{basic:int,named:error}}.Close$invoke"(ptr %invoke.func.value, i32 %invoke.func.typecode, ptr undef) #8
  %9 = extractvalue %runtime._interface %8, 1
  call void @runtime.trackPointer(ptr %9, ptr undef) #8
  %10 = extractvalue %runtime._interface %8, 0
  %.not3 = icmp eq i32 %10, 0
  br i1 %.not3, label %if.done2, label %if.then1

if.then1:                                         ; preds = %if.done
  call void @os.Exit(i32 1, ptr undef) #8
  br label %if.done2

if.done2:                                         ; preds = %if.then1, %if.done
  ret void
}

declare void @os.Exit(i32, ptr) #0

declare %runtime._interface @"interface:{Close:func:{}{named:error},Write:func:{slice:basic:uint8}{basic:int,named:error}}.Close$invoke"(ptr, i32, ptr) #7

attributes #0 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #2 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "tinygo-methods"="reflect/methods.Close() error; reflect/methods.Read([]uint8) (int, error); reflect/methods.ReadAt([]uint8, int64) (int, error); reflect/methods.Seek(int64, int) (int64, error); reflect/methods.Write([]uint8) (int, error)" }
attributes #3 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "wasm-import-module"="env" "wasm-import-name"="tinygo_unwind" }
attributes #4 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "wasm-import-module"="env" "wasm-import-name"="tinygo_launch" }
attributes #5 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "wasm-import-module"="env" "wasm-import-name"="tinygo_rewind" }
attributes #6 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "tinygo-methods"="reflect/methods.Error() string" }
attributes #7 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "tinygo-invoke"="reflect/methods.Close() error" "tinygo-methods"="reflect/methods.Close() error; reflect/methods.Write([]uint8) (int, error)" }
attributes #8 = { nounwind }
