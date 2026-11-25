; ModuleID = 'inlineasm.go'
source_filename = "inlineasm.go"
target datalayout = "e-m:e-p:32:32-Fi8-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "thumbv7m-unknown-unknown-eabi"

%runtime._interface = type { ptr, ptr }

@"runtime/gc.layout:36-5455f5ff0f" = linkonce_odr unnamed_addr constant { i32, [5 x i8] } { i32 36, [5 x i8] c"TU\F5\FF\0F" }
@"runtime.hashmapType:string:interface{}" = linkonce_odr unnamed_addr constant { ptr, ptr, ptr } { ptr inttoptr (i32 69 to ptr), ptr inttoptr (i32 67 to ptr), ptr @"runtime/gc.layout:36-5455f5ff0f" }
@"reflect/types.type:basic:uintptr" = linkonce_odr constant { i8, ptr } { i8 -52, ptr @"reflect/types.type:pointer:basic:uintptr" }, align 4
@"reflect/types.type:pointer:basic:uintptr" = linkonce_odr constant { i8, i8, i16, ptr } { i8 -43, i8 0, i16 0, ptr @"reflect/types.type:basic:uintptr" }, align 4
@"main$string" = internal unnamed_addr constant [5 x i8] c"state", align 1

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #0 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @main.asm(ptr %context) unnamed_addr #0 {
entry:
  call void asm sideeffect "nop", "~{memory}"() #3
  ret void
}

; Function Attrs: nounwind
define hidden i32 @main.asmFull(ptr %context) unnamed_addr #0 {
entry:
  %0 = call i32 asm sideeffect "mrs $0, PRIMASK", "=&r,~{memory}"() #3
  ret i32 %0
}

; Function Attrs: nounwind
define hidden void @main.asmFullInput(i32 %state, ptr %context) unnamed_addr #0 {
entry:
  %hashmap.value = alloca %runtime._interface, align 4
  %0 = call ptr @runtime.hashmapMakeGeneric(i32 8, i32 8, i32 1, ptr nonnull @"runtime.hashmapType:string:interface{}", ptr null, ptr nonnull @runtime.hashmapStringPtrHash, ptr null, ptr nonnull @runtime.hashmapStringEqual, ptr undef) #3
  %pack.int = inttoptr i32 %state to ptr
  call void @llvm.lifetime.start.p0(ptr nonnull %hashmap.value)
  store ptr @"reflect/types.type:basic:uintptr", ptr %hashmap.value, align 4
  %hashmap.value.repack1 = getelementptr inbounds nuw i8, ptr %hashmap.value, i32 4
  store ptr %pack.int, ptr %hashmap.value.repack1, align 4
  call void @runtime.hashmapStringSet(ptr %0, ptr nonnull @"main$string", i32 5, ptr nonnull %hashmap.value, ptr undef) #3
  call void @llvm.lifetime.end.p0(ptr nonnull %hashmap.value)
  call void asm sideeffect "msr PRIMASK, ${0}", "r,~{memory}"(i32 %state) #3
  ret void
}

declare i32 @runtime.hashmapStringPtrHash(ptr, i32, i32, ptr) #1

declare i1 @runtime.hashmapStringEqual(ptr, ptr, i32, ptr) #1

declare ptr @runtime.hashmapMakeGeneric(i32, i32, i32, ptr, ptr, ptr, ptr, ptr, ptr) #1

; Function Attrs: nocallback nofree nosync nounwind willreturn memory(argmem: readwrite)
declare void @llvm.lifetime.start.p0(ptr captures(none)) #2

declare void @runtime.hashmapStringSet(ptr dereferenceable_or_null(52), ptr readonly, i32 range(i32 0, -2147483648), ptr captures(none), ptr) #1

; Function Attrs: nocallback nofree nosync nounwind willreturn memory(argmem: readwrite)
declare void @llvm.lifetime.end.p0(ptr captures(none)) #2

; Function Attrs: nounwind
define hidden void @main.main(ptr %context) unnamed_addr #0 {
entry:
  ret void
}

attributes #0 = { nounwind "target-features"="+armv7-m,+hwdiv,+soft-float,+thumb-mode,-aes,-bf16,-cdecp0,-cdecp1,-cdecp2,-cdecp3,-cdecp4,-cdecp5,-cdecp6,-cdecp7,-crc,-crypto,-d32,-dotprod,-dsp,-fp-armv8,-fp-armv8d16,-fp-armv8d16sp,-fp-armv8sp,-fp16,-fp16fml,-fp64,-fpregs,-fullfp16,-hwdiv-arm,-i8mm,-lob,-mve,-mve.fp,-neon,-pacbti,-ras,-sb,-sha2,-vfp2,-vfp2sp,-vfp3,-vfp3d16,-vfp3d16sp,-vfp3sp,-vfp4,-vfp4d16,-vfp4d16sp,-vfp4sp" }
attributes #1 = { "target-features"="+armv7-m,+hwdiv,+soft-float,+thumb-mode,-aes,-bf16,-cdecp0,-cdecp1,-cdecp2,-cdecp3,-cdecp4,-cdecp5,-cdecp6,-cdecp7,-crc,-crypto,-d32,-dotprod,-dsp,-fp-armv8,-fp-armv8d16,-fp-armv8d16sp,-fp-armv8sp,-fp16,-fp16fml,-fp64,-fpregs,-fullfp16,-hwdiv-arm,-i8mm,-lob,-mve,-mve.fp,-neon,-pacbti,-ras,-sb,-sha2,-vfp2,-vfp2sp,-vfp3,-vfp3d16,-vfp3d16sp,-vfp3sp,-vfp4,-vfp4d16,-vfp4d16sp,-vfp4sp" }
attributes #2 = { nocallback nofree nosync nounwind willreturn memory(argmem: readwrite) }
attributes #3 = { nounwind }
