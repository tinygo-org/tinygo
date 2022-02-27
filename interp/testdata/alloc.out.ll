; ModuleID = 'testdata/alloc.ll'
source_filename = "testdata/alloc.ll"
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
@interp.alloc.10 = private global [12 x i8] zeroinitializer, align 8
@interp.alloc.11 = private global [7 x i8] zeroinitializer, align 4
@interp.alloc.12 = private global [3 x i8] zeroinitializer, align 2
@interp.alloc.13 = private global [0 x i8] zeroinitializer
@interp.alloc.14 = private global [3 x i8*] zeroinitializer, align 8
@interp.alloc.15 = private global [5 x { i8*, [2 x i32] }] zeroinitializer, align 16
@interp.alloc.16 = private global { [3 x i8*], [2 x i32], [2 x i8*], [6 x i32], [2 x i8*], [3 x i32], [2 x i8*], [2 x i32], i8*, [2 x i32], i8* } zeroinitializer, align 16
@interp.alloc.17 = private global [3 x { [3 x i8*], [2 x i32], [2 x i8*], [6 x i32], [2 x i8*], [3 x i32], [2 x i8*], [2 x i32], i8*, [2 x i32], i8* }] zeroinitializer, align 16
@interp.alloc.18 = private global { i8*, [60 x i32], i8* } zeroinitializer, align 16

declare i8* @runtime.alloc(i32, i8*) unnamed_addr

define void @runtime.initAll() unnamed_addr {
interpreted:
  store i32 ptrtoint ([12 x i8]* @interp.alloc.10 to i32), i32* bitcast (i8** @pointerFree12 to i32*), align 4
  store i32 ptrtoint ([7 x i8]* @interp.alloc.11 to i32), i32* bitcast (i8** @pointerFree7 to i32*), align 4
  store i32 ptrtoint ([3 x i8]* @interp.alloc.12 to i32), i32* bitcast (i8** @pointerFree3 to i32*), align 4
  store i32 ptrtoint ([0 x i8]* @interp.alloc.13 to i32), i32* bitcast (i8** @pointerFree0 to i32*), align 4
  store i32 ptrtoint ([3 x i8*]* @interp.alloc.14 to i32), i32* bitcast (i8** @layout1 to i32*), align 4
  store i32 ptrtoint ([5 x { i8*, [2 x i32] }]* @interp.alloc.15 to i32), i32* bitcast (i8** @layout2 to i32*), align 4
  store i32 ptrtoint ({ [3 x i8*], [2 x i32], [2 x i8*], [6 x i32], [2 x i8*], [3 x i32], [2 x i8*], [2 x i32], i8*, [2 x i32], i8* }* @interp.alloc.16 to i32), i32* bitcast (i8** @layout3 to i32*), align 4
  store i32 ptrtoint ([3 x { [3 x i8*], [2 x i32], [2 x i8*], [6 x i32], [2 x i8*], [3 x i32], [2 x i8*], [2 x i32], i8*, [2 x i32], i8* }]* @interp.alloc.17 to i32), i32* bitcast (i8** @layout4 to i32*), align 4
  store i32 ptrtoint ({ i8*, [60 x i32], i8* }* @interp.alloc.18 to i32), i32* bitcast (i8** @bigobj1 to i32*), align 4
  ret void
}

define internal void @main.init() unnamed_addr {
  %pointerFree12 = call i8* @runtime.alloc(i32 12, i8* nonnull inttoptr (i32 3 to i8*))
  store i8* %pointerFree12, i8** @pointerFree12, align 4
  %pointerFree7 = call i8* @runtime.alloc(i32 7, i8* nonnull inttoptr (i32 3 to i8*))
  store i8* %pointerFree7, i8** @pointerFree7, align 4
  %pointerFree3 = call i8* @runtime.alloc(i32 3, i8* nonnull inttoptr (i32 3 to i8*))
  store i8* %pointerFree3, i8** @pointerFree3, align 4
  %pointerFree0 = call i8* @runtime.alloc(i32 0, i8* nonnull inttoptr (i32 3 to i8*))
  store i8* %pointerFree0, i8** @pointerFree0, align 4
  %layout1 = call i8* @runtime.alloc(i32 12, i8* nonnull inttoptr (i32 67 to i8*))
  store i8* %layout1, i8** @layout1, align 4
  %layout2 = call i8* @runtime.alloc(i32 60, i8* nonnull inttoptr (i32 71 to i8*))
  store i8* %layout2, i8** @layout2, align 4
  %layout3 = call i8* @runtime.alloc(i32 104, i8* nonnull inttoptr (i32 -1827137035 to i8*))
  store i8* %layout3, i8** @layout3, align 4
  %layout4 = call i8* @runtime.alloc(i32 312, i8* nonnull inttoptr (i32 -1827137035 to i8*))
  store i8* %layout4, i8** @layout4, align 4
  %bigobj1 = call i8* @runtime.alloc(i32 248, i8* bitcast ({ i32, [8 x i8] }* @"runtime/gc.layout:62-2000000000000001" to i8*))
  store i8* %bigobj1, i8** @bigobj1, align 4
  ret void
}
