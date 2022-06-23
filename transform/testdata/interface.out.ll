target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

@"reflect/types.type:basic:uint8" = linkonce_odr constant { i8, i8* } { i8 8, i8* getelementptr inbounds ({ i8, i8* }, { i8, i8* }* @"reflect/types.type:pointer:basic:uint8", i32 0, i32 0) }, align 4
@"reflect/types.type:pointer:basic:uint8" = linkonce_odr constant { i8, i8* } { i8 21, i8* getelementptr inbounds ({ i8, i8* }, { i8, i8* }* @"reflect/types.type:basic:uint8", i32 0, i32 0) }, align 4
@"reflect/types.type:basic:int" = linkonce_odr constant { i8, i8* } { i8 2, i8* getelementptr inbounds ({ i8, i8* }, { i8, i8* }* @"reflect/types.type:pointer:basic:int", i32 0, i32 0) }, align 4
@"reflect/types.type:pointer:basic:int" = linkonce_odr constant { i8, i8* } { i8 21, i8* getelementptr inbounds ({ i8, i8* }, { i8, i8* }* @"reflect/types.type:basic:int", i32 0, i32 0) }, align 4
@"reflect/types.type:pointer:named:Number" = linkonce_odr constant { i8, i8* } { i8 21, i8* getelementptr inbounds ({ i8, i8*, i8* }, { i8, i8*, i8* }* @"reflect/types.type:named:Number", i32 0, i32 0) }, align 4
@"reflect/types.type:named:Number" = linkonce_odr constant { i8, i8*, i8* } { i8 34, i8* getelementptr inbounds ({ i8, i8* }, { i8, i8* }* @"reflect/types.type:pointer:named:Number", i32 0, i32 0), i8* getelementptr inbounds ({ i8, i8* }, { i8, i8* }* @"reflect/types.type:basic:int", i32 0, i32 0) }, align 4

declare void @runtime.printuint8(i8)

declare void @runtime.printint16(i16)

declare void @runtime.printint32(i32)

declare void @runtime.printptr(i32)

declare void @runtime.printnl()

declare void @runtime.nilPanic(i8*)

define void @printInterfaces() {
  call void @printInterface(i8* getelementptr inbounds ({ i8, i8* }, { i8, i8* }* @"reflect/types.type:basic:int", i32 0, i32 0), i8* inttoptr (i32 5 to i8*))
  call void @printInterface(i8* getelementptr inbounds ({ i8, i8* }, { i8, i8* }* @"reflect/types.type:basic:uint8", i32 0, i32 0), i8* inttoptr (i8 120 to i8*))
  call void @printInterface(i8* getelementptr inbounds ({ i8, i8*, i8* }, { i8, i8*, i8* }* @"reflect/types.type:named:Number", i32 0, i32 0), i8* inttoptr (i32 3 to i8*))
  ret void
}

define void @printInterface(i8* %typecode, i8* %value) {
  %isUnmatched = call i1 @"Unmatched$typeassert"(i8* %typecode)
  br i1 %isUnmatched, label %typeswitch.Unmatched, label %typeswitch.notUnmatched

typeswitch.Unmatched:                             ; preds = %0
  %unmatched = ptrtoint i8* %value to i32
  call void @runtime.printptr(i32 %unmatched)
  call void @runtime.printnl()
  ret void

typeswitch.notUnmatched:                          ; preds = %0
  %isDoubler = call i1 @"Doubler$typeassert"(i8* %typecode)
  br i1 %isDoubler, label %typeswitch.Doubler, label %typeswitch.notDoubler

typeswitch.Doubler:                               ; preds = %typeswitch.notUnmatched
  %doubler.result = call i32 @"Doubler.Double$invoke"(i8* %value, i8* %typecode, i8* undef)
  call void @runtime.printint32(i32 %doubler.result)
  ret void

typeswitch.notDoubler:                            ; preds = %typeswitch.notUnmatched
  %typeassert.ok = icmp eq i8* getelementptr inbounds ({ i8, i8* }, { i8, i8* }* @"reflect/types.type:basic:uint8", i32 0, i32 0), %typecode
  br i1 %typeassert.ok, label %typeswitch.byte, label %typeswitch.notByte

typeswitch.byte:                                  ; preds = %typeswitch.notDoubler
  %byte = ptrtoint i8* %value to i8
  call void @runtime.printuint8(i8 %byte)
  call void @runtime.printnl()
  ret void

typeswitch.notByte:                               ; preds = %typeswitch.notDoubler
  br i1 false, label %typeswitch.int16, label %typeswitch.notInt16

typeswitch.int16:                                 ; preds = %typeswitch.notByte
  %int16 = ptrtoint i8* %value to i16
  call void @runtime.printint16(i16 %int16)
  call void @runtime.printnl()
  ret void

typeswitch.notInt16:                              ; preds = %typeswitch.notByte
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

define internal i32 @"Doubler.Double$invoke"(i8* %receiver, i8* %actualType, i8* %context) unnamed_addr #0 {
entry:
  %"named:Number.icmp" = icmp eq i8* %actualType, getelementptr inbounds ({ i8, i8*, i8* }, { i8, i8*, i8* }* @"reflect/types.type:named:Number", i32 0, i32 0)
  br i1 %"named:Number.icmp", label %"named:Number", label %"named:Number.next"

"named:Number":                                   ; preds = %entry
  %0 = call i32 @"(Number).Double$invoke"(i8* %receiver, i8* undef)
  ret i32 %0

"named:Number.next":                              ; preds = %entry
  call void @runtime.nilPanic(i8* undef)
  unreachable
}

define internal i1 @"Doubler$typeassert"(i8* %actualType) unnamed_addr #1 {
entry:
  %"named:Number.icmp" = icmp eq i8* %actualType, getelementptr inbounds ({ i8, i8*, i8* }, { i8, i8*, i8* }* @"reflect/types.type:named:Number", i32 0, i32 0)
  br i1 %"named:Number.icmp", label %then, label %"named:Number.next"

then:                                             ; preds = %entry
  ret i1 true

"named:Number.next":                              ; preds = %entry
  ret i1 false
}

define internal i1 @"Unmatched$typeassert"(i8* %actualType) unnamed_addr #2 {
entry:
  ret i1 false

then:                                             ; No predecessors!
  ret i1 true
}

attributes #0 = { "tinygo-invoke"="reflect/methods.Double() int" "tinygo-methods"="reflect/methods.Double() int" }
attributes #1 = { "tinygo-methods"="reflect/methods.Double() int" }
attributes #2 = { "tinygo-methods"="reflect/methods.NeverImplementedMethod()" }
