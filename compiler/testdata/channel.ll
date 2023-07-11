; ModuleID = 'channel.go'
source_filename = "channel.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime.channelBlockedList = type { ptr, ptr, ptr, { ptr, i32, i32 } }
%runtime.chanSelectState = type { ptr, ptr }

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #0

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #1

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #2 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @main.chanIntSend(ptr dereferenceable_or_null(32) %ch, ptr %context) unnamed_addr #2 {
entry:
  %chan.blockedList = alloca %runtime.channelBlockedList, align 8
  %chan.value = alloca i32, align 4
  call void @llvm.lifetime.start.p0(i64 4, ptr nonnull %chan.value)
  store i32 3, ptr %chan.value, align 4
  call void @llvm.lifetime.start.p0(i64 24, ptr nonnull %chan.blockedList)
  call void @runtime.chanSend(ptr %ch, ptr nonnull %chan.value, ptr nonnull %chan.blockedList, ptr undef) #4
  call void @llvm.lifetime.end.p0(i64 24, ptr nonnull %chan.blockedList)
  call void @llvm.lifetime.end.p0(i64 4, ptr nonnull %chan.value)
  ret void
}

; Function Attrs: argmemonly nocallback nofree nosync nounwind willreturn
declare void @llvm.lifetime.start.p0(i64 immarg, ptr nocapture) #3

declare void @runtime.chanSend(ptr dereferenceable_or_null(32), ptr, ptr dereferenceable_or_null(24), ptr) #1

; Function Attrs: argmemonly nocallback nofree nosync nounwind willreturn
declare void @llvm.lifetime.end.p0(i64 immarg, ptr nocapture) #3

; Function Attrs: nounwind
define hidden void @main.chanIntRecv(ptr dereferenceable_or_null(32) %ch, ptr %context) unnamed_addr #2 {
entry:
  %chan.blockedList = alloca %runtime.channelBlockedList, align 8
  %chan.value = alloca i32, align 4
  call void @llvm.lifetime.start.p0(i64 4, ptr nonnull %chan.value)
  call void @llvm.lifetime.start.p0(i64 24, ptr nonnull %chan.blockedList)
  %0 = call i1 @runtime.chanRecv(ptr %ch, ptr nonnull %chan.value, ptr nonnull %chan.blockedList, ptr undef) #4
  call void @llvm.lifetime.end.p0(i64 4, ptr nonnull %chan.value)
  call void @llvm.lifetime.end.p0(i64 24, ptr nonnull %chan.blockedList)
  ret void
}

declare i1 @runtime.chanRecv(ptr dereferenceable_or_null(32), ptr, ptr dereferenceable_or_null(24), ptr) #1

; Function Attrs: nounwind
define hidden void @main.chanZeroSend(ptr dereferenceable_or_null(32) %ch, ptr %context) unnamed_addr #2 {
entry:
  %complit = alloca {}, align 8
  %chan.blockedList = alloca %runtime.channelBlockedList, align 8
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr nonnull %complit, ptr nonnull %stackalloc, ptr undef) #4
  call void @llvm.lifetime.start.p0(i64 24, ptr nonnull %chan.blockedList)
  call void @runtime.chanSend(ptr %ch, ptr null, ptr nonnull %chan.blockedList, ptr undef) #4
  call void @llvm.lifetime.end.p0(i64 24, ptr nonnull %chan.blockedList)
  ret void
}

; Function Attrs: nounwind
define hidden void @main.chanZeroRecv(ptr dereferenceable_or_null(32) %ch, ptr %context) unnamed_addr #2 {
entry:
  %chan.blockedList = alloca %runtime.channelBlockedList, align 8
  call void @llvm.lifetime.start.p0(i64 24, ptr nonnull %chan.blockedList)
  %0 = call i1 @runtime.chanRecv(ptr %ch, ptr null, ptr nonnull %chan.blockedList, ptr undef) #4
  call void @llvm.lifetime.end.p0(i64 24, ptr nonnull %chan.blockedList)
  ret void
}

; Function Attrs: nounwind
define hidden void @main.selectZeroRecv(ptr dereferenceable_or_null(32) %ch1, ptr dereferenceable_or_null(32) %ch2, ptr %context) unnamed_addr #2 {
entry:
  %select.states.alloca = alloca [2 x %runtime.chanSelectState], align 8
  %select.send.value = alloca i32, align 4
  store i32 1, ptr %select.send.value, align 4
  call void @llvm.lifetime.start.p0(i64 16, ptr nonnull %select.states.alloca)
  store ptr %ch1, ptr %select.states.alloca, align 8
  %select.states.alloca.repack1 = getelementptr inbounds %runtime.chanSelectState, ptr %select.states.alloca, i32 0, i32 1
  store ptr %select.send.value, ptr %select.states.alloca.repack1, align 4
  %0 = getelementptr inbounds [2 x %runtime.chanSelectState], ptr %select.states.alloca, i32 0, i32 1
  store ptr %ch2, ptr %0, align 8
  %.repack3 = getelementptr inbounds [2 x %runtime.chanSelectState], ptr %select.states.alloca, i32 0, i32 1, i32 1
  store ptr null, ptr %.repack3, align 4
  %select.result = call { i32, i1 } @runtime.tryChanSelect(ptr undef, ptr nonnull %select.states.alloca, i32 2, i32 2, ptr undef) #4
  call void @llvm.lifetime.end.p0(i64 16, ptr nonnull %select.states.alloca)
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

declare { i32, i1 } @runtime.tryChanSelect(ptr, ptr, i32, i32, ptr) #1

attributes #0 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #1 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #2 = { nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #3 = { argmemonly nocallback nofree nosync nounwind willreturn }
attributes #4 = { nounwind }
