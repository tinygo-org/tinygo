target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

@tbl = external global [16 x i32], align 4

declare void @runtime.slicePanic()

define void @testVoidRet() {
  ret void
}

define {i32*, i32, i32} @testAgg({i32*, i32, i32} %slice) {
  %slicezero = insertvalue {i32*, i32, i32} %slice, i32 0, 1
  ret {i32*, i32, i32} %slicezero
}

define i32 @testLoop(i32 %x, i32 %y) {
entry:
  br label %loopHeader
loopHeader:
  %i = phi i32 [ 0, %entry ], [ %next, %loopBody ]
  %v = phi i32 [ 0, %entry ], [ %sum, %loopBody ]
  %ok = icmp ult i32 %i, %x
  br i1 %ok, label %loopBody, label %exit
loopBody:
  %sum = add i32 %v, %y
  %next = add i32 %i, 1
  br label %loopHeader
exit:
  ret i32 %v
}

define {i32*, i32, i32} @testPtrMath({i32*, i32, i32} %s, i32 %x) {
entry:
  %cap = extractvalue {i32*, i32, i32} %s, 2
  %ok = icmp ult i32 %x, %cap
  br i1 %ok, label %slice, label %panic
slice:
  %base = extractvalue {i32*, i32, i32} %s, 0
  %ptr = getelementptr i32, i32* %base, i32 %x
  %slice.0 = insertvalue {i32*, i32, i32} %s, i32* %ptr, 0
  %len = extractvalue {i32*, i32, i32} %s, 1
  %newLen = sub i32 %len, %x
  %slice.1 = insertvalue {i32*, i32, i32} %slice.0, i32 %newLen, 1
  %newCap = sub i32 %cap, %x
  %slice.2 = insertvalue {i32*, i32, i32} %slice.1, i32 %newCap, 2
  ret {i32*, i32, i32} %slice.2
panic:
  call void @runtime.slicePanic()
  unreachable
}

define i32 @testLookup(i32 %x) {
entry:
  switch i32 %x, label %def [ i32 0, label %lookup.0
                              i32 1, label %lookup.1
                              i32 5, label %lookup.5 ]
lookup.0:
  br label %exit
lookup.1:
  br label %exit
lookup.5:
  br label %exit
def:
  %v = mul i32 %x, 7
  br label %exit
exit:
  %p = phi i32 [ 6, %lookup.0 ], [ 1, %lookup.1 ], [ 7, %lookup.5 ], [ %v, %def ]
  %res = mul i32 %p, %p
  ret i32 %res
}

define i32 @testCall({i32*, i32, i32} %s) {
entry:
  %ptr = extractvalue {i32*, i32, i32} %s, 0
  %len = extractvalue {i32*, i32, i32} %s, 1
  switch i32 %len, label %def [   i32 0, label %zero
                                  i32 1, label %one ]
zero:
  ret i32 0
one:
  %v = load i32, i32* %ptr
  ret i32 %v
def:
  %split = udiv i32 %len, 2
  %leftLen = sub i32 %len, %split
  %left = insertvalue {i32*, i32, i32} %s, i32 %leftLen, 1
  %x = call i32 @testCall({i32*, i32, i32} %left)
  %right = call {i32*, i32, i32} @testPtrMath({i32*, i32, i32} %s, i32 %split)
  %y = call i32 @testCall({i32*, i32, i32} %right)
  %sum = add i32 %x, %y
  ret i32 %sum
}

define i32 @memLookup(i8 %idx) {
entry:
  %ok = icmp ult i8 %idx, 16
  br i1 %ok, label %lookup, label %else
lookup:
  %ptr = getelementptr inbounds [16 x i32], [16 x i32]* @tbl, i32 0, i8 %idx
  %v = load i32, i32* %ptr
  ret i32 %v
else:
  ret i32 -1
}

define void @spinWait(i32* %on) {
entry:
  br label %wait
wait:
  %load = load atomic i32, i32* %on seq_cst, align 4
  %ok = trunc i32 %load to i1
  br i1 %ok, label %exit, label %wait
exit:
  ret void
}

define void @spinNotify(i32* %on) {
  store atomic i32 1, i32* %on seq_cst, align 4
  ret void
}

define void @alloca() {
  %mem = alloca i32, i32 2
  ret void
}

define void @binIntOps(i32 %x, i32 %y) {
  %add = add i32 %x, %y
  %sub = sub i32 %x, %y
  %mul = mul i32 %x, %y
  %udiv = udiv i32 %x, %y
  %sdiv = sdiv i32 %x, %y
  %urem = urem i32 %x, %y
  %srem = srem i32 %x, %y
  %shl = shl i32 %x, %y
  %lshr = lshr i32 %x, %y
  %ashr = ashr i32 %x, %y
  %and = and i32 %x, %y
  %or = or i32 %x, %y
  %xor = xor i32 %x, %y
  ret void
}
