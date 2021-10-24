; ModuleID = 'goroutine.go'
source_filename = "goroutine.go"
target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-wasi"

%runtime.funcValueWithSignature = type { i32, i8* }
%runtime.channel = type { i32, i32, i8, %runtime.channelBlockedList*, i32, i32, i32, i8* }
%runtime.channelBlockedList = type { %runtime.channelBlockedList*, %"internal/task.Task"*, %runtime.chanSelectState*, { %runtime.channelBlockedList*, i32, i32 } }
%"internal/task.Task" = type { %"internal/task.Task"*, i8*, i64, %"internal/task.state" }
%"internal/task.state" = type { i8* }
%runtime.chanSelectState = type { %runtime.channel*, i8* }

@"main.regularFunctionGoroutine$pack" = private unnamed_addr constant { i32, i8* } { i32 5, i8* undef }
@"main.inlineFunctionGoroutine$pack" = private unnamed_addr constant { i32, i8* } { i32 5, i8* undef }
@"reflect/types.funcid:func:{basic:int}{}" = external constant i8
@"main.closureFunctionGoroutine$1$withSignature" = linkonce_odr constant %runtime.funcValueWithSignature { i32 ptrtoint (void (i32, i8*, i8*)* @"main.closureFunctionGoroutine$1" to i32), i8* @"reflect/types.funcid:func:{basic:int}{}" }
@"main.startInterfaceMethod$string" = internal unnamed_addr constant [4 x i8] c"test", align 1

declare noalias nonnull i8* @runtime.alloc(i32, i8*, i8*)

; Function Attrs: nounwind
define hidden void @main.init(i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @main.regularFunctionGoroutine(i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  call void @"internal/task.start"(i32 ptrtoint (void (i32, i8*, i8*)* @main.regularFunction to i32), i8* bitcast ({ i32, i8* }* @"main.regularFunctionGoroutine$pack" to i8*), i32 undef, i8* undef, i8* null) #0
  ret void
}

declare void @main.regularFunction(i32, i8*, i8*)

declare void @"internal/task.start"(i32, i8*, i32, i8*, i8*)

; Function Attrs: nounwind
define hidden void @main.inlineFunctionGoroutine(i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  call void @"internal/task.start"(i32 ptrtoint (void (i32, i8*, i8*)* @"main.inlineFunctionGoroutine$1" to i32), i8* bitcast ({ i32, i8* }* @"main.inlineFunctionGoroutine$pack" to i8*), i32 undef, i8* undef, i8* null) #0
  ret void
}

; Function Attrs: nounwind
define hidden void @"main.inlineFunctionGoroutine$1"(i32 %x, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @main.closureFunctionGoroutine(i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %n = call i8* @runtime.alloc(i32 4, i8* undef, i8* null) #0
  %0 = bitcast i8* %n to i32*
  store i32 3, i32* %0, align 4
  %1 = call i8* @runtime.alloc(i32 8, i8* undef, i8* null) #0
  %2 = bitcast i8* %1 to i32*
  store i32 5, i32* %2, align 4
  %3 = getelementptr inbounds i8, i8* %1, i32 4
  %4 = bitcast i8* %3 to i8**
  store i8* %n, i8** %4, align 4
  call void @"internal/task.start"(i32 ptrtoint (void (i32, i8*, i8*)* @"main.closureFunctionGoroutine$1" to i32), i8* nonnull %1, i32 undef, i8* undef, i8* null) #0
  %5 = load i32, i32* %0, align 4
  call void @runtime.printint32(i32 %5, i8* undef, i8* null) #0
  ret void
}

; Function Attrs: nounwind
define hidden void @"main.closureFunctionGoroutine$1"(i32 %x, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %unpack.ptr = bitcast i8* %context to i32*
  store i32 7, i32* %unpack.ptr, align 4
  ret void
}

declare void @runtime.printint32(i32, i8*, i8*)

; Function Attrs: nounwind
define hidden void @main.funcGoroutine(i8* %fn.context, i32 %fn.funcptr, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %0 = call i32 @runtime.getFuncPtr(i8* %fn.context, i32 %fn.funcptr, i8* nonnull @"reflect/types.funcid:func:{basic:int}{}", i8* undef, i8* null) #0
  %1 = call i8* @runtime.alloc(i32 8, i8* undef, i8* null) #0
  %2 = bitcast i8* %1 to i32*
  store i32 5, i32* %2, align 4
  %3 = getelementptr inbounds i8, i8* %1, i32 4
  %4 = bitcast i8* %3 to i8**
  store i8* %fn.context, i8** %4, align 4
  call void @"internal/task.start"(i32 %0, i8* nonnull %1, i32 undef, i8* undef, i8* null) #0
  ret void
}

declare i32 @runtime.getFuncPtr(i8*, i32, i8* dereferenceable_or_null(1), i8*, i8*)

; Function Attrs: nounwind
define hidden void @main.recoverBuiltinGoroutine(i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @main.copyBuiltinGoroutine(i8* %dst.data, i32 %dst.len, i32 %dst.cap, i8* %src.data, i32 %src.len, i32 %src.cap, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %copy.n = call i32 @runtime.sliceCopy(i8* %dst.data, i8* %src.data, i32 %dst.len, i32 %src.len, i32 1, i8* undef, i8* null) #0
  ret void
}

declare i32 @runtime.sliceCopy(i8* nocapture writeonly, i8* nocapture readonly, i32, i32, i32, i8*, i8*)

; Function Attrs: nounwind
define hidden void @main.closeBuiltinGoroutine(%runtime.channel* dereferenceable_or_null(32) %ch, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  call void @runtime.chanClose(%runtime.channel* %ch, i8* undef, i8* null) #0
  ret void
}

declare void @runtime.chanClose(%runtime.channel* dereferenceable_or_null(32), i8*, i8*)

; Function Attrs: nounwind
define hidden void @main.startInterfaceMethod(i32 %itf.typecode, i8* %itf.value, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %0 = call i8* @runtime.alloc(i32 16, i8* undef, i8* null) #0
  %1 = bitcast i8* %0 to i8**
  store i8* %itf.value, i8** %1, align 4
  %2 = getelementptr inbounds i8, i8* %0, i32 4
  %.repack = bitcast i8* %2 to i8**
  store i8* getelementptr inbounds ([4 x i8], [4 x i8]* @"main.startInterfaceMethod$string", i32 0, i32 0), i8** %.repack, align 4
  %.repack1 = getelementptr inbounds i8, i8* %0, i32 8
  %3 = bitcast i8* %.repack1 to i32*
  store i32 4, i32* %3, align 4
  %4 = getelementptr inbounds i8, i8* %0, i32 12
  %5 = bitcast i8* %4 to i32*
  store i32 %itf.typecode, i32* %5, align 4
  call void @"internal/task.start"(i32 ptrtoint (void (i8*, i8*, i32, i32, i8*, i8*)* @"interface:{Print:func:{basic:string}{}}.Print$invoke" to i32), i8* nonnull %0, i32 undef, i8* undef, i8* null) #0
  ret void
}

declare void @"interface:{Print:func:{basic:string}{}}.Print$invoke"(i8*, i8*, i32, i32, i8*, i8*) #1

attributes #0 = { nounwind }
attributes #1 = { "tinygo-invoke"="reflect/methods.Print(string)" "tinygo-methods"="reflect/methods.Print(string)" }
