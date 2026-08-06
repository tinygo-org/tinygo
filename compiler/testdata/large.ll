; ModuleID = 'large.go'
source_filename = "large.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-i128:128-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime._string = type { ptr, i32 }
%runtime.deferFrame = type { ptr, ptr, [0 x ptr], ptr, i8, %runtime._interface, ptr }
%runtime._interface = type { ptr, ptr }
%runtime.channelOp = type { ptr, ptr, i32, ptr }
%runtime.chanSelectState = type { ptr, ptr }

@tinygo_rewinding = external global i8
@tinygo_panic_rewinding = external global i8
@"runtime/gc.layout:258-000000000000000000000000000000000000000000000000000000000000000002" = linkonce_odr unnamed_addr constant { i32, [33 x i8] } { i32 258, [33 x i8] c"\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\02" }
@"reflect/types.typeid:named:main.largeValue" = external constant i8
@llvm.used = appending global [15 x ptr] [ptr @"(main.largeReceiver).makeLargeValue", ptr @"(main.largeReceiver).readLargeValue", ptr @main.makeLargeValue, ptr @main.makeZeroLargeValue, ptr @main.readLargeValue, ptr @main.deferLargeValue, ptr @main.goLargeValue, ptr @main.makeLargeResults, ptr @main.makeTwoLargeResults, ptr @main.makeMixedLargeResults, ptr @main.chooseLargeValue, ptr @main.makePointerLargeValue, ptr @main.useLargeMap, ptr @main.useLargeChannel, ptr @main.selectLargeChannel]
@"main$string" = internal unnamed_addr constant [31 x i8] c"blocking select matched no case", align 1
@"main$pack" = internal unnamed_addr constant { %runtime._string } { %runtime._string { ptr @"main$string", i32 31 } }
@"reflect/types.type:basic:string" = linkonce_odr constant { i8, ptr } { i8 81, ptr @"reflect/types.type:pointer:basic:string" }, align 4
@"reflect/types.type:pointer:basic:string" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:basic:string" }, align 4

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @"(main.largeReceiver).makeLargeValue"(ptr dereferenceable_or_null(1025) %return, ptr readonly dereferenceable_or_null(1025) %receiver, ptr %context) unnamed_addr #1 {
entry:
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 1 dereferenceable(1025) %return, ptr noundef nonnull align 1 dereferenceable(1025) %receiver, i32 1025, i1 false)
  ret void
}

; Function Attrs: nocallback nofree nounwind willreturn memory(argmem: readwrite)
declare void @llvm.memcpy.p0.p0.i32(ptr noalias nocapture writeonly, ptr noalias nocapture readonly, i32, i1 immarg) #2

; Function Attrs: nounwind
define hidden i8 @"(main.largeReceiver).readLargeValue"(ptr readonly dereferenceable_or_null(1025) %receiver, ptr readonly dereferenceable_or_null(1025) %value, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %value1 = call align 1 dereferenceable(1025) ptr @runtime.alloc(i32 1025, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %value1, ptr nonnull %stackalloc, ptr undef) #13
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 1 dereferenceable(1025) %value1, ptr noundef nonnull align 1 dereferenceable(1025) %value, i32 1025, i1 false)
  %0 = getelementptr inbounds nuw i8, ptr %value1, i32 1024
  %1 = load i8, ptr %0, align 1
  ret i8 %1
}

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #3

; Function Attrs: nounwind
define hidden void @main.makeLargeValue(ptr dereferenceable_or_null(1025) %return, i8 %value, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %result = call align 1 dereferenceable(1025) ptr @runtime.alloc(i32 1025, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %result, ptr nonnull %stackalloc, ptr undef) #13
  %0 = getelementptr inbounds nuw i8, ptr %result, i32 1024
  store i8 %value, ptr %0, align 1
  %1 = call align 1 dereferenceable(1025) ptr @runtime.alloc(i32 1025, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %1, ptr nonnull %stackalloc, ptr undef) #13
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 1 dereferenceable(1025) %1, ptr noundef nonnull align 1 dereferenceable(1025) %result, i32 1025, i1 false)
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 1 dereferenceable(1025) %return, ptr noundef nonnull align 1 dereferenceable(1025) %1, i32 1025, i1 false)
  ret void
}

; Function Attrs: nounwind
define hidden void @main.makeZeroLargeValue(ptr dereferenceable_or_null(1025) %return, ptr %context) unnamed_addr #1 {
entry:
  store [1025 x i8] zeroinitializer, ptr %return, align 1
  ret void
}

