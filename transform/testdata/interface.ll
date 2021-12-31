target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

%runtime.typecodeID = type { %runtime.typecodeID*, i32, %runtime.interfaceMethodInfo*, %runtime.typecodeID*, i32 }
%runtime.interfaceMethodInfo = type { i8*, i32 }

@"reflect/types.type:basic:uint8" = private constant %runtime.typecodeID zeroinitializer
@"reflect/types.typeid:basic:uint8" = external constant i8
@"reflect/types.typeid:basic:int16" = external constant i8
@"reflect/types.type:basic:int" = private constant %runtime.typecodeID zeroinitializer
@"reflect/methods.NeverImplementedMethod()" = linkonce_odr constant i8 0
@"reflect/methods.Double() int" = linkonce_odr constant i8 0
@"Number$methodset" = private constant [1 x %runtime.interfaceMethodInfo] [%runtime.interfaceMethodInfo { i8* @"reflect/methods.Double() int", i32 ptrtoint (i32 (i8*, i8*)* @"(Number).Double$invoke" to i32) }]
@"reflect/types.type:named:Number" = private constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:basic:int", i32 0, %runtime.interfaceMethodInfo* getelementptr inbounds ([1 x %runtime.interfaceMethodInfo], [1 x %runtime.interfaceMethodInfo]* @"Number$methodset", i32 0, i32 0), %runtime.typecodeID* null, i32 0 }

declare i1 @runtime.typeAssert(i32, i8*)
declare void @runtime.printuint8(i8)
declare void @runtime.printint16(i16)
declare void @runtime.printint32(i32)
declare void @runtime.printptr(i32)
declare void @runtime.printnl()
declare void @runtime.nilPanic(i8*)

define void @printInterfaces() {
  call void @printInterface(i32 ptrtoint (%runtime.typecodeID* @"reflect/types.type:basic:int" to i32), i8* inttoptr (i32 5 to i8*))
  call void @printInterface(i32 ptrtoint (%runtime.typecodeID* @"reflect/types.type:basic:uint8" to i32), i8* inttoptr (i8 120 to i8*))
  call void @printInterface(i32 ptrtoint (%runtime.typecodeID* @"reflect/types.type:named:Number" to i32), i8* inttoptr (i32 3 to i8*))

  ret void
}

define void @printInterface(i32 %typecode, i8* %value) {
  %isUnmatched = call i1 @Unmatched$typeassert(i32 %typecode)
  br i1 %isUnmatched, label %typeswitch.Unmatched, label %typeswitch.notUnmatched

typeswitch.Unmatched:
  %unmatched = ptrtoint i8* %value to i32
  call void @runtime.printptr(i32 %unmatched)
  call void @runtime.printnl()
  ret void

typeswitch.notUnmatched:
  %isDoubler = call i1 @Doubler$typeassert(i32 %typecode)
  br i1 %isDoubler, label %typeswitch.Doubler, label %typeswitch.notDoubler

typeswitch.Doubler:
  %doubler.result = call i32 @"Doubler.Double$invoke"(i8* %value, i32 %typecode, i8* undef)
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

define i32 @"(Number).Double"(i32 %receiver, i8* %context) {
  %ret = mul i32 %receiver, 2
  ret i32 %ret
}

define i32 @"(Number).Double$invoke"(i8* %receiverPtr, i8* %context) {
  %receiver = ptrtoint i8* %receiverPtr to i32
  %ret = call i32 @"(Number).Double"(i32 %receiver, i8* undef)
  ret i32 %ret
}

declare i32 @"Doubler.Double$invoke"(i8* %receiver, i32 %typecode, i8* %context) #0

declare i1 @Doubler$typeassert(i32 %typecode) #1

declare i1 @Unmatched$typeassert(i32 %typecode) #2

attributes #0 = { "tinygo-invoke"="reflect/methods.Double() int" "tinygo-methods"="reflect/methods.Double() int" }
attributes #1 = { "tinygo-methods"="reflect/methods.Double() int" }
attributes #2 = { "tinygo-methods"="reflect/methods.NeverImplementedMethod()" }
