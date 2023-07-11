invoke: clean wasm_exec
	tinygo build -o ./html/wasm.wasm -target wasm -no-debug ./invoke/wasm.go
	cp ./invoke/wasm.js ./html/
	cp ./invoke/index.html ./html/

export: clean wasm_exec
	tinygo build -o ./html/wasm.wasm -target wasm -no-debug ./export/wasm.go
	cp ./export/wasm.js ./html/
	cp ./export/index.html ./html/

callback: clean wasm_exec
	tinygo build -o ./html/wasm.wasm -target wasm ./callback/wasm.go
	cp ./callback/wasm.js ./html/
	cp ./callback/index.html ./html/

slices: clean wasm_exec
	tinygo build -o ./html/wasm.wasm -target wasm -no-debug ./slices/wasm.go
	cp ./slices/wasm.js ./html/
	cp ./slices/index.html ./html/


main: clean wasm_exec
	tinygo build -o ./html/wasm.wasm -target wasm -no-debug ./main/main.go
	cp ./main/index.html ./html/

wasm_exec:
	cp `tinygo env TINYGOROOT`/targets/wasm_exec.js ./html/

clean:
	rm -rf ./html
	mkdir ./html
