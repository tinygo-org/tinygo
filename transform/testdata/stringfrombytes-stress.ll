declare { ptr, i64 } @runtime.stringFromBytes(ptr nocapture readonly, i64, i64)

declare i8 @readByte(ptr nocapture readonly) memory(argmem: read)
declare i8 @two(ptr nocapture readonly, ptr readonly) memory(argmem: read)
declare ptr @identity(ptr) memory(none)
declare void @unknown()
declare i8 @va(ptr nocapture readonly, ...) memory(argmem: read)
declare void @runtime.trackPointer(ptr, ptr, ptr)
declare i8 @writableParam(ptr nocapture) #0
declare i8 @readAndWrite(ptr nocapture readonly, ptr nocapture) #0
declare i1 @runtime.hashmapStringGet(ptr readonly nocapture, ptr readonly nocapture, i64, ptr nocapture, i64, ptr) #0
declare void @llvm.lifetime.start.p0(i64 immarg, ptr nocapture)
declare void @llvm.lifetime.end.p0(i64 immarg, ptr nocapture)
declare void @llvm.memcpy.p0.p0.i64(ptr noalias nocapture writeonly, ptr noalias nocapture readonly, i64, i1 immarg)

@g = global i8 0
@slot = global ptr null

define i8 @atomicRMW(ptr nocapture readonly %s) {
  %old = atomicrmw add ptr @g, i8 1 seq_cst
  %r = load i8, ptr %s
  ret i8 %r
}

define i8 @atomicCmpxchg(ptr nocapture readonly %s) {
  %old = cmpxchg ptr @g, i8 0, i8 1 seq_cst seq_cst
  %r = load i8, ptr %s
  ret i8 %r
}

define i8 @fenceCallee(ptr nocapture readonly %s) {
  fence seq_cst
  %r = load i8, ptr %s
  ret i8 %r
}

define i8 @writeGlobal(ptr nocapture readonly %s) {
  store i8 1, ptr @g
  %r = load i8, ptr %s
  ret i8 %r
}

define void @store1(ptr %x) {
  store i8 1, ptr %x
  ret void
}

define i8 @nest(ptr nocapture readonly %s, ptr nocapture %w) {
  call void @store1(ptr %w)
  %r = load i8, ptr %s
  ret i8 %r
}

define i8 @nestLocal(ptr nocapture readonly %s) {
  %l = alloca i8
  call void @store1(ptr %l)
  %r = load i8, ptr %s
  ret i8 %r
}

; The source array is written between the conversion and the use.

define i8 @keep_lifetimeEnd() {
  %a = alloca i8
  call void @llvm.lifetime.start.p0(i64 1, ptr %a)
  store i8 65, ptr %a
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %a, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  call void @llvm.lifetime.end.p0(i64 1, ptr %a)
  %r = call i8 @readByte(ptr %p)
  ret i8 %r
}

define i8 @keep_lifetimeStart() {
  %a = alloca i8
  store i8 65, ptr %a
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %a, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  call void @llvm.lifetime.start.p0(i64 1, ptr %a)
  %r = call i8 @readByte(ptr %p)
  ret i8 %r
}

define i8 @keep_srcFromCall() {
  %a = alloca i8
  store i8 65, ptr %a
  %src = call ptr @identity(ptr %a)
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  store i8 66, ptr %a
  %r = call i8 @readByte(ptr %p)
  ret i8 %r
}

define i8 @keep_srcFromLoad() {
  %a = alloca i8
  %pp = alloca ptr
  store i8 65, ptr %a
  store ptr %a, ptr %pp
  %src = load ptr, ptr %pp
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  store i8 66, ptr %a
  %r = call i8 @readByte(ptr %p)
  ret i8 %r
}

define i8 @keep_srcFromPhi(i1 %c) {
entry:
  %a = alloca i8
  %b = alloca i8
  store i8 65, ptr %a
  store i8 67, ptr %b
  br i1 %c, label %x, label %y
x:
  br label %j
y:
  br label %j
j:
  %src = phi ptr [ %a, %x ], [ %b, %y ]
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  store i8 66, ptr %b
  %r = call i8 @readByte(ptr %p)
  ret i8 %r
}

define i8 @keep_memcpyBetween() {
  %a = alloca i8
  %b = alloca i8
  store i8 65, ptr %a
  store i8 67, ptr %b
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %a, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  call void @llvm.memcpy.p0.p0.i64(ptr %a, ptr %b, i64 1, i1 false)
  %r = call i8 @readByte(ptr %p)
  ret i8 %r
}

define i8 @keep_atomicBetween() {
  %a = alloca i8
  store i8 65, ptr %a
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %a, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  %old = atomicrmw xchg ptr %a, i8 66 seq_cst
  %r = call i8 @readByte(ptr %p)
  ret i8 %r
}