; Function Attrs: nounwind
define hidden i8 @main.passZeroLargeValue(ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %"main.largeValue{}:main.largeValue" = call align 1 dereferenceable(1025) ptr @runtime.alloc(i32 1025, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %"main.largeValue{}:main.largeValue", ptr nonnull %stackalloc, ptr undef) #13
  store [1025 x i8] zeroinitializer, ptr %"main.largeValue{}:main.largeValue", align 1
  %0 = call i8 @main.readLargeValue(ptr nonnull %"main.largeValue{}:main.largeValue", ptr undef)
  ret i8 %0
}

; Function Attrs: nounwind
define hidden i8 @main.readLargeValue(ptr readonly dereferenceable_or_null(1025) %value, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %value1 = call align 1 dereferenceable(1025) ptr @runtime.alloc(i32 1025, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %value1, ptr nonnull %stackalloc, ptr undef) #13
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 1 dereferenceable(1025) %value1, ptr noundef nonnull align 1 dereferenceable(1025) %value, i32 1025, i1 false)
  %0 = getelementptr inbounds nuw i8, ptr %value1, i32 1024
  %1 = load i8, ptr %0, align 1
  ret i8 %1
}

; Function Attrs: nounwind
define hidden i8 @main.useLargeValue(ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %call.result = call align 1 dereferenceable(1025) ptr @runtime.alloc(i32 1025, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %call.result, ptr nonnull %stackalloc, ptr undef) #13
  call void @main.makeLargeValue(ptr nonnull %call.result, i8 42, ptr undef)
  %0 = call i8 @main.readLargeValue(ptr nonnull %call.result, ptr undef)
  ret i8 %0
}

; Function Attrs: nounwind
define hidden i8 @main.useLargeFunctionValue(ptr %fn.context, ptr %fn.funcptr, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %call.result = call align 1 dereferenceable(1025) ptr @runtime.alloc(i32 1025, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %call.result, ptr nonnull %stackalloc, ptr undef) #13
  call void @main.makeLargeValue(ptr nonnull %call.result, i8 42, ptr undef)
  %0 = icmp eq ptr %fn.funcptr, null
  br i1 %0, label %fpcall.throw, label %fpcall.next

fpcall.next:                                      ; preds = %entry
  %1 = call i8 %fn.funcptr(ptr nonnull %call.result, ptr %fn.context) #13
  ret i8 %1

fpcall.throw:                                     ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #13
  br label %unwind.return

unwind.return:                                    ; preds = %fpcall.throw
  ret i8 undef
}

declare void @runtime.nilPanic(ptr) #0

; Function Attrs: nounwind
define hidden i8 @main.useLargeInterface(ptr %value.typecode, ptr %value.value, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %call.result = call align 1 dereferenceable(1025) ptr @runtime.alloc(i32 1025, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %call.result, ptr nonnull %stackalloc, ptr undef) #13
  call void @"interface:{main.makeLargeValue:func:{}{named:main.largeValue},main.readLargeValue:func:{named:main.largeValue}{basic:uint8}}.makeLargeValue$invoke"(ptr nonnull %call.result, ptr %value.value, ptr %value.typecode, ptr undef) #13
  %0 = call i8 @"interface:{main.makeLargeValue:func:{}{named:main.largeValue},main.readLargeValue:func:{named:main.largeValue}{basic:uint8}}.readLargeValue$invoke"(ptr %value.value, ptr nonnull %call.result, ptr %value.typecode, ptr undef) #13
  ret i8 %0
}

declare void @"interface:{main.makeLargeValue:func:{}{named:main.largeValue},main.readLargeValue:func:{named:main.largeValue}{basic:uint8}}.makeLargeValue$invoke"(ptr, ptr, ptr, ptr) #4

declare i8 @"interface:{main.makeLargeValue:func:{}{named:main.largeValue},main.readLargeValue:func:{named:main.largeValue}{basic:uint8}}.readLargeValue$invoke"(ptr, ptr, ptr, ptr) #5

