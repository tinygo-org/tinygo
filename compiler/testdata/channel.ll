; ModuleID = 'channel.go'
source_filename = "channel.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-i128:128-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime._string = type { ptr, i32 }
%runtime.channelOp = type { ptr, ptr, i32, ptr }
%runtime.chanSelectState = type { ptr, ptr }

@"main$string" = internal unnamed_addr constant [31 x i8] c"blocking select matched no case", align 1
@"main$pack" = internal unnamed_addr constant { %runtime._string } { %runtime._string { ptr @"main$string", i32 31 } }
@"reflect/types.type:basic:string" = linkonce_odr constant { i8, ptr } { i8 81, ptr @"reflect/types.type:pointer:basic:string" }, align 4
@"reflect/types.type:pointer:basic:string" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:basic:string" }, align 4

declare void @runtime.trackPointer(ptr readonly captures(none), ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @main.chanIntSend(ptr dereferenceable_or_null(36) %ch, ptr %context) unnamed_addr #1 {
entry:
  %chan.op = alloca %runtime.channelOp, align 8
  %chan.value = alloca i32, align 4
  call void @llvm.lifetime.start.p0(ptr nonnull %chan.value)
  store i32 3, ptr %chan.value, align 4
  call void @llvm.lifetime.start.p0(ptr nonnull %chan.op)
  call void @runtime.chanSend(ptr %ch, ptr nonnull %chan.value, ptr nonnull %chan.op, ptr undef) #3
  call void @llvm.lifetime.end.p0(ptr nonnull %chan.op)
  call void @llvm.lifetime.end.p0(ptr nonnull %chan.value)
  ret void
}

; Function Attrs: nocallback nofree nosync nounwind willreturn memory(argmem: readwrite)
declare void @llvm.lifetime.start.p0(ptr captures(none)) #2

declare void @runtime.chanSend(ptr dereferenceable_or_null(36), ptr, ptr dereferenceable_or_null(16), ptr) #0

; Function Attrs: nocallback nofree nosync nounwind willreturn memory(argmem: readwrite)
declare void @llvm.lifetime.end.p0(ptr captures(none)) #2

; Function Attrs: nounwind
define hidden void @main.chanIntRecv(ptr dereferenceable_or_null(36) %ch, ptr %context) unnamed_addr #1 {
entry:
  %chan.op = alloca %runtime.channelOp, align 8
  %chan.value = alloca i32, align 4
  call void @llvm.lifetime.start.p0(ptr nonnull %chan.value)
  call void @llvm.lifetime.start.p0(ptr nonnull %chan.op)
  %0 = call i1 @runtime.chanRecv(ptr %ch, ptr nonnull %chan.value, ptr nonnull %chan.op, ptr undef) #3
  call void @llvm.lifetime.end.p0(ptr nonnull %chan.value)
  call void @llvm.lifetime.end.p0(ptr nonnull %chan.op)
  ret void
}

declare i1 @runtime.chanRecv(ptr dereferenceable_or_null(36), ptr, ptr dereferenceable_or_null(16), ptr) #0

; Function Attrs: nounwind
define hidden void @main.chanZeroSend(ptr dereferenceable_or_null(36) %ch, ptr %context) unnamed_addr #1 {
entry:
  %chan.op = alloca %runtime.channelOp, align 8
  call void @llvm.lifetime.start.p0(ptr nonnull %chan.op)
  call void @runtime.chanSend(ptr %ch, ptr null, ptr nonnull %chan.op, ptr undef) #3
  call void @llvm.lifetime.end.p0(ptr nonnull %chan.op)
  ret void
}

; Function Attrs: nounwind
define hidden void @main.chanZeroRecv(ptr dereferenceable_or_null(36) %ch, ptr %context) unnamed_addr #1 {
entry:
  %chan.op = alloca %runtime.channelOp, align 8
  call void @llvm.lifetime.start.p0(ptr nonnull %chan.op)
  %0 = call i1 @runtime.chanRecv(ptr %ch, ptr null, ptr nonnull %chan.op, ptr undef) #3
  call void @llvm.lifetime.end.p0(ptr nonnull %chan.op)
  ret void
}