define i8 @keep_callBetween(ptr %src) {
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  call void @unknown()
  %r = call i8 @readByte(ptr %p)
  ret i8 %r
}

define i8 @keep_storeMayAlias(ptr %src, ptr %q) {
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  store i8 66, ptr %q
  %r = call i8 @readByte(ptr %p)
  ret i8 %r
}

define i8 @keep_stringToWritableParam(ptr %src) {
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  %r = call i8 @writableParam(ptr %p)
  ret i8 %r
}

define i8 @keep_writableArgViaSelect(i1 %c) {
  %a = alloca i8
  %b = alloca i8
  store i8 65, ptr %a
  store i8 67, ptr %b
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %a, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  %out = select i1 %c, ptr %a, ptr %b
  %r = call i8 @readAndWrite(ptr %p, ptr %out)
  ret i8 %r
}

define i8 @keep_writableArgViaGEP() {
  %a = alloca [4 x i8]
  store i8 65, ptr %a
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %a, i64 4, i64 4)
  %p = extractvalue { ptr, i64 } %s, 0
  %out = getelementptr i8, ptr %a, i64 2
  %r = call i8 @readAndWrite(ptr %p, ptr %out)
  ret i8 %r
}

define i8 @keep_ptrTwice(ptr %src) {
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  %r = call i8 @two(ptr %p, ptr %p)
  ret i8 %r
}

define i8 @keep_secondConversion() {
  %a = alloca i8
  store i8 65, ptr %a
  %s1 = call { ptr, i64 } @runtime.stringFromBytes(ptr %a, i64 1, i64 1)
  %p1 = extractvalue { ptr, i64 } %s1, 0
  store i8 66, ptr %a
  %s2 = call { ptr, i64 } @runtime.stringFromBytes(ptr %a, i64 1, i64 1)
  %p2 = extractvalue { ptr, i64 } %s2, 0
  %r1 = call i8 @readByte(ptr %p1)
  %r2 = call i8 @readByte(ptr %p2)
  %r = add i8 %r1, %r2
  ret i8 %r
}

define i8 @keep_calleeWritesGlobal(ptr %src) {
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  %r = call i8 @writeGlobal(ptr %p)
  ret i8 %r
}

define i8 @keep_nestedWriteToArg() {
  %a = alloca i8
  store i8 65, ptr %a
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %a, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  %r = call i8 @nest(ptr %p, ptr %a)
  ret i8 %r
}

define i1 @keep_mapSlotIsSource(ptr %m) {
  %a = alloca [8 x i8]
  store i64 0, ptr %a
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %a, i64 8, i64 8)
  %p = extractvalue { ptr, i64 } %s, 0
  %ok = call i1 @runtime.hashmapStringGet(ptr %m, ptr %p, i64 8, ptr %a, i64 8, ptr undef)
  ret i1 %ok
}

define i1 @keep_mapKeyAliasesSlot(ptr %m) {
  %slot = alloca [8 x i8]
  store i64 0, ptr %slot
  %src = call ptr @identity(ptr %slot)
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 8, i64 8)
  %p = extractvalue { ptr, i64 } %s, 0
  %ok = call i1 @runtime.hashmapStringGet(ptr %m, ptr %p, i64 8, ptr %slot, i64 8, ptr undef)
  ret i1 %ok
}

; The copy can be removed.

define i8 @drop_storeOtherAlloca() {
  %a = alloca i8
  %b = alloca i8
  store i8 65, ptr %a
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %a, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  store i8 66, ptr %b
  %r = call i8 @readByte(ptr %p)
  ret i8 %r
}

define i8 @drop_lifetimeOther() {
  %a = alloca i8
  %b = alloca i8
  store i8 65, ptr %a
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %a, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  call void @llvm.lifetime.end.p0(i64 1, ptr %b)
  %r = call i8 @readByte(ptr %p)
  ret i8 %r
}

define i8 @drop_loadBetween(ptr %src, ptr %q) {
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  %x = load i8, ptr %q
  %r = call i8 @readByte(ptr %p)
  %y = add i8 %r, %x
  ret i8 %y
}

define i8 @drop_nestedLocal(ptr %src) {
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  %r = call i8 @nestLocal(ptr %p)
  ret i8 %r
}

define i1 @drop_mapUnknownSource(ptr %m, ptr %src) {
  %slot = alloca i64
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 4, i64 4)
  %p = extractvalue { ptr, i64 } %s, 0
  %ok = call i1 @runtime.hashmapStringGet(ptr %m, ptr %p, i64 4, ptr %slot, i64 8, ptr undef)
  ret i1 %ok
}

define i1 @drop_mapArraySource(ptr %m) {
  %arr = alloca [4 x i8]
  %slot = alloca i64
  store i32 0, ptr %arr
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %arr, i64 4, i64 4)
  %p = extractvalue { ptr, i64 } %s, 0
  %ok = call i1 @runtime.hashmapStringGet(ptr %m, ptr %p, i64 4, ptr %slot, i64 8, ptr undef)
  ret i1 %ok
}

