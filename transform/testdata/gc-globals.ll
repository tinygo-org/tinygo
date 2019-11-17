target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-unknown-wasm"

%runtime._string = type { i8*, i32 }
%runtime._interface = type { i32, i8* }

@globalInt = constant i32 5
@globalString = constant %runtime._string zeroinitializer
@globalInterface = constant %runtime._interface zeroinitializer
@runtime.trackedGlobalsLength = external global i32
@runtime.trackedGlobalsBitmap = external global [0 x i8]
@runtime.trackedGlobalsStart = external global i32

define void @main() {
  %1 = load i32, i32* @globalInt
  %2 = load %runtime._string, %runtime._string* @globalString
  %3 = load %runtime._interface, %runtime._interface* @globalInterface
  ret void
}

define void @runtime.markGlobals() {
  ; Very small subset of what runtime.markGlobals would really do.
  ; Just enough to make sure the transformation is correct.
  %1 = load i32, i32* @runtime.trackedGlobalsStart
  %2 = load i32, i32* @runtime.trackedGlobalsLength
  %3 = getelementptr inbounds [0 x i8], [0 x i8]* @runtime.trackedGlobalsBitmap, i32 0, i32 0
  %4 = load i8, i8* %3
  ret void
}
