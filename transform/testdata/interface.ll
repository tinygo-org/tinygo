target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

%runtime.typecodeID = type { %runtime.typecodeID*, i32, %runtime.interfaceMethodInfo* }
%runtime.interfaceMethodInfo = type { i8*, i32 }

@"reflect/types.type:basic:uint8" = external constant %runtime.typecodeID
@"reflect/types.typeid:basic:uint8" = external constant i8
@"reflect/types.typeid:basic:int16" = external constant i8
@"reflect/types.type:basic:int" = external constant %runtime.typecodeID
@"func NeverImplementedMethod()" = external constant i8
@"Unmatched$interface" = private constant [1 x i8*] [i8* @"func NeverImplementedMethod()"]
@"func Double() int" = external constant i8
@"Doubler$interface" = private constant [1 x i8*] [i8* @"func Double() int"]
@"Number$methodset" = private constant [1 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { i8* @"func Double() int", i32 ptrtoint (i32 (i8*, i8*)* @"(Number).Double$invoke" to i32) }]
@"reflect/types.type:named:Number" = private constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:basic:int", i32 0, %runtime.interfaceMethodInfo* getelementptr inbounds ([1 x %runtime.interfaceMethodInfo], [1 x %runtime.interfaceMethodInfo]* @"Number$methodset", i32 0, i32 0) }

declare i1 @runtime.interfaceImplements(i32, i8**)
declare i1 @runtime.typeAssert(i32, i8*)
declare i32 @runtime.interfaceMethod(i32, i8**, i8*)
declare void @runtime.printuint8(i8)
declare void @runtime.printint16(i16)
declare void @runtime.printint32(i32)
declare void @runtime.printptr(i32)
declare void @runtime.printnl()
declare void @runtime.nilPanic(i8*, i8*)

define void @printInterfaces() {
  call void @printInterface(i32 ptrtoint (%runtime.typecodeID* @"reflect/types.type:basic:int" to i32), i8* inttoptr (i32 5 to i8*))
  call void @printInterface(i32 ptrtoint (%runtime.typecodeID* @"reflect/types.type:basic:uint8" to i32), i8* inttoptr (i8 120 to i8*))
  call void @printInterface(i32 ptrtoint (%runtime.typecodeID* @"reflect/types.type:named:Number" to i32), i8* inttoptr (i32 3 to i8*))

  ret void
}

define void @printInterface(i32 %typecode, i8* %value) {
  %isUnmatched = call i1 @runtime.interfaceImplements(i32 %typecode, i8** getelementptr inbounds ([1 x i8*], [1 x i8*]* @"Unmatched$interface", i32 0, i32 0))
  br i1 %isUnmatched, label %typeswitch.Unmatched, label %typeswitch.notUnmatched

typeswitch.Unmatched:
  %unmatched = ptrtoint i8* %value to i32
  call void @runtime.printptr(i32 %unmatched)
  call void @runtime.printnl()
  ret void

typeswitch.notUnmatched:
  %isDoubler = call i1 @runtime.interfaceImplements(i32 %typecode, i8** getelementptr inbounds ([1 x i8*], [1 x i8*]* @"Doubler$interface", i32 0, i32 0))
  br i1 %isDoubler, label %typeswitch.Doubler, label %typeswitch.notDoubler

typeswitch.Doubler:
  %doubler.func = call i32 @runtime.interfaceMethod(i32 %typecode, i8** getelementptr inbounds ([1 x i8*], [1 x i8*]* @"Doubler$interface", i32 0, i32 0), i8* nonnull @"func Double() int")
  %doubler.func.cast = inttoptr i32 %doubler.func to i32 (i8*, i8*)*
  %doubler.result = call i32 %doubler.func.cast(i8* %value, i8* null)
  call void @runtime.printint32(i32 %doubler.result)
  ret void

typeswitch.notDoubler:
  %isByte = call i1 @runtime.typeAssert(i32 %typecode, i8* nonnull @"reflect/types.typeid:basic:uint8")
  br i1 %isByte, label %typeswitch.byte, label %typeswitch.notByte

typeswitch.byte:
  %byte = ptrtoint i8* %value to i8
  call void @runtime.printuint8(i8 %byte)
  call void @runtime.printnl()
  ret void

typeswitch.notByte:
  ; this is a type assert that always fails
  %isInt16 = call i1 @runtime.typeAssert(i32 %typecode, i8* nonnull @"reflect/types.typeid:basic:int16")
  br i1 %isInt16, label %typeswitch.int16, label %typeswitch.notInt16

typeswitch.int16:
  %int16 = ptrtoint i8* %value to i16
  call void @runtime.printint16(i16 %int16)
  call void @runtime.printnl()
  ret void

typeswitch.notInt16:
  ret void
}

define i32 @"(Number).Double"(i32 %receiver, i8* %parentHandle) {
  %ret = mul i32 %receiver, 2
  ret i32 %ret
}

define i32 @"(Number).Double$invoke"(i8* %receiverPtr, i8* %parentHandle) {
  %receiver = ptrtoint i8* %receiverPtr to i32
  %ret = call i32 @"(Number).Double"(i32 %receiver, i8* null)
  ret i32 %ret
}
