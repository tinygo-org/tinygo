; ModuleID = 'wrapper.go'
source_filename = "wrapper.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime.typecodeID = type { ptr, i32, ptr, ptr, i32 }
%runtime.structField = type { ptr, ptr, ptr, i1 }
%runtime.interfaceMethodInfo = type { ptr, i32 }
%runtime._interface = type { i32, ptr }
%"github.com/tinygo-org/tinygo/compiler/testdata/os.File" = type { ptr }

@"reflect/types.type:pointer:named:github.com/tinygo-org/tinygo/compiler/testdata/os.File" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:github.com/tinygo-org/tinygo/compiler/testdata/os.File", i32 0, ptr @"*github.com/tinygo-org/tinygo/compiler/testdata/os.File$methodset", ptr null, i32 0 }
@"reflect/types.type:named:github.com/tinygo-org/tinygo/compiler/testdata/os.File" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{#file:pointer:named:github.com/tinygo-org/tinygo/compiler/testdata/os.file}", i32 0, ptr @"github.com/tinygo-org/tinygo/compiler/testdata/os.File$methodset", ptr @"reflect/types.type:pointer:named:github.com/tinygo-org/tinygo/compiler/testdata/os.File", i32 0 }
@"reflect/types.type:struct:{#file:pointer:named:github.com/tinygo-org/tinygo/compiler/testdata/os.file}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields", i32 0, ptr @"struct{*github.com/tinygo-org/tinygo/compiler/testdata/os.file}$methodset", ptr @"reflect/types.type:pointer:struct:{#file:pointer:named:github.com/tinygo-org/tinygo/compiler/testdata/os.file}", i32 0 }
@"reflect/types.structFields" = private unnamed_addr global [1 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:pointer:named:github.com/tinygo-org/tinygo/compiler/testdata/os.file", ptr @"reflect/types.structFieldName.8", ptr null, i1 true }]
@"reflect/types.type:pointer:named:github.com/tinygo-org/tinygo/compiler/testdata/os.file" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:github.com/tinygo-org/tinygo/compiler/testdata/os.file", i32 0, ptr @"*github.com/tinygo-org/tinygo/compiler/testdata/os.file$methodset", ptr null, i32 0 }
@"reflect/types.type:named:github.com/tinygo-org/tinygo/compiler/testdata/os.file" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{handle:named:github.com/tinygo-org/tinygo/compiler/testdata/os.FileHandle,name:basic:string,dirinfo:pointer:named:github.com/tinygo-org/tinygo/compiler/testdata/os.dirInfo}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:github.com/tinygo-org/tinygo/compiler/testdata/os.file", i32 0 }
@"reflect/types.type:struct:{handle:named:github.com/tinygo-org/tinygo/compiler/testdata/os.FileHandle,name:basic:string,dirinfo:pointer:named:github.com/tinygo-org/tinygo/compiler/testdata/os.dirInfo}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.1", i32 0, ptr null, ptr @"reflect/types.type:pointer:struct:{handle:named:github.com/tinygo-org/tinygo/compiler/testdata/os.FileHandle,name:basic:string,dirinfo:pointer:named:github.com/tinygo-org/tinygo/compiler/testdata/os.dirInfo}", i32 0 }
@"reflect/types.structFields.1" = private unnamed_addr global [3 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:named:github.com/tinygo-org/tinygo/compiler/testdata/os.FileHandle", ptr @"reflect/types.structFieldName", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:basic:string", ptr @"reflect/types.structFieldName.2", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:pointer:named:github.com/tinygo-org/tinygo/compiler/testdata/os.dirInfo", ptr @"reflect/types.structFieldName.7", ptr null, i1 false }]
@"reflect/types.type:named:github.com/tinygo-org/tinygo/compiler/testdata/os.FileHandle" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:interface:{Close:func:{}{named:error}}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:github.com/tinygo-org/tinygo/compiler/testdata/os.FileHandle", i32 ptrtoint (ptr @"interface:{Close:func:{}{named:error}}.$typeassert" to i32) }
@"reflect/types.type:interface:{Close:func:{}{named:error}}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.interface:interface{Close() error}$interface", i32 0, ptr null, ptr @"reflect/types.type:pointer:interface:{Close:func:{}{named:error}}", i32 ptrtoint (ptr @"interface:{Close:func:{}{named:error}}.$typeassert" to i32) }
@"reflect/methods.Close() error" = linkonce_odr constant i8 0, align 1
@"reflect/types.interface:interface{Close() error}$interface" = linkonce_odr constant [1 x ptr] [ptr @"reflect/methods.Close() error"]
@"reflect/types.type:pointer:interface:{Close:func:{}{named:error}}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:interface:{Close:func:{}{named:error}}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.type:pointer:named:github.com/tinygo-org/tinygo/compiler/testdata/os.FileHandle" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:github.com/tinygo-org/tinygo/compiler/testdata/os.FileHandle", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.structFieldName" = private unnamed_addr global [6 x i8] c"handle"
@"reflect/types.type:basic:string" = linkonce_odr constant %runtime.typecodeID { ptr null, i32 0, ptr null, ptr @"reflect/types.type:pointer:basic:string", i32 0 }
@"reflect/types.type:pointer:basic:string" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:basic:string", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.structFieldName.2" = private unnamed_addr global [4 x i8] c"name"
@"reflect/types.type:pointer:named:github.com/tinygo-org/tinygo/compiler/testdata/os.dirInfo" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:github.com/tinygo-org/tinygo/compiler/testdata/os.dirInfo", i32 0, ptr @"*github.com/tinygo-org/tinygo/compiler/testdata/os.dirInfo$methodset", ptr null, i32 0 }
@"reflect/types.type:named:github.com/tinygo-org/tinygo/compiler/testdata/os.dirInfo" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{nbuf:basic:int,bufp:basic:int,buf:array:8184:basic:uint8}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:github.com/tinygo-org/tinygo/compiler/testdata/os.dirInfo", i32 0 }
@"reflect/types.type:struct:{nbuf:basic:int,bufp:basic:int,buf:array:8184:basic:uint8}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.structFields.3", i32 0, ptr null, ptr @"reflect/types.type:pointer:struct:{nbuf:basic:int,bufp:basic:int,buf:array:8184:basic:uint8}", i32 0 }
@"reflect/types.structFields.3" = private unnamed_addr global [3 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:basic:int", ptr @"reflect/types.structFieldName.4", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:basic:int", ptr @"reflect/types.structFieldName.5", ptr null, i1 false }, %runtime.structField { ptr @"reflect/types.type:array:8184:basic:uint8", ptr @"reflect/types.structFieldName.6", ptr null, i1 false }]
@"reflect/types.type:basic:int" = linkonce_odr constant %runtime.typecodeID { ptr null, i32 0, ptr null, ptr @"reflect/types.type:pointer:basic:int", i32 0 }
@"reflect/types.type:pointer:basic:int" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:basic:int", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.structFieldName.4" = private unnamed_addr global [4 x i8] c"nbuf"
@"reflect/types.structFieldName.5" = private unnamed_addr global [4 x i8] c"bufp"
@"reflect/types.type:array:8184:basic:uint8" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:basic:uint8", i32 8184, ptr null, ptr @"reflect/types.type:pointer:array:8184:basic:uint8", i32 0 }
@"reflect/types.type:basic:uint8" = linkonce_odr constant %runtime.typecodeID { ptr null, i32 0, ptr null, ptr @"reflect/types.type:pointer:basic:uint8", i32 0 }
@"reflect/types.type:pointer:basic:uint8" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:basic:uint8", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.type:pointer:array:8184:basic:uint8" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:array:8184:basic:uint8", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.structFieldName.6" = private unnamed_addr global [3 x i8] c"buf"
@"reflect/types.type:pointer:struct:{nbuf:basic:int,bufp:basic:int,buf:array:8184:basic:uint8}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{nbuf:basic:int,bufp:basic:int,buf:array:8184:basic:uint8}", i32 0, ptr null, ptr null, i32 0 }
@"github.com/tinygo-org/tinygo/compiler/testdata/os.$methods.close()" = linkonce_odr constant i8 0, align 1
@"*github.com/tinygo-org/tinygo/compiler/testdata/os.dirInfo$methodset" = linkonce_odr constant [1 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"github.com/tinygo-org/tinygo/compiler/testdata/os.$methods.close()", i32 ptrtoint (ptr @"(*github.com/tinygo-org/tinygo/compiler/testdata/os.dirInfo).close" to i32) }]
@"reflect/types.structFieldName.7" = private unnamed_addr global [7 x i8] c"dirinfo"
@"reflect/types.type:pointer:struct:{handle:named:github.com/tinygo-org/tinygo/compiler/testdata/os.FileHandle,name:basic:string,dirinfo:pointer:named:github.com/tinygo-org/tinygo/compiler/testdata/os.dirInfo}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{handle:named:github.com/tinygo-org/tinygo/compiler/testdata/os.FileHandle,name:basic:string,dirinfo:pointer:named:github.com/tinygo-org/tinygo/compiler/testdata/os.dirInfo}", i32 0, ptr null, ptr null, i32 0 }
@"github.com/tinygo-org/tinygo/compiler/testdata/os.$methods.close() error" = linkonce_odr constant i8 0, align 1
@"*github.com/tinygo-org/tinygo/compiler/testdata/os.file$methodset" = linkonce_odr constant [1 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"github.com/tinygo-org/tinygo/compiler/testdata/os.$methods.close() error", i32 ptrtoint (ptr @"(*github.com/tinygo-org/tinygo/compiler/testdata/os.file).close" to i32) }]
@"reflect/types.structFieldName.8" = private unnamed_addr global [4 x i8] c"file"
@"struct{*github.com/tinygo-org/tinygo/compiler/testdata/os.file}$methodset" = linkonce_odr constant [1 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"github.com/tinygo-org/tinygo/compiler/testdata/os.$methods.close() error", i32 ptrtoint (ptr @"(struct{*github.com/tinygo-org/tinygo/compiler/testdata/os.file}).close$invoke" to i32) }]
@"reflect/types.type:pointer:struct:{#file:pointer:named:github.com/tinygo-org/tinygo/compiler/testdata/os.file}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:struct:{#file:pointer:named:github.com/tinygo-org/tinygo/compiler/testdata/os.file}", i32 0, ptr @"*struct{*github.com/tinygo-org/tinygo/compiler/testdata/os.file}$methodset", ptr null, i32 0 }
@"*struct{*github.com/tinygo-org/tinygo/compiler/testdata/os.file}$methodset" = linkonce_odr constant [1 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"github.com/tinygo-org/tinygo/compiler/testdata/os.$methods.close() error", i32 ptrtoint (ptr @"(*struct{*github.com/tinygo-org/tinygo/compiler/testdata/os.file}).close" to i32) }]
@"github.com/tinygo-org/tinygo/compiler/testdata/os.File$methodset" = linkonce_odr constant [1 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"github.com/tinygo-org/tinygo/compiler/testdata/os.$methods.close() error", i32 ptrtoint (ptr @"(github.com/tinygo-org/tinygo/compiler/testdata/os.File).close$invoke" to i32) }]
@"reflect/methods.Read([]uint8) (int, error)" = linkonce_odr constant i8 0, align 1
@"reflect/methods.Write([]uint8) (int, error)" = linkonce_odr constant i8 0, align 1
@"*github.com/tinygo-org/tinygo/compiler/testdata/os.File$methodset" = linkonce_odr constant [4 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { ptr @"reflect/methods.Close() error", i32 ptrtoint (ptr @"(*github.com/tinygo-org/tinygo/compiler/testdata/os.File).Close" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Read([]uint8) (int, error)", i32 ptrtoint (ptr @"(*github.com/tinygo-org/tinygo/compiler/testdata/os.File).Read" to i32) }, %runtime.interfaceMethodInfo { ptr @"reflect/methods.Write([]uint8) (int, error)", i32 ptrtoint (ptr @"(*github.com/tinygo-org/tinygo/compiler/testdata/os.File).Write" to i32) }, %runtime.interfaceMethodInfo { ptr @"github.com/tinygo-org/tinygo/compiler/testdata/os.$methods.close() error", i32 ptrtoint (ptr @"(*github.com/tinygo-org/tinygo/compiler/testdata/os.File).close" to i32) }]
@main.stdin = hidden global %runtime._interface zeroinitializer, align 8
@main.cmd = hidden global ptr null, align 4

declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #0

declare void @runtime.trackPointer(ptr nocapture readonly, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  %new = call ptr @runtime.alloc(i32 4, ptr nonnull inttoptr (i32 67 to ptr), ptr undef) #5
  call void @runtime.trackPointer(ptr nonnull %new, ptr undef) #5
  call void @runtime.trackPointer(ptr nonnull %new, ptr undef) #5
  store i32 ptrtoint (ptr @"reflect/types.type:pointer:named:github.com/tinygo-org/tinygo/compiler/testdata/os.File" to i32), ptr @main.stdin, align 8
  store ptr %new, ptr getelementptr inbounds (%runtime._interface, ptr @main.stdin, i32 0, i32 1), align 4
  %new1 = call ptr @runtime.alloc(i32 32, ptr nonnull inttoptr (i32 2449 to ptr), ptr undef) #5
  call void @runtime.trackPointer(ptr nonnull %new1, ptr undef) #5
  store ptr %new1, ptr @main.cmd, align 4
  ret void
}

declare i1 @"interface:{Close:func:{}{named:error}}.$typeassert"(i32) #2

declare void @"(*github.com/tinygo-org/tinygo/compiler/testdata/os.dirInfo).close"(ptr dereferenceable_or_null(8192), ptr) #0

declare %runtime._interface @"(*github.com/tinygo-org/tinygo/compiler/testdata/os.file).close"(ptr dereferenceable_or_null(20), ptr) #0

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(struct{*github.com/tinygo-org/tinygo/compiler/testdata/os.file}).close"(ptr dereferenceable_or_null(20) %f.file, ptr %context) unnamed_addr #1 {
entry:
  %f = alloca { ptr }, align 8
  store ptr null, ptr %f, align 8
  call void @runtime.trackPointer(ptr nonnull %f, ptr undef) #5
  store ptr %f.file, ptr %f, align 8
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #5
  %1 = call %runtime._interface @"(*github.com/tinygo-org/tinygo/compiler/testdata/os.file).close"(ptr %0, ptr undef) #5
  %2 = extractvalue %runtime._interface %1, 1
  call void @runtime.trackPointer(ptr %2, ptr undef) #5
  ret %runtime._interface %1

deref.throw:                                      ; preds = %entry
  unreachable
}

declare void @runtime.nilPanic(ptr) #0

; Function Attrs: nounwind
define linkonce_odr %runtime._interface @"(struct{*github.com/tinygo-org/tinygo/compiler/testdata/os.file}).close$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
entry:
  %ret = call %runtime._interface @"(struct{*github.com/tinygo-org/tinygo/compiler/testdata/os.file}).close"(ptr %0, ptr %1)
  ret %runtime._interface %ret
}

; Function Attrs: argmemonly nocallback nofree nosync nounwind willreturn
declare void @llvm.lifetime.start.p0(i64 immarg, ptr nocapture) #3

; Function Attrs: argmemonly nocallback nofree nosync nounwind willreturn
declare void @llvm.lifetime.end.p0(i64 immarg, ptr nocapture) #3

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(*struct{*github.com/tinygo-org/tinygo/compiler/testdata/os.file}).close"(ptr dereferenceable_or_null(4) %f, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #5
  %2 = call %runtime._interface @"(*github.com/tinygo-org/tinygo/compiler/testdata/os.file).close"(ptr %1, ptr undef) #5
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #5
  ret %runtime._interface %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #5
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(github.com/tinygo-org/tinygo/compiler/testdata/os.File).close"(ptr dereferenceable_or_null(20) %f.file, ptr %context) unnamed_addr #1 {
entry:
  %f = alloca %"github.com/tinygo-org/tinygo/compiler/testdata/os.File", align 8
  store ptr null, ptr %f, align 8
  call void @runtime.trackPointer(ptr nonnull %f, ptr undef) #5
  store ptr %f.file, ptr %f, align 8
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load ptr, ptr %f, align 8
  call void @runtime.trackPointer(ptr %0, ptr undef) #5
  %1 = call %runtime._interface @"(*github.com/tinygo-org/tinygo/compiler/testdata/os.file).close"(ptr %0, ptr undef) #5
  %2 = extractvalue %runtime._interface %1, 1
  call void @runtime.trackPointer(ptr %2, ptr undef) #5
  ret %runtime._interface %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr %runtime._interface @"(github.com/tinygo-org/tinygo/compiler/testdata/os.File).close$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
entry:
  %ret = call %runtime._interface @"(github.com/tinygo-org/tinygo/compiler/testdata/os.File).close"(ptr %0, ptr %1)
  ret %runtime._interface %ret
}

