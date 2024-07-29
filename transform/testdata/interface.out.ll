target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

@"reflect/types.type:basic:uint8" = linkonce_odr constant { i8, ptr } { i8 8, ptr @"reflect/types.type:pointer:basic:uint8" }, align 4
@"reflect/types.type:pointer:basic:uint8" = linkonce_odr constant { i8, ptr } { i8 21, ptr @"reflect/types.type:basic:uint8" }, align 4
@"reflect/types.type:basic:int" = linkonce_odr constant { i8, ptr } { i8 2, ptr @"reflect/types.type:pointer:basic:int" }, align 4
@"reflect/types.type:pointer:basic:int" = linkonce_odr constant { i8, ptr } { i8 21, ptr @"reflect/types.type:basic:int" }, align 4
@"reflect/types.type:pointer:named:Number" = linkonce_odr constant { i8, ptr } { i8 21, ptr @"reflect/types.type:named:Number" }, align 4
@"reflect/types.type:named:Number" = linkonce_odr constant { i8, ptr, ptr } { i8 34, ptr @"reflect/types.type:pointer:named:Number", ptr @"reflect/types.type:basic:int" }, align 4

declare void @runtime.printuint8(i8)

declare void @runtime.printint16(i16)

declare void @runtime.printnl()

define void @printInterfaces() {
  call void @printInterface(ptr @"reflect/types.type:basic:int", ptr inttoptr (i32 5 to ptr))
  call void @printInterface(ptr @"reflect/types.type:basic:uint8", ptr inttoptr (i8 120 to ptr))
  call void @printInterface(ptr @"reflect/types.type:named:Number", ptr inttoptr (i32 3 to ptr))
  ret void
}

define void @printInterface(ptr %typecode, ptr %value) {
  %typeassert.ok = icmp eq ptr @"reflect/types.type:basic:uint8", %typecode
  br i1 %typeassert.ok, label %typeswitch.byte, label %typeswitch.notByte

typeswitch.byte:                                  ; preds = %0
  %byte = ptrtoint ptr %value to i8
  call void @runtime.printuint8(i8 %byte)
  call void @runtime.printnl()
  ret void

typeswitch.notByte:                               ; preds = %0
  br i1 false, label %typeswitch.int16, label %typeswitch.notInt16

typeswitch.int16:                                 ; preds = %typeswitch.notByte
  %int16 = ptrtoint ptr %value to i16
  call void @runtime.printint16(i16 %int16)
  call void @runtime.printnl()
  ret void

typeswitch.notInt16:                              ; preds = %typeswitch.notByte
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
