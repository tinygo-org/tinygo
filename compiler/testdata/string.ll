; ModuleID = 'string.go'
source_filename = "string.go"
target datalayout = "e-m:e-p:32:32-p270:32:32-p271:32:32-p272:64:64-f64:32:64-f80:32-n8:16:32-S128"
target triple = "i686--linux"

declare noalias nonnull i8* @runtime.alloc(i32, i8*, i8*)

define hidden void @main.init(i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret void
}

define hidden i32 @main.stringLen(i8* %s.data, i32 %s.len, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret i32 %s.len
}

define hidden i8 @main.stringIndex(i8* %s.data, i32 %s.len, i32 %index, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %.not = icmp ult i32 %index, %s.len
  br i1 %.not, label %lookup.next, label %lookup.throw

lookup.throw:                                     ; preds = %entry
  call void @runtime.lookupPanic(i8* undef, i8* null)
  unreachable

lookup.next:                                      ; preds = %entry
  %0 = getelementptr inbounds i8, i8* %s.data, i32 %index
  %1 = load i8, i8* %0, align 1
  ret i8 %1
}

declare void @runtime.lookupPanic(i8*, i8*)

define hidden i1 @main.stringCompareEqual(i8* %s1.data, i32 %s1.len, i8* %s2.data, i32 %s2.len, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %0 = call i1 @runtime.stringEqual(i8* %s1.data, i32 %s1.len, i8* %s2.data, i32 %s2.len, i8* undef, i8* null)
  ret i1 %0
}

declare i1 @runtime.stringEqual(i8*, i32, i8*, i32, i8*, i8*)

define hidden i1 @main.stringCompareUnequal(i8* %s1.data, i32 %s1.len, i8* %s2.data, i32 %s2.len, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %0 = call i1 @runtime.stringEqual(i8* %s1.data, i32 %s1.len, i8* %s2.data, i32 %s2.len, i8* undef, i8* null)
  %1 = xor i1 %0, true
  ret i1 %1
}

define hidden i1 @main.stringCompareLarger(i8* %s1.data, i32 %s1.len, i8* %s2.data, i32 %s2.len, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %0 = call i1 @runtime.stringLess(i8* %s1.data, i32 %s1.len, i8* %s2.data, i32 %s2.len, i8* undef, i8* null)
  %1 = xor i1 %0, true
  ret i1 %1
}

declare i1 @runtime.stringLess(i8*, i32, i8*, i32, i8*, i8*)
