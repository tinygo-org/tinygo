target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-unknown-wasm"

%runtime.stackChainObject = type { %runtime.stackChainObject*, i32 }

@runtime.stackChainStart = global %runtime.stackChainObject* null
@someGlobal = global i8 3

declare void @runtime.trackPointer(i8* nocapture readonly)

declare noalias nonnull i8* @runtime.alloc(i32)

define i8* @getPointer() {
  ret i8* @someGlobal
}

define i8* @needsStackSlots() {
  %gc.stackobject = alloca { %runtime.stackChainObject*, i32, i8* }
  store { %runtime.stackChainObject*, i32, i8* } { %runtime.stackChainObject* null, i32 1, i8* null }, { %runtime.stackChainObject*, i32, i8* }* %gc.stackobject
  %1 = load %runtime.stackChainObject*, %runtime.stackChainObject** @runtime.stackChainStart
  %2 = getelementptr { %runtime.stackChainObject*, i32, i8* }, { %runtime.stackChainObject*, i32, i8* }* %gc.stackobject, i32 0, i32 0
  store %runtime.stackChainObject* %1, %runtime.stackChainObject** %2
  %3 = bitcast { %runtime.stackChainObject*, i32, i8* }* %gc.stackobject to %runtime.stackChainObject*
  store %runtime.stackChainObject* %3, %runtime.stackChainObject** @runtime.stackChainStart
  %ptr = call i8* @runtime.alloc(i32 4)
  %4 = getelementptr { %runtime.stackChainObject*, i32, i8* }, { %runtime.stackChainObject*, i32, i8* }* %gc.stackobject, i32 0, i32 2
  store i8* %ptr, i8** %4
  store %runtime.stackChainObject* %1, %runtime.stackChainObject** @runtime.stackChainStart
  ret i8* %ptr
}

define i8* @needsStackSlots2() {
  %gc.stackobject = alloca { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }
  store { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* } { %runtime.stackChainObject* null, i32 4, i8* null, i8* null, i8* null, i8* null }, { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }* %gc.stackobject
  %1 = load %runtime.stackChainObject*, %runtime.stackChainObject** @runtime.stackChainStart
  %2 = getelementptr { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }, { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }* %gc.stackobject, i32 0, i32 0
  store %runtime.stackChainObject* %1, %runtime.stackChainObject** %2
  %3 = bitcast { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }* %gc.stackobject to %runtime.stackChainObject*
  store %runtime.stackChainObject* %3, %runtime.stackChainObject** @runtime.stackChainStart
  %ptr1 = call i8* @getPointer()
  %4 = getelementptr { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }, { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }* %gc.stackobject, i32 0, i32 4
  store i8* %ptr1, i8** %4
  %5 = getelementptr { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }, { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }* %gc.stackobject, i32 0, i32 3
  store i8* %ptr1, i8** %5
  %6 = getelementptr { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }, { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }* %gc.stackobject, i32 0, i32 2
  store i8* %ptr1, i8** %6
  %ptr2 = getelementptr i8, i8* @someGlobal, i32 0
  %unused = call i8* @runtime.alloc(i32 4)
  %7 = getelementptr { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }, { %runtime.stackChainObject*, i32, i8*, i8*, i8*, i8* }* %gc.stackobject, i32 0, i32 5
  store i8* %unused, i8** %7
  store %runtime.stackChainObject* %1, %runtime.stackChainObject** @runtime.stackChainStart
  ret i8* %ptr1
}

define i8* @noAllocatingFunction() {
  %ptr = call i8* @getPointer()
  ret i8* %ptr
}
