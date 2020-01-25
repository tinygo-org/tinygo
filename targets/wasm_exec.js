// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// This file has been modified for use by the TinyGo compiler.

(() => {
	// Map web browser API and Node.js API to a single common API (preferring web standards over Node.js API).
	const isNodeJS = typeof process !== "undefined";
	if (isNodeJS) {
		global.require = require;
		global.fs = require("fs");

		const nodeCrypto = require("crypto");
		global.crypto = {
			getRandomValues(b) {
				nodeCrypto.randomFillSync(b);
			},
		};

		global.performance = {
			now() {
				const [sec, nsec] = process.hrtime();
				return sec * 1000 + nsec / 1000000;
			},
		};

		const util = require("util");
		global.TextEncoder = util.TextEncoder;
		global.TextDecoder = util.TextDecoder;
	} else {
		if (typeof window !== "undefined") {
			window.global = window;
		} else if (typeof self !== "undefined") {
			self.global = self;
		} else {
			throw new Error("cannot export Go (neither window nor self is defined)");
		}

		let outputBuf = "";
		global.fs = {
			constants: { O_WRONLY: -1, O_RDWR: -1, O_CREAT: -1, O_TRUNC: -1, O_APPEND: -1, O_EXCL: -1 }, // unused
			writeSync(fd, buf) {
				outputBuf += decoder.decode(buf);
				const nl = outputBuf.lastIndexOf("\n");
				if (nl != -1) {
					console.log(outputBuf.substr(0, nl));
					outputBuf = outputBuf.substr(nl + 1);
				}
				return buf.length;
			},
			write(fd, buf, offset, length, position, callback) {
				if (offset !== 0 || length !== buf.length || position !== null) {
					throw new Error("not implemented");
				}
				const n = this.writeSync(fd, buf);
				callback(null, n);
			},
			open(path, flags, mode, callback) {
				const err = new Error("not implemented");
				err.code = "ENOSYS";
				callback(err);
			},
			fsync(fd, callback) {
				callback(null);
			},
		};
	}

	const encoder = new TextEncoder("utf-8");
	const decoder = new TextDecoder("utf-8");
	var logLine = [];

	global.Go = class {
		constructor() {
			this._callbackTimeouts = new Map();
			this._nextCallbackTimeoutID = 1;

			const mem = () => {
				// The buffer may change when requesting more memory.
				return new DataView(this._inst.exports.memory.buffer);
			}

			const setInt64 = (addr, v) => {
				mem().setUint32(addr + 0, v, true);
				mem().setUint32(addr + 4, Math.floor(v / 4294967296), true);
			}

			const getInt64 = (addr) => {
				const low = mem().getUint32(addr + 0, true);
				const high = mem().getInt32(addr + 4, true);
				return low + high * 4294967296;
			}

			const loadValue = (addr) => {
				const f = mem().getFloat64(addr, true);
				if (f === 0) {
					return undefined;
				}
				if (!isNaN(f)) {
					return f;
				}

				const id = mem().getUint32(addr, true);
				return this._values[id];
			}

			const storeValue = (addr, v) => {
				const nanHead = 0x7FF80000;

				if (typeof v === "number") {
					if (isNaN(v)) {
						mem().setUint32(addr + 4, nanHead, true);
						mem().setUint32(addr, 0, true);
						return;
					}
					if (v === 0) {
						mem().setUint32(addr + 4, nanHead, true);
						mem().setUint32(addr, 1, true);
						return;
					}
					mem().setFloat64(addr, v, true);
					return;
				}

				switch (v) {
					case undefined:
						mem().setFloat64(addr, 0, true);
						return;
					case null:
						mem().setUint32(addr + 4, nanHead, true);
						mem().setUint32(addr, 2, true);
						return;
					case true:
						mem().setUint32(addr + 4, nanHead, true);
						mem().setUint32(addr, 3, true);
						return;
					case false:
						mem().setUint32(addr + 4, nanHead, true);
						mem().setUint32(addr, 4, true);
						return;
				}

				let ref = this._refs.get(v);
				if (ref === undefined) {
					ref = this._values.length;
					this._values.push(v);
					this._refs.set(v, ref);
				}
				let typeFlag = 0;
				switch (typeof v) {
					case "string":
						typeFlag = 1;
						break;
					case "symbol":
						typeFlag = 2;
						break;
					case "function":
						typeFlag = 3;
						break;
				}
				mem().setUint32(addr + 4, nanHead | typeFlag, true);
				mem().setUint32(addr, ref, true);
			}

			const loadSlice = (array, len, cap) => {
				return new Uint8Array(this._inst.exports.memory.buffer, array, len);
			}

			const loadSliceOfValues = (array, len, cap) => {
				const a = new Array(len);
				for (let i = 0; i < len; i++) {
					a[i] = loadValue(array + i * 8);
				}
				return a;
			}

			const loadString = (ptr, len) => {
				return decoder.decode(new DataView(this._inst.exports.memory.buffer, ptr, len));
			}

			const timeOrigin = Date.now() - performance.now();
			this.importObject = {
				wasi_unstable: {
					// https://github.com/bytecodealliance/wasmtime/blob/master/docs/WASI-api.md#__wasi_fd_write
					fd_write: function(fd, iovs_ptr, iovs_len, nwritten_ptr) {
						let nwritten = 0;
						if (fd == 1) {
							for (let iovs_i=0; iovs_i<iovs_len;iovs_i++) {
								let iov_ptr = iovs_ptr+iovs_i*8; // assuming wasm32
								let ptr = mem().getUint32(iov_ptr + 0, true);
								let len = mem().getUint32(iov_ptr + 4, true);
								for (let i=0; i<len; i++) {
									let c = mem().getUint8(ptr+i);
									if (c == 13) { // CR
										// ignore
									} else if (c == 10) { // LF
										// write line
										let line = decoder.decode(new Uint8Array(logLine));
										logLine = [];
										console.log(line);
									} else {
										logLine.push(c);
									}
								}
							}
						} else {
							console.error('invalid file descriptor:', fd);
						}
						mem().setUint32(nwritten_ptr, nwritten, true);
						return 0;
					},
				},
				env: {
					// func ticks() float64
					"runtime.ticks": () => {
						return timeOrigin + performance.now();
					},

					// func sleepTicks(timeout float64)
					"runtime.sleepTicks": (timeout) => {
						// Do not sleep, only reactivate scheduler after the given timeout.
						setTimeout(this._inst.exports.go_scheduler, timeout);
					},

					// func stringVal(value string) ref
					"syscall/js.stringVal": (ret_ptr, value_ptr, value_len) => {
						const s = loadString(value_ptr, value_len);
						storeValue(ret_ptr, s);
					},

					// func valueGet(v ref, p string) ref
					"syscall/js.valueGet": (retval, v_addr, p_ptr, p_len) => {
						let prop = loadString(p_ptr, p_len);
						let value = loadValue(v_addr);
						let result = Reflect.get(value, prop);
						storeValue(retval, result);
					},

					// func valueSet(v ref, p string, x ref)
					"syscall/js.valueSet": (v_addr, p_ptr, p_len, x_addr) => {
						const v = loadValue(v_addr);
						const p = loadString(p_ptr, p_len);
						const x = loadValue(x_addr);
						Reflect.set(v, p, x);
					},

					// func valueIndex(v ref, i int) ref
					"syscall/js.valueIndex": (ret_addr, v_addr, i) => {
						storeValue(ret_addr, Reflect.get(loadValue(v_addr), i));
					},

					// valueSetIndex(v ref, i int, x ref)
					"syscall/js.valueSetIndex": (v_addr, i, x_addr) => {
						Reflect.set(loadValue(v_addr), i, loadValue(x_addr));
					},

					// func valueCall(v ref, m string, args []ref) (ref, bool)
					"syscall/js.valueCall": (ret_addr, v_addr, m_ptr, m_len, args_ptr, args_len, args_cap) => {
						const v = loadValue(v_addr);
						const name = loadString(m_ptr, m_len);
						const args = loadSliceOfValues(args_ptr, args_len, args_cap);
						try {
							const m = Reflect.get(v, name);
							storeValue(ret_addr, Reflect.apply(m, v, args));
							mem().setUint8(ret_addr + 8, 1);
						} catch (err) {
							storeValue(ret_addr, err);
							mem().setUint8(ret_addr + 8, 0);
						}
					},

					// func valueInvoke(v ref, args []ref) (ref, bool)
					"syscall/js.valueInvoke": (ret_addr, v_addr, args_ptr, args_len, args_cap) => {
						try {
							const v = loadValue(v_addr);
							const args = loadSliceOfValues(args_ptr, args_len, args_cap);
							storeValue(ret_addr, Reflect.apply(v, undefined, args));
							mem().setUint8(ret_addr + 8, 1);
						} catch (err) {
							storeValue(ret_addr, err);
							mem().setUint8(ret_addr + 8, 0);
						}
					},

					// func valueNew(v ref, args []ref) (ref, bool)
					"syscall/js.valueNew": (ret_addr, v_addr, args_ptr, args_len, args_cap) => {
						const v = loadValue(v_addr);
						const args = loadSliceOfValues(args_ptr, args_len, args_cap);
						try {
							storeValue(ret_addr, Reflect.construct(v, args));
							mem().setUint8(ret_addr + 8, 1);
						} catch (err) {
							storeValue(ret_addr, err);
							mem().setUint8(ret_addr+ 8, 0);
						}
					},

					// func valueLength(v ref) int
					"syscall/js.valueLength": (v_addr) => {
						return loadValue(v_addr).length;
					},

					// valuePrepareString(v ref) (ref, int)
					"syscall/js.valuePrepareString": (ret_addr, v_addr) => {
						const s = String(loadValue(v_addr));
						const str = encoder.encode(s);
						storeValue(ret_addr, str);
						setInt64(ret_addr + 8, str.length);
					},

					// valueLoadString(v ref, b []byte)
					"syscall/js.valueLoadString": (v_addr, slice_ptr, slice_len, slice_cap) => {
						const str = loadValue(v_addr);
						loadSlice(slice_ptr, slice_len, slice_cap).set(str);
					},

					// func valueInstanceOf(v ref, t ref) bool
					//"syscall/js.valueInstanceOf": (sp) => {
					//	mem().setUint8(sp + 24, loadValue(sp + 8) instanceof loadValue(sp + 16));
					//},
				}
			};
		}

		async run(instance) {
			this._inst = instance;
			this._values = [ // TODO: garbage collection
				NaN,
				0,
				null,
				true,
				false,
				global,
				this._inst.exports.memory,
				this,
			];
			this._refs = new Map();
			this._callbackShutdown = false;
			this.exited = false;

			const mem = new DataView(this._inst.exports.memory.buffer)

			while (true) {
				const callbackPromise = new Promise((resolve) => {
					this._resolveCallbackPromise = () => {
						if (this.exited) {
							throw new Error("bad callback: Go program has already exited");
						}
						setTimeout(resolve, 0); // make sure it is asynchronous
					};
				});
				this._inst.exports._start();
				if (this.exited) {
					break;
				}
				await callbackPromise;
			}
		}

		_resume() {
			if (this.exited) {
				throw new Error("Go program has already exited");
			}
			this._inst.exports.resume();
			if (this.exited) {
				this._resolveExitPromise();
			}
		}

		_makeFuncWrapper(id) {
			const go = this;
			return function () {
				const event = { id: id, this: this, args: arguments };
				go._pendingEvent = event;
				go._resume();
				return event.result;
			};
		}
	}

	if (isNodeJS) {
		if (process.argv.length != 3) {
			process.stderr.write("usage: go_js_wasm_exec [wasm binary] [arguments]\n");
			process.exit(1);
		}

		const go = new Go();
		WebAssembly.instantiate(fs.readFileSync(process.argv[2]), go.importObject).then((result) => {
			process.on("exit", (code) => { // Node.js exits if no callback is pending
				if (code === 0 && !go.exited) {
					// deadlock, make Go print error and stack traces
					go._callbackShutdown = true;
				}
			});
			return go.run(result.instance);
		}).catch((err) => {
			throw err;
		});
	}
})();
