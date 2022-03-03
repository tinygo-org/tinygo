target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@extern = external global i8

define void @runtime.initAll(i8* %context) unnamed_addr {
  ; If all uses of the alloca run at runtime it should not be created at runtime.
  %a = call i8 @lookup(i8 1)

  ; If the alloca is used at runtime, it should be created and the runtime instructions should be wrapped by lifetime intrinsics.
  %x = load i8, i8* @extern
  %b = call i8 @lookup(i8 %x)
  store i8 %b, i8* @extern

  ret void
}

define i8 @lookup(i8 %x) {
  %tbl = alloca i8, i32 4
  store i8 1, i8* %tbl
  %tbl.1 = getelementptr i8, i8* %tbl, i8 1
  store i8 2, i8* %tbl.1
  %tbl.2 = getelementptr i8, i8* %tbl, i8 2
  store i8 3, i8* %tbl.2
  %tbl.3 = getelementptr i8, i8* %tbl, i8 3
  store i8 4, i8* %tbl.3
  %elem = getelementptr i8, i8* %tbl, i8 %x
  %v = load i8, i8* %elem
  ret i8 %v
}