; Function Attrs: nounwind
define hidden void @main.deferLargeValue(ptr readonly dereferenceable_or_null(1025) %value, ptr %context) unnamed_addr #1 {
entry:
  %defer.alloca = alloca { i32, ptr, ptr }, align 8
  %deferframe.buf = alloca %runtime.deferFrame, align 8
  %deferPtr = getelementptr inbounds nuw i8, ptr %deferframe.buf, i32 24
  %0 = call ptr @llvm.stacksave.p0()
  call void @runtime.setupDeferFrame(ptr nonnull %deferframe.buf, ptr %0, ptr undef) #13
  %stackalloc = alloca i8, align 1
  %defer.next = load ptr, ptr %deferPtr, align 4
  call void @runtime.trackPointer(ptr nonnull %defer.alloca, ptr nonnull %stackalloc, ptr undef) #13
  store i32 0, ptr %defer.alloca, align 4
  %defer.alloca.repack15 = getelementptr inbounds nuw i8, ptr %defer.alloca, i32 4
  store ptr %defer.next, ptr %defer.alloca.repack15, align 4
  %defer.alloca.repack17 = getelementptr inbounds nuw i8, ptr %defer.alloca, i32 8
  store ptr %value, ptr %defer.alloca.repack17, align 4
  store ptr %defer.alloca, ptr %deferPtr, align 4
  br label %rundefers.block

rundefers.after:                                  ; preds = %rundefers.end
  call void @runtime.destroyDeferFrame(ptr nonnull %deferframe.buf, ptr undef) #13
  %unwind = call i1 @runtime.unwindPending(ptr undef) #13
  br i1 %unwind, label %unwind.return, label %unwind.continue

unwind.continue:                                  ; preds = %rundefers.after
  ret void

rundefers.block:                                  ; preds = %entry
  br label %rundefers.loophead

rundefers.loophead:                               ; preds = %rundefers.callback0, %rundefers.block
  %1 = load ptr, ptr %deferPtr, align 4
  %stackIsNil = icmp eq ptr %1, null
  br i1 %stackIsNil, label %rundefers.end, label %rundefers.loop

rundefers.loop:                                   ; preds = %rundefers.loophead
  %stack.next.gep = getelementptr inbounds nuw i8, ptr %1, i32 4
  %stack.next = load ptr, ptr %stack.next.gep, align 4
  store ptr %stack.next, ptr %deferPtr, align 4
  %callback = load i32, ptr %1, align 4
  switch i32 %callback, label %rundefers.default [
    i32 0, label %rundefers.callback0
  ]

rundefers.callback0:                              ; preds = %rundefers.loop
  %gep = getelementptr inbounds nuw i8, ptr %1, i32 8
  %param = load ptr, ptr %gep, align 4
  %2 = call i8 @main.deferLargeValue.asyncifycatch.0(ptr %param, ptr undef) #13
  call void @runtime.clearUnwind(ptr undef) #13
  br label %rundefers.loophead

rundefers.default:                                ; preds = %rundefers.loop
  unreachable

rundefers.end:                                    ; preds = %rundefers.loophead
  br label %rundefers.after

recover:                                          ; preds = %rundefers.end3
  br i1 poison, label %unwind.return, label %unwind.continue2

unwind.continue2:                                 ; preds = %recover
  ret void

lpad:                                             ; No predecessors!
  br label %rundefers.loophead6

rundefers.loophead6:                              ; preds = %rundefers.callback012, %lpad
  br i1 poison, label %rundefers.end3, label %rundefers.loop5

rundefers.loop5:                                  ; preds = %rundefers.loophead6
  switch i32 poison, label %rundefers.default4 [
    i32 0, label %rundefers.callback012
  ]

rundefers.callback012:                            ; preds = %rundefers.loop5
  br label %rundefers.loophead6

rundefers.default4:                               ; preds = %rundefers.loop5
  unreachable

rundefers.end3:                                   ; preds = %rundefers.loophead6
  br label %recover

unwind.return:                                    ; preds = %recover, %rundefers.after
  ret void
}

; Function Attrs: nocallback nofree nosync nounwind willreturn
declare ptr @llvm.stacksave.p0() #6

declare void @runtime.setupDeferFrame(ptr dereferenceable_or_null(28), ptr, ptr) #0

declare void @runtime.destroyDeferFrame(ptr dereferenceable_or_null(28), ptr) #0

declare i1 @runtime.unwindPending(ptr) #0

; Function Attrs: noinline
define internal i8 @main.deferLargeValue.asyncifycatch.0(ptr %0, ptr %1) #7 {
entry:
  %2 = call i8 @main.readLargeValue(ptr %0, ptr %1)
  %3 = call i1 @runtime.unwindPending(ptr undef)
  br i1 %3, label %unwind.stop, label %return

