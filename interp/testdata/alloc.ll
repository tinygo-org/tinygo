target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32--wasi"

@"runtime/gc.layout:62-2000000000000001" = linkonce_odr unnamed_addr constant { i32, [8 x i8] } { i32 62, [8 x i8] c" \00\00\00\00\00\00\01" }
@pointerFree12 = global i8* null
@pointerFree7 = global i8* null
@pointerFree3 = global i8* null
@pointerFree0 = global i8* null
@layout1 = global i8* null
@layout2 = global i8* null
@layout3 = global i8* null
@layout4 = global i8* null
@bigobj1 = global i8* null

declare i8* @runtime.alloc(i32, i8*) unnamed_addr

define void @runtime.initAll() unnamed_addr {
  call void @main.init()
  ret void
}

define internal void @main.init() unnamed_addr {
  ; Object that's word-aligned.
  %pointerFree12 = call i8* @runtime.alloc(i32 12, i8* inttoptr (i32 3 to i8*))
  store i8* %pointerFree12, i8** @pointerFree12
  ; Object larger than a word but not word-aligned.
  %pointerFree7 = call i8* @runtime.alloc(i32 7, i8* inttoptr (i32 3 to i8*))
  store i8* %pointerFree7, i8** @pointerFree7
  ; Object smaller than a word (and of course not word-aligned).
  %pointerFree3 = call i8* @runtime.alloc(i32 3, i8* inttoptr (i32 3 to i8*))
  store i8* %pointerFree3, i8** @pointerFree3
  ; Zero-sized object.
  %pointerFree0 = call i8* @runtime.alloc(i32 0, i8* inttoptr (i32 3 to i8*))
  store i8* %pointerFree0, i8** @pointerFree0

  ; Object made out of 3 pointers.
  %layout1 = call i8* @runtime.alloc(i32 12, i8* inttoptr (i32 67 to i8*))
  store i8* %layout1, i8** @layout1
  ; Array (or slice) of 5 slices.
  %layout2 = call i8* @runtime.alloc(i32 60, i8* inttoptr (i32 71 to i8*))
  store i8* %layout2, i8** @layout2
  ; Oddly shaped object, using all bits in the layout integer.
  %layout3 = call i8* @runtime.alloc(i32 104, i8* inttoptr (i32 2467830261 to i8*))
  store i8* %layout3, i8** @layout3
  ; ...repeated.
  %layout4 = call i8* @runtime.alloc(i32 312, i8* inttoptr (i32 2467830261 to i8*))
  store i8* %layout4, i8** @layout4

  ; Large object that needs to be stored in a separate global.
  %bigobj1 = call i8* @runtime.alloc(i32 248, i8* bitcast ({ i32, [8 x i8] }* @"runtime/gc.layout:62-2000000000000001" to i8*))
  store i8* %bigobj1, i8** @bigobj1
  ret void
}