; ModuleID = 'exec.go'
source_filename = "exec.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime._interface = type { i32, ptr }
%runtime.typecodeID = type { ptr, i32, ptr, ptr, i32 }
%runtime.structField = type { ptr, ptr, ptr, i1 }
%runtime.interfaceMethodInfo = type { ptr, i32 }
%runtime._string = type { ptr, i32 }
%main.Error = type { %runtime._string, %runtime._interface }
%main.wrappedError = type { %runtime._string, %runtime._interface }
%main.Cmd = type { %runtime._string, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._string, %runtime._interface, %runtime._interface, %runtime._interface, { ptr, i32, i32 }, ptr, ptr, ptr, %runtime._interface, %runtime._interface, { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, ptr, ptr, %runtime._interface }
%main.ExitError = type { ptr, { ptr, i32, i32 } }
%main.prefixSuffixSaver = type { i32, { ptr, i32, i32 }, { ptr, i32, i32 }, i32, i64 }
%runtime.chanSelectState = type { ptr, ptr }
%os.ProcAttr = type { %runtime._string, { ptr, i32, i32 }, { ptr, i32, i32 }, ptr }
%runtime.channelBlockedList = type { ptr, ptr, ptr, { ptr, i32, i32 } }
%"internal/task.stackState" = type { i32, i32 }
%"internal/task.state" = type { i32, ptr, %"internal/task.stackState", i1 }
%"internal/task.Queue" = type { ptr, ptr }
%sync.Once = type { i1, %sync.Mutex }
%sync.Mutex = type { i1, %"internal/task.Stack" }
%"internal/task.Stack" = type { ptr }
%main.closeOnce = type { ptr, %sync.Once, %runtime._interface }
%runtime._defer = type { i32, ptr }

@"main$string" = internal unnamed_addr constant [57 x i8] c"cannot run executable found relative to current directory", align 1
@main.ErrDot = hidden global %runtime._interface zeroinitializer, align 8
@"main$string.1" = internal unnamed_addr constant [6 x i8] c"exec: ", align 1
@"main$string.2" = internal unnamed_addr constant [2 x i8] c": ", align 1
@"main$string.3" = internal unnamed_addr constant [2 x i8] c": ", align 1
@"main$string.4" = internal unnamed_addr constant [24 x i8] c"exec: Stdout already set", align 1
@"main$string.5" = internal unnamed_addr constant [24 x i8] c"exec: Stderr already set", align 1
@"reflect/types.type:pointer:named:bytes.Buffer" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:bytes.Buffer", i32 0, ptr @"*bytes.Buffer$methodset", ptr null, i32 0 }
@"reflect/types.type:named:bytes.Buffer" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{buf:slice:basic:uint8,off:basic:int,lastRead:named:bytes.readOp}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:bytes.Buffer", i32 0 }
@"reflect/types.type:struct:{buf:slice:basic:uint8,off:basic:int,lastRead:named:bytes.readOp}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields", i32 0, ptr null, ptr @"reflect/types.type:pointer:struct:{buf:slice:basic:uint8,off:basic:int,lastRead:named:bytes.readOp}", i32 0 }
@"reflect/types.structFields" = private unnamed_addr global [3 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:slice:basic:uint8", ptr @"reflect/types.structFieldName", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:basic:int", ptr @"reflect/types.structFieldName.6", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:named:bytes.readOp", ptr @"reflect/types.structFieldName.7", ptr null, i1 false }]
@"reflect/types.type:slice:basic:uint8" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:basic:uint8", i32 0, ptr null, ptr @"reflect/types.type:pointer:slice:basic:uint8", i32 0 }
@"reflect/types.type:basic:uint8" = linkonce_odr constant %runtime.typecodeID { ptr null, i32 0, ptr null, ptr @"reflect/types.type:pointer:basic:uint8", i32 0 }
@"reflect/types.type:pointer:basic:uint8" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:basic:uint8", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.type:pointer:slice:basic:uint8" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:slice:basic:uint8", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.structFieldName" = private unnamed_addr global [3 x i8] c"buf"
@"reflect/types.type:basic:int" = linkonce_odr constant %runtime.typecodeID { ptr null, i32 0, ptr null, ptr @"reflect/types.type:pointer:basic:int", i32 0 }
@"reflect/types.type:pointer:basic:int" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:basic:int", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.structFieldName.6" = private unnamed_addr global [3 x i8] c"off"
@"reflect/types.type:named:bytes.readOp" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:basic:int8", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:bytes.readOp", i32 0 }
@"reflect/types.type:basic:int8" = linkonce_odr constant %runtime.typecodeID { ptr null, i32 0, ptr null, ptr @"reflect/types.type:pointer:basic:int8", i32 0 }
@"reflect/types.type:pointer:basic:int8" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:basic:int8", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.type:pointer:named:bytes.readOp" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:bytes.readOp", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.structFieldName.7" = private unnamed_addr global [8 x i8] c"lastRead"
@"reflect/types.type:pointer:struct:{buf:slice:basic:uint8,off:basic:int,lastRead:named:bytes.readOp}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{buf:slice:basic:uint8,off:basic:int,lastRead:named:bytes.readOp}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/methods.Bytes() []uint8" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Cap() int" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Grow(int)" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Len() int" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Next(int) []uint8" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Read([]uint8) (int, error)" = linkonce_odr constant i8 0, align 1
@"reflect/methods.ReadByte() (uint8, error)" = linkonce_odr constant i8 0, align 1
@"reflect/methods.ReadBytes(uint8) ([]uint8, error)" = linkonce_odr constant i8 0, align 1
@"reflect/methods.ReadFrom(io.Reader) (int64, error)" = linkonce_odr constant i8 0, align 1
@"reflect/methods.ReadRune() (int32, int, error)" = linkonce_odr constant i8 0, align 1
@"reflect/methods.ReadString(uint8) (string, error)" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Reset()" = linkonce_odr constant i8 0, align 1
@"reflect/methods.String() string" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Truncate(int)" = linkonce_odr constant i8 0, align 1
@"reflect/methods.UnreadByte() error" = linkonce_odr constant i8 0, align 1
@"reflect/methods.UnreadRune() error" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Write([]uint8) (int, error)" = linkonce_odr constant i8 0, align 1
@"reflect/methods.WriteByte(uint8) error" = linkonce_odr constant i8 0, align 1
@"reflect/methods.WriteRune(int32) (int, error)" = linkonce_odr constant i8 0, align 1
@"reflect/methods.WriteString(string) (int, error)" = linkonce_odr constant i8 0, align 1
@"reflect/methods.WriteTo(io.Writer) (int64, error)" = linkonce_odr constant i8 0, align 1
@"bytes.$methods.empty() bool" = linkonce_odr constant i8 0, align 1
@"bytes.$methods.grow(int) int" = linkonce_odr constant i8 0, align 1
@"bytes.$methods.readSlice(uint8) ([]uint8, error)" = linkonce_odr constant i8 0, align 1
@"bytes.$methods.tryGrowByReslice(int) (int, bool)" = linkonce_odr constant i8 0, align 1
@"*bytes.Buffer$methodset" = linkonce_odr constant [25 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.Bytes() []uint8", i32 ptrtoint (ptr @"(*bytes.Buffer).Bytes" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Cap() int", i32 ptrtoint (ptr @"(*bytes.Buffer).Cap" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Grow(int)", i32 ptrtoint (ptr @"(*bytes.Buffer).Grow" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Len() int", i32 ptrtoint (ptr @"(*bytes.Buffer).Len" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Next(int) []uint8", i32 ptrtoint (ptr @"(*bytes.Buffer).Next" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Read([]uint8) (int, error)", i32 ptrtoint (ptr @"(*bytes.Buffer).Read" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadByte() (uint8, error)", i32 ptrtoint (ptr @"(*bytes.Buffer).ReadByte" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadBytes(uint8) ([]uint8, error)", i32 ptrtoint (ptr @"(*bytes.Buffer).ReadBytes" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadFrom(io.Reader) (int64, error)", i32 ptrtoint (ptr @"(*bytes.Buffer).ReadFrom" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadRune() (int32, int, error)", i32 ptrtoint (ptr @"(*bytes.Buffer).ReadRune" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadString(uint8) (string, error)", i32 ptrtoint (ptr @"(*bytes.Buffer).ReadString" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Reset()", i32 ptrtoint (ptr @"(*bytes.Buffer).Reset" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.String() string", i32 ptrtoint (ptr @"(*bytes.Buffer).String" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Truncate(int)", i32 ptrtoint (ptr @"(*bytes.Buffer).Truncate" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.UnreadByte() error", i32 ptrtoint (ptr @"(*bytes.Buffer).UnreadByte" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.UnreadRune() error", i32 ptrtoint (ptr @"(*bytes.Buffer).UnreadRune" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Write([]uint8) (int, error)", i32 ptrtoint (ptr @"(*bytes.Buffer).Write" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.WriteByte(uint8) error", i32 ptrtoint (ptr @"(*bytes.Buffer).WriteByte" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.WriteRune(int32) (int, error)", i32 ptrtoint (ptr @"(*bytes.Buffer).WriteRune" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.WriteString(string) (int, error)", i32 ptrtoint (ptr @"(*bytes.Buffer).WriteString" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.WriteTo(io.Writer) (int64, error)", i32 ptrtoint (ptr @"(*bytes.Buffer).WriteTo" to i32) }, %runtime.interfaceMethodInfo { ptr @"bytes.$methods.empty() bool", i32 ptrtoint (ptr @"(*bytes.Buffer).empty" to i32) }, %runtime.interfaceMethodInfo { ptr @"bytes.$methods.grow(int) int", i32 ptrtoint (ptr @"(*bytes.Buffer).grow" to i32) }, %runtime.interfaceMethodInfo { ptr @"bytes.$methods.readSlice(uint8) ([]uint8, error)", i32 ptrtoint (ptr @"(*bytes.Buffer).readSlice" to i32) }, %runtime.interfaceMethodInfo { ptr @"bytes.$methods.tryGrowByReslice(int) (int, bool)", i32 ptrtoint (ptr @"(*bytes.Buffer).tryGrowByReslice" to i32) }]
@"main$string.8" = internal unnamed_addr constant [24 x i8] c"exec: Stdout already set", align 1
@"reflect/types.type:pointer:named:main.prefixSuffixSaver" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:main.prefixSuffixSaver", i32 0, ptr @"*main.prefixSuffixSaver$methodset", ptr null, i32 0 }
@"reflect/types.type:named:main.prefixSuffixSaver" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{N:basic:int,prefix:slice:basic:uint8,suffix:slice:basic:uint8,suffixOff:basic:int,skipped:basic:int64}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:main.prefixSuffixSaver", i32 0 }
@"reflect/types.type:struct:{N:basic:int,prefix:slice:basic:uint8,suffix:slice:basic:uint8,suffixOff:basic:int,skipped:basic:int64}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.9", i32 0, ptr null, ptr @"reflect/types.type:pointer:struct:{N:basic:int,prefix:slice:basic:uint8,suffix:slice:basic:uint8,suffixOff:basic:int,skipped:basic:int64}", i32 0 }
@"reflect/types.structFields.9" = private unnamed_addr global [5 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:basic:int", ptr @"reflect/types.structFieldName.10", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:slice:basic:uint8", ptr @"reflect/types.structFieldName.11", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:slice:basic:uint8", ptr @"reflect/types.structFieldName.12", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:basic:int", ptr @"reflect/types.structFieldName.13", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:basic:int64", ptr @"reflect/types.structFieldName.14", ptr null, i1 false }]
@"reflect/types.structFieldName.10" = private unnamed_addr global [1 x i8] c"N"
@"reflect/types.structFieldName.11" = private unnamed_addr global [6 x i8] c"prefix"
@"reflect/types.structFieldName.12" = private unnamed_addr global [6 x i8] c"suffix"
@"reflect/types.structFieldName.13" = private unnamed_addr global [9 x i8] c"suffixOff"
@"reflect/types.type:basic:int64" = linkonce_odr constant %runtime.typecodeID { ptr null, i32 0, ptr null, ptr @"reflect/types.type:pointer:basic:int64", i32 0 }
@"reflect/types.type:pointer:basic:int64" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:basic:int64", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.structFieldName.14" = private unnamed_addr global [7 x i8] c"skipped"
@"reflect/types.type:pointer:struct:{N:basic:int,prefix:slice:basic:uint8,suffix:slice:basic:uint8,suffixOff:basic:int,skipped:basic:int64}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{N:basic:int,prefix:slice:basic:uint8,suffix:slice:basic:uint8,suffixOff:basic:int,skipped:basic:int64}", i32 0, ptr null, ptr null, i32 0 }
@"main.$methods.fill(*[]uint8, []uint8) []uint8" = linkonce_odr constant i8 0, align 1
@"*main.prefixSuffixSaver$methodset" = linkonce_odr constant [3 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.Bytes() []uint8", i32 ptrtoint (ptr @"(*main.prefixSuffixSaver).Bytes" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Write([]uint8) (int, error)", i32 ptrtoint (ptr @"(*main.prefixSuffixSaver).Write" to i32) }, %runtime.interfaceMethodInfo { ptr @"main.$methods.fill(*[]uint8, []uint8) []uint8", i32 ptrtoint (ptr @"(*main.prefixSuffixSaver).fill" to i32) }]
@"reflect/types.typeid:pointer:named:main.ExitError" = external constant i8
@"reflect/types.typeid:pointer:named:main.prefixSuffixSaver" = external constant i8
@"main$string.15" = internal unnamed_addr constant [16 x i8] c"exec: no command", align 1
@"main$string.16" = internal unnamed_addr constant [21 x i8] c"exec: already started", align 1
@"main$string.17" = internal unnamed_addr constant [24 x i8] c"exec: Stderr already set", align 1
@"main$string.18" = internal unnamed_addr constant [38 x i8] c"exec: StderrPipe after process started", align 1
@"reflect/types.type:pointer:named:os.File" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:os.File", i32 0, ptr @"*os.File$methodset", ptr null, i32 0 }
@"reflect/types.type:named:os.File" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{#file:pointer:named:os.file}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:os.File", i32 0 }
@"reflect/types.type:struct:{#file:pointer:named:os.file}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.19", i32 0, ptr null, ptr @"reflect/types.type:pointer:struct:{#file:pointer:named:os.file}", i32 0 }
@"reflect/types.structFields.19" = private unnamed_addr global [1 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:pointer:named:os.file", ptr @"reflect/types.structFieldName.23", ptr null, i1 true }]
@"reflect/types.type:pointer:named:os.file" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:os.file", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.type:named:os.file" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{handle:named:os.FileHandle,name:basic:string}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:os.file", i32 0 }
@"reflect/types.type:struct:{handle:named:os.FileHandle,name:basic:string}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.20", i32 0, ptr null, ptr @"reflect/types.type:pointer:struct:{handle:named:os.FileHandle,name:basic:string}", i32 0 }
@"reflect/types.structFields.20" = private unnamed_addr global [2 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:named:os.FileHandle", ptr @"reflect/types.structFieldName.21", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:basic:string", ptr @"reflect/types.structFieldName.22", ptr null, i1 false }]
@"reflect/types.type:named:os.FileHandle" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:interface:{Close:func:{}{named:error},Read:func:{slice:basic:uint8}{basic:int,named:error},ReadAt:func:{slice:basic:uint8,basic:int64}{basic:int,named:error},Seek:func:{basic:int64,basic:int}{basic:int64,named:error},Write:func:{slice:basic:uint8}{basic:int,named:error}}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:os.FileHandle", i32 ptrtoint (ptr @"interface:{Close:func:{}{named:error},Read:func:{slice:basic:uint8}{basic:int,named:error},ReadAt:func:{slice:basic:uint8,basic:int64}{basic:int,named:error},Seek:func:{basic:int64,basic:int}{basic:int64,named:error},Write:func:{slice:basic:uint8}{basic:int,named:error}}.$typeassert" to i32) }
@"reflect/types.type:interface:{Close:func:{}{named:error},Read:func:{slice:basic:uint8}{basic:int,named:error},ReadAt:func:{slice:basic:uint8,basic:int64}{basic:int,named:error},Seek:func:{basic:int64,basic:int}{basic:int64,named:error},Write:func:{slice:basic:uint8}{basic:int,named:error}}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.interface:interface{Close() (err error); Read(b []byte) (n int, err error); ReadAt(b []byte, offset int64) (n int, err error); Seek(offset int64, whence int) (newoffset int64, err error); Write(b []byte) (n int, err error)}$interface", i32 0, ptr null, ptr @"reflect/types.type:pointer:interface:{Close:func:{}{named:error},Read:func:{slice:basic:uint8}{basic:int,named:error},ReadAt:func:{slice:basic:uint8,basic:int64}{basic:int,named:error},Seek:func:{basic:int64,basic:int}{basic:int64,named:error},Write:func:{slice:basic:uint8}{basic:int,named:error}}", i32 ptrtoint (ptr @"interface:{Close:func:{}{named:error},Read:func:{slice:basic:uint8}{basic:int,named:error},ReadAt:func:{slice:basic:uint8,basic:int64}{basic:int,named:error},Seek:func:{basic:int64,basic:int}{basic:int64,named:error},Write:func:{slice:basic:uint8}{basic:int,named:error}}.$typeassert" to i32) }
@"reflect/methods.Close() error" = linkonce_odr constant i8 0, align 1
@"reflect/methods.ReadAt([]uint8, int64) (int, error)" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Seek(int64, int) (int64, error)" = linkonce_odr constant i8 0, align 1
@"reflect/types.interface:interface{Close() (err error); Read(b []byte) (n int, err error); ReadAt(b []byte, offset int64) (n int, err error); Seek(offset int64, whence int) (newoffset int64, err error); Write(b []byte) (n int, err error)}$interface" = linkonce_odr constant [5 x ptr] [ptr @"reflect/methods.Close() error", ptr @"reflect/methods.Read([]uint8) (int, error)", ptr @"reflect/methods.ReadAt([]uint8, int64) (int, error)", ptr @"reflect/methods.Seek(int64, int) (int64, error)", ptr @"reflect/methods.Write([]uint8) (int, error)"]
@"reflect/types.type:pointer:interface:{Close:func:{}{named:error},Read:func:{slice:basic:uint8}{basic:int,named:error},ReadAt:func:{slice:basic:uint8,basic:int64}{basic:int,named:error},Seek:func:{basic:int64,basic:int}{basic:int64,named:error},Write:func:{slice:basic:uint8}{basic:int,named:error}}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:interface:{Close:func:{}{named:error},Read:func:{slice:basic:uint8}{basic:int,named:error},ReadAt:func:{slice:basic:uint8,basic:int64}{basic:int,named:error},Seek:func:{basic:int64,basic:int}{basic:int64,named:error},Write:func:{slice:basic:uint8}{basic:int,named:error}}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.type:pointer:named:os.FileHandle" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:os.FileHandle", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.structFieldName.21" = private unnamed_addr global [6 x i8] c"handle"
@"reflect/types.type:basic:string" = linkonce_odr constant %runtime.typecodeID { ptr null, i32 0, ptr null, ptr @"reflect/types.type:pointer:basic:string", i32 0 }
@"reflect/types.type:pointer:basic:string" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:basic:string", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.structFieldName.22" = private unnamed_addr global [4 x i8] c"name"
@"reflect/types.type:pointer:struct:{handle:named:os.FileHandle,name:basic:string}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{handle:named:os.FileHandle,name:basic:string}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.structFieldName.23" = private unnamed_addr global [4 x i8] c"file"
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
@"os.$methods.readdir(int, os.readdirMode) ([]string, []io/fs.DirEntry, []io/fs.FileInfo, error)" = linkonce_odr constant i8 0, align 1
@"*os.File$methodset" = linkonce_odr constant [17 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.Close() error", i32 ptrtoint (ptr @"(*os.File).Close" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Fd() uintptr", i32 ptrtoint (ptr @"(*os.File).Fd" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Name() string", i32 ptrtoint (ptr @"(*os.File).Name" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Read([]uint8) (int, error)", i32 ptrtoint (ptr @"(*os.File).Read" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadAt([]uint8, int64) (int, error)", i32 ptrtoint (ptr @"(*os.File).ReadAt" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadDir(int) ([]io/fs.DirEntry, error)", i32 ptrtoint (ptr @"(*os.File).ReadDir" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Readdir(int) ([]io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(*os.File).Readdir" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Readdirnames(int) ([]string, error)", i32 ptrtoint (ptr @"(*os.File).Readdirnames" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Seek(int64, int) (int64, error)", i32 ptrtoint (ptr @"(*os.File).Seek" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Stat() (io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(*os.File).Stat" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Sync() error", i32 ptrtoint (ptr @"(*os.File).Sync" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.SyscallConn() (syscall.RawConn, error)", i32 ptrtoint (ptr @"(*os.File).SyscallConn" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Truncate(int64) error", i32 ptrtoint (ptr @"(*os.File).Truncate" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Write([]uint8) (int, error)", i32 ptrtoint (ptr @"(*os.File).Write" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.WriteAt([]uint8, int64) (int, error)", i32 ptrtoint (ptr @"(*os.File).WriteAt" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.WriteString(string) (int, error)", i32 ptrtoint (ptr @"(*os.File).WriteString" to i32) }, %runtime.interfaceMethodInfo { ptr @"os.$methods.readdir(int, os.readdirMode) ([]string, []io/fs.DirEntry, []io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(*os.File).readdir" to i32) }]
@"main$string.24" = internal unnamed_addr constant [23 x i8] c"exec: Stdin already set", align 1
@"main$string.25" = internal unnamed_addr constant [37 x i8] c"exec: StdinPipe after process started", align 1
@"reflect/types.type:pointer:named:main.closeOnce" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:main.closeOnce", i32 0, ptr @"*main.closeOnce$methodset", ptr null, i32 0 }
@"reflect/types.type:named:main.closeOnce" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{#File:pointer:named:os.File,once:named:sync.Once,err:named:error}", i32 0, ptr @"main.closeOnce$methodset", ptr @"reflect/types.type:pointer:named:main.closeOnce", i32 0 }
@"reflect/types.type:struct:{#File:pointer:named:os.File,once:named:sync.Once,err:named:error}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.26", i32 0, ptr @"struct{*os.File; once sync.Once; err error}$methodset", ptr @"reflect/types.type:pointer:struct:{#File:pointer:named:os.File,once:named:sync.Once,err:named:error}", i32 0 }
@"reflect/types.structFields.26" = private unnamed_addr global [3 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:pointer:named:os.File", ptr @"reflect/types.structFieldName.27", ptr null, i1 true }, %runtime.structField { ptr @"reflect/types.type:named:sync.Once", ptr @"reflect/types.structFieldName.53", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:named:error", ptr @"reflect/types.structFieldName.54", ptr null, i1 false }]
@"reflect/types.structFieldName.27" = private unnamed_addr global [4 x i8] c"File"
@"reflect/types.type:named:sync.Once" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{done:basic:bool,m:named:sync.Mutex}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:sync.Once", i32 0 }
@"reflect/types.type:struct:{done:basic:bool,m:named:sync.Mutex}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.28", i32 0, ptr null, ptr @"reflect/types.type:pointer:struct:{done:basic:bool,m:named:sync.Mutex}", i32 0 }
@"reflect/types.structFields.28" = private unnamed_addr global [2 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:basic:bool", ptr @"reflect/types.structFieldName.29", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:named:sync.Mutex", ptr @"reflect/types.structFieldName.52", ptr null, i1 false }]
@"reflect/types.type:basic:bool" = linkonce_odr constant %runtime.typecodeID { ptr null, i32 0, ptr null, ptr @"reflect/types.type:pointer:basic:bool", i32 0 }
@"reflect/types.type:pointer:basic:bool" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:basic:bool", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.structFieldName.29" = private unnamed_addr global [4 x i8] c"done"
@"reflect/types.type:named:sync.Mutex" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{locked:basic:bool,blocked:named:internal/task.Stack}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:sync.Mutex", i32 0 }
@"reflect/types.type:struct:{locked:basic:bool,blocked:named:internal/task.Stack}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.30", i32 0, ptr null, ptr @"reflect/types.type:pointer:struct:{locked:basic:bool,blocked:named:internal/task.Stack}", i32 0 }
@"reflect/types.structFields.30" = private unnamed_addr global [2 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:basic:bool", ptr @"reflect/types.structFieldName.31", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:named:internal/task.Stack", ptr @"reflect/types.structFieldName.51", ptr null, i1 false }]
@"reflect/types.structFieldName.31" = private unnamed_addr global [6 x i8] c"locked"
@"reflect/types.type:named:internal/task.Stack" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{top:pointer:named:internal/task.Task}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:internal/task.Stack", i32 0 }
@"reflect/types.type:struct:{top:pointer:named:internal/task.Task}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.32", i32 0, ptr null, ptr @"reflect/types.type:pointer:struct:{top:pointer:named:internal/task.Task}", i32 0 }
@"reflect/types.structFields.32" = private unnamed_addr global [1 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:pointer:named:internal/task.Task", ptr @"reflect/types.structFieldName.50", ptr null, i1 false }]
@"reflect/types.type:pointer:named:internal/task.Task" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:internal/task.Task", i32 0, ptr @"*internal/task.Task$methodset", ptr null, i32 0 }
@"reflect/types.type:named:internal/task.Task" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{Next:pointer:named:internal/task.Task,Ptr:basic:unsafe.Pointer,Data:basic:uint64,gcData:named:internal/task.gcData,state:named:internal/task.state,DeferFrame:basic:unsafe.Pointer}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:internal/task.Task", i32 0 }
@"reflect/types.type:struct:{Next:pointer:named:internal/task.Task,Ptr:basic:unsafe.Pointer,Data:basic:uint64,gcData:named:internal/task.gcData,state:named:internal/task.state,DeferFrame:basic:unsafe.Pointer}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.33", i32 0, ptr null, ptr @"reflect/types.type:pointer:struct:{Next:pointer:named:internal/task.Task,Ptr:basic:unsafe.Pointer,Data:basic:uint64,gcData:named:internal/task.gcData,state:named:internal/task.state,DeferFrame:basic:unsafe.Pointer}", i32 0 }
@"reflect/types.structFields.33" = private unnamed_addr global [6 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:pointer:named:internal/task.Task", ptr @"reflect/types.structFieldName.34", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:basic:unsafe.Pointer", ptr @"reflect/types.structFieldName.35", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:basic:uint64", ptr @"reflect/types.structFieldName.36", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:named:internal/task.gcData", ptr @"reflect/types.structFieldName.39", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:named:internal/task.state", ptr @"reflect/types.structFieldName.48", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:basic:unsafe.Pointer", ptr @"reflect/types.structFieldName.49", ptr null, i1 false }]
@"reflect/types.structFieldName.34" = private unnamed_addr global [4 x i8] c"Next"
@"reflect/types.type:basic:unsafe.Pointer" = linkonce_odr constant %runtime.typecodeID { ptr null, i32 0, ptr null, ptr @"reflect/types.type:pointer:basic:unsafe.Pointer", i32 0 }
@"reflect/types.type:pointer:basic:unsafe.Pointer" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:basic:unsafe.Pointer", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.structFieldName.35" = private unnamed_addr global [3 x i8] c"Ptr"
@"reflect/types.type:basic:uint64" = linkonce_odr constant %runtime.typecodeID { ptr null, i32 0, ptr null, ptr @"reflect/types.type:pointer:basic:uint64", i32 0 }
@"reflect/types.type:pointer:basic:uint64" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:basic:uint64", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.structFieldName.36" = private unnamed_addr global [4 x i8] c"Data"
@"reflect/types.type:named:internal/task.gcData" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{stackChain:basic:unsafe.Pointer}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:internal/task.gcData", i32 0 }
@"reflect/types.type:struct:{stackChain:basic:unsafe.Pointer}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.37", i32 0, ptr null, ptr @"reflect/types.type:pointer:struct:{stackChain:basic:unsafe.Pointer}", i32 0 }
@"reflect/types.structFields.37" = private unnamed_addr global [1 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:basic:unsafe.Pointer", ptr @"reflect/types.structFieldName.38", ptr null, i1 false }]
@"reflect/types.structFieldName.38" = private unnamed_addr global [10 x i8] c"stackChain"
@"reflect/types.type:pointer:struct:{stackChain:basic:unsafe.Pointer}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{stackChain:basic:unsafe.Pointer}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.type:pointer:named:internal/task.gcData" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:internal/task.gcData", i32 0, ptr @"*internal/task.gcData$methodset", ptr null, i32 0 }
@"internal/task.$methods.swap()" = linkonce_odr constant i8 0, align 1
@"*internal/task.gcData$methodset" = linkonce_odr constant [1 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"internal/task.$methods.swap()", i32 ptrtoint (ptr @"(*internal/task.gcData).swap" to i32) }]
@"reflect/types.structFieldName.39" = private unnamed_addr global [6 x i8] c"gcData"
@"reflect/types.type:named:internal/task.state" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{entry:basic:uintptr,args:basic:unsafe.Pointer,#stackState:named:internal/task.stackState,launched:basic:bool}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:internal/task.state", i32 0 }
@"reflect/types.type:struct:{entry:basic:uintptr,args:basic:unsafe.Pointer,#stackState:named:internal/task.stackState,launched:basic:bool}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.40", i32 0, ptr null, ptr @"reflect/types.type:pointer:struct:{entry:basic:uintptr,args:basic:unsafe.Pointer,#stackState:named:internal/task.stackState,launched:basic:bool}", i32 0 }
@"reflect/types.structFields.40" = private unnamed_addr global [4 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:basic:uintptr", ptr @"reflect/types.structFieldName.41", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:basic:unsafe.Pointer", ptr @"reflect/types.structFieldName.42", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:named:internal/task.stackState", ptr @"reflect/types.structFieldName.46", ptr null, i1 true }, %runtime.structField { ptr @"reflect/types.type:basic:bool", ptr @"reflect/types.structFieldName.47", ptr null, i1 false }]
@"reflect/types.type:basic:uintptr" = linkonce_odr constant %runtime.typecodeID { ptr null, i32 0, ptr null, ptr @"reflect/types.type:pointer:basic:uintptr", i32 0 }
@"reflect/types.type:pointer:basic:uintptr" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:basic:uintptr", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.structFieldName.41" = private unnamed_addr global [5 x i8] c"entry"
@"reflect/types.structFieldName.42" = private unnamed_addr global [4 x i8] c"args"
@"reflect/types.type:named:internal/task.stackState" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{asyncifysp:basic:uintptr,csp:basic:uintptr}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:internal/task.stackState", i32 0 }
@"reflect/types.type:struct:{asyncifysp:basic:uintptr,csp:basic:uintptr}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.43", i32 0, ptr null, ptr @"reflect/types.type:pointer:struct:{asyncifysp:basic:uintptr,csp:basic:uintptr}", i32 0 }
@"reflect/types.structFields.43" = private unnamed_addr global [2 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:basic:uintptr", ptr @"reflect/types.structFieldName.44", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:basic:uintptr", ptr @"reflect/types.structFieldName.45", ptr null, i1 false }]
@"reflect/types.structFieldName.44" = private unnamed_addr global [10 x i8] c"asyncifysp"
@"reflect/types.structFieldName.45" = private unnamed_addr global [3 x i8] c"csp"
@"reflect/types.type:pointer:struct:{asyncifysp:basic:uintptr,csp:basic:uintptr}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{asyncifysp:basic:uintptr,csp:basic:uintptr}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.type:pointer:named:internal/task.stackState" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:internal/task.stackState", i32 0, ptr @"*internal/task.stackState$methodset", ptr null, i32 0 }
@"internal/task.$methods.unwind()" = linkonce_odr constant i8 0, align 1
@"*internal/task.stackState$methodset" = linkonce_odr constant [1 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"internal/task.$methods.unwind()", i32 ptrtoint (ptr @tinygo_unwind to i32) }]
@"reflect/types.structFieldName.46" = private unnamed_addr global [10 x i8] c"stackState"
@"reflect/types.structFieldName.47" = private unnamed_addr global [8 x i8] c"launched"
@"reflect/types.type:pointer:struct:{entry:basic:uintptr,args:basic:unsafe.Pointer,#stackState:named:internal/task.stackState,launched:basic:bool}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{entry:basic:uintptr,args:basic:unsafe.Pointer,#stackState:named:internal/task.stackState,launched:basic:bool}", i32 0, ptr @"*struct{entry uintptr; args unsafe.Pointer; internal/task.stackState; launched bool}$methodset", ptr null, i32 0 }
@"*struct{entry uintptr; args unsafe.Pointer; internal/task.stackState; launched bool}$methodset" = linkonce_odr constant [1 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"internal/task.$methods.unwind()", i32 ptrtoint (ptr @"(*struct{entry uintptr; args unsafe.Pointer; internal/task.stackState; launched bool}).unwind$wrapper for func (*internal/task.stackState).unwind()" to i32) }]
@"reflect/types.type:pointer:named:internal/task.state" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:internal/task.state", i32 0, ptr @"*internal/task.state$methodset", ptr null, i32 0 }
@"internal/task.$methods.initialize(uintptr, unsafe.Pointer, uintptr)" = linkonce_odr constant i8 0, align 1
@"internal/task.$methods.launch()" = linkonce_odr constant i8 0, align 1
@"internal/task.$methods.rewind()" = linkonce_odr constant i8 0, align 1
@"*internal/task.state$methodset" = linkonce_odr constant [4 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"internal/task.$methods.initialize(uintptr, unsafe.Pointer, uintptr)", i32 ptrtoint (ptr @"(*internal/task.state).initialize" to i32) }, %runtime.interfaceMethodInfo { ptr @"internal/task.$methods.launch()", i32 ptrtoint (ptr @tinygo_launch to i32) }, %runtime.interfaceMethodInfo { ptr @"internal/task.$methods.rewind()", i32 ptrtoint (ptr @tinygo_rewind to i32) }, %runtime.interfaceMethodInfo { ptr @"internal/task.$methods.unwind()", i32 ptrtoint (ptr @"(*internal/task.state).unwind$wrapper for func (*internal/task.stackState).unwind()" to i32) }]
@"reflect/types.structFieldName.48" = private unnamed_addr global [5 x i8] c"state"
@"reflect/types.structFieldName.49" = private unnamed_addr global [10 x i8] c"DeferFrame"
@"reflect/types.type:pointer:struct:{Next:pointer:named:internal/task.Task,Ptr:basic:unsafe.Pointer,Data:basic:uint64,gcData:named:internal/task.gcData,state:named:internal/task.state,DeferFrame:basic:unsafe.Pointer}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{Next:pointer:named:internal/task.Task,Ptr:basic:unsafe.Pointer,Data:basic:uint64,gcData:named:internal/task.gcData,state:named:internal/task.state,DeferFrame:basic:unsafe.Pointer}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/methods.Resume()" = linkonce_odr constant i8 0, align 1
@"internal/task.$methods.tail() *internal/task.Task" = linkonce_odr constant i8 0, align 1
@"*internal/task.Task$methodset" = linkonce_odr constant [2 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.Resume()", i32 ptrtoint (ptr @"(*internal/task.Task).Resume" to i32) }, %runtime.interfaceMethodInfo { ptr @"internal/task.$methods.tail() *internal/task.Task", i32 ptrtoint (ptr @"(*internal/task.Task).tail" to i32) }]
@"reflect/types.structFieldName.50" = private unnamed_addr global [3 x i8] c"top"
@"reflect/types.type:pointer:struct:{top:pointer:named:internal/task.Task}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{top:pointer:named:internal/task.Task}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.type:pointer:named:internal/task.Stack" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:internal/task.Stack", i32 0, ptr @"*internal/task.Stack$methodset", ptr null, i32 0 }
@"reflect/methods.Pop() *internal/task.Task" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Push(*internal/task.Task)" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Queue() internal/task.Queue" = linkonce_odr constant i8 0, align 1
@"*internal/task.Stack$methodset" = linkonce_odr constant [3 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.Pop() *internal/task.Task", i32 ptrtoint (ptr @"(*internal/task.Stack).Pop" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Push(*internal/task.Task)", i32 ptrtoint (ptr @"(*internal/task.Stack).Push" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Queue() internal/task.Queue", i32 ptrtoint (ptr @"(*internal/task.Stack).Queue" to i32) }]
@"reflect/types.structFieldName.51" = private unnamed_addr global [7 x i8] c"blocked"
@"reflect/types.type:pointer:struct:{locked:basic:bool,blocked:named:internal/task.Stack}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{locked:basic:bool,blocked:named:internal/task.Stack}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.type:pointer:named:sync.Mutex" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:sync.Mutex", i32 0, ptr @"*sync.Mutex$methodset", ptr null, i32 0 }
@"reflect/methods.Lock()" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Unlock()" = linkonce_odr constant i8 0, align 1
@"*sync.Mutex$methodset" = linkonce_odr constant [2 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.Lock()", i32 ptrtoint (ptr @"(*sync.Mutex).Lock" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Unlock()", i32 ptrtoint (ptr @"(*sync.Mutex).Unlock" to i32) }]
@"reflect/types.structFieldName.52" = private unnamed_addr global [1 x i8] c"m"
@"reflect/types.type:pointer:struct:{done:basic:bool,m:named:sync.Mutex}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{done:basic:bool,m:named:sync.Mutex}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.type:pointer:named:sync.Once" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:sync.Once", i32 0, ptr @"*sync.Once$methodset", ptr null, i32 0 }
@"reflect/methods.Do(func())" = linkonce_odr constant i8 0, align 1
@"*sync.Once$methodset" = linkonce_odr constant [1 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.Do(func())", i32 ptrtoint (ptr @"(*sync.Once).Do" to i32) }]
@"reflect/types.structFieldName.53" = private unnamed_addr global [4 x i8] c"once"
@"reflect/types.type:named:error" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:interface:{Error:func:{}{basic:string}}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:error", i32 ptrtoint (ptr @"interface:{Error:func:{}{basic:string}}.$typeassert" to i32) }
@"reflect/types.type:interface:{Error:func:{}{basic:string}}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.interface:interface{Error() string}$interface", i32 0, ptr null, ptr @"reflect/types.type:pointer:interface:{Error:func:{}{basic:string}}", i32 ptrtoint (ptr @"interface:{Error:func:{}{basic:string}}.$typeassert" to i32) }
@"reflect/methods.Error() string" = linkonce_odr constant i8 0, align 1
@"reflect/types.interface:interface{Error() string}$interface" = linkonce_odr constant [1 x ptr] [ptr @"reflect/methods.Error() string"]
@"reflect/types.type:pointer:interface:{Error:func:{}{basic:string}}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:interface:{Error:func:{}{basic:string}}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.type:pointer:named:error" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:error", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.structFieldName.54" = private unnamed_addr global [3 x i8] c"err"
@"struct{*os.File; once sync.Once; err error}$methodset" = linkonce_odr constant [17 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.Close() error", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).Close$wrapper for func (*os.File).Close() (err error)$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Fd() uintptr", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).Fd$wrapper for func (*os.File).Fd() uintptr$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Name() string", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).Name$wrapper for func (*os.File).Name() string$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Read([]uint8) (int, error)", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).Read$wrapper for func (*os.File).Read(b []byte) (n int, err error)$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadAt([]uint8, int64) (int, error)", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).ReadAt$wrapper for func (*os.File).ReadAt(b []byte, offset int64) (n int, err error)$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadDir(int) ([]io/fs.DirEntry, error)", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).ReadDir$wrapper for func (*os.File).ReadDir(n int) ([]io/fs.DirEntry, error)$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Readdir(int) ([]io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).Readdir$wrapper for func (*os.File).Readdir(n int) ([]io/fs.FileInfo, error)$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Readdirnames(int) ([]string, error)", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).Readdirnames$wrapper for func (*os.File).Readdirnames(n int) (names []string, err error)$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Seek(int64, int) (int64, error)", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).Seek$wrapper for func (*os.File).Seek(offset int64, whence int) (ret int64, err error)$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Stat() (io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).Stat$wrapper for func (*os.File).Stat() (io/fs.FileInfo, error)$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Sync() error", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).Sync$wrapper for func (*os.File).Sync() error$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.SyscallConn() (syscall.RawConn, error)", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).SyscallConn$wrapper for func (*os.File).SyscallConn() (syscall.RawConn, error)$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Truncate(int64) error", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).Truncate$wrapper for func (*os.File).Truncate(size int64) error$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Write([]uint8) (int, error)", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).Write$wrapper for func (*os.File).Write(b []byte) (n int, err error)$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.WriteAt([]uint8, int64) (int, error)", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).WriteAt$wrapper for func (*os.File).WriteAt(b []byte, off int64) (n int, err error)$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.WriteString(string) (int, error)", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).WriteString$wrapper for func (*os.File).WriteString(s string) (n int, err error)$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"os.$methods.readdir(int, os.readdirMode) ([]string, []io/fs.DirEntry, []io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(struct{*os.File; once sync.Once; err error}).readdir$wrapper for func (*os.File).readdir(n int, mode os.readdirMode) (names []string, dirents []io/fs.DirEntry, infos []io/fs.FileInfo, err error)$invoke" to i32) }]
@"reflect/types.type:pointer:struct:{#File:pointer:named:os.File,once:named:sync.Once,err:named:error}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{#File:pointer:named:os.File,once:named:sync.Once,err:named:error}", i32 0, ptr @"*struct{*os.File; once sync.Once; err error}$methodset", ptr null, i32 0 }
@"*struct{*os.File; once sync.Once; err error}$methodset" = linkonce_odr constant [17 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.Close() error", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).Close$wrapper for func (*os.File).Close() (err error)" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Fd() uintptr", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).Fd$wrapper for func (*os.File).Fd() uintptr" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Name() string", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).Name$wrapper for func (*os.File).Name() string" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Read([]uint8) (int, error)", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).Read$wrapper for func (*os.File).Read(b []byte) (n int, err error)" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadAt([]uint8, int64) (int, error)", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).ReadAt$wrapper for func (*os.File).ReadAt(b []byte, offset int64) (n int, err error)" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadDir(int) ([]io/fs.DirEntry, error)", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).ReadDir$wrapper for func (*os.File).ReadDir(n int) ([]io/fs.DirEntry, error)" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Readdir(int) ([]io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).Readdir$wrapper for func (*os.File).Readdir(n int) ([]io/fs.FileInfo, error)" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Readdirnames(int) ([]string, error)", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).Readdirnames$wrapper for func (*os.File).Readdirnames(n int) (names []string, err error)" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Seek(int64, int) (int64, error)", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).Seek$wrapper for func (*os.File).Seek(offset int64, whence int) (ret int64, err error)" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Stat() (io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).Stat$wrapper for func (*os.File).Stat() (io/fs.FileInfo, error)" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Sync() error", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).Sync$wrapper for func (*os.File).Sync() error" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.SyscallConn() (syscall.RawConn, error)", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).SyscallConn$wrapper for func (*os.File).SyscallConn() (syscall.RawConn, error)" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Truncate(int64) error", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).Truncate$wrapper for func (*os.File).Truncate(size int64) error" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Write([]uint8) (int, error)", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).Write$wrapper for func (*os.File).Write(b []byte) (n int, err error)" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.WriteAt([]uint8, int64) (int, error)", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).WriteAt$wrapper for func (*os.File).WriteAt(b []byte, off int64) (n int, err error)" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.WriteString(string) (int, error)", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).WriteString$wrapper for func (*os.File).WriteString(s string) (n int, err error)" to i32) }, %runtime.interfaceMethodInfo { ptr @"os.$methods.readdir(int, os.readdirMode) ([]string, []io/fs.DirEntry, []io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(*struct{*os.File; once sync.Once; err error}).readdir$wrapper for func (*os.File).readdir(n int, mode os.readdirMode) (names []string, dirents []io/fs.DirEntry, infos []io/fs.FileInfo, err error)" to i32) }]
@"main.closeOnce$methodset" = linkonce_odr constant [16 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.Fd() uintptr", i32 ptrtoint (ptr @"(main.closeOnce).Fd$wrapper for func (*os.File).Fd() uintptr$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Name() string", i32 ptrtoint (ptr @"(main.closeOnce).Name$wrapper for func (*os.File).Name() string$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Read([]uint8) (int, error)", i32 ptrtoint (ptr @"(main.closeOnce).Read$wrapper for func (*os.File).Read(b []byte) (n int, err error)$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadAt([]uint8, int64) (int, error)", i32 ptrtoint (ptr @"(main.closeOnce).ReadAt$wrapper for func (*os.File).ReadAt(b []byte, offset int64) (n int, err error)$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadDir(int) ([]io/fs.DirEntry, error)", i32 ptrtoint (ptr @"(main.closeOnce).ReadDir$wrapper for func (*os.File).ReadDir(n int) ([]io/fs.DirEntry, error)$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Readdir(int) ([]io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(main.closeOnce).Readdir$wrapper for func (*os.File).Readdir(n int) ([]io/fs.FileInfo, error)$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Readdirnames(int) ([]string, error)", i32 ptrtoint (ptr @"(main.closeOnce).Readdirnames$wrapper for func (*os.File).Readdirnames(n int) (names []string, err error)$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Seek(int64, int) (int64, error)", i32 ptrtoint (ptr @"(main.closeOnce).Seek$wrapper for func (*os.File).Seek(offset int64, whence int) (ret int64, err error)$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Stat() (io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(main.closeOnce).Stat$wrapper for func (*os.File).Stat() (io/fs.FileInfo, error)$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Sync() error", i32 ptrtoint (ptr @"(main.closeOnce).Sync$wrapper for func (*os.File).Sync() error$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.SyscallConn() (syscall.RawConn, error)", i32 ptrtoint (ptr @"(main.closeOnce).SyscallConn$wrapper for func (*os.File).SyscallConn() (syscall.RawConn, error)$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Truncate(int64) error", i32 ptrtoint (ptr @"(main.closeOnce).Truncate$wrapper for func (*os.File).Truncate(size int64) error$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Write([]uint8) (int, error)", i32 ptrtoint (ptr @"(main.closeOnce).Write$wrapper for func (*os.File).Write(b []byte) (n int, err error)$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.WriteAt([]uint8, int64) (int, error)", i32 ptrtoint (ptr @"(main.closeOnce).WriteAt$wrapper for func (*os.File).WriteAt(b []byte, off int64) (n int, err error)$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.WriteString(string) (int, error)", i32 ptrtoint (ptr @"(main.closeOnce).WriteString$wrapper for func (*os.File).WriteString(s string) (n int, err error)$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"os.$methods.readdir(int, os.readdirMode) ([]string, []io/fs.DirEntry, []io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(main.closeOnce).readdir$wrapper for func (*os.File).readdir(n int, mode os.readdirMode) (names []string, dirents []io/fs.DirEntry, infos []io/fs.FileInfo, err error)$invoke" to i32) }]
@"main.$methods.close()" = linkonce_odr constant i8 0, align 1
@"*main.closeOnce$methodset" = linkonce_odr constant [18 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.Close() error", i32 ptrtoint (ptr @"(*main.closeOnce).Close" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Fd() uintptr", i32 ptrtoint (ptr @"(*main.closeOnce).Fd$wrapper for func (*os.File).Fd() uintptr" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Name() string", i32 ptrtoint (ptr @"(*main.closeOnce).Name$wrapper for func (*os.File).Name() string" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Read([]uint8) (int, error)", i32 ptrtoint (ptr @"(*main.closeOnce).Read$wrapper for func (*os.File).Read(b []byte) (n int, err error)" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadAt([]uint8, int64) (int, error)", i32 ptrtoint (ptr @"(*main.closeOnce).ReadAt$wrapper for func (*os.File).ReadAt(b []byte, offset int64) (n int, err error)" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ReadDir(int) ([]io/fs.DirEntry, error)", i32 ptrtoint (ptr @"(*main.closeOnce).ReadDir$wrapper for func (*os.File).ReadDir(n int) ([]io/fs.DirEntry, error)" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Readdir(int) ([]io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(*main.closeOnce).Readdir$wrapper for func (*os.File).Readdir(n int) ([]io/fs.FileInfo, error)" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Readdirnames(int) ([]string, error)", i32 ptrtoint (ptr @"(*main.closeOnce).Readdirnames$wrapper for func (*os.File).Readdirnames(n int) (names []string, err error)" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Seek(int64, int) (int64, error)", i32 ptrtoint (ptr @"(*main.closeOnce).Seek$wrapper for func (*os.File).Seek(offset int64, whence int) (ret int64, err error)" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Stat() (io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(*main.closeOnce).Stat$wrapper for func (*os.File).Stat() (io/fs.FileInfo, error)" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Sync() error", i32 ptrtoint (ptr @"(*main.closeOnce).Sync$wrapper for func (*os.File).Sync() error" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.SyscallConn() (syscall.RawConn, error)", i32 ptrtoint (ptr @"(*main.closeOnce).SyscallConn$wrapper for func (*os.File).SyscallConn() (syscall.RawConn, error)" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Truncate(int64) error", i32 ptrtoint (ptr @"(*main.closeOnce).Truncate$wrapper for func (*os.File).Truncate(size int64) error" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Write([]uint8) (int, error)", i32 ptrtoint (ptr @"(*main.closeOnce).Write$wrapper for func (*os.File).Write(b []byte) (n int, err error)" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.WriteAt([]uint8, int64) (int, error)", i32 ptrtoint (ptr @"(*main.closeOnce).WriteAt$wrapper for func (*os.File).WriteAt(b []byte, off int64) (n int, err error)" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.WriteString(string) (int, error)", i32 ptrtoint (ptr @"(*main.closeOnce).WriteString$wrapper for func (*os.File).WriteString(s string) (n int, err error)" to i32) }, %runtime.interfaceMethodInfo { ptr @"main.$methods.close()", i32 ptrtoint (ptr @"(*main.closeOnce).close" to i32) }, %runtime.interfaceMethodInfo { ptr @"os.$methods.readdir(int, os.readdirMode) ([]string, []io/fs.DirEntry, []io/fs.FileInfo, error)", i32 ptrtoint (ptr @"(*main.closeOnce).readdir$wrapper for func (*os.File).readdir(n int, mode os.readdirMode) (names []string, dirents []io/fs.DirEntry, infos []io/fs.FileInfo, err error)" to i32) }]
@"main$string.55" = internal unnamed_addr constant [24 x i8] c"exec: Stdout already set", align 1
@"main$string.56" = internal unnamed_addr constant [38 x i8] c"exec: StdoutPipe after process started", align 1
@"main$string.57" = internal unnamed_addr constant [1 x i8] c" ", align 1
@"main$string.58" = internal unnamed_addr constant [17 x i8] c"exec: not started", align 1
@"main$string.59" = internal unnamed_addr constant [29 x i8] c"exec: Wait was already called", align 1
@"reflect/types.type:pointer:named:main.ExitError" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:main.ExitError", i32 0, ptr @"*main.ExitError$methodset", ptr null, i32 0 }
@"reflect/types.type:named:main.ExitError" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{#ProcessState:pointer:named:os.ProcessState,Stderr:slice:basic:uint8}", i32 0, ptr @"main.ExitError$methodset", ptr @"reflect/types.type:pointer:named:main.ExitError", i32 0 }
@"reflect/types.type:struct:{#ProcessState:pointer:named:os.ProcessState,Stderr:slice:basic:uint8}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.60", i32 0, ptr @"struct{*os.ProcessState; Stderr []byte}$methodset", ptr @"reflect/types.type:pointer:struct:{#ProcessState:pointer:named:os.ProcessState,Stderr:slice:basic:uint8}", i32 0 }
@"reflect/types.structFields.60" = private unnamed_addr global [2 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:pointer:named:os.ProcessState", ptr @"reflect/types.structFieldName.62", ptr null, i1 true }, %runtime.structField { ptr @"reflect/types.type:slice:basic:uint8", ptr @"reflect/types.structFieldName.63", ptr null, i1 false }]
@"reflect/types.type:pointer:named:os.ProcessState" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:os.ProcessState", i32 0, ptr @"*os.ProcessState$methodset", ptr null, i32 0 }
@"reflect/types.type:named:os.ProcessState" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:os.ProcessState", i32 0 }
@"reflect/types.type:struct:{}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.61", i32 0, ptr null, ptr @"reflect/types.type:pointer:struct:{}", i32 0 }
@"reflect/types.structFields.61" = private unnamed_addr global [0 x %runtime.structField] zeroinitializer
@"reflect/types.type:pointer:struct:{}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/methods.ExitCode() int" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Success() bool" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Sys() interface{}" = linkonce_odr constant i8 0, align 1
@"*os.ProcessState$methodset" = linkonce_odr constant [4 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.ExitCode() int", i32 ptrtoint (ptr @"(*os.ProcessState).ExitCode" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.String() string", i32 ptrtoint (ptr @"(*os.ProcessState).String" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Success() bool", i32 ptrtoint (ptr @"(*os.ProcessState).Success" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Sys() interface{}", i32 ptrtoint (ptr @"(*os.ProcessState).Sys" to i32) }]
@"reflect/types.structFieldName.62" = private unnamed_addr global [12 x i8] c"ProcessState"
@"reflect/types.structFieldName.63" = private unnamed_addr global [6 x i8] c"Stderr"
@"struct{*os.ProcessState; Stderr []byte}$methodset" = linkonce_odr constant [4 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.ExitCode() int", i32 ptrtoint (ptr @"(struct{*os.ProcessState; Stderr []byte}).ExitCode$wrapper for func (*os.ProcessState).ExitCode() int$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.String() string", i32 ptrtoint (ptr @"(struct{*os.ProcessState; Stderr []byte}).String$wrapper for func (*os.ProcessState).String() string$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Success() bool", i32 ptrtoint (ptr @"(struct{*os.ProcessState; Stderr []byte}).Success$wrapper for func (*os.ProcessState).Success() bool$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Sys() interface{}", i32 ptrtoint (ptr @"(struct{*os.ProcessState; Stderr []byte}).Sys$wrapper for func (*os.ProcessState).Sys() interface{}$invoke" to i32) }]
@"reflect/types.type:pointer:struct:{#ProcessState:pointer:named:os.ProcessState,Stderr:slice:basic:uint8}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{#ProcessState:pointer:named:os.ProcessState,Stderr:slice:basic:uint8}", i32 0, ptr @"*struct{*os.ProcessState; Stderr []byte}$methodset", ptr null, i32 0 }
@"*struct{*os.ProcessState; Stderr []byte}$methodset" = linkonce_odr constant [4 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.ExitCode() int", i32 ptrtoint (ptr @"(*struct{*os.ProcessState; Stderr []byte}).ExitCode$wrapper for func (*os.ProcessState).ExitCode() int" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.String() string", i32 ptrtoint (ptr @"(*struct{*os.ProcessState; Stderr []byte}).String$wrapper for func (*os.ProcessState).String() string" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Success() bool", i32 ptrtoint (ptr @"(*struct{*os.ProcessState; Stderr []byte}).Success$wrapper for func (*os.ProcessState).Success() bool" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Sys() interface{}", i32 ptrtoint (ptr @"(*struct{*os.ProcessState; Stderr []byte}).Sys$wrapper for func (*os.ProcessState).Sys() interface{}" to i32) }]
@"main.ExitError$methodset" = linkonce_odr constant [4 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.ExitCode() int", i32 ptrtoint (ptr @"(main.ExitError).ExitCode$wrapper for func (*os.ProcessState).ExitCode() int$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.String() string", i32 ptrtoint (ptr @"(main.ExitError).String$wrapper for func (*os.ProcessState).String() string$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Success() bool", i32 ptrtoint (ptr @"(main.ExitError).Success$wrapper for func (*os.ProcessState).Success() bool$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Sys() interface{}", i32 ptrtoint (ptr @"(main.ExitError).Sys$wrapper for func (*os.ProcessState).Sys() interface{}$invoke" to i32) }]
@"*main.ExitError$methodset" = linkonce_odr constant [5 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.Error() string", i32 ptrtoint (ptr @"(*main.ExitError).Error" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.ExitCode() int", i32 ptrtoint (ptr @"(*main.ExitError).ExitCode$wrapper for func (*os.ProcessState).ExitCode() int" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.String() string", i32 ptrtoint (ptr @"(*main.ExitError).String$wrapper for func (*os.ProcessState).String() string" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Success() bool", i32 ptrtoint (ptr @"(*main.ExitError).Success$wrapper for func (*os.ProcessState).Success() bool" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Sys() interface{}", i32 ptrtoint (ptr @"(*main.ExitError).Sys$wrapper for func (*os.ProcessState).Sys() interface{}" to i32) }]
@"main$string.64" = internal unnamed_addr constant [9 x i8] c"/dev/null", align 1
@"reflect/types.typeid:pointer:named:os.File" = external constant i8
@os.ErrProcessDone = external global %runtime._interface, align 4
@"main$string.65" = internal unnamed_addr constant [33 x i8] c"exec: error sending signal to Cmd", align 1
@"reflect/types.type:named:main.wrappedError" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{prefix:basic:string,err:named:error}", i32 0, ptr @"main.wrappedError$methodset", ptr @"reflect/types.type:pointer:named:main.wrappedError", i32 0 }
@"reflect/types.type:struct:{prefix:basic:string,err:named:error}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.66", i32 0, ptr null, ptr @"reflect/types.type:pointer:struct:{prefix:basic:string,err:named:error}", i32 0 }
@"reflect/types.structFields.66" = private unnamed_addr global [2 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:basic:string", ptr @"reflect/types.structFieldName.67", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:named:error", ptr @"reflect/types.structFieldName.68", ptr null, i1 false }]
@"reflect/types.structFieldName.67" = private unnamed_addr global [6 x i8] c"prefix"
@"reflect/types.structFieldName.68" = private unnamed_addr global [3 x i8] c"err"
@"reflect/types.type:pointer:struct:{prefix:basic:string,err:named:error}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{prefix:basic:string,err:named:error}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/methods.Unwrap() error" = linkonce_odr constant i8 0, align 1
@"main.wrappedError$methodset" = linkonce_odr constant [2 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.Error() string", i32 ptrtoint (ptr @"(main.wrappedError).Error$invoke" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Unwrap() error", i32 ptrtoint (ptr @"(main.wrappedError).Unwrap$invoke" to i32) }]
@"reflect/types.type:pointer:named:main.wrappedError" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:main.wrappedError", i32 0, ptr @"*main.wrappedError$methodset", ptr null, i32 0 }
@"main$string.69" = internal unnamed_addr constant [17 x i8] c"main.wrappedError", align 1
@"main$string.70" = internal unnamed_addr constant [5 x i8] c"Error", align 1
@"main$string.71" = internal unnamed_addr constant [17 x i8] c"main.wrappedError", align 1
@"main$string.72" = internal unnamed_addr constant [6 x i8] c"Unwrap", align 1
@"*main.wrappedError$methodset" = linkonce_odr constant [2 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.Error() string", i32 ptrtoint (ptr @"(*main.wrappedError).Error$wrapper for func (main.wrappedError).Error() string" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Unwrap() error", i32 ptrtoint (ptr @"(*main.wrappedError).Unwrap$wrapper for func (main.wrappedError).Unwrap() error" to i32) }]
@"main$string.73" = internal unnamed_addr constant [31 x i8] c"blocking select matched no case", align 1
@"main$pack" = internal unnamed_addr constant { %runtime._string } { %runtime._string { ptr @"main$string.73", i32 31 } }
@"main$string.74" = internal unnamed_addr constant [9 x i8] c"/dev/null", align 1
@"runtime/gc.layout:42-2c926b9a925" = linkonce_odr unnamed_addr constant { i32, [6 x i8] } { i32 42, [6 x i8] c"\02\C9&\B9\A9%" }
@"main$string.75" = internal unnamed_addr constant [11 x i8] c"nil Context", align 1
@"main$pack.76" = internal unnamed_addr constant { %runtime._string } { %runtime._string { ptr @"main$string.75", i32 11 } }
@"main$string.77" = internal unnamed_addr constant [2 x i8] c"./", align 1
@"main$string.78" = internal unnamed_addr constant [14 x i8] c"\0A... omitting ", align 1
@"main$string.79" = internal unnamed_addr constant [11 x i8] c" bytes ...\0A", align 1
@"main$string.80" = internal unnamed_addr constant [1 x i8] c"=", align 1
@"main$string.81" = internal unnamed_addr constant [1 x i8] c"=", align 1
@"main$string.82" = internal unnamed_addr constant [1 x i8] c"=", align 1
@"main$string.83" = internal unnamed_addr constant [10 x i8] c"SYSTEMROOT", align 1
@"main$string.84" = internal unnamed_addr constant [10 x i8] c"SYSTEMROOT", align 1
@"main$string.85" = internal unnamed_addr constant [11 x i8] c"SYSTEMROOT=", align 1

declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #0

declare void @runtime.trackPointer(ptr nocapture readonly, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  %0 = call %runtime._interface @errors.New(ptr nonnull @"main$string", i32 57, ptr undef) #14
  %1 = extractvalue %runtime._interface %0, 1
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %.elt = extractvalue %runtime._interface %0, 0
  store i32 %.elt, ptr @main.ErrDot, align 8
  %.elt1 = extractvalue %runtime._interface %0, 1
  store ptr %.elt1, ptr getelementptr inbounds (%runtime._interface, ptr @main.ErrDot, i32 0, i32 1), align 4
  ret void
}

declare %runtime._interface @errors.New(ptr, i32, ptr) #0

; Function Attrs: nounwind
define hidden void @main.main(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden { %runtime._string, %runtime._interface } @main.LookPath(ptr %arg0.data, i32 %arg0.len, ptr %context) unnamed_addr #1 {
entry:
  ret { %runtime._string, %runtime._interface } zeroinitializer
}

; Function Attrs: nounwind
define hidden i1 @main.skipStdinCopyError(i32 %arg0.typecode, ptr %arg0.value, ptr %context) unnamed_addr #1 {
entry:
  ret i1 false
}

; Function Attrs: nounwind
define hidden %runtime._string @"(*main.Error).Error"(ptr dereferenceable_or_null(16) %e, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %e, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %.unpack = load ptr, ptr %e, align 4
  %.elt5 = getelementptr inbounds %runtime._string, ptr %e, i32 0, i32 1
  %.unpack6 = load i32, ptr %.elt5, align 4
  call void @runtime.trackPointer(ptr %.unpack, ptr undef) #14
  %1 = call %runtime._string @strconv.Quote(ptr %.unpack, i32 %.unpack6, ptr undef) #14
  %2 = extractvalue %runtime._string %1, 0
  call void @runtime.trackPointer(ptr %2, ptr undef) #14
  %3 = extractvalue %runtime._string %1, 0
  %4 = extractvalue %runtime._string %1, 1
  %5 = call %runtime._string @runtime.stringConcat(ptr nonnull @"main$string.1", i32 6, ptr %3, i32 %4, ptr undef) #14
  %6 = extractvalue %runtime._string %5, 0
  call void @runtime.trackPointer(ptr %6, ptr undef) #14
  %7 = extractvalue %runtime._string %5, 0
  %8 = extractvalue %runtime._string %5, 1
  %9 = call %runtime._string @runtime.stringConcat(ptr %7, i32 %8, ptr nonnull @"main$string.2", i32 2, ptr undef) #14
  %10 = extractvalue %runtime._string %9, 0
  call void @runtime.trackPointer(ptr %10, ptr undef) #14
  br i1 false, label %gep.throw1, label %gep.next2

gep.next2:                                        ; preds = %deref.next
  br i1 false, label %deref.throw3, label %deref.next4

deref.next4:                                      ; preds = %gep.next2
  %11 = getelementptr inbounds %main.Error, ptr %e, i32 0, i32 1
  %.unpack7 = load i32, ptr %11, align 4
  %.elt8 = getelementptr inbounds %main.Error, ptr %e, i32 0, i32 1, i32 1
  %.unpack9 = load ptr, ptr %.elt8, align 4
  call void @runtime.trackPointer(ptr %.unpack9, ptr undef) #14
  %12 = call %runtime._string @"interface:{Error:func:{}{basic:string}}.Error$invoke"(ptr %.unpack9, i32 %.unpack7, ptr undef) #14
  %13 = extractvalue %runtime._string %12, 0
  call void @runtime.trackPointer(ptr %13, ptr undef) #14
  %14 = extractvalue %runtime._string %9, 0
  %15 = extractvalue %runtime._string %9, 1
  %16 = extractvalue %runtime._string %12, 0
  %17 = extractvalue %runtime._string %12, 1
  %18 = call %runtime._string @runtime.stringConcat(ptr %14, i32 %15, ptr %16, i32 %17, ptr undef) #14
  %19 = extractvalue %runtime._string %18, 0
  call void @runtime.trackPointer(ptr %19, ptr undef) #14
  ret %runtime._string %18

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable

gep.throw1:                                       ; preds = %deref.next
  unreachable

deref.throw3:                                     ; preds = %gep.next2
  unreachable
}

declare void @runtime.nilPanic(ptr) #0

declare %runtime._string @strconv.Quote(ptr, i32, ptr) #0

declare %runtime._string @runtime.stringConcat(ptr dereferenceable_or_null(1), i32, ptr dereferenceable_or_null(1), i32, ptr) #0

declare %runtime._string @"interface:{Error:func:{}{basic:string}}.Error$invoke"(ptr, i32, ptr) #2

; Function Attrs: nounwind
define hidden %runtime._interface @"(*main.Error).Unwrap"(ptr dereferenceable_or_null(16) %e, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %e, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = getelementptr inbounds %main.Error, ptr %e, i32 0, i32 1
  %.unpack = load i32, ptr %1, align 4
  %2 = insertvalue %runtime._interface undef, i32 %.unpack, 0
  %.elt1 = getelementptr inbounds %main.Error, ptr %e, i32 0, i32 1, i32 1
  %.unpack2 = load ptr, ptr %.elt1, align 4
  %3 = insertvalue %runtime._interface %2, ptr %.unpack2, 1
  call void @runtime.trackPointer(ptr %.unpack2, ptr undef) #14
  ret %runtime._interface %3

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define hidden %runtime._string @"(main.wrappedError).Error"(%main.wrappedError %w, ptr %context) unnamed_addr #1 {
entry:
  %w1 = alloca %main.wrappedError, align 8
  store ptr null, ptr %w1, align 8
  %w1.repack5 = getelementptr inbounds %runtime._string, ptr %w1, i32 0, i32 1
  store i32 0, ptr %w1.repack5, align 4
  %w1.repack4 = getelementptr inbounds %main.wrappedError, ptr %w1, i32 0, i32 1
  store i32 0, ptr %w1.repack4, align 8
  %w1.repack4.repack6 = getelementptr inbounds %main.wrappedError, ptr %w1, i32 0, i32 1, i32 1
  store ptr null, ptr %w1.repack4.repack6, align 4
  call void @runtime.trackPointer(ptr nonnull %w1, ptr undef) #14
  %w.elt = extractvalue %main.wrappedError %w, 0
  %w.elt.elt = extractvalue %runtime._string %w.elt, 0
  store ptr %w.elt.elt, ptr %w1, align 8
  %w1.repack9 = getelementptr inbounds %runtime._string, ptr %w1, i32 0, i32 1
  %w.elt.elt10 = extractvalue %runtime._string %w.elt, 1
  store i32 %w.elt.elt10, ptr %w1.repack9, align 4
  %w1.repack7 = getelementptr inbounds %main.wrappedError, ptr %w1, i32 0, i32 1
  %w.elt8 = extractvalue %main.wrappedError %w, 1
  %w.elt8.elt = extractvalue %runtime._interface %w.elt8, 0
  store i32 %w.elt8.elt, ptr %w1.repack7, align 8
  %w1.repack7.repack11 = getelementptr inbounds %main.wrappedError, ptr %w1, i32 0, i32 1, i32 1
  %w.elt8.elt12 = extractvalue %runtime._interface %w.elt8, 1
  store ptr %w.elt8.elt12, ptr %w1.repack7.repack11, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %.unpack = load ptr, ptr %w1, align 8
  %.elt13 = getelementptr inbounds %runtime._string, ptr %w1, i32 0, i32 1
  %.unpack14 = load i32, ptr %.elt13, align 4
  call void @runtime.trackPointer(ptr %.unpack, ptr undef) #14
  %0 = call %runtime._string @runtime.stringConcat(ptr %.unpack, i32 %.unpack14, ptr nonnull @"main$string.3", i32 2, ptr undef) #14
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  br i1 false, label %deref.throw2, label %deref.next3

deref.next3:                                      ; preds = %deref.next
  %2 = getelementptr inbounds %main.wrappedError, ptr %w1, i32 0, i32 1
  %.unpack15 = load i32, ptr %2, align 8
  %.elt16 = getelementptr inbounds %main.wrappedError, ptr %w1, i32 0, i32 1, i32 1
  %.unpack17 = load ptr, ptr %.elt16, align 4
  call void @runtime.trackPointer(ptr %.unpack17, ptr undef) #14
  %3 = call %runtime._string @"interface:{Error:func:{}{basic:string}}.Error$invoke"(ptr %.unpack17, i32 %.unpack15, ptr undef) #14
  %4 = extractvalue %runtime._string %3, 0
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  %5 = extractvalue %runtime._string %0, 0
  %6 = extractvalue %runtime._string %0, 1
  %7 = extractvalue %runtime._string %3, 0
  %8 = extractvalue %runtime._string %3, 1
  %9 = call %runtime._string @runtime.stringConcat(ptr %5, i32 %6, ptr %7, i32 %8, ptr undef) #14
  %10 = extractvalue %runtime._string %9, 0
  call void @runtime.trackPointer(ptr %10, ptr undef) #14
  ret %runtime._string %9

deref.throw:                                      ; preds = %entry
  unreachable

deref.throw2:                                     ; preds = %deref.next
  unreachable
}

; Function Attrs: nounwind
define hidden %runtime._interface @"(main.wrappedError).Unwrap"(%main.wrappedError %w, ptr %context) unnamed_addr #1 {
entry:
  %w1 = alloca %main.wrappedError, align 8
  store ptr null, ptr %w1, align 8
  %w1.repack3 = getelementptr inbounds %runtime._string, ptr %w1, i32 0, i32 1
  store i32 0, ptr %w1.repack3, align 4
  %w1.repack2 = getelementptr inbounds %main.wrappedError, ptr %w1, i32 0, i32 1
  store i32 0, ptr %w1.repack2, align 8
  %w1.repack2.repack4 = getelementptr inbounds %main.wrappedError, ptr %w1, i32 0, i32 1, i32 1
  store ptr null, ptr %w1.repack2.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %w1, ptr undef) #14
  %w.elt = extractvalue %main.wrappedError %w, 0
  %w.elt.elt = extractvalue %runtime._string %w.elt, 0
  store ptr %w.elt.elt, ptr %w1, align 8
  %w1.repack7 = getelementptr inbounds %runtime._string, ptr %w1, i32 0, i32 1
  %w.elt.elt8 = extractvalue %runtime._string %w.elt, 1
  store i32 %w.elt.elt8, ptr %w1.repack7, align 4
  %w1.repack5 = getelementptr inbounds %main.wrappedError, ptr %w1, i32 0, i32 1
  %w.elt6 = extractvalue %main.wrappedError %w, 1
  %w.elt6.elt = extractvalue %runtime._interface %w.elt6, 0
  store i32 %w.elt6.elt, ptr %w1.repack5, align 8
  %w1.repack5.repack9 = getelementptr inbounds %main.wrappedError, ptr %w1, i32 0, i32 1, i32 1
  %w.elt6.elt10 = extractvalue %runtime._interface %w.elt6, 1
  store ptr %w.elt6.elt10, ptr %w1.repack5.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = getelementptr inbounds %main.wrappedError, ptr %w1, i32 0, i32 1
  %.unpack = load i32, ptr %0, align 8
  %1 = insertvalue %runtime._interface undef, i32 %.unpack, 0
  %.elt11 = getelementptr inbounds %main.wrappedError, ptr %w1, i32 0, i32 1, i32 1
  %.unpack12 = load ptr, ptr %.elt11, align 4
  %2 = insertvalue %runtime._interface %1, ptr %.unpack12, 1
  call void @runtime.trackPointer(ptr %.unpack12, ptr undef) #14
  ret %runtime._interface %2

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define hidden { { ptr, i32, i32 }, %runtime._interface } @"(*main.Cmd).CombinedOutput"(ptr dereferenceable_or_null(168) %c, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %c, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 5
  %.unpack = load i32, ptr %1, align 4
  %.elt13 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 5, i32 1
  %.unpack14 = load ptr, ptr %.elt13, align 4
  call void @runtime.trackPointer(ptr %.unpack14, ptr undef) #14
  %.not = icmp eq i32 %.unpack, 0
  br i1 %.not, label %if.done, label %if.then

if.then:                                          ; preds = %deref.next
  %2 = call %runtime._interface @errors.New(ptr nonnull @"main$string.4", i32 24, ptr undef) #14
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %4 = insertvalue { { ptr, i32, i32 }, %runtime._interface } zeroinitializer, %runtime._interface %2, 1
  ret { { ptr, i32, i32 }, %runtime._interface } %4

if.done:                                          ; preds = %deref.next
  br i1 false, label %gep.throw3, label %gep.next4

gep.next4:                                        ; preds = %if.done
  br i1 false, label %deref.throw5, label %deref.next6

deref.next6:                                      ; preds = %gep.next4
  %5 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 6
  %.unpack15 = load i32, ptr %5, align 4
  %.elt16 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 6, i32 1
  %.unpack17 = load ptr, ptr %.elt16, align 4
  call void @runtime.trackPointer(ptr %.unpack17, ptr undef) #14
  %.not18 = icmp eq i32 %.unpack15, 0
  br i1 %.not18, label %if.done2, label %if.then1

if.then1:                                         ; preds = %deref.next6
  %6 = call %runtime._interface @errors.New(ptr nonnull @"main$string.5", i32 24, ptr undef) #14
  %7 = extractvalue %runtime._interface %6, 1
  call void @runtime.trackPointer(ptr %7, ptr undef) #14
  %8 = insertvalue { { ptr, i32, i32 }, %runtime._interface } zeroinitializer, %runtime._interface %6, 1
  ret { { ptr, i32, i32 }, %runtime._interface } %8

if.done2:                                         ; preds = %deref.next6
  %b = call ptr @runtime.alloc(i32 20, ptr nonnull inttoptr (i32 75 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %b, ptr undef) #14
  br i1 false, label %gep.throw7, label %gep.next8

gep.next8:                                        ; preds = %if.done2
  call void @runtime.trackPointer(ptr nonnull %b, ptr undef) #14
  br i1 false, label %store.throw, label %store.next

store.next:                                       ; preds = %gep.next8
  %9 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 5
  store i32 ptrtoint (ptr @"reflect/types.type:pointer:named:bytes.Buffer" to i32), ptr %9, align 4
  %.repack19 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 5, i32 1
  store ptr %b, ptr %.repack19, align 4
  br i1 false, label %gep.throw9, label %gep.next10

gep.next10:                                       ; preds = %store.next
  call void @runtime.trackPointer(ptr nonnull %b, ptr undef) #14
  br i1 false, label %store.throw11, label %store.next12

store.next12:                                     ; preds = %gep.next10
  %10 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 6
  store i32 ptrtoint (ptr @"reflect/types.type:pointer:named:bytes.Buffer" to i32), ptr %10, align 4
  %.repack21 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 6, i32 1
  store ptr %b, ptr %.repack21, align 4
  %11 = call %runtime._interface @"(*main.Cmd).Run"(ptr nonnull %c, ptr undef)
  %12 = extractvalue %runtime._interface %11, 1
  call void @runtime.trackPointer(ptr %12, ptr undef) #14
  %13 = call { ptr, i32, i32 } @"(*bytes.Buffer).Bytes"(ptr nonnull %b, ptr undef) #14
  %14 = extractvalue { ptr, i32, i32 } %13, 0
  call void @runtime.trackPointer(ptr %14, ptr undef) #14
  %15 = insertvalue { { ptr, i32, i32 }, %runtime._interface } zeroinitializer, { ptr, i32, i32 } %13, 0
  %16 = insertvalue { { ptr, i32, i32 }, %runtime._interface } %15, %runtime._interface %11, 1
  ret { { ptr, i32, i32 }, %runtime._interface } %16

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable

gep.throw3:                                       ; preds = %if.done
  unreachable

deref.throw5:                                     ; preds = %gep.next4
  unreachable

gep.throw7:                                       ; preds = %if.done2
  unreachable

store.throw:                                      ; preds = %gep.next8
  unreachable

gep.throw9:                                       ; preds = %store.next
  unreachable

store.throw11:                                    ; preds = %gep.next10
  unreachable
}

declare { ptr, i32, i32 } @"(*bytes.Buffer).Bytes"(ptr dereferenceable_or_null(20), ptr) #0

declare i32 @"(*bytes.Buffer).Cap"(ptr dereferenceable_or_null(20), ptr) #0

declare void @"(*bytes.Buffer).Grow"(ptr dereferenceable_or_null(20), i32, ptr) #0

declare i32 @"(*bytes.Buffer).Len"(ptr dereferenceable_or_null(20), ptr) #0

declare { ptr, i32, i32 } @"(*bytes.Buffer).Next"(ptr dereferenceable_or_null(20), i32, ptr) #0

declare { i32, %runtime._interface } @"(*bytes.Buffer).Read"(ptr dereferenceable_or_null(20), ptr, i32, i32, ptr) #0

declare { i8, %runtime._interface } @"(*bytes.Buffer).ReadByte"(ptr dereferenceable_or_null(20), ptr) #0

declare { { ptr, i32, i32 }, %runtime._interface } @"(*bytes.Buffer).ReadBytes"(ptr dereferenceable_or_null(20), i8, ptr) #0

declare { i64, %runtime._interface } @"(*bytes.Buffer).ReadFrom"(ptr dereferenceable_or_null(20), i32, ptr, ptr) #0

declare { i32, i32, %runtime._interface } @"(*bytes.Buffer).ReadRune"(ptr dereferenceable_or_null(20), ptr) #0

declare { %runtime._string, %runtime._interface } @"(*bytes.Buffer).ReadString"(ptr dereferenceable_or_null(20), i8, ptr) #0

declare void @"(*bytes.Buffer).Reset"(ptr dereferenceable_or_null(20), ptr) #0

declare %runtime._string @"(*bytes.Buffer).String"(ptr dereferenceable_or_null(20), ptr) #0

declare void @"(*bytes.Buffer).Truncate"(ptr dereferenceable_or_null(20), i32, ptr) #0

declare %runtime._interface @"(*bytes.Buffer).UnreadByte"(ptr dereferenceable_or_null(20), ptr) #0

declare %runtime._interface @"(*bytes.Buffer).UnreadRune"(ptr dereferenceable_or_null(20), ptr) #0

declare { i32, %runtime._interface } @"(*bytes.Buffer).Write"(ptr dereferenceable_or_null(20), ptr, i32, i32, ptr) #0

declare %runtime._interface @"(*bytes.Buffer).WriteByte"(ptr dereferenceable_or_null(20), i8, ptr) #0

declare { i32, %runtime._interface } @"(*bytes.Buffer).WriteRune"(ptr dereferenceable_or_null(20), i32, ptr) #0

declare { i32, %runtime._interface } @"(*bytes.Buffer).WriteString"(ptr dereferenceable_or_null(20), ptr, i32, ptr) #0

declare { i64, %runtime._interface } @"(*bytes.Buffer).WriteTo"(ptr dereferenceable_or_null(20), i32, ptr, ptr) #0

declare i1 @"(*bytes.Buffer).empty"(ptr dereferenceable_or_null(20), ptr) #0

declare i32 @"(*bytes.Buffer).grow"(ptr dereferenceable_or_null(20), i32, ptr) #0

declare { { ptr, i32, i32 }, %runtime._interface } @"(*bytes.Buffer).readSlice"(ptr dereferenceable_or_null(20), i8, ptr) #0

declare { i32, i1 } @"(*bytes.Buffer).tryGrowByReslice"(ptr dereferenceable_or_null(20), i32, ptr) #0

; Function Attrs: nounwind
define hidden %runtime._interface @"(*main.Cmd).Run"(ptr dereferenceable_or_null(168) %c, ptr %context) unnamed_addr #1 {
entry:
  %0 = call %runtime._interface @"(*main.Cmd).Start"(ptr %c, ptr undef)
  %1 = extractvalue %runtime._interface %0, 1
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = extractvalue %runtime._interface %0, 0
  %.not = icmp eq i32 %2, 0
  br i1 %.not, label %if.done, label %if.then

if.then:                                          ; preds = %entry
  ret %runtime._interface %0

if.done:                                          ; preds = %entry
  %3 = call %runtime._interface @"(*main.Cmd).Wait"(ptr %c, ptr undef)
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  ret %runtime._interface %3
}

; Function Attrs: nounwind
define hidden { ptr, i32, i32 } @"(*main.Cmd).Environ"(ptr dereferenceable_or_null(168) %c, ptr %context) unnamed_addr #1 {
entry:
  %0 = call { { ptr, i32, i32 }, %runtime._interface } @"(*main.Cmd).environ"(ptr %c, ptr undef)
  %1 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %0, 0
  %2 = extractvalue { ptr, i32, i32 } %1, 0
  call void @runtime.trackPointer(ptr %2, ptr undef) #14
  %3 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %0, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  %5 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %0, 0
  ret { ptr, i32, i32 } %5
}

; Function Attrs: nounwind
define hidden { { ptr, i32, i32 }, %runtime._interface } @"(*main.Cmd).environ"(ptr dereferenceable_or_null(168) %c, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %c, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 2
  %.unpack = load ptr, ptr %1, align 4
  %.elt1 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 2, i32 1
  %.unpack2 = load i32, ptr %.elt1, align 4
  %.elt3 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 2, i32 2
  %.unpack4 = load i32, ptr %.elt3, align 4
  call void @runtime.trackPointer(ptr %.unpack, ptr undef) #14
  %2 = call { ptr, i32, i32 } @main.dedupEnv(ptr %.unpack, i32 %.unpack2, i32 %.unpack4, ptr undef)
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %4 = extractvalue { ptr, i32, i32 } %2, 0
  %5 = extractvalue { ptr, i32, i32 } %2, 1
  %6 = extractvalue { ptr, i32, i32 } %2, 2
  %7 = call { ptr, i32, i32 } @main.addCriticalEnv(ptr %4, i32 %5, i32 %6, ptr undef)
  %8 = extractvalue { ptr, i32, i32 } %7, 0
  call void @runtime.trackPointer(ptr %8, ptr undef) #14
  %9 = insertvalue { { ptr, i32, i32 }, %runtime._interface } zeroinitializer, { ptr, i32, i32 } %7, 0
  %10 = insertvalue { { ptr, i32, i32 }, %runtime._interface } %9, %runtime._interface zeroinitializer, 1
  ret { { ptr, i32, i32 }, %runtime._interface } %10

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define hidden { { ptr, i32, i32 }, %runtime._interface } @"(*main.Cmd).Output"(ptr dereferenceable_or_null(168) %c, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %c, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 5
  %.unpack = load i32, ptr %1, align 4
  %.elt32 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 5, i32 1
  %.unpack33 = load ptr, ptr %.elt32, align 4
  call void @runtime.trackPointer(ptr %.unpack33, ptr undef) #14
  %.not = icmp eq i32 %.unpack, 0
  br i1 %.not, label %if.done, label %if.then

if.then:                                          ; preds = %deref.next
  %2 = call %runtime._interface @errors.New(ptr nonnull @"main$string.8", i32 24, ptr undef) #14
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %4 = insertvalue { { ptr, i32, i32 }, %runtime._interface } zeroinitializer, %runtime._interface %2, 1
  ret { { ptr, i32, i32 }, %runtime._interface } %4

if.done:                                          ; preds = %deref.next
  %stdout = call ptr @runtime.alloc(i32 20, ptr nonnull inttoptr (i32 75 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %stdout, ptr undef) #14
  br i1 false, label %gep.throw6, label %gep.next7

gep.next7:                                        ; preds = %if.done
  call void @runtime.trackPointer(ptr nonnull %stdout, ptr undef) #14
  br i1 false, label %store.throw, label %store.next

store.next:                                       ; preds = %gep.next7
  %5 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 5
  store i32 ptrtoint (ptr @"reflect/types.type:pointer:named:bytes.Buffer" to i32), ptr %5, align 4
  %.repack34 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 5, i32 1
  store ptr %stdout, ptr %.repack34, align 4
  br i1 false, label %gep.throw8, label %gep.next9

gep.next9:                                        ; preds = %store.next
  br i1 false, label %deref.throw10, label %deref.next11

deref.next11:                                     ; preds = %gep.next9
  %6 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 6
  %.unpack36 = load i32, ptr %6, align 4
  %.elt37 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 6, i32 1
  %.unpack38 = load ptr, ptr %.elt37, align 4
  call void @runtime.trackPointer(ptr %.unpack38, ptr undef) #14
  %7 = icmp eq i32 %.unpack36, 0
  br i1 %7, label %if.then1, label %if.done2

if.then1:                                         ; preds = %deref.next11
  br i1 false, label %gep.throw12, label %gep.next13

gep.next13:                                       ; preds = %if.then1
  %8 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 6
  %complit = call ptr @runtime.alloc(i32 40, ptr nonnull inttoptr (i32 1173 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %complit, ptr undef) #14
  br i1 false, label %store.throw14, label %store.next15

store.next15:                                     ; preds = %gep.next13
  store i32 32768, ptr %complit, align 4
  call void @runtime.trackPointer(ptr nonnull %complit, ptr undef) #14
  br i1 false, label %store.throw16, label %store.next17

store.next17:                                     ; preds = %store.next15
  store i32 ptrtoint (ptr @"reflect/types.type:pointer:named:main.prefixSuffixSaver" to i32), ptr %8, align 4
  %.repack48 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 6, i32 1
  store ptr %complit, ptr %.repack48, align 4
  br label %if.done2

if.done2:                                         ; preds = %store.next17, %deref.next11
  %9 = call %runtime._interface @"(*main.Cmd).Run"(ptr nonnull %c, ptr undef)
  %10 = extractvalue %runtime._interface %9, 1
  call void @runtime.trackPointer(ptr %10, ptr undef) #14
  %11 = extractvalue %runtime._interface %9, 0
  %.not39 = icmp eq i32 %11, 0
  br i1 %.not39, label %if.done5, label %cond.true

cond.true:                                        ; preds = %if.done2
  br i1 %7, label %if.then3, label %if.done5

if.then3:                                         ; preds = %cond.true
  %interface.type = extractvalue %runtime._interface %9, 0
  %typecode = call i1 @runtime.typeAssert(i32 %interface.type, ptr nonnull @"reflect/types.typeid:pointer:named:main.ExitError", ptr undef) #14
  br i1 %typecode, label %typeassert.ok, label %typeassert.next

typeassert.next:                                  ; preds = %typeassert.ok, %if.then3
  %typeassert.value = phi ptr [ null, %if.then3 ], [ %typeassert.value.ptr, %typeassert.ok ]
  br i1 %typecode, label %if.then4, label %if.done5

typeassert.ok:                                    ; preds = %if.then3
  %typeassert.value.ptr = extractvalue %runtime._interface %9, 1
  br label %typeassert.next

if.then4:                                         ; preds = %typeassert.next
  %12 = icmp eq ptr %typeassert.value, null
  br i1 %12, label %gep.throw18, label %gep.next19

gep.next19:                                       ; preds = %if.then4
  %13 = getelementptr inbounds %main.ExitError, ptr %typeassert.value, i32 0, i32 1
  br i1 false, label %gep.throw20, label %gep.next21

gep.next21:                                       ; preds = %gep.next19
  br i1 false, label %deref.throw22, label %deref.next23

deref.next23:                                     ; preds = %gep.next21
  %14 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 6
  %.unpack40 = load i32, ptr %14, align 4
  %.elt41 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 6, i32 1
  %.unpack42 = load ptr, ptr %.elt41, align 4
  call void @runtime.trackPointer(ptr %.unpack42, ptr undef) #14
  %typecode25 = call i1 @runtime.typeAssert(i32 %.unpack40, ptr nonnull @"reflect/types.typeid:pointer:named:main.prefixSuffixSaver", ptr undef) #14
  br i1 %typecode25, label %typeassert.ok26, label %typeassert.next27

typeassert.next27:                                ; preds = %typeassert.ok26, %deref.next23
  %typeassert.value29 = phi ptr [ null, %deref.next23 ], [ %.unpack42, %typeassert.ok26 ]
  call void @runtime.interfaceTypeAssert(i1 %typecode25, ptr undef) #14
  %15 = call { ptr, i32, i32 } @"(*main.prefixSuffixSaver).Bytes"(ptr %typeassert.value29, ptr undef)
  %16 = extractvalue { ptr, i32, i32 } %15, 0
  call void @runtime.trackPointer(ptr %16, ptr undef) #14
  br i1 false, label %store.throw30, label %store.next31

store.next31:                                     ; preds = %typeassert.next27
  %.elt = extractvalue { ptr, i32, i32 } %15, 0
  store ptr %.elt, ptr %13, align 4
  %.repack43 = getelementptr inbounds %main.ExitError, ptr %typeassert.value, i32 0, i32 1, i32 1
  %.elt44 = extractvalue { ptr, i32, i32 } %15, 1
  store i32 %.elt44, ptr %.repack43, align 4
  %.repack45 = getelementptr inbounds %main.ExitError, ptr %typeassert.value, i32 0, i32 1, i32 2
  %.elt46 = extractvalue { ptr, i32, i32 } %15, 2
  store i32 %.elt46, ptr %.repack45, align 4
  br label %if.done5

typeassert.ok26:                                  ; preds = %deref.next23
  br label %typeassert.next27

if.done5:                                         ; preds = %store.next31, %typeassert.next, %cond.true, %if.done2
  %17 = call { ptr, i32, i32 } @"(*bytes.Buffer).Bytes"(ptr nonnull %stdout, ptr undef) #14
  %18 = extractvalue { ptr, i32, i32 } %17, 0
  call void @runtime.trackPointer(ptr %18, ptr undef) #14
  %19 = insertvalue { { ptr, i32, i32 }, %runtime._interface } zeroinitializer, { ptr, i32, i32 } %17, 0
  %20 = insertvalue { { ptr, i32, i32 }, %runtime._interface } %19, %runtime._interface %9, 1
  ret { { ptr, i32, i32 }, %runtime._interface } %20

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable

gep.throw6:                                       ; preds = %if.done
  unreachable

store.throw:                                      ; preds = %gep.next7
  unreachable

gep.throw8:                                       ; preds = %store.next
  unreachable

deref.throw10:                                    ; preds = %gep.next9
  unreachable

gep.throw12:                                      ; preds = %if.then1
  unreachable

store.throw14:                                    ; preds = %gep.next13
  unreachable

store.throw16:                                    ; preds = %store.next15
  unreachable

gep.throw18:                                      ; preds = %if.then4
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

gep.throw20:                                      ; preds = %gep.next19
  unreachable

deref.throw22:                                    ; preds = %gep.next21
  unreachable

store.throw30:                                    ; preds = %typeassert.next27
  unreachable
}

; Function Attrs: nounwind
define hidden { ptr, i32, i32 } @"(*main.prefixSuffixSaver).Bytes"(ptr dereferenceable_or_null(40) %w, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %w, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 2
  %.unpack = load ptr, ptr %1, align 4
  call void @runtime.trackPointer(ptr %.unpack, ptr undef) #14
  %2 = icmp eq ptr %.unpack, null
  br i1 %2, label %if.then, label %if.done

if.then:                                          ; preds = %deref.next
  %3 = icmp eq ptr %w, null
  br i1 %3, label %gep.throw3, label %gep.next4

gep.next4:                                        ; preds = %if.then
  br i1 false, label %deref.throw5, label %deref.next6

deref.next6:                                      ; preds = %gep.next4
  %4 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 1
  %.unpack98 = load ptr, ptr %4, align 4
  %5 = insertvalue { ptr, i32, i32 } undef, ptr %.unpack98, 0
  %.elt99 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 1, i32 1
  %.unpack100 = load i32, ptr %.elt99, align 4
  %6 = insertvalue { ptr, i32, i32 } %5, i32 %.unpack100, 1
  %.elt101 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 1, i32 2
  %.unpack102 = load i32, ptr %.elt101, align 4
  %7 = insertvalue { ptr, i32, i32 } %6, i32 %.unpack102, 2
  call void @runtime.trackPointer(ptr %.unpack98, ptr undef) #14
  ret { ptr, i32, i32 } %7

if.done:                                          ; preds = %deref.next
  %8 = icmp eq ptr %w, null
  br i1 %8, label %gep.throw7, label %gep.next8

gep.next8:                                        ; preds = %if.done
  br i1 false, label %deref.throw9, label %deref.next10

deref.next10:                                     ; preds = %gep.next8
  %9 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 4
  %10 = load i64, ptr %9, align 8
  %11 = icmp eq i64 %10, 0
  br i1 %11, label %if.then1, label %if.done2

if.then1:                                         ; preds = %deref.next10
  %12 = icmp eq ptr %w, null
  br i1 %12, label %gep.throw11, label %gep.next12

gep.next12:                                       ; preds = %if.then1
  br i1 false, label %deref.throw13, label %deref.next14

deref.next14:                                     ; preds = %gep.next12
  %13 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 1
  %.unpack88 = load ptr, ptr %13, align 4
  %.elt89 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 1, i32 1
  %.unpack90 = load i32, ptr %.elt89, align 4
  %.elt91 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 1, i32 2
  %.unpack92 = load i32, ptr %.elt91, align 4
  call void @runtime.trackPointer(ptr %.unpack88, ptr undef) #14
  %14 = icmp eq ptr %w, null
  br i1 %14, label %gep.throw15, label %gep.next16

gep.next16:                                       ; preds = %deref.next14
  br i1 false, label %deref.throw17, label %deref.next18

deref.next18:                                     ; preds = %gep.next16
  %15 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 2
  %.unpack93 = load ptr, ptr %15, align 4
  %.elt94 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 2, i32 1
  %.unpack95 = load i32, ptr %.elt94, align 4
  call void @runtime.trackPointer(ptr %.unpack93, ptr undef) #14
  %append.new = call { ptr, i32, i32 } @runtime.sliceAppend(ptr %.unpack88, ptr %.unpack93, i32 %.unpack90, i32 %.unpack92, i32 %.unpack95, i32 1, ptr undef) #14
  %append.newPtr = extractvalue { ptr, i32, i32 } %append.new, 0
  %append.newLen = extractvalue { ptr, i32, i32 } %append.new, 1
  %append.newCap = extractvalue { ptr, i32, i32 } %append.new, 2
  %16 = insertvalue { ptr, i32, i32 } undef, ptr %append.newPtr, 0
  %17 = insertvalue { ptr, i32, i32 } %16, i32 %append.newLen, 1
  %18 = insertvalue { ptr, i32, i32 } %17, i32 %append.newCap, 2
  call void @runtime.trackPointer(ptr %append.newPtr, ptr undef) #14
  ret { ptr, i32, i32 } %18

if.done2:                                         ; preds = %deref.next10
  %buf = call ptr @runtime.alloc(i32 20, ptr nonnull inttoptr (i32 75 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %buf, ptr undef) #14
  %19 = icmp eq ptr %w, null
  br i1 %19, label %gep.throw19, label %gep.next20

gep.next20:                                       ; preds = %if.done2
  br i1 false, label %deref.throw21, label %deref.next22

deref.next22:                                     ; preds = %gep.next20
  %20 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 1
  %.unpack63 = load ptr, ptr %20, align 4
  %.elt64 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 1, i32 1
  %.unpack65 = load i32, ptr %.elt64, align 4
  call void @runtime.trackPointer(ptr %.unpack63, ptr undef) #14
  br i1 false, label %gep.throw23, label %gep.next24

gep.next24:                                       ; preds = %deref.next22
  br i1 false, label %deref.throw25, label %deref.next26

deref.next26:                                     ; preds = %gep.next24
  %21 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 2
  %.unpack68 = load ptr, ptr %21, align 4
  %.elt69 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 2, i32 1
  %.unpack70 = load i32, ptr %.elt69, align 4
  call void @runtime.trackPointer(ptr %.unpack68, ptr undef) #14
  %22 = add i32 %.unpack65, %.unpack70
  %23 = add i32 %22, 50
  call void @"(*bytes.Buffer).Grow"(ptr nonnull %buf, i32 %23, ptr undef) #14
  br i1 false, label %gep.throw28, label %gep.next29

gep.next29:                                       ; preds = %deref.next26
  br i1 false, label %deref.throw30, label %deref.next31

deref.next31:                                     ; preds = %gep.next29
  %24 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 1
  %.unpack73 = load ptr, ptr %24, align 4
  %.elt74 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 1, i32 1
  %.unpack75 = load i32, ptr %.elt74, align 4
  %.elt76 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 1, i32 2
  %.unpack77 = load i32, ptr %.elt76, align 4
  call void @runtime.trackPointer(ptr %.unpack73, ptr undef) #14
  %25 = call { i32, %runtime._interface } @"(*bytes.Buffer).Write"(ptr nonnull %buf, ptr %.unpack73, i32 %.unpack75, i32 %.unpack77, ptr undef) #14
  %26 = extractvalue { i32, %runtime._interface } %25, 1
  %27 = extractvalue %runtime._interface %26, 1
  call void @runtime.trackPointer(ptr %27, ptr undef) #14
  %28 = call { i32, %runtime._interface } @"(*bytes.Buffer).WriteString"(ptr nonnull %buf, ptr nonnull @"main$string.78", i32 14, ptr undef) #14
  %29 = extractvalue { i32, %runtime._interface } %28, 1
  %30 = extractvalue %runtime._interface %29, 1
  call void @runtime.trackPointer(ptr %30, ptr undef) #14
  br i1 false, label %gep.throw32, label %gep.next33

gep.next33:                                       ; preds = %deref.next31
  br i1 false, label %deref.throw34, label %deref.next35

deref.next35:                                     ; preds = %gep.next33
  %31 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 4
  %32 = load i64, ptr %31, align 8
  %33 = call %runtime._string @strconv.FormatInt(i64 %32, i32 10, ptr undef) #14
  %34 = extractvalue %runtime._string %33, 0
  call void @runtime.trackPointer(ptr %34, ptr undef) #14
  %35 = extractvalue %runtime._string %33, 0
  %36 = extractvalue %runtime._string %33, 1
  %37 = call { i32, %runtime._interface } @"(*bytes.Buffer).WriteString"(ptr nonnull %buf, ptr %35, i32 %36, ptr undef) #14
  %38 = extractvalue { i32, %runtime._interface } %37, 1
  %39 = extractvalue %runtime._interface %38, 1
  call void @runtime.trackPointer(ptr %39, ptr undef) #14
  %40 = call { i32, %runtime._interface } @"(*bytes.Buffer).WriteString"(ptr nonnull %buf, ptr nonnull @"main$string.79", i32 11, ptr undef) #14
  %41 = extractvalue { i32, %runtime._interface } %40, 1
  %42 = extractvalue %runtime._interface %41, 1
  call void @runtime.trackPointer(ptr %42, ptr undef) #14
  br i1 false, label %gep.throw36, label %gep.next37

gep.next37:                                       ; preds = %deref.next35
  br i1 false, label %deref.throw38, label %deref.next39

deref.next39:                                     ; preds = %gep.next37
  %43 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 2
  %.unpack78 = load ptr, ptr %43, align 4
  %.elt79 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 2, i32 1
  %.unpack80 = load i32, ptr %.elt79, align 4
  %.elt81 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 2, i32 2
  %.unpack82 = load i32, ptr %.elt81, align 4
  call void @runtime.trackPointer(ptr %.unpack78, ptr undef) #14
  br i1 false, label %gep.throw40, label %gep.next41

gep.next41:                                       ; preds = %deref.next39
  br i1 false, label %deref.throw42, label %deref.next43

deref.next43:                                     ; preds = %gep.next41
  %44 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 3
  %45 = load i32, ptr %44, align 4
  %slice.lowhigh = icmp ult i32 %.unpack80, %45
  %slice.highmax = icmp ugt i32 %.unpack80, %.unpack82
  %slice.lowmax = or i1 %slice.lowhigh, %slice.highmax
  br i1 %slice.lowmax, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %deref.next43
  %46 = getelementptr inbounds i8, ptr %.unpack78, i32 %45
  %47 = sub i32 %.unpack80, %45
  %48 = sub i32 %.unpack82, %45
  %49 = call { i32, %runtime._interface } @"(*bytes.Buffer).Write"(ptr nonnull %buf, ptr %46, i32 %47, i32 %48, ptr undef) #14
  %50 = extractvalue { i32, %runtime._interface } %49, 1
  %51 = extractvalue %runtime._interface %50, 1
  call void @runtime.trackPointer(ptr %51, ptr undef) #14
  br i1 false, label %gep.throw44, label %gep.next45

gep.next45:                                       ; preds = %slice.next
  br i1 false, label %deref.throw46, label %deref.next47

deref.next47:                                     ; preds = %gep.next45
  %52 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 2
  %.unpack83 = load ptr, ptr %52, align 4
  %.elt86 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 2, i32 2
  %.unpack87 = load i32, ptr %.elt86, align 4
  call void @runtime.trackPointer(ptr %.unpack83, ptr undef) #14
  br i1 false, label %gep.throw48, label %gep.next49

gep.next49:                                       ; preds = %deref.next47
  br i1 false, label %deref.throw50, label %deref.next51

deref.next51:                                     ; preds = %gep.next49
  %53 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 3
  %54 = load i32, ptr %53, align 4
  %slice.highmax53 = icmp ugt i32 %54, %.unpack87
  br i1 %slice.highmax53, label %slice.throw57, label %slice.next58

slice.next58:                                     ; preds = %deref.next51
  %55 = call { i32, %runtime._interface } @"(*bytes.Buffer).Write"(ptr nonnull %buf, ptr %.unpack83, i32 %54, i32 %.unpack87, ptr undef) #14
  %56 = extractvalue { i32, %runtime._interface } %55, 1
  %57 = extractvalue %runtime._interface %56, 1
  call void @runtime.trackPointer(ptr %57, ptr undef) #14
  %58 = call { ptr, i32, i32 } @"(*bytes.Buffer).Bytes"(ptr nonnull %buf, ptr undef) #14
  %59 = extractvalue { ptr, i32, i32 } %58, 0
  call void @runtime.trackPointer(ptr %59, ptr undef) #14
  ret { ptr, i32, i32 } %58

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable

gep.throw3:                                       ; preds = %if.then
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw5:                                     ; preds = %gep.next4
  unreachable

gep.throw7:                                       ; preds = %if.done
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw9:                                     ; preds = %gep.next8
  unreachable

gep.throw11:                                      ; preds = %if.then1
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw13:                                    ; preds = %gep.next12
  unreachable

gep.throw15:                                      ; preds = %deref.next14
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw17:                                    ; preds = %gep.next16
  unreachable

gep.throw19:                                      ; preds = %if.done2
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw21:                                    ; preds = %gep.next20
  unreachable

gep.throw23:                                      ; preds = %deref.next22
  unreachable

deref.throw25:                                    ; preds = %gep.next24
  unreachable

gep.throw28:                                      ; preds = %deref.next26
  unreachable

deref.throw30:                                    ; preds = %gep.next29
  unreachable

gep.throw32:                                      ; preds = %deref.next31
  unreachable

deref.throw34:                                    ; preds = %gep.next33
  unreachable

gep.throw36:                                      ; preds = %deref.next35
  unreachable

deref.throw38:                                    ; preds = %gep.next37
  unreachable

gep.throw40:                                      ; preds = %deref.next39
  unreachable

deref.throw42:                                    ; preds = %gep.next41
  unreachable

slice.throw:                                      ; preds = %deref.next43
  call void @runtime.slicePanic(ptr undef) #14
  unreachable

gep.throw44:                                      ; preds = %slice.next
  unreachable

deref.throw46:                                    ; preds = %gep.next45
  unreachable

gep.throw48:                                      ; preds = %deref.next47
  unreachable

deref.throw50:                                    ; preds = %gep.next49
  unreachable

slice.throw57:                                    ; preds = %deref.next51
  call void @runtime.slicePanic(ptr undef) #14
  unreachable
}

; Function Attrs: nounwind
define hidden { i32, %runtime._interface } @"(*main.prefixSuffixSaver).Write"(ptr dereferenceable_or_null(40) %w, ptr %p.data, i32 %p.len, i32 %p.cap, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %w, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  %1 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 1
  %2 = call { ptr, i32, i32 } @"(*main.prefixSuffixSaver).fill"(ptr nonnull %w, ptr nonnull %1, ptr %p.data, i32 %p.len, i32 %p.cap, ptr undef)
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %len2 = extractvalue { ptr, i32, i32 } %2, 1
  br i1 false, label %gep.throw3, label %gep.next4

gep.next4:                                        ; preds = %gep.next
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next4
  %4 = load i32, ptr %w, align 4
  %5 = sub i32 %len2, %4
  %6 = icmp sgt i32 %5, 0
  br i1 %6, label %if.then, label %if.done

if.then:                                          ; preds = %deref.next
  %7 = extractvalue { ptr, i32, i32 } %2, 1
  %8 = extractvalue { ptr, i32, i32 } %2, 2
  %slice.lowhigh = icmp ult i32 %7, %5
  %slice.highmax = icmp ugt i32 %7, %8
  %slice.lowmax = or i1 %slice.lowhigh, %slice.highmax
  br i1 %slice.lowmax, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %if.then
  %9 = extractvalue { ptr, i32, i32 } %2, 0
  %10 = getelementptr inbounds i8, ptr %9, i32 %5
  %11 = sub i32 %7, %5
  %12 = sub i32 %8, %5
  %13 = insertvalue { ptr, i32, i32 } undef, ptr %10, 0
  %14 = insertvalue { ptr, i32, i32 } %13, i32 %11, 1
  %15 = insertvalue { ptr, i32, i32 } %14, i32 %12, 2
  br i1 false, label %gep.throw5, label %gep.next6

gep.next6:                                        ; preds = %slice.next
  %16 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 4
  br i1 false, label %deref.throw7, label %deref.next8

deref.next8:                                      ; preds = %gep.next6
  br i1 false, label %store.throw, label %store.next

store.next:                                       ; preds = %deref.next8
  %17 = load i64, ptr %16, align 8
  %18 = sext i32 %5 to i64
  %19 = add i64 %17, %18
  store i64 %19, ptr %16, align 8
  br label %if.done

if.done:                                          ; preds = %store.next, %deref.next
  %20 = phi { ptr, i32, i32 } [ %2, %deref.next ], [ %15, %store.next ]
  %21 = extractvalue { ptr, i32, i32 } %20, 0
  call void @runtime.trackPointer(ptr %21, ptr undef) #14
  br i1 false, label %gep.throw9, label %gep.next10

gep.next10:                                       ; preds = %if.done
  %22 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 2
  %23 = extractvalue { ptr, i32, i32 } %20, 0
  %24 = extractvalue { ptr, i32, i32 } %20, 1
  %25 = extractvalue { ptr, i32, i32 } %20, 2
  %26 = call { ptr, i32, i32 } @"(*main.prefixSuffixSaver).fill"(ptr nonnull %w, ptr nonnull %22, ptr %23, i32 %24, i32 %25, ptr undef)
  %27 = extractvalue { ptr, i32, i32 } %26, 0
  call void @runtime.trackPointer(ptr %27, ptr undef) #14
  br label %for.loop

for.loop:                                         ; preds = %store.next57, %deref.next53, %gep.next10
  %28 = phi { ptr, i32, i32 } [ %26, %gep.next10 ], [ %44, %deref.next53 ], [ %44, %store.next57 ]
  %29 = extractvalue { ptr, i32, i32 } %28, 0
  call void @runtime.trackPointer(ptr %29, ptr undef) #14
  %len11 = extractvalue { ptr, i32, i32 } %28, 1
  %30 = icmp sgt i32 %len11, 0
  br i1 %30, label %for.body, label %for.done

for.body:                                         ; preds = %for.loop
  br i1 false, label %gep.throw12, label %gep.next13

gep.next13:                                       ; preds = %for.body
  br i1 false, label %deref.throw14, label %deref.next15

deref.next15:                                     ; preds = %gep.next13
  %31 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 2
  %.unpack = load ptr, ptr %31, align 4
  %.elt58 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 2, i32 1
  %.unpack59 = load i32, ptr %.elt58, align 4
  %.elt60 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 2, i32 2
  %.unpack61 = load i32, ptr %.elt60, align 4
  call void @runtime.trackPointer(ptr %.unpack, ptr undef) #14
  br i1 false, label %gep.throw16, label %gep.next17

gep.next17:                                       ; preds = %deref.next15
  br i1 false, label %deref.throw18, label %deref.next19

deref.next19:                                     ; preds = %gep.next17
  %32 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 3
  %33 = load i32, ptr %32, align 4
  %slice.lowhigh20 = icmp ult i32 %.unpack59, %33
  %slice.highmax21 = icmp ugt i32 %.unpack59, %.unpack61
  %slice.lowmax23 = or i1 %slice.lowhigh20, %slice.highmax21
  br i1 %slice.lowmax23, label %slice.throw25, label %slice.next26

slice.next26:                                     ; preds = %deref.next19
  %34 = getelementptr inbounds i8, ptr %.unpack, i32 %33
  %35 = sub i32 %.unpack59, %33
  %copy.srcLen = extractvalue { ptr, i32, i32 } %28, 1
  %copy.srcArray = extractvalue { ptr, i32, i32 } %28, 0
  %copy.n = call i32 @runtime.sliceCopy(ptr %34, ptr %copy.srcArray, i32 %35, i32 %copy.srcLen, i32 1, ptr undef) #14
  %36 = extractvalue { ptr, i32, i32 } %28, 1
  %37 = extractvalue { ptr, i32, i32 } %28, 2
  %slice.lowhigh27 = icmp ult i32 %36, %copy.n
  %slice.highmax28 = icmp ugt i32 %36, %37
  %slice.lowmax30 = or i1 %slice.lowhigh27, %slice.highmax28
  br i1 %slice.lowmax30, label %slice.throw32, label %slice.next33

slice.next33:                                     ; preds = %slice.next26
  %38 = extractvalue { ptr, i32, i32 } %28, 0
  %39 = getelementptr inbounds i8, ptr %38, i32 %copy.n
  %40 = sub i32 %36, %copy.n
  %41 = sub i32 %37, %copy.n
  %42 = insertvalue { ptr, i32, i32 } undef, ptr %39, 0
  %43 = insertvalue { ptr, i32, i32 } %42, i32 %40, 1
  %44 = insertvalue { ptr, i32, i32 } %43, i32 %41, 2
  br i1 false, label %gep.throw34, label %gep.next35

gep.next35:                                       ; preds = %slice.next33
  %45 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 4
  br i1 false, label %deref.throw36, label %deref.next37

deref.next37:                                     ; preds = %gep.next35
  br i1 false, label %store.throw38, label %store.next39

store.next39:                                     ; preds = %deref.next37
  %46 = load i64, ptr %45, align 8
  %47 = sext i32 %copy.n to i64
  %48 = add i64 %46, %47
  store i64 %48, ptr %45, align 8
  br i1 false, label %gep.throw40, label %gep.next41

gep.next41:                                       ; preds = %store.next39
  %49 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 3
  br i1 false, label %deref.throw42, label %deref.next43

deref.next43:                                     ; preds = %gep.next41
  br i1 false, label %store.throw44, label %store.next45

store.next45:                                     ; preds = %deref.next43
  %50 = load i32, ptr %49, align 4
  %51 = add i32 %50, %copy.n
  store i32 %51, ptr %49, align 4
  br i1 false, label %gep.throw46, label %gep.next47

gep.next47:                                       ; preds = %store.next45
  br i1 false, label %deref.throw48, label %deref.next49

deref.next49:                                     ; preds = %gep.next47
  %52 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 3
  %53 = load i32, ptr %52, align 4
  br i1 false, label %gep.throw50, label %gep.next51

gep.next51:                                       ; preds = %deref.next49
  br i1 false, label %deref.throw52, label %deref.next53

deref.next53:                                     ; preds = %gep.next51
  %54 = load i32, ptr %w, align 4
  %55 = icmp eq i32 %53, %54
  br i1 %55, label %if.then1, label %for.loop

if.then1:                                         ; preds = %deref.next53
  br i1 false, label %gep.throw54, label %gep.next55

gep.next55:                                       ; preds = %if.then1
  br i1 false, label %store.throw56, label %store.next57

store.next57:                                     ; preds = %gep.next55
  %56 = getelementptr inbounds %main.prefixSuffixSaver, ptr %w, i32 0, i32 3
  store i32 0, ptr %56, align 4
  br label %for.loop

for.done:                                         ; preds = %for.loop
  %57 = insertvalue { i32, %runtime._interface } zeroinitializer, i32 %p.len, 0
  %58 = insertvalue { i32, %runtime._interface } %57, %runtime._interface zeroinitializer, 1
  ret { i32, %runtime._interface } %58

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

gep.throw3:                                       ; preds = %gep.next
  unreachable

deref.throw:                                      ; preds = %gep.next4
  unreachable

slice.throw:                                      ; preds = %if.then
  call void @runtime.slicePanic(ptr undef) #14
  unreachable

gep.throw5:                                       ; preds = %slice.next
  unreachable

deref.throw7:                                     ; preds = %gep.next6
  unreachable

store.throw:                                      ; preds = %deref.next8
  unreachable

gep.throw9:                                       ; preds = %if.done
  unreachable

gep.throw12:                                      ; preds = %for.body
  unreachable

deref.throw14:                                    ; preds = %gep.next13
  unreachable

gep.throw16:                                      ; preds = %deref.next15
  unreachable

deref.throw18:                                    ; preds = %gep.next17
  unreachable

slice.throw25:                                    ; preds = %deref.next19
  call void @runtime.slicePanic(ptr undef) #14
  unreachable

slice.throw32:                                    ; preds = %slice.next26
  call void @runtime.slicePanic(ptr undef) #14
  unreachable

gep.throw34:                                      ; preds = %slice.next33
  unreachable

deref.throw36:                                    ; preds = %gep.next35
  unreachable

store.throw38:                                    ; preds = %deref.next37
  unreachable

gep.throw40:                                      ; preds = %store.next39
  unreachable

deref.throw42:                                    ; preds = %gep.next41
  unreachable

store.throw44:                                    ; preds = %deref.next43
  unreachable

gep.throw46:                                      ; preds = %store.next45
  unreachable

deref.throw48:                                    ; preds = %gep.next47
  unreachable

gep.throw50:                                      ; preds = %deref.next49
  unreachable

deref.throw52:                                    ; preds = %gep.next51
  unreachable

gep.throw54:                                      ; preds = %if.then1
  unreachable

store.throw56:                                    ; preds = %gep.next55
  unreachable
}

; Function Attrs: nounwind
define hidden { ptr, i32, i32 } @"(*main.prefixSuffixSaver).fill"(ptr dereferenceable_or_null(40) %w, ptr dereferenceable_or_null(12) %dst, ptr %p.data, i32 %p.len, i32 %p.cap, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %w, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = icmp eq ptr %dst, null
  br i1 %1, label %deref.throw1, label %deref.next2

deref.next2:                                      ; preds = %deref.next
  %2 = load i32, ptr %w, align 4
  %.unpack = load ptr, ptr %dst, align 4
  %.elt13 = getelementptr inbounds { ptr, i32, i32 }, ptr %dst, i32 0, i32 1
  %.unpack14 = load i32, ptr %.elt13, align 4
  call void @runtime.trackPointer(ptr %.unpack, ptr undef) #14
  %3 = sub i32 %2, %.unpack14
  %4 = icmp sgt i32 %3, 0
  br i1 %4, label %if.then, label %if.done

if.then:                                          ; preds = %deref.next2
  %5 = call i32 @main.minInt(i32 %p.len, i32 %3, ptr undef)
  br i1 false, label %deref.throw4, label %deref.next5

deref.next5:                                      ; preds = %if.then
  %.unpack17 = load ptr, ptr %dst, align 4
  %.elt18 = getelementptr inbounds { ptr, i32, i32 }, ptr %dst, i32 0, i32 1
  %.unpack19 = load i32, ptr %.elt18, align 4
  %.elt20 = getelementptr inbounds { ptr, i32, i32 }, ptr %dst, i32 0, i32 2
  %.unpack21 = load i32, ptr %.elt20, align 4
  call void @runtime.trackPointer(ptr %.unpack17, ptr undef) #14
  %slice.highmax = icmp ugt i32 %5, %p.cap
  br i1 %slice.highmax, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %deref.next5
  %append.new = call { ptr, i32, i32 } @runtime.sliceAppend(ptr %.unpack17, ptr %p.data, i32 %.unpack19, i32 %.unpack21, i32 %5, i32 1, ptr undef) #14
  %append.newPtr = extractvalue { ptr, i32, i32 } %append.new, 0
  call void @runtime.trackPointer(ptr %append.newPtr, ptr undef) #14
  br i1 false, label %store.throw, label %store.next

store.next:                                       ; preds = %slice.next
  %append.newLen = extractvalue { ptr, i32, i32 } %append.new, 1
  %append.newCap = extractvalue { ptr, i32, i32 } %append.new, 2
  store ptr %append.newPtr, ptr %dst, align 4
  %dst.repack22 = getelementptr inbounds { ptr, i32, i32 }, ptr %dst, i32 0, i32 1
  store i32 %append.newLen, ptr %dst.repack22, align 4
  %dst.repack24 = getelementptr inbounds { ptr, i32, i32 }, ptr %dst, i32 0, i32 2
  store i32 %append.newCap, ptr %dst.repack24, align 4
  %slice.lowhigh6 = icmp ugt i32 %5, %p.len
  %slice.highmax7 = icmp ugt i32 %p.len, %p.cap
  %slice.lowmax9 = or i1 %slice.lowhigh6, %slice.highmax7
  br i1 %slice.lowmax9, label %slice.throw11, label %slice.next12

slice.next12:                                     ; preds = %store.next
  %6 = getelementptr inbounds i8, ptr %p.data, i32 %5
  %7 = sub i32 %p.len, %5
  %8 = sub i32 %p.cap, %5
  br label %if.done

if.done:                                          ; preds = %slice.next12, %deref.next2
  %p.data.pn = phi ptr [ %p.data, %deref.next2 ], [ %6, %slice.next12 ]
  %p.len.pn = phi i32 [ %p.len, %deref.next2 ], [ %7, %slice.next12 ]
  %p.cap.pn = phi i32 [ %p.cap, %deref.next2 ], [ %8, %slice.next12 ]
  %.pn26 = insertvalue { ptr, i32, i32 } zeroinitializer, ptr %p.data.pn, 0
  %.pn = insertvalue { ptr, i32, i32 } %.pn26, i32 %p.len.pn, 1
  %9 = insertvalue { ptr, i32, i32 } %.pn, i32 %p.cap.pn, 2
  call void @runtime.trackPointer(ptr %p.data.pn, ptr undef) #14
  ret { ptr, i32, i32 } %9

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable

deref.throw1:                                     ; preds = %deref.next
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw4:                                     ; preds = %if.then
  unreachable

slice.throw:                                      ; preds = %deref.next5
  call void @runtime.slicePanic(ptr undef) #14
  unreachable

store.throw:                                      ; preds = %slice.next
  unreachable

slice.throw11:                                    ; preds = %store.next
  call void @runtime.slicePanic(ptr undef) #14
  unreachable
}

declare i1 @runtime.typeAssert(i32, ptr dereferenceable_or_null(1), ptr) #0

declare void @runtime.interfaceTypeAssert(i1, ptr) #0

; Function Attrs: nounwind
define hidden %runtime._interface @"(*main.Cmd).Start"(ptr dereferenceable_or_null(168) %c, ptr %context) unnamed_addr #1 {
entry:
  %select.states.alloca = alloca [1 x %runtime.chanSelectState], align 8
  %0 = icmp eq ptr %c, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %.unpack = load ptr, ptr %c, align 4
  %.elt222 = getelementptr inbounds %runtime._string, ptr %c, i32 0, i32 1
  %.unpack223 = load i32, ptr %.elt222, align 4
  call void @runtime.trackPointer(ptr %.unpack, ptr undef) #14
  %1 = call i1 @runtime.stringEqual(ptr %.unpack, i32 %.unpack223, ptr null, i32 0, ptr undef) #14
  br i1 %1, label %cond.true, label %if.done

cond.true:                                        ; preds = %deref.next
  %2 = icmp eq ptr %c, null
  br i1 %2, label %gep.throw24, label %gep.next25

gep.next25:                                       ; preds = %cond.true
  br i1 false, label %deref.throw26, label %deref.next27

deref.next27:                                     ; preds = %gep.next25
  %3 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 12
  %.unpack406 = load i32, ptr %3, align 4
  %.elt407 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 12, i32 1
  %.unpack408 = load ptr, ptr %.elt407, align 4
  call void @runtime.trackPointer(ptr %.unpack408, ptr undef) #14
  %4 = icmp eq i32 %.unpack406, 0
  br i1 %4, label %cond.true1, label %if.done

cond.true1:                                       ; preds = %deref.next27
  %5 = icmp eq ptr %c, null
  br i1 %5, label %gep.throw28, label %gep.next29

gep.next29:                                       ; preds = %cond.true1
  br i1 false, label %deref.throw30, label %deref.next31

deref.next31:                                     ; preds = %gep.next29
  %6 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 19
  %.unpack410 = load i32, ptr %6, align 4
  %.elt411 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 19, i32 1
  %.unpack412 = load ptr, ptr %.elt411, align 4
  call void @runtime.trackPointer(ptr %.unpack412, ptr undef) #14
  %7 = icmp eq i32 %.unpack410, 0
  br i1 %7, label %if.then, label %if.done

if.then:                                          ; preds = %deref.next31
  %8 = icmp eq ptr %c, null
  br i1 %8, label %gep.throw32, label %gep.next33

gep.next33:                                       ; preds = %if.then
  %9 = call %runtime._interface @errors.New(ptr nonnull @"main$string.15", i32 16, ptr undef) #14
  %10 = extractvalue %runtime._interface %9, 1
  call void @runtime.trackPointer(ptr %10, ptr undef) #14
  br i1 false, label %store.throw, label %store.next

store.next:                                       ; preds = %gep.next33
  %11 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 12
  %.elt413 = extractvalue %runtime._interface %9, 0
  store i32 %.elt413, ptr %11, align 4
  %.repack414 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 12, i32 1
  %.elt415 = extractvalue %runtime._interface %9, 1
  store ptr %.elt415, ptr %.repack414, align 4
  br label %if.done

if.done:                                          ; preds = %store.next, %deref.next31, %deref.next27, %deref.next
  br i1 false, label %gep.throw34, label %gep.next35

gep.next35:                                       ; preds = %if.done
  br i1 false, label %deref.throw36, label %deref.next37

deref.next37:                                     ; preds = %gep.next35
  %12 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 12
  %.unpack224 = load i32, ptr %12, align 4
  %.elt225 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 12, i32 1
  %.unpack226 = load ptr, ptr %.elt225, align 4
  call void @runtime.trackPointer(ptr %.unpack226, ptr undef) #14
  %.not = icmp eq i32 %.unpack224, 0
  br i1 %.not, label %cond.false, label %if.then2

if.then2:                                         ; preds = %deref.next61, %deref.next37
  %13 = icmp eq ptr %c, null
  br i1 %13, label %gep.throw38, label %gep.next39

gep.next39:                                       ; preds = %if.then2
  br i1 false, label %deref.throw40, label %deref.next41

deref.next41:                                     ; preds = %gep.next39
  %14 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14
  %.unpack381 = load ptr, ptr %14, align 4
  %.elt382 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 1
  %.unpack383 = load i32, ptr %.elt382, align 4
  %.elt384 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 2
  %.unpack385 = load i32, ptr %.elt384, align 4
  call void @runtime.trackPointer(ptr %.unpack381, ptr undef) #14
  call void @"(*main.Cmd).closeDescriptors"(ptr %c, ptr %.unpack381, i32 %.unpack383, i32 %.unpack385, ptr undef)
  %15 = icmp eq ptr %c, null
  br i1 %15, label %gep.throw42, label %gep.next43

gep.next43:                                       ; preds = %deref.next41
  br i1 false, label %deref.throw44, label %deref.next45

deref.next45:                                     ; preds = %gep.next43
  %16 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15
  %.unpack387 = load ptr, ptr %16, align 4
  %.elt388 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 1
  %.unpack389 = load i32, ptr %.elt388, align 4
  %.elt390 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 2
  %.unpack391 = load i32, ptr %.elt390, align 4
  call void @runtime.trackPointer(ptr %.unpack387, ptr undef) #14
  call void @"(*main.Cmd).closeDescriptors"(ptr %c, ptr %.unpack387, i32 %.unpack389, i32 %.unpack391, ptr undef)
  %17 = icmp eq ptr %c, null
  br i1 %17, label %gep.throw46, label %gep.next47

gep.next47:                                       ; preds = %deref.next45
  br i1 false, label %deref.throw48, label %deref.next49

deref.next49:                                     ; preds = %gep.next47
  %18 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 19
  %.unpack393 = load i32, ptr %18, align 4
  %.elt394 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 19, i32 1
  %.unpack395 = load ptr, ptr %.elt394, align 4
  call void @runtime.trackPointer(ptr %.unpack395, ptr undef) #14
  %.not396 = icmp eq i32 %.unpack393, 0
  br i1 %.not396, label %if.done4, label %if.then3

if.then3:                                         ; preds = %deref.next49
  %19 = icmp eq ptr %c, null
  br i1 %19, label %gep.throw50, label %gep.next51

gep.next51:                                       ; preds = %if.then3
  br i1 false, label %deref.throw52, label %deref.next53

deref.next53:                                     ; preds = %gep.next51
  %20 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 19
  %.unpack402 = load i32, ptr %20, align 4
  %21 = insertvalue %runtime._interface undef, i32 %.unpack402, 0
  %.elt403 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 19, i32 1
  %.unpack404 = load ptr, ptr %.elt403, align 4
  %22 = insertvalue %runtime._interface %21, ptr %.unpack404, 1
  call void @runtime.trackPointer(ptr %.unpack404, ptr undef) #14
  ret %runtime._interface %22

if.done4:                                         ; preds = %deref.next49
  %23 = icmp eq ptr %c, null
  br i1 %23, label %gep.throw54, label %gep.next55

gep.next55:                                       ; preds = %if.done4
  br i1 false, label %deref.throw56, label %deref.next57

deref.next57:                                     ; preds = %gep.next55
  %24 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 12
  %.unpack398 = load i32, ptr %24, align 4
  %25 = insertvalue %runtime._interface undef, i32 %.unpack398, 0
  %.elt399 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 12, i32 1
  %.unpack400 = load ptr, ptr %.elt399, align 4
  %26 = insertvalue %runtime._interface %25, ptr %.unpack400, 1
  call void @runtime.trackPointer(ptr %.unpack400, ptr undef) #14
  ret %runtime._interface %26

cond.false:                                       ; preds = %deref.next37
  br i1 false, label %gep.throw58, label %gep.next59

gep.next59:                                       ; preds = %cond.false
  br i1 false, label %deref.throw60, label %deref.next61

deref.next61:                                     ; preds = %gep.next59
  %27 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 19
  %.unpack227 = load i32, ptr %27, align 4
  %.elt228 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 19, i32 1
  %.unpack229 = load ptr, ptr %.elt228, align 4
  call void @runtime.trackPointer(ptr %.unpack229, ptr undef) #14
  %.not230 = icmp eq i32 %.unpack227, 0
  br i1 %.not230, label %if.done5, label %if.then2

if.done5:                                         ; preds = %deref.next61
  br i1 false, label %if.then6, label %if.done9

if.then6:                                         ; preds = %if.done5
  br i1 poison, label %gep.throw62, label %gep.next63

gep.next63:                                       ; preds = %if.then6
  br i1 poison, label %deref.throw64, label %deref.next65

deref.next65:                                     ; preds = %gep.next63
  br i1 poison, label %gep.throw66, label %gep.next67

gep.next67:                                       ; preds = %deref.next65
  br i1 poison, label %deref.throw68, label %deref.next69

deref.next69:                                     ; preds = %gep.next67
  br i1 poison, label %if.then7, label %if.done8

if.then7:                                         ; preds = %deref.next69
  br i1 poison, label %gep.throw70, label %gep.next71

gep.next71:                                       ; preds = %if.then7
  br i1 poison, label %deref.throw72, label %deref.next73

deref.next73:                                     ; preds = %gep.next71
  br i1 poison, label %gep.throw74, label %gep.next75

gep.next75:                                       ; preds = %deref.next73
  br i1 poison, label %deref.throw76, label %deref.next77

deref.next77:                                     ; preds = %gep.next75
  ret %runtime._interface poison

if.done8:                                         ; preds = %deref.next69
  br i1 poison, label %gep.throw78, label %gep.next79

gep.next79:                                       ; preds = %if.done8
  br i1 poison, label %store.throw80, label %store.next81

store.next81:                                     ; preds = %gep.next79
  br label %if.done9

if.done9:                                         ; preds = %store.next81, %if.done5
  br i1 false, label %gep.throw82, label %gep.next83

gep.next83:                                       ; preds = %if.done9
  br i1 false, label %deref.throw84, label %deref.next85

deref.next85:                                     ; preds = %gep.next83
  %28 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 9
  %29 = load ptr, ptr %28, align 4
  call void @runtime.trackPointer(ptr %29, ptr undef) #14
  %.not231 = icmp eq ptr %29, null
  br i1 %.not231, label %if.done11, label %if.then10

if.then10:                                        ; preds = %deref.next85
  %30 = call %runtime._interface @errors.New(ptr nonnull @"main$string.16", i32 21, ptr undef) #14
  %31 = extractvalue %runtime._interface %30, 1
  call void @runtime.trackPointer(ptr %31, ptr undef) #14
  ret %runtime._interface %30

if.done11:                                        ; preds = %deref.next85
  br i1 false, label %gep.throw86, label %gep.next87

gep.next87:                                       ; preds = %if.done11
  br i1 false, label %deref.throw88, label %deref.next89

deref.next89:                                     ; preds = %gep.next87
  %32 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 11
  %.unpack232 = load i32, ptr %32, align 4
  %.elt233 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 11, i32 1
  %.unpack234 = load ptr, ptr %.elt233, align 4
  call void @runtime.trackPointer(ptr %.unpack234, ptr undef) #14
  %.not235 = icmp eq i32 %.unpack232, 0
  br i1 %.not235, label %if.done13, label %if.then12

if.then12:                                        ; preds = %deref.next89
  %33 = icmp eq ptr %c, null
  br i1 %33, label %gep.throw90, label %gep.next91

gep.next91:                                       ; preds = %if.then12
  br i1 false, label %deref.throw92, label %deref.next93

deref.next93:                                     ; preds = %gep.next91
  %34 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 11
  %.unpack358 = load i32, ptr %34, align 4
  %.elt359 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 11, i32 1
  %.unpack360 = load ptr, ptr %.elt359, align 4
  call void @runtime.trackPointer(ptr %.unpack360, ptr undef) #14
  %35 = call ptr @"interface:{Deadline:func:{}{named:time.Time,basic:bool},Done:func:{}{chan:struct:{}},Err:func:{}{named:error},Value:func:{interface:{}}{interface:{}}}.Done$invoke"(ptr %.unpack360, i32 %.unpack358, ptr undef) #14
  call void @runtime.trackPointer(ptr %35, ptr undef) #14
  call void @llvm.lifetime.start.p0(i64 8, ptr nonnull %select.states.alloca)
  store ptr %35, ptr %select.states.alloca, align 8
  %select.states.alloca.repack362 = getelementptr inbounds %runtime.chanSelectState, ptr %select.states.alloca, i32 0, i32 1
  store ptr null, ptr %select.states.alloca.repack362, align 4
  %select.result = call { i32, i1 } @runtime.tryChanSelect(ptr undef, ptr nonnull %select.states.alloca, i32 1, i32 1, ptr undef) #14
  call void @llvm.lifetime.end.p0(i64 8, ptr nonnull %select.states.alloca)
  %36 = extractvalue { i32, i1 } %select.result, 0
  %37 = icmp eq i32 %36, 0
  br i1 %37, label %select.body, label %if.done13

select.body:                                      ; preds = %deref.next93
  %38 = icmp eq ptr %c, null
  br i1 %38, label %gep.throw94, label %gep.next95

gep.next95:                                       ; preds = %select.body
  br i1 false, label %deref.throw96, label %deref.next97

deref.next97:                                     ; preds = %gep.next95
  %39 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14
  %.unpack365 = load ptr, ptr %39, align 4
  %.elt366 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 1
  %.unpack367 = load i32, ptr %.elt366, align 4
  %.elt368 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 2
  %.unpack369 = load i32, ptr %.elt368, align 4
  call void @runtime.trackPointer(ptr %.unpack365, ptr undef) #14
  call void @"(*main.Cmd).closeDescriptors"(ptr %c, ptr %.unpack365, i32 %.unpack367, i32 %.unpack369, ptr undef)
  %40 = icmp eq ptr %c, null
  br i1 %40, label %gep.throw98, label %gep.next99

gep.next99:                                       ; preds = %deref.next97
  br i1 false, label %deref.throw100, label %deref.next101

deref.next101:                                    ; preds = %gep.next99
  %41 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15
  %.unpack371 = load ptr, ptr %41, align 4
  %.elt372 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 1
  %.unpack373 = load i32, ptr %.elt372, align 4
  %.elt374 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 2
  %.unpack375 = load i32, ptr %.elt374, align 4
  call void @runtime.trackPointer(ptr %.unpack371, ptr undef) #14
  call void @"(*main.Cmd).closeDescriptors"(ptr %c, ptr %.unpack371, i32 %.unpack373, i32 %.unpack375, ptr undef)
  %42 = icmp eq ptr %c, null
  br i1 %42, label %gep.throw102, label %gep.next103

gep.next103:                                      ; preds = %deref.next101
  br i1 false, label %deref.throw104, label %deref.next105

deref.next105:                                    ; preds = %gep.next103
  %43 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 11
  %.unpack377 = load i32, ptr %43, align 4
  %.elt378 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 11, i32 1
  %.unpack379 = load ptr, ptr %.elt378, align 4
  call void @runtime.trackPointer(ptr %.unpack379, ptr undef) #14
  %44 = call %runtime._interface @"interface:{Deadline:func:{}{named:time.Time,basic:bool},Done:func:{}{chan:struct:{}},Err:func:{}{named:error},Value:func:{interface:{}}{interface:{}}}.Err$invoke"(ptr %.unpack379, i32 %.unpack377, ptr undef) #14
  %45 = extractvalue %runtime._interface %44, 1
  call void @runtime.trackPointer(ptr %45, ptr undef) #14
  ret %runtime._interface %44

if.done13:                                        ; preds = %deref.next93, %deref.next89
  br i1 false, label %gep.throw108, label %gep.next109

gep.next109:                                      ; preds = %if.done13
  %46 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 13
  br i1 false, label %gep.throw110, label %gep.next111

gep.next111:                                      ; preds = %gep.next109
  br i1 false, label %deref.throw112, label %deref.next113

deref.next113:                                    ; preds = %gep.next111
  %47 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 7
  %.unpack236 = load ptr, ptr %47, align 4
  %.elt237 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 7, i32 1
  %.unpack238 = load i32, ptr %.elt237, align 4
  call void @runtime.trackPointer(ptr %.unpack236, ptr undef) #14
  %48 = add i32 %.unpack238, 3
  %slice.maxcap = icmp ugt i32 %48, 1073741823
  br i1 %slice.maxcap, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %deref.next113
  %makeslice.cap = shl i32 %48, 2
  %makeslice.buf = call ptr @runtime.alloc(i32 %makeslice.cap, ptr nonnull inttoptr (i32 67 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %makeslice.buf, ptr undef) #14
  br i1 false, label %store.throw114, label %store.next115

store.next115:                                    ; preds = %slice.next
  store ptr %makeslice.buf, ptr %46, align 4
  %.repack241 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 13, i32 1
  store i32 0, ptr %.repack241, align 4
  %.repack243 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 13, i32 2
  store i32 %48, ptr %.repack243, align 4
  %slicelit = call ptr @runtime.alloc(i32 24, ptr nonnull inttoptr (i32 197 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %slicelit, ptr undef) #14
  %slicelit.repack245 = getelementptr inbounds { ptr, ptr }, ptr %slicelit, i32 0, i32 1
  store ptr @"(*main.Cmd).stdin$thunk$thunk for func (*main.Cmd).stdin() (f *os.File, err error)", ptr %slicelit.repack245, align 4
  %.repack246 = getelementptr inbounds [3 x { ptr, ptr }], ptr %slicelit, i32 0, i32 1, i32 1
  store ptr @"(*main.Cmd).stdout$thunk$thunk for func (*main.Cmd).stdout() (f *os.File, err error)", ptr %.repack246, align 4
  %.repack247 = getelementptr inbounds [3 x { ptr, ptr }], ptr %slicelit, i32 0, i32 2, i32 1
  store ptr @"(*main.Cmd).stderr$thunk$thunk for func (*main.Cmd).stderr() (f *os.File, err error)", ptr %.repack247, align 4
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %store.next133, %store.next115
  %49 = phi i32 [ -1, %store.next115 ], [ %50, %store.next133 ]
  %50 = add i32 %49, 1
  %51 = icmp slt i32 %50, 3
  br i1 %51, label %rangeindex.body, label %rangeindex.done

rangeindex.body:                                  ; preds = %rangeindex.loop
  %52 = icmp ugt i32 %50, 2
  br i1 %52, label %lookup.throw, label %lookup.next

lookup.next:                                      ; preds = %rangeindex.body
  %53 = getelementptr inbounds { ptr, ptr }, ptr %slicelit, i32 %50
  %.unpack330 = load ptr, ptr %53, align 4
  %.elt331 = getelementptr inbounds { ptr, ptr }, ptr %slicelit, i32 %50, i32 1
  %.unpack332 = load ptr, ptr %.elt331, align 4
  call void @runtime.trackPointer(ptr %.unpack330, ptr undef) #14
  call void @runtime.trackPointer(ptr %.unpack332, ptr undef) #14
  %54 = icmp eq ptr %.unpack332, null
  br i1 %54, label %fpcall.throw, label %fpcall.next

fpcall.next:                                      ; preds = %lookup.next
  %55 = call { ptr, %runtime._interface } %.unpack332(ptr %c, ptr %.unpack330) #14
  %56 = extractvalue { ptr, %runtime._interface } %55, 0
  call void @runtime.trackPointer(ptr %56, ptr undef) #14
  %57 = extractvalue { ptr, %runtime._interface } %55, 1
  %58 = extractvalue %runtime._interface %57, 1
  call void @runtime.trackPointer(ptr %58, ptr undef) #14
  %59 = extractvalue { ptr, %runtime._interface } %55, 0
  %60 = extractvalue { ptr, %runtime._interface } %55, 1
  %61 = extractvalue %runtime._interface %60, 0
  %.not333 = icmp eq i32 %61, 0
  br i1 %.not333, label %if.done15, label %if.then14

if.then14:                                        ; preds = %fpcall.next
  %62 = icmp eq ptr %c, null
  br i1 %62, label %gep.throw117, label %gep.next118

gep.next118:                                      ; preds = %if.then14
  br i1 false, label %deref.throw119, label %deref.next120

deref.next120:                                    ; preds = %gep.next118
  %63 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14
  %.unpack346 = load ptr, ptr %63, align 4
  %.elt347 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 1
  %.unpack348 = load i32, ptr %.elt347, align 4
  %.elt349 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 2
  %.unpack350 = load i32, ptr %.elt349, align 4
  call void @runtime.trackPointer(ptr %.unpack346, ptr undef) #14
  call void @"(*main.Cmd).closeDescriptors"(ptr %c, ptr %.unpack346, i32 %.unpack348, i32 %.unpack350, ptr undef)
  %64 = icmp eq ptr %c, null
  br i1 %64, label %gep.throw121, label %gep.next122

gep.next122:                                      ; preds = %deref.next120
  br i1 false, label %deref.throw123, label %deref.next124

deref.next124:                                    ; preds = %gep.next122
  %65 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15
  %.unpack352 = load ptr, ptr %65, align 4
  %.elt353 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 1
  %.unpack354 = load i32, ptr %.elt353, align 4
  %.elt355 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 2
  %.unpack356 = load i32, ptr %.elt355, align 4
  call void @runtime.trackPointer(ptr %.unpack352, ptr undef) #14
  call void @"(*main.Cmd).closeDescriptors"(ptr %c, ptr %.unpack352, i32 %.unpack354, i32 %.unpack356, ptr undef)
  ret %runtime._interface %60

if.done15:                                        ; preds = %fpcall.next
  %66 = icmp eq ptr %c, null
  br i1 %66, label %gep.throw125, label %gep.next126

gep.next126:                                      ; preds = %if.done15
  %67 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 13
  br i1 false, label %gep.throw127, label %gep.next128

gep.next128:                                      ; preds = %gep.next126
  br i1 false, label %deref.throw129, label %deref.next130

deref.next130:                                    ; preds = %gep.next128
  %68 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 13
  %.unpack335 = load ptr, ptr %68, align 4
  %.elt336 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 13, i32 1
  %.unpack337 = load i32, ptr %.elt336, align 4
  %.elt338 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 13, i32 2
  %.unpack339 = load i32, ptr %.elt338, align 4
  call void @runtime.trackPointer(ptr %.unpack335, ptr undef) #14
  %varargs = call ptr @runtime.alloc(i32 4, ptr nonnull inttoptr (i32 67 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %varargs, ptr undef) #14
  store ptr %59, ptr %varargs, align 4
  %append.new = call { ptr, i32, i32 } @runtime.sliceAppend(ptr %.unpack335, ptr nonnull %varargs, i32 %.unpack337, i32 %.unpack339, i32 1, i32 4, ptr undef) #14
  %append.newPtr = extractvalue { ptr, i32, i32 } %append.new, 0
  call void @runtime.trackPointer(ptr %append.newPtr, ptr undef) #14
  br i1 false, label %store.throw132, label %store.next133

store.next133:                                    ; preds = %deref.next130
  %append.newLen = extractvalue { ptr, i32, i32 } %append.new, 1
  %append.newCap = extractvalue { ptr, i32, i32 } %append.new, 2
  store ptr %append.newPtr, ptr %67, align 4
  %.repack341 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 13, i32 1
  store i32 %append.newLen, ptr %.repack341, align 4
  %.repack343 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 13, i32 2
  store i32 %append.newCap, ptr %.repack343, align 4
  br label %rangeindex.loop

rangeindex.done:                                  ; preds = %rangeindex.loop
  br i1 false, label %gep.throw134, label %gep.next135

gep.next135:                                      ; preds = %rangeindex.done
  %69 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 13
  br i1 false, label %gep.throw136, label %gep.next137

gep.next137:                                      ; preds = %gep.next135
  br i1 false, label %deref.throw138, label %deref.next139

deref.next139:                                    ; preds = %gep.next137
  %70 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 13
  %.unpack248 = load ptr, ptr %70, align 4
  %.elt249 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 13, i32 1
  %.unpack250 = load i32, ptr %.elt249, align 4
  %.elt251 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 13, i32 2
  %.unpack252 = load i32, ptr %.elt251, align 4
  call void @runtime.trackPointer(ptr %.unpack248, ptr undef) #14
  br i1 false, label %gep.throw140, label %gep.next141

gep.next141:                                      ; preds = %deref.next139
  br i1 false, label %deref.throw142, label %deref.next143

deref.next143:                                    ; preds = %gep.next141
  %71 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 7
  %.unpack253 = load ptr, ptr %71, align 4
  %.elt254 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 7, i32 1
  %.unpack255 = load i32, ptr %.elt254, align 4
  call void @runtime.trackPointer(ptr %.unpack253, ptr undef) #14
  %append.new149 = call { ptr, i32, i32 } @runtime.sliceAppend(ptr %.unpack248, ptr %.unpack253, i32 %.unpack250, i32 %.unpack252, i32 %.unpack255, i32 4, ptr undef) #14
  %append.newPtr150 = extractvalue { ptr, i32, i32 } %append.new149, 0
  call void @runtime.trackPointer(ptr %append.newPtr150, ptr undef) #14
  br i1 false, label %store.throw153, label %store.next154

store.next154:                                    ; preds = %deref.next143
  %append.newLen151 = extractvalue { ptr, i32, i32 } %append.new149, 1
  %append.newCap152 = extractvalue { ptr, i32, i32 } %append.new149, 2
  store ptr %append.newPtr150, ptr %69, align 4
  %.repack258 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 13, i32 1
  store i32 %append.newLen151, ptr %.repack258, align 4
  %.repack260 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 13, i32 2
  store i32 %append.newCap152, ptr %.repack260, align 4
  %72 = call { { ptr, i32, i32 }, %runtime._interface } @"(*main.Cmd).environ"(ptr nonnull %c, ptr undef)
  %73 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %72, 0
  %74 = extractvalue { ptr, i32, i32 } %73, 0
  call void @runtime.trackPointer(ptr %74, ptr undef) #14
  %75 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %72, 1
  %76 = extractvalue %runtime._interface %75, 1
  call void @runtime.trackPointer(ptr %76, ptr undef) #14
  %77 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %72, 0
  %78 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %72, 1
  %79 = extractvalue %runtime._interface %78, 0
  %.not262 = icmp eq i32 %79, 0
  br i1 %.not262, label %if.done17, label %if.then16

if.then16:                                        ; preds = %store.next154
  ret %runtime._interface %78

if.done17:                                        ; preds = %store.next154
  br i1 false, label %gep.throw155, label %gep.next156

gep.next156:                                      ; preds = %if.done17
  %80 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 9
  br i1 false, label %gep.throw157, label %gep.next158

gep.next158:                                      ; preds = %gep.next156
  br i1 false, label %deref.throw159, label %deref.next160

deref.next160:                                    ; preds = %gep.next158
  %.unpack263 = load ptr, ptr %c, align 4
  %.elt264 = getelementptr inbounds %runtime._string, ptr %c, i32 0, i32 1
  %.unpack265 = load i32, ptr %.elt264, align 4
  call void @runtime.trackPointer(ptr %.unpack263, ptr undef) #14
  %81 = call { ptr, i32, i32 } @"(*main.Cmd).argv"(ptr nonnull %c, ptr undef)
  %82 = extractvalue { ptr, i32, i32 } %81, 0
  call void @runtime.trackPointer(ptr %82, ptr undef) #14
  %complit = call ptr @runtime.alloc(i32 36, ptr nonnull inttoptr (i32 18771 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %complit, ptr undef) #14
  br i1 false, label %gep.throw161, label %gep.next162

gep.next162:                                      ; preds = %deref.next160
  br i1 false, label %deref.throw163, label %deref.next164

deref.next164:                                    ; preds = %gep.next162
  %83 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 3
  %.unpack266 = load ptr, ptr %83, align 4
  %.elt267 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 3, i32 1
  %.unpack268 = load i32, ptr %.elt267, align 4
  call void @runtime.trackPointer(ptr %.unpack266, ptr undef) #14
  %84 = getelementptr inbounds %os.ProcAttr, ptr %complit, i32 0, i32 2
  br i1 false, label %gep.throw165, label %gep.next166

gep.next166:                                      ; preds = %deref.next164
  br i1 false, label %deref.throw167, label %deref.next168

deref.next168:                                    ; preds = %gep.next166
  %85 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 13
  %.unpack269 = load ptr, ptr %85, align 4
  %.elt270 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 13, i32 1
  %.unpack271 = load i32, ptr %.elt270, align 4
  %.elt272 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 13, i32 2
  %.unpack273 = load i32, ptr %.elt272, align 4
  call void @runtime.trackPointer(ptr %.unpack269, ptr undef) #14
  %86 = getelementptr inbounds %os.ProcAttr, ptr %complit, i32 0, i32 1
  %87 = getelementptr inbounds %os.ProcAttr, ptr %complit, i32 0, i32 3
  br i1 false, label %gep.throw169, label %gep.next170

gep.next170:                                      ; preds = %deref.next168
  br i1 false, label %deref.throw171, label %deref.next172

deref.next172:                                    ; preds = %gep.next170
  %88 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 8
  %89 = load ptr, ptr %88, align 4
  call void @runtime.trackPointer(ptr %89, ptr undef) #14
  br i1 false, label %store.throw173, label %store.next174

store.next174:                                    ; preds = %deref.next172
  store ptr %.unpack266, ptr %complit, align 4
  %complit.repack274 = getelementptr inbounds %runtime._string, ptr %complit, i32 0, i32 1
  store i32 %.unpack268, ptr %complit.repack274, align 4
  br i1 false, label %store.throw175, label %store.next176

store.next176:                                    ; preds = %store.next174
  store ptr %.unpack269, ptr %84, align 4
  %.repack276 = getelementptr inbounds %os.ProcAttr, ptr %complit, i32 0, i32 2, i32 1
  store i32 %.unpack271, ptr %.repack276, align 4
  %.repack278 = getelementptr inbounds %os.ProcAttr, ptr %complit, i32 0, i32 2, i32 2
  store i32 %.unpack273, ptr %.repack278, align 4
  br i1 false, label %store.throw177, label %store.next178

store.next178:                                    ; preds = %store.next176
  %.elt = extractvalue { ptr, i32, i32 } %77, 0
  store ptr %.elt, ptr %86, align 4
  %.repack280 = getelementptr inbounds %os.ProcAttr, ptr %complit, i32 0, i32 1, i32 1
  %.elt281 = extractvalue { ptr, i32, i32 } %77, 1
  store i32 %.elt281, ptr %.repack280, align 4
  %.repack282 = getelementptr inbounds %os.ProcAttr, ptr %complit, i32 0, i32 1, i32 2
  %.elt283 = extractvalue { ptr, i32, i32 } %77, 2
  store i32 %.elt283, ptr %.repack282, align 4
  br i1 false, label %store.throw179, label %store.next180

store.next180:                                    ; preds = %store.next178
  store ptr %89, ptr %87, align 4
  %90 = extractvalue { ptr, i32, i32 } %81, 0
  %91 = extractvalue { ptr, i32, i32 } %81, 1
  %92 = extractvalue { ptr, i32, i32 } %81, 2
  %93 = call { ptr, %runtime._interface } @os.StartProcess(ptr %.unpack263, i32 %.unpack265, ptr %90, i32 %91, i32 %92, ptr nonnull %complit, ptr undef) #14
  %94 = extractvalue { ptr, %runtime._interface } %93, 0
  call void @runtime.trackPointer(ptr %94, ptr undef) #14
  %95 = extractvalue { ptr, %runtime._interface } %93, 1
  %96 = extractvalue %runtime._interface %95, 1
  call void @runtime.trackPointer(ptr %96, ptr undef) #14
  br i1 false, label %store.throw181, label %store.next182

store.next182:                                    ; preds = %store.next180
  %97 = extractvalue { ptr, %runtime._interface } %93, 0
  store ptr %97, ptr %80, align 4
  %98 = extractvalue { ptr, %runtime._interface } %93, 1
  %99 = extractvalue %runtime._interface %98, 0
  %.not284 = icmp eq i32 %99, 0
  br i1 %.not284, label %if.done19, label %if.then18

if.then18:                                        ; preds = %store.next182
  br i1 false, label %gep.throw183, label %gep.next184

gep.next184:                                      ; preds = %if.then18
  br i1 false, label %deref.throw185, label %deref.next186

deref.next186:                                    ; preds = %gep.next184
  %100 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14
  %.unpack318 = load ptr, ptr %100, align 4
  %.elt319 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 1
  %.unpack320 = load i32, ptr %.elt319, align 4
  %.elt321 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 2
  %.unpack322 = load i32, ptr %.elt321, align 4
  call void @runtime.trackPointer(ptr %.unpack318, ptr undef) #14
  call void @"(*main.Cmd).closeDescriptors"(ptr nonnull %c, ptr %.unpack318, i32 %.unpack320, i32 %.unpack322, ptr undef)
  br i1 false, label %gep.throw187, label %gep.next188

gep.next188:                                      ; preds = %deref.next186
  br i1 false, label %deref.throw189, label %deref.next190

deref.next190:                                    ; preds = %gep.next188
  %101 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15
  %.unpack324 = load ptr, ptr %101, align 4
  %.elt325 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 1
  %.unpack326 = load i32, ptr %.elt325, align 4
  %.elt327 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 2
  %.unpack328 = load i32, ptr %.elt327, align 4
  call void @runtime.trackPointer(ptr %.unpack324, ptr undef) #14
  call void @"(*main.Cmd).closeDescriptors"(ptr nonnull %c, ptr %.unpack324, i32 %.unpack326, i32 %.unpack328, ptr undef)
  ret %runtime._interface %98

if.done19:                                        ; preds = %store.next182
  br i1 false, label %gep.throw191, label %gep.next192

gep.next192:                                      ; preds = %if.done19
  br i1 false, label %deref.throw193, label %deref.next194

deref.next194:                                    ; preds = %gep.next192
  %102 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14
  %.unpack286 = load ptr, ptr %102, align 4
  %.elt287 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 1
  %.unpack288 = load i32, ptr %.elt287, align 4
  %.elt289 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 2
  %.unpack290 = load i32, ptr %.elt289, align 4
  call void @runtime.trackPointer(ptr %.unpack286, ptr undef) #14
  call void @"(*main.Cmd).closeDescriptors"(ptr nonnull %c, ptr %.unpack286, i32 %.unpack288, i32 %.unpack290, ptr undef)
  br i1 false, label %gep.throw195, label %gep.next196

gep.next196:                                      ; preds = %deref.next194
  br i1 false, label %deref.throw197, label %deref.next198

deref.next198:                                    ; preds = %gep.next196
  %103 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 16
  %.unpack292 = load ptr, ptr %103, align 4
  %.elt293 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 16, i32 1
  %.unpack294 = load i32, ptr %.elt293, align 4
  call void @runtime.trackPointer(ptr %.unpack292, ptr undef) #14
  %104 = icmp sgt i32 %.unpack294, 0
  br i1 %104, label %if.then20, label %if.done23

if.then20:                                        ; preds = %deref.next198
  %errc = call ptr @runtime.alloc(i32 4, ptr nonnull inttoptr (i32 67 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %errc, ptr undef) #14
  br i1 false, label %gep.throw200, label %gep.next201

gep.next201:                                      ; preds = %if.then20
  br i1 false, label %deref.throw202, label %deref.next203

deref.next203:                                    ; preds = %gep.next201
  %105 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 16
  %.unpack298 = load ptr, ptr %105, align 4
  %.elt299 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 16, i32 1
  %.unpack300 = load i32, ptr %.elt299, align 4
  call void @runtime.trackPointer(ptr %.unpack298, ptr undef) #14
  %106 = icmp ugt i32 %.unpack300, 268435454
  br i1 %106, label %chan.throw, label %chan.next

chan.next:                                        ; preds = %deref.next203
  %107 = call ptr @runtime.chanMake(i32 8, i32 %.unpack300, ptr undef) #14
  call void @runtime.trackPointer(ptr %107, ptr undef) #14
  store ptr %107, ptr %errc, align 4
  br i1 false, label %gep.throw205, label %gep.next206

gep.next206:                                      ; preds = %chan.next
  %108 = load ptr, ptr %errc, align 4
  call void @runtime.trackPointer(ptr %108, ptr undef) #14
  br i1 false, label %store.throw207, label %store.next208

store.next208:                                    ; preds = %gep.next206
  %109 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 17
  store ptr %108, ptr %109, align 4
  br i1 false, label %gep.throw209, label %gep.next210

gep.next210:                                      ; preds = %store.next208
  br i1 false, label %deref.throw211, label %deref.next212

deref.next212:                                    ; preds = %gep.next210
  %110 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 16
  %.unpack304 = load ptr, ptr %110, align 4
  %.elt305 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 16, i32 1
  %.unpack306 = load i32, ptr %.elt305, align 4
  call void @runtime.trackPointer(ptr %.unpack304, ptr undef) #14
  br label %rangeindex.loop21

rangeindex.loop21:                                ; preds = %lookup.next217, %deref.next212
  %111 = phi i32 [ -1, %deref.next212 ], [ %112, %lookup.next217 ]
  %112 = add i32 %111, 1
  %113 = icmp slt i32 %112, %.unpack306
  br i1 %113, label %rangeindex.body22, label %if.done23

rangeindex.body22:                                ; preds = %rangeindex.loop21
  %.not309 = icmp ult i32 %112, %.unpack306
  br i1 %.not309, label %lookup.next217, label %lookup.throw216

lookup.next217:                                   ; preds = %rangeindex.body22
  %114 = getelementptr inbounds { ptr, ptr }, ptr %.unpack304, i32 %112
  %.unpack311 = load ptr, ptr %114, align 4
  %.elt312 = getelementptr inbounds { ptr, ptr }, ptr %.unpack304, i32 %112, i32 1
  %.unpack313 = load ptr, ptr %.elt312, align 4
  call void @runtime.trackPointer(ptr %.unpack311, ptr undef) #14
  call void @runtime.trackPointer(ptr %.unpack313, ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %errc, ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull @"(*main.Cmd).Start$1", ptr undef) #14
  %115 = call ptr @runtime.alloc(i32 12, ptr null, ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %115, ptr undef) #14
  store ptr %.unpack311, ptr %115, align 4
  %.repack315 = getelementptr inbounds { ptr, ptr }, ptr %115, i32 0, i32 1
  store ptr %.unpack313, ptr %.repack315, align 4
  %116 = getelementptr inbounds { { ptr, ptr }, ptr }, ptr %115, i32 0, i32 1
  store ptr %errc, ptr %116, align 4
  call void @"internal/task.start"(i32 ptrtoint (ptr @"(*main.Cmd).Start$1$gowrapper" to i32), ptr nonnull %115, i32 16384, ptr undef) #14
  br label %rangeindex.loop21

if.done23:                                        ; preds = %rangeindex.loop21, %deref.next198
  br i1 false, label %gep.throw218, label %gep.next219

gep.next219:                                      ; preds = %if.done23
  %117 = call ptr @"(*main.Cmd).watchCtx"(ptr nonnull %c, ptr undef)
  call void @runtime.trackPointer(ptr %117, ptr undef) #14
  br i1 false, label %store.throw220, label %store.next221

store.next221:                                    ; preds = %gep.next219
  %118 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 18
  store ptr %117, ptr %118, align 4
  ret %runtime._interface zeroinitializer

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable

gep.throw24:                                      ; preds = %cond.true
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw26:                                    ; preds = %gep.next25
  unreachable

gep.throw28:                                      ; preds = %cond.true1
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw30:                                    ; preds = %gep.next29
  unreachable

gep.throw32:                                      ; preds = %if.then
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

store.throw:                                      ; preds = %gep.next33
  unreachable

gep.throw34:                                      ; preds = %if.done
  unreachable

deref.throw36:                                    ; preds = %gep.next35
  unreachable

gep.throw38:                                      ; preds = %if.then2
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw40:                                    ; preds = %gep.next39
  unreachable

gep.throw42:                                      ; preds = %deref.next41
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw44:                                    ; preds = %gep.next43
  unreachable

gep.throw46:                                      ; preds = %deref.next45
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw48:                                    ; preds = %gep.next47
  unreachable

gep.throw50:                                      ; preds = %if.then3
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw52:                                    ; preds = %gep.next51
  unreachable

gep.throw54:                                      ; preds = %if.done4
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw56:                                    ; preds = %gep.next55
  unreachable

gep.throw58:                                      ; preds = %cond.false
  unreachable

deref.throw60:                                    ; preds = %gep.next59
  unreachable

gep.throw62:                                      ; preds = %if.then6
  unreachable

deref.throw64:                                    ; preds = %gep.next63
  unreachable

gep.throw66:                                      ; preds = %deref.next65
  unreachable

deref.throw68:                                    ; preds = %gep.next67
  unreachable

gep.throw70:                                      ; preds = %if.then7
  unreachable

deref.throw72:                                    ; preds = %gep.next71
  unreachable

gep.throw74:                                      ; preds = %deref.next73
  unreachable

deref.throw76:                                    ; preds = %gep.next75
  unreachable

gep.throw78:                                      ; preds = %if.done8
  unreachable

store.throw80:                                    ; preds = %gep.next79
  unreachable

gep.throw82:                                      ; preds = %if.done9
  unreachable

deref.throw84:                                    ; preds = %gep.next83
  unreachable

gep.throw86:                                      ; preds = %if.done11
  unreachable

deref.throw88:                                    ; preds = %gep.next87
  unreachable

gep.throw90:                                      ; preds = %if.then12
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw92:                                    ; preds = %gep.next91
  unreachable

gep.throw94:                                      ; preds = %select.body
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw96:                                    ; preds = %gep.next95
  unreachable

gep.throw98:                                      ; preds = %deref.next97
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw100:                                   ; preds = %gep.next99
  unreachable

gep.throw102:                                     ; preds = %deref.next101
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw104:                                   ; preds = %gep.next103
  unreachable

gep.throw108:                                     ; preds = %if.done13
  unreachable

gep.throw110:                                     ; preds = %gep.next109
  unreachable

deref.throw112:                                   ; preds = %gep.next111
  unreachable

slice.throw:                                      ; preds = %deref.next113
  call void @runtime.slicePanic(ptr undef) #14
  unreachable

store.throw114:                                   ; preds = %slice.next
  unreachable

lookup.throw:                                     ; preds = %rangeindex.body
  call void @runtime.lookupPanic(ptr undef) #14
  unreachable

fpcall.throw:                                     ; preds = %lookup.next
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

gep.throw117:                                     ; preds = %if.then14
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw119:                                   ; preds = %gep.next118
  unreachable

gep.throw121:                                     ; preds = %deref.next120
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw123:                                   ; preds = %gep.next122
  unreachable

gep.throw125:                                     ; preds = %if.done15
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

gep.throw127:                                     ; preds = %gep.next126
  unreachable

deref.throw129:                                   ; preds = %gep.next128
  unreachable

store.throw132:                                   ; preds = %deref.next130
  unreachable

gep.throw134:                                     ; preds = %rangeindex.done
  unreachable

gep.throw136:                                     ; preds = %gep.next135
  unreachable

deref.throw138:                                   ; preds = %gep.next137
  unreachable

gep.throw140:                                     ; preds = %deref.next139
  unreachable

deref.throw142:                                   ; preds = %gep.next141
  unreachable

store.throw153:                                   ; preds = %deref.next143
  unreachable

gep.throw155:                                     ; preds = %if.done17
  unreachable

gep.throw157:                                     ; preds = %gep.next156
  unreachable

deref.throw159:                                   ; preds = %gep.next158
  unreachable

gep.throw161:                                     ; preds = %deref.next160
  unreachable

deref.throw163:                                   ; preds = %gep.next162
  unreachable

gep.throw165:                                     ; preds = %deref.next164
  unreachable

deref.throw167:                                   ; preds = %gep.next166
  unreachable

gep.throw169:                                     ; preds = %deref.next168
  unreachable

deref.throw171:                                   ; preds = %gep.next170
  unreachable

store.throw173:                                   ; preds = %deref.next172
  unreachable

store.throw175:                                   ; preds = %store.next174
  unreachable

store.throw177:                                   ; preds = %store.next176
  unreachable

store.throw179:                                   ; preds = %store.next178
  unreachable

store.throw181:                                   ; preds = %store.next180
  unreachable

gep.throw183:                                     ; preds = %if.then18
  unreachable

deref.throw185:                                   ; preds = %gep.next184
  unreachable

gep.throw187:                                     ; preds = %deref.next186
  unreachable

deref.throw189:                                   ; preds = %gep.next188
  unreachable

gep.throw191:                                     ; preds = %if.done19
  unreachable

deref.throw193:                                   ; preds = %gep.next192
  unreachable

gep.throw195:                                     ; preds = %deref.next194
  unreachable

deref.throw197:                                   ; preds = %gep.next196
  unreachable

gep.throw200:                                     ; preds = %if.then20
  unreachable

deref.throw202:                                   ; preds = %gep.next201
  unreachable

chan.throw:                                       ; preds = %deref.next203
  call void @runtime.chanMakePanic(ptr undef) #14
  unreachable

gep.throw205:                                     ; preds = %chan.next
  unreachable

store.throw207:                                   ; preds = %gep.next206
  unreachable

gep.throw209:                                     ; preds = %store.next208
  unreachable

deref.throw211:                                   ; preds = %gep.next210
  unreachable

lookup.throw216:                                  ; preds = %rangeindex.body22
  call void @runtime.lookupPanic(ptr undef) #14
  unreachable

gep.throw218:                                     ; preds = %if.done23
  unreachable

store.throw220:                                   ; preds = %gep.next219
  unreachable
}

; Function Attrs: nounwind
define hidden %runtime._interface @"(*main.Cmd).Wait"(ptr dereferenceable_or_null(168) %c, ptr %context) unnamed_addr #1 {
entry:
  %chan.blockedList46 = alloca %runtime.channelBlockedList, align 8
  %chan.value45 = alloca %runtime._interface, align 8
  %chan.blockedList = alloca %runtime.channelBlockedList, align 8
  %chan.value = alloca %runtime._interface, align 8
  %0 = icmp eq ptr %c, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 9
  %2 = load ptr, ptr %1, align 4
  call void @runtime.trackPointer(ptr %2, ptr undef) #14
  %3 = icmp eq ptr %2, null
  br i1 %3, label %if.then, label %if.done

if.then:                                          ; preds = %deref.next
  %4 = call %runtime._interface @errors.New(ptr nonnull @"main$string.58", i32 17, ptr undef) #14
  %5 = extractvalue %runtime._interface %4, 1
  call void @runtime.trackPointer(ptr %5, ptr undef) #14
  ret %runtime._interface %4

if.done:                                          ; preds = %deref.next
  %6 = icmp eq ptr %c, null
  br i1 %6, label %gep.throw13, label %gep.next14

gep.next14:                                       ; preds = %if.done
  br i1 false, label %deref.throw15, label %deref.next16

deref.next16:                                     ; preds = %gep.next14
  %7 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 10
  %8 = load ptr, ptr %7, align 4
  call void @runtime.trackPointer(ptr %8, ptr undef) #14
  %.not = icmp eq ptr %8, null
  br i1 %.not, label %if.done2, label %if.then1

if.then1:                                         ; preds = %deref.next16
  %9 = call %runtime._interface @errors.New(ptr nonnull @"main$string.59", i32 29, ptr undef) #14
  %10 = extractvalue %runtime._interface %9, 1
  call void @runtime.trackPointer(ptr %10, ptr undef) #14
  ret %runtime._interface %9

if.done2:                                         ; preds = %deref.next16
  br i1 false, label %gep.throw17, label %gep.next18

gep.next18:                                       ; preds = %if.done2
  br i1 false, label %deref.throw19, label %deref.next20

deref.next20:                                     ; preds = %gep.next18
  %11 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 9
  %12 = load ptr, ptr %11, align 4
  call void @runtime.trackPointer(ptr %12, ptr undef) #14
  %13 = call { ptr, %runtime._interface } @"(*os.Process).Wait"(ptr %12, ptr undef) #14
  %14 = extractvalue { ptr, %runtime._interface } %13, 0
  call void @runtime.trackPointer(ptr %14, ptr undef) #14
  %15 = extractvalue { ptr, %runtime._interface } %13, 1
  %16 = extractvalue %runtime._interface %15, 1
  call void @runtime.trackPointer(ptr %16, ptr undef) #14
  %17 = extractvalue { ptr, %runtime._interface } %13, 0
  %18 = extractvalue { ptr, %runtime._interface } %13, 1
  %19 = extractvalue %runtime._interface %18, 0
  %20 = icmp eq i32 %19, 0
  br i1 %20, label %cond.true, label %if.done4

cond.true:                                        ; preds = %deref.next20
  %21 = call i1 @"(*os.ProcessState).Success"(ptr %17, ptr undef) #14
  br i1 %21, label %if.done4, label %if.then3

if.then3:                                         ; preds = %cond.true
  %complit = call ptr @runtime.alloc(i32 16, ptr nonnull inttoptr (i32 201 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %complit, ptr undef) #14
  br i1 false, label %store.throw, label %store.next

store.next:                                       ; preds = %if.then3
  store ptr %17, ptr %complit, align 4
  %22 = insertvalue %runtime._interface { i32 ptrtoint (ptr @"reflect/types.type:pointer:named:main.ExitError" to i32), ptr undef }, ptr %complit, 1
  call void @runtime.trackPointer(ptr nonnull %complit, ptr undef) #14
  br label %if.done4

if.done4:                                         ; preds = %store.next, %cond.true, %deref.next20
  %23 = phi %runtime._interface [ %18, %deref.next20 ], [ %18, %cond.true ], [ %22, %store.next ]
  %24 = extractvalue %runtime._interface %23, 1
  call void @runtime.trackPointer(ptr %24, ptr undef) #14
  br i1 false, label %gep.throw21, label %gep.next22

gep.next22:                                       ; preds = %if.done4
  br i1 false, label %store.throw23, label %store.next24

store.next24:                                     ; preds = %gep.next22
  %25 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 10
  store ptr %17, ptr %25, align 4
  br i1 false, label %gep.throw25, label %gep.next26

gep.next26:                                       ; preds = %store.next24
  br i1 false, label %deref.throw27, label %deref.next28

deref.next28:                                     ; preds = %gep.next26
  %26 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 16
  %.unpack = load ptr, ptr %26, align 4
  %.elt56 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 16, i32 1
  %.unpack57 = load i32, ptr %.elt56, align 4
  call void @runtime.trackPointer(ptr %.unpack, ptr undef) #14
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %if.then6, %cond.true5, %deref.next32, %deref.next28
  %27 = phi %runtime._interface [ zeroinitializer, %deref.next28 ], [ %27, %deref.next32 ], [ %27, %cond.true5 ], [ %chan.received76, %if.then6 ]
  %28 = phi i32 [ -1, %deref.next28 ], [ %30, %deref.next32 ], [ %30, %cond.true5 ], [ %30, %if.then6 ]
  %29 = extractvalue %runtime._interface %27, 1
  call void @runtime.trackPointer(ptr %29, ptr undef) #14
  %30 = add i32 %28, 1
  %31 = icmp slt i32 %30, %.unpack57
  br i1 %31, label %rangeindex.body, label %rangeindex.done

rangeindex.body:                                  ; preds = %rangeindex.loop
  br i1 false, label %gep.throw29, label %gep.next30

gep.next30:                                       ; preds = %rangeindex.body
  br i1 false, label %deref.throw31, label %deref.next32

deref.next32:                                     ; preds = %gep.next30
  %32 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 17
  %33 = load ptr, ptr %32, align 4
  call void @runtime.trackPointer(ptr %33, ptr undef) #14
  call void @llvm.lifetime.start.p0(i64 8, ptr nonnull %chan.value)
  call void @llvm.lifetime.start.p0(i64 24, ptr nonnull %chan.blockedList)
  %34 = call i1 @runtime.chanRecv(ptr %33, ptr nonnull %chan.value, ptr nonnull %chan.blockedList, ptr undef) #14
  %chan.received.unpack = load i32, ptr %chan.value, align 8
  %35 = insertvalue %runtime._interface undef, i32 %chan.received.unpack, 0
  %chan.received.elt74 = getelementptr inbounds %runtime._interface, ptr %chan.value, i32 0, i32 1
  %chan.received.unpack75 = load ptr, ptr %chan.received.elt74, align 4
  %chan.received76 = insertvalue %runtime._interface %35, ptr %chan.received.unpack75, 1
  call void @llvm.lifetime.end.p0(i64 8, ptr nonnull %chan.value)
  call void @llvm.lifetime.end.p0(i64 24, ptr nonnull %chan.blockedList)
  call void @runtime.trackPointer(ptr %chan.received.unpack75, ptr undef) #14
  %.not77 = icmp eq i32 %chan.received.unpack, 0
  br i1 %.not77, label %rangeindex.loop, label %cond.true5

cond.true5:                                       ; preds = %deref.next32
  %36 = extractvalue %runtime._interface %27, 0
  %37 = icmp eq i32 %36, 0
  br i1 %37, label %if.then6, label %rangeindex.loop

if.then6:                                         ; preds = %cond.true5
  br label %rangeindex.loop

rangeindex.done:                                  ; preds = %rangeindex.loop
  br i1 false, label %gep.throw33, label %gep.next34

gep.next34:                                       ; preds = %rangeindex.done
  br i1 false, label %store.throw35, label %store.next36

store.next36:                                     ; preds = %gep.next34
  %38 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 16
  store ptr null, ptr %38, align 4
  %.repack60 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 16, i32 1
  store i32 0, ptr %.repack60, align 4
  %.repack61 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 16, i32 2
  store i32 0, ptr %.repack61, align 4
  br i1 false, label %gep.throw37, label %gep.next38

gep.next38:                                       ; preds = %store.next36
  br i1 false, label %deref.throw39, label %deref.next40

deref.next40:                                     ; preds = %gep.next38
  %39 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 18
  %40 = load ptr, ptr %39, align 4
  call void @runtime.trackPointer(ptr %40, ptr undef) #14
  %.not62 = icmp eq ptr %40, null
  br i1 %.not62, label %if.done10, label %if.then7

if.then7:                                         ; preds = %deref.next40
  br i1 false, label %gep.throw41, label %gep.next42

gep.next42:                                       ; preds = %if.then7
  br i1 false, label %deref.throw43, label %deref.next44

deref.next44:                                     ; preds = %gep.next42
  %41 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 18
  %42 = load ptr, ptr %41, align 4
  call void @runtime.trackPointer(ptr %42, ptr undef) #14
  call void @llvm.lifetime.start.p0(i64 8, ptr nonnull %chan.value45)
  call void @llvm.lifetime.start.p0(i64 24, ptr nonnull %chan.blockedList46)
  %43 = call i1 @runtime.chanRecv(ptr %42, ptr nonnull %chan.value45, ptr nonnull %chan.blockedList46, ptr undef) #14
  %chan.received47.unpack = load i32, ptr %chan.value45, align 8
  %44 = insertvalue %runtime._interface undef, i32 %chan.received47.unpack, 0
  %chan.received47.elt70 = getelementptr inbounds %runtime._interface, ptr %chan.value45, i32 0, i32 1
  %chan.received47.unpack71 = load ptr, ptr %chan.received47.elt70, align 4
  %chan.received4772 = insertvalue %runtime._interface %44, ptr %chan.received47.unpack71, 1
  call void @llvm.lifetime.end.p0(i64 8, ptr nonnull %chan.value45)
  call void @llvm.lifetime.end.p0(i64 24, ptr nonnull %chan.blockedList46)
  call void @runtime.trackPointer(ptr %chan.received47.unpack71, ptr undef) #14
  %.not73 = icmp eq i32 %chan.received47.unpack, 0
  br i1 %.not73, label %if.done10, label %cond.true8

cond.true8:                                       ; preds = %deref.next44
  %45 = extractvalue %runtime._interface %23, 0
  %46 = icmp eq i32 %45, 0
  br i1 %46, label %if.then9, label %if.done10

if.then9:                                         ; preds = %cond.true8
  br label %if.done10

if.done10:                                        ; preds = %if.then9, %cond.true8, %deref.next44, %deref.next40
  %47 = phi %runtime._interface [ %23, %deref.next40 ], [ %23, %deref.next44 ], [ %23, %cond.true8 ], [ %chan.received4772, %if.then9 ]
  %48 = extractvalue %runtime._interface %47, 1
  call void @runtime.trackPointer(ptr %48, ptr undef) #14
  %49 = extractvalue %runtime._interface %47, 0
  %50 = icmp eq i32 %49, 0
  br i1 %50, label %if.then11, label %if.done12

if.then11:                                        ; preds = %if.done10
  br label %if.done12

if.done12:                                        ; preds = %if.then11, %if.done10
  %51 = phi %runtime._interface [ %47, %if.done10 ], [ %27, %if.then11 ]
  %52 = extractvalue %runtime._interface %51, 1
  call void @runtime.trackPointer(ptr %52, ptr undef) #14
  br i1 false, label %gep.throw48, label %gep.next49

gep.next49:                                       ; preds = %if.done12
  br i1 false, label %deref.throw50, label %deref.next51

deref.next51:                                     ; preds = %gep.next49
  %53 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15
  %.unpack63 = load ptr, ptr %53, align 4
  %.elt64 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 1
  %.unpack65 = load i32, ptr %.elt64, align 4
  %.elt66 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 2
  %.unpack67 = load i32, ptr %.elt66, align 4
  call void @runtime.trackPointer(ptr %.unpack63, ptr undef) #14
  call void @"(*main.Cmd).closeDescriptors"(ptr nonnull %c, ptr %.unpack63, i32 %.unpack65, i32 %.unpack67, ptr undef)
  br i1 false, label %gep.throw52, label %gep.next53

gep.next53:                                       ; preds = %deref.next51
  br i1 false, label %store.throw54, label %store.next55

store.next55:                                     ; preds = %gep.next53
  %54 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15
  store ptr null, ptr %54, align 4
  %.repack68 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 1
  store i32 0, ptr %.repack68, align 4
  %.repack69 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 2
  store i32 0, ptr %.repack69, align 4
  ret %runtime._interface %51

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable

gep.throw13:                                      ; preds = %if.done
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw15:                                    ; preds = %gep.next14
  unreachable

gep.throw17:                                      ; preds = %if.done2
  unreachable

deref.throw19:                                    ; preds = %gep.next18
  unreachable

store.throw:                                      ; preds = %if.then3
  unreachable

gep.throw21:                                      ; preds = %if.done4
  unreachable

store.throw23:                                    ; preds = %gep.next22
  unreachable

gep.throw25:                                      ; preds = %store.next24
  unreachable

deref.throw27:                                    ; preds = %gep.next26
  unreachable

gep.throw29:                                      ; preds = %rangeindex.body
  unreachable

deref.throw31:                                    ; preds = %gep.next30
  unreachable

gep.throw33:                                      ; preds = %rangeindex.done
  unreachable

store.throw35:                                    ; preds = %gep.next34
  unreachable

gep.throw37:                                      ; preds = %store.next36
  unreachable

deref.throw39:                                    ; preds = %gep.next38
  unreachable

gep.throw41:                                      ; preds = %if.then7
  unreachable

deref.throw43:                                    ; preds = %gep.next42
  unreachable

gep.throw48:                                      ; preds = %if.done12
  unreachable

deref.throw50:                                    ; preds = %gep.next49
  unreachable

gep.throw52:                                      ; preds = %deref.next51
  unreachable

store.throw54:                                    ; preds = %gep.next53
  unreachable
}

declare i1 @runtime.stringEqual(ptr, i32, ptr, i32, ptr) #0

; Function Attrs: nounwind
define hidden void @"(*main.Cmd).closeDescriptors"(ptr dereferenceable_or_null(168) %c, ptr %closers.data, i32 %closers.len, i32 %closers.cap, ptr %context) unnamed_addr #1 {
entry:
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %lookup.next, %entry
  %0 = phi i32 [ -1, %entry ], [ %1, %lookup.next ]
  %1 = add i32 %0, 1
  %2 = icmp slt i32 %1, %closers.len
  br i1 %2, label %rangeindex.body, label %rangeindex.done

rangeindex.body:                                  ; preds = %rangeindex.loop
  %.not = icmp ult i32 %1, %closers.len
  br i1 %.not, label %lookup.next, label %lookup.throw

lookup.next:                                      ; preds = %rangeindex.body
  %3 = getelementptr inbounds %runtime._interface, ptr %closers.data, i32 %1
  %.unpack = load i32, ptr %3, align 4
  %.elt1 = getelementptr inbounds %runtime._interface, ptr %closers.data, i32 %1, i32 1
  %.unpack2 = load ptr, ptr %.elt1, align 4
  call void @runtime.trackPointer(ptr %.unpack2, ptr undef) #14
  %4 = call %runtime._interface @"interface:{Close:func:{}{named:error}}.Close$invoke"(ptr %.unpack2, i32 %.unpack, ptr undef) #14
  %5 = extractvalue %runtime._interface %4, 1
  call void @runtime.trackPointer(ptr %5, ptr undef) #14
  br label %rangeindex.loop

rangeindex.done:                                  ; preds = %rangeindex.loop
  ret void

lookup.throw:                                     ; preds = %rangeindex.body
  call void @runtime.lookupPanic(ptr undef) #14
  unreachable
}

; Function Attrs: nounwind
define hidden { %runtime._string, %runtime._interface } @main.lookExtensions(ptr %path.data, i32 %path.len, ptr %dir.data, i32 %dir.len, ptr %context) unnamed_addr #1 {
entry:
  %0 = insertvalue %runtime._string zeroinitializer, ptr %path.data, 0
  %1 = insertvalue %runtime._string %0, i32 %path.len, 1
  %2 = call %runtime._string @"path/filepath.Base"(ptr %path.data, i32 %path.len, ptr undef) #14
  %3 = extractvalue %runtime._string %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %4 = extractvalue %runtime._string %2, 0
  %5 = extractvalue %runtime._string %2, 1
  %6 = call i1 @runtime.stringEqual(ptr %4, i32 %5, ptr %path.data, i32 %path.len, ptr undef) #14
  br i1 %6, label %if.then, label %if.done

if.then:                                          ; preds = %entry
  %7 = call %runtime._string @runtime.stringConcat(ptr nonnull @"main$string.77", i32 2, ptr %path.data, i32 %path.len, ptr undef) #14
  %8 = extractvalue %runtime._string %7, 0
  call void @runtime.trackPointer(ptr %8, ptr undef) #14
  br label %if.done

if.done:                                          ; preds = %if.then, %entry
  %9 = phi %runtime._string [ %1, %entry ], [ %7, %if.then ]
  %10 = extractvalue %runtime._string %9, 0
  call void @runtime.trackPointer(ptr %10, ptr undef) #14
  %11 = call i1 @runtime.stringEqual(ptr %dir.data, i32 %dir.len, ptr null, i32 0, ptr undef) #14
  br i1 %11, label %if.then1, label %if.done2

if.then1:                                         ; preds = %if.done
  %12 = extractvalue %runtime._string %9, 0
  %13 = extractvalue %runtime._string %9, 1
  %14 = call { %runtime._string, %runtime._interface } @main.LookPath(ptr %12, i32 %13, ptr undef)
  %15 = extractvalue { %runtime._string, %runtime._interface } %14, 0
  %16 = extractvalue %runtime._string %15, 0
  call void @runtime.trackPointer(ptr %16, ptr undef) #14
  %17 = extractvalue { %runtime._string, %runtime._interface } %14, 1
  %18 = extractvalue %runtime._interface %17, 1
  call void @runtime.trackPointer(ptr %18, ptr undef) #14
  ret { %runtime._string, %runtime._interface } %14

if.done2:                                         ; preds = %if.done
  %19 = extractvalue %runtime._string %9, 0
  %20 = extractvalue %runtime._string %9, 1
  %21 = call %runtime._string @"path/filepath.VolumeName"(ptr %19, i32 %20, ptr undef) #14
  %22 = extractvalue %runtime._string %21, 0
  call void @runtime.trackPointer(ptr %22, ptr undef) #14
  %23 = extractvalue %runtime._string %21, 0
  %24 = extractvalue %runtime._string %21, 1
  %25 = call i1 @runtime.stringEqual(ptr %23, i32 %24, ptr null, i32 0, ptr undef) #14
  br i1 %25, label %if.done4, label %if.then3

if.then3:                                         ; preds = %if.done2
  %26 = extractvalue %runtime._string %9, 0
  %27 = extractvalue %runtime._string %9, 1
  %28 = call { %runtime._string, %runtime._interface } @main.LookPath(ptr %26, i32 %27, ptr undef)
  %29 = extractvalue { %runtime._string, %runtime._interface } %28, 0
  %30 = extractvalue %runtime._string %29, 0
  call void @runtime.trackPointer(ptr %30, ptr undef) #14
  %31 = extractvalue { %runtime._string, %runtime._interface } %28, 1
  %32 = extractvalue %runtime._interface %31, 1
  call void @runtime.trackPointer(ptr %32, ptr undef) #14
  ret { %runtime._string, %runtime._interface } %28

if.done4:                                         ; preds = %if.done2
  %len = extractvalue %runtime._string %9, 1
  %33 = icmp sgt i32 %len, 1
  br i1 %33, label %cond.true, label %if.done6

cond.true:                                        ; preds = %if.done4
  %len9 = extractvalue %runtime._string %9, 1
  %34 = icmp eq i32 %len9, 0
  br i1 %34, label %lookup.throw, label %lookup.next

lookup.next:                                      ; preds = %cond.true
  %35 = extractvalue %runtime._string %9, 0
  %36 = load i8, ptr %35, align 1
  %37 = call i1 @os.IsPathSeparator(i8 %36, ptr undef) #14
  br i1 %37, label %if.then5, label %if.done6

if.then5:                                         ; preds = %lookup.next
  %38 = extractvalue %runtime._string %9, 0
  %39 = extractvalue %runtime._string %9, 1
  %40 = call { %runtime._string, %runtime._interface } @main.LookPath(ptr %38, i32 %39, ptr undef)
  %41 = extractvalue { %runtime._string, %runtime._interface } %40, 0
  %42 = extractvalue %runtime._string %41, 0
  call void @runtime.trackPointer(ptr %42, ptr undef) #14
  %43 = extractvalue { %runtime._string, %runtime._interface } %40, 1
  %44 = extractvalue %runtime._interface %43, 1
  call void @runtime.trackPointer(ptr %44, ptr undef) #14
  ret { %runtime._string, %runtime._interface } %40

if.done6:                                         ; preds = %lookup.next, %if.done4
  %varargs = call ptr @runtime.alloc(i32 16, ptr nonnull inttoptr (i32 69 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %varargs, ptr undef) #14
  store ptr %dir.data, ptr %varargs, align 4
  %varargs.repack10 = getelementptr inbounds %runtime._string, ptr %varargs, i32 0, i32 1
  store i32 %dir.len, ptr %varargs.repack10, align 4
  %45 = getelementptr inbounds [2 x %runtime._string], ptr %varargs, i32 0, i32 1
  %.elt = extractvalue %runtime._string %9, 0
  store ptr %.elt, ptr %45, align 4
  %.repack12 = getelementptr inbounds [2 x %runtime._string], ptr %varargs, i32 0, i32 1, i32 1
  %.elt13 = extractvalue %runtime._string %9, 1
  store i32 %.elt13, ptr %.repack12, align 4
  %46 = call %runtime._string @"path/filepath.Join"(ptr nonnull %varargs, i32 2, i32 2, ptr undef) #14
  %47 = extractvalue %runtime._string %46, 0
  call void @runtime.trackPointer(ptr %47, ptr undef) #14
  %48 = extractvalue %runtime._string %46, 0
  %49 = extractvalue %runtime._string %46, 1
  %50 = call { %runtime._string, %runtime._interface } @main.LookPath(ptr %48, i32 %49, ptr undef)
  %51 = extractvalue { %runtime._string, %runtime._interface } %50, 0
  %52 = extractvalue %runtime._string %51, 0
  call void @runtime.trackPointer(ptr %52, ptr undef) #14
  %53 = extractvalue { %runtime._string, %runtime._interface } %50, 1
  %54 = extractvalue %runtime._interface %53, 1
  call void @runtime.trackPointer(ptr %54, ptr undef) #14
  %55 = extractvalue { %runtime._string, %runtime._interface } %50, 1
  %56 = extractvalue %runtime._interface %55, 0
  %.not = icmp eq i32 %56, 0
  br i1 %.not, label %if.done8, label %if.then7

if.then7:                                         ; preds = %if.done6
  %57 = insertvalue { %runtime._string, %runtime._interface } zeroinitializer, %runtime._interface %55, 1
  ret { %runtime._string, %runtime._interface } %57

if.done8:                                         ; preds = %if.done6
  %58 = extractvalue { %runtime._string, %runtime._interface } %50, 0
  %59 = extractvalue %runtime._string %58, 0
  %60 = extractvalue %runtime._string %58, 1
  %61 = extractvalue %runtime._string %46, 0
  %62 = extractvalue %runtime._string %46, 1
  %63 = call %runtime._string @strings.TrimPrefix(ptr %59, i32 %60, ptr %61, i32 %62, ptr undef) #14
  %64 = extractvalue %runtime._string %63, 0
  call void @runtime.trackPointer(ptr %64, ptr undef) #14
  %65 = extractvalue %runtime._string %9, 0
  %66 = extractvalue %runtime._string %9, 1
  %67 = extractvalue %runtime._string %63, 0
  %68 = extractvalue %runtime._string %63, 1
  %69 = call %runtime._string @runtime.stringConcat(ptr %65, i32 %66, ptr %67, i32 %68, ptr undef) #14
  %70 = extractvalue %runtime._string %69, 0
  call void @runtime.trackPointer(ptr %70, ptr undef) #14
  %71 = insertvalue { %runtime._string, %runtime._interface } zeroinitializer, %runtime._string %69, 0
  %72 = insertvalue { %runtime._string, %runtime._interface } %71, %runtime._interface zeroinitializer, 1
  ret { %runtime._string, %runtime._interface } %72

lookup.throw:                                     ; preds = %cond.true
  call void @runtime.lookupPanic(ptr undef) #14
  unreachable
}

declare ptr @"interface:{Deadline:func:{}{named:time.Time,basic:bool},Done:func:{}{chan:struct:{}},Err:func:{}{named:error},Value:func:{interface:{}}{interface:{}}}.Done$invoke"(ptr, i32, ptr) #3

; Function Attrs: argmemonly nocallback nofree nosync nounwind willreturn
declare void @llvm.lifetime.start.p0(i64 immarg, ptr nocapture) #4

declare { i32, i1 } @runtime.tryChanSelect(ptr, ptr, i32, i32, ptr) #0

; Function Attrs: argmemonly nocallback nofree nosync nounwind willreturn
declare void @llvm.lifetime.end.p0(i64 immarg, ptr nocapture) #4

declare %runtime._interface @"interface:{Deadline:func:{}{named:time.Time,basic:bool},Done:func:{}{chan:struct:{}},Err:func:{}{named:error},Value:func:{interface:{}}{interface:{}}}.Err$invoke"(ptr, i32, ptr) #5

declare void @runtime.slicePanic(ptr) #0

; Function Attrs: nounwind
define linkonce_odr hidden { ptr, %runtime._interface } @"(*main.Cmd).stdin$thunk$thunk for func (*main.Cmd).stdin() (f *os.File, err error)"(ptr dereferenceable_or_null(168) %c, ptr %context) unnamed_addr #1 {
entry:
  %0 = call { ptr, %runtime._interface } @"(*main.Cmd).stdin"(ptr %c, ptr undef)
  %1 = extractvalue { ptr, %runtime._interface } %0, 0
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = extractvalue { ptr, %runtime._interface } %0, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret { ptr, %runtime._interface } %0
}

; Function Attrs: nounwind
define hidden { ptr, %runtime._interface } @"(*main.Cmd).stdin"(ptr dereferenceable_or_null(168) %c, ptr %context) unnamed_addr #1 {
entry:
  %c7 = call ptr @runtime.alloc(i32 4, ptr nonnull inttoptr (i32 67 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %c7, ptr undef) #14
  store ptr %c, ptr %c7, align 4
  call void @runtime.trackPointer(ptr %c, ptr undef) #14
  %0 = icmp eq ptr %c, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 4
  %.unpack = load i32, ptr %1, align 4
  %.elt75 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 4, i32 1
  %.unpack76 = load ptr, ptr %.elt75, align 4
  call void @runtime.trackPointer(ptr %.unpack76, ptr undef) #14
  %2 = icmp eq i32 %.unpack, 0
  br i1 %2, label %if.then, label %if.done2

if.then:                                          ; preds = %deref.next
  %3 = call { ptr, %runtime._interface } @os.Open(ptr nonnull @"main$string.64", i32 9, ptr undef) #14
  %4 = extractvalue { ptr, %runtime._interface } %3, 0
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  %5 = extractvalue { ptr, %runtime._interface } %3, 1
  %6 = extractvalue %runtime._interface %5, 1
  call void @runtime.trackPointer(ptr %6, ptr undef) #14
  %7 = extractvalue { ptr, %runtime._interface } %3, 0
  %8 = extractvalue { ptr, %runtime._interface } %3, 1
  %9 = extractvalue %runtime._interface %8, 0
  %.not113 = icmp eq i32 %9, 0
  br i1 %.not113, label %if.done, label %if.then1

if.then1:                                         ; preds = %if.then
  ret { ptr, %runtime._interface } %3

if.done:                                          ; preds = %if.then
  %10 = load ptr, ptr %c7, align 4
  call void @runtime.trackPointer(ptr %10, ptr undef) #14
  %11 = icmp eq ptr %10, null
  br i1 %11, label %gep.throw8, label %gep.next9

gep.next9:                                        ; preds = %if.done
  %12 = load ptr, ptr %c7, align 4
  call void @runtime.trackPointer(ptr %12, ptr undef) #14
  %13 = icmp eq ptr %12, null
  br i1 %13, label %gep.throw10, label %gep.next11

gep.next11:                                       ; preds = %gep.next9
  br i1 false, label %deref.throw12, label %deref.next13

deref.next13:                                     ; preds = %gep.next11
  %14 = getelementptr inbounds %main.Cmd, ptr %12, i32 0, i32 14
  %.unpack114 = load ptr, ptr %14, align 4
  %.elt115 = getelementptr inbounds %main.Cmd, ptr %12, i32 0, i32 14, i32 1
  %.unpack116 = load i32, ptr %.elt115, align 4
  %.elt117 = getelementptr inbounds %main.Cmd, ptr %12, i32 0, i32 14, i32 2
  %.unpack118 = load i32, ptr %.elt117, align 4
  call void @runtime.trackPointer(ptr %.unpack114, ptr undef) #14
  %varargs = call ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 133 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %varargs, ptr undef) #14
  call void @runtime.trackPointer(ptr %7, ptr undef) #14
  store i32 ptrtoint (ptr @"reflect/types.type:pointer:named:os.File" to i32), ptr %varargs, align 4
  %varargs.repack119 = getelementptr inbounds %runtime._interface, ptr %varargs, i32 0, i32 1
  store ptr %7, ptr %varargs.repack119, align 4
  %append.new = call { ptr, i32, i32 } @runtime.sliceAppend(ptr %.unpack114, ptr nonnull %varargs, i32 %.unpack116, i32 %.unpack118, i32 1, i32 8, ptr undef) #14
  %append.newPtr = extractvalue { ptr, i32, i32 } %append.new, 0
  call void @runtime.trackPointer(ptr %append.newPtr, ptr undef) #14
  br i1 false, label %store.throw, label %store.next

store.next:                                       ; preds = %deref.next13
  %15 = getelementptr inbounds %main.Cmd, ptr %10, i32 0, i32 14
  %append.newLen = extractvalue { ptr, i32, i32 } %append.new, 1
  %append.newCap = extractvalue { ptr, i32, i32 } %append.new, 2
  store ptr %append.newPtr, ptr %15, align 4
  %.repack121 = getelementptr inbounds %main.Cmd, ptr %10, i32 0, i32 14, i32 1
  store i32 %append.newLen, ptr %.repack121, align 4
  %.repack123 = getelementptr inbounds %main.Cmd, ptr %10, i32 0, i32 14, i32 2
  store i32 %append.newCap, ptr %.repack123, align 4
  ret { ptr, %runtime._interface } %3

if.done2:                                         ; preds = %deref.next
  %16 = load ptr, ptr %c7, align 4
  call void @runtime.trackPointer(ptr %16, ptr undef) #14
  %17 = icmp eq ptr %16, null
  br i1 %17, label %gep.throw14, label %gep.next15

gep.next15:                                       ; preds = %if.done2
  br i1 false, label %deref.throw16, label %deref.next17

deref.next17:                                     ; preds = %gep.next15
  %18 = getelementptr inbounds %main.Cmd, ptr %16, i32 0, i32 4
  %.unpack77 = load i32, ptr %18, align 4
  %.elt78 = getelementptr inbounds %main.Cmd, ptr %16, i32 0, i32 4, i32 1
  %.unpack79 = load ptr, ptr %.elt78, align 4
  call void @runtime.trackPointer(ptr %.unpack79, ptr undef) #14
  %typecode = call i1 @runtime.typeAssert(i32 %.unpack77, ptr nonnull @"reflect/types.typeid:pointer:named:os.File", ptr undef) #14
  br i1 %typecode, label %typeassert.ok, label %typeassert.next

typeassert.next:                                  ; preds = %typeassert.ok, %deref.next17
  %typeassert.value = phi ptr [ null, %deref.next17 ], [ %.unpack79, %typeassert.ok ]
  br i1 %typecode, label %if.then3, label %if.done4

typeassert.ok:                                    ; preds = %deref.next17
  br label %typeassert.next

if.then3:                                         ; preds = %typeassert.next
  %19 = insertvalue { ptr, %runtime._interface } zeroinitializer, ptr %typeassert.value, 0
  %20 = insertvalue { ptr, %runtime._interface } %19, %runtime._interface zeroinitializer, 1
  ret { ptr, %runtime._interface } %20

if.done4:                                         ; preds = %typeassert.next
  %pw = call ptr @runtime.alloc(i32 4, ptr nonnull inttoptr (i32 67 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %pw, ptr undef) #14
  %21 = call { ptr, ptr, %runtime._interface } @os.Pipe(ptr undef) #14
  %22 = extractvalue { ptr, ptr, %runtime._interface } %21, 0
  call void @runtime.trackPointer(ptr %22, ptr undef) #14
  %23 = extractvalue { ptr, ptr, %runtime._interface } %21, 1
  call void @runtime.trackPointer(ptr %23, ptr undef) #14
  %24 = extractvalue { ptr, ptr, %runtime._interface } %21, 2
  %25 = extractvalue %runtime._interface %24, 1
  call void @runtime.trackPointer(ptr %25, ptr undef) #14
  %26 = extractvalue { ptr, ptr, %runtime._interface } %21, 0
  %27 = extractvalue { ptr, ptr, %runtime._interface } %21, 1
  store ptr %27, ptr %pw, align 4
  %28 = extractvalue { ptr, ptr, %runtime._interface } %21, 2
  %29 = extractvalue %runtime._interface %28, 0
  %.not = icmp eq i32 %29, 0
  br i1 %.not, label %if.done6, label %if.then5

if.then5:                                         ; preds = %if.done4
  %30 = insertvalue { ptr, %runtime._interface } zeroinitializer, %runtime._interface %28, 1
  ret { ptr, %runtime._interface } %30

if.done6:                                         ; preds = %if.done4
  %31 = load ptr, ptr %c7, align 4
  call void @runtime.trackPointer(ptr %31, ptr undef) #14
  %32 = icmp eq ptr %31, null
  br i1 %32, label %gep.throw18, label %gep.next19

gep.next19:                                       ; preds = %if.done6
  %33 = getelementptr inbounds %main.Cmd, ptr %31, i32 0, i32 14
  %34 = load ptr, ptr %c7, align 4
  call void @runtime.trackPointer(ptr %34, ptr undef) #14
  %35 = icmp eq ptr %34, null
  br i1 %35, label %gep.throw20, label %gep.next21

gep.next21:                                       ; preds = %gep.next19
  br i1 false, label %deref.throw22, label %deref.next23

deref.next23:                                     ; preds = %gep.next21
  %36 = getelementptr inbounds %main.Cmd, ptr %34, i32 0, i32 14
  %.unpack80 = load ptr, ptr %36, align 4
  %.elt81 = getelementptr inbounds %main.Cmd, ptr %34, i32 0, i32 14, i32 1
  %.unpack82 = load i32, ptr %.elt81, align 4
  %.elt83 = getelementptr inbounds %main.Cmd, ptr %34, i32 0, i32 14, i32 2
  %.unpack84 = load i32, ptr %.elt83, align 4
  call void @runtime.trackPointer(ptr %.unpack80, ptr undef) #14
  %varargs24 = call ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 133 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %varargs24, ptr undef) #14
  call void @runtime.trackPointer(ptr %26, ptr undef) #14
  store i32 ptrtoint (ptr @"reflect/types.type:pointer:named:os.File" to i32), ptr %varargs24, align 4
  %varargs24.repack85 = getelementptr inbounds %runtime._interface, ptr %varargs24, i32 0, i32 1
  store ptr %26, ptr %varargs24.repack85, align 4
  %append.new31 = call { ptr, i32, i32 } @runtime.sliceAppend(ptr %.unpack80, ptr nonnull %varargs24, i32 %.unpack82, i32 %.unpack84, i32 1, i32 8, ptr undef) #14
  %append.newPtr32 = extractvalue { ptr, i32, i32 } %append.new31, 0
  call void @runtime.trackPointer(ptr %append.newPtr32, ptr undef) #14
  br i1 false, label %store.throw35, label %store.next36

store.next36:                                     ; preds = %deref.next23
  %append.newLen33 = extractvalue { ptr, i32, i32 } %append.new31, 1
  %append.newCap34 = extractvalue { ptr, i32, i32 } %append.new31, 2
  store ptr %append.newPtr32, ptr %33, align 4
  %.repack87 = getelementptr inbounds %main.Cmd, ptr %31, i32 0, i32 14, i32 1
  store i32 %append.newLen33, ptr %.repack87, align 4
  %.repack89 = getelementptr inbounds %main.Cmd, ptr %31, i32 0, i32 14, i32 2
  store i32 %append.newCap34, ptr %.repack89, align 4
  %37 = load ptr, ptr %c7, align 4
  call void @runtime.trackPointer(ptr %37, ptr undef) #14
  %38 = icmp eq ptr %37, null
  br i1 %38, label %gep.throw37, label %gep.next38

gep.next38:                                       ; preds = %store.next36
  %39 = getelementptr inbounds %main.Cmd, ptr %37, i32 0, i32 15
  %40 = load ptr, ptr %c7, align 4
  call void @runtime.trackPointer(ptr %40, ptr undef) #14
  %41 = icmp eq ptr %40, null
  br i1 %41, label %gep.throw39, label %gep.next40

gep.next40:                                       ; preds = %gep.next38
  br i1 false, label %deref.throw41, label %deref.next42

deref.next42:                                     ; preds = %gep.next40
  %42 = getelementptr inbounds %main.Cmd, ptr %40, i32 0, i32 15
  %.unpack91 = load ptr, ptr %42, align 4
  %.elt92 = getelementptr inbounds %main.Cmd, ptr %40, i32 0, i32 15, i32 1
  %.unpack93 = load i32, ptr %.elt92, align 4
  %.elt94 = getelementptr inbounds %main.Cmd, ptr %40, i32 0, i32 15, i32 2
  %.unpack95 = load i32, ptr %.elt94, align 4
  call void @runtime.trackPointer(ptr %.unpack91, ptr undef) #14
  %43 = load ptr, ptr %pw, align 4
  call void @runtime.trackPointer(ptr %43, ptr undef) #14
  %varargs43 = call ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 133 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %varargs43, ptr undef) #14
  call void @runtime.trackPointer(ptr %43, ptr undef) #14
  store i32 ptrtoint (ptr @"reflect/types.type:pointer:named:os.File" to i32), ptr %varargs43, align 4
  %varargs43.repack96 = getelementptr inbounds %runtime._interface, ptr %varargs43, i32 0, i32 1
  store ptr %43, ptr %varargs43.repack96, align 4
  %append.new50 = call { ptr, i32, i32 } @runtime.sliceAppend(ptr %.unpack91, ptr nonnull %varargs43, i32 %.unpack93, i32 %.unpack95, i32 1, i32 8, ptr undef) #14
  %append.newPtr51 = extractvalue { ptr, i32, i32 } %append.new50, 0
  call void @runtime.trackPointer(ptr %append.newPtr51, ptr undef) #14
  br i1 false, label %store.throw54, label %store.next55

store.next55:                                     ; preds = %deref.next42
  %append.newLen52 = extractvalue { ptr, i32, i32 } %append.new50, 1
  %append.newCap53 = extractvalue { ptr, i32, i32 } %append.new50, 2
  store ptr %append.newPtr51, ptr %39, align 4
  %.repack98 = getelementptr inbounds %main.Cmd, ptr %37, i32 0, i32 15, i32 1
  store i32 %append.newLen52, ptr %.repack98, align 4
  %.repack100 = getelementptr inbounds %main.Cmd, ptr %37, i32 0, i32 15, i32 2
  store i32 %append.newCap53, ptr %.repack100, align 4
  %44 = load ptr, ptr %c7, align 4
  call void @runtime.trackPointer(ptr %44, ptr undef) #14
  %45 = icmp eq ptr %44, null
  br i1 %45, label %gep.throw56, label %gep.next57

gep.next57:                                       ; preds = %store.next55
  %46 = load ptr, ptr %c7, align 4
  call void @runtime.trackPointer(ptr %46, ptr undef) #14
  %47 = icmp eq ptr %46, null
  br i1 %47, label %gep.throw58, label %gep.next59

gep.next59:                                       ; preds = %gep.next57
  br i1 false, label %deref.throw60, label %deref.next61

deref.next61:                                     ; preds = %gep.next59
  %48 = getelementptr inbounds %main.Cmd, ptr %46, i32 0, i32 16
  %.unpack102 = load ptr, ptr %48, align 4
  %.elt103 = getelementptr inbounds %main.Cmd, ptr %46, i32 0, i32 16, i32 1
  %.unpack104 = load i32, ptr %.elt103, align 4
  %.elt105 = getelementptr inbounds %main.Cmd, ptr %46, i32 0, i32 16, i32 2
  %.unpack106 = load i32, ptr %.elt105, align 4
  call void @runtime.trackPointer(ptr %.unpack102, ptr undef) #14
  %49 = call ptr @runtime.alloc(i32 8, ptr null, ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %49, ptr undef) #14
  store ptr %pw, ptr %49, align 4
  %50 = getelementptr inbounds { ptr, ptr }, ptr %49, i32 0, i32 1
  store ptr %c7, ptr %50, align 4
  call void @runtime.trackPointer(ptr nonnull %49, ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull @"(*main.Cmd).stdin$1", ptr undef) #14
  %varargs62 = call ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 197 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %varargs62, ptr undef) #14
  store ptr %49, ptr %varargs62, align 4
  %varargs62.repack107 = getelementptr inbounds { ptr, ptr }, ptr %varargs62, i32 0, i32 1
  store ptr @"(*main.Cmd).stdin$1", ptr %varargs62.repack107, align 4
  %append.new69 = call { ptr, i32, i32 } @runtime.sliceAppend(ptr %.unpack102, ptr nonnull %varargs62, i32 %.unpack104, i32 %.unpack106, i32 1, i32 8, ptr undef) #14
  %append.newPtr70 = extractvalue { ptr, i32, i32 } %append.new69, 0
  call void @runtime.trackPointer(ptr %append.newPtr70, ptr undef) #14
  br i1 false, label %store.throw73, label %store.next74

store.next74:                                     ; preds = %deref.next61
  %51 = getelementptr inbounds %main.Cmd, ptr %44, i32 0, i32 16
  %append.newLen71 = extractvalue { ptr, i32, i32 } %append.new69, 1
  %append.newCap72 = extractvalue { ptr, i32, i32 } %append.new69, 2
  store ptr %append.newPtr70, ptr %51, align 4
  %.repack109 = getelementptr inbounds %main.Cmd, ptr %44, i32 0, i32 16, i32 1
  store i32 %append.newLen71, ptr %.repack109, align 4
  %.repack111 = getelementptr inbounds %main.Cmd, ptr %44, i32 0, i32 16, i32 2
  store i32 %append.newCap72, ptr %.repack111, align 4
  %52 = insertvalue { ptr, %runtime._interface } zeroinitializer, ptr %26, 0
  %53 = insertvalue { ptr, %runtime._interface } %52, %runtime._interface zeroinitializer, 1
  ret { ptr, %runtime._interface } %53

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable

gep.throw8:                                       ; preds = %if.done
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

gep.throw10:                                      ; preds = %gep.next9
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw12:                                    ; preds = %gep.next11
  unreachable

store.throw:                                      ; preds = %deref.next13
  unreachable

gep.throw14:                                      ; preds = %if.done2
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw16:                                    ; preds = %gep.next15
  unreachable

gep.throw18:                                      ; preds = %if.done6
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

gep.throw20:                                      ; preds = %gep.next19
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw22:                                    ; preds = %gep.next21
  unreachable

store.throw35:                                    ; preds = %deref.next23
  unreachable

gep.throw37:                                      ; preds = %store.next36
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

gep.throw39:                                      ; preds = %gep.next38
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw41:                                    ; preds = %gep.next40
  unreachable

store.throw54:                                    ; preds = %deref.next42
  unreachable

gep.throw56:                                      ; preds = %store.next55
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

gep.throw58:                                      ; preds = %gep.next57
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw60:                                    ; preds = %gep.next59
  unreachable

store.throw73:                                    ; preds = %deref.next61
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { ptr, %runtime._interface } @"(*main.Cmd).stdout$thunk$thunk for func (*main.Cmd).stdout() (f *os.File, err error)"(ptr dereferenceable_or_null(168) %c, ptr %context) unnamed_addr #1 {
entry:
  %0 = call { ptr, %runtime._interface } @"(*main.Cmd).stdout"(ptr %c, ptr undef)
  %1 = extractvalue { ptr, %runtime._interface } %0, 0
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = extractvalue { ptr, %runtime._interface } %0, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret { ptr, %runtime._interface } %0
}

; Function Attrs: nounwind
define hidden { ptr, %runtime._interface } @"(*main.Cmd).stdout"(ptr dereferenceable_or_null(168) %c, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %c, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 5
  %.unpack = load i32, ptr %1, align 4
  %.elt1 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 5, i32 1
  %.unpack2 = load ptr, ptr %.elt1, align 4
  call void @runtime.trackPointer(ptr %.unpack2, ptr undef) #14
  %2 = call { ptr, %runtime._interface } @"(*main.Cmd).writerDescriptor"(ptr nonnull %c, i32 %.unpack, ptr %.unpack2, ptr undef)
  %3 = extractvalue { ptr, %runtime._interface } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %4 = extractvalue { ptr, %runtime._interface } %2, 1
  %5 = extractvalue %runtime._interface %4, 1
  call void @runtime.trackPointer(ptr %5, ptr undef) #14
  ret { ptr, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { ptr, %runtime._interface } @"(*main.Cmd).stderr$thunk$thunk for func (*main.Cmd).stderr() (f *os.File, err error)"(ptr dereferenceable_or_null(168) %c, ptr %context) unnamed_addr #1 {
entry:
  %0 = call { ptr, %runtime._interface } @"(*main.Cmd).stderr"(ptr %c, ptr undef)
  %1 = extractvalue { ptr, %runtime._interface } %0, 0
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = extractvalue { ptr, %runtime._interface } %0, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret { ptr, %runtime._interface } %0
}

; Function Attrs: nounwind
define hidden { ptr, %runtime._interface } @"(*main.Cmd).stderr"(ptr dereferenceable_or_null(168) %c, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %c, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 6
  %.unpack = load i32, ptr %1, align 4
  %.elt17 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 6, i32 1
  %.unpack18 = load ptr, ptr %.elt17, align 4
  call void @runtime.trackPointer(ptr %.unpack18, ptr undef) #14
  %.not = icmp eq i32 %.unpack, 0
  br i1 %.not, label %if.done, label %cond.true

cond.true:                                        ; preds = %deref.next
  br i1 false, label %gep.throw1, label %gep.next2

gep.next2:                                        ; preds = %cond.true
  br i1 false, label %deref.throw3, label %deref.next4

deref.next4:                                      ; preds = %gep.next2
  %2 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 6
  %.unpack22 = load i32, ptr %2, align 4
  %.elt23 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 6, i32 1
  %.unpack24 = load ptr, ptr %.elt23, align 4
  call void @runtime.trackPointer(ptr %.unpack24, ptr undef) #14
  br i1 false, label %gep.throw5, label %gep.next6

gep.next6:                                        ; preds = %deref.next4
  br i1 false, label %deref.throw7, label %deref.next8

deref.next8:                                      ; preds = %gep.next6
  %3 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 5
  %.unpack25 = load i32, ptr %3, align 4
  %.elt26 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 5, i32 1
  %.unpack27 = load ptr, ptr %.elt26, align 4
  call void @runtime.trackPointer(ptr %.unpack27, ptr undef) #14
  %4 = call i1 @main.interfaceEqual(i32 %.unpack22, ptr %.unpack24, i32 %.unpack25, ptr %.unpack27, ptr undef)
  br i1 %4, label %if.then, label %if.done

if.then:                                          ; preds = %deref.next8
  br i1 false, label %gep.throw9, label %gep.next10

gep.next10:                                       ; preds = %if.then
  br i1 false, label %deref.throw11, label %deref.next12

deref.next12:                                     ; preds = %gep.next10
  %5 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 13
  %.unpack28 = load ptr, ptr %5, align 4
  %.elt29 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 13, i32 1
  %.unpack30 = load i32, ptr %.elt29, align 4
  call void @runtime.trackPointer(ptr %.unpack28, ptr undef) #14
  %6 = icmp ult i32 %.unpack30, 2
  br i1 %6, label %lookup.throw, label %lookup.next

lookup.next:                                      ; preds = %deref.next12
  %7 = getelementptr inbounds ptr, ptr %.unpack28, i32 1
  %8 = load ptr, ptr %7, align 4
  call void @runtime.trackPointer(ptr %8, ptr undef) #14
  %9 = insertvalue { ptr, %runtime._interface } zeroinitializer, ptr %8, 0
  %10 = insertvalue { ptr, %runtime._interface } %9, %runtime._interface zeroinitializer, 1
  ret { ptr, %runtime._interface } %10

if.done:                                          ; preds = %deref.next8, %deref.next
  br i1 false, label %gep.throw13, label %gep.next14

gep.next14:                                       ; preds = %if.done
  br i1 false, label %deref.throw15, label %deref.next16

deref.next16:                                     ; preds = %gep.next14
  %11 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 6
  %.unpack19 = load i32, ptr %11, align 4
  %.elt20 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 6, i32 1
  %.unpack21 = load ptr, ptr %.elt20, align 4
  call void @runtime.trackPointer(ptr %.unpack21, ptr undef) #14
  %12 = call { ptr, %runtime._interface } @"(*main.Cmd).writerDescriptor"(ptr nonnull %c, i32 %.unpack19, ptr %.unpack21, ptr undef)
  %13 = extractvalue { ptr, %runtime._interface } %12, 0
  call void @runtime.trackPointer(ptr %13, ptr undef) #14
  %14 = extractvalue { ptr, %runtime._interface } %12, 1
  %15 = extractvalue %runtime._interface %14, 1
  call void @runtime.trackPointer(ptr %15, ptr undef) #14
  ret { ptr, %runtime._interface } %12

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable

gep.throw1:                                       ; preds = %cond.true
  unreachable

deref.throw3:                                     ; preds = %gep.next2
  unreachable

gep.throw5:                                       ; preds = %deref.next4
  unreachable

deref.throw7:                                     ; preds = %gep.next6
  unreachable

gep.throw9:                                       ; preds = %if.then
  unreachable

deref.throw11:                                    ; preds = %gep.next10
  unreachable

lookup.throw:                                     ; preds = %deref.next12
  call void @runtime.lookupPanic(ptr undef) #14
  unreachable

gep.throw13:                                      ; preds = %if.done
  unreachable

deref.throw15:                                    ; preds = %gep.next14
  unreachable
}

declare void @runtime.lookupPanic(ptr) #0

declare { ptr, i32, i32 } @runtime.sliceAppend(ptr, ptr nocapture readonly, i32, i32, i32, i32, ptr) #0

; Function Attrs: nounwind
define hidden { ptr, i32, i32 } @"(*main.Cmd).argv"(ptr dereferenceable_or_null(168) %c, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %c, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 1
  %.unpack = load ptr, ptr %1, align 4
  %.elt9 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 1, i32 1
  %.unpack10 = load i32, ptr %.elt9, align 4
  call void @runtime.trackPointer(ptr %.unpack, ptr undef) #14
  %2 = icmp sgt i32 %.unpack10, 0
  br i1 %2, label %if.then, label %if.done

if.then:                                          ; preds = %deref.next
  br i1 false, label %gep.throw1, label %gep.next2

gep.next2:                                        ; preds = %if.then
  br i1 false, label %deref.throw3, label %deref.next4

deref.next4:                                      ; preds = %gep.next2
  %3 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 1
  %.unpack18 = load ptr, ptr %3, align 4
  %4 = insertvalue { ptr, i32, i32 } undef, ptr %.unpack18, 0
  %.elt19 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 1, i32 1
  %.unpack20 = load i32, ptr %.elt19, align 4
  %5 = insertvalue { ptr, i32, i32 } %4, i32 %.unpack20, 1
  %.elt21 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 1, i32 2
  %.unpack22 = load i32, ptr %.elt21, align 4
  %6 = insertvalue { ptr, i32, i32 } %5, i32 %.unpack22, 2
  call void @runtime.trackPointer(ptr %.unpack18, ptr undef) #14
  ret { ptr, i32, i32 } %6

if.done:                                          ; preds = %deref.next
  %slicelit = call ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 69 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %slicelit, ptr undef) #14
  br i1 false, label %gep.throw5, label %gep.next6

gep.next6:                                        ; preds = %if.done
  br i1 false, label %deref.throw7, label %deref.next8

deref.next8:                                      ; preds = %gep.next6
  %.unpack13 = load ptr, ptr %c, align 4
  %.elt14 = getelementptr inbounds %runtime._string, ptr %c, i32 0, i32 1
  %.unpack15 = load i32, ptr %.elt14, align 4
  call void @runtime.trackPointer(ptr %.unpack13, ptr undef) #14
  store ptr %.unpack13, ptr %slicelit, align 4
  %slicelit.repack16 = getelementptr inbounds %runtime._string, ptr %slicelit, i32 0, i32 1
  store i32 %.unpack15, ptr %slicelit.repack16, align 4
  %7 = insertvalue { ptr, i32, i32 } undef, ptr %slicelit, 0
  %8 = insertvalue { ptr, i32, i32 } %7, i32 1, 1
  %9 = insertvalue { ptr, i32, i32 } %8, i32 1, 2
  ret { ptr, i32, i32 } %9

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable

gep.throw1:                                       ; preds = %if.then
  unreachable

deref.throw3:                                     ; preds = %gep.next2
  unreachable

gep.throw5:                                       ; preds = %if.done
  unreachable

deref.throw7:                                     ; preds = %gep.next6
  unreachable
}

declare { ptr, %runtime._interface } @os.StartProcess(ptr, i32, ptr, i32, i32, ptr dereferenceable_or_null(36), ptr) #0

declare void @runtime.chanMakePanic(ptr) #0

declare ptr @runtime.chanMake(i32, i32, ptr) #0

; Function Attrs: nounwind
define internal void @"(*main.Cmd).Start$1"(ptr %fn.context, ptr %fn.funcptr, ptr %context) unnamed_addr #1 {
entry:
  %chan.blockedList = alloca %runtime.channelBlockedList, align 8
  %chan.value = alloca %runtime._interface, align 8
  %0 = load ptr, ptr %context, align 4
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = icmp eq ptr %fn.funcptr, null
  br i1 %1, label %fpcall.throw, label %fpcall.next

fpcall.next:                                      ; preds = %entry
  %2 = call %runtime._interface %fn.funcptr(ptr %fn.context) #14
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  call void @llvm.lifetime.start.p0(i64 8, ptr nonnull %chan.value)
  %.elt = extractvalue %runtime._interface %2, 0
  store i32 %.elt, ptr %chan.value, align 8
  %chan.value.repack1 = getelementptr inbounds %runtime._interface, ptr %chan.value, i32 0, i32 1
  %.elt2 = extractvalue %runtime._interface %2, 1
  store ptr %.elt2, ptr %chan.value.repack1, align 4
  call void @llvm.lifetime.start.p0(i64 24, ptr nonnull %chan.blockedList)
  call void @runtime.chanSend(ptr %0, ptr nonnull %chan.value, ptr nonnull %chan.blockedList, ptr undef) #14
  call void @llvm.lifetime.end.p0(i64 24, ptr nonnull %chan.blockedList)
  call void @llvm.lifetime.end.p0(i64 8, ptr nonnull %chan.value)
  ret void

fpcall.throw:                                     ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable
}

declare void @runtime.deadlock(ptr) #0

; Function Attrs: nounwind
define linkonce_odr void @"(*main.Cmd).Start$1$gowrapper"(ptr %0) unnamed_addr #6 {
entry:
  %1 = load ptr, ptr %0, align 4
  %2 = getelementptr inbounds { ptr, ptr, ptr }, ptr %0, i32 0, i32 1
  %3 = load ptr, ptr %2, align 4
  %4 = getelementptr inbounds { ptr, ptr, ptr }, ptr %0, i32 0, i32 2
  %5 = load ptr, ptr %4, align 4
  call void @"(*main.Cmd).Start$1"(ptr %1, ptr %3, ptr %5)
  call void @runtime.deadlock(ptr undef) #14
  unreachable
}

declare void @"internal/task.start"(i32, ptr, i32, ptr) #0

; Function Attrs: nounwind
define hidden ptr @"(*main.Cmd).watchCtx"(ptr dereferenceable_or_null(168) %c, ptr %context) unnamed_addr #1 {
entry:
  %c1 = call ptr @runtime.alloc(i32 4, ptr nonnull inttoptr (i32 67 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %c1, ptr undef) #14
  store ptr %c, ptr %c1, align 4
  call void @runtime.trackPointer(ptr %c, ptr undef) #14
  %0 = icmp eq ptr %c, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 11
  %.unpack = load i32, ptr %1, align 4
  %.elt2 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 11, i32 1
  %.unpack3 = load ptr, ptr %.elt2, align 4
  call void @runtime.trackPointer(ptr %.unpack3, ptr undef) #14
  %2 = icmp eq i32 %.unpack, 0
  br i1 %2, label %if.then, label %if.done

if.then:                                          ; preds = %deref.next
  ret ptr null

if.done:                                          ; preds = %deref.next
  %errc = call ptr @runtime.alloc(i32 4, ptr nonnull inttoptr (i32 67 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %errc, ptr undef) #14
  %3 = call ptr @runtime.chanMake(i32 8, i32 0, ptr undef) #14
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  store ptr %3, ptr %errc, align 4
  %4 = call ptr @runtime.alloc(i32 8, ptr null, ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %4, ptr undef) #14
  store ptr %errc, ptr %4, align 4
  %5 = getelementptr inbounds { ptr, ptr }, ptr %4, i32 0, i32 1
  store ptr %c1, ptr %5, align 4
  call void @runtime.trackPointer(ptr nonnull %4, ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull @"(*main.Cmd).watchCtx$1", ptr undef) #14
  call void @"internal/task.start"(i32 ptrtoint (ptr @"(*main.Cmd).watchCtx$1$gowrapper" to i32), ptr nonnull %4, i32 16384, ptr undef) #14
  %6 = load ptr, ptr %errc, align 4
  call void @runtime.trackPointer(ptr %6, ptr undef) #14
  ret ptr %6

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

declare void @runtime.chanSend(ptr dereferenceable_or_null(32), ptr, ptr dereferenceable_or_null(24), ptr) #0

; Function Attrs: nounwind
define hidden { %runtime._interface, %runtime._interface } @"(*main.Cmd).StderrPipe"(ptr dereferenceable_or_null(168) %c, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %c, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 6
  %.unpack = load i32, ptr %1, align 4
  %.elt38 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 6, i32 1
  %.unpack39 = load ptr, ptr %.elt38, align 4
  call void @runtime.trackPointer(ptr %.unpack39, ptr undef) #14
  %.not = icmp eq i32 %.unpack, 0
  br i1 %.not, label %if.done, label %if.then

if.then:                                          ; preds = %deref.next
  %2 = call %runtime._interface @errors.New(ptr nonnull @"main$string.17", i32 24, ptr undef) #14
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %4 = insertvalue { %runtime._interface, %runtime._interface } zeroinitializer, %runtime._interface %2, 1
  ret { %runtime._interface, %runtime._interface } %4

if.done:                                          ; preds = %deref.next
  br i1 false, label %gep.throw5, label %gep.next6

gep.next6:                                        ; preds = %if.done
  br i1 false, label %deref.throw7, label %deref.next8

deref.next8:                                      ; preds = %gep.next6
  %5 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 9
  %6 = load ptr, ptr %5, align 4
  call void @runtime.trackPointer(ptr %6, ptr undef) #14
  %.not40 = icmp eq ptr %6, null
  br i1 %.not40, label %if.done2, label %if.then1

if.then1:                                         ; preds = %deref.next8
  %7 = call %runtime._interface @errors.New(ptr nonnull @"main$string.18", i32 38, ptr undef) #14
  %8 = extractvalue %runtime._interface %7, 1
  call void @runtime.trackPointer(ptr %8, ptr undef) #14
  %9 = insertvalue { %runtime._interface, %runtime._interface } zeroinitializer, %runtime._interface %7, 1
  ret { %runtime._interface, %runtime._interface } %9

if.done2:                                         ; preds = %deref.next8
  %10 = call { ptr, ptr, %runtime._interface } @os.Pipe(ptr undef) #14
  %11 = extractvalue { ptr, ptr, %runtime._interface } %10, 0
  call void @runtime.trackPointer(ptr %11, ptr undef) #14
  %12 = extractvalue { ptr, ptr, %runtime._interface } %10, 1
  call void @runtime.trackPointer(ptr %12, ptr undef) #14
  %13 = extractvalue { ptr, ptr, %runtime._interface } %10, 2
  %14 = extractvalue %runtime._interface %13, 1
  call void @runtime.trackPointer(ptr %14, ptr undef) #14
  %15 = extractvalue { ptr, ptr, %runtime._interface } %10, 0
  %16 = extractvalue { ptr, ptr, %runtime._interface } %10, 1
  %17 = extractvalue { ptr, ptr, %runtime._interface } %10, 2
  %18 = extractvalue %runtime._interface %17, 0
  %.not41 = icmp eq i32 %18, 0
  br i1 %.not41, label %if.done4, label %if.then3

if.then3:                                         ; preds = %if.done2
  %19 = insertvalue { %runtime._interface, %runtime._interface } zeroinitializer, %runtime._interface %17, 1
  ret { %runtime._interface, %runtime._interface } %19

if.done4:                                         ; preds = %if.done2
  br i1 false, label %gep.throw9, label %gep.next10

gep.next10:                                       ; preds = %if.done4
  call void @runtime.trackPointer(ptr %16, ptr undef) #14
  br i1 false, label %store.throw, label %store.next

store.next:                                       ; preds = %gep.next10
  %20 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 6
  store i32 ptrtoint (ptr @"reflect/types.type:pointer:named:os.File" to i32), ptr %20, align 4
  %.repack42 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 6, i32 1
  store ptr %16, ptr %.repack42, align 4
  br i1 false, label %gep.throw11, label %gep.next12

gep.next12:                                       ; preds = %store.next
  %21 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14
  br i1 false, label %gep.throw13, label %gep.next14

gep.next14:                                       ; preds = %gep.next12
  br i1 false, label %deref.throw15, label %deref.next16

deref.next16:                                     ; preds = %gep.next14
  %22 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14
  %.unpack44 = load ptr, ptr %22, align 4
  %.elt45 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 1
  %.unpack46 = load i32, ptr %.elt45, align 4
  %.elt47 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 2
  %.unpack48 = load i32, ptr %.elt47, align 4
  call void @runtime.trackPointer(ptr %.unpack44, ptr undef) #14
  %varargs = call ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 133 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %varargs, ptr undef) #14
  call void @runtime.trackPointer(ptr %16, ptr undef) #14
  store i32 ptrtoint (ptr @"reflect/types.type:pointer:named:os.File" to i32), ptr %varargs, align 4
  %varargs.repack49 = getelementptr inbounds %runtime._interface, ptr %varargs, i32 0, i32 1
  store ptr %16, ptr %varargs.repack49, align 4
  %append.new = call { ptr, i32, i32 } @runtime.sliceAppend(ptr %.unpack44, ptr nonnull %varargs, i32 %.unpack46, i32 %.unpack48, i32 1, i32 8, ptr undef) #14
  %append.newPtr = extractvalue { ptr, i32, i32 } %append.new, 0
  call void @runtime.trackPointer(ptr %append.newPtr, ptr undef) #14
  br i1 false, label %store.throw17, label %store.next18

store.next18:                                     ; preds = %deref.next16
  %append.newLen = extractvalue { ptr, i32, i32 } %append.new, 1
  %append.newCap = extractvalue { ptr, i32, i32 } %append.new, 2
  store ptr %append.newPtr, ptr %21, align 4
  %.repack51 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 1
  store i32 %append.newLen, ptr %.repack51, align 4
  %.repack53 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 2
  store i32 %append.newCap, ptr %.repack53, align 4
  br i1 false, label %gep.throw19, label %gep.next20

gep.next20:                                       ; preds = %store.next18
  br i1 false, label %gep.throw21, label %gep.next22

gep.next22:                                       ; preds = %gep.next20
  br i1 false, label %deref.throw23, label %deref.next24

deref.next24:                                     ; preds = %gep.next22
  %23 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15
  %.unpack55 = load ptr, ptr %23, align 4
  %.elt56 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 1
  %.unpack57 = load i32, ptr %.elt56, align 4
  %.elt58 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 2
  %.unpack59 = load i32, ptr %.elt58, align 4
  call void @runtime.trackPointer(ptr %.unpack55, ptr undef) #14
  %varargs25 = call ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 133 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %varargs25, ptr undef) #14
  call void @runtime.trackPointer(ptr %15, ptr undef) #14
  store i32 ptrtoint (ptr @"reflect/types.type:pointer:named:os.File" to i32), ptr %varargs25, align 4
  %varargs25.repack60 = getelementptr inbounds %runtime._interface, ptr %varargs25, i32 0, i32 1
  store ptr %15, ptr %varargs25.repack60, align 4
  %append.new32 = call { ptr, i32, i32 } @runtime.sliceAppend(ptr %.unpack55, ptr nonnull %varargs25, i32 %.unpack57, i32 %.unpack59, i32 1, i32 8, ptr undef) #14
  %append.newPtr33 = extractvalue { ptr, i32, i32 } %append.new32, 0
  call void @runtime.trackPointer(ptr %append.newPtr33, ptr undef) #14
  br i1 false, label %store.throw36, label %store.next37

store.next37:                                     ; preds = %deref.next24
  %24 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15
  %append.newLen34 = extractvalue { ptr, i32, i32 } %append.new32, 1
  %append.newCap35 = extractvalue { ptr, i32, i32 } %append.new32, 2
  store ptr %append.newPtr33, ptr %24, align 4
  %.repack62 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 1
  store i32 %append.newLen34, ptr %.repack62, align 4
  %.repack64 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 2
  store i32 %append.newCap35, ptr %.repack64, align 4
  %25 = insertvalue %runtime._interface { i32 ptrtoint (ptr @"reflect/types.type:pointer:named:os.File" to i32), ptr undef }, ptr %15, 1
  call void @runtime.trackPointer(ptr %15, ptr undef) #14
  %26 = insertvalue { %runtime._interface, %runtime._interface } zeroinitializer, %runtime._interface %25, 0
  %27 = insertvalue { %runtime._interface, %runtime._interface } %26, %runtime._interface zeroinitializer, 1
  ret { %runtime._interface, %runtime._interface } %27

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable

gep.throw5:                                       ; preds = %if.done
  unreachable

deref.throw7:                                     ; preds = %gep.next6
  unreachable

gep.throw9:                                       ; preds = %if.done4
  unreachable

store.throw:                                      ; preds = %gep.next10
  unreachable

gep.throw11:                                      ; preds = %store.next
  unreachable

gep.throw13:                                      ; preds = %gep.next12
  unreachable

deref.throw15:                                    ; preds = %gep.next14
  unreachable

store.throw17:                                    ; preds = %deref.next16
  unreachable

gep.throw19:                                      ; preds = %store.next18
  unreachable

gep.throw21:                                      ; preds = %gep.next20
  unreachable

deref.throw23:                                    ; preds = %gep.next22
  unreachable

store.throw36:                                    ; preds = %deref.next24
  unreachable
}

declare { ptr, ptr, %runtime._interface } @os.Pipe(ptr) #0

declare i1 @"interface:{Close:func:{}{named:error},Read:func:{slice:basic:uint8}{basic:int,named:error},ReadAt:func:{slice:basic:uint8,basic:int64}{basic:int,named:error},Seek:func:{basic:int64,basic:int}{basic:int64,named:error},Write:func:{slice:basic:uint8}{basic:int,named:error}}.$typeassert"(i32) #7

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

; Function Attrs: nounwind
define hidden { %runtime._interface, %runtime._interface } @"(*main.Cmd).StdinPipe"(ptr dereferenceable_or_null(168) %c, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %c, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 4
  %.unpack = load i32, ptr %1, align 4
  %.elt40 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 4, i32 1
  %.unpack41 = load ptr, ptr %.elt40, align 4
  call void @runtime.trackPointer(ptr %.unpack41, ptr undef) #14
  %.not = icmp eq i32 %.unpack, 0
  br i1 %.not, label %if.done, label %if.then

if.then:                                          ; preds = %deref.next
  %2 = call %runtime._interface @errors.New(ptr nonnull @"main$string.24", i32 23, ptr undef) #14
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %4 = insertvalue { %runtime._interface, %runtime._interface } zeroinitializer, %runtime._interface %2, 1
  ret { %runtime._interface, %runtime._interface } %4

if.done:                                          ; preds = %deref.next
  br i1 false, label %gep.throw5, label %gep.next6

gep.next6:                                        ; preds = %if.done
  br i1 false, label %deref.throw7, label %deref.next8

deref.next8:                                      ; preds = %gep.next6
  %5 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 9
  %6 = load ptr, ptr %5, align 4
  call void @runtime.trackPointer(ptr %6, ptr undef) #14
  %.not42 = icmp eq ptr %6, null
  br i1 %.not42, label %if.done2, label %if.then1

if.then1:                                         ; preds = %deref.next8
  %7 = call %runtime._interface @errors.New(ptr nonnull @"main$string.25", i32 37, ptr undef) #14
  %8 = extractvalue %runtime._interface %7, 1
  call void @runtime.trackPointer(ptr %8, ptr undef) #14
  %9 = insertvalue { %runtime._interface, %runtime._interface } zeroinitializer, %runtime._interface %7, 1
  ret { %runtime._interface, %runtime._interface } %9

if.done2:                                         ; preds = %deref.next8
  %10 = call { ptr, ptr, %runtime._interface } @os.Pipe(ptr undef) #14
  %11 = extractvalue { ptr, ptr, %runtime._interface } %10, 0
  call void @runtime.trackPointer(ptr %11, ptr undef) #14
  %12 = extractvalue { ptr, ptr, %runtime._interface } %10, 1
  call void @runtime.trackPointer(ptr %12, ptr undef) #14
  %13 = extractvalue { ptr, ptr, %runtime._interface } %10, 2
  %14 = extractvalue %runtime._interface %13, 1
  call void @runtime.trackPointer(ptr %14, ptr undef) #14
  %15 = extractvalue { ptr, ptr, %runtime._interface } %10, 0
  %16 = extractvalue { ptr, ptr, %runtime._interface } %10, 1
  %17 = extractvalue { ptr, ptr, %runtime._interface } %10, 2
  %18 = extractvalue %runtime._interface %17, 0
  %.not43 = icmp eq i32 %18, 0
  br i1 %.not43, label %if.done4, label %if.then3

if.then3:                                         ; preds = %if.done2
  %19 = insertvalue { %runtime._interface, %runtime._interface } zeroinitializer, %runtime._interface %17, 1
  ret { %runtime._interface, %runtime._interface } %19

if.done4:                                         ; preds = %if.done2
  br i1 false, label %gep.throw9, label %gep.next10

gep.next10:                                       ; preds = %if.done4
  call void @runtime.trackPointer(ptr %15, ptr undef) #14
  br i1 false, label %store.throw, label %store.next

store.next:                                       ; preds = %gep.next10
  %20 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 4
  store i32 ptrtoint (ptr @"reflect/types.type:pointer:named:os.File" to i32), ptr %20, align 4
  %.repack44 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 4, i32 1
  store ptr %15, ptr %.repack44, align 4
  br i1 false, label %gep.throw11, label %gep.next12

gep.next12:                                       ; preds = %store.next
  %21 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14
  br i1 false, label %gep.throw13, label %gep.next14

gep.next14:                                       ; preds = %gep.next12
  br i1 false, label %deref.throw15, label %deref.next16

deref.next16:                                     ; preds = %gep.next14
  %22 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14
  %.unpack46 = load ptr, ptr %22, align 4
  %.elt47 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 1
  %.unpack48 = load i32, ptr %.elt47, align 4
  %.elt49 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 2
  %.unpack50 = load i32, ptr %.elt49, align 4
  call void @runtime.trackPointer(ptr %.unpack46, ptr undef) #14
  %varargs = call ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 133 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %varargs, ptr undef) #14
  call void @runtime.trackPointer(ptr %15, ptr undef) #14
  store i32 ptrtoint (ptr @"reflect/types.type:pointer:named:os.File" to i32), ptr %varargs, align 4
  %varargs.repack51 = getelementptr inbounds %runtime._interface, ptr %varargs, i32 0, i32 1
  store ptr %15, ptr %varargs.repack51, align 4
  %append.new = call { ptr, i32, i32 } @runtime.sliceAppend(ptr %.unpack46, ptr nonnull %varargs, i32 %.unpack48, i32 %.unpack50, i32 1, i32 8, ptr undef) #14
  %append.newPtr = extractvalue { ptr, i32, i32 } %append.new, 0
  call void @runtime.trackPointer(ptr %append.newPtr, ptr undef) #14
  br i1 false, label %store.throw17, label %store.next18

store.next18:                                     ; preds = %deref.next16
  %append.newLen = extractvalue { ptr, i32, i32 } %append.new, 1
  %append.newCap = extractvalue { ptr, i32, i32 } %append.new, 2
  store ptr %append.newPtr, ptr %21, align 4
  %.repack53 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 1
  store i32 %append.newLen, ptr %.repack53, align 4
  %.repack55 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 2
  store i32 %append.newCap, ptr %.repack55, align 4
  %complit = call ptr @runtime.alloc(i32 24, ptr nonnull inttoptr (i32 2637 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %complit, ptr undef) #14
  br i1 false, label %store.throw19, label %store.next20

store.next20:                                     ; preds = %store.next18
  store ptr %16, ptr %complit, align 4
  br i1 false, label %gep.throw21, label %gep.next22

gep.next22:                                       ; preds = %store.next20
  br i1 false, label %gep.throw23, label %gep.next24

gep.next24:                                       ; preds = %gep.next22
  br i1 false, label %deref.throw25, label %deref.next26

deref.next26:                                     ; preds = %gep.next24
  %23 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15
  %.unpack57 = load ptr, ptr %23, align 4
  %.elt58 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 1
  %.unpack59 = load i32, ptr %.elt58, align 4
  %.elt60 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 2
  %.unpack61 = load i32, ptr %.elt60, align 4
  call void @runtime.trackPointer(ptr %.unpack57, ptr undef) #14
  %varargs27 = call ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 133 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %varargs27, ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %complit, ptr undef) #14
  store i32 ptrtoint (ptr @"reflect/types.type:pointer:named:main.closeOnce" to i32), ptr %varargs27, align 4
  %varargs27.repack62 = getelementptr inbounds %runtime._interface, ptr %varargs27, i32 0, i32 1
  store ptr %complit, ptr %varargs27.repack62, align 4
  %append.new34 = call { ptr, i32, i32 } @runtime.sliceAppend(ptr %.unpack57, ptr nonnull %varargs27, i32 %.unpack59, i32 %.unpack61, i32 1, i32 8, ptr undef) #14
  %append.newPtr35 = extractvalue { ptr, i32, i32 } %append.new34, 0
  call void @runtime.trackPointer(ptr %append.newPtr35, ptr undef) #14
  br i1 false, label %store.throw38, label %store.next39

store.next39:                                     ; preds = %deref.next26
  %24 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15
  %append.newLen36 = extractvalue { ptr, i32, i32 } %append.new34, 1
  %append.newCap37 = extractvalue { ptr, i32, i32 } %append.new34, 2
  store ptr %append.newPtr35, ptr %24, align 4
  %.repack64 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 1
  store i32 %append.newLen36, ptr %.repack64, align 4
  %.repack66 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 2
  store i32 %append.newCap37, ptr %.repack66, align 4
  %25 = insertvalue %runtime._interface { i32 ptrtoint (ptr @"reflect/types.type:pointer:named:main.closeOnce" to i32), ptr undef }, ptr %complit, 1
  call void @runtime.trackPointer(ptr nonnull %complit, ptr undef) #14
  %26 = insertvalue { %runtime._interface, %runtime._interface } zeroinitializer, %runtime._interface %25, 0
  %27 = insertvalue { %runtime._interface, %runtime._interface } %26, %runtime._interface zeroinitializer, 1
  ret { %runtime._interface, %runtime._interface } %27

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable

gep.throw5:                                       ; preds = %if.done
  unreachable

deref.throw7:                                     ; preds = %gep.next6
  unreachable

gep.throw9:                                       ; preds = %if.done4
  unreachable

store.throw:                                      ; preds = %gep.next10
  unreachable

gep.throw11:                                      ; preds = %store.next
  unreachable

gep.throw13:                                      ; preds = %gep.next12
  unreachable

deref.throw15:                                    ; preds = %gep.next14
  unreachable

store.throw17:                                    ; preds = %deref.next16
  unreachable

store.throw19:                                    ; preds = %store.next18
  unreachable

gep.throw21:                                      ; preds = %store.next20
  unreachable

gep.throw23:                                      ; preds = %gep.next22
  unreachable

deref.throw25:                                    ; preds = %gep.next24
  unreachable

store.throw38:                                    ; preds = %deref.next26
  unreachable
}

declare void @"(*internal/task.gcData).swap"(ptr dereferenceable_or_null(4), ptr) #0

declare void @tinygo_unwind(ptr nocapture dereferenceable_or_null(8)) #8

; Function Attrs: nounwind
define linkonce_odr hidden void @"(*struct{entry uintptr; args unsafe.Pointer; internal/task.stackState; launched bool}).unwind$wrapper for func (*internal/task.stackState).unwind()"(ptr dereferenceable_or_null(20) %arg0, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %arg0, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  %1 = getelementptr inbounds { i32, ptr, %"internal/task.stackState", i1 }, ptr %arg0, i32 0, i32 2
  call void @tinygo_unwind(ptr nonnull %1) #14
  ret void

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable
}

declare void @"(*internal/task.state).initialize"(ptr dereferenceable_or_null(20), i32, ptr, i32, ptr) #0

declare void @tinygo_launch(ptr nocapture dereferenceable_or_null(20)) #9

declare void @tinygo_rewind(ptr nocapture dereferenceable_or_null(20)) #10

; Function Attrs: nounwind
define linkonce_odr hidden void @"(*internal/task.state).unwind$wrapper for func (*internal/task.stackState).unwind()"(ptr dereferenceable_or_null(20) %arg0, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %arg0, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  %1 = getelementptr inbounds %"internal/task.state", ptr %arg0, i32 0, i32 2
  call void @tinygo_unwind(ptr nonnull %1) #14
  ret void

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
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

declare i1 @"interface:{Error:func:{}{basic:string}}.$typeassert"(i32) #11

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(struct{*os.File; once sync.Once; err error}).Close$wrapper for func (*os.File).Close() (err error)"({ ptr, %sync.Once, %runtime._interface } %f, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call %runtime._interface @"(*os.File).Close"(ptr %0, ptr undef) #14
  %2 = extractvalue %runtime._interface %1, 1
  call void @runtime.trackPointer(ptr %2, ptr undef) #14
  ret %runtime._interface %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr %runtime._interface @"(struct{*os.File; once sync.Once; err error}).Close$wrapper for func (*os.File).Close() (err error)$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
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
  %ret = call %runtime._interface @"(struct{*os.File; once sync.Once; err error}).Close$wrapper for func (*os.File).Close() (err error)"({ ptr, %sync.Once, %runtime._interface } %5, ptr %1)
  ret %runtime._interface %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden i32 @"(struct{*os.File; once sync.Once; err error}).Fd$wrapper for func (*os.File).Fd() uintptr"({ ptr, %sync.Once, %runtime._interface } %f, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call i32 @"(*os.File).Fd"(ptr %0, ptr undef) #14
  ret i32 %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr i32 @"(struct{*os.File; once sync.Once; err error}).Fd$wrapper for func (*os.File).Fd() uintptr$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
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
  %ret = call i32 @"(struct{*os.File; once sync.Once; err error}).Fd$wrapper for func (*os.File).Fd() uintptr"({ ptr, %sync.Once, %runtime._interface } %5, ptr %1)
  ret i32 %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._string @"(struct{*os.File; once sync.Once; err error}).Name$wrapper for func (*os.File).Name() string"({ ptr, %sync.Once, %runtime._interface } %f, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call %runtime._string @"(*os.File).Name"(ptr %0, ptr undef) #14
  %2 = extractvalue %runtime._string %1, 0
  call void @runtime.trackPointer(ptr %2, ptr undef) #14
  ret %runtime._string %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr %runtime._string @"(struct{*os.File; once sync.Once; err error}).Name$wrapper for func (*os.File).Name() string$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
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
  %ret = call %runtime._string @"(struct{*os.File; once sync.Once; err error}).Name$wrapper for func (*os.File).Name() string"({ ptr, %sync.Once, %runtime._interface } %5, ptr %1)
  ret %runtime._string %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Read$wrapper for func (*os.File).Read(b []byte) (n int, err error)"({ ptr, %sync.Once, %runtime._interface } %f, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call { i32, %runtime._interface } @"(*os.File).Read"(ptr %0, ptr %b.data, i32 %b.len, i32 %b.cap, ptr undef) #14
  %2 = extractvalue { i32, %runtime._interface } %1, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret { i32, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Read$wrapper for func (*os.File).Read(b []byte) (n int, err error)$invoke"(ptr %0, ptr %1, i32 %2, i32 %3, ptr %4) unnamed_addr #1 {
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
  %ret = call { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Read$wrapper for func (*os.File).Read(b []byte) (n int, err error)"({ ptr, %sync.Once, %runtime._interface } %8, ptr %1, i32 %2, i32 %3, ptr %4)
  ret { i32, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).ReadAt$wrapper for func (*os.File).ReadAt(b []byte, offset int64) (n int, err error)"({ ptr, %sync.Once, %runtime._interface } %f, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %offset, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call { i32, %runtime._interface } @"(*os.File).ReadAt"(ptr %0, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %offset, ptr undef) #14
  %2 = extractvalue { i32, %runtime._interface } %1, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret { i32, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).ReadAt$wrapper for func (*os.File).ReadAt(b []byte, offset int64) (n int, err error)$invoke"(ptr %0, ptr %1, i32 %2, i32 %3, i64 %4, ptr %5) unnamed_addr #1 {
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
  %ret = call { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).ReadAt$wrapper for func (*os.File).ReadAt(b []byte, offset int64) (n int, err error)"({ ptr, %sync.Once, %runtime._interface } %9, ptr %1, i32 %2, i32 %3, i64 %4, ptr %5)
  ret { i32, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).ReadDir$wrapper for func (*os.File).ReadDir(n int) ([]io/fs.DirEntry, error)"({ ptr, %sync.Once, %runtime._interface } %f, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).ReadDir"(ptr %0, i32 %n, ptr undef) #14
  %2 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %1, 0
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %4 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %1, 1
  %5 = extractvalue %runtime._interface %4, 1
  call void @runtime.trackPointer(ptr %5, ptr undef) #14
  ret { { ptr, i32, i32 }, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { { ptr, i32, i32 }, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).ReadDir$wrapper for func (*os.File).ReadDir(n int) ([]io/fs.DirEntry, error)$invoke"(ptr %0, i32 %1, ptr %2) unnamed_addr #1 {
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
  %ret = call { { ptr, i32, i32 }, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).ReadDir$wrapper for func (*os.File).ReadDir(n int) ([]io/fs.DirEntry, error)"({ ptr, %sync.Once, %runtime._interface } %6, i32 %1, ptr %2)
  ret { { ptr, i32, i32 }, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Readdir$wrapper for func (*os.File).Readdir(n int) ([]io/fs.FileInfo, error)"({ ptr, %sync.Once, %runtime._interface } %f, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).Readdir"(ptr %0, i32 %n, ptr undef) #14
  %2 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %1, 0
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %4 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %1, 1
  %5 = extractvalue %runtime._interface %4, 1
  call void @runtime.trackPointer(ptr %5, ptr undef) #14
  ret { { ptr, i32, i32 }, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { { ptr, i32, i32 }, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Readdir$wrapper for func (*os.File).Readdir(n int) ([]io/fs.FileInfo, error)$invoke"(ptr %0, i32 %1, ptr %2) unnamed_addr #1 {
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
  %ret = call { { ptr, i32, i32 }, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Readdir$wrapper for func (*os.File).Readdir(n int) ([]io/fs.FileInfo, error)"({ ptr, %sync.Once, %runtime._interface } %6, i32 %1, ptr %2)
  ret { { ptr, i32, i32 }, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Readdirnames$wrapper for func (*os.File).Readdirnames(n int) (names []string, err error)"({ ptr, %sync.Once, %runtime._interface } %f, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).Readdirnames"(ptr %0, i32 %n, ptr undef) #14
  %2 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %1, 0
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %4 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %1, 1
  %5 = extractvalue %runtime._interface %4, 1
  call void @runtime.trackPointer(ptr %5, ptr undef) #14
  ret { { ptr, i32, i32 }, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { { ptr, i32, i32 }, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Readdirnames$wrapper for func (*os.File).Readdirnames(n int) (names []string, err error)$invoke"(ptr %0, i32 %1, ptr %2) unnamed_addr #1 {
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
  %ret = call { { ptr, i32, i32 }, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Readdirnames$wrapper for func (*os.File).Readdirnames(n int) (names []string, err error)"({ ptr, %sync.Once, %runtime._interface } %6, i32 %1, ptr %2)
  ret { { ptr, i32, i32 }, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { i64, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Seek$wrapper for func (*os.File).Seek(offset int64, whence int) (ret int64, err error)"({ ptr, %sync.Once, %runtime._interface } %f, i64 %offset, i32 %whence, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call { i64, %runtime._interface } @"(*os.File).Seek"(ptr %0, i64 %offset, i32 %whence, ptr undef) #14
  %2 = extractvalue { i64, %runtime._interface } %1, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret { i64, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { i64, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Seek$wrapper for func (*os.File).Seek(offset int64, whence int) (ret int64, err error)$invoke"(ptr %0, i64 %1, i32 %2, ptr %3) unnamed_addr #1 {
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
  %ret = call { i64, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Seek$wrapper for func (*os.File).Seek(offset int64, whence int) (ret int64, err error)"({ ptr, %sync.Once, %runtime._interface } %7, i64 %1, i32 %2, ptr %3)
  ret { i64, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { %runtime._interface, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Stat$wrapper for func (*os.File).Stat() (io/fs.FileInfo, error)"({ ptr, %sync.Once, %runtime._interface } %f, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call { %runtime._interface, %runtime._interface } @"(*os.File).Stat"(ptr %0, ptr undef) #14
  %2 = extractvalue { %runtime._interface, %runtime._interface } %1, 0
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %4 = extractvalue { %runtime._interface, %runtime._interface } %1, 1
  %5 = extractvalue %runtime._interface %4, 1
  call void @runtime.trackPointer(ptr %5, ptr undef) #14
  ret { %runtime._interface, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { %runtime._interface, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Stat$wrapper for func (*os.File).Stat() (io/fs.FileInfo, error)$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
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
  %ret = call { %runtime._interface, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Stat$wrapper for func (*os.File).Stat() (io/fs.FileInfo, error)"({ ptr, %sync.Once, %runtime._interface } %5, ptr %1)
  ret { %runtime._interface, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(struct{*os.File; once sync.Once; err error}).Sync$wrapper for func (*os.File).Sync() error"({ ptr, %sync.Once, %runtime._interface } %f, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call %runtime._interface @"(*os.File).Sync"(ptr %0, ptr undef) #14
  %2 = extractvalue %runtime._interface %1, 1
  call void @runtime.trackPointer(ptr %2, ptr undef) #14
  ret %runtime._interface %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr %runtime._interface @"(struct{*os.File; once sync.Once; err error}).Sync$wrapper for func (*os.File).Sync() error$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
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
  %ret = call %runtime._interface @"(struct{*os.File; once sync.Once; err error}).Sync$wrapper for func (*os.File).Sync() error"({ ptr, %sync.Once, %runtime._interface } %5, ptr %1)
  ret %runtime._interface %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { %runtime._interface, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).SyscallConn$wrapper for func (*os.File).SyscallConn() (syscall.RawConn, error)"({ ptr, %sync.Once, %runtime._interface } %f, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call { %runtime._interface, %runtime._interface } @"(*os.File).SyscallConn"(ptr %0, ptr undef) #14
  %2 = extractvalue { %runtime._interface, %runtime._interface } %1, 0
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %4 = extractvalue { %runtime._interface, %runtime._interface } %1, 1
  %5 = extractvalue %runtime._interface %4, 1
  call void @runtime.trackPointer(ptr %5, ptr undef) #14
  ret { %runtime._interface, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { %runtime._interface, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).SyscallConn$wrapper for func (*os.File).SyscallConn() (syscall.RawConn, error)$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
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
  %ret = call { %runtime._interface, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).SyscallConn$wrapper for func (*os.File).SyscallConn() (syscall.RawConn, error)"({ ptr, %sync.Once, %runtime._interface } %5, ptr %1)
  ret { %runtime._interface, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(struct{*os.File; once sync.Once; err error}).Truncate$wrapper for func (*os.File).Truncate(size int64) error"({ ptr, %sync.Once, %runtime._interface } %f, i64 %size, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call %runtime._interface @"(*os.File).Truncate"(ptr %0, i64 %size, ptr undef) #14
  %2 = extractvalue %runtime._interface %1, 1
  call void @runtime.trackPointer(ptr %2, ptr undef) #14
  ret %runtime._interface %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr %runtime._interface @"(struct{*os.File; once sync.Once; err error}).Truncate$wrapper for func (*os.File).Truncate(size int64) error$invoke"(ptr %0, i64 %1, ptr %2) unnamed_addr #1 {
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
  %ret = call %runtime._interface @"(struct{*os.File; once sync.Once; err error}).Truncate$wrapper for func (*os.File).Truncate(size int64) error"({ ptr, %sync.Once, %runtime._interface } %6, i64 %1, ptr %2)
  ret %runtime._interface %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Write$wrapper for func (*os.File).Write(b []byte) (n int, err error)"({ ptr, %sync.Once, %runtime._interface } %f, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call { i32, %runtime._interface } @"(*os.File).Write"(ptr %0, ptr %b.data, i32 %b.len, i32 %b.cap, ptr undef) #14
  %2 = extractvalue { i32, %runtime._interface } %1, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret { i32, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Write$wrapper for func (*os.File).Write(b []byte) (n int, err error)$invoke"(ptr %0, ptr %1, i32 %2, i32 %3, ptr %4) unnamed_addr #1 {
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
  %ret = call { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).Write$wrapper for func (*os.File).Write(b []byte) (n int, err error)"({ ptr, %sync.Once, %runtime._interface } %8, ptr %1, i32 %2, i32 %3, ptr %4)
  ret { i32, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).WriteAt$wrapper for func (*os.File).WriteAt(b []byte, off int64) (n int, err error)"({ ptr, %sync.Once, %runtime._interface } %f, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %off, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call { i32, %runtime._interface } @"(*os.File).WriteAt"(ptr %0, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %off, ptr undef) #14
  %2 = extractvalue { i32, %runtime._interface } %1, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret { i32, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).WriteAt$wrapper for func (*os.File).WriteAt(b []byte, off int64) (n int, err error)$invoke"(ptr %0, ptr %1, i32 %2, i32 %3, i64 %4, ptr %5) unnamed_addr #1 {
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
  %ret = call { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).WriteAt$wrapper for func (*os.File).WriteAt(b []byte, off int64) (n int, err error)"({ ptr, %sync.Once, %runtime._interface } %9, ptr %1, i32 %2, i32 %3, i64 %4, ptr %5)
  ret { i32, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).WriteString$wrapper for func (*os.File).WriteString(s string) (n int, err error)"({ ptr, %sync.Once, %runtime._interface } %f, ptr %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call { i32, %runtime._interface } @"(*os.File).WriteString"(ptr %0, ptr %s.data, i32 %s.len, ptr undef) #14
  %2 = extractvalue { i32, %runtime._interface } %1, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret { i32, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).WriteString$wrapper for func (*os.File).WriteString(s string) (n int, err error)$invoke"(ptr %0, ptr %1, i32 %2, ptr %3) unnamed_addr #1 {
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
  %ret = call { i32, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).WriteString$wrapper for func (*os.File).WriteString(s string) (n int, err error)"({ ptr, %sync.Once, %runtime._interface } %7, ptr %1, i32 %2, ptr %3)
  ret { i32, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).readdir$wrapper for func (*os.File).readdir(n int, mode os.readdirMode) (names []string, dirents []io/fs.DirEntry, infos []io/fs.FileInfo, err error)"({ ptr, %sync.Once, %runtime._interface } %f, i32 %n, i32 %mode, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca { ptr, %sync.Once, %runtime._interface }, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds { ptr, %sync.Once, %runtime._interface }, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(*os.File).readdir"(ptr %0, i32 %n, i32 %mode, ptr undef) #14
  %2 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %1, 0
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %4 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %1, 1
  %5 = extractvalue { ptr, i32, i32 } %4, 0
  call void @runtime.trackPointer(ptr %5, ptr undef) #14
  %6 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %1, 2
  %7 = extractvalue { ptr, i32, i32 } %6, 0
  call void @runtime.trackPointer(ptr %7, ptr undef) #14
  %8 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %1, 3
  %9 = extractvalue %runtime._interface %8, 1
  call void @runtime.trackPointer(ptr %9, ptr undef) #14
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
define linkonce_odr { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).readdir$wrapper for func (*os.File).readdir(n int, mode os.readdirMode) (names []string, dirents []io/fs.DirEntry, infos []io/fs.FileInfo, err error)$invoke"(ptr %0, i32 %1, i32 %2, ptr %3) unnamed_addr #1 {
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
  %ret = call { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(struct{*os.File; once sync.Once; err error}).readdir$wrapper for func (*os.File).readdir(n int, mode os.readdirMode) (names []string, dirents []io/fs.DirEntry, infos []io/fs.FileInfo, err error)"({ ptr, %sync.Once, %runtime._interface } %7, i32 %1, i32 %2, ptr %3)
  ret { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(*struct{*os.File; once sync.Once; err error}).Close$wrapper for func (*os.File).Close() (err error)"(ptr dereferenceable_or_null(24) %f, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call %runtime._interface @"(*os.File).Close"(ptr %1, ptr undef) #14
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret %runtime._interface %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden i32 @"(*struct{*os.File; once sync.Once; err error}).Fd$wrapper for func (*os.File).Fd() uintptr"(ptr dereferenceable_or_null(24) %f, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call i32 @"(*os.File).Fd"(ptr %1, ptr undef) #14
  ret i32 %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._string @"(*struct{*os.File; once sync.Once; err error}).Name$wrapper for func (*os.File).Name() string"(ptr dereferenceable_or_null(24) %f, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call %runtime._string @"(*os.File).Name"(ptr %1, ptr undef) #14
  %3 = extractvalue %runtime._string %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret %runtime._string %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(*struct{*os.File; once sync.Once; err error}).Read$wrapper for func (*os.File).Read(b []byte) (n int, err error)"(ptr dereferenceable_or_null(24) %f, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call { i32, %runtime._interface } @"(*os.File).Read"(ptr %1, ptr %b.data, i32 %b.len, i32 %b.cap, ptr undef) #14
  %3 = extractvalue { i32, %runtime._interface } %2, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  ret { i32, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(*struct{*os.File; once sync.Once; err error}).ReadAt$wrapper for func (*os.File).ReadAt(b []byte, offset int64) (n int, err error)"(ptr dereferenceable_or_null(24) %f, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %offset, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call { i32, %runtime._interface } @"(*os.File).ReadAt"(ptr %1, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %offset, ptr undef) #14
  %3 = extractvalue { i32, %runtime._interface } %2, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  ret { i32, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, %runtime._interface } @"(*struct{*os.File; once sync.Once; err error}).ReadDir$wrapper for func (*os.File).ReadDir(n int) ([]io/fs.DirEntry, error)"(ptr dereferenceable_or_null(24) %f, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).ReadDir"(ptr %1, i32 %n, ptr undef) #14
  %3 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %2, 0
  %4 = extractvalue { ptr, i32, i32 } %3, 0
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  %5 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %2, 1
  %6 = extractvalue %runtime._interface %5, 1
  call void @runtime.trackPointer(ptr %6, ptr undef) #14
  ret { { ptr, i32, i32 }, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, %runtime._interface } @"(*struct{*os.File; once sync.Once; err error}).Readdir$wrapper for func (*os.File).Readdir(n int) ([]io/fs.FileInfo, error)"(ptr dereferenceable_or_null(24) %f, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).Readdir"(ptr %1, i32 %n, ptr undef) #14
  %3 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %2, 0
  %4 = extractvalue { ptr, i32, i32 } %3, 0
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  %5 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %2, 1
  %6 = extractvalue %runtime._interface %5, 1
  call void @runtime.trackPointer(ptr %6, ptr undef) #14
  ret { { ptr, i32, i32 }, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, %runtime._interface } @"(*struct{*os.File; once sync.Once; err error}).Readdirnames$wrapper for func (*os.File).Readdirnames(n int) (names []string, err error)"(ptr dereferenceable_or_null(24) %f, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).Readdirnames"(ptr %1, i32 %n, ptr undef) #14
  %3 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %2, 0
  %4 = extractvalue { ptr, i32, i32 } %3, 0
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  %5 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %2, 1
  %6 = extractvalue %runtime._interface %5, 1
  call void @runtime.trackPointer(ptr %6, ptr undef) #14
  ret { { ptr, i32, i32 }, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { i64, %runtime._interface } @"(*struct{*os.File; once sync.Once; err error}).Seek$wrapper for func (*os.File).Seek(offset int64, whence int) (ret int64, err error)"(ptr dereferenceable_or_null(24) %f, i64 %offset, i32 %whence, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call { i64, %runtime._interface } @"(*os.File).Seek"(ptr %1, i64 %offset, i32 %whence, ptr undef) #14
  %3 = extractvalue { i64, %runtime._interface } %2, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  ret { i64, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { %runtime._interface, %runtime._interface } @"(*struct{*os.File; once sync.Once; err error}).Stat$wrapper for func (*os.File).Stat() (io/fs.FileInfo, error)"(ptr dereferenceable_or_null(24) %f, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call { %runtime._interface, %runtime._interface } @"(*os.File).Stat"(ptr %1, ptr undef) #14
  %3 = extractvalue { %runtime._interface, %runtime._interface } %2, 0
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  %5 = extractvalue { %runtime._interface, %runtime._interface } %2, 1
  %6 = extractvalue %runtime._interface %5, 1
  call void @runtime.trackPointer(ptr %6, ptr undef) #14
  ret { %runtime._interface, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(*struct{*os.File; once sync.Once; err error}).Sync$wrapper for func (*os.File).Sync() error"(ptr dereferenceable_or_null(24) %f, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call %runtime._interface @"(*os.File).Sync"(ptr %1, ptr undef) #14
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret %runtime._interface %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { %runtime._interface, %runtime._interface } @"(*struct{*os.File; once sync.Once; err error}).SyscallConn$wrapper for func (*os.File).SyscallConn() (syscall.RawConn, error)"(ptr dereferenceable_or_null(24) %f, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call { %runtime._interface, %runtime._interface } @"(*os.File).SyscallConn"(ptr %1, ptr undef) #14
  %3 = extractvalue { %runtime._interface, %runtime._interface } %2, 0
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  %5 = extractvalue { %runtime._interface, %runtime._interface } %2, 1
  %6 = extractvalue %runtime._interface %5, 1
  call void @runtime.trackPointer(ptr %6, ptr undef) #14
  ret { %runtime._interface, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(*struct{*os.File; once sync.Once; err error}).Truncate$wrapper for func (*os.File).Truncate(size int64) error"(ptr dereferenceable_or_null(24) %f, i64 %size, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call %runtime._interface @"(*os.File).Truncate"(ptr %1, i64 %size, ptr undef) #14
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret %runtime._interface %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(*struct{*os.File; once sync.Once; err error}).Write$wrapper for func (*os.File).Write(b []byte) (n int, err error)"(ptr dereferenceable_or_null(24) %f, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call { i32, %runtime._interface } @"(*os.File).Write"(ptr %1, ptr %b.data, i32 %b.len, i32 %b.cap, ptr undef) #14
  %3 = extractvalue { i32, %runtime._interface } %2, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  ret { i32, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(*struct{*os.File; once sync.Once; err error}).WriteAt$wrapper for func (*os.File).WriteAt(b []byte, off int64) (n int, err error)"(ptr dereferenceable_or_null(24) %f, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %off, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call { i32, %runtime._interface } @"(*os.File).WriteAt"(ptr %1, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %off, ptr undef) #14
  %3 = extractvalue { i32, %runtime._interface } %2, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  ret { i32, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(*struct{*os.File; once sync.Once; err error}).WriteString$wrapper for func (*os.File).WriteString(s string) (n int, err error)"(ptr dereferenceable_or_null(24) %f, ptr %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call { i32, %runtime._interface } @"(*os.File).WriteString"(ptr %1, ptr %s.data, i32 %s.len, ptr undef) #14
  %3 = extractvalue { i32, %runtime._interface } %2, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  ret { i32, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(*struct{*os.File; once sync.Once; err error}).readdir$wrapper for func (*os.File).readdir(n int, mode os.readdirMode) (names []string, dirents []io/fs.DirEntry, infos []io/fs.FileInfo, err error)"(ptr dereferenceable_or_null(24) %f, i32 %n, i32 %mode, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(*os.File).readdir"(ptr %1, i32 %n, i32 %mode, ptr undef) #14
  %3 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %2, 0
  %4 = extractvalue { ptr, i32, i32 } %3, 0
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  %5 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %2, 1
  %6 = extractvalue { ptr, i32, i32 } %5, 0
  call void @runtime.trackPointer(ptr %6, ptr undef) #14
  %7 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %2, 2
  %8 = extractvalue { ptr, i32, i32 } %7, 0
  call void @runtime.trackPointer(ptr %8, ptr undef) #14
  %9 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %2, 3
  %10 = extractvalue %runtime._interface %9, 1
  call void @runtime.trackPointer(ptr %10, ptr undef) #14
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
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden i32 @"(main.closeOnce).Fd$wrapper for func (*os.File).Fd() uintptr"(%main.closeOnce %f, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call i32 @"(*os.File).Fd"(ptr %0, ptr undef) #14
  ret i32 %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr i32 @"(main.closeOnce).Fd$wrapper for func (*os.File).Fd() uintptr$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
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
  %ret = call i32 @"(main.closeOnce).Fd$wrapper for func (*os.File).Fd() uintptr"(%main.closeOnce %5, ptr %1)
  ret i32 %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._string @"(main.closeOnce).Name$wrapper for func (*os.File).Name() string"(%main.closeOnce %f, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call %runtime._string @"(*os.File).Name"(ptr %0, ptr undef) #14
  %2 = extractvalue %runtime._string %1, 0
  call void @runtime.trackPointer(ptr %2, ptr undef) #14
  ret %runtime._string %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr %runtime._string @"(main.closeOnce).Name$wrapper for func (*os.File).Name() string$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
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
  %ret = call %runtime._string @"(main.closeOnce).Name$wrapper for func (*os.File).Name() string"(%main.closeOnce %5, ptr %1)
  ret %runtime._string %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(main.closeOnce).Read$wrapper for func (*os.File).Read(b []byte) (n int, err error)"(%main.closeOnce %f, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call { i32, %runtime._interface } @"(*os.File).Read"(ptr %0, ptr %b.data, i32 %b.len, i32 %b.cap, ptr undef) #14
  %2 = extractvalue { i32, %runtime._interface } %1, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret { i32, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { i32, %runtime._interface } @"(main.closeOnce).Read$wrapper for func (*os.File).Read(b []byte) (n int, err error)$invoke"(ptr %0, ptr %1, i32 %2, i32 %3, ptr %4) unnamed_addr #1 {
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
  %ret = call { i32, %runtime._interface } @"(main.closeOnce).Read$wrapper for func (*os.File).Read(b []byte) (n int, err error)"(%main.closeOnce %8, ptr %1, i32 %2, i32 %3, ptr %4)
  ret { i32, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(main.closeOnce).ReadAt$wrapper for func (*os.File).ReadAt(b []byte, offset int64) (n int, err error)"(%main.closeOnce %f, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %offset, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call { i32, %runtime._interface } @"(*os.File).ReadAt"(ptr %0, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %offset, ptr undef) #14
  %2 = extractvalue { i32, %runtime._interface } %1, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret { i32, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { i32, %runtime._interface } @"(main.closeOnce).ReadAt$wrapper for func (*os.File).ReadAt(b []byte, offset int64) (n int, err error)$invoke"(ptr %0, ptr %1, i32 %2, i32 %3, i64 %4, ptr %5) unnamed_addr #1 {
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
  %ret = call { i32, %runtime._interface } @"(main.closeOnce).ReadAt$wrapper for func (*os.File).ReadAt(b []byte, offset int64) (n int, err error)"(%main.closeOnce %9, ptr %1, i32 %2, i32 %3, i64 %4, ptr %5)
  ret { i32, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, %runtime._interface } @"(main.closeOnce).ReadDir$wrapper for func (*os.File).ReadDir(n int) ([]io/fs.DirEntry, error)"(%main.closeOnce %f, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).ReadDir"(ptr %0, i32 %n, ptr undef) #14
  %2 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %1, 0
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %4 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %1, 1
  %5 = extractvalue %runtime._interface %4, 1
  call void @runtime.trackPointer(ptr %5, ptr undef) #14
  ret { { ptr, i32, i32 }, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { { ptr, i32, i32 }, %runtime._interface } @"(main.closeOnce).ReadDir$wrapper for func (*os.File).ReadDir(n int) ([]io/fs.DirEntry, error)$invoke"(ptr %0, i32 %1, ptr %2) unnamed_addr #1 {
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
  %ret = call { { ptr, i32, i32 }, %runtime._interface } @"(main.closeOnce).ReadDir$wrapper for func (*os.File).ReadDir(n int) ([]io/fs.DirEntry, error)"(%main.closeOnce %6, i32 %1, ptr %2)
  ret { { ptr, i32, i32 }, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, %runtime._interface } @"(main.closeOnce).Readdir$wrapper for func (*os.File).Readdir(n int) ([]io/fs.FileInfo, error)"(%main.closeOnce %f, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).Readdir"(ptr %0, i32 %n, ptr undef) #14
  %2 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %1, 0
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %4 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %1, 1
  %5 = extractvalue %runtime._interface %4, 1
  call void @runtime.trackPointer(ptr %5, ptr undef) #14
  ret { { ptr, i32, i32 }, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { { ptr, i32, i32 }, %runtime._interface } @"(main.closeOnce).Readdir$wrapper for func (*os.File).Readdir(n int) ([]io/fs.FileInfo, error)$invoke"(ptr %0, i32 %1, ptr %2) unnamed_addr #1 {
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
  %ret = call { { ptr, i32, i32 }, %runtime._interface } @"(main.closeOnce).Readdir$wrapper for func (*os.File).Readdir(n int) ([]io/fs.FileInfo, error)"(%main.closeOnce %6, i32 %1, ptr %2)
  ret { { ptr, i32, i32 }, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, %runtime._interface } @"(main.closeOnce).Readdirnames$wrapper for func (*os.File).Readdirnames(n int) (names []string, err error)"(%main.closeOnce %f, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).Readdirnames"(ptr %0, i32 %n, ptr undef) #14
  %2 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %1, 0
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %4 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %1, 1
  %5 = extractvalue %runtime._interface %4, 1
  call void @runtime.trackPointer(ptr %5, ptr undef) #14
  ret { { ptr, i32, i32 }, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { { ptr, i32, i32 }, %runtime._interface } @"(main.closeOnce).Readdirnames$wrapper for func (*os.File).Readdirnames(n int) (names []string, err error)$invoke"(ptr %0, i32 %1, ptr %2) unnamed_addr #1 {
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
  %ret = call { { ptr, i32, i32 }, %runtime._interface } @"(main.closeOnce).Readdirnames$wrapper for func (*os.File).Readdirnames(n int) (names []string, err error)"(%main.closeOnce %6, i32 %1, ptr %2)
  ret { { ptr, i32, i32 }, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { i64, %runtime._interface } @"(main.closeOnce).Seek$wrapper for func (*os.File).Seek(offset int64, whence int) (ret int64, err error)"(%main.closeOnce %f, i64 %offset, i32 %whence, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call { i64, %runtime._interface } @"(*os.File).Seek"(ptr %0, i64 %offset, i32 %whence, ptr undef) #14
  %2 = extractvalue { i64, %runtime._interface } %1, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret { i64, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { i64, %runtime._interface } @"(main.closeOnce).Seek$wrapper for func (*os.File).Seek(offset int64, whence int) (ret int64, err error)$invoke"(ptr %0, i64 %1, i32 %2, ptr %3) unnamed_addr #1 {
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
  %ret = call { i64, %runtime._interface } @"(main.closeOnce).Seek$wrapper for func (*os.File).Seek(offset int64, whence int) (ret int64, err error)"(%main.closeOnce %7, i64 %1, i32 %2, ptr %3)
  ret { i64, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { %runtime._interface, %runtime._interface } @"(main.closeOnce).Stat$wrapper for func (*os.File).Stat() (io/fs.FileInfo, error)"(%main.closeOnce %f, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call { %runtime._interface, %runtime._interface } @"(*os.File).Stat"(ptr %0, ptr undef) #14
  %2 = extractvalue { %runtime._interface, %runtime._interface } %1, 0
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %4 = extractvalue { %runtime._interface, %runtime._interface } %1, 1
  %5 = extractvalue %runtime._interface %4, 1
  call void @runtime.trackPointer(ptr %5, ptr undef) #14
  ret { %runtime._interface, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { %runtime._interface, %runtime._interface } @"(main.closeOnce).Stat$wrapper for func (*os.File).Stat() (io/fs.FileInfo, error)$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
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
  %ret = call { %runtime._interface, %runtime._interface } @"(main.closeOnce).Stat$wrapper for func (*os.File).Stat() (io/fs.FileInfo, error)"(%main.closeOnce %5, ptr %1)
  ret { %runtime._interface, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(main.closeOnce).Sync$wrapper for func (*os.File).Sync() error"(%main.closeOnce %f, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call %runtime._interface @"(*os.File).Sync"(ptr %0, ptr undef) #14
  %2 = extractvalue %runtime._interface %1, 1
  call void @runtime.trackPointer(ptr %2, ptr undef) #14
  ret %runtime._interface %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr %runtime._interface @"(main.closeOnce).Sync$wrapper for func (*os.File).Sync() error$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
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
  %ret = call %runtime._interface @"(main.closeOnce).Sync$wrapper for func (*os.File).Sync() error"(%main.closeOnce %5, ptr %1)
  ret %runtime._interface %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { %runtime._interface, %runtime._interface } @"(main.closeOnce).SyscallConn$wrapper for func (*os.File).SyscallConn() (syscall.RawConn, error)"(%main.closeOnce %f, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call { %runtime._interface, %runtime._interface } @"(*os.File).SyscallConn"(ptr %0, ptr undef) #14
  %2 = extractvalue { %runtime._interface, %runtime._interface } %1, 0
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %4 = extractvalue { %runtime._interface, %runtime._interface } %1, 1
  %5 = extractvalue %runtime._interface %4, 1
  call void @runtime.trackPointer(ptr %5, ptr undef) #14
  ret { %runtime._interface, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { %runtime._interface, %runtime._interface } @"(main.closeOnce).SyscallConn$wrapper for func (*os.File).SyscallConn() (syscall.RawConn, error)$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
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
  %ret = call { %runtime._interface, %runtime._interface } @"(main.closeOnce).SyscallConn$wrapper for func (*os.File).SyscallConn() (syscall.RawConn, error)"(%main.closeOnce %5, ptr %1)
  ret { %runtime._interface, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(main.closeOnce).Truncate$wrapper for func (*os.File).Truncate(size int64) error"(%main.closeOnce %f, i64 %size, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call %runtime._interface @"(*os.File).Truncate"(ptr %0, i64 %size, ptr undef) #14
  %2 = extractvalue %runtime._interface %1, 1
  call void @runtime.trackPointer(ptr %2, ptr undef) #14
  ret %runtime._interface %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr %runtime._interface @"(main.closeOnce).Truncate$wrapper for func (*os.File).Truncate(size int64) error$invoke"(ptr %0, i64 %1, ptr %2) unnamed_addr #1 {
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
  %ret = call %runtime._interface @"(main.closeOnce).Truncate$wrapper for func (*os.File).Truncate(size int64) error"(%main.closeOnce %6, i64 %1, ptr %2)
  ret %runtime._interface %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(main.closeOnce).Write$wrapper for func (*os.File).Write(b []byte) (n int, err error)"(%main.closeOnce %f, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call { i32, %runtime._interface } @"(*os.File).Write"(ptr %0, ptr %b.data, i32 %b.len, i32 %b.cap, ptr undef) #14
  %2 = extractvalue { i32, %runtime._interface } %1, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret { i32, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { i32, %runtime._interface } @"(main.closeOnce).Write$wrapper for func (*os.File).Write(b []byte) (n int, err error)$invoke"(ptr %0, ptr %1, i32 %2, i32 %3, ptr %4) unnamed_addr #1 {
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
  %ret = call { i32, %runtime._interface } @"(main.closeOnce).Write$wrapper for func (*os.File).Write(b []byte) (n int, err error)"(%main.closeOnce %8, ptr %1, i32 %2, i32 %3, ptr %4)
  ret { i32, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(main.closeOnce).WriteAt$wrapper for func (*os.File).WriteAt(b []byte, off int64) (n int, err error)"(%main.closeOnce %f, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %off, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call { i32, %runtime._interface } @"(*os.File).WriteAt"(ptr %0, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %off, ptr undef) #14
  %2 = extractvalue { i32, %runtime._interface } %1, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret { i32, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { i32, %runtime._interface } @"(main.closeOnce).WriteAt$wrapper for func (*os.File).WriteAt(b []byte, off int64) (n int, err error)$invoke"(ptr %0, ptr %1, i32 %2, i32 %3, i64 %4, ptr %5) unnamed_addr #1 {
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
  %ret = call { i32, %runtime._interface } @"(main.closeOnce).WriteAt$wrapper for func (*os.File).WriteAt(b []byte, off int64) (n int, err error)"(%main.closeOnce %9, ptr %1, i32 %2, i32 %3, i64 %4, ptr %5)
  ret { i32, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(main.closeOnce).WriteString$wrapper for func (*os.File).WriteString(s string) (n int, err error)"(%main.closeOnce %f, ptr %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call { i32, %runtime._interface } @"(*os.File).WriteString"(ptr %0, ptr %s.data, i32 %s.len, ptr undef) #14
  %2 = extractvalue { i32, %runtime._interface } %1, 1
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret { i32, %runtime._interface } %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr { i32, %runtime._interface } @"(main.closeOnce).WriteString$wrapper for func (*os.File).WriteString(s string) (n int, err error)$invoke"(ptr %0, ptr %1, i32 %2, ptr %3) unnamed_addr #1 {
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
  %ret = call { i32, %runtime._interface } @"(main.closeOnce).WriteString$wrapper for func (*os.File).WriteString(s string) (n int, err error)"(%main.closeOnce %7, ptr %1, i32 %2, ptr %3)
  ret { i32, %runtime._interface } %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(main.closeOnce).readdir$wrapper for func (*os.File).readdir(n int, mode os.readdirMode) (names []string, dirents []io/fs.DirEntry, infos []io/fs.FileInfo, err error)"(%main.closeOnce %f, i32 %n, i32 %mode, ptr %context) unnamed_addr #1 {
entry:
  %f1 = alloca %main.closeOnce, align 8
  store ptr null, ptr %f1, align 8
  %f1.repack2 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 1
  store %sync.Once zeroinitializer, ptr %f1.repack2, align 4
  %f1.repack3 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2
  store i32 0, ptr %f1.repack3, align 8
  %f1.repack3.repack4 = getelementptr inbounds %main.closeOnce, ptr %f1, i32 0, i32 2, i32 1
  store ptr null, ptr %f1.repack3.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %f1, ptr undef) #14
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
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(*os.File).readdir"(ptr %0, i32 %n, i32 %mode, ptr undef) #14
  %2 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %1, 0
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %4 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %1, 1
  %5 = extractvalue { ptr, i32, i32 } %4, 0
  call void @runtime.trackPointer(ptr %5, ptr undef) #14
  %6 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %1, 2
  %7 = extractvalue { ptr, i32, i32 } %6, 0
  call void @runtime.trackPointer(ptr %7, ptr undef) #14
  %8 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %1, 3
  %9 = extractvalue %runtime._interface %8, 1
  call void @runtime.trackPointer(ptr %9, ptr undef) #14
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
define linkonce_odr { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(main.closeOnce).readdir$wrapper for func (*os.File).readdir(n int, mode os.readdirMode) (names []string, dirents []io/fs.DirEntry, infos []io/fs.FileInfo, err error)$invoke"(ptr %0, i32 %1, i32 %2, ptr %3) unnamed_addr #1 {
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
  %ret = call { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(main.closeOnce).readdir$wrapper for func (*os.File).readdir(n int, mode os.readdirMode) (names []string, dirents []io/fs.DirEntry, infos []io/fs.FileInfo, err error)"(%main.closeOnce %7, i32 %1, i32 %2, ptr %3)
  ret { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %ret
}

; Function Attrs: nounwind
define hidden %runtime._interface @"(*main.closeOnce).Close"(ptr dereferenceable_or_null(24) %c, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %c, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  %1 = getelementptr inbounds %main.closeOnce, ptr %c, i32 0, i32 1
  call void @runtime.trackPointer(ptr nonnull %c, ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull @"(*main.closeOnce).close$bound$bound method wrapper for func (*main.closeOnce).close()", ptr undef) #14
  call void @"(*sync.Once).Do"(ptr nonnull %1, ptr nonnull %c, ptr nonnull @"(*main.closeOnce).close$bound$bound method wrapper for func (*main.closeOnce).close()", ptr undef) #14
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
  call void @runtime.trackPointer(ptr %.unpack4, ptr undef) #14
  ret %runtime._interface %4

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

gep.throw1:                                       ; preds = %gep.next
  unreachable

deref.throw:                                      ; preds = %gep.next2
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden i32 @"(*main.closeOnce).Fd$wrapper for func (*os.File).Fd() uintptr"(ptr dereferenceable_or_null(24) %f, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call i32 @"(*os.File).Fd"(ptr %1, ptr undef) #14
  ret i32 %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._string @"(*main.closeOnce).Name$wrapper for func (*os.File).Name() string"(ptr dereferenceable_or_null(24) %f, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call %runtime._string @"(*os.File).Name"(ptr %1, ptr undef) #14
  %3 = extractvalue %runtime._string %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret %runtime._string %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(*main.closeOnce).Read$wrapper for func (*os.File).Read(b []byte) (n int, err error)"(ptr dereferenceable_or_null(24) %f, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call { i32, %runtime._interface } @"(*os.File).Read"(ptr %1, ptr %b.data, i32 %b.len, i32 %b.cap, ptr undef) #14
  %3 = extractvalue { i32, %runtime._interface } %2, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  ret { i32, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(*main.closeOnce).ReadAt$wrapper for func (*os.File).ReadAt(b []byte, offset int64) (n int, err error)"(ptr dereferenceable_or_null(24) %f, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %offset, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call { i32, %runtime._interface } @"(*os.File).ReadAt"(ptr %1, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %offset, ptr undef) #14
  %3 = extractvalue { i32, %runtime._interface } %2, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  ret { i32, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, %runtime._interface } @"(*main.closeOnce).ReadDir$wrapper for func (*os.File).ReadDir(n int) ([]io/fs.DirEntry, error)"(ptr dereferenceable_or_null(24) %f, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).ReadDir"(ptr %1, i32 %n, ptr undef) #14
  %3 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %2, 0
  %4 = extractvalue { ptr, i32, i32 } %3, 0
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  %5 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %2, 1
  %6 = extractvalue %runtime._interface %5, 1
  call void @runtime.trackPointer(ptr %6, ptr undef) #14
  ret { { ptr, i32, i32 }, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, %runtime._interface } @"(*main.closeOnce).Readdir$wrapper for func (*os.File).Readdir(n int) ([]io/fs.FileInfo, error)"(ptr dereferenceable_or_null(24) %f, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).Readdir"(ptr %1, i32 %n, ptr undef) #14
  %3 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %2, 0
  %4 = extractvalue { ptr, i32, i32 } %3, 0
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  %5 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %2, 1
  %6 = extractvalue %runtime._interface %5, 1
  call void @runtime.trackPointer(ptr %6, ptr undef) #14
  ret { { ptr, i32, i32 }, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, %runtime._interface } @"(*main.closeOnce).Readdirnames$wrapper for func (*os.File).Readdirnames(n int) (names []string, err error)"(ptr dereferenceable_or_null(24) %f, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call { { ptr, i32, i32 }, %runtime._interface } @"(*os.File).Readdirnames"(ptr %1, i32 %n, ptr undef) #14
  %3 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %2, 0
  %4 = extractvalue { ptr, i32, i32 } %3, 0
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  %5 = extractvalue { { ptr, i32, i32 }, %runtime._interface } %2, 1
  %6 = extractvalue %runtime._interface %5, 1
  call void @runtime.trackPointer(ptr %6, ptr undef) #14
  ret { { ptr, i32, i32 }, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { i64, %runtime._interface } @"(*main.closeOnce).Seek$wrapper for func (*os.File).Seek(offset int64, whence int) (ret int64, err error)"(ptr dereferenceable_or_null(24) %f, i64 %offset, i32 %whence, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call { i64, %runtime._interface } @"(*os.File).Seek"(ptr %1, i64 %offset, i32 %whence, ptr undef) #14
  %3 = extractvalue { i64, %runtime._interface } %2, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  ret { i64, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { %runtime._interface, %runtime._interface } @"(*main.closeOnce).Stat$wrapper for func (*os.File).Stat() (io/fs.FileInfo, error)"(ptr dereferenceable_or_null(24) %f, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call { %runtime._interface, %runtime._interface } @"(*os.File).Stat"(ptr %1, ptr undef) #14
  %3 = extractvalue { %runtime._interface, %runtime._interface } %2, 0
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  %5 = extractvalue { %runtime._interface, %runtime._interface } %2, 1
  %6 = extractvalue %runtime._interface %5, 1
  call void @runtime.trackPointer(ptr %6, ptr undef) #14
  ret { %runtime._interface, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(*main.closeOnce).Sync$wrapper for func (*os.File).Sync() error"(ptr dereferenceable_or_null(24) %f, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call %runtime._interface @"(*os.File).Sync"(ptr %1, ptr undef) #14
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret %runtime._interface %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { %runtime._interface, %runtime._interface } @"(*main.closeOnce).SyscallConn$wrapper for func (*os.File).SyscallConn() (syscall.RawConn, error)"(ptr dereferenceable_or_null(24) %f, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call { %runtime._interface, %runtime._interface } @"(*os.File).SyscallConn"(ptr %1, ptr undef) #14
  %3 = extractvalue { %runtime._interface, %runtime._interface } %2, 0
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  %5 = extractvalue { %runtime._interface, %runtime._interface } %2, 1
  %6 = extractvalue %runtime._interface %5, 1
  call void @runtime.trackPointer(ptr %6, ptr undef) #14
  ret { %runtime._interface, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(*main.closeOnce).Truncate$wrapper for func (*os.File).Truncate(size int64) error"(ptr dereferenceable_or_null(24) %f, i64 %size, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call %runtime._interface @"(*os.File).Truncate"(ptr %1, i64 %size, ptr undef) #14
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret %runtime._interface %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(*main.closeOnce).Write$wrapper for func (*os.File).Write(b []byte) (n int, err error)"(ptr dereferenceable_or_null(24) %f, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call { i32, %runtime._interface } @"(*os.File).Write"(ptr %1, ptr %b.data, i32 %b.len, i32 %b.cap, ptr undef) #14
  %3 = extractvalue { i32, %runtime._interface } %2, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  ret { i32, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(*main.closeOnce).WriteAt$wrapper for func (*os.File).WriteAt(b []byte, off int64) (n int, err error)"(ptr dereferenceable_or_null(24) %f, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %off, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call { i32, %runtime._interface } @"(*os.File).WriteAt"(ptr %1, ptr %b.data, i32 %b.len, i32 %b.cap, i64 %off, ptr undef) #14
  %3 = extractvalue { i32, %runtime._interface } %2, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  ret { i32, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { i32, %runtime._interface } @"(*main.closeOnce).WriteString$wrapper for func (*os.File).WriteString(s string) (n int, err error)"(ptr dereferenceable_or_null(24) %f, ptr %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call { i32, %runtime._interface } @"(*os.File).WriteString"(ptr %1, ptr %s.data, i32 %s.len, ptr undef) #14
  %3 = extractvalue { i32, %runtime._interface } %2, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  ret { i32, %runtime._interface } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
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
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call %runtime._interface @"(*os.File).Close"(ptr %1, ptr undef) #14
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
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
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

gep.throw1:                                       ; preds = %gep.next
  unreachable

deref.throw:                                      ; preds = %gep.next2
  unreachable

store.throw:                                      ; preds = %deref.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(*main.closeOnce).readdir$wrapper for func (*os.File).readdir(n int, mode os.readdirMode) (names []string, dirents []io/fs.DirEntry, infos []io/fs.FileInfo, err error)"(ptr dereferenceable_or_null(24) %f, i32 %n, i32 %mode, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } @"(*os.File).readdir"(ptr %1, i32 %n, i32 %mode, ptr undef) #14
  %3 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %2, 0
  %4 = extractvalue { ptr, i32, i32 } %3, 0
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  %5 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %2, 1
  %6 = extractvalue { ptr, i32, i32 } %5, 0
  call void @runtime.trackPointer(ptr %6, ptr undef) #14
  %7 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %2, 2
  %8 = extractvalue { ptr, i32, i32 } %7, 0
  call void @runtime.trackPointer(ptr %8, ptr undef) #14
  %9 = extractvalue { { ptr, i32, i32 }, { ptr, i32, i32 }, { ptr, i32, i32 }, %runtime._interface } %2, 3
  %10 = extractvalue %runtime._interface %9, 1
  call void @runtime.trackPointer(ptr %10, ptr undef) #14
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
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define hidden { %runtime._interface, %runtime._interface } @"(*main.Cmd).StdoutPipe"(ptr dereferenceable_or_null(168) %c, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %c, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 5
  %.unpack = load i32, ptr %1, align 4
  %.elt38 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 5, i32 1
  %.unpack39 = load ptr, ptr %.elt38, align 4
  call void @runtime.trackPointer(ptr %.unpack39, ptr undef) #14
  %.not = icmp eq i32 %.unpack, 0
  br i1 %.not, label %if.done, label %if.then

if.then:                                          ; preds = %deref.next
  %2 = call %runtime._interface @errors.New(ptr nonnull @"main$string.55", i32 24, ptr undef) #14
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %4 = insertvalue { %runtime._interface, %runtime._interface } zeroinitializer, %runtime._interface %2, 1
  ret { %runtime._interface, %runtime._interface } %4

if.done:                                          ; preds = %deref.next
  br i1 false, label %gep.throw5, label %gep.next6

gep.next6:                                        ; preds = %if.done
  br i1 false, label %deref.throw7, label %deref.next8

deref.next8:                                      ; preds = %gep.next6
  %5 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 9
  %6 = load ptr, ptr %5, align 4
  call void @runtime.trackPointer(ptr %6, ptr undef) #14
  %.not40 = icmp eq ptr %6, null
  br i1 %.not40, label %if.done2, label %if.then1

if.then1:                                         ; preds = %deref.next8
  %7 = call %runtime._interface @errors.New(ptr nonnull @"main$string.56", i32 38, ptr undef) #14
  %8 = extractvalue %runtime._interface %7, 1
  call void @runtime.trackPointer(ptr %8, ptr undef) #14
  %9 = insertvalue { %runtime._interface, %runtime._interface } zeroinitializer, %runtime._interface %7, 1
  ret { %runtime._interface, %runtime._interface } %9

if.done2:                                         ; preds = %deref.next8
  %10 = call { ptr, ptr, %runtime._interface } @os.Pipe(ptr undef) #14
  %11 = extractvalue { ptr, ptr, %runtime._interface } %10, 0
  call void @runtime.trackPointer(ptr %11, ptr undef) #14
  %12 = extractvalue { ptr, ptr, %runtime._interface } %10, 1
  call void @runtime.trackPointer(ptr %12, ptr undef) #14
  %13 = extractvalue { ptr, ptr, %runtime._interface } %10, 2
  %14 = extractvalue %runtime._interface %13, 1
  call void @runtime.trackPointer(ptr %14, ptr undef) #14
  %15 = extractvalue { ptr, ptr, %runtime._interface } %10, 0
  %16 = extractvalue { ptr, ptr, %runtime._interface } %10, 1
  %17 = extractvalue { ptr, ptr, %runtime._interface } %10, 2
  %18 = extractvalue %runtime._interface %17, 0
  %.not41 = icmp eq i32 %18, 0
  br i1 %.not41, label %if.done4, label %if.then3

if.then3:                                         ; preds = %if.done2
  %19 = insertvalue { %runtime._interface, %runtime._interface } zeroinitializer, %runtime._interface %17, 1
  ret { %runtime._interface, %runtime._interface } %19

if.done4:                                         ; preds = %if.done2
  br i1 false, label %gep.throw9, label %gep.next10

gep.next10:                                       ; preds = %if.done4
  call void @runtime.trackPointer(ptr %16, ptr undef) #14
  br i1 false, label %store.throw, label %store.next

store.next:                                       ; preds = %gep.next10
  %20 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 5
  store i32 ptrtoint (ptr @"reflect/types.type:pointer:named:os.File" to i32), ptr %20, align 4
  %.repack42 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 5, i32 1
  store ptr %16, ptr %.repack42, align 4
  br i1 false, label %gep.throw11, label %gep.next12

gep.next12:                                       ; preds = %store.next
  %21 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14
  br i1 false, label %gep.throw13, label %gep.next14

gep.next14:                                       ; preds = %gep.next12
  br i1 false, label %deref.throw15, label %deref.next16

deref.next16:                                     ; preds = %gep.next14
  %22 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14
  %.unpack44 = load ptr, ptr %22, align 4
  %.elt45 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 1
  %.unpack46 = load i32, ptr %.elt45, align 4
  %.elt47 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 2
  %.unpack48 = load i32, ptr %.elt47, align 4
  call void @runtime.trackPointer(ptr %.unpack44, ptr undef) #14
  %varargs = call ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 133 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %varargs, ptr undef) #14
  call void @runtime.trackPointer(ptr %16, ptr undef) #14
  store i32 ptrtoint (ptr @"reflect/types.type:pointer:named:os.File" to i32), ptr %varargs, align 4
  %varargs.repack49 = getelementptr inbounds %runtime._interface, ptr %varargs, i32 0, i32 1
  store ptr %16, ptr %varargs.repack49, align 4
  %append.new = call { ptr, i32, i32 } @runtime.sliceAppend(ptr %.unpack44, ptr nonnull %varargs, i32 %.unpack46, i32 %.unpack48, i32 1, i32 8, ptr undef) #14
  %append.newPtr = extractvalue { ptr, i32, i32 } %append.new, 0
  call void @runtime.trackPointer(ptr %append.newPtr, ptr undef) #14
  br i1 false, label %store.throw17, label %store.next18

store.next18:                                     ; preds = %deref.next16
  %append.newLen = extractvalue { ptr, i32, i32 } %append.new, 1
  %append.newCap = extractvalue { ptr, i32, i32 } %append.new, 2
  store ptr %append.newPtr, ptr %21, align 4
  %.repack51 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 1
  store i32 %append.newLen, ptr %.repack51, align 4
  %.repack53 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 2
  store i32 %append.newCap, ptr %.repack53, align 4
  br i1 false, label %gep.throw19, label %gep.next20

gep.next20:                                       ; preds = %store.next18
  br i1 false, label %gep.throw21, label %gep.next22

gep.next22:                                       ; preds = %gep.next20
  br i1 false, label %deref.throw23, label %deref.next24

deref.next24:                                     ; preds = %gep.next22
  %23 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15
  %.unpack55 = load ptr, ptr %23, align 4
  %.elt56 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 1
  %.unpack57 = load i32, ptr %.elt56, align 4
  %.elt58 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 2
  %.unpack59 = load i32, ptr %.elt58, align 4
  call void @runtime.trackPointer(ptr %.unpack55, ptr undef) #14
  %varargs25 = call ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 133 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %varargs25, ptr undef) #14
  call void @runtime.trackPointer(ptr %15, ptr undef) #14
  store i32 ptrtoint (ptr @"reflect/types.type:pointer:named:os.File" to i32), ptr %varargs25, align 4
  %varargs25.repack60 = getelementptr inbounds %runtime._interface, ptr %varargs25, i32 0, i32 1
  store ptr %15, ptr %varargs25.repack60, align 4
  %append.new32 = call { ptr, i32, i32 } @runtime.sliceAppend(ptr %.unpack55, ptr nonnull %varargs25, i32 %.unpack57, i32 %.unpack59, i32 1, i32 8, ptr undef) #14
  %append.newPtr33 = extractvalue { ptr, i32, i32 } %append.new32, 0
  call void @runtime.trackPointer(ptr %append.newPtr33, ptr undef) #14
  br i1 false, label %store.throw36, label %store.next37

store.next37:                                     ; preds = %deref.next24
  %24 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15
  %append.newLen34 = extractvalue { ptr, i32, i32 } %append.new32, 1
  %append.newCap35 = extractvalue { ptr, i32, i32 } %append.new32, 2
  store ptr %append.newPtr33, ptr %24, align 4
  %.repack62 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 1
  store i32 %append.newLen34, ptr %.repack62, align 4
  %.repack64 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 2
  store i32 %append.newCap35, ptr %.repack64, align 4
  %25 = insertvalue %runtime._interface { i32 ptrtoint (ptr @"reflect/types.type:pointer:named:os.File" to i32), ptr undef }, ptr %15, 1
  call void @runtime.trackPointer(ptr %15, ptr undef) #14
  %26 = insertvalue { %runtime._interface, %runtime._interface } zeroinitializer, %runtime._interface %25, 0
  %27 = insertvalue { %runtime._interface, %runtime._interface } %26, %runtime._interface zeroinitializer, 1
  ret { %runtime._interface, %runtime._interface } %27

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable

gep.throw5:                                       ; preds = %if.done
  unreachable

deref.throw7:                                     ; preds = %gep.next6
  unreachable

gep.throw9:                                       ; preds = %if.done4
  unreachable

store.throw:                                      ; preds = %gep.next10
  unreachable

gep.throw11:                                      ; preds = %store.next
  unreachable

gep.throw13:                                      ; preds = %gep.next12
  unreachable

deref.throw15:                                    ; preds = %gep.next14
  unreachable

store.throw17:                                    ; preds = %deref.next16
  unreachable

gep.throw19:                                      ; preds = %store.next18
  unreachable

gep.throw21:                                      ; preds = %gep.next20
  unreachable

deref.throw23:                                    ; preds = %gep.next22
  unreachable

store.throw36:                                    ; preds = %deref.next24
  unreachable
}

; Function Attrs: nounwind
define hidden %runtime._string @"(*main.Cmd).String"(ptr dereferenceable_or_null(168) %c, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %c, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 12
  %.unpack = load i32, ptr %1, align 4
  %.elt17 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 12, i32 1
  %.unpack18 = load ptr, ptr %.elt17, align 4
  call void @runtime.trackPointer(ptr %.unpack18, ptr undef) #14
  %.not = icmp eq i32 %.unpack, 0
  br i1 %.not, label %cond.false, label %if.then

if.then:                                          ; preds = %deref.next8, %deref.next
  br i1 false, label %gep.throw1, label %gep.next2

gep.next2:                                        ; preds = %if.then
  br i1 false, label %deref.throw3, label %deref.next4

deref.next4:                                      ; preds = %gep.next2
  %2 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 1
  %.unpack36 = load ptr, ptr %2, align 4
  %.elt37 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 1, i32 1
  %.unpack38 = load i32, ptr %.elt37, align 4
  %.elt39 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 1, i32 2
  %.unpack40 = load i32, ptr %.elt39, align 4
  call void @runtime.trackPointer(ptr %.unpack36, ptr undef) #14
  %3 = call %runtime._string @strings.Join(ptr %.unpack36, i32 %.unpack38, i32 %.unpack40, ptr nonnull @"main$string.57", i32 1, ptr undef) #14
  %4 = extractvalue %runtime._string %3, 0
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  ret %runtime._string %3

cond.false:                                       ; preds = %deref.next
  br i1 false, label %gep.throw5, label %gep.next6

gep.next6:                                        ; preds = %cond.false
  br i1 false, label %deref.throw7, label %deref.next8

deref.next8:                                      ; preds = %gep.next6
  %5 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 19
  %.unpack19 = load i32, ptr %5, align 4
  %.elt20 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 19, i32 1
  %.unpack21 = load ptr, ptr %.elt20, align 4
  call void @runtime.trackPointer(ptr %.unpack21, ptr undef) #14
  %.not22 = icmp eq i32 %.unpack19, 0
  br i1 %.not22, label %if.done, label %if.then

if.done:                                          ; preds = %deref.next8
  %new = call ptr @runtime.alloc(i32 16, ptr nonnull inttoptr (i32 201 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %new, ptr undef) #14
  br i1 false, label %gep.throw9, label %gep.next10

gep.next10:                                       ; preds = %if.done
  br i1 false, label %deref.throw11, label %deref.next12

deref.next12:                                     ; preds = %gep.next10
  %.unpack23 = load ptr, ptr %c, align 4
  %.elt24 = getelementptr inbounds %runtime._string, ptr %c, i32 0, i32 1
  %.unpack25 = load i32, ptr %.elt24, align 4
  call void @runtime.trackPointer(ptr %.unpack23, ptr undef) #14
  %6 = call { i32, %runtime._interface } @"(*strings.Builder).WriteString"(ptr nonnull %new, ptr %.unpack23, i32 %.unpack25, ptr undef) #14
  %7 = extractvalue { i32, %runtime._interface } %6, 1
  %8 = extractvalue %runtime._interface %7, 1
  call void @runtime.trackPointer(ptr %8, ptr undef) #14
  br i1 false, label %gep.throw13, label %gep.next14

gep.next14:                                       ; preds = %deref.next12
  br i1 false, label %deref.throw15, label %deref.next16

deref.next16:                                     ; preds = %gep.next14
  %9 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 1
  %.unpack26 = load ptr, ptr %9, align 4
  %.elt27 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 1, i32 1
  %.unpack28 = load i32, ptr %.elt27, align 4
  %.elt29 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 1, i32 2
  %.unpack30 = load i32, ptr %.elt29, align 4
  call void @runtime.trackPointer(ptr %.unpack26, ptr undef) #14
  %10 = add i32 %.unpack28, -1
  %.not31 = icmp ult i32 %10, %.unpack30
  br i1 %.not31, label %slice.next, label %slice.throw

slice.next:                                       ; preds = %deref.next16
  %11 = getelementptr inbounds %runtime._string, ptr %.unpack26, i32 1
  %12 = add i32 %.unpack28, -1
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %lookup.next, %slice.next
  %13 = phi i32 [ -1, %slice.next ], [ %14, %lookup.next ]
  %14 = add i32 %13, 1
  %15 = icmp slt i32 %14, %12
  br i1 %15, label %rangeindex.body, label %rangeindex.done

rangeindex.body:                                  ; preds = %rangeindex.loop
  %.not32 = icmp ult i32 %14, %12
  br i1 %.not32, label %lookup.next, label %lookup.throw

lookup.next:                                      ; preds = %rangeindex.body
  %16 = getelementptr inbounds %runtime._string, ptr %11, i32 %14
  %.unpack33 = load ptr, ptr %16, align 4
  %.elt34 = getelementptr inbounds %runtime._string, ptr %16, i32 0, i32 1
  %.unpack35 = load i32, ptr %.elt34, align 4
  call void @runtime.trackPointer(ptr %.unpack33, ptr undef) #14
  %17 = call %runtime._interface @"(*strings.Builder).WriteByte"(ptr nonnull %new, i8 32, ptr undef) #14
  %18 = extractvalue %runtime._interface %17, 1
  call void @runtime.trackPointer(ptr %18, ptr undef) #14
  %19 = call { i32, %runtime._interface } @"(*strings.Builder).WriteString"(ptr nonnull %new, ptr %.unpack33, i32 %.unpack35, ptr undef) #14
  %20 = extractvalue { i32, %runtime._interface } %19, 1
  %21 = extractvalue %runtime._interface %20, 1
  call void @runtime.trackPointer(ptr %21, ptr undef) #14
  br label %rangeindex.loop

rangeindex.done:                                  ; preds = %rangeindex.loop
  %22 = call %runtime._string @"(*strings.Builder).String"(ptr nonnull %new, ptr undef) #14
  %23 = extractvalue %runtime._string %22, 0
  call void @runtime.trackPointer(ptr %23, ptr undef) #14
  ret %runtime._string %22

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable

gep.throw1:                                       ; preds = %if.then
  unreachable

deref.throw3:                                     ; preds = %gep.next2
  unreachable

gep.throw5:                                       ; preds = %cond.false
  unreachable

deref.throw7:                                     ; preds = %gep.next6
  unreachable

gep.throw9:                                       ; preds = %if.done
  unreachable

deref.throw11:                                    ; preds = %gep.next10
  unreachable

gep.throw13:                                      ; preds = %deref.next12
  unreachable

deref.throw15:                                    ; preds = %gep.next14
  unreachable

slice.throw:                                      ; preds = %deref.next16
  call void @runtime.slicePanic(ptr undef) #14
  unreachable

lookup.throw:                                     ; preds = %rangeindex.body
  call void @runtime.lookupPanic(ptr undef) #14
  unreachable
}

declare %runtime._string @strings.Join(ptr, i32, i32, ptr, i32, ptr) #0

declare { i32, %runtime._interface } @"(*strings.Builder).WriteString"(ptr dereferenceable_or_null(16), ptr, i32, ptr) #0

declare %runtime._interface @"(*strings.Builder).WriteByte"(ptr dereferenceable_or_null(16), i8, ptr) #0

declare %runtime._string @"(*strings.Builder).String"(ptr dereferenceable_or_null(16), ptr) #0

declare { ptr, %runtime._interface } @"(*os.Process).Wait"(ptr dereferenceable_or_null(4), ptr) #0

declare i1 @"(*os.ProcessState).Success"(ptr, ptr) #0

declare i32 @"(*os.ProcessState).ExitCode"(ptr, ptr) #0

declare %runtime._string @"(*os.ProcessState).String"(ptr, ptr) #0

declare %runtime._interface @"(*os.ProcessState).Sys"(ptr, ptr) #0

; Function Attrs: nounwind
define linkonce_odr hidden i32 @"(struct{*os.ProcessState; Stderr []byte}).ExitCode$wrapper for func (*os.ProcessState).ExitCode() int"({ ptr, { ptr, i32, i32 } } %p, ptr %context) unnamed_addr #1 {
entry:
  %p1 = alloca { ptr, { ptr, i32, i32 } }, align 8
  store ptr null, ptr %p1, align 8
  %p1.repack2 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %p1, i32 0, i32 1
  store ptr null, ptr %p1.repack2, align 4
  %p1.repack2.repack3 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %p1, i32 0, i32 1, i32 1
  store i32 0, ptr %p1.repack2.repack3, align 8
  %p1.repack2.repack4 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %p1, i32 0, i32 1, i32 2
  store i32 0, ptr %p1.repack2.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %p1, ptr undef) #14
  %p.elt = extractvalue { ptr, { ptr, i32, i32 } } %p, 0
  store ptr %p.elt, ptr %p1, align 8
  %p1.repack5 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %p1, i32 0, i32 1
  %p.elt6 = extractvalue { ptr, { ptr, i32, i32 } } %p, 1
  %p.elt6.elt = extractvalue { ptr, i32, i32 } %p.elt6, 0
  store ptr %p.elt6.elt, ptr %p1.repack5, align 4
  %p1.repack5.repack7 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %p1, i32 0, i32 1, i32 1
  %p.elt6.elt8 = extractvalue { ptr, i32, i32 } %p.elt6, 1
  store i32 %p.elt6.elt8, ptr %p1.repack5.repack7, align 8
  %p1.repack5.repack9 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %p1, i32 0, i32 1, i32 2
  %p.elt6.elt10 = extractvalue { ptr, i32, i32 } %p.elt6, 2
  store i32 %p.elt6.elt10, ptr %p1.repack5.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %p1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call i32 @"(*os.ProcessState).ExitCode"(ptr %0, ptr undef) #14
  ret i32 %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr i32 @"(struct{*os.ProcessState; Stderr []byte}).ExitCode$wrapper for func (*os.ProcessState).ExitCode() int$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %2 = insertvalue { ptr, { ptr, i32, i32 } } undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %0, i32 0, i32 1
  %.unpack2.unpack = load ptr, ptr %.elt1, align 4
  %3 = insertvalue { ptr, i32, i32 } undef, ptr %.unpack2.unpack, 0
  %.unpack2.elt3 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %0, i32 0, i32 1, i32 1
  %.unpack2.unpack4 = load i32, ptr %.unpack2.elt3, align 4
  %4 = insertvalue { ptr, i32, i32 } %3, i32 %.unpack2.unpack4, 1
  %.unpack2.elt5 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %0, i32 0, i32 1, i32 2
  %.unpack2.unpack6 = load i32, ptr %.unpack2.elt5, align 4
  %.unpack27 = insertvalue { ptr, i32, i32 } %4, i32 %.unpack2.unpack6, 2
  %5 = insertvalue { ptr, { ptr, i32, i32 } } %2, { ptr, i32, i32 } %.unpack27, 1
  %ret = call i32 @"(struct{*os.ProcessState; Stderr []byte}).ExitCode$wrapper for func (*os.ProcessState).ExitCode() int"({ ptr, { ptr, i32, i32 } } %5, ptr %1)
  ret i32 %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._string @"(struct{*os.ProcessState; Stderr []byte}).String$wrapper for func (*os.ProcessState).String() string"({ ptr, { ptr, i32, i32 } } %p, ptr %context) unnamed_addr #1 {
entry:
  %p1 = alloca { ptr, { ptr, i32, i32 } }, align 8
  store ptr null, ptr %p1, align 8
  %p1.repack2 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %p1, i32 0, i32 1
  store ptr null, ptr %p1.repack2, align 4
  %p1.repack2.repack3 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %p1, i32 0, i32 1, i32 1
  store i32 0, ptr %p1.repack2.repack3, align 8
  %p1.repack2.repack4 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %p1, i32 0, i32 1, i32 2
  store i32 0, ptr %p1.repack2.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %p1, ptr undef) #14
  %p.elt = extractvalue { ptr, { ptr, i32, i32 } } %p, 0
  store ptr %p.elt, ptr %p1, align 8
  %p1.repack5 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %p1, i32 0, i32 1
  %p.elt6 = extractvalue { ptr, { ptr, i32, i32 } } %p, 1
  %p.elt6.elt = extractvalue { ptr, i32, i32 } %p.elt6, 0
  store ptr %p.elt6.elt, ptr %p1.repack5, align 4
  %p1.repack5.repack7 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %p1, i32 0, i32 1, i32 1
  %p.elt6.elt8 = extractvalue { ptr, i32, i32 } %p.elt6, 1
  store i32 %p.elt6.elt8, ptr %p1.repack5.repack7, align 8
  %p1.repack5.repack9 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %p1, i32 0, i32 1, i32 2
  %p.elt6.elt10 = extractvalue { ptr, i32, i32 } %p.elt6, 2
  store i32 %p.elt6.elt10, ptr %p1.repack5.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %p1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call %runtime._string @"(*os.ProcessState).String"(ptr %0, ptr undef) #14
  %2 = extractvalue %runtime._string %1, 0
  call void @runtime.trackPointer(ptr %2, ptr undef) #14
  ret %runtime._string %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr %runtime._string @"(struct{*os.ProcessState; Stderr []byte}).String$wrapper for func (*os.ProcessState).String() string$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %2 = insertvalue { ptr, { ptr, i32, i32 } } undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %0, i32 0, i32 1
  %.unpack2.unpack = load ptr, ptr %.elt1, align 4
  %3 = insertvalue { ptr, i32, i32 } undef, ptr %.unpack2.unpack, 0
  %.unpack2.elt3 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %0, i32 0, i32 1, i32 1
  %.unpack2.unpack4 = load i32, ptr %.unpack2.elt3, align 4
  %4 = insertvalue { ptr, i32, i32 } %3, i32 %.unpack2.unpack4, 1
  %.unpack2.elt5 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %0, i32 0, i32 1, i32 2
  %.unpack2.unpack6 = load i32, ptr %.unpack2.elt5, align 4
  %.unpack27 = insertvalue { ptr, i32, i32 } %4, i32 %.unpack2.unpack6, 2
  %5 = insertvalue { ptr, { ptr, i32, i32 } } %2, { ptr, i32, i32 } %.unpack27, 1
  %ret = call %runtime._string @"(struct{*os.ProcessState; Stderr []byte}).String$wrapper for func (*os.ProcessState).String() string"({ ptr, { ptr, i32, i32 } } %5, ptr %1)
  ret %runtime._string %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden i1 @"(struct{*os.ProcessState; Stderr []byte}).Success$wrapper for func (*os.ProcessState).Success() bool"({ ptr, { ptr, i32, i32 } } %p, ptr %context) unnamed_addr #1 {
entry:
  %p1 = alloca { ptr, { ptr, i32, i32 } }, align 8
  store ptr null, ptr %p1, align 8
  %p1.repack2 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %p1, i32 0, i32 1
  store ptr null, ptr %p1.repack2, align 4
  %p1.repack2.repack3 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %p1, i32 0, i32 1, i32 1
  store i32 0, ptr %p1.repack2.repack3, align 8
  %p1.repack2.repack4 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %p1, i32 0, i32 1, i32 2
  store i32 0, ptr %p1.repack2.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %p1, ptr undef) #14
  %p.elt = extractvalue { ptr, { ptr, i32, i32 } } %p, 0
  store ptr %p.elt, ptr %p1, align 8
  %p1.repack5 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %p1, i32 0, i32 1
  %p.elt6 = extractvalue { ptr, { ptr, i32, i32 } } %p, 1
  %p.elt6.elt = extractvalue { ptr, i32, i32 } %p.elt6, 0
  store ptr %p.elt6.elt, ptr %p1.repack5, align 4
  %p1.repack5.repack7 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %p1, i32 0, i32 1, i32 1
  %p.elt6.elt8 = extractvalue { ptr, i32, i32 } %p.elt6, 1
  store i32 %p.elt6.elt8, ptr %p1.repack5.repack7, align 8
  %p1.repack5.repack9 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %p1, i32 0, i32 1, i32 2
  %p.elt6.elt10 = extractvalue { ptr, i32, i32 } %p.elt6, 2
  store i32 %p.elt6.elt10, ptr %p1.repack5.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %p1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call i1 @"(*os.ProcessState).Success"(ptr %0, ptr undef) #14
  ret i1 %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr i1 @"(struct{*os.ProcessState; Stderr []byte}).Success$wrapper for func (*os.ProcessState).Success() bool$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %2 = insertvalue { ptr, { ptr, i32, i32 } } undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %0, i32 0, i32 1
  %.unpack2.unpack = load ptr, ptr %.elt1, align 4
  %3 = insertvalue { ptr, i32, i32 } undef, ptr %.unpack2.unpack, 0
  %.unpack2.elt3 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %0, i32 0, i32 1, i32 1
  %.unpack2.unpack4 = load i32, ptr %.unpack2.elt3, align 4
  %4 = insertvalue { ptr, i32, i32 } %3, i32 %.unpack2.unpack4, 1
  %.unpack2.elt5 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %0, i32 0, i32 1, i32 2
  %.unpack2.unpack6 = load i32, ptr %.unpack2.elt5, align 4
  %.unpack27 = insertvalue { ptr, i32, i32 } %4, i32 %.unpack2.unpack6, 2
  %5 = insertvalue { ptr, { ptr, i32, i32 } } %2, { ptr, i32, i32 } %.unpack27, 1
  %ret = call i1 @"(struct{*os.ProcessState; Stderr []byte}).Success$wrapper for func (*os.ProcessState).Success() bool"({ ptr, { ptr, i32, i32 } } %5, ptr %1)
  ret i1 %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(struct{*os.ProcessState; Stderr []byte}).Sys$wrapper for func (*os.ProcessState).Sys() interface{}"({ ptr, { ptr, i32, i32 } } %p, ptr %context) unnamed_addr #1 {
entry:
  %p1 = alloca { ptr, { ptr, i32, i32 } }, align 8
  store ptr null, ptr %p1, align 8
  %p1.repack2 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %p1, i32 0, i32 1
  store ptr null, ptr %p1.repack2, align 4
  %p1.repack2.repack3 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %p1, i32 0, i32 1, i32 1
  store i32 0, ptr %p1.repack2.repack3, align 8
  %p1.repack2.repack4 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %p1, i32 0, i32 1, i32 2
  store i32 0, ptr %p1.repack2.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %p1, ptr undef) #14
  %p.elt = extractvalue { ptr, { ptr, i32, i32 } } %p, 0
  store ptr %p.elt, ptr %p1, align 8
  %p1.repack5 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %p1, i32 0, i32 1
  %p.elt6 = extractvalue { ptr, { ptr, i32, i32 } } %p, 1
  %p.elt6.elt = extractvalue { ptr, i32, i32 } %p.elt6, 0
  store ptr %p.elt6.elt, ptr %p1.repack5, align 4
  %p1.repack5.repack7 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %p1, i32 0, i32 1, i32 1
  %p.elt6.elt8 = extractvalue { ptr, i32, i32 } %p.elt6, 1
  store i32 %p.elt6.elt8, ptr %p1.repack5.repack7, align 8
  %p1.repack5.repack9 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %p1, i32 0, i32 1, i32 2
  %p.elt6.elt10 = extractvalue { ptr, i32, i32 } %p.elt6, 2
  store i32 %p.elt6.elt10, ptr %p1.repack5.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %p1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call %runtime._interface @"(*os.ProcessState).Sys"(ptr %0, ptr undef) #14
  %2 = extractvalue %runtime._interface %1, 1
  call void @runtime.trackPointer(ptr %2, ptr undef) #14
  ret %runtime._interface %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr %runtime._interface @"(struct{*os.ProcessState; Stderr []byte}).Sys$wrapper for func (*os.ProcessState).Sys() interface{}$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %2 = insertvalue { ptr, { ptr, i32, i32 } } undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %0, i32 0, i32 1
  %.unpack2.unpack = load ptr, ptr %.elt1, align 4
  %3 = insertvalue { ptr, i32, i32 } undef, ptr %.unpack2.unpack, 0
  %.unpack2.elt3 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %0, i32 0, i32 1, i32 1
  %.unpack2.unpack4 = load i32, ptr %.unpack2.elt3, align 4
  %4 = insertvalue { ptr, i32, i32 } %3, i32 %.unpack2.unpack4, 1
  %.unpack2.elt5 = getelementptr inbounds { ptr, { ptr, i32, i32 } }, ptr %0, i32 0, i32 1, i32 2
  %.unpack2.unpack6 = load i32, ptr %.unpack2.elt5, align 4
  %.unpack27 = insertvalue { ptr, i32, i32 } %4, i32 %.unpack2.unpack6, 2
  %5 = insertvalue { ptr, { ptr, i32, i32 } } %2, { ptr, i32, i32 } %.unpack27, 1
  %ret = call %runtime._interface @"(struct{*os.ProcessState; Stderr []byte}).Sys$wrapper for func (*os.ProcessState).Sys() interface{}"({ ptr, { ptr, i32, i32 } } %5, ptr %1)
  ret %runtime._interface %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden i32 @"(*struct{*os.ProcessState; Stderr []byte}).ExitCode$wrapper for func (*os.ProcessState).ExitCode() int"(ptr dereferenceable_or_null(16) %p, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %p, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %p, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call i32 @"(*os.ProcessState).ExitCode"(ptr %1, ptr undef) #14
  ret i32 %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._string @"(*struct{*os.ProcessState; Stderr []byte}).String$wrapper for func (*os.ProcessState).String() string"(ptr dereferenceable_or_null(16) %p, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %p, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %p, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call %runtime._string @"(*os.ProcessState).String"(ptr %1, ptr undef) #14
  %3 = extractvalue %runtime._string %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret %runtime._string %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden i1 @"(*struct{*os.ProcessState; Stderr []byte}).Success$wrapper for func (*os.ProcessState).Success() bool"(ptr dereferenceable_or_null(16) %p, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %p, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %p, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call i1 @"(*os.ProcessState).Success"(ptr %1, ptr undef) #14
  ret i1 %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(*struct{*os.ProcessState; Stderr []byte}).Sys$wrapper for func (*os.ProcessState).Sys() interface{}"(ptr dereferenceable_or_null(16) %p, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %p, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %p, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call %runtime._interface @"(*os.ProcessState).Sys"(ptr %1, ptr undef) #14
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret %runtime._interface %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden i32 @"(main.ExitError).ExitCode$wrapper for func (*os.ProcessState).ExitCode() int"(%main.ExitError %p, ptr %context) unnamed_addr #1 {
entry:
  %p1 = alloca %main.ExitError, align 8
  store ptr null, ptr %p1, align 8
  %p1.repack2 = getelementptr inbounds %main.ExitError, ptr %p1, i32 0, i32 1
  store ptr null, ptr %p1.repack2, align 4
  %p1.repack2.repack3 = getelementptr inbounds %main.ExitError, ptr %p1, i32 0, i32 1, i32 1
  store i32 0, ptr %p1.repack2.repack3, align 8
  %p1.repack2.repack4 = getelementptr inbounds %main.ExitError, ptr %p1, i32 0, i32 1, i32 2
  store i32 0, ptr %p1.repack2.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %p1, ptr undef) #14
  %p.elt = extractvalue %main.ExitError %p, 0
  store ptr %p.elt, ptr %p1, align 8
  %p1.repack5 = getelementptr inbounds %main.ExitError, ptr %p1, i32 0, i32 1
  %p.elt6 = extractvalue %main.ExitError %p, 1
  %p.elt6.elt = extractvalue { ptr, i32, i32 } %p.elt6, 0
  store ptr %p.elt6.elt, ptr %p1.repack5, align 4
  %p1.repack5.repack7 = getelementptr inbounds %main.ExitError, ptr %p1, i32 0, i32 1, i32 1
  %p.elt6.elt8 = extractvalue { ptr, i32, i32 } %p.elt6, 1
  store i32 %p.elt6.elt8, ptr %p1.repack5.repack7, align 8
  %p1.repack5.repack9 = getelementptr inbounds %main.ExitError, ptr %p1, i32 0, i32 1, i32 2
  %p.elt6.elt10 = extractvalue { ptr, i32, i32 } %p.elt6, 2
  store i32 %p.elt6.elt10, ptr %p1.repack5.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %p1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call i32 @"(*os.ProcessState).ExitCode"(ptr %0, ptr undef) #14
  ret i32 %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr i32 @"(main.ExitError).ExitCode$wrapper for func (*os.ProcessState).ExitCode() int$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %2 = insertvalue %main.ExitError undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds %main.ExitError, ptr %0, i32 0, i32 1
  %.unpack2.unpack = load ptr, ptr %.elt1, align 4
  %3 = insertvalue { ptr, i32, i32 } undef, ptr %.unpack2.unpack, 0
  %.unpack2.elt3 = getelementptr inbounds %main.ExitError, ptr %0, i32 0, i32 1, i32 1
  %.unpack2.unpack4 = load i32, ptr %.unpack2.elt3, align 4
  %4 = insertvalue { ptr, i32, i32 } %3, i32 %.unpack2.unpack4, 1
  %.unpack2.elt5 = getelementptr inbounds %main.ExitError, ptr %0, i32 0, i32 1, i32 2
  %.unpack2.unpack6 = load i32, ptr %.unpack2.elt5, align 4
  %.unpack27 = insertvalue { ptr, i32, i32 } %4, i32 %.unpack2.unpack6, 2
  %5 = insertvalue %main.ExitError %2, { ptr, i32, i32 } %.unpack27, 1
  %ret = call i32 @"(main.ExitError).ExitCode$wrapper for func (*os.ProcessState).ExitCode() int"(%main.ExitError %5, ptr %1)
  ret i32 %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._string @"(main.ExitError).String$wrapper for func (*os.ProcessState).String() string"(%main.ExitError %p, ptr %context) unnamed_addr #1 {
entry:
  %p1 = alloca %main.ExitError, align 8
  store ptr null, ptr %p1, align 8
  %p1.repack2 = getelementptr inbounds %main.ExitError, ptr %p1, i32 0, i32 1
  store ptr null, ptr %p1.repack2, align 4
  %p1.repack2.repack3 = getelementptr inbounds %main.ExitError, ptr %p1, i32 0, i32 1, i32 1
  store i32 0, ptr %p1.repack2.repack3, align 8
  %p1.repack2.repack4 = getelementptr inbounds %main.ExitError, ptr %p1, i32 0, i32 1, i32 2
  store i32 0, ptr %p1.repack2.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %p1, ptr undef) #14
  %p.elt = extractvalue %main.ExitError %p, 0
  store ptr %p.elt, ptr %p1, align 8
  %p1.repack5 = getelementptr inbounds %main.ExitError, ptr %p1, i32 0, i32 1
  %p.elt6 = extractvalue %main.ExitError %p, 1
  %p.elt6.elt = extractvalue { ptr, i32, i32 } %p.elt6, 0
  store ptr %p.elt6.elt, ptr %p1.repack5, align 4
  %p1.repack5.repack7 = getelementptr inbounds %main.ExitError, ptr %p1, i32 0, i32 1, i32 1
  %p.elt6.elt8 = extractvalue { ptr, i32, i32 } %p.elt6, 1
  store i32 %p.elt6.elt8, ptr %p1.repack5.repack7, align 8
  %p1.repack5.repack9 = getelementptr inbounds %main.ExitError, ptr %p1, i32 0, i32 1, i32 2
  %p.elt6.elt10 = extractvalue { ptr, i32, i32 } %p.elt6, 2
  store i32 %p.elt6.elt10, ptr %p1.repack5.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %p1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call %runtime._string @"(*os.ProcessState).String"(ptr %0, ptr undef) #14
  %2 = extractvalue %runtime._string %1, 0
  call void @runtime.trackPointer(ptr %2, ptr undef) #14
  ret %runtime._string %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr %runtime._string @"(main.ExitError).String$wrapper for func (*os.ProcessState).String() string$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %2 = insertvalue %main.ExitError undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds %main.ExitError, ptr %0, i32 0, i32 1
  %.unpack2.unpack = load ptr, ptr %.elt1, align 4
  %3 = insertvalue { ptr, i32, i32 } undef, ptr %.unpack2.unpack, 0
  %.unpack2.elt3 = getelementptr inbounds %main.ExitError, ptr %0, i32 0, i32 1, i32 1
  %.unpack2.unpack4 = load i32, ptr %.unpack2.elt3, align 4
  %4 = insertvalue { ptr, i32, i32 } %3, i32 %.unpack2.unpack4, 1
  %.unpack2.elt5 = getelementptr inbounds %main.ExitError, ptr %0, i32 0, i32 1, i32 2
  %.unpack2.unpack6 = load i32, ptr %.unpack2.elt5, align 4
  %.unpack27 = insertvalue { ptr, i32, i32 } %4, i32 %.unpack2.unpack6, 2
  %5 = insertvalue %main.ExitError %2, { ptr, i32, i32 } %.unpack27, 1
  %ret = call %runtime._string @"(main.ExitError).String$wrapper for func (*os.ProcessState).String() string"(%main.ExitError %5, ptr %1)
  ret %runtime._string %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden i1 @"(main.ExitError).Success$wrapper for func (*os.ProcessState).Success() bool"(%main.ExitError %p, ptr %context) unnamed_addr #1 {
entry:
  %p1 = alloca %main.ExitError, align 8
  store ptr null, ptr %p1, align 8
  %p1.repack2 = getelementptr inbounds %main.ExitError, ptr %p1, i32 0, i32 1
  store ptr null, ptr %p1.repack2, align 4
  %p1.repack2.repack3 = getelementptr inbounds %main.ExitError, ptr %p1, i32 0, i32 1, i32 1
  store i32 0, ptr %p1.repack2.repack3, align 8
  %p1.repack2.repack4 = getelementptr inbounds %main.ExitError, ptr %p1, i32 0, i32 1, i32 2
  store i32 0, ptr %p1.repack2.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %p1, ptr undef) #14
  %p.elt = extractvalue %main.ExitError %p, 0
  store ptr %p.elt, ptr %p1, align 8
  %p1.repack5 = getelementptr inbounds %main.ExitError, ptr %p1, i32 0, i32 1
  %p.elt6 = extractvalue %main.ExitError %p, 1
  %p.elt6.elt = extractvalue { ptr, i32, i32 } %p.elt6, 0
  store ptr %p.elt6.elt, ptr %p1.repack5, align 4
  %p1.repack5.repack7 = getelementptr inbounds %main.ExitError, ptr %p1, i32 0, i32 1, i32 1
  %p.elt6.elt8 = extractvalue { ptr, i32, i32 } %p.elt6, 1
  store i32 %p.elt6.elt8, ptr %p1.repack5.repack7, align 8
  %p1.repack5.repack9 = getelementptr inbounds %main.ExitError, ptr %p1, i32 0, i32 1, i32 2
  %p.elt6.elt10 = extractvalue { ptr, i32, i32 } %p.elt6, 2
  store i32 %p.elt6.elt10, ptr %p1.repack5.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %p1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call i1 @"(*os.ProcessState).Success"(ptr %0, ptr undef) #14
  ret i1 %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr i1 @"(main.ExitError).Success$wrapper for func (*os.ProcessState).Success() bool$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %2 = insertvalue %main.ExitError undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds %main.ExitError, ptr %0, i32 0, i32 1
  %.unpack2.unpack = load ptr, ptr %.elt1, align 4
  %3 = insertvalue { ptr, i32, i32 } undef, ptr %.unpack2.unpack, 0
  %.unpack2.elt3 = getelementptr inbounds %main.ExitError, ptr %0, i32 0, i32 1, i32 1
  %.unpack2.unpack4 = load i32, ptr %.unpack2.elt3, align 4
  %4 = insertvalue { ptr, i32, i32 } %3, i32 %.unpack2.unpack4, 1
  %.unpack2.elt5 = getelementptr inbounds %main.ExitError, ptr %0, i32 0, i32 1, i32 2
  %.unpack2.unpack6 = load i32, ptr %.unpack2.elt5, align 4
  %.unpack27 = insertvalue { ptr, i32, i32 } %4, i32 %.unpack2.unpack6, 2
  %5 = insertvalue %main.ExitError %2, { ptr, i32, i32 } %.unpack27, 1
  %ret = call i1 @"(main.ExitError).Success$wrapper for func (*os.ProcessState).Success() bool"(%main.ExitError %5, ptr %1)
  ret i1 %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(main.ExitError).Sys$wrapper for func (*os.ProcessState).Sys() interface{}"(%main.ExitError %p, ptr %context) unnamed_addr #1 {
entry:
  %p1 = alloca %main.ExitError, align 8
  store ptr null, ptr %p1, align 8
  %p1.repack2 = getelementptr inbounds %main.ExitError, ptr %p1, i32 0, i32 1
  store ptr null, ptr %p1.repack2, align 4
  %p1.repack2.repack3 = getelementptr inbounds %main.ExitError, ptr %p1, i32 0, i32 1, i32 1
  store i32 0, ptr %p1.repack2.repack3, align 8
  %p1.repack2.repack4 = getelementptr inbounds %main.ExitError, ptr %p1, i32 0, i32 1, i32 2
  store i32 0, ptr %p1.repack2.repack4, align 4
  call void @runtime.trackPointer(ptr nonnull %p1, ptr undef) #14
  %p.elt = extractvalue %main.ExitError %p, 0
  store ptr %p.elt, ptr %p1, align 8
  %p1.repack5 = getelementptr inbounds %main.ExitError, ptr %p1, i32 0, i32 1
  %p.elt6 = extractvalue %main.ExitError %p, 1
  %p.elt6.elt = extractvalue { ptr, i32, i32 } %p.elt6, 0
  store ptr %p.elt6.elt, ptr %p1.repack5, align 4
  %p1.repack5.repack7 = getelementptr inbounds %main.ExitError, ptr %p1, i32 0, i32 1, i32 1
  %p.elt6.elt8 = extractvalue { ptr, i32, i32 } %p.elt6, 1
  store i32 %p.elt6.elt8, ptr %p1.repack5.repack7, align 8
  %p1.repack5.repack9 = getelementptr inbounds %main.ExitError, ptr %p1, i32 0, i32 1, i32 2
  %p.elt6.elt10 = extractvalue { ptr, i32, i32 } %p.elt6, 2
  store i32 %p.elt6.elt10, ptr %p1.repack5.repack9, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %p1, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #14
  %1 = call %runtime._interface @"(*os.ProcessState).Sys"(ptr %0, ptr undef) #14
  %2 = extractvalue %runtime._interface %1, 1
  call void @runtime.trackPointer(ptr %2, ptr undef) #14
  ret %runtime._interface %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr %runtime._interface @"(main.ExitError).Sys$wrapper for func (*os.ProcessState).Sys() interface{}$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
entry:
  %.unpack = load ptr, ptr %0, align 4
  %2 = insertvalue %main.ExitError undef, ptr %.unpack, 0
  %.elt1 = getelementptr inbounds %main.ExitError, ptr %0, i32 0, i32 1
  %.unpack2.unpack = load ptr, ptr %.elt1, align 4
  %3 = insertvalue { ptr, i32, i32 } undef, ptr %.unpack2.unpack, 0
  %.unpack2.elt3 = getelementptr inbounds %main.ExitError, ptr %0, i32 0, i32 1, i32 1
  %.unpack2.unpack4 = load i32, ptr %.unpack2.elt3, align 4
  %4 = insertvalue { ptr, i32, i32 } %3, i32 %.unpack2.unpack4, 1
  %.unpack2.elt5 = getelementptr inbounds %main.ExitError, ptr %0, i32 0, i32 1, i32 2
  %.unpack2.unpack6 = load i32, ptr %.unpack2.elt5, align 4
  %.unpack27 = insertvalue { ptr, i32, i32 } %4, i32 %.unpack2.unpack6, 2
  %5 = insertvalue %main.ExitError %2, { ptr, i32, i32 } %.unpack27, 1
  %ret = call %runtime._interface @"(main.ExitError).Sys$wrapper for func (*os.ProcessState).Sys() interface{}"(%main.ExitError %5, ptr %1)
  ret %runtime._interface %ret
}

; Function Attrs: nounwind
define hidden %runtime._string @"(*main.ExitError).Error"(ptr dereferenceable_or_null(16) %e, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %e, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %e, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call %runtime._string @"(*os.ProcessState).String"(ptr %1, ptr undef) #14
  %3 = extractvalue %runtime._string %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret %runtime._string %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden i32 @"(*main.ExitError).ExitCode$wrapper for func (*os.ProcessState).ExitCode() int"(ptr dereferenceable_or_null(16) %p, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %p, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %p, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call i32 @"(*os.ProcessState).ExitCode"(ptr %1, ptr undef) #14
  ret i32 %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._string @"(*main.ExitError).String$wrapper for func (*os.ProcessState).String() string"(ptr dereferenceable_or_null(16) %p, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %p, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %p, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call %runtime._string @"(*os.ProcessState).String"(ptr %1, ptr undef) #14
  %3 = extractvalue %runtime._string %2, 0
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret %runtime._string %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden i1 @"(*main.ExitError).Success$wrapper for func (*os.ProcessState).Success() bool"(ptr dereferenceable_or_null(16) %p, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %p, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %p, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call i1 @"(*os.ProcessState).Success"(ptr %1, ptr undef) #14
  ret i1 %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(*main.ExitError).Sys$wrapper for func (*os.ProcessState).Sys() interface{}"(ptr dereferenceable_or_null(16) %p, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %p, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %p, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = call %runtime._interface @"(*os.ProcessState).Sys"(ptr %1, ptr undef) #14
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  ret %runtime._interface %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

declare i1 @runtime.chanRecv(ptr dereferenceable_or_null(32), ptr, ptr dereferenceable_or_null(24), ptr) #0

declare %runtime._interface @"interface:{Close:func:{}{named:error}}.Close$invoke"(ptr, i32, ptr) #12

; Function Attrs: nounwind
define hidden { ptr, i32, i32 } @main.dedupEnv(ptr %env.data, i32 %env.len, i32 %env.cap, ptr %context) unnamed_addr #1 {
entry:
  %0 = call { ptr, i32, i32 } @main.dedupEnvCase(i1 false, ptr %env.data, i32 %env.len, i32 %env.cap, ptr undef)
  %1 = extractvalue { ptr, i32, i32 } %0, 0
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  ret { ptr, i32, i32 } %0
}

; Function Attrs: nounwind
define hidden { ptr, i32, i32 } @main.addCriticalEnv(ptr %env.data, i32 %env.len, i32 %env.cap, ptr %context) unnamed_addr #1 {
entry:
  %0 = insertvalue { ptr, i32, i32 } zeroinitializer, ptr %env.data, 0
  %1 = insertvalue { ptr, i32, i32 } %0, i32 %env.len, 1
  %2 = insertvalue { ptr, i32, i32 } %1, i32 %env.cap, 2
  br i1 true, label %if.then, label %if.done

if.then:                                          ; preds = %entry
  ret { ptr, i32, i32 } %2

if.done:                                          ; preds = %entry
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %if.done1, %lookup.next, %if.done
  br i1 poison, label %rangeindex.body, label %rangeindex.done

rangeindex.body:                                  ; preds = %rangeindex.loop
  br i1 poison, label %lookup.throw, label %lookup.next

lookup.next:                                      ; preds = %rangeindex.body
  br i1 poison, label %if.done1, label %rangeindex.loop

if.done1:                                         ; preds = %lookup.next
  br i1 poison, label %if.then2, label %rangeindex.loop

if.then2:                                         ; preds = %if.done1
  ret { ptr, i32, i32 } %2

rangeindex.done:                                  ; preds = %rangeindex.loop
  ret { ptr, i32, i32 } poison

lookup.throw:                                     ; preds = %rangeindex.body
  unreachable
}

; Function Attrs: nounwind
define hidden i1 @main.interfaceEqual(i32 %a.typecode, ptr %a.value, i32 %b.typecode, ptr %b.value, ptr %context) unnamed_addr #1 {
entry:
  %defer.alloca = alloca { i32, ptr }, align 8
  %deferPtr = alloca ptr, align 4
  store ptr null, ptr %deferPtr, align 4
  call void @runtime.trackPointer(ptr nonnull %defer.alloca, ptr undef) #14
  store i32 0, ptr %defer.alloca, align 8
  %defer.alloca.repack1 = getelementptr inbounds { i32, ptr }, ptr %defer.alloca, i32 0, i32 1
  store ptr null, ptr %defer.alloca.repack1, align 4
  store ptr %defer.alloca, ptr %deferPtr, align 4
  %0 = call i1 @runtime.interfaceEqual(i32 %a.typecode, ptr %a.value, i32 %b.typecode, ptr %b.value, ptr undef) #14
  br label %rundefers.loophead

rundefers.loophead:                               ; preds = %rundefers.callback0, %entry
  %1 = load ptr, ptr %deferPtr, align 4
  %stackIsNil = icmp eq ptr %1, null
  br i1 %stackIsNil, label %rundefers.end, label %rundefers.loop

rundefers.loop:                                   ; preds = %rundefers.loophead
  %stack.next.gep = getelementptr inbounds %runtime._defer, ptr %1, i32 0, i32 1
  %stack.next = load ptr, ptr %stack.next.gep, align 4
  store ptr %stack.next, ptr %deferPtr, align 4
  %callback = load i32, ptr %1, align 4
  switch i32 %callback, label %rundefers.default [
    i32 0, label %rundefers.callback0
  ]

rundefers.callback0:                              ; preds = %rundefers.loop
  call void @"main.interfaceEqual$1"(ptr undef)
  br label %rundefers.loophead

rundefers.default:                                ; preds = %rundefers.loop
  unreachable

rundefers.end:                                    ; preds = %rundefers.loophead
  ret i1 %0

recover:                                          ; No predecessors!
  ret i1 false
}

; Function Attrs: nounwind
define hidden { ptr, %runtime._interface } @"(*main.Cmd).writerDescriptor"(ptr dereferenceable_or_null(168) %c, i32 %w.typecode, ptr %w.value, ptr %context) unnamed_addr #1 {
entry:
  %w = call ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 133 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %w, ptr undef) #14
  store i32 %w.typecode, ptr %w, align 4
  %w.repack66 = getelementptr inbounds %runtime._interface, ptr %w, i32 0, i32 1
  store ptr %w.value, ptr %w.repack66, align 4
  call void @runtime.trackPointer(ptr %w.value, ptr undef) #14
  %0 = icmp eq i32 %w.typecode, 0
  br i1 %0, label %if.then, label %if.done2

if.then:                                          ; preds = %entry
  %1 = call { ptr, %runtime._interface } @os.OpenFile(ptr nonnull @"main$string.74", i32 9, i32 1, i32 0, ptr undef) #14
  %2 = extractvalue { ptr, %runtime._interface } %1, 0
  call void @runtime.trackPointer(ptr %2, ptr undef) #14
  %3 = extractvalue { ptr, %runtime._interface } %1, 1
  %4 = extractvalue %runtime._interface %3, 1
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  %5 = extractvalue { ptr, %runtime._interface } %1, 0
  %6 = extractvalue { ptr, %runtime._interface } %1, 1
  %7 = extractvalue %runtime._interface %6, 0
  %.not105 = icmp eq i32 %7, 0
  br i1 %.not105, label %if.done, label %if.then1

if.then1:                                         ; preds = %if.then
  ret { ptr, %runtime._interface } %1

if.done:                                          ; preds = %if.then
  %8 = icmp eq ptr %c, null
  br i1 %8, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %if.done
  br i1 false, label %gep.throw7, label %gep.next8

gep.next8:                                        ; preds = %gep.next
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next8
  %9 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14
  %.unpack106 = load ptr, ptr %9, align 4
  %.elt107 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 1
  %.unpack108 = load i32, ptr %.elt107, align 4
  %.elt109 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 2
  %.unpack110 = load i32, ptr %.elt109, align 4
  call void @runtime.trackPointer(ptr %.unpack106, ptr undef) #14
  %varargs = call ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 133 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %varargs, ptr undef) #14
  call void @runtime.trackPointer(ptr %5, ptr undef) #14
  store i32 ptrtoint (ptr @"reflect/types.type:pointer:named:os.File" to i32), ptr %varargs, align 4
  %varargs.repack111 = getelementptr inbounds %runtime._interface, ptr %varargs, i32 0, i32 1
  store ptr %5, ptr %varargs.repack111, align 4
  %append.new = call { ptr, i32, i32 } @runtime.sliceAppend(ptr %.unpack106, ptr nonnull %varargs, i32 %.unpack108, i32 %.unpack110, i32 1, i32 8, ptr undef) #14
  %append.newPtr = extractvalue { ptr, i32, i32 } %append.new, 0
  call void @runtime.trackPointer(ptr %append.newPtr, ptr undef) #14
  br i1 false, label %store.throw, label %store.next

store.next:                                       ; preds = %deref.next
  %10 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14
  %append.newLen = extractvalue { ptr, i32, i32 } %append.new, 1
  %append.newCap = extractvalue { ptr, i32, i32 } %append.new, 2
  store ptr %append.newPtr, ptr %10, align 4
  %.repack113 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 1
  store i32 %append.newLen, ptr %.repack113, align 4
  %.repack115 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 2
  store i32 %append.newCap, ptr %.repack115, align 4
  ret { ptr, %runtime._interface } %1

if.done2:                                         ; preds = %entry
  %.unpack = load i32, ptr %w, align 4
  %.elt70 = getelementptr inbounds %runtime._interface, ptr %w, i32 0, i32 1
  %.unpack71 = load ptr, ptr %.elt70, align 4
  call void @runtime.trackPointer(ptr %.unpack71, ptr undef) #14
  %typecode = call i1 @runtime.typeAssert(i32 %.unpack, ptr nonnull @"reflect/types.typeid:pointer:named:os.File", ptr undef) #14
  br i1 %typecode, label %typeassert.ok, label %typeassert.next

typeassert.next:                                  ; preds = %typeassert.ok, %if.done2
  %typeassert.value = phi ptr [ null, %if.done2 ], [ %.unpack71, %typeassert.ok ]
  br i1 %typecode, label %if.then3, label %if.done4

typeassert.ok:                                    ; preds = %if.done2
  br label %typeassert.next

if.then3:                                         ; preds = %typeassert.next
  %11 = insertvalue { ptr, %runtime._interface } zeroinitializer, ptr %typeassert.value, 0
  %12 = insertvalue { ptr, %runtime._interface } %11, %runtime._interface zeroinitializer, 1
  ret { ptr, %runtime._interface } %12

if.done4:                                         ; preds = %typeassert.next
  %pr = call ptr @runtime.alloc(i32 4, ptr nonnull inttoptr (i32 67 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %pr, ptr undef) #14
  %13 = call { ptr, ptr, %runtime._interface } @os.Pipe(ptr undef) #14
  %14 = extractvalue { ptr, ptr, %runtime._interface } %13, 0
  call void @runtime.trackPointer(ptr %14, ptr undef) #14
  %15 = extractvalue { ptr, ptr, %runtime._interface } %13, 1
  call void @runtime.trackPointer(ptr %15, ptr undef) #14
  %16 = extractvalue { ptr, ptr, %runtime._interface } %13, 2
  %17 = extractvalue %runtime._interface %16, 1
  call void @runtime.trackPointer(ptr %17, ptr undef) #14
  %18 = extractvalue { ptr, ptr, %runtime._interface } %13, 0
  store ptr %18, ptr %pr, align 4
  %19 = extractvalue { ptr, ptr, %runtime._interface } %13, 1
  %20 = extractvalue { ptr, ptr, %runtime._interface } %13, 2
  %21 = extractvalue %runtime._interface %20, 0
  %.not = icmp eq i32 %21, 0
  br i1 %.not, label %if.done6, label %if.then5

if.then5:                                         ; preds = %if.done4
  %22 = insertvalue { ptr, %runtime._interface } zeroinitializer, %runtime._interface %20, 1
  ret { ptr, %runtime._interface } %22

if.done6:                                         ; preds = %if.done4
  %23 = icmp eq ptr %c, null
  br i1 %23, label %gep.throw9, label %gep.next10

gep.next10:                                       ; preds = %if.done6
  %24 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14
  br i1 false, label %gep.throw11, label %gep.next12

gep.next12:                                       ; preds = %gep.next10
  br i1 false, label %deref.throw13, label %deref.next14

deref.next14:                                     ; preds = %gep.next12
  %25 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14
  %.unpack72 = load ptr, ptr %25, align 4
  %.elt73 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 1
  %.unpack74 = load i32, ptr %.elt73, align 4
  %.elt75 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 2
  %.unpack76 = load i32, ptr %.elt75, align 4
  call void @runtime.trackPointer(ptr %.unpack72, ptr undef) #14
  %varargs15 = call ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 133 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %varargs15, ptr undef) #14
  call void @runtime.trackPointer(ptr %19, ptr undef) #14
  store i32 ptrtoint (ptr @"reflect/types.type:pointer:named:os.File" to i32), ptr %varargs15, align 4
  %varargs15.repack77 = getelementptr inbounds %runtime._interface, ptr %varargs15, i32 0, i32 1
  store ptr %19, ptr %varargs15.repack77, align 4
  %append.new22 = call { ptr, i32, i32 } @runtime.sliceAppend(ptr %.unpack72, ptr nonnull %varargs15, i32 %.unpack74, i32 %.unpack76, i32 1, i32 8, ptr undef) #14
  %append.newPtr23 = extractvalue { ptr, i32, i32 } %append.new22, 0
  call void @runtime.trackPointer(ptr %append.newPtr23, ptr undef) #14
  br i1 false, label %store.throw26, label %store.next27

store.next27:                                     ; preds = %deref.next14
  %append.newLen24 = extractvalue { ptr, i32, i32 } %append.new22, 1
  %append.newCap25 = extractvalue { ptr, i32, i32 } %append.new22, 2
  store ptr %append.newPtr23, ptr %24, align 4
  %.repack79 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 1
  store i32 %append.newLen24, ptr %.repack79, align 4
  %.repack81 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 14, i32 2
  store i32 %append.newCap25, ptr %.repack81, align 4
  br i1 false, label %gep.throw28, label %gep.next29

gep.next29:                                       ; preds = %store.next27
  %26 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15
  br i1 false, label %gep.throw30, label %gep.next31

gep.next31:                                       ; preds = %gep.next29
  br i1 false, label %deref.throw32, label %deref.next33

deref.next33:                                     ; preds = %gep.next31
  %27 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15
  %.unpack83 = load ptr, ptr %27, align 4
  %.elt84 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 1
  %.unpack85 = load i32, ptr %.elt84, align 4
  %.elt86 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 2
  %.unpack87 = load i32, ptr %.elt86, align 4
  call void @runtime.trackPointer(ptr %.unpack83, ptr undef) #14
  %28 = load ptr, ptr %pr, align 4
  call void @runtime.trackPointer(ptr %28, ptr undef) #14
  %varargs34 = call ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 133 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %varargs34, ptr undef) #14
  call void @runtime.trackPointer(ptr %28, ptr undef) #14
  store i32 ptrtoint (ptr @"reflect/types.type:pointer:named:os.File" to i32), ptr %varargs34, align 4
  %varargs34.repack88 = getelementptr inbounds %runtime._interface, ptr %varargs34, i32 0, i32 1
  store ptr %28, ptr %varargs34.repack88, align 4
  %append.new41 = call { ptr, i32, i32 } @runtime.sliceAppend(ptr %.unpack83, ptr nonnull %varargs34, i32 %.unpack85, i32 %.unpack87, i32 1, i32 8, ptr undef) #14
  %append.newPtr42 = extractvalue { ptr, i32, i32 } %append.new41, 0
  call void @runtime.trackPointer(ptr %append.newPtr42, ptr undef) #14
  br i1 false, label %store.throw45, label %store.next46

store.next46:                                     ; preds = %deref.next33
  %append.newLen43 = extractvalue { ptr, i32, i32 } %append.new41, 1
  %append.newCap44 = extractvalue { ptr, i32, i32 } %append.new41, 2
  store ptr %append.newPtr42, ptr %26, align 4
  %.repack90 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 1
  store i32 %append.newLen43, ptr %.repack90, align 4
  %.repack92 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 15, i32 2
  store i32 %append.newCap44, ptr %.repack92, align 4
  br i1 false, label %gep.throw47, label %gep.next48

gep.next48:                                       ; preds = %store.next46
  br i1 false, label %gep.throw49, label %gep.next50

gep.next50:                                       ; preds = %gep.next48
  br i1 false, label %deref.throw51, label %deref.next52

deref.next52:                                     ; preds = %gep.next50
  %29 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 16
  %.unpack94 = load ptr, ptr %29, align 4
  %.elt95 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 16, i32 1
  %.unpack96 = load i32, ptr %.elt95, align 4
  %.elt97 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 16, i32 2
  %.unpack98 = load i32, ptr %.elt97, align 4
  call void @runtime.trackPointer(ptr %.unpack94, ptr undef) #14
  %30 = call ptr @runtime.alloc(i32 8, ptr null, ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %30, ptr undef) #14
  store ptr %w, ptr %30, align 4
  %31 = getelementptr inbounds { ptr, ptr }, ptr %30, i32 0, i32 1
  store ptr %pr, ptr %31, align 4
  call void @runtime.trackPointer(ptr nonnull %30, ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull @"(*main.Cmd).writerDescriptor$1", ptr undef) #14
  %varargs53 = call ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 197 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %varargs53, ptr undef) #14
  store ptr %30, ptr %varargs53, align 4
  %varargs53.repack99 = getelementptr inbounds { ptr, ptr }, ptr %varargs53, i32 0, i32 1
  store ptr @"(*main.Cmd).writerDescriptor$1", ptr %varargs53.repack99, align 4
  %append.new60 = call { ptr, i32, i32 } @runtime.sliceAppend(ptr %.unpack94, ptr nonnull %varargs53, i32 %.unpack96, i32 %.unpack98, i32 1, i32 8, ptr undef) #14
  %append.newPtr61 = extractvalue { ptr, i32, i32 } %append.new60, 0
  call void @runtime.trackPointer(ptr %append.newPtr61, ptr undef) #14
  br i1 false, label %store.throw64, label %store.next65

store.next65:                                     ; preds = %deref.next52
  %32 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 16
  %append.newLen62 = extractvalue { ptr, i32, i32 } %append.new60, 1
  %append.newCap63 = extractvalue { ptr, i32, i32 } %append.new60, 2
  store ptr %append.newPtr61, ptr %32, align 4
  %.repack101 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 16, i32 1
  store i32 %append.newLen62, ptr %.repack101, align 4
  %.repack103 = getelementptr inbounds %main.Cmd, ptr %c, i32 0, i32 16, i32 2
  store i32 %append.newCap63, ptr %.repack103, align 4
  %33 = insertvalue { ptr, %runtime._interface } zeroinitializer, ptr %19, 0
  %34 = insertvalue { ptr, %runtime._interface } %33, %runtime._interface zeroinitializer, 1
  ret { ptr, %runtime._interface } %34

gep.throw:                                        ; preds = %if.done
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

gep.throw7:                                       ; preds = %gep.next
  unreachable

deref.throw:                                      ; preds = %gep.next8
  unreachable

store.throw:                                      ; preds = %deref.next
  unreachable

gep.throw9:                                       ; preds = %if.done6
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

gep.throw11:                                      ; preds = %gep.next10
  unreachable

deref.throw13:                                    ; preds = %gep.next12
  unreachable

store.throw26:                                    ; preds = %deref.next14
  unreachable

gep.throw28:                                      ; preds = %store.next27
  unreachable

gep.throw30:                                      ; preds = %gep.next29
  unreachable

deref.throw32:                                    ; preds = %gep.next31
  unreachable

store.throw45:                                    ; preds = %deref.next33
  unreachable

gep.throw47:                                      ; preds = %store.next46
  unreachable

gep.throw49:                                      ; preds = %gep.next48
  unreachable

deref.throw51:                                    ; preds = %gep.next50
  unreachable

store.throw64:                                    ; preds = %deref.next52
  unreachable
}

declare { ptr, %runtime._interface } @os.Open(ptr, i32, ptr) #0

; Function Attrs: nounwind
define internal %runtime._interface @"(*main.Cmd).stdin$1"(ptr %context) unnamed_addr #1 {
entry:
  %0 = load ptr, ptr %context, align 4
  %1 = getelementptr inbounds { ptr, ptr }, ptr %context, i32 0, i32 1
  %2 = load ptr, ptr %1, align 4
  %3 = load ptr, ptr %0, align 4
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %4 = load ptr, ptr %2, align 4
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  %5 = icmp eq ptr %4, null
  br i1 %5, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %6 = getelementptr inbounds %main.Cmd, ptr %4, i32 0, i32 4
  %.unpack = load i32, ptr %6, align 4
  %.elt3 = getelementptr inbounds %main.Cmd, ptr %4, i32 0, i32 4, i32 1
  %.unpack4 = load ptr, ptr %.elt3, align 4
  call void @runtime.trackPointer(ptr %.unpack4, ptr undef) #14
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %7 = call { i64, %runtime._interface } @io.Copy(i32 ptrtoint (ptr @"reflect/types.type:pointer:named:os.File" to i32), ptr %3, i32 %.unpack, ptr %.unpack4, ptr undef) #14
  %8 = extractvalue { i64, %runtime._interface } %7, 1
  %9 = extractvalue %runtime._interface %8, 1
  call void @runtime.trackPointer(ptr %9, ptr undef) #14
  %10 = extractvalue { i64, %runtime._interface } %7, 1
  %11 = extractvalue %runtime._interface %10, 0
  %12 = extractvalue %runtime._interface %10, 1
  %13 = call i1 @main.skipStdinCopyError(i32 %11, ptr %12, ptr undef)
  br i1 %13, label %if.then, label %if.done

if.then:                                          ; preds = %deref.next
  br label %if.done

if.done:                                          ; preds = %if.then, %deref.next
  %14 = phi %runtime._interface [ %10, %deref.next ], [ zeroinitializer, %if.then ]
  %15 = extractvalue %runtime._interface %14, 1
  call void @runtime.trackPointer(ptr %15, ptr undef) #14
  %16 = load ptr, ptr %0, align 4
  call void @runtime.trackPointer(ptr %16, ptr undef) #14
  %17 = call %runtime._interface @"(*os.File).Close"(ptr %16, ptr undef) #14
  %18 = extractvalue %runtime._interface %17, 1
  call void @runtime.trackPointer(ptr %18, ptr undef) #14
  %19 = extractvalue %runtime._interface %14, 0
  %20 = icmp eq i32 %19, 0
  br i1 %20, label %if.then1, label %if.done2

if.then1:                                         ; preds = %if.done
  br label %if.done2

if.done2:                                         ; preds = %if.then1, %if.done
  %21 = phi %runtime._interface [ %14, %if.done ], [ %17, %if.then1 ]
  %22 = extractvalue %runtime._interface %21, 1
  call void @runtime.trackPointer(ptr %22, ptr undef) #14
  ret %runtime._interface %21

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

declare { i64, %runtime._interface } @io.Copy(i32, ptr, i32, ptr, ptr) #0

; Function Attrs: nounwind
define internal void @"(*main.Cmd).watchCtx$1"(ptr %context) unnamed_addr #1 {
entry:
  %complit = alloca %main.wrappedError, align 8
  %chan.blockedList = alloca %runtime.channelBlockedList, align 8
  %chan.value = alloca %runtime._interface, align 8
  %select.block.alloca = alloca [2 x %runtime.channelBlockedList], align 8
  %select.states.alloca = alloca [2 x %runtime.chanSelectState], align 8
  %select.send.value = alloca %runtime._interface, align 8
  %0 = load ptr, ptr %context, align 4
  %1 = getelementptr inbounds { ptr, ptr }, ptr %context, i32 0, i32 1
  %2 = load ptr, ptr %1, align 4
  %3 = load ptr, ptr %0, align 4
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %4 = load ptr, ptr %2, align 4
  call void @runtime.trackPointer(ptr %4, ptr undef) #14
  %5 = icmp eq ptr %4, null
  br i1 %5, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %6 = getelementptr inbounds %main.Cmd, ptr %4, i32 0, i32 11
  %.unpack = load i32, ptr %6, align 4
  %.elt16 = getelementptr inbounds %main.Cmd, ptr %4, i32 0, i32 11, i32 1
  %.unpack17 = load ptr, ptr %.elt16, align 4
  call void @runtime.trackPointer(ptr %.unpack17, ptr undef) #14
  %7 = call ptr @"interface:{Deadline:func:{}{named:time.Time,basic:bool},Done:func:{}{chan:struct:{}},Err:func:{}{named:error},Value:func:{interface:{}}{interface:{}}}.Done$invoke"(ptr %.unpack17, i32 %.unpack, ptr undef) #14
  call void @runtime.trackPointer(ptr %7, ptr undef) #14
  store i32 0, ptr %select.send.value, align 8
  %select.send.value.repack18 = getelementptr inbounds %runtime._interface, ptr %select.send.value, i32 0, i32 1
  store ptr null, ptr %select.send.value.repack18, align 4
  call void @llvm.lifetime.start.p0(i64 16, ptr nonnull %select.states.alloca)
  store ptr %3, ptr %select.states.alloca, align 8
  %select.states.alloca.repack19 = getelementptr inbounds %runtime.chanSelectState, ptr %select.states.alloca, i32 0, i32 1
  store ptr %select.send.value, ptr %select.states.alloca.repack19, align 4
  %8 = getelementptr inbounds [2 x %runtime.chanSelectState], ptr %select.states.alloca, i32 0, i32 1
  store ptr %7, ptr %8, align 8
  %.repack21 = getelementptr inbounds [2 x %runtime.chanSelectState], ptr %select.states.alloca, i32 0, i32 1, i32 1
  store ptr null, ptr %.repack21, align 4
  call void @llvm.lifetime.start.p0(i64 48, ptr nonnull %select.block.alloca)
  %select.result = call { i32, i1 } @runtime.chanSelect(ptr undef, ptr nonnull %select.states.alloca, i32 2, i32 2, ptr nonnull %select.block.alloca, i32 2, i32 2, ptr undef) #14
  call void @llvm.lifetime.end.p0(i64 48, ptr nonnull %select.block.alloca)
  call void @llvm.lifetime.end.p0(i64 16, ptr nonnull %select.states.alloca)
  %9 = extractvalue { i32, i1 } %select.result, 0
  %10 = icmp eq i32 %9, 0
  br i1 %10, label %select.body, label %select.next

select.body:                                      ; preds = %deref.next
  ret void

select.next:                                      ; preds = %deref.next
  %11 = icmp eq i32 %9, 1
  br i1 %11, label %select.body1, label %select.next3

select.body1:                                     ; preds = %select.next
  %12 = load ptr, ptr %2, align 4
  call void @runtime.trackPointer(ptr %12, ptr undef) #14
  %13 = icmp eq ptr %12, null
  br i1 %13, label %gep.throw4, label %gep.next5

gep.next5:                                        ; preds = %select.body1
  br i1 false, label %deref.throw6, label %deref.next7

deref.next7:                                      ; preds = %gep.next5
  %14 = getelementptr inbounds %main.Cmd, ptr %12, i32 0, i32 9
  %15 = load ptr, ptr %14, align 4
  call void @runtime.trackPointer(ptr %15, ptr undef) #14
  %16 = call %runtime._interface @"(*os.Process).Kill"(ptr %15, ptr undef) #14
  %17 = extractvalue %runtime._interface %16, 1
  call void @runtime.trackPointer(ptr %17, ptr undef) #14
  %18 = extractvalue %runtime._interface %16, 0
  %19 = icmp eq i32 %18, 0
  br i1 %19, label %if.then, label %if.else

if.then:                                          ; preds = %deref.next7
  %20 = load ptr, ptr %2, align 4
  call void @runtime.trackPointer(ptr %20, ptr undef) #14
  %21 = icmp eq ptr %20, null
  br i1 %21, label %gep.throw8, label %gep.next9

gep.next9:                                        ; preds = %if.then
  br i1 false, label %deref.throw10, label %deref.next11

deref.next11:                                     ; preds = %gep.next9
  %22 = getelementptr inbounds %main.Cmd, ptr %20, i32 0, i32 11
  %.unpack52 = load i32, ptr %22, align 4
  %.elt53 = getelementptr inbounds %main.Cmd, ptr %20, i32 0, i32 11, i32 1
  %.unpack54 = load ptr, ptr %.elt53, align 4
  call void @runtime.trackPointer(ptr %.unpack54, ptr undef) #14
  %23 = call %runtime._interface @"interface:{Deadline:func:{}{named:time.Time,basic:bool},Done:func:{}{chan:struct:{}},Err:func:{}{named:error},Value:func:{interface:{}}{interface:{}}}.Err$invoke"(ptr %.unpack54, i32 %.unpack52, ptr undef) #14
  %24 = extractvalue %runtime._interface %23, 1
  call void @runtime.trackPointer(ptr %24, ptr undef) #14
  br label %if.done

if.done:                                          ; preds = %store.next15, %if.else, %deref.next11
  %25 = phi %runtime._interface [ %23, %deref.next11 ], [ zeroinitializer, %if.else ], [ %33, %store.next15 ]
  %26 = extractvalue %runtime._interface %25, 1
  call void @runtime.trackPointer(ptr %26, ptr undef) #14
  %27 = load ptr, ptr %0, align 4
  call void @runtime.trackPointer(ptr %27, ptr undef) #14
  call void @llvm.lifetime.start.p0(i64 8, ptr nonnull %chan.value)
  %.elt48 = extractvalue %runtime._interface %25, 0
  store i32 %.elt48, ptr %chan.value, align 8
  %chan.value.repack49 = getelementptr inbounds %runtime._interface, ptr %chan.value, i32 0, i32 1
  %.elt50 = extractvalue %runtime._interface %25, 1
  store ptr %.elt50, ptr %chan.value.repack49, align 4
  call void @llvm.lifetime.start.p0(i64 24, ptr nonnull %chan.blockedList)
  call void @runtime.chanSend(ptr %27, ptr nonnull %chan.value, ptr nonnull %chan.blockedList, ptr undef) #14
  call void @llvm.lifetime.end.p0(i64 24, ptr nonnull %chan.blockedList)
  call void @llvm.lifetime.end.p0(i64 8, ptr nonnull %chan.value)
  ret void

if.else:                                          ; preds = %deref.next7
  %.unpack23 = load i32, ptr @os.ErrProcessDone, align 4
  %.unpack24 = load ptr, ptr getelementptr inbounds (%runtime._interface, ptr @os.ErrProcessDone, i32 0, i32 1), align 4
  call void @runtime.trackPointer(ptr %.unpack24, ptr undef) #14
  %28 = extractvalue %runtime._interface %16, 0
  %29 = extractvalue %runtime._interface %16, 1
  %30 = call i1 @errors.Is(i32 %28, ptr %29, i32 %.unpack23, ptr %.unpack24, ptr undef) #14
  br i1 %30, label %if.done, label %if.then2

if.then2:                                         ; preds = %if.else
  store ptr null, ptr %complit, align 8
  %complit.repack26 = getelementptr inbounds %runtime._string, ptr %complit, i32 0, i32 1
  store i32 0, ptr %complit.repack26, align 4
  %complit.repack25 = getelementptr inbounds %main.wrappedError, ptr %complit, i32 0, i32 1
  store i32 0, ptr %complit.repack25, align 8
  %complit.repack25.repack27 = getelementptr inbounds %main.wrappedError, ptr %complit, i32 0, i32 1, i32 1
  store ptr null, ptr %complit.repack25.repack27, align 4
  call void @runtime.trackPointer(ptr nonnull %complit, ptr undef) #14
  %31 = getelementptr inbounds %main.wrappedError, ptr %complit, i32 0, i32 1
  br i1 false, label %store.throw, label %store.next

store.next:                                       ; preds = %if.then2
  store ptr @"main$string.65", ptr %complit, align 8
  %complit.repack28 = getelementptr inbounds %runtime._string, ptr %complit, i32 0, i32 1
  store i32 33, ptr %complit.repack28, align 4
  br i1 false, label %store.throw14, label %store.next15

store.next15:                                     ; preds = %store.next
  %.elt = extractvalue %runtime._interface %16, 0
  store i32 %.elt, ptr %31, align 8
  %.repack29 = getelementptr inbounds %main.wrappedError, ptr %complit, i32 0, i32 1, i32 1
  %.elt30 = extractvalue %runtime._interface %16, 1
  store ptr %.elt30, ptr %.repack29, align 4
  %.unpack32.unpack = load ptr, ptr %complit, align 8
  %.unpack32.elt35 = getelementptr inbounds %runtime._string, ptr %complit, i32 0, i32 1
  %.unpack32.unpack36 = load i32, ptr %.unpack32.elt35, align 4
  %.elt33 = getelementptr inbounds %main.wrappedError, ptr %complit, i32 0, i32 1
  %.unpack34.unpack = load i32, ptr %.elt33, align 8
  %.unpack34.elt38 = getelementptr inbounds %main.wrappedError, ptr %complit, i32 0, i32 1, i32 1
  %.unpack34.unpack39 = load ptr, ptr %.unpack34.elt38, align 4
  call void @runtime.trackPointer(ptr %.unpack32.unpack, ptr undef) #14
  call void @runtime.trackPointer(ptr %.unpack34.unpack39, ptr undef) #14
  %32 = call ptr @runtime.alloc(i32 16, ptr null, ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %32, ptr undef) #14
  store ptr %.unpack32.unpack, ptr %32, align 4
  %.repack44 = getelementptr inbounds %runtime._string, ptr %32, i32 0, i32 1
  store i32 %.unpack32.unpack36, ptr %.repack44, align 4
  %.repack42 = getelementptr inbounds %main.wrappedError, ptr %32, i32 0, i32 1
  store i32 %.unpack34.unpack, ptr %.repack42, align 4
  %.repack42.repack46 = getelementptr inbounds %main.wrappedError, ptr %32, i32 0, i32 1, i32 1
  store ptr %.unpack34.unpack39, ptr %.repack42.repack46, align 4
  %33 = insertvalue %runtime._interface { i32 ptrtoint (ptr @"reflect/types.type:named:main.wrappedError" to i32), ptr undef }, ptr %32, 1
  call void @runtime.trackPointer(ptr nonnull %32, ptr undef) #14
  br label %if.done

select.next3:                                     ; preds = %select.next
  call void @runtime.trackPointer(ptr nonnull @"main$pack", ptr undef) #14
  call void @runtime._panic(i32 ptrtoint (ptr @"reflect/types.type:basic:string" to i32), ptr nonnull @"main$pack", ptr undef) #14
  unreachable

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable

gep.throw4:                                       ; preds = %select.body1
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw6:                                     ; preds = %gep.next5
  unreachable

gep.throw8:                                       ; preds = %if.then
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

deref.throw10:                                    ; preds = %gep.next9
  unreachable

store.throw:                                      ; preds = %if.then2
  unreachable

store.throw14:                                    ; preds = %store.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr void @"(*main.Cmd).watchCtx$1$gowrapper"(ptr %0) unnamed_addr #13 {
entry:
  call void @"(*main.Cmd).watchCtx$1"(ptr %0)
  call void @runtime.deadlock(ptr undef) #14
  unreachable
}

declare { i32, i1 } @runtime.chanSelect(ptr, ptr, i32, i32, ptr, i32, i32, ptr) #0

declare %runtime._interface @"(*os.Process).Kill"(ptr dereferenceable_or_null(4), ptr) #0

declare i1 @errors.Is(i32, ptr, i32, ptr, ptr) #0

; Function Attrs: nounwind
define linkonce_odr %runtime._string @"(main.wrappedError).Error$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
entry:
  %.unpack.unpack = load ptr, ptr %0, align 4
  %2 = insertvalue %runtime._string undef, ptr %.unpack.unpack, 0
  %.unpack.elt3 = getelementptr inbounds %runtime._string, ptr %0, i32 0, i32 1
  %.unpack.unpack4 = load i32, ptr %.unpack.elt3, align 4
  %.unpack5 = insertvalue %runtime._string %2, i32 %.unpack.unpack4, 1
  %3 = insertvalue %main.wrappedError undef, %runtime._string %.unpack5, 0
  %.elt1 = getelementptr inbounds %main.wrappedError, ptr %0, i32 0, i32 1
  %.unpack2.unpack = load i32, ptr %.elt1, align 4
  %4 = insertvalue %runtime._interface undef, i32 %.unpack2.unpack, 0
  %.unpack2.elt6 = getelementptr inbounds %main.wrappedError, ptr %0, i32 0, i32 1, i32 1
  %.unpack2.unpack7 = load ptr, ptr %.unpack2.elt6, align 4
  %.unpack28 = insertvalue %runtime._interface %4, ptr %.unpack2.unpack7, 1
  %5 = insertvalue %main.wrappedError %3, %runtime._interface %.unpack28, 1
  %ret = call %runtime._string @"(main.wrappedError).Error"(%main.wrappedError %5, ptr %1)
  ret %runtime._string %ret
}

; Function Attrs: nounwind
define linkonce_odr %runtime._interface @"(main.wrappedError).Unwrap$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
entry:
  %.unpack.unpack = load ptr, ptr %0, align 4
  %2 = insertvalue %runtime._string undef, ptr %.unpack.unpack, 0
  %.unpack.elt3 = getelementptr inbounds %runtime._string, ptr %0, i32 0, i32 1
  %.unpack.unpack4 = load i32, ptr %.unpack.elt3, align 4
  %.unpack5 = insertvalue %runtime._string %2, i32 %.unpack.unpack4, 1
  %3 = insertvalue %main.wrappedError undef, %runtime._string %.unpack5, 0
  %.elt1 = getelementptr inbounds %main.wrappedError, ptr %0, i32 0, i32 1
  %.unpack2.unpack = load i32, ptr %.elt1, align 4
  %4 = insertvalue %runtime._interface undef, i32 %.unpack2.unpack, 0
  %.unpack2.elt6 = getelementptr inbounds %main.wrappedError, ptr %0, i32 0, i32 1, i32 1
  %.unpack2.unpack7 = load ptr, ptr %.unpack2.elt6, align 4
  %.unpack28 = insertvalue %runtime._interface %4, ptr %.unpack2.unpack7, 1
  %5 = insertvalue %main.wrappedError %3, %runtime._interface %.unpack28, 1
  %ret = call %runtime._interface @"(main.wrappedError).Unwrap"(%main.wrappedError %5, ptr %1)
  ret %runtime._interface %ret
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._string @"(*main.wrappedError).Error$wrapper for func (main.wrappedError).Error() string"(ptr dereferenceable_or_null(16) %w, ptr %context) unnamed_addr #1 {
entry:
  call void @runtime.trackPointer(ptr %w, ptr undef) #14
  %0 = icmp eq ptr %w, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %.unpack.unpack = load ptr, ptr %w, align 4
  %1 = insertvalue %runtime._string undef, ptr %.unpack.unpack, 0
  %.unpack.elt3 = getelementptr inbounds %runtime._string, ptr %w, i32 0, i32 1
  %.unpack.unpack4 = load i32, ptr %.unpack.elt3, align 4
  %.unpack5 = insertvalue %runtime._string %1, i32 %.unpack.unpack4, 1
  %2 = insertvalue %main.wrappedError undef, %runtime._string %.unpack5, 0
  %.elt1 = getelementptr inbounds %main.wrappedError, ptr %w, i32 0, i32 1
  %.unpack2.unpack = load i32, ptr %.elt1, align 4
  %3 = insertvalue %runtime._interface undef, i32 %.unpack2.unpack, 0
  %.unpack2.elt6 = getelementptr inbounds %main.wrappedError, ptr %w, i32 0, i32 1, i32 1
  %.unpack2.unpack7 = load ptr, ptr %.unpack2.elt6, align 4
  %.unpack28 = insertvalue %runtime._interface %3, ptr %.unpack2.unpack7, 1
  %4 = insertvalue %main.wrappedError %2, %runtime._interface %.unpack28, 1
  call void @runtime.trackPointer(ptr %.unpack.unpack, ptr undef) #14
  call void @runtime.trackPointer(ptr %.unpack2.unpack7, ptr undef) #14
  %5 = call %runtime._string @"(main.wrappedError).Error"(%main.wrappedError %4, ptr undef)
  %6 = extractvalue %runtime._string %5, 0
  call void @runtime.trackPointer(ptr %6, ptr undef) #14
  ret %runtime._string %5

deref.throw:                                      ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(*main.wrappedError).Unwrap$wrapper for func (main.wrappedError).Unwrap() error"(ptr dereferenceable_or_null(16) %w, ptr %context) unnamed_addr #1 {
entry:
  call void @runtime.trackPointer(ptr %w, ptr undef) #14
  %0 = icmp eq ptr %w, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %.unpack.unpack = load ptr, ptr %w, align 4
  %1 = insertvalue %runtime._string undef, ptr %.unpack.unpack, 0
  %.unpack.elt3 = getelementptr inbounds %runtime._string, ptr %w, i32 0, i32 1
  %.unpack.unpack4 = load i32, ptr %.unpack.elt3, align 4
  %.unpack5 = insertvalue %runtime._string %1, i32 %.unpack.unpack4, 1
  %2 = insertvalue %main.wrappedError undef, %runtime._string %.unpack5, 0
  %.elt1 = getelementptr inbounds %main.wrappedError, ptr %w, i32 0, i32 1
  %.unpack2.unpack = load i32, ptr %.elt1, align 4
  %3 = insertvalue %runtime._interface undef, i32 %.unpack2.unpack, 0
  %.unpack2.elt6 = getelementptr inbounds %main.wrappedError, ptr %w, i32 0, i32 1, i32 1
  %.unpack2.unpack7 = load ptr, ptr %.unpack2.elt6, align 4
  %.unpack28 = insertvalue %runtime._interface %3, ptr %.unpack2.unpack7, 1
  %4 = insertvalue %main.wrappedError %2, %runtime._interface %.unpack28, 1
  call void @runtime.trackPointer(ptr %.unpack.unpack, ptr undef) #14
  call void @runtime.trackPointer(ptr %.unpack2.unpack7, ptr undef) #14
  %5 = call %runtime._interface @"(main.wrappedError).Unwrap"(%main.wrappedError %4, ptr undef)
  %6 = extractvalue %runtime._interface %5, 1
  call void @runtime.trackPointer(ptr %6, ptr undef) #14
  ret %runtime._interface %5

deref.throw:                                      ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #14
  unreachable
}

declare void @runtime._panic(i32, ptr, ptr) #0

declare { ptr, %runtime._interface } @os.OpenFile(ptr, i32, i32, i32, ptr) #0

; Function Attrs: nounwind
define internal %runtime._interface @"(*main.Cmd).writerDescriptor$1"(ptr %context) unnamed_addr #1 {
entry:
  %0 = load ptr, ptr %context, align 4
  %1 = getelementptr inbounds { ptr, ptr }, ptr %context, i32 0, i32 1
  %2 = load ptr, ptr %1, align 4
  %.unpack = load i32, ptr %0, align 4
  %.elt1 = getelementptr inbounds %runtime._interface, ptr %0, i32 0, i32 1
  %.unpack2 = load ptr, ptr %.elt1, align 4
  call void @runtime.trackPointer(ptr %.unpack2, ptr undef) #14
  %3 = load ptr, ptr %2, align 4
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  %4 = call { i64, %runtime._interface } @io.Copy(i32 %.unpack, ptr %.unpack2, i32 ptrtoint (ptr @"reflect/types.type:pointer:named:os.File" to i32), ptr %3, ptr undef) #14
  %5 = extractvalue { i64, %runtime._interface } %4, 1
  %6 = extractvalue %runtime._interface %5, 1
  call void @runtime.trackPointer(ptr %6, ptr undef) #14
  %7 = extractvalue { i64, %runtime._interface } %4, 1
  %8 = load ptr, ptr %2, align 4
  call void @runtime.trackPointer(ptr %8, ptr undef) #14
  %9 = call %runtime._interface @"(*os.File).Close"(ptr %8, ptr undef) #14
  %10 = extractvalue %runtime._interface %9, 1
  call void @runtime.trackPointer(ptr %10, ptr undef) #14
  ret %runtime._interface %7
}

; Function Attrs: nounwind
define hidden ptr @main.Command(ptr %name.data, i32 %name.len, ptr %arg.data, i32 %arg.len, i32 %arg.cap, ptr %context) unnamed_addr #1 {
entry:
  %complit = call ptr @runtime.alloc(i32 168, ptr nonnull @"runtime/gc.layout:42-2c926b9a925", ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %complit, ptr undef) #14
  %0 = getelementptr inbounds %main.Cmd, ptr %complit, i32 0, i32 1
  %slicelit = call ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 69 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %slicelit, ptr undef) #14
  store ptr %name.data, ptr %slicelit, align 4
  %slicelit.repack10 = getelementptr inbounds %runtime._string, ptr %slicelit, i32 0, i32 1
  store i32 %name.len, ptr %slicelit.repack10, align 4
  %append.new = call { ptr, i32, i32 } @runtime.sliceAppend(ptr nonnull %slicelit, ptr %arg.data, i32 1, i32 1, i32 %arg.len, i32 8, ptr undef) #14
  %append.newPtr = extractvalue { ptr, i32, i32 } %append.new, 0
  %append.newLen = extractvalue { ptr, i32, i32 } %append.new, 1
  %append.newCap = extractvalue { ptr, i32, i32 } %append.new, 2
  call void @runtime.trackPointer(ptr %append.newPtr, ptr undef) #14
  br i1 false, label %store.throw, label %store.next

store.next:                                       ; preds = %entry
  store ptr %name.data, ptr %complit, align 4
  %complit.repack12 = getelementptr inbounds %runtime._string, ptr %complit, i32 0, i32 1
  store i32 %name.len, ptr %complit.repack12, align 4
  br i1 false, label %store.throw4, label %store.next5

store.next5:                                      ; preds = %store.next
  store ptr %append.newPtr, ptr %0, align 4
  %.repack14 = getelementptr inbounds %main.Cmd, ptr %complit, i32 0, i32 1, i32 1
  store i32 %append.newLen, ptr %.repack14, align 4
  %.repack16 = getelementptr inbounds %main.Cmd, ptr %complit, i32 0, i32 1, i32 2
  store i32 %append.newCap, ptr %.repack16, align 4
  %1 = call %runtime._string @"path/filepath.Base"(ptr %name.data, i32 %name.len, ptr undef) #14
  %2 = extractvalue %runtime._string %1, 0
  call void @runtime.trackPointer(ptr %2, ptr undef) #14
  %3 = extractvalue %runtime._string %1, 0
  %4 = extractvalue %runtime._string %1, 1
  %5 = call i1 @runtime.stringEqual(ptr %3, i32 %4, ptr %name.data, i32 %name.len, ptr undef) #14
  br i1 %5, label %if.then, label %if.done3

if.then:                                          ; preds = %store.next5
  %6 = call { %runtime._string, %runtime._interface } @main.LookPath(ptr %name.data, i32 %name.len, ptr undef)
  %7 = extractvalue { %runtime._string, %runtime._interface } %6, 0
  %8 = extractvalue %runtime._string %7, 0
  call void @runtime.trackPointer(ptr %8, ptr undef) #14
  %9 = extractvalue { %runtime._string, %runtime._interface } %6, 1
  %10 = extractvalue %runtime._interface %9, 1
  call void @runtime.trackPointer(ptr %10, ptr undef) #14
  %11 = extractvalue { %runtime._string, %runtime._interface } %6, 0
  %12 = extractvalue { %runtime._string, %runtime._interface } %6, 1
  %13 = extractvalue %runtime._string %11, 0
  %14 = extractvalue %runtime._string %11, 1
  %15 = call i1 @runtime.stringEqual(ptr %13, i32 %14, ptr null, i32 0, ptr undef) #14
  br i1 %15, label %if.done, label %if.then1

if.then1:                                         ; preds = %if.then
  br i1 false, label %store.throw6, label %store.next7

store.next7:                                      ; preds = %if.then1
  %.elt20 = extractvalue %runtime._string %11, 0
  store ptr %.elt20, ptr %complit, align 4
  %complit.repack21 = getelementptr inbounds %runtime._string, ptr %complit, i32 0, i32 1
  %.elt22 = extractvalue %runtime._string %11, 1
  store i32 %.elt22, ptr %complit.repack21, align 4
  br label %if.done

if.done:                                          ; preds = %store.next7, %if.then
  %16 = extractvalue %runtime._interface %12, 0
  %.not = icmp eq i32 %16, 0
  br i1 %.not, label %if.done3, label %if.then2

if.then2:                                         ; preds = %if.done
  br i1 false, label %store.throw8, label %store.next9

store.next9:                                      ; preds = %if.then2
  %17 = getelementptr inbounds %main.Cmd, ptr %complit, i32 0, i32 12
  %.elt = extractvalue %runtime._interface %12, 0
  store i32 %.elt, ptr %17, align 4
  %.repack18 = getelementptr inbounds %main.Cmd, ptr %complit, i32 0, i32 12, i32 1
  %.elt19 = extractvalue %runtime._interface %12, 1
  store ptr %.elt19, ptr %.repack18, align 4
  br label %if.done3

if.done3:                                         ; preds = %store.next9, %if.done, %store.next5
  ret ptr %complit

store.throw:                                      ; preds = %entry
  unreachable

store.throw4:                                     ; preds = %store.next
  unreachable

store.throw6:                                     ; preds = %if.then1
  unreachable

store.throw8:                                     ; preds = %if.then2
  unreachable
}

declare %runtime._string @"path/filepath.Base"(ptr, i32, ptr) #0

; Function Attrs: nounwind
define hidden ptr @main.CommandContext(i32 %ctx.typecode, ptr %ctx.value, ptr %name.data, i32 %name.len, ptr %arg.data, i32 %arg.len, i32 %arg.cap, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq i32 %ctx.typecode, 0
  br i1 %0, label %if.then, label %if.done

if.then:                                          ; preds = %entry
  call void @runtime.trackPointer(ptr nonnull @"main$pack.76", ptr undef) #14
  call void @runtime._panic(i32 ptrtoint (ptr @"reflect/types.type:basic:string" to i32), ptr nonnull @"main$pack.76", ptr undef) #14
  unreachable

if.done:                                          ; preds = %entry
  %1 = call ptr @main.Command(ptr %name.data, i32 %name.len, ptr %arg.data, i32 %arg.len, i32 %arg.cap, ptr undef)
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  %2 = icmp eq ptr %1, null
  br i1 %2, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %if.done
  br i1 false, label %store.throw, label %store.next

store.next:                                       ; preds = %gep.next
  %3 = getelementptr inbounds %main.Cmd, ptr %1, i32 0, i32 11
  store i32 %ctx.typecode, ptr %3, align 4
  %.repack1 = getelementptr inbounds %main.Cmd, ptr %1, i32 0, i32 11, i32 1
  store ptr %ctx.value, ptr %.repack1, align 4
  ret ptr %1

gep.throw:                                        ; preds = %if.done
  call void @runtime.nilPanic(ptr undef) #14
  unreachable

store.throw:                                      ; preds = %gep.next
  unreachable
}

declare i1 @runtime.interfaceEqual(i32, ptr, i32, ptr, ptr) #0

; Function Attrs: nounwind
define internal void @"main.interfaceEqual$1"(ptr %context) unnamed_addr #1 {
entry:
  %0 = call %runtime._interface @runtime._recover(i1 false, ptr undef) #14
  %1 = extractvalue %runtime._interface %0, 1
  call void @runtime.trackPointer(ptr %1, ptr undef) #14
  ret void
}

declare %runtime._interface @runtime._recover(i1, ptr) #0

declare %runtime._string @"path/filepath.VolumeName"(ptr, i32, ptr) #0

declare i1 @os.IsPathSeparator(i8, ptr) #0

declare %runtime._string @"path/filepath.Join"(ptr, i32, i32, ptr) #0

declare %runtime._string @strings.TrimPrefix(ptr, i32, ptr, i32, ptr) #0

; Function Attrs: nounwind
define linkonce_odr hidden void @"(*main.closeOnce).close$bound$bound method wrapper for func (*main.closeOnce).close()"(ptr %context) unnamed_addr #1 {
entry:
  call void @"(*main.closeOnce).close"(ptr %context, ptr undef)
  ret void
}

declare %runtime._string @strconv.FormatInt(i64, i32, ptr) #0

declare i32 @runtime.sliceCopy(ptr nocapture writeonly, ptr nocapture readonly, i32, i32, i32, ptr) #0

; Function Attrs: nounwind
define hidden i32 @main.minInt(i32 %a, i32 %b, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp slt i32 %a, %b
  br i1 %0, label %if.then, label %if.done

if.then:                                          ; preds = %entry
  ret i32 %a

if.done:                                          ; preds = %entry
  ret i32 %b
}

; Function Attrs: nounwind
define hidden { ptr, i32, i32 } @main.dedupEnvCase(i1 %caseInsensitive, ptr %env.data, i32 %env.len, i32 %env.cap, ptr %context) unnamed_addr #1 {
entry:
  %hashmap.value26 = alloca i1, align 1
  %hashmap.value = alloca i1, align 1
  %slice.maxcap = icmp ugt i32 %env.len, 536870911
  br i1 %slice.maxcap, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %entry
  %makeslice.cap = shl i32 %env.len, 3
  %makeslice.buf = call ptr @runtime.alloc(i32 %makeslice.cap, ptr nonnull inttoptr (i32 69 to ptr), ptr undef) #14
  %0 = insertvalue { ptr, i32, i32 } undef, ptr %makeslice.buf, 0
  %1 = insertvalue { ptr, i32, i32 } %0, i32 0, 1
  %2 = insertvalue { ptr, i32, i32 } %1, i32 %env.len, 2
  call void @runtime.trackPointer(ptr nonnull %makeslice.buf, ptr undef) #14
  %3 = call ptr @runtime.hashmapMake(i8 8, i8 1, i32 %env.len, i8 1, ptr undef) #14
  call void @runtime.trackPointer(ptr %3, ptr undef) #14
  br label %for.loop

for.loop:                                         ; preds = %for.post, %slice.next
  %4 = phi { ptr, i32, i32 } [ %2, %slice.next ], [ %22, %for.post ]
  %5 = phi i32 [ %env.len, %slice.next ], [ %24, %for.post ]
  %6 = extractvalue { ptr, i32, i32 } %4, 0
  call void @runtime.trackPointer(ptr %6, ptr undef) #14
  %7 = icmp sgt i32 %5, 0
  br i1 %7, label %for.body, label %for.done

for.body:                                         ; preds = %for.loop
  %8 = add i32 %5, -1
  %.not68 = icmp ult i32 %8, %env.len
  br i1 %.not68, label %lookup.next, label %lookup.throw

lookup.next:                                      ; preds = %for.body
  %9 = getelementptr inbounds %runtime._string, ptr %env.data, i32 %8
  %.unpack69 = load ptr, ptr %9, align 4
  %.elt70 = getelementptr inbounds %runtime._string, ptr %env.data, i32 %8, i32 1
  %.unpack71 = load i32, ptr %.elt70, align 4
  call void @runtime.trackPointer(ptr %.unpack69, ptr undef) #14
  %10 = call i32 @strings.Index(ptr %.unpack69, i32 %.unpack71, ptr nonnull @"main$string.80", i32 1, ptr undef) #14
  %11 = icmp eq i32 %10, 0
  br i1 %11, label %if.then, label %if.done

if.then:                                          ; preds = %lookup.next
  %slice.lowhigh12 = icmp eq i32 %.unpack71, 0
  br i1 %slice.lowhigh12, label %slice.throw17, label %slice.next18

slice.next18:                                     ; preds = %if.then
  %12 = getelementptr inbounds i8, ptr %.unpack69, i32 1
  %13 = add i32 %.unpack71, -1
  %14 = call i32 @strings.Index(ptr nonnull %12, i32 %13, ptr nonnull @"main$string.81", i32 1, ptr undef) #14
  %15 = add i32 %14, 1
  br label %if.done

if.done:                                          ; preds = %slice.next18, %lookup.next
  %16 = phi i32 [ %10, %lookup.next ], [ %15, %slice.next18 ]
  %17 = icmp slt i32 %16, 0
  br i1 %17, label %if.then1, label %if.done3

if.then1:                                         ; preds = %if.done
  %18 = call i1 @runtime.stringEqual(ptr %.unpack69, i32 %.unpack71, ptr null, i32 0, ptr undef) #14
  br i1 %18, label %for.post, label %if.then2

if.then2:                                         ; preds = %if.then1
  %varargs = call ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 69 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %varargs, ptr undef) #14
  store ptr %.unpack69, ptr %varargs, align 4
  %varargs.repack74 = getelementptr inbounds %runtime._string, ptr %varargs, i32 0, i32 1
  store i32 %.unpack71, ptr %varargs.repack74, align 4
  %append.srcBuf = extractvalue { ptr, i32, i32 } %4, 0
  %append.srcLen = extractvalue { ptr, i32, i32 } %4, 1
  %append.srcCap = extractvalue { ptr, i32, i32 } %4, 2
  %append.new = call { ptr, i32, i32 } @runtime.sliceAppend(ptr %append.srcBuf, ptr nonnull %varargs, i32 %append.srcLen, i32 %append.srcCap, i32 1, i32 8, ptr undef) #14
  %append.newPtr = extractvalue { ptr, i32, i32 } %append.new, 0
  %append.newLen = extractvalue { ptr, i32, i32 } %append.new, 1
  %append.newCap = extractvalue { ptr, i32, i32 } %append.new, 2
  %19 = insertvalue { ptr, i32, i32 } undef, ptr %append.newPtr, 0
  %20 = insertvalue { ptr, i32, i32 } %19, i32 %append.newLen, 1
  %21 = insertvalue { ptr, i32, i32 } %20, i32 %append.newCap, 2
  call void @runtime.trackPointer(ptr %append.newPtr, ptr undef) #14
  br label %for.post

for.post:                                         ; preds = %if.done6, %if.done5, %if.then2, %if.then1
  %22 = phi { ptr, i32, i32 } [ %4, %if.then1 ], [ %4, %if.done5 ], [ %39, %if.done6 ], [ %21, %if.then2 ]
  %23 = extractvalue { ptr, i32, i32 } %22, 0
  call void @runtime.trackPointer(ptr %23, ptr undef) #14
  %24 = add i32 %5, -1
  br label %for.loop

if.done3:                                         ; preds = %if.done
  %slice.maxcap21 = icmp ugt i32 %16, %.unpack71
  br i1 %slice.maxcap21, label %slice.throw24, label %slice.next25

slice.next25:                                     ; preds = %if.done3
  %25 = insertvalue %runtime._string undef, ptr %.unpack69, 0
  %26 = insertvalue %runtime._string %25, i32 %16, 1
  br i1 %caseInsensitive, label %if.then4, label %if.done5

if.then4:                                         ; preds = %slice.next25
  %27 = call %runtime._string @strings.ToLower(ptr %.unpack69, i32 %16, ptr undef) #14
  %28 = extractvalue %runtime._string %27, 0
  call void @runtime.trackPointer(ptr %28, ptr undef) #14
  br label %if.done5

if.done5:                                         ; preds = %if.then4, %slice.next25
  %29 = phi %runtime._string [ %26, %slice.next25 ], [ %27, %if.then4 ]
  %30 = extractvalue %runtime._string %29, 0
  call void @runtime.trackPointer(ptr %30, ptr undef) #14
  call void @llvm.lifetime.start.p0(i64 1, ptr nonnull %hashmap.value)
  %31 = extractvalue %runtime._string %29, 0
  %32 = extractvalue %runtime._string %29, 1
  %33 = call i1 @runtime.hashmapStringGet(ptr %3, ptr %31, i32 %32, ptr nonnull %hashmap.value, i32 1, ptr undef) #14
  %34 = load i1, ptr %hashmap.value, align 1
  call void @llvm.lifetime.end.p0(i64 1, ptr nonnull %hashmap.value)
  br i1 %34, label %for.post, label %if.done6

if.done6:                                         ; preds = %if.done5
  call void @llvm.lifetime.start.p0(i64 1, ptr nonnull %hashmap.value26)
  store i1 true, ptr %hashmap.value26, align 1
  %35 = extractvalue %runtime._string %29, 0
  %36 = extractvalue %runtime._string %29, 1
  call void @runtime.hashmapStringSet(ptr %3, ptr %35, i32 %36, ptr nonnull %hashmap.value26, ptr undef) #14
  call void @llvm.lifetime.end.p0(i64 1, ptr nonnull %hashmap.value26)
  %varargs27 = call ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 69 to ptr), ptr undef) #14
  call void @runtime.trackPointer(ptr nonnull %varargs27, ptr undef) #14
  store ptr %.unpack69, ptr %varargs27, align 4
  %varargs27.repack72 = getelementptr inbounds %runtime._string, ptr %varargs27, i32 0, i32 1
  store i32 %.unpack71, ptr %varargs27.repack72, align 4
  %append.srcBuf29 = extractvalue { ptr, i32, i32 } %4, 0
  %append.srcLen30 = extractvalue { ptr, i32, i32 } %4, 1
  %append.srcCap31 = extractvalue { ptr, i32, i32 } %4, 2
  %append.new34 = call { ptr, i32, i32 } @runtime.sliceAppend(ptr %append.srcBuf29, ptr nonnull %varargs27, i32 %append.srcLen30, i32 %append.srcCap31, i32 1, i32 8, ptr undef) #14
  %append.newPtr35 = extractvalue { ptr, i32, i32 } %append.new34, 0
  %append.newLen36 = extractvalue { ptr, i32, i32 } %append.new34, 1
  %append.newCap37 = extractvalue { ptr, i32, i32 } %append.new34, 2
  %37 = insertvalue { ptr, i32, i32 } undef, ptr %append.newPtr35, 0
  %38 = insertvalue { ptr, i32, i32 } %37, i32 %append.newLen36, 1
  %39 = insertvalue { ptr, i32, i32 } %38, i32 %append.newCap37, 2
  call void @runtime.trackPointer(ptr %append.newPtr35, ptr undef) #14
  br label %for.post

for.done:                                         ; preds = %for.loop
  br label %for.loop7

for.loop7:                                        ; preds = %lookup.next55, %for.done
  %40 = phi i32 [ 0, %for.done ], [ %49, %lookup.next55 ]
  %len38 = extractvalue { ptr, i32, i32 } %4, 1
  %41 = sdiv i32 %len38, 2
  %42 = icmp slt i32 %40, %41
  br i1 %42, label %for.body8, label %for.done9

for.body8:                                        ; preds = %for.loop7
  %len39 = extractvalue { ptr, i32, i32 } %4, 1
  %43 = xor i32 %40, -1
  %44 = add i32 %len39, %43
  %indexaddr.len41 = extractvalue { ptr, i32, i32 } %4, 1
  %.not = icmp ult i32 %40, %indexaddr.len41
  br i1 %.not, label %lookup.next43, label %lookup.throw42

lookup.next43:                                    ; preds = %for.body8
  %indexaddr.ptr40 = extractvalue { ptr, i32, i32 } %4, 0
  %45 = getelementptr inbounds %runtime._string, ptr %indexaddr.ptr40, i32 %40
  %indexaddr.len45 = extractvalue { ptr, i32, i32 } %4, 1
  %.not56 = icmp ult i32 %44, %indexaddr.len45
  br i1 %.not56, label %lookup.next47, label %lookup.throw46

lookup.next47:                                    ; preds = %lookup.next43
  %indexaddr.ptr44 = extractvalue { ptr, i32, i32 } %4, 0
  %46 = getelementptr inbounds %runtime._string, ptr %indexaddr.ptr44, i32 %44
  %indexaddr.len49 = extractvalue { ptr, i32, i32 } %4, 1
  %.not57 = icmp ult i32 %44, %indexaddr.len49
  br i1 %.not57, label %lookup.next51, label %lookup.throw50

lookup.next51:                                    ; preds = %lookup.next47
  %indexaddr.ptr48 = extractvalue { ptr, i32, i32 } %4, 0
  %47 = getelementptr inbounds %runtime._string, ptr %indexaddr.ptr48, i32 %44
  %.unpack = load ptr, ptr %47, align 4
  %.elt58 = getelementptr inbounds %runtime._string, ptr %indexaddr.ptr48, i32 %44, i32 1
  %.unpack59 = load i32, ptr %.elt58, align 4
  call void @runtime.trackPointer(ptr %.unpack, ptr undef) #14
  %indexaddr.len53 = extractvalue { ptr, i32, i32 } %4, 1
  %.not60 = icmp ult i32 %40, %indexaddr.len53
  br i1 %.not60, label %lookup.next55, label %lookup.throw54

lookup.next55:                                    ; preds = %lookup.next51
  %indexaddr.ptr52 = extractvalue { ptr, i32, i32 } %4, 0
  %48 = getelementptr inbounds %runtime._string, ptr %indexaddr.ptr52, i32 %40
  %.unpack61 = load ptr, ptr %48, align 4
  %.elt62 = getelementptr inbounds %runtime._string, ptr %indexaddr.ptr52, i32 %40, i32 1
  %.unpack63 = load i32, ptr %.elt62, align 4
  call void @runtime.trackPointer(ptr %.unpack61, ptr undef) #14
  store ptr %.unpack, ptr %45, align 4
  %.repack64 = getelementptr inbounds %runtime._string, ptr %indexaddr.ptr40, i32 %40, i32 1
  store i32 %.unpack59, ptr %.repack64, align 4
  store ptr %.unpack61, ptr %46, align 4
  %.repack66 = getelementptr inbounds %runtime._string, ptr %indexaddr.ptr44, i32 %44, i32 1
  store i32 %.unpack63, ptr %.repack66, align 4
  %49 = add i32 %40, 1
  br label %for.loop7

for.done9:                                        ; preds = %for.loop7
  ret { ptr, i32, i32 } %4

slice.throw:                                      ; preds = %entry
  call void @runtime.slicePanic(ptr undef) #14
  unreachable

lookup.throw:                                     ; preds = %for.body
  call void @runtime.lookupPanic(ptr undef) #14
  unreachable

slice.throw17:                                    ; preds = %if.then
  call void @runtime.slicePanic(ptr undef) #14
  unreachable

slice.throw24:                                    ; preds = %if.done3
  call void @runtime.slicePanic(ptr undef) #14
  unreachable

lookup.throw42:                                   ; preds = %for.body8
  call void @runtime.lookupPanic(ptr undef) #14
  unreachable

lookup.throw46:                                   ; preds = %lookup.next43
  call void @runtime.lookupPanic(ptr undef) #14
  unreachable

lookup.throw50:                                   ; preds = %lookup.next47
  call void @runtime.lookupPanic(ptr undef) #14
  unreachable

lookup.throw54:                                   ; preds = %lookup.next51
  call void @runtime.lookupPanic(ptr undef) #14
  unreachable
}

declare ptr @runtime.hashmapMake(i8, i8, i32, i8, ptr) #0

declare i32 @strings.Index(ptr, i32, ptr, i32, ptr) #0

declare %runtime._string @strings.ToLower(ptr, i32, ptr) #0

declare i1 @runtime.hashmapStringGet(ptr dereferenceable_or_null(32), ptr, i32, ptr, i32, ptr) #0

declare void @runtime.hashmapStringSet(ptr dereferenceable_or_null(32), ptr, i32, ptr, ptr) #0

declare { %runtime._string, %runtime._string, i1 } @strings.Cut(ptr, i32, ptr, i32, ptr) #0

declare i1 @strings.EqualFold(ptr, i32, ptr, i32, ptr) #0

declare %runtime._string @os.Getenv(ptr, i32, ptr) #0

attributes #0 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #2 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "tinygo-invoke"="reflect/methods.Error() string" "tinygo-methods"="reflect/methods.Error() string" }
attributes #3 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "tinygo-invoke"="reflect/methods.Done() <-chan (struct{})" "tinygo-methods"="reflect/methods.Deadline() (time.Time, bool); reflect/methods.Done() <-chan (struct{}); reflect/methods.Err() error; reflect/methods.Value(interface{}) interface{}" }
attributes #4 = { argmemonly nocallback nofree nosync nounwind willreturn }
attributes #5 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "tinygo-invoke"="reflect/methods.Err() error" "tinygo-methods"="reflect/methods.Deadline() (time.Time, bool); reflect/methods.Done() <-chan (struct{}); reflect/methods.Err() error; reflect/methods.Value(interface{}) interface{}" }
attributes #6 = { nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "tinygo-gowrapper"="(*main.Cmd).Start$1" }
attributes #7 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "tinygo-methods"="reflect/methods.Close() error; reflect/methods.Read([]uint8) (int, error); reflect/methods.ReadAt([]uint8, int64) (int, error); reflect/methods.Seek(int64, int) (int64, error); reflect/methods.Write([]uint8) (int, error)" }
attributes #8 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "wasm-import-module"="env" "wasm-import-name"="tinygo_unwind" }
attributes #9 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "wasm-import-module"="env" "wasm-import-name"="tinygo_launch" }
attributes #10 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "wasm-import-module"="env" "wasm-import-name"="tinygo_rewind" }
attributes #11 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "tinygo-methods"="reflect/methods.Error() string" }
attributes #12 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "tinygo-invoke"="reflect/methods.Close() error" "tinygo-methods"="reflect/methods.Close() error" }
attributes #13 = { nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "tinygo-gowrapper"="(*main.Cmd).watchCtx$1" }
attributes #14 = { nounwind }