unwind.stop:                                      ; preds = %entry
  call void @runtime.asyncifyStopUnwindImport()
  call void @runtime.savePanicReplay(i32 ptrtoint (ptr @tinygo.asyncify.panicreplay.0 to i32), i32 ptrtoint (ptr @main.deferLargeValue.asyncifycatch.0 to i32), ptr undef)
  br label %return

return:                                           ; preds = %unwind.stop, %entry
  ret i8 %2
}

; Function Attrs: noinline
define internal void @tinygo.asyncify.panicreplay.0(i32 %0) #7 {
entry:
  %1 = call ptr @runtime.panicRewindData(ptr undef)
  %2 = call ptr @runtime.panicRewindStackPointer(ptr undef)
  %3 = inttoptr i32 %0 to ptr
  call void @tinygo_set_panic_rewind_stack_pointer(ptr %2, ptr undef)
  store i8 1, ptr @tinygo_rewinding, align 1
  store i8 1, ptr @tinygo_panic_rewinding, align 1
  call void @runtime.asyncifyStartRewindImport(ptr %1)
  %4 = call i8 %3(ptr undef, ptr undef)
  unreachable
}

declare ptr @runtime.panicRewindData(ptr) #0

declare ptr @runtime.panicRewindStackPointer(ptr) #0

declare void @tinygo_set_panic_rewind_stack_pointer(ptr, ptr) #0

declare void @runtime.asyncifyStartRewindImport(ptr nocapture) #8

declare void @runtime.asyncifyStopUnwindImport() #9

declare void @runtime.savePanicReplay(i32, i32, ptr) #0

declare void @runtime.clearUnwind(ptr) #0

; Function Attrs: nounwind
define hidden void @main.goLargeValue(ptr readonly dereferenceable_or_null(1025) %value, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %go.param = call align 1 dereferenceable(1025) ptr @runtime.alloc(i32 1025, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %go.param, ptr nonnull %stackalloc, ptr undef) #13
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 1 dereferenceable(1025) %go.param, ptr noundef nonnull align 1 dereferenceable(1025) %value, i32 1025, i1 false)
  call void @"internal/task.start"(i32 ptrtoint (ptr @"main.readLargeValue$gowrapper" to i32), ptr nonnull %go.param, i32 65536, ptr undef) #13
  ret void
}

declare void @runtime.exitGoroutine(ptr) #0

; Function Attrs: nounwind
define linkonce_odr void @"main.readLargeValue$gowrapper"(ptr %0) unnamed_addr #10 {
entry:
  %1 = call i8 @main.readLargeValue(ptr %0, ptr undef)
  call void @runtime.exitGoroutine(ptr undef) #13
  unreachable
}

declare void @"internal/task.start"(i32, ptr, i32, ptr) #0

; Function Attrs: nounwind
define hidden void @main.makeLargeResults(ptr dereferenceable_or_null(1026) %return, i8 %value, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %call.result = call align 1 dereferenceable(1025) ptr @runtime.alloc(i32 1025, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %call.result, ptr nonnull %stackalloc, ptr undef) #13
  call void @main.makeLargeValue(ptr nonnull %call.result, i8 %value, ptr undef)
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 1 dereferenceable(1025) %return, ptr noundef nonnull align 1 dereferenceable(1025) %call.result, i32 1025, i1 false)
  %0 = getelementptr inbounds nuw i8, ptr %return, i32 1025
  store i8 %value, ptr %0, align 1
  ret void
}

; Function Attrs: nounwind
define hidden void @main.makeTwoLargeResults(ptr dereferenceable_or_null(2050) %return, i8 %value, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %call.result = call align 1 dereferenceable(1025) ptr @runtime.alloc(i32 1025, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %call.result, ptr nonnull %stackalloc, ptr undef) #13
  call void @main.makeLargeValue(ptr nonnull %call.result, i8 %value, ptr undef)
  %0 = add i8 %value, 1
  %call.result1 = call align 1 dereferenceable(1025) ptr @runtime.alloc(i32 1025, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %call.result1, ptr nonnull %stackalloc, ptr undef) #13
  call void @main.makeLargeValue(ptr nonnull %call.result1, i8 %0, ptr undef)
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 1 dereferenceable(1025) %return, ptr noundef nonnull align 1 dereferenceable(1025) %call.result, i32 1025, i1 false)
  %1 = getelementptr inbounds nuw i8, ptr %return, i32 1025
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 1 dereferenceable(1025) %1, ptr noundef nonnull align 1 dereferenceable(1025) %call.result1, i32 1025, i1 false)
  ret void
}

