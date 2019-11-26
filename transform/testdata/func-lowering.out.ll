target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-unknown-wasm"

%runtime.typecodeID = type { %runtime.typecodeID*, i32 }
%runtime.funcValueWithSignature = type { i32, %runtime.typecodeID* }

@"reflect/types.type:func:{basic:int8}{}" = external constant %runtime.typecodeID
@"reflect/types.type:func:{basic:uint8}{}" = external constant %runtime.typecodeID
@"reflect/types.type:func:{basic:int}{}" = external constant %runtime.typecodeID
@"funcInt8$withSignature" = constant %runtime.funcValueWithSignature { i32 ptrtoint (void (i8, i8*, i8*)* @funcInt8 to i32), %runtime.typecodeID* @"reflect/types.type:func:{basic:int8}{}" }
@"func1Uint8$withSignature" = constant %runtime.funcValueWithSignature { i32 ptrtoint (void (i8, i8*, i8*)* @func1Uint8 to i32), %runtime.typecodeID* @"reflect/types.type:func:{basic:uint8}{}" }
@"func2Uint8$withSignature" = constant %runtime.funcValueWithSignature { i32 ptrtoint (void (i8, i8*, i8*)* @func2Uint8 to i32), %runtime.typecodeID* @"reflect/types.type:func:{basic:uint8}{}" }
@"main$withSignature" = constant %runtime.funcValueWithSignature { i32 ptrtoint (void (i32, i8*, i8*)* @"main$1" to i32), %runtime.typecodeID* @"reflect/types.type:func:{basic:int}{}" }
@"main$2$withSignature" = constant %runtime.funcValueWithSignature { i32 ptrtoint (void (i32, i8*, i8*)* @"main$2" to i32), %runtime.typecodeID* @"reflect/types.type:func:{basic:int}{}" }

declare i32 @runtime.getFuncPtr(i8*, i32, %runtime.typecodeID*, i8*, i8*)

declare i32 @runtime.makeGoroutine(i32, i8*, i8*)

declare void @runtime.nilPanic(i8*, i8*)

declare i1 @runtime.isnil(i8*, i8*, i8*)

declare void @"main$1"(i32, i8*, i8*)

declare void @"main$2"(i32, i8*, i8*)

declare void @funcInt8(i8, i8*, i8*)

declare void @func1Uint8(i8, i8*, i8*)

declare void @func2Uint8(i8, i8*, i8*)

; Call a function of which only one function with this signature is used as a
; function value. This means that lowering it to IR is trivial: simply check
; whether the func value is nil, and if not, call that one function directly.
define void @runFunc1(i8*, i32, i8, i8* %context, i8* %parentHandle) {
entry:
  %3 = icmp eq i32 %1, 0
  %4 = select i1 %3, void (i8, i8*, i8*)* null, void (i8, i8*, i8*)* @funcInt8
  %5 = bitcast void (i8, i8*, i8*)* %4 to i8*
  %6 = call i1 @runtime.isnil(i8* %5, i8* undef, i8* null)
  br i1 %6, label %fpcall.nil, label %fpcall.next

fpcall.nil:
  call void @runtime.nilPanic(i8* undef, i8* null)
  unreachable

fpcall.next:
  call void %4(i8 %2, i8* %0, i8* undef)
  ret void
}

; There are two functions with this signature used in a func value. That means
; that we'll have to check at runtime which of the two it is (or whether the
; func value is nil). This call will thus be lowered to a switch statement.
define void @runFunc2(i8*, i32, i8, i8* %context, i8* %parentHandle) {
entry:
  br i1 false, label %fpcall.nil, label %fpcall.next

fpcall.nil:
  call void @runtime.nilPanic(i8* undef, i8* null)
  unreachable

fpcall.next:
  switch i32 %1, label %func.default [
    i32 0, label %func.nil
    i32 1, label %func.call1
    i32 2, label %func.call2
  ]

func.nil:
  call void @runtime.nilPanic(i8* undef, i8* null)
  unreachable

func.call1:
  call void @func1Uint8(i8 %2, i8* %0, i8* undef)
  br label %func.next

func.call2:
  call void @func2Uint8(i8 %2, i8* %0, i8* undef)
  br label %func.next

func.next:
  ret void

func.default:
  unreachable
}

; Special case for runtime.makeGoroutine.
define void @sleepFuncValue(i8*, i32, i8* nocapture readnone %context, i8* nocapture readnone %parentHandle) {
entry:
  switch i32 %1, label %func.default [
    i32 0, label %func.nil
    i32 1, label %func.call1
    i32 2, label %func.call2
  ]

func.nil:
  call void @runtime.nilPanic(i8* undef, i8* null)
  unreachable

func.call1:
  %2 = call i32 @runtime.makeGoroutine(i32 ptrtoint (void (i32, i8*, i8*)* @"main$1" to i32), i8* undef, i8* null)
  %3 = inttoptr i32 %2 to void (i32, i8*, i8*)*
  call void %3(i32 8, i8* %0, i8* null)
  br label %func.next

func.call2:
  %4 = call i32 @runtime.makeGoroutine(i32 ptrtoint (void (i32, i8*, i8*)* @"main$2" to i32), i8* undef, i8* null)
  %5 = inttoptr i32 %4 to void (i32, i8*, i8*)*
  call void %5(i32 8, i8* %0, i8* null)
  br label %func.next

func.next:
  ret void

func.default:
  unreachable
}