; Function Attrs: nounwind
define hidden void @main.selectZeroRecv(ptr dereferenceable_or_null(36) %ch1, ptr dereferenceable_or_null(36) %ch2, ptr %context) unnamed_addr #1 {
entry:
  %select.states.alloca = alloca [2 x %runtime.chanSelectState], align 8
  %select.send.value = alloca i32, align 4
  store i32 1, ptr %select.send.value, align 4
  call void @llvm.lifetime.start.p0(ptr nonnull %select.states.alloca)
  store ptr %ch1, ptr %select.states.alloca, align 4
  %select.states.alloca.repack1 = getelementptr inbounds nuw i8, ptr %select.states.alloca, i32 4
  store ptr %select.send.value, ptr %select.states.alloca.repack1, align 4
  %0 = getelementptr inbounds nuw i8, ptr %select.states.alloca, i32 8
  store ptr %ch2, ptr %0, align 4
  %.repack3 = getelementptr inbounds nuw i8, ptr %select.states.alloca, i32 12
  store ptr null, ptr %.repack3, align 4
  %select.result = call { i32, i1 } @runtime.chanSelect(ptr undef, ptr nonnull %select.states.alloca, i32 2, i32 2, ptr null, i32 0, i32 0, ptr undef) #3
  call void @llvm.lifetime.end.p0(ptr nonnull %select.states.alloca)
  %1 = extractvalue { i32, i1 } %select.result, 0
  %2 = icmp eq i32 %1, 0
  br i1 %2, label %select.done, label %select.next

select.done:                                      ; preds = %select.body, %select.next, %entry
  ret void

select.next:                                      ; preds = %entry
  %3 = icmp eq i32 %1, 1
  br i1 %3, label %select.body, label %select.done

select.body:                                      ; preds = %select.next
  br label %select.done
}

declare { i32, i1 } @runtime.chanSelect(ptr, ptr, i32, i32, ptr, i32, i32, ptr) #0

; Function Attrs: nounwind
define hidden i1 @main.selectNonBlockingSend(ptr dereferenceable_or_null(36) %ch, i32 %value, ptr %context) unnamed_addr #1 {
entry:
  %select.send.value = alloca i32, align 4
  call void @llvm.lifetime.start.p0(ptr nonnull %select.send.value)
  store i32 %value, ptr %select.send.value, align 4
  %select.sent = call i1 @runtime.chanTrySend(ptr %ch, ptr nonnull %select.send.value, ptr undef) #3
  call void @llvm.lifetime.end.p0(ptr nonnull %select.send.value)
  br i1 %select.sent, label %select.body, label %select.next

select.body:                                      ; preds = %entry
  ret i1 true

select.next:                                      ; preds = %entry
  ret i1 false
}

declare i1 @runtime.chanTrySend(ptr dereferenceable_or_null(36), ptr, ptr) #0

; Function Attrs: nounwind
define hidden { i32, i1, i1 } @main.selectNonBlockingRecv(ptr dereferenceable_or_null(36) %ch, ptr %context) unnamed_addr #1 {
entry:
  %select.recvbuf = alloca i32, align 4
  %stackalloc = alloca i8, align 1
  call void @llvm.lifetime.start.p0(ptr nonnull %select.recvbuf)
  %select.recv = call { i1, i1 } @runtime.chanTryRecv(ptr %ch, ptr nonnull %select.recvbuf, ptr undef) #3
  %select.received = extractvalue { i1, i1 } %select.recv, 0
  call void @runtime.trackPointer(ptr nonnull %select.recvbuf, ptr nonnull %stackalloc, ptr undef) #3
  br i1 %select.received, label %select.body, label %select.next

select.body:                                      ; preds = %entry
  %select.recv.ok = extractvalue { i1, i1 } %select.recv, 1
  %select.received1 = load i32, ptr %select.recvbuf, align 4
  %0 = insertvalue { i32, i1, i1 } zeroinitializer, i32 %select.received1, 0
  %1 = insertvalue { i32, i1, i1 } %0, i1 %select.recv.ok, 1
  %2 = insertvalue { i32, i1, i1 } %1, i1 true, 2
  ret { i32, i1, i1 } %2

select.next:                                      ; preds = %entry
  ret { i32, i1, i1 } zeroinitializer
}

declare { i1, i1 } @runtime.chanTryRecv(ptr dereferenceable_or_null(36), ptr, ptr) #0

; Function Attrs: nounwind
define hidden i1 @main.selectNonBlockingZeroSend(ptr dereferenceable_or_null(36) %ch, ptr %context) unnamed_addr #1 {
entry:
  %select.sent = call i1 @runtime.chanTrySend(ptr %ch, ptr null, ptr undef) #3
  br i1 %select.sent, label %select.body, label %select.next

select.body:                                      ; preds = %entry
  ret i1 true

select.next:                                      ; preds = %entry
  ret i1 false
}