; Function Attrs: nounwind
define hidden void @main.makeMixedLargeResults(ptr dereferenceable_or_null(2051) %return, i8 %value, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %call.result = call align 1 dereferenceable(1025) ptr @runtime.alloc(i32 1025, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %call.result, ptr nonnull %stackalloc, ptr undef) #13
  call void @main.makeLargeValue(ptr nonnull %call.result, i8 %value, ptr undef)
  %0 = add i8 %value, 1
  %1 = add i8 %value, 2
  %call.result1 = call align 1 dereferenceable(1025) ptr @runtime.alloc(i32 1025, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %call.result1, ptr nonnull %stackalloc, ptr undef) #13
  call void @main.makeLargeValue(ptr nonnull %call.result1, i8 %1, ptr undef)
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 1 dereferenceable(1025) %return, ptr noundef nonnull align 1 dereferenceable(1025) %call.result, i32 1025, i1 false)
  %2 = getelementptr inbounds nuw i8, ptr %return, i32 1025
  store i8 %0, ptr %2, align 1
  %3 = getelementptr inbounds nuw i8, ptr %return, i32 1026
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 1 dereferenceable(1025) %3, ptr noundef nonnull align 1 dereferenceable(1025) %call.result1, i32 1025, i1 false)
  ret void
}

; Function Attrs: nounwind
define hidden void @main.chooseLargeValue(ptr dereferenceable_or_null(1025) %return, i1 %flag, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %call.result = call align 1 dereferenceable(1025) ptr @runtime.alloc(i32 1025, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %call.result, ptr nonnull %stackalloc, ptr undef) #13
  call void @main.makeLargeValue(ptr nonnull %call.result, i8 1, ptr undef)
  br i1 %flag, label %if.then, label %if.done

if.then:                                          ; preds = %entry
  %call.result1 = call align 1 dereferenceable(1025) ptr @runtime.alloc(i32 1025, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %call.result1, ptr nonnull %stackalloc, ptr undef) #13
  call void @main.makeLargeValue(ptr nonnull %call.result1, i8 42, ptr undef)
  br label %if.done

if.done:                                          ; preds = %if.then, %entry
  %0 = phi ptr [ %call.result, %entry ], [ %call.result1, %if.then ]
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 1 dereferenceable(1025) %return, ptr noundef nonnull align 1 dereferenceable(1025) %0, i32 1025, i1 false)
  ret void
}

; Function Attrs: nounwind
define hidden void @main.makePointerLargeValue(ptr dereferenceable_or_null(1032) %return, ptr dereferenceable_or_null(1) %value, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %complit = call align 4 dereferenceable(1032) ptr @runtime.alloc(i32 1032, ptr nonnull @"runtime/gc.layout:258-000000000000000000000000000000000000000000000000000000000000000002", ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %complit, ptr nonnull %stackalloc, ptr undef) #13
  br i1 false, label %store.throw, label %store.next

store.next:                                       ; preds = %entry
  %0 = getelementptr inbounds nuw i8, ptr %complit, i32 1028
  store ptr %value, ptr %0, align 4
  %1 = call align 4 dereferenceable(1032) ptr @runtime.alloc(i32 1032, ptr nonnull @"runtime/gc.layout:258-000000000000000000000000000000000000000000000000000000000000000002", ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %1, ptr nonnull %stackalloc, ptr undef) #13
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 4 dereferenceable(1032) %1, ptr noundef nonnull align 4 dereferenceable(1032) %complit, i32 1032, i1 false)
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 1 dereferenceable(1032) %return, ptr noundef nonnull align 4 dereferenceable(1032) %1, i32 1032, i1 false)
  ret void

store.throw:                                      ; preds = %entry
  br label %unwind.return

unwind.return:                                    ; preds = %store.throw
  ret void
}

