target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

@"reflect/types.type:basic:uint8" = linkonce_odr constant { i8, ptr } { i8 8, ptr @"reflect/types.type:pointer:basic:uint8" }, align 4
@"reflect/types.type:pointer:basic:uint8" = linkonce_odr constant { i8, ptr } { i8 21, ptr @"reflect/types.type:basic:uint8" }, align 4
@"reflect/types.typeid:basic:uint8" = external constant i8
@"reflect/types.typeid:basic:int16" = external constant i8
@"reflect/types.type:basic:int" = linkonce_odr constant { i8, ptr } { i8 2, ptr @"reflect/types.type:pointer:basic:int" }, align 4
@"reflect/types.type:pointer:basic:int" = linkonce_odr constant { i8, ptr } { i8 21, ptr @"reflect/types.type:basic:int" }, align 4
@"reflect/methods.NeverImplementedMethod()" = linkonce_odr constant i8 0
@"reflect/methods.Double() int" = linkonce_odr constant i8 0
@"Number$methodset" = linkonce_odr unnamed_addr constant { i32, [1 x ptr], { ptr } } { i32 1, [1 x ptr] [ptr @"reflect/methods.Double() int"], { ptr } { ptr @"(Number).Double$invoke" } }
@"reflect/types.type:named:Number" = linkonce_odr constant { ptr, i8, ptr, ptr } { ptr @"Number$methodset", i8 34, ptr @"reflect/types.type:pointer:named:Number", ptr @"reflect/types.type:basic:int" }, align 4
@"reflect/types.type:pointer:named:Number" = linkonce_odr constant { i8, ptr } { i8 21, ptr getelementptr inbounds ({ ptr, i8, ptr, ptr }, ptr @"reflect/types.type:named:Number", i32 0, i32 1) }, align 4

declare i1 @runtime.typeAssert(ptr, ptr)
declare void @runtime.printuint8(i8)
declare void @runtime.printint16(i16)
declare void @runtime.printint32(i32)
declare void @runtime.printptr(i32)
declare void @runtime.printnl()
declare void @runtime.nilPanic(ptr)

define void @printInterfaces() {
  call void @printInterface(ptr @"reflect/types.type:basic:int", ptr inttoptr (i32 5 to ptr))
  call void @printInterface(ptr @"reflect/types.type:basic:uint8", ptr inttoptr (i8 120 to ptr))
  call void @printInterface(ptr getelementptr inbounds ({ ptr, i8, ptr, ptr }, ptr @"reflect/types.type:named:Number", i32 0, i32 1), ptr inttoptr (i32 3 to ptr))

  ret void
}

define void @printInterface(ptr %typecode, ptr %value) {
  %isUnmatched = call i1 @Unmatched$typeassert(ptr %typecode)
  br i1 %isUnmatched, label %typeswitch.Unmatched, label %typeswitch.notUnmatched

typeswitch.Unmatched:
  %unmatched = ptrtoint ptr %value to i32
  call void @runtime.printptr(i32 %unmatched)
  call void @runtime.printnl()
  ret void

typeswitch.notUnmatched:
  %isDoubler = call i1 @Doubler$typeassert(ptr %typecode)
  br i1 %isDoubler, label %typeswitch.Doubler, label %typeswitch.notDoubler

typeswitch.Doubler:
  %doubler.result = call i32 @"Doubler.Double$invoke"(ptr %value, ptr %typecode, ptr undef)
  call void @runtime.printint32(i32 %doubler.result)
  ret void

typeswitch.notDoubler:
  %isByte = call i1 @runtime.typeAssert(ptr %typecode, ptr nonnull @"reflect/types.typeid:basic:uint8")
  br i1 %isByte, label %typeswitch.byte, label %typeswitch.notByte

typeswitch.byte:
  %byte = ptrtoint ptr %value to i8
  call void @runtime.printuint8(i8 %byte)
  call void @runtime.printnl()
  ret void

typeswitch.notByte:
  ; this is a type assert that always fails
  %isInt16 = call i1 @runtime.typeAssert(ptr %typecode, ptr nonnull @"reflect/types.typeid:basic:int16")
  br i1 %isInt16, label %typeswitch.int16, label %typeswitch.notInt16

typeswitch.int16:
  %int16 = ptrtoint ptr %value to i16
  call void @runtime.printint16(i16 %int16)
  call void @runtime.printnl()
  ret void

typeswitch.notInt16:
  ret void
}

define i32 @"(Number).Double"(i32 %receiver, ptr %context) {
  %ret = mul i32 %receiver, 2
  ret i32 %ret
}

define i32 @"(Number).Double$invoke"(ptr %receiverPtr, ptr %context) {
  %receiver = ptrtoint ptr %receiverPtr to i32
  %ret = call i32 @"(Number).Double"(i32 %receiver, ptr undef)
  ret i32 %ret
}

declare i32 @"Doubler.Double$invoke"(ptr %receiver, ptr %typecode, ptr %context) #0

declare i1 @Doubler$typeassert(ptr %typecode) #1

declare i1 @Unmatched$typeassert(ptr %typecode) #2

attributes #0 = { "tinygo-invoke"="reflect/methods.Double() int" "tinygo-methods"="reflect/methods.Double() int" }
attributes #1 = { "tinygo-methods"="reflect/methods.Double() int" }
attributes #2 = { "tinygo-methods"="reflect/methods.NeverImplementedMethod()" }
