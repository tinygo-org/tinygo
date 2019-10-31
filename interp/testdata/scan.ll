target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

define i64 @returnsConst() {
  ret i64 0
}

define i64 @returnsArg(i64 %arg) {
  ret i64 %arg
}

declare i64 @externalCall()

define i64 @externalCallOnly() {
  %result = call i64 @externalCall()
  ret i64 0
}

define i64 @externalCallAndReturn() {
  %result = call i64 @externalCall()
  ret i64 %result
}

define i64 @externalCallBranch() {
  %result = call i64 @externalCall()
  %zero = icmp eq i64 %result, 0
  br i1 %zero, label %if.then, label %if.done

if.then:
  ret i64 2

if.done:
  ret i64 4
}

@cleanGlobalInt = global i64 5
define i64 @readCleanGlobal() {
  %global = load i64, i64* @cleanGlobalInt
  ret i64 %global
}

@dirtyGlobalInt = global i64 5
define i64 @readDirtyGlobal() {
  %global = load i64, i64* @dirtyGlobalInt
  ret i64 %global
}

@functionPointer = global i64()* null
define i64 @callFunctionPointer() {
  %fp = load i64()*, i64()** @functionPointer
  %result = call i64 %fp()
  ret i64 %result
}