; Function Attrs: nounwind
define hidden i8 @main.assertLargeValue(ptr %value.typecode, ptr %value.value, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %large = call align 1 dereferenceable(1025) ptr @runtime.alloc(i32 1025, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %large, ptr nonnull %stackalloc, ptr undef) #13
  %typecode = call i1 @runtime.typeAssert(ptr %value.typecode, ptr nonnull @"reflect/types.typeid:named:main.largeValue", ptr undef) #13
  %typeassert.result = call align 1 dereferenceable(1026) ptr @runtime.alloc(i32 1026, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %typeassert.result, ptr nonnull %stackalloc, ptr undef) #13
  call void @llvm.memset.p0.i32(ptr noundef nonnull align 1 dereferenceable(1026) %typeassert.result, i8 0, i32 1026, i1 false)
  br i1 %typecode, label %typeassert.ok, label %typeassert.next

typeassert.next:                                  ; preds = %typeassert.ok, %entry
  %0 = getelementptr inbounds nuw i8, ptr %typeassert.result, i32 1025
  store i1 %typecode, ptr %0, align 1
  %t2 = call align 1 dereferenceable(1025) ptr @runtime.alloc(i32 1025, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %t2, ptr nonnull %stackalloc, ptr undef) #13
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 1 dereferenceable(1025) %t2, ptr noundef nonnull align 1 dereferenceable(1025) %typeassert.result, i32 1025, i1 false)
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 1 dereferenceable(1025) %large, ptr noundef nonnull align 1 dereferenceable(1025) %t2, i32 1025, i1 false)
  br i1 %typecode, label %if.done, label %if.then

typeassert.ok:                                    ; preds = %entry
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 1 dereferenceable(1025) %typeassert.result, ptr noundef nonnull align 1 dereferenceable(1025) %value.value, i32 1025, i1 false)
  br label %typeassert.next

if.done:                                          ; preds = %typeassert.next
  %1 = getelementptr inbounds nuw i8, ptr %large, i32 1024
  %2 = load i8, ptr %1, align 1
  ret i8 %2

if.then:                                          ; preds = %typeassert.next
  ret i8 0
}

declare i1 @runtime.typeAssert(ptr, ptr dereferenceable_or_null(1), ptr) #0

; Function Attrs: nocallback nofree nounwind willreturn memory(argmem: write)
declare void @llvm.memset.p0.i32(ptr nocapture writeonly, i8, i32, i1 immarg) #11

; Function Attrs: nounwind
define hidden i8 @main.useLargeMap(ptr readonly dereferenceable_or_null(1025) %key, ptr readonly dereferenceable_or_null(1025) %value, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call ptr @runtime.hashmapMakeGeneric(i32 1025, i32 1025, i32 1, ptr null, ptr nonnull @runtime.hash32, ptr null, ptr nonnull @runtime.memequal, ptr undef) #13
  call void @runtime.trackPointer(ptr %0, ptr nonnull %stackalloc, ptr undef) #13
  call void @runtime.hashmapBinarySet(ptr %0, ptr %key, ptr %value, ptr undef) #13
  %result = call align 1 dereferenceable(1025) ptr @runtime.alloc(i32 1025, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %result, ptr nonnull %stackalloc, ptr undef) #13
  %hashmap.result = call align 1 dereferenceable(1026) ptr @runtime.alloc(i32 1026, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %hashmap.result, ptr nonnull %stackalloc, ptr undef) #13
  %1 = call i1 @runtime.hashmapBinaryGet(ptr %0, ptr %key, ptr nonnull %hashmap.result, i32 1025, ptr undef) #13
  %2 = getelementptr inbounds nuw i8, ptr %hashmap.result, i32 1025
  store i1 %1, ptr %2, align 1
  %t3 = call align 1 dereferenceable(1025) ptr @runtime.alloc(i32 1025, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %t3, ptr nonnull %stackalloc, ptr undef) #13
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 1 dereferenceable(1025) %t3, ptr noundef nonnull align 1 dereferenceable(1025) %hashmap.result, i32 1025, i1 false)
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 1 dereferenceable(1025) %result, ptr noundef nonnull align 1 dereferenceable(1025) %t3, i32 1025, i1 false)
  %3 = getelementptr inbounds nuw i8, ptr %hashmap.result, i32 1025
  %t4 = load i1, ptr %3, align 1
  br i1 %t4, label %if.done, label %if.then

if.done:                                          ; preds = %entry
  %4 = getelementptr inbounds nuw i8, ptr %result, i32 1024
  %5 = load i8, ptr %4, align 1
  ret i8 %5

if.then:                                          ; preds = %entry
  ret i8 0
}

declare i32 @runtime.hash32(ptr, i32, i32, ptr) #0

