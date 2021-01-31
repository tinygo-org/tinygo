; ModuleID = 'slice.go'
source_filename = "slice.go"
target datalayout = "e-m:e-p:32:32-p270:32:32-p271:32:32-p272:64:64-f64:32:64-f80:32-n8:16:32-S128"
target triple = "i686--linux"

declare noalias nonnull i8* @runtime.alloc(i32, i8*, i8*)

define hidden void @main.init(i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret void
}

define hidden i32 @main.sliceLen(i32* %ints.data, i32 %ints.len, i32 %ints.cap, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret i32 %ints.len
}

define hidden i32 @main.sliceCap(i32* %ints.data, i32 %ints.len, i32 %ints.cap, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret i32 %ints.cap
}