; Function Attrs: nounwind
define hidden { i1, i1 } @main.selectNonBlockingZeroRecv(ptr dereferenceable_or_null(36) %ch, ptr %context) unnamed_addr #1 {
entry:
  %select.recv = call { i1, i1 } @runtime.chanTryRecv(ptr %ch, ptr null, ptr undef) #3
  %select.received = extractvalue { i1, i1 } %select.recv, 0
  br i1 %select.received, label %select.body, label %select.next

select.body:                                      ; preds = %entry
  %select.recv.ok = extractvalue { i1, i1 } %select.recv, 1
  %0 = insertvalue { i1, i1 } zeroinitializer, i1 %select.recv.ok, 0
  %1 = insertvalue { i1, i1 } %0, i1 true, 1
  ret { i1, i1 } %1

select.next:                                      ; preds = %entry
  ret { i1, i1 } zeroinitializer
}

; Function Attrs: nounwind
define hidden { i32, i1 } @main.selectBlocking(ptr dereferenceable_or_null(36) %ch1, ptr dereferenceable_or_null(36) %ch2, ptr %context) unnamed_addr #1 {
entry:
  %select.block.alloca = alloca [2 x %runtime.channelOp], align 8
  %select.states.alloca = alloca [2 x %runtime.chanSelectState], align 8
  %select.recvbuf.alloca = alloca [4 x i8], align 4
  %stackalloc = alloca i8, align 1
  call void @llvm.lifetime.start.p0(ptr nonnull %select.recvbuf.alloca)
  call void @llvm.lifetime.start.p0(ptr nonnull %select.states.alloca)
  store ptr %ch1, ptr %select.states.alloca, align 4
  %select.states.alloca.repack4 = getelementptr inbounds nuw i8, ptr %select.states.alloca, i32 4
  store ptr null, ptr %select.states.alloca.repack4, align 4
  %0 = getelementptr inbounds nuw i8, ptr %select.states.alloca, i32 8
  store ptr %ch2, ptr %0, align 4
  %.repack6 = getelementptr inbounds nuw i8, ptr %select.states.alloca, i32 12
  store ptr null, ptr %.repack6, align 4
  call void @llvm.lifetime.start.p0(ptr nonnull %select.block.alloca)
  %select.result = call { i32, i1 } @runtime.chanSelect(ptr nonnull %select.recvbuf.alloca, ptr nonnull %select.states.alloca, i32 2, i32 2, ptr nonnull %select.block.alloca, i32 2, i32 2, ptr undef) #3
  call void @llvm.lifetime.end.p0(ptr nonnull %select.block.alloca)
  call void @llvm.lifetime.end.p0(ptr nonnull %select.states.alloca)
  call void @runtime.trackPointer(ptr nonnull %select.recvbuf.alloca, ptr nonnull %stackalloc, ptr undef) #3
  %1 = extractvalue { i32, i1 } %select.result, 0
  %2 = icmp eq i32 %1, 0
  br i1 %2, label %select.body, label %select.next

select.body:                                      ; preds = %entry
  %select.received = load i32, ptr %select.recvbuf.alloca, align 4
  %3 = extractvalue { i32, i1 } %select.result, 1
  %4 = insertvalue { i32, i1 } zeroinitializer, i32 %select.received, 0
  %5 = insertvalue { i32, i1 } %4, i1 %3, 1
  ret { i32, i1 } %5

select.next:                                      ; preds = %entry
  %6 = icmp eq i32 %1, 1
  br i1 %6, label %select.body1, label %select.next2

select.body1:                                     ; preds = %select.next
  %select.received3 = load i32, ptr %select.recvbuf.alloca, align 4
  %7 = extractvalue { i32, i1 } %select.result, 1
  %8 = insertvalue { i32, i1 } zeroinitializer, i32 %select.received3, 0
  %9 = insertvalue { i32, i1 } %8, i1 %7, 1
  ret { i32, i1 } %9

select.next2:                                     ; preds = %select.next
  call void @runtime.trackPointer(ptr nonnull @"reflect/types.type:basic:string", ptr nonnull %stackalloc, ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull @"main$pack", ptr nonnull %stackalloc, ptr undef) #3
  call void @runtime._panic(ptr nonnull @"reflect/types.type:basic:string", ptr nonnull @"main$pack", ptr undef) #3
  unreachable
}

declare void @runtime._panic(ptr, ptr, ptr) #0

attributes #0 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { nocallback nofree nosync nounwind willreturn memory(argmem: readwrite) }
attributes #3 = { nounwind }