declare i1 @runtime.memequal(ptr, ptr, i32, ptr) #0

declare ptr @runtime.hashmapMakeGeneric(i32, i32, i32, ptr, ptr, ptr, ptr, ptr) #0

declare void @runtime.hashmapBinarySet(ptr dereferenceable_or_null(48), ptr, ptr, ptr) #0

declare i1 @runtime.hashmapBinaryGet(ptr dereferenceable_or_null(48), ptr, ptr, i32, ptr) #0

; Function Attrs: nounwind
define hidden i8 @main.useLargeChannel(ptr dereferenceable_or_null(36) %ch, ptr readonly dereferenceable_or_null(1025) %value, ptr %context) unnamed_addr #1 {
entry:
  %chan.op1 = alloca %runtime.channelOp, align 8
  %chan.op = alloca %runtime.channelOp, align 8
  %stackalloc = alloca i8, align 1
  call void @llvm.lifetime.start.p0(ptr nonnull %chan.op)
  call void @runtime.chanSend(ptr %ch, ptr %value, ptr nonnull %chan.op, ptr undef) #13
  call void @llvm.lifetime.end.p0(ptr nonnull %chan.op)
  %result = call align 1 dereferenceable(1025) ptr @runtime.alloc(i32 1025, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %result, ptr nonnull %stackalloc, ptr undef) #13
  %chan.result = call align 1 dereferenceable(1026) ptr @runtime.alloc(i32 1026, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %chan.result, ptr nonnull %stackalloc, ptr undef) #13
  call void @llvm.lifetime.start.p0(ptr nonnull %chan.op1)
  %0 = call i1 @runtime.chanRecv(ptr %ch, ptr nonnull %chan.result, ptr nonnull %chan.op1, ptr undef) #13
  %1 = getelementptr inbounds nuw i8, ptr %chan.result, i32 1025
  store i1 %0, ptr %1, align 1
  call void @llvm.lifetime.end.p0(ptr nonnull %chan.op1)
  %t2 = call align 1 dereferenceable(1025) ptr @runtime.alloc(i32 1025, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %t2, ptr nonnull %stackalloc, ptr undef) #13
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 1 dereferenceable(1025) %t2, ptr noundef nonnull align 1 dereferenceable(1025) %chan.result, i32 1025, i1 false)
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 1 dereferenceable(1025) %result, ptr noundef nonnull align 1 dereferenceable(1025) %t2, i32 1025, i1 false)
  %2 = getelementptr inbounds nuw i8, ptr %chan.result, i32 1025
  %t3 = load i1, ptr %2, align 1
  br i1 %t3, label %if.done, label %if.then

if.done:                                          ; preds = %entry
  %3 = getelementptr inbounds nuw i8, ptr %result, i32 1024
  %4 = load i8, ptr %3, align 1
  ret i8 %4

if.then:                                          ; preds = %entry
  ret i8 0
}

; Function Attrs: nocallback nofree nosync nounwind willreturn memory(argmem: readwrite)
declare void @llvm.lifetime.start.p0(ptr nocapture) #12

declare void @runtime.chanSend(ptr dereferenceable_or_null(36), ptr, ptr dereferenceable_or_null(16), ptr) #0

; Function Attrs: nocallback nofree nosync nounwind willreturn memory(argmem: readwrite)
declare void @llvm.lifetime.end.p0(ptr nocapture) #12

declare i1 @runtime.chanRecv(ptr dereferenceable_or_null(36), ptr, ptr dereferenceable_or_null(16), ptr) #0

