target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.phiNodesResultA = global i8 0
@main.phiNodesResultB = global i8 0

define void @runtime.initAll() {
  call void @main.init()
  ret void
}

; PHI nodes always use the value from the previous block, even in a loop. This
; means that the loop below should swap the values %a and %b on each iteration.
; Previously there was a bug which resulted in %b getting the value 3 on the
; second iteration while it should have gotten 1 (from the first iteration of
; %for.loop).
define internal void @main.init() {
entry:
  br label %for.loop

for.loop:
  %a = phi i8 [ 1, %entry ], [ %b, %for.loop ]
  %b = phi i8 [ 3, %entry ], [ %a, %for.loop ]
  %icmp = icmp eq i8 %a, 3
  br i1 %icmp, label %for.done, label %for.loop

for.done:
  store i8 %a, ptr @main.phiNodesResultA
  store i8 %b, ptr @main.phiNodesResultB
  ret void
}