define i1 @drop_mapPhiSourceOtherSlot(ptr %m, i1 %c) {
entry:
  %a = alloca [4 x i8]
  %b = alloca [4 x i8]
  %slot = alloca i64
  br i1 %c, label %x, label %y
x:
  br label %j
y:
  br label %j
j:
  %src = phi ptr [ %a, %x ], [ %b, %y ]
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 4, i64 4)
  %p = extractvalue { ptr, i64 } %s, 0
  %ok = call i1 @runtime.hashmapStringGet(ptr %m, ptr %p, i64 4, ptr %slot, i64 8, ptr undef)
  ret i1 %ok
}

define i1 @drop_mapSelectSlot(ptr %m, ptr %src, i1 %c) {
  %s1 = alloca i64
  %s2 = alloca i64
  %slot = select i1 %c, ptr %s1, ptr %s2
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 4, i64 4)
  %p = extractvalue { ptr, i64 } %s, 0
  %ok = call i1 @runtime.hashmapStringGet(ptr %m, ptr %p, i64 4, ptr %slot, i64 8, ptr undef)
  ret i1 %ok
}

define i8 @drop_trackPointerBetween(ptr %src) {
  %slot = alloca i8
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  call void @runtime.trackPointer(ptr %src, ptr %slot, ptr undef)
  %r = call i8 @readByte(ptr %p)
  ret i8 %r
}

define i8 @drop_trackPointerOfString(ptr %src) {
  %slot = alloca i8
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  call void @runtime.trackPointer(ptr %p, ptr %slot, ptr undef)
  %r = call i8 @readByte(ptr %p)
  ret i8 %r
}

define i8 @keep_trackPointerThenStore(ptr %src, ptr %q) {
  %slot = alloca i8
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  call void @runtime.trackPointer(ptr %src, ptr %slot, ptr undef)
  store i8 66, ptr %q
  %r = call i8 @readByte(ptr %p)
  ret i8 %r
}

attributes #0 = { "tinygo-writes-args-only" }

; Callees the pass cannot see into, or that write with atomic operations.

define i8 @keep_indirectCallee(ptr %src, ptr %fn) {
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  %r = call i8 %fn(ptr %p)
  ret i8 %r
}

define i8 @keep_calleeAtomicRMW() {
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr @g, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  %r = call i8 @atomicRMW(ptr %p)
  ret i8 %r
}

define i8 @keep_calleeAtomicCmpxchg() {
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr @g, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  %r = call i8 @atomicCmpxchg(ptr %p)
  ret i8 %r
}

define i8 @keep_calleeFence() {
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr @g, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  %r = call i8 @fenceCallee(ptr %p)
  ret i8 %r
}

; The string pointer is used outside a call, or in a variadic position.

define i8 @keep_storedPointer(ptr %src) {
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  store ptr %p, ptr @slot
  ret i8 0
}

define i8 @keep_variadicPosition(ptr %src, ptr %other) {
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  %r = call i8 (ptr, ...) @va(ptr %other, ptr %p)
  ret i8 %r
}

; The source is an alloca reached through a GEP, an escaped address or an
; integer round trip, and the alloca is written before the use.

define i8 @keep_gepSourceStoreBase() {
  %a = alloca [8 x i8]
  store i8 65, ptr %a
  %g = getelementptr i8, ptr %a, i64 0
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %g, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  store i8 66, ptr %a
  %r = call i8 @readByte(ptr %p)
  ret i8 %r
}

define i8 @keep_escapedViaGep() {
  %a = alloca [8 x i8]
  store i8 65, ptr %a
  %g = getelementptr i8, ptr %a, i64 0
  store ptr %g, ptr @slot
  %src = load ptr, ptr @slot
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  store i8 66, ptr %a
  %r = call i8 @readByte(ptr %p)
  ret i8 %r
}

define i8 @keep_escapedViaPtrToInt() {
  %a = alloca [8 x i8]
  store i8 65, ptr %a
  %i = ptrtoint ptr %a to i64
  %src = inttoptr i64 %i to ptr
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  store i8 66, ptr %a
  %r = call i8 @readByte(ptr %p)
  ret i8 %r
}

; The use is in another block, where a write can come first.

define i8 @keep_useInNextBlock(ptr %src, ptr %q) {
entry:
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  br label %next
next:
  store i8 66, ptr %q
  %r = call i8 @readByte(ptr %p)
  ret i8 %r
}

define i8 @keep_useOnOnePath(ptr %src, ptr %q, i1 %c) {
entry:
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  br i1 %c, label %write, label %read
write:
  store i8 66, ptr %q
  br label %read
read:
  %r = call i8 @readByte(ptr %p)
  ret i8 %r
}