; Function Attrs: nounwind
define hidden i8 @main.selectLargeChannel(ptr dereferenceable_or_null(36) %ch, ptr readonly dereferenceable_or_null(1025) %value, ptr %context) unnamed_addr #1 {
entry:
  %select.block.alloca = alloca [2 x %runtime.channelOp], align 8
  %select.states.alloca = alloca [2 x %runtime.chanSelectState], align 8
  %select.recvbuf.alloca = alloca [1025 x i8], align 1
  %stackalloc = alloca i8, align 1
  call void @llvm.lifetime.start.p0(ptr nonnull %select.recvbuf.alloca)
  call void @llvm.lifetime.start.p0(ptr nonnull %select.states.alloca)
  store ptr %ch, ptr %select.states.alloca, align 4
  %select.states.alloca.repack3 = getelementptr inbounds nuw i8, ptr %select.states.alloca, i32 4
  store ptr %value, ptr %select.states.alloca.repack3, align 4
  %0 = getelementptr inbounds nuw i8, ptr %select.states.alloca, i32 8
  store ptr %ch, ptr %0, align 4
  %.repack5 = getelementptr inbounds nuw i8, ptr %select.states.alloca, i32 12
  store ptr null, ptr %.repack5, align 4
  call void @llvm.lifetime.start.p0(ptr nonnull %select.block.alloca)
  %select.result = call { i32, i1 } @runtime.chanSelect(ptr nonnull %select.recvbuf.alloca, ptr nonnull %select.states.alloca, i32 2, i32 2, ptr nonnull %select.block.alloca, i32 2, i32 2, ptr undef) #13
  call void @llvm.lifetime.end.p0(ptr nonnull %select.block.alloca)
  call void @llvm.lifetime.end.p0(ptr nonnull %select.states.alloca)
  call void @runtime.trackPointer(ptr nonnull %select.recvbuf.alloca, ptr nonnull %stackalloc, ptr undef) #13
  %1 = extractvalue { i32, i1 } %select.result, 0
  %2 = icmp eq i32 %1, 0
  br i1 %2, label %select.body, label %select.next

select.body:                                      ; preds = %entry
  ret i8 0

select.next:                                      ; preds = %entry
  %3 = icmp eq i32 %1, 1
  br i1 %3, label %select.body1, label %select.next2

select.body1:                                     ; preds = %select.next
  %result = call align 1 dereferenceable(1025) ptr @runtime.alloc(i32 1025, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %result, ptr nonnull %stackalloc, ptr undef) #13
  %select.received = call align 1 dereferenceable(1025) ptr @runtime.alloc(i32 1025, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull %select.received, ptr nonnull %stackalloc, ptr undef) #13
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 1 dereferenceable(1025) %select.received, ptr noundef nonnull align 1 dereferenceable(1025) %select.recvbuf.alloca, i32 1025, i1 false)
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 1 dereferenceable(1025) %result, ptr noundef nonnull align 1 dereferenceable(1025) %select.received, i32 1025, i1 false)
  %4 = getelementptr inbounds nuw i8, ptr %result, i32 1024
  %5 = load i8, ptr %4, align 1
  ret i8 %5

select.next2:                                     ; preds = %select.next
  call void @runtime.trackPointer(ptr nonnull @"reflect/types.type:basic:string", ptr nonnull %stackalloc, ptr undef) #13
  call void @runtime.trackPointer(ptr nonnull @"main$pack", ptr nonnull %stackalloc, ptr undef) #13
  call void @runtime._panic(ptr nonnull @"reflect/types.type:basic:string", ptr nonnull @"main$pack", ptr undef) #13
  br label %unwind.return

unwind.return:                                    ; preds = %select.next2
  ret i8 undef
}

declare { i32, i1 } @runtime.chanSelect(ptr, ptr, i32, i32, ptr, i32, i32, ptr) #0

declare void @runtime._panic(ptr, ptr, ptr) #0

attributes #0 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { nocallback nofree nounwind willreturn memory(argmem: readwrite) }
attributes #3 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #4 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "tinygo-indirect-result"="true" "tinygo-invoke"="main.$methods.makeLargeValue:func:{}{named:main.largeValue}" "tinygo-methods"="main.$methods.makeLargeValue:func:{}{named:main.largeValue}; main.$methods.readLargeValue:func:{named:main.largeValue}{basic:uint8}" }
attributes #5 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "tinygo-invoke"="main.$methods.readLargeValue:func:{named:main.largeValue}{basic:uint8}" "tinygo-methods"="main.$methods.makeLargeValue:func:{}{named:main.largeValue}; main.$methods.readLargeValue:func:{named:main.largeValue}{basic:uint8}" }
attributes #6 = { nocallback nofree nosync nounwind willreturn }
attributes #7 = { noinline }
attributes #8 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "wasm-import-module"="asyncify" "wasm-import-name"="start_rewind" }
attributes #9 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "wasm-import-module"="asyncify" "wasm-import-name"="stop_unwind" }
attributes #10 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "tinygo-gowrapper"="main.readLargeValue" }
attributes #11 = { nocallback nofree nounwind willreturn memory(argmem: write) }
attributes #12 = { nocallback nofree nosync nounwind willreturn memory(argmem: readwrite) }
attributes #13 = { nounwind }
