; ModuleID = 'channel.go'
source_filename = "channel.go"
target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-wasi"

%runtime.channel = type { i32, i32, i8, %runtime.channelBlockedList*, i32, i32, i32, i8* }
%runtime.channelBlockedList = type { %runtime.channelBlockedList*, %"internal/task.Task"*, %runtime.chanSelectState*, { %runtime.channelBlockedList*, i32, i32 } }
%"internal/task.Task" = type { %"internal/task.Task"*, i8*, i64, %"internal/task.state" }
%"internal/task.state" = type { i8* }
%runtime.chanSelectState = type { %runtime.channel*, i8* }

declare noalias nonnull i8* @runtime.alloc(i32, i8*, i8*, i8*)

; Function Attrs: nounwind
define hidden void @main.init(i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @main.chanIntSend(%runtime.channel* dereferenceable_or_null(32) %ch, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %chan.blockedList = alloca %runtime.channelBlockedList, align 8
  %chan.value = alloca i32, align 4
  %chan.value.bitcast = bitcast i32* %chan.value to i8*
  call void @llvm.lifetime.start.p0i8(i64 4, i8* nonnull %chan.value.bitcast)
  store i32 3, i32* %chan.value, align 4
  %chan.blockedList.bitcast = bitcast %runtime.channelBlockedList* %chan.blockedList to i8*
  call void @llvm.lifetime.start.p0i8(i64 24, i8* nonnull %chan.blockedList.bitcast)
  call void @runtime.chanSend(%runtime.channel* %ch, i8* nonnull %chan.value.bitcast, %runtime.channelBlockedList* nonnull %chan.blockedList, i8* undef, i8* null) #0
  call void @llvm.lifetime.end.p0i8(i64 24, i8* nonnull %chan.blockedList.bitcast)
  call void @llvm.lifetime.end.p0i8(i64 4, i8* nonnull %chan.value.bitcast)
  ret void
}

; Function Attrs: argmemonly nounwind willreturn
declare void @llvm.lifetime.start.p0i8(i64 immarg, i8* nocapture) #1

declare void @runtime.chanSend(%runtime.channel* dereferenceable_or_null(32), i8*, %runtime.channelBlockedList* dereferenceable_or_null(24), i8*, i8*)

; Function Attrs: argmemonly nounwind willreturn
declare void @llvm.lifetime.end.p0i8(i64 immarg, i8* nocapture) #1

; Function Attrs: nounwind
define hidden void @main.chanIntRecv(%runtime.channel* dereferenceable_or_null(32) %ch, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %chan.blockedList = alloca %runtime.channelBlockedList, align 8
  %chan.value = alloca i32, align 4
  %chan.value.bitcast = bitcast i32* %chan.value to i8*
  call void @llvm.lifetime.start.p0i8(i64 4, i8* nonnull %chan.value.bitcast)
  %chan.blockedList.bitcast = bitcast %runtime.channelBlockedList* %chan.blockedList to i8*
  call void @llvm.lifetime.start.p0i8(i64 24, i8* nonnull %chan.blockedList.bitcast)
  %0 = call i1 @runtime.chanRecv(%runtime.channel* %ch, i8* nonnull %chan.value.bitcast, %runtime.channelBlockedList* nonnull %chan.blockedList, i8* undef, i8* null) #0
  call void @llvm.lifetime.end.p0i8(i64 4, i8* nonnull %chan.value.bitcast)
  call void @llvm.lifetime.end.p0i8(i64 24, i8* nonnull %chan.blockedList.bitcast)
  ret void
}

declare i1 @runtime.chanRecv(%runtime.channel* dereferenceable_or_null(32), i8*, %runtime.channelBlockedList* dereferenceable_or_null(24), i8*, i8*)

; Function Attrs: nounwind
define hidden void @main.chanZeroSend(%runtime.channel* dereferenceable_or_null(32) %ch, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %chan.blockedList = alloca %runtime.channelBlockedList, align 8
  %chan.blockedList.bitcast = bitcast %runtime.channelBlockedList* %chan.blockedList to i8*
  call void @llvm.lifetime.start.p0i8(i64 24, i8* nonnull %chan.blockedList.bitcast)
  call void @runtime.chanSend(%runtime.channel* %ch, i8* null, %runtime.channelBlockedList* nonnull %chan.blockedList, i8* undef, i8* null) #0
  call void @llvm.lifetime.end.p0i8(i64 24, i8* nonnull %chan.blockedList.bitcast)
  ret void
}

; Function Attrs: nounwind
define hidden void @main.chanZeroRecv(%runtime.channel* dereferenceable_or_null(32) %ch, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %chan.blockedList = alloca %runtime.channelBlockedList, align 8
  %chan.blockedList.bitcast = bitcast %runtime.channelBlockedList* %chan.blockedList to i8*
  call void @llvm.lifetime.start.p0i8(i64 24, i8* nonnull %chan.blockedList.bitcast)
  %0 = call i1 @runtime.chanRecv(%runtime.channel* %ch, i8* null, %runtime.channelBlockedList* nonnull %chan.blockedList, i8* undef, i8* null) #0
  call void @llvm.lifetime.end.p0i8(i64 24, i8* nonnull %chan.blockedList.bitcast)
  ret void
}

; Function Attrs: nounwind
define hidden void @main.selectZeroRecv(%runtime.channel* dereferenceable_or_null(32) %ch1, %runtime.channel* dereferenceable_or_null(32) %ch2, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %select.states.alloca = alloca [2 x %runtime.chanSelectState], align 8
  %select.send.value = alloca i32, align 4
  store i32 1, i32* %select.send.value, align 4
  %select.states.alloca.bitcast = bitcast [2 x %runtime.chanSelectState]* %select.states.alloca to i8*
  call void @llvm.lifetime.start.p0i8(i64 16, i8* nonnull %select.states.alloca.bitcast)
  %.repack = getelementptr inbounds [2 x %runtime.chanSelectState], [2 x %runtime.chanSelectState]* %select.states.alloca, i32 0, i32 0, i32 0
  store %runtime.channel* %ch1, %runtime.channel** %.repack, align 8
  %.repack1 = getelementptr inbounds [2 x %runtime.chanSelectState], [2 x %runtime.chanSelectState]* %select.states.alloca, i32 0, i32 0, i32 1
  %0 = bitcast i8** %.repack1 to i32**
  store i32* %select.send.value, i32** %0, align 4
  %.repack3 = getelementptr inbounds [2 x %runtime.chanSelectState], [2 x %runtime.chanSelectState]* %select.states.alloca, i32 0, i32 1, i32 0
  store %runtime.channel* %ch2, %runtime.channel** %.repack3, align 8
  %.repack4 = getelementptr inbounds [2 x %runtime.chanSelectState], [2 x %runtime.chanSelectState]* %select.states.alloca, i32 0, i32 1, i32 1
  store i8* null, i8** %.repack4, align 4
  %select.states = getelementptr inbounds [2 x %runtime.chanSelectState], [2 x %runtime.chanSelectState]* %select.states.alloca, i32 0, i32 0
  %select.result = call { i32, i1 } @runtime.tryChanSelect(i8* undef, %runtime.chanSelectState* nonnull %select.states, i32 2, i32 2, i8* undef, i8* null) #0
  call void @llvm.lifetime.end.p0i8(i64 16, i8* nonnull %select.states.alloca.bitcast)
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

declare { i32, i1 } @runtime.tryChanSelect(i8*, %runtime.chanSelectState*, i32, i32, i8*, i8*)

attributes #0 = { nounwind }
attributes #1 = { argmemonly nounwind willreturn }
