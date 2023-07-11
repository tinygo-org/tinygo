target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32--wasi"

@"runtime/gc.layout:62-2000000000000001" = linkonce_odr unnamed_addr constant { i32, [8 x i8] } { i32 62, [8 x i8] c"\01\00\00\00\00\00\00 " }
@pointerFree12 = global ptr null
@pointerFree7 = global ptr null
@pointerFree3 = global ptr null
@pointerFree0 = global ptr null
@layout1 = global ptr null
@layout2 = global ptr null
@layout3 = global ptr null
@layout4 = global ptr null
@bigobj1 = global ptr null

declare ptr @runtime.alloc(i32, ptr) unnamed_addr

define void @runtime.initAll() unnamed_addr {
  call void @main.init()
  ret void
}

define internal void @main.init() unnamed_addr {
  ; Object that's word-aligned.
  %pointerFree12 = call ptr @runtime.alloc(i32 12, ptr inttoptr (i32 3 to ptr))
  store ptr %pointerFree12, ptr @pointerFree12
  ; Object larger than a word but not word-aligned.
  %pointerFree7 = call ptr @runtime.alloc(i32 7, ptr inttoptr (i32 3 to ptr))
  store ptr %pointerFree7, ptr @pointerFree7
  ; Object smaller than a word (and of course not word-aligned).
  %pointerFree3 = call ptr @runtime.alloc(i32 3, ptr inttoptr (i32 3 to ptr))
  store ptr %pointerFree3, ptr @pointerFree3
  ; Zero-sized object.
  %pointerFree0 = call ptr @runtime.alloc(i32 0, ptr inttoptr (i32 3 to ptr))
  store ptr %pointerFree0, ptr @pointerFree0

  ; Object made out of 3 pointers.
  %layout1 = call ptr @runtime.alloc(i32 12, ptr inttoptr (i32 67 to ptr))
  store ptr %layout1, ptr @layout1
  ; Array (or slice) of 5 slices.
  %layout2 = call ptr @runtime.alloc(i32 60, ptr inttoptr (i32 71 to ptr))
  store ptr %layout2, ptr @layout2
  ; Oddly shaped object, using all bits in the layout integer.
  %layout3 = call ptr @runtime.alloc(i32 104, ptr inttoptr (i32 2467830261 to ptr))
  store ptr %layout3, ptr @layout3
  ; ...repeated.
  %layout4 = call ptr @runtime.alloc(i32 312, ptr inttoptr (i32 2467830261 to ptr))
  store ptr %layout4, ptr @layout4

  ; Large object that needs to be stored in a separate global.
  %bigobj1 = call ptr @runtime.alloc(i32 248, ptr @"runtime/gc.layout:62-2000000000000001")
  store ptr %bigobj1, ptr @bigobj1
  ret void
}
