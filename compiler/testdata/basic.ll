target datalayout = "e-m:e-p:32:32-Fi8-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

%runtime.channel = type { i32, i32, i8, %runtime.channelBlockedList*, i32, i32, i32, i8* }
%runtime.channelBlockedList = type { %runtime.channelBlockedList*, %"internal/task.Task"*, %runtime.chanSelectState*, { %runtime.channelBlockedList*, i32, i32 } }
%"internal/task.Task" = type opaque
%runtime.chanSelectState = type { %runtime.channel*, i8* }
%runtime._string = type { i8*, i32 }

@"main.stringEqual$string" = internal unnamed_addr constant [1 x i8] c"s"

define internal i32 @main.add(i32 %x, i32 %y, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %0 = add i32 %x, %y
  ret i32 %0
}

define internal void @main.closeChan(%runtime.channel* dereferenceable_or_null(32) %ch, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  call void @runtime.chanClose(%runtime.channel* %ch, i8* undef, i8* null)
  ret void
}

declare void @runtime.chanClose(%runtime.channel* dereferenceable_or_null(32), i8*, i8*)

define internal void @main.init(i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret void
}

define internal i1 @main.stringEqual(i8* %s.data, i32 %s.len, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %0 = insertvalue %runtime._string zeroinitializer, i8* %s.data, 0
  %1 = insertvalue %runtime._string %0, i32 %s.len, 1
  %2 = extractvalue %runtime._string %1, 0
  %3 = extractvalue %runtime._string %1, 1
  %4 = call i1 @runtime.stringEqual(i8* %2, i32 %3, i8* getelementptr inbounds ([1 x i8], [1 x i8]* @"main.stringEqual$string", i32 0, i32 0), i32 1, i8* undef, i8* null)
  ret i1 %4
}

declare i1 @runtime.stringEqual(i8*, i32, i8*, i32, i8*, i8*)
