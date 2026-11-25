declare void @runtime.lookupPanic()

define i8 @main.get(ptr %p, i64 %i, i64 %n) {
entry:
  %oob = icmp uge i64 %i, %n
  br i1 %oob, label %lookup.throw, label %lookup.next

lookup.next:
  %a = getelementptr i8, ptr %p, i64 %i
  %v = load i8, ptr %a
  ret i8 %v

lookup.throw:
  call void @runtime.lookupPanic()
  unreachable
}

define i8 @main.other(ptr %p) {
  %v = load i8, ptr %p
  ret i8 %v
}