declare %runtime._interface @"(*github.com/tinygo-org/tinygo/compiler/testdata/os.File).Close"(ptr dereferenceable_or_null(4), ptr) #0

declare { i32, %runtime._interface } @"(*github.com/tinygo-org/tinygo/compiler/testdata/os.File).Read"(ptr dereferenceable_or_null(4), ptr, i32, i32, ptr) #0

declare { i32, %runtime._interface } @"(*github.com/tinygo-org/tinygo/compiler/testdata/os.File).Write"(ptr dereferenceable_or_null(4), ptr, i32, i32, ptr) #0

; Function Attrs: nounwind
define linkonce_odr hidden %runtime._interface @"(*github.com/tinygo-org/tinygo/compiler/testdata/os.File).close"(ptr dereferenceable_or_null(4) %f, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %f, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %1 = load ptr, ptr %f, align 4
  call void @runtime.trackPointer(ptr %1, ptr undef) #5
  %2 = call %runtime._interface @"(*github.com/tinygo-org/tinygo/compiler/testdata/os.file).close"(ptr %1, ptr undef) #5
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #5
  ret %runtime._interface %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #5
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define hidden void @main.main(ptr %context) unnamed_addr #1 {
entry:
  %0 = load ptr, ptr @main.cmd, align 4
  call void @runtime.trackPointer(ptr %0, ptr undef) #5
  %1 = call { %runtime._interface, %runtime._interface } @"(*github.com/tinygo-org/tinygo/compiler/testdata/os/exec.Cmd).StdinPipe"(ptr %0, ptr undef) #5
  %2 = extractvalue { %runtime._interface, %runtime._interface } %1, 0
  %3 = extractvalue %runtime._interface %2, 1
  call void @runtime.trackPointer(ptr %3, ptr undef) #5
  %4 = extractvalue { %runtime._interface, %runtime._interface } %1, 1
  %5 = extractvalue %runtime._interface %4, 1
  call void @runtime.trackPointer(ptr %5, ptr undef) #5
  %6 = extractvalue { %runtime._interface, %runtime._interface } %1, 1
  %7 = extractvalue %runtime._interface %6, 0
  %.not = icmp eq i32 %7, 0
  br i1 %.not, label %if.done, label %if.then

if.then:                                          ; preds = %entry
  %8 = extractvalue %runtime._interface %6, 0
  %9 = extractvalue %runtime._interface %6, 1
  call void @runtime._panic(i32 %8, ptr %9, ptr undef) #5
  unreachable

if.done:                                          ; preds = %entry
  %10 = extractvalue { %runtime._interface, %runtime._interface } %1, 0
  %invoke.func.typecode = extractvalue %runtime._interface %10, 0
  %invoke.func.value = extractvalue %runtime._interface %10, 1
  %11 = call %runtime._interface @"interface:{Close:func:{}{named:error},Write:func:{slice:basic:uint8}{basic:int,named:error}}.Close$invoke"(ptr %invoke.func.value, i32 %invoke.func.typecode, ptr undef) #5
  %12 = extractvalue %runtime._interface %11, 1
  call void @runtime.trackPointer(ptr %12, ptr undef) #5
  %13 = extractvalue %runtime._interface %11, 0
  %.not3 = icmp eq i32 %13, 0
  br i1 %.not3, label %if.done2, label %if.then1

if.then1:                                         ; preds = %if.done
  %14 = extractvalue %runtime._interface %11, 0
  %15 = extractvalue %runtime._interface %11, 1
  call void @runtime._panic(i32 %14, ptr %15, ptr undef) #5
  unreachable

if.done2:                                         ; preds = %if.done
  ret void
}

declare { %runtime._interface, %runtime._interface } @"(*github.com/tinygo-org/tinygo/compiler/testdata/os/exec.Cmd).StdinPipe"(ptr dereferenceable_or_null(32), ptr) #0

declare void @runtime._panic(i32, ptr, ptr) #0

declare %runtime._interface @"interface:{Close:func:{}{named:error},Write:func:{slice:basic:uint8}{basic:int,named:error}}.Close$invoke"(ptr, i32, ptr) #4

attributes #0 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #2 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "tinygo-methods"="reflect/methods.Close() error" }
attributes #3 = { argmemonly nocallback nofree nosync nounwind willreturn }
attributes #4 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "tinygo-invoke"="reflect/methods.Close() error" "tinygo-methods"="reflect/methods.Close() error; reflect/methods.Write([]uint8) (int, error)" }
attributes #5 = { nounwind }
