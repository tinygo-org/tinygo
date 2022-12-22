package main

import (
	"bytes"
	"flag"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

var tmpl = template.Must(template.New("go").Funcs(template.FuncMap{
	"mul": func(x, y int) int {
		return x * y
	},
	"tuple": func(v ...interface{}) []interface{} {
		return v
	},
	"title": strings.Title,
}).Parse(`//go:build baremetal && !tinygo.wasm

// Automatically generated file. DO NOT EDIT.
// This file implements standins for non-native atomics using critical sections.

package runtime

import (
	_ "unsafe"
	"runtime/interrupt"
)

// Documentation:
// * https://llvm.org/docs/Atomics.html
// * https://gcc.gnu.org/onlinedocs/gcc/_005f_005fsync-Builtins.html
//
// Some atomic operations are emitted inline while others are emitted as libcalls.
// How many are emitted as libcalls depends on the MCU arch and core variant.

{{- define "load"}}{{$bits := mul . 8 -}}
//export __atomic_load_{{.}}
func __atomic_load_{{.}}(ptr *uint{{$bits}}, ordering uintptr) uint{{$bits}} {
	// The LLVM docs for this say that there is a val argument after the pointer.
	// That is a typo, and the GCC docs omit it.
	mask := interrupt.Disable()
	val := *ptr
	interrupt.Restore(mask)
	return val
}
{{end}}
{{- define "store"}}{{$bits := mul . 8 -}}
//export __atomic_store_{{.}}
func __atomic_store_{{.}}(ptr *uint{{$bits}}, val uint{{$bits}}, ordering uintptr) {
	mask := interrupt.Disable()
	*ptr = val
	interrupt.Restore(mask)
}
{{end}}
{{- define "cas"}}{{$bits := mul . 8 -}}
//go:inline
func doAtomicCAS{{$bits}}(ptr *uint{{$bits}}, expected, desired uint{{$bits}}) uint{{$bits}} {
	mask := interrupt.Disable()
	old := *ptr
	if old == expected {
		*ptr = desired
	}
	interrupt.Restore(mask)
	return old
}

//export __sync_val_compare_and_swap_{{.}}
func __sync_val_compare_and_swap_{{.}}(ptr *uint{{$bits}}, expected, desired uint{{$bits}}) uint{{$bits}} {
	return doAtomicCAS{{$bits}}(ptr, expected, desired)
}

//export __atomic_compare_exchange_{{.}}
func __atomic_compare_exchange_{{.}}(ptr, expected *uint{{$bits}}, desired uint{{$bits}}, successOrder, failureOrder uintptr) bool {
	exp := *expected
	old := doAtomicCAS{{$bits}}(ptr, exp, desired)
	return old == exp
}
{{end}}
{{- define "swap"}}{{$bits := mul . 8 -}}
//go:inline
func doAtomicSwap{{$bits}}(ptr *uint{{$bits}}, new uint{{$bits}}) uint{{$bits}} {
	mask := interrupt.Disable()
	old := *ptr
	*ptr = new
	interrupt.Restore(mask)
	return old
}

//export __sync_lock_test_and_set_{{.}}
func __sync_lock_test_and_set_{{.}}(ptr *uint{{$bits}}, new uint{{$bits}}) uint{{$bits}} {
	return doAtomicSwap{{$bits}}(ptr, new)
}

//export __atomic_exchange_{{.}}
func __atomic_exchange_{{.}}(ptr *uint{{$bits}}, new uint{{$bits}}, ordering uintptr) uint{{$bits}} {
	return doAtomicSwap{{$bits}}(ptr, new)
}
{{end}}
{{- define "rmw"}}
	{{- $opname := index . 0}}
	{{- $bytes := index . 1}}{{$bits := mul $bytes 8}}
	{{- $signed := index . 2}}
	{{- $opdef := index . 3}}

{{- $type := printf "int%d" $bits}}
{{- if not $signed}}{{$type = printf "u%s" $type}}{{end -}}
{{- $opfn := printf "doAtomic%s%d" (title $opname) $bits}}

//go:inline
func {{$opfn}}(ptr *{{$type}}, value {{$type}}) (old, new {{$type}}) {
	mask := interrupt.Disable()
	old = *ptr
	{{$opdef}}
	*ptr = new
	interrupt.Restore(mask)
	return old, new
}

//export __atomic_fetch_{{$opname}}_{{$bytes}}
func __atomic_fetch_{{$opname}}_{{$bytes}}(ptr *{{$type}}, value {{$type}}, ordering uintptr) {{$type}} {
	old, _ := {{$opfn}}(ptr, value)
	return old
}

//export __sync_fetch_and_{{$opname}}_{{$bytes}}
func __sync_fetch_and_{{$opname}}_{{$bytes}}(ptr *{{$type}}, value {{$type}}) {{$type}} {
	old, _ := {{$opfn}}(ptr, value)
	return old
}

//export __atomic_{{$opname}}_fetch_{{$bytes}}
func __atomic_{{$opname}}_fetch_{{$bytes}}(ptr *{{$type}}, value {{$type}}, ordering uintptr) {{$type}} {
	_, new := {{$opfn}}(ptr, value)
	return new
}
{{end}}
{{- define "atomics"}}
// {{mul . 8}}-bit atomics.

{{/* These atomics are accessible directly from sync/atomic. */ -}}
{{template "load" .}}
{{template "store" .}}
{{template "cas" .}}
{{template "swap" .}}
{{template "rmw" (tuple "add" . false "new = old + value")}}

{{- end}}
{{template "atomics" 2 -}}
{{template "atomics" 4 -}}
{{template "atomics" 8}}
`))

func main() {
	var out string
	flag.StringVar(&out, "out", "-", "output path")
	flag.Parse()
	f := os.Stdout
	if out != "-" {
		var err error
		f, err = os.Create(out)
		if err != nil {
			panic(err)
		}
		defer f.Close()
	}
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, nil)
	if err != nil {
		panic(err)
	}
	cmd := exec.Command("gofmt")
	cmd.Stdin = &buf
	cmd.Stdout = f
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}
