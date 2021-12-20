target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-unknown-wasm"

%runtime._string = type { i8*, i32 }
%runtime._interface = type { i32, i8* }

@globalInt = global i32 5
@constString = constant %runtime._string zeroinitializer
@constInterface = constant %runtime._interface zeroinitializer
@runtime.trackedGlobalsLength = internal global i32 4
@runtime.trackedGlobalsBitmap = external global [0 x i8]
@runtime.trackedGlobalsStart = internal global i32 ptrtoint ({ %runtime._string, %runtime._interface }* @tinygo.trackedGlobals to i32)
@tinygo.trackedGlobals = internal unnamed_addr global { %runtime._string, %runtime._interface } zeroinitializer
@runtime.trackedGlobalsBitmap.1 = internal global [1 x i8] c"\09"

define void @main() {
  %1 = load i32, i32* @globalInt, align 4
  %2 = load %runtime._string, %runtime._string* getelementptr inbounds ({ %runtime._string, %runtime._interface }, { %runtime._string, %runtime._interface }* @tinygo.trackedGlobals, i32 0, i32 0), align 4
  %3 = load %runtime._interface, %runtime._interface* getelementptr inbounds ({ %runtime._string, %runtime._interface }, { %runtime._string, %runtime._interface }* @tinygo.trackedGlobals, i32 0, i32 1), align 4
  %4 = load %runtime._string, %runtime._string* @constString, align 4
  %5 = load %runtime._interface, %runtime._interface* @constInterface, align 4
  ret void
}

define void @runtime.markGlobals() {
  %1 = load i32, i32* @runtime.trackedGlobalsStart, align 4
  %2 = load i32, i32* @runtime.trackedGlobalsLength, align 4
  %3 = getelementptr inbounds [0 x i8], [0 x i8]* bitcast ([1 x i8]* @runtime.trackedGlobalsBitmap.1 to [0 x i8]*), i32 0, i32 0
  %4 = load i8, i8* %3, align 1
  ret void
}
