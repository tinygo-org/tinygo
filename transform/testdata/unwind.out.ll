target datalayout = "e-p:32:32"
target triple = "wasm32-unknown-unknown"

@runtime.unwindPendingSignal = internal unnamed_addr global i1 false
@value = local_unnamed_addr global i32 0

declare void @external() local_unnamed_addr

; Function Attrs: mustprogress nofree norecurse nounwind willreturn memory(readwrite, argmem: none, inaccessiblemem: write)
define noundef i1 @checkSafe() local_unnamed_addr #0 {
entry:
  %unwind.entry = load i1, ptr @runtime.unwindPendingSignal, align 1
  %0 = xor i1 %unwind.entry, true
  tail call void @llvm.assume(i1 %0)
  store i32 1, ptr @value, align 4
  ret i1 false
}

; Function Attrs: mustprogress nofree norecurse nounwind willreturn memory(readwrite, argmem: none, inaccessiblemem: write)
define noundef i1 @checkPanic() local_unnamed_addr #0 {
entry:
  %unwind.entry = load i1, ptr @runtime.unwindPendingSignal, align 1
  %0 = xor i1 %unwind.entry, true
  tail call void @llvm.assume(i1 %0)
  store i1 true, ptr @runtime.unwindPendingSignal, align 1
  ret i1 true
}

define i1 @checkExternal() local_unnamed_addr {
entry:
  %unwind.entry = load i1, ptr @runtime.unwindPendingSignal, align 1
  %0 = xor i1 %unwind.entry, true
  tail call void @llvm.assume(i1 %0)
  tail call void @external()
  %unwind.i = load i1, ptr @runtime.unwindPendingSignal, align 1
  ret i1 %unwind.i
}

define i1 @checkIndirect(ptr nocapture readonly %fn) local_unnamed_addr {
entry:
  %unwind.entry = load i1, ptr @runtime.unwindPendingSignal, align 1
  %0 = xor i1 %unwind.entry, true
  tail call void @llvm.assume(i1 %0)
  tail call void %fn()
  %unwind.i = load i1, ptr @runtime.unwindPendingSignal, align 1
  ret i1 %unwind.i
}

; Function Attrs: mustprogress nofree norecurse nounwind willreturn memory(read, argmem: none, inaccessiblemem: none)
define i1 @runtime.unwindPending() local_unnamed_addr #1 {
entry:
  %unwind = load i1, ptr @runtime.unwindPendingSignal, align 1
  ret i1 %unwind
}

; Function Attrs: mustprogress nocallback nofree nounwind willreturn memory(inaccessiblemem: write)
declare void @llvm.assume(i1 noundef) #2

attributes #0 = { mustprogress nofree norecurse nounwind willreturn memory(readwrite, argmem: none, inaccessiblemem: write) }
attributes #1 = { mustprogress nofree norecurse nounwind willreturn memory(read, argmem: none, inaccessiblemem: none) }
attributes #2 = { mustprogress nocallback nofree nounwind willreturn memory(inaccessiblemem: write) }
