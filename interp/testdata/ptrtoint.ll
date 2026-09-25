target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.v1 = global i32 0
@main.v2 = global i64 0
@main.v3 = global i8 0
@main.global = global i8 0

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init()
  ret void
}

define internal void @main.init() unnamed_addr {
entry:
  ; Narrow the receiver from a pointer, like an interface method thunk.
  %v1 = call i32 @main.narrow.id.unpack(ptr inttoptr (i8 1 to ptr))
  store i32 %v1, ptr @main.v1

  ; Widen a narrow integer through a pointer.
  %ptr = inttoptr i16 -1 to ptr
  %v2 = ptrtoint ptr %ptr to i64
  store i64 %v2, ptr @main.v2

  ; A real pointer can't be truncated at compile time.
  %v3 = ptrtoint ptr @main.global to i8
  store i8 %v3, ptr @main.v3
  ret void
}

define internal i32 @main.narrow.id.unpack(ptr %receiver) unnamed_addr {
entry:
  %unpack.int = ptrtoint ptr %receiver to i8
  %ret = call i32 @main.narrow.id(i8 %unpack.int)
  ret i32 %ret
}

define internal i32 @main.narrow.id(i8 %h) unnamed_addr {
entry:
  %cmp = icmp eq i8 %h, 1
  br i1 %cmp, label %if.then, label %if.done

if.then:
  ret i32 10

if.done:
  ret i32 0
}
