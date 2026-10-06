// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// This file has been modified for use by the TinyGo compiler.
"use strict";

(() => {
	const enosys = () => {
		const err = new Error("not implemented");
		err.code = "ENOSYS";
		return err;
	};

	if (!globalThis.fs) {
		let outputBuf = "";
		globalThis.fs = {
			constants: { O_WRONLY: -1, O_RDWR: -1, O_CREAT: -1, O_TRUNC: -1, O_APPEND: -1, O_EXCL: -1, O_DIRECTORY: -1 }, // unused
			writeSync(fd, buf) {
				outputBuf += decoder.decode(buf);
				const nl = outputBuf.lastIndexOf("\n");
				if (nl != -1) {
					console.log(outputBuf.substring(0, nl));
					outputBuf = outputBuf.substring(nl + 1);
				}
				return buf.length;
			},
			write(fd, buf, offset, length, position, callback) {
				if (offset !== 0 || length !== buf.length || position !== null) {
					callback(enosys());
					return;
				}
				const n = this.writeSync(fd, buf);
				callback(null, n);
			},
			chmod(path, mode, callback) { callback(enosys()); },
			chown(path, uid, gid, callback) { callback(enosys()); },
			close(fd, callback) { callback(enosys()); },
			fchmod(fd, mode, callback) { callback(enosys()); },
			fchown(fd, uid, gid, callback) { callback(enosys()); },
			fstat(fd, callback) { callback(enosys()); },
			fsync(fd, callback) { callback(null); },
			ftruncate(fd, length, callback) { callback(enosys()); },
			lchown(path, uid, gid, callback) { callback(enosys()); },
			link(path, link, callback) { callback(enosys()); },
			lstat(path, callback) { callback(enosys()); },
			mkdir(path, perm, callback) { callback(enosys()); },
			open(path, flags, mode, callback) { callback(enosys()); },
			read(fd, buffer, offset, length, position, callback) { callback(enosys()); },
			readdir(path, callback) { callback(enosys()); },
			readlink(path, callback) { callback(enosys()); },
			rename(from, to, callback) { callback(enosys()); },
			rmdir(path, callback) { callback(enosys()); },
			stat(path, callback) { callback(enosys()); },
			symlink(path, link, callback) { callback(enosys()); },
			truncate(path, length, callback) { callback(enosys()); },
			unlink(path, callback) { callback(enosys()); },
			utimes(path, atime, mtime, callback) { callback(enosys()); },
		};
	}

	if (!globalThis.process) {
		globalThis.process = {
			getuid() { return -1; },
			getgid() { return -1; },
			geteuid() { return -1; },
			getegid() { return -1; },
			getgroups() { throw enosys(); },
			pid: -1,
			ppid: -1,
			umask() { throw enosys(); },
			cwd() { throw enosys(); },
			chdir() { throw enosys(); },
		}
	}

	if (!globalThis.crypto) {
		throw new Error("globalThis.crypto is not available, polyfill required (crypto.getRandomValues only)");
	}

	if (!globalThis.performance) {
		throw new Error("globalThis.performance is not available, polyfill required (performance.now only)");
	}

	if (!globalThis.TextEncoder) {
		throw new Error("globalThis.TextEncoder is not available, polyfill required");
	}

	if (!globalThis.TextDecoder) {
		throw new Error("globalThis.TextDecoder is not available, polyfill required");
	}

	const encoder = new TextEncoder("utf-8");
	const decoder = new TextDecoder("utf-8");
	let reinterpretBuf = new DataView(new ArrayBuffer(8));
	var logLines = { 1: [], 2: [] }; // unfinished stdout and stderr lines
	const wasmExit = {}; // thrown to exit via proc_exit (not an error)

	// WASI snapshot_preview1 errno values and the jsfs error codes they answer.
	// https://github.com/WebAssembly/WASI/blob/snapshot-01/phases/snapshot/docs.md#-errno-enumu16
	const wasiErrno = {
		E2BIG: 1, EACCES: 2, EAGAIN: 6, EBADF: 8, EBUSY: 10, EEXIST: 20, EFAULT: 21,
		EFBIG: 22, EINTR: 27, EINVAL: 28, EIO: 29, EISDIR: 31, ELOOP: 32, EMFILE: 33,
		ENAMETOOLONG: 37, ENODEV: 43, ENOENT: 44, ENOMEM: 48, ENOSPC: 51, ENOSYS: 52,
		ENOTDIR: 54, ENOTEMPTY: 55, ENOTSUP: 58, EOPNOTSUPP: 58, ENOTTY: 59, EPERM: 63,
		EPIPE: 64, EROFS: 69, ESPIPE: 70, EXDEV: 75,
	};

	// jsfsWasi serves the WASI filesystem imports from bottle's jsfs.sync, so os
	// works on one in-memory tree shared with the page. The only preopen is "/".
	const jsfsWasi = (jsfs, buffer) => {
		const sync = jsfs.sync;
		const C = sync.constants;
		const enc = new TextEncoder();
		const dec = new TextDecoder();
		const u8 = () => new Uint8Array(buffer());
		const dv = () => new DataView(buffer());
		const str = (ptr, len) => dec.decode(new Uint8Array(buffer(), ptr >>> 0, len >>> 0));
		const errno = (e) => wasiErrno[e && e.code] || 29;
		const call = (f) => { try { return f() || 0; } catch (e) { return errno(e); } };
		const fail = (code) => { const e = new Error(code); e.code = code; throw e; };

		// Rights are not enforced here; jsfs checks access when a file is opened.
		const ALL = 0x1fffffffn, TTY = ALL & ~((1n << 2n) | (1n << 5n));
		const FD_READ = 1n << 1n, FD_WRITE = 1n << 6n;
		const OFLAG_CREAT = 1, OFLAG_DIRECTORY = 2, OFLAG_EXCL = 4, OFLAG_TRUNC = 8;
		const FDFLAG_APPEND = 1, LOOKUP_FOLLOW = 1;
		const FT_CHR = 2, FT_DIR = 3, FT_REG = 4, FT_LNK = 7;

		const filetype = (mode) => {
			switch (mode & 0o170000) {
			case 0o060000: return 1;
			case 0o020000: return FT_CHR;
			case 0o040000: return FT_DIR;
			case 0o100000: return FT_REG;
			case 0o140000: return 6;
			case 0o120000: return FT_LNK;
			}
			return 0;
		};
		const ns = (ms) => BigInt(Math.round((ms || 0) * 1e6));

		// WASI fd -> {kind, path, jfd, pos, append}. Files hold a jsfs fd and keep
		// their own offset, so append and seek work without reaching into jsfs.
		const fds = new Map([
			[0, { kind: "stdio" }], [1, { kind: "stdio" }], [2, { kind: "stdio" }],
			[3, { kind: "dir", path: "/" }],
		]);
		const entry = (fd) => fds.get(fd) || fail("EBADF");
		const file = (fd) => {
			const e = entry(fd);
			if (e.kind === "dir") fail("EISDIR");
			if (e.kind !== "file") fail("ESPIPE");
			return e;
		};
		const add = (e) => {
			let fd = 4;
			while (fds.has(fd)) fd++;
			fds.set(fd, e);
			return fd;
		};
		const join = (dir, rel) => (rel.startsWith("/") ? rel : (dir === "/" ? "" : dir) + "/" + rel);
		const at = (dirfd, ptr, len) => {
			const d = entry(dirfd);
			if (d.kind !== "dir") fail("ENOTDIR");
			return join(d.path, str(ptr, len));
		};

		const putFilestat = (ptr, st) => {
			const v = dv();
			ptr >>>= 0;
			v.setBigUint64(ptr, BigInt(st.dev || 0), true);
			v.setBigUint64(ptr + 8, BigInt(st.ino || 0), true);
			v.setUint8(ptr + 16, filetype(st.mode));
			v.setBigUint64(ptr + 24, BigInt(st.nlink || 1), true);
			v.setBigUint64(ptr + 32, BigInt(st.size || 0), true);
			v.setBigUint64(ptr + 40, ns(st.atimeMs), true);
			v.setBigUint64(ptr + 48, ns(st.mtimeMs), true);
			v.setBigUint64(ptr + 56, ns(st.ctimeMs), true);
		};
		const statOf = (e) => {
			if (e.kind === "stdio") return { mode: 0o020000, nlink: 1 };
			if (e.kind === "dir") return sync.stat(e.path);
			return sync.fstat(e.jfd);
		};

		// iovs walks a ciovec/iovec array, calling f(ptr, len) until it returns
		// fewer bytes than asked for.
		const iovs = (ptr, n, f) => {
			const v = dv();
			let total = 0;
			for (let i = 0; i < n; i++) {
				const p = v.getUint32(ptr + i * 8, true), l = v.getUint32(ptr + i * 8 + 4, true);
				const got = f(p, l);
				total += got;
				if (got < l) break;
			}
			return total;
		};

		// Stdin keeps what a pulled chunk had beyond the read, rather than drop it.
		let stdinRest = null;
		const readStdin = (p, l) => {
			let chunk = stdinRest;
			stdinRest = null;
			if (!chunk) chunk = jsfs.stdio.stdin();
			if (!chunk || chunk.length === 0) return 0;
			const n = Math.min(l, chunk.length);
			u8().set(chunk.subarray(0, n), p);
			if (n < chunk.length) stdinRest = chunk.slice(n);
			return n;
		};

		const readAt = (e, p, l, pos) => sync.read(e.jfd, u8(), p, l, pos);
		const writeAt = (e, p, l, pos) => sync.write(e.jfd, u8(), p, l, pos);
		const writeOut = (fd, p, l) => globalThis.fs.writeSync(fd, u8().slice(p, p + l));

		const setTimes = (path, follow, atim, mtim, flags) => {
			const st = follow ? sync.stat(path) : sync.lstat(path);
			const nowS = Date.now() / 1000;
			let a = st.atimeMs / 1000, m = st.mtimeMs / 1000;
			if (flags & 1) a = Number(atim) / 1e9;
			if (flags & 2) a = nowS;
			if (flags & 4) m = Number(mtim) / 1e9;
			if (flags & 8) m = nowS;
			sync.utimes(path, a, m);
		};

		return {
			fd_prestat_get: (fd, ptr) => {
				if (fd !== 3 || !fds.has(3)) return wasiErrno.EBADF;
				dv().setUint8(ptr >>> 0, 0);
				dv().setUint32((ptr >>> 0) + 4, 1, true);
				return 0;
			},
			fd_prestat_dir_name: (fd, ptr, len) => {
				if (fd !== 3 || !fds.has(3)) return wasiErrno.EBADF;
				if (len < 1) return wasiErrno.ENAMETOOLONG;
				u8()[ptr >>> 0] = 0x2f;
				return 0;
			},

			path_open: (dirfd, dirflags, pathPtr, pathLen, oflags, rightsBase, rightsInh, fdflags, fdPtr) => call(() => {
				const path = at(dirfd, pathPtr, pathLen);
				const follow = (dirflags & LOOKUP_FOLLOW) !== 0;
				const rd = (rightsBase & FD_READ) !== 0n, wr = (rightsBase & FD_WRITE) !== 0n;
				let st = null;
				try { st = follow ? sync.stat(path) : sync.lstat(path); } catch (e) { if (e.code !== "ENOENT") throw e; }
				if (st && (oflags & OFLAG_CREAT) && (oflags & OFLAG_EXCL)) fail("EEXIST");
				if (oflags & OFLAG_DIRECTORY) {
					if (!st) fail("ENOENT");
					if (filetype(st.mode) !== FT_DIR) fail("ENOTDIR");
				}
				let fd;
				if (st && filetype(st.mode) === FT_DIR) {
					if (wr || (oflags & OFLAG_TRUNC)) fail("EISDIR");
					fd = add({ kind: "dir", path });
				} else {
					if (st && filetype(st.mode) === FT_LNK) fail("ELOOP");
					let flags = wr ? (rd ? C.O_RDWR : C.O_WRONLY) : C.O_RDONLY;
					if (oflags & OFLAG_CREAT) flags |= C.O_CREAT;
					if (oflags & OFLAG_EXCL) flags |= C.O_EXCL;
					if (oflags & OFLAG_TRUNC) flags |= C.O_TRUNC;
					const jfd = sync.open(path, flags, 0o644);
					fd = add({ kind: "file", path, jfd, pos: 0, append: (fdflags & FDFLAG_APPEND) !== 0 });
				}
				dv().setUint32(fdPtr >>> 0, fd, true);
			}),

			fd_close: (fd) => call(() => {
				const e = entry(fd);
				if (e.kind === "file") sync.close(e.jfd);
				if (e.kind !== "stdio") fds.delete(fd);
			}),

			fd_read: (fd, iovsPtr, iovsLen, nPtr) => call(() => {
				const e = entry(fd);
				let n;
				if (e.kind === "stdio") n = fd === 0 ? iovs(iovsPtr >>> 0, iovsLen, readStdin) : fail("EBADF");
				else n = iovs(iovsPtr >>> 0, iovsLen, (p, l) => {
					const got = readAt(file(fd), p, l, e.pos);
					e.pos += got;
					return got;
				});
				dv().setUint32(nPtr >>> 0, n, true);
			}),
			fd_pread: (fd, iovsPtr, iovsLen, offset, nPtr) => call(() => {
				const e = file(fd);
				let pos = Number(offset);
				const n = iovs(iovsPtr >>> 0, iovsLen, (p, l) => {
					const got = readAt(e, p, l, pos);
					pos += got;
					return got;
				});
				dv().setUint32(nPtr >>> 0, n, true);
			}),
			fd_write: (fd, iovsPtr, iovsLen, nPtr) => call(() => {
				const e = entry(fd);
				let n;
				if (e.kind === "stdio") n = fd === 0 ? fail("EBADF") : iovs(iovsPtr >>> 0, iovsLen, (p, l) => writeOut(fd, p, l));
				else n = iovs(iovsPtr >>> 0, iovsLen, (p, l) => {
					const f = file(fd);
					const pos = f.append ? sync.fstat(f.jfd).size : f.pos;
					const got = writeAt(f, p, l, pos);
					f.pos = pos + got;
					return got;
				});
				dv().setUint32(nPtr >>> 0, n, true);
			}),
			fd_pwrite: (fd, iovsPtr, iovsLen, offset, nPtr) => call(() => {
				const e = file(fd);
				let pos = Number(offset);
				const n = iovs(iovsPtr >>> 0, iovsLen, (p, l) => {
					const got = writeAt(e, p, l, pos);
					pos += got;
					return got;
				});
				dv().setUint32(nPtr >>> 0, n, true);
			}),
			fd_seek: (fd, offset, whence, ptr) => call(() => {
				const e = file(fd);
				const base = whence === 0 ? 0 : whence === 1 ? e.pos : whence === 2 ? sync.fstat(e.jfd).size : fail("EINVAL");
				const pos = base + Number(offset);
				if (pos < 0) fail("EINVAL");
				e.pos = pos;
				dv().setBigUint64(ptr >>> 0, BigInt(pos), true);
			}),
			fd_tell: (fd, ptr) => call(() => { dv().setBigUint64(ptr >>> 0, BigInt(file(fd).pos), true); }),

			fd_fdstat_get: (fd, ptr) => call(() => {
				const e = entry(fd);
				const v = dv();
				ptr >>>= 0;
				v.setUint8(ptr, e.kind === "stdio" ? FT_CHR : e.kind === "dir" ? FT_DIR : filetype(statOf(e).mode));
				v.setUint16(ptr + 2, e.append ? FDFLAG_APPEND : 0, true);
				v.setBigUint64(ptr + 8, e.kind === "stdio" ? TTY : ALL, true);
				v.setBigUint64(ptr + 16, ALL, true);
			}),
			fd_fdstat_set_flags: (fd, flags) => call(() => {
				const e = entry(fd);
				if (e.kind === "file") e.append = (flags & FDFLAG_APPEND) !== 0;
			}),
			fd_fdstat_set_rights: (fd) => call(() => { entry(fd); }),
			fd_filestat_get: (fd, ptr) => call(() => { putFilestat(ptr, statOf(entry(fd))); }),
			fd_filestat_set_size: (fd, size) => call(() => { sync.ftruncate(file(fd).jfd, Number(size)); }),
			fd_filestat_set_times: (fd, atim, mtim, flags) => call(() => {
				const e = entry(fd);
				if (e.kind === "stdio") fail("EBADF");
				setTimes(e.path, true, atim, mtim, flags);
			}),
			fd_sync: (fd) => call(() => { const e = entry(fd); if (e.kind === "file") sync.fsync(e.jfd); }),
			fd_datasync: (fd) => call(() => { const e = entry(fd); if (e.kind === "file") sync.fsync(e.jfd); }),
			fd_advise: (fd) => call(() => { entry(fd); }),
			fd_allocate: (fd, offset, len) => call(() => {
				const e = file(fd);
				const end = Number(offset) + Number(len);
				if (sync.fstat(e.jfd).size < end) sync.ftruncate(e.jfd, end);
			}),
			fd_renumber: (from, to) => call(() => {
				const e = entry(from);
				const old = entry(to);
				if (old.kind === "file") sync.close(old.jfd);
				fds.set(to, e);
				fds.delete(from);
			}),

			// Entries are ".", ".." and the directory's names in jsfs order. A cookie
			// is the index of the next entry, so a listing resumes where it stopped.
			fd_readdir: (fd, bufPtr, bufLen, cookie, usedPtr) => call(() => {
				const e = entry(fd);
				if (e.kind !== "dir") fail("ENOTDIR");
				const names = [".", ".."].concat(sync.readdir(e.path));
				const parent = e.path === "/" ? "/" : e.path.slice(0, e.path.lastIndexOf("/")) || "/";
				bufPtr >>>= 0;
				let used = 0;
				for (let i = Number(cookie); i < names.length && used < bufLen; i++) {
					const name = names[i];
					let st;
					try {
						st = name === "." ? sync.stat(e.path) : name === ".." ? sync.stat(parent) : sync.lstat(join(e.path, name));
					} catch (err) {
						st = { ino: 0, mode: 0 };
					}
					const nb = enc.encode(name);
					const rec = new Uint8Array(24 + nb.length);
					const rv = new DataView(rec.buffer);
					rv.setBigUint64(0, BigInt(i + 1), true);
					rv.setBigUint64(8, BigInt(st.ino || 0), true);
					rv.setUint32(16, nb.length, true);
					rv.setUint8(20, filetype(st.mode));
					rec.set(nb, 24);
					const n = Math.min(rec.length, bufLen - used);
					u8().set(rec.subarray(0, n), bufPtr + used);
					used += n;
				}
				dv().setUint32(usedPtr >>> 0, used, true);
			}),

			path_filestat_get: (dirfd, flags, pathPtr, pathLen, ptr) => call(() => {
				const path = at(dirfd, pathPtr, pathLen);
				putFilestat(ptr, (flags & LOOKUP_FOLLOW) ? sync.stat(path) : sync.lstat(path));
			}),
			path_filestat_set_times: (dirfd, flags, pathPtr, pathLen, atim, mtim, fstFlags) => call(() => {
				setTimes(at(dirfd, pathPtr, pathLen), (flags & LOOKUP_FOLLOW) !== 0, atim, mtim, fstFlags);
			}),
			path_create_directory: (dirfd, pathPtr, pathLen) => call(() => { sync.mkdir(at(dirfd, pathPtr, pathLen), 0o755); }),
			path_remove_directory: (dirfd, pathPtr, pathLen) => call(() => { sync.rmdir(at(dirfd, pathPtr, pathLen)); }),
			path_unlink_file: (dirfd, pathPtr, pathLen) => call(() => { sync.unlink(at(dirfd, pathPtr, pathLen)); }),
			path_rename: (fd1, p1, l1, fd2, p2, l2) => call(() => { sync.rename(at(fd1, p1, l1), at(fd2, p2, l2)); }),
			path_link: (fd1, flags, p1, l1, fd2, p2, l2) => call(() => { sync.link(at(fd1, p1, l1), at(fd2, p2, l2)); }),
			path_symlink: (tPtr, tLen, dirfd, pathPtr, pathLen) => call(() => { sync.symlink(str(tPtr, tLen), at(dirfd, pathPtr, pathLen)); }),
			path_readlink: (dirfd, pathPtr, pathLen, bufPtr, bufLen, usedPtr) => call(() => {
				const b = enc.encode(sync.readlink(at(dirfd, pathPtr, pathLen)));
				const n = Math.min(b.length, bufLen >>> 0);
				u8().set(b.subarray(0, n), bufPtr >>> 0);
				dv().setUint32(usedPtr >>> 0, n, true);
			}),
		};
	};

	globalThis.Go = class {
		constructor() {
			this.argv = ["js"];
			this.env = {};
			this._callbackTimeouts = new Map();
			this._nextCallbackTimeoutID = 1;

			const mem = () => {
				// The buffer may change when requesting more memory.
				return new DataView(this._inst.exports.memory.buffer);
			}

			const unboxValue = (v_ref) => {
				reinterpretBuf.setBigInt64(0, v_ref, true);
				const f = reinterpretBuf.getFloat64(0, true);
				if (f === 0) {
					return undefined;
				}
				if (!isNaN(f)) {
					return f;
				}

				const id = v_ref & 0xffffffffn;
				return this._values[id];
			}


			const loadValue = (addr) => {
				let v_ref = mem().getBigUint64(addr, true);
				return unboxValue(v_ref);
			}

			const boxValue = (v) => {
				const nanHead = 0x7FF80000n;

				if (typeof v === "number") {
					if (isNaN(v)) {
						return nanHead << 32n;
					}
					if (v === 0) {
						return (nanHead << 32n) | 1n;
					}
					reinterpretBuf.setFloat64(0, v, true);
					return reinterpretBuf.getBigInt64(0, true);
				}

				switch (v) {
					case undefined:
						return 0n;
					case null:
						return (nanHead << 32n) | 2n;
					case true:
						return (nanHead << 32n) | 3n;
					case false:
						return (nanHead << 32n) | 4n;
				}

				let id = this._ids.get(v);
				if (id === undefined) {
					id = this._idPool.pop();
					if (id === undefined) {
						id = BigInt(this._values.length);
					}
					this._values[id] = v;
					this._goRefCounts[id] = 0;
					this._ids.set(v, id);
				}
				this._goRefCounts[id]++;
				let typeFlag = 1n;
				switch (typeof v) {
					case "string":
						typeFlag = 2n;
						break;
					case "symbol":
						typeFlag = 3n;
						break;
					case "function":
						typeFlag = 4n;
						break;
				}
				return id | ((nanHead | typeFlag) << 32n);
			}

			const storeValue = (addr, v) => {
				let v_ref = boxValue(v);
				mem().setBigUint64(addr, v_ref, true);
			}

			// writeStrings copies list, joined by NUL, into memory when it fits in n
			// bytes, and returns its length so the caller can size a buffer.
			const writeStrings = (list, buf, n) => {
				if (!Array.isArray(list) || list.length === 0) {
					return 0;
				}
				const bytes = encoder.encode(list.join("\0"));
				buf >>>= 0;
				n >>>= 0;
				if (bytes.length <= n) {
					new Uint8Array(this._inst.exports.memory.buffer, buf, bytes.length).set(bytes);
				}
				return bytes.length;
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

			const wasi_EBADF = 8;
			const wasi_ENOSYS = 52;
			this.importObject = {
				wasi_snapshot_preview1: {
					// https://github.com/WebAssembly/WASI/blob/snapshot-01/phases/snapshot/docs.md
					fd_write: function(fd, iovs_ptr, iovs_len, nwritten_ptr) {
						iovs_ptr >>>= 0;
						iovs_len >>>= 0;
						nwritten_ptr >>>= 0;
						let nwritten = 0;
						if (fd == 1 || fd == 2) {
							const logLine = logLines[fd];
							for (let iovs_i = 0; iovs_i < iovs_len; iovs_i++) {
								let iov_ptr = iovs_ptr + iovs_i * 8; // assuming wasm32
								let ptr = mem().getUint32(iov_ptr + 0, true);
								let len = mem().getUint32(iov_ptr + 4, true);
								nwritten += len;
								for (let i = 0; i < len; i++) {
									let c = mem().getUint8(ptr + i);
									if (c == 13) { // CR
										// ignore
									} else if (c == 10) { // LF
										// write line
										let line = decoder.decode(new Uint8Array(logLine));
										logLine.length = 0;
										console.log(line); // stderr too, as when it went through putchar
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
					fd_read: () => wasi_ENOSYS,
					fd_close: () => wasi_ENOSYS,
					fd_fdstat_get: () => wasi_ENOSYS,
					fd_prestat_get: () => wasi_EBADF, // wasi-libc relies on this errno value
					fd_prestat_dir_name: () => wasi_ENOSYS,
					fd_seek: () => wasi_ENOSYS,
					path_open: () => wasi_ENOSYS,
					proc_exit: (code) => {
						this.exited = true;
						this.exitCode = code;
						this._resolveExitPromise();
						throw wasmExit;
					},
					random_get: (bufPtr, bufLen) => {
						bufPtr >>>= 0;
						bufLen >>>= 0;
						crypto.getRandomValues(loadSlice(bufPtr, bufLen));
						return 0;
					},
				},
				gojs: {
					// func argvString(buf unsafe.Pointer, n uint32) uint32
					"runtime.argvString": (buf, n) => writeStrings(this.argv, buf, n),

					// func envString(buf unsafe.Pointer, n uint32) uint32
					"runtime.envString": (buf, n) => writeStrings(Object.entries(this.env || {}).map(([k, v]) => k + "=" + v), buf, n),

					// func ticks() int64
					"runtime.ticks": () => {
						return BigInt(Math.floor(performance.now() * 1e6));
					},

					"runtime.walltime": () => BigInt(Date.now()) * 1000000n,

					// func getRandomData(r []byte)
					"runtime.getRandomData": (slice_ptr, slice_len, slice_cap) => {
						slice_ptr >>>= 0;
						slice_len >>>= 0;
						crypto.getRandomValues(loadSlice(slice_ptr, slice_len, slice_cap));
					},

					// func sleepTicks(timeout int64)
					"runtime.sleepTicks": (timeout) => {
						// Do not sleep, only reactivate the scheduler after the given
						// timeout, keeping exactly one pending wakeup.
						// Round up to a whole millisecond. Some hosts (Cloudflare
						// Workers) only advance the clock by a timer's delay rounded
						// down to a millisecond, so a fractional remainder would be
						// slept again and again without the clock ever moving.
						const ms = Math.ceil(Number(timeout) / 1e6);
						const due = performance.now() + ms;
						if (this._scheduledWakeup !== undefined) {
							if (this._scheduledWakeupDue <= due) return;
							clearTimeout(this._scheduledWakeup);
						}
						this._scheduledWakeupDue = due;
						this._scheduledWakeup = setTimeout(() => {
							this._scheduledWakeup = undefined;
							this._wake();
						}, ms);
					},

					// func finalizeRef(v ref)
					"syscall/js.finalizeRef": (v_ref) => {
						// Note: TinyGo does not support finalizers so this is only called
						// for one specific case, by js.go:jsString. and can/might leak memory.
						const id = v_ref & 0xffffffffn;
						if (this._goRefCounts?.[id] !== undefined) {
							this._goRefCounts[id]--;
							if (this._goRefCounts[id] === 0) {
								const v = this._values[id];
								this._values[id] = null;
								this._ids.delete(v);
								this._idPool.push(id);
							}
						} else {
							console.error("syscall/js.finalizeRef: unknown id", id);
						}
					},

					// func stringVal(value string) ref
					"syscall/js.stringVal": (value_ptr, value_len) => {
						value_ptr >>>= 0;
						value_len >>>= 0;
						const s = loadString(value_ptr, value_len);
						return boxValue(s);
					},

					// func valueGet(v ref, p string) ref
					"syscall/js.valueGet": (v_ref, p_ptr, p_len) => {
						p_ptr >>>= 0;
						p_len >>>= 0;
						let prop = loadString(p_ptr, p_len);
						let v = unboxValue(v_ref);
						let result = Reflect.get(v, prop);
						return boxValue(result);
					},

					// func valueSet(v ref, p string, x ref)
					"syscall/js.valueSet": (v_ref, p_ptr, p_len, x_ref) => {
						p_ptr >>>= 0;
						p_len >>>= 0;
						const v = unboxValue(v_ref);
						const p = loadString(p_ptr, p_len);
						const x = unboxValue(x_ref);
						Reflect.set(v, p, x);
					},

					// func valueDelete(v ref, p string)
					"syscall/js.valueDelete": (v_ref, p_ptr, p_len) => {
						p_ptr >>>= 0;
						p_len >>>= 0;
						const v = unboxValue(v_ref);
						const p = loadString(p_ptr, p_len);
						Reflect.deleteProperty(v, p);
					},

					// func valueIndex(v ref, i int) ref
					"syscall/js.valueIndex": (v_ref, i) => {
						return boxValue(Reflect.get(unboxValue(v_ref), i));
					},

					// valueSetIndex(v ref, i int, x ref)
					"syscall/js.valueSetIndex": (v_ref, i, x_ref) => {
						Reflect.set(unboxValue(v_ref), i, unboxValue(x_ref));
					},

					// func valueCall(v ref, m string, args []ref) (ref, bool)
					"syscall/js.valueCall": (ret_addr, v_ref, m_ptr, m_len, args_ptr, args_len, args_cap) => {
						ret_addr >>>= 0;
						m_ptr >>>= 0;
						m_len >>>= 0;
						args_ptr >>>= 0;
						args_len >>>= 0;
						const v = unboxValue(v_ref);
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
					"syscall/js.valueInvoke": (ret_addr, v_ref, args_ptr, args_len, args_cap) => {
						ret_addr >>>= 0;
						args_ptr >>>= 0;
						args_len >>>= 0;
						try {
							const v = unboxValue(v_ref);
							const args = loadSliceOfValues(args_ptr, args_len, args_cap);
							storeValue(ret_addr, Reflect.apply(v, undefined, args));
							mem().setUint8(ret_addr + 8, 1);
						} catch (err) {
							storeValue(ret_addr, err);
							mem().setUint8(ret_addr + 8, 0);
						}
					},

					// func valueNew(v ref, args []ref) (ref, bool)
					"syscall/js.valueNew": (ret_addr, v_ref, args_ptr, args_len, args_cap) => {
						ret_addr >>>= 0;
						args_ptr >>>= 0;
						args_len >>>= 0;
						const v = unboxValue(v_ref);
						const args = loadSliceOfValues(args_ptr, args_len, args_cap);
						try {
							storeValue(ret_addr, Reflect.construct(v, args));
							mem().setUint8(ret_addr + 8, 1);
						} catch (err) {
							storeValue(ret_addr, err);
							mem().setUint8(ret_addr + 8, 0);
						}
					},

					// func valueLength(v ref) int
					"syscall/js.valueLength": (v_ref) => {
						return unboxValue(v_ref).length;
					},

					// valuePrepareString(v ref) (ref, int)
					"syscall/js.valuePrepareString": (ret_addr, v_ref) => {
						ret_addr >>>= 0;
						const s = String(unboxValue(v_ref));
						const str = encoder.encode(s);
						storeValue(ret_addr, str);
						mem().setInt32(ret_addr + 8, str.length, true);
					},

					// valueLoadString(v ref, b []byte)
					"syscall/js.valueLoadString": (v_ref, slice_ptr, slice_len, slice_cap) => {
						slice_ptr >>>= 0;
						slice_len >>>= 0;
						const str = unboxValue(v_ref);
						loadSlice(slice_ptr, slice_len, slice_cap).set(str);
					},

					// func valueInstanceOf(v ref, t ref) bool
					"syscall/js.valueInstanceOf": (v_ref, t_ref) => {
						return unboxValue(v_ref) instanceof unboxValue(t_ref);
					},

					// func copyBytesToGo(dst []byte, src ref) (int, bool)
					"syscall/js.copyBytesToGo": (ret_addr, dest_addr, dest_len, dest_cap, src_ref) => {
						ret_addr >>>= 0;
						dest_addr >>>= 0;
						dest_len >>>= 0;
						let num_bytes_copied_addr = ret_addr;
						let returned_status_addr = ret_addr + 4; // Address of returned boolean status variable

						const dst = loadSlice(dest_addr, dest_len);
						const src = unboxValue(src_ref);
						if (!(src instanceof Uint8Array || src instanceof Uint8ClampedArray)) {
							mem().setUint8(returned_status_addr, 0); // Return "not ok" status
							return;
						}
						const toCopy = src.subarray(0, dst.length);
						dst.set(toCopy);
						mem().setUint32(num_bytes_copied_addr, toCopy.length, true);
						mem().setUint8(returned_status_addr, 1); // Return "ok" status
					},

					// copyBytesToJS(dst ref, src []byte) (int, bool)
					// Originally copied from upstream Go project, then modified:
					//   https://github.com/golang/go/blob/3f995c3f3b43033013013e6c7ccc93a9b1411ca9/misc/wasm/wasm_exec.js#L404-L416
					"syscall/js.copyBytesToJS": (ret_addr, dst_ref, src_addr, src_len, src_cap) => {
						ret_addr >>>= 0;
						src_addr >>>= 0;
						src_len >>>= 0;
						let num_bytes_copied_addr = ret_addr;
						let returned_status_addr = ret_addr + 4; // Address of returned boolean status variable

						const dst = unboxValue(dst_ref);
						const src = loadSlice(src_addr, src_len);
						if (!(dst instanceof Uint8Array || dst instanceof Uint8ClampedArray)) {
							mem().setUint8(returned_status_addr, 0); // Return "not ok" status
							return;
						}
						const toCopy = src.subarray(0, dst.length);
						dst.set(toCopy);
						mem().setUint32(num_bytes_copied_addr, toCopy.length, true);
						mem().setUint8(returned_status_addr, 1); // Return "ok" status
					},
				}
			};

			// Go 1.20 uses 'env'. Go 1.21 uses 'gojs'.
			// For compatibility, we use both as long as Go 1.20 is supported.
			this.importObject.env = this.importObject.gojs;

			// With bottle's jsfs on the page, os reaches its filesystem through the
			// WASI imports. Otherwise there are no preopens and every open fails.
			const wasi = this.importObject.wasi_snapshot_preview1;
			this._jsfs = globalThis.jsfs && globalThis.jsfs.sync ? globalThis.jsfs : null;
			if (this._jsfs) {
				Object.assign(wasi, jsfsWasi(this._jsfs, () => this._inst.exports.memory.buffer));
			}
			// wasi-libc exits when the environment sizes fail, so they answer empty.
			const sizes = (a, b) => { mem().setUint32(a >>> 0, 0, true); mem().setUint32(b >>> 0, 0, true); return 0; };
			const stubs = {
				args_sizes_get: sizes, args_get: () => 0,
				environ_sizes_get: sizes, environ_get: () => 0,
				clock_time_get: (id, precision, ptr) => {
					const t = id === 0 ? BigInt(Date.now()) * 1000000n : BigInt(Math.round((timeOrigin + performance.now()) * 1e6));
					mem().setBigUint64(ptr >>> 0, t, true);
					return 0;
				},
				sched_yield: () => 0,
			};
			for (const name of ["args_get", "args_sizes_get", "environ_get", "environ_sizes_get",
				"clock_res_get", "clock_time_get", "fd_advise", "fd_allocate", "fd_close",
				"fd_datasync", "fd_fdstat_get", "fd_fdstat_set_flags", "fd_fdstat_set_rights",
				"fd_filestat_get", "fd_filestat_set_size", "fd_filestat_set_times", "fd_pread",
				"fd_prestat_dir_name", "fd_prestat_get", "fd_pwrite", "fd_read", "fd_readdir",
				"fd_renumber", "fd_seek", "fd_sync", "fd_tell", "fd_write", "path_create_directory",
				"path_filestat_get", "path_filestat_set_times", "path_link", "path_open",
				"path_readlink", "path_remove_directory", "path_rename", "path_symlink",
				"path_unlink_file", "poll_oneoff", "proc_raise", "sched_yield", "sock_accept",
				"sock_recv", "sock_send", "sock_shutdown"]) {
				if (!wasi[name]) wasi[name] = stubs[name] || (() => wasi_ENOSYS);
			}
		}

		async run(instance) {
			this._inst = instance;
			this._values = [ // JS values that Go currently has references to, indexed by reference id
				NaN,
				0,
				null,
				true,
				false,
				globalThis,
				this,
			];
			this._goRefCounts = []; // number of references that Go has to a JS value, indexed by reference id
			this._ids = new Map();  // mapping from JS values to reference ids
			this._idPool = [];      // unused ids that have been garbage collected
			this.exited = false;    // whether the Go program has exited
			// A wakeup left pending by a previous run would otherwise suppress the
			// first one this run asks for, and the scheduler would never start.
			if (this._scheduledWakeup !== undefined) {
				clearTimeout(this._scheduledWakeup);
				this._scheduledWakeup = undefined;
			}
			this.exitCode = 0;
			// syscall/js.handleEvent reads _pendingEvent and returns early only when
			// it IsNull(). Leaving it `undefined` is not null, so handleEvent falls
			// through to cb.Get("id") and panics with "call of Value.Get on
			// undefined", which surfaces as RuntimeError: unreachable. Upstream Go's
			// wasm_exec.js initializes this in its constructor; this line restores
			// parity. Any resume() that runs without a pending event -- and resume()
			// always spawns a handleEvent goroutine -- depends on it.
			this._pendingEvent = null; // event awaiting dispatch by syscall/js.handleEvent
			// os starts in $PWD. Unless the caller chose one, that is the jsfs
			// working directory, which proc.js sets for each process it spawns.
			if (this._jsfs && !("PWD" in (this.env || {}))) {
				this.env = Object.assign({}, this.env, { PWD: this._jsfs.getCwd() });
			}

			if (this._inst.exports._start) {
				let exitPromise = new Promise((resolve, reject) => {
					this._resolveExitPromise = resolve;
				});

				// Run program, but catch the wasmExit exception that's thrown
				// to return back here.
				try {
					this._inst.exports._start();
				} catch (e) {
					if (e !== wasmExit) throw e;
				}

				await exitPromise;
				return this.exitCode;
			} else {
				this._inst.exports._initialize();
			}
		}

		// _wake runs the scheduler for a timer the program asked for. It is a
		// method so a host such as proc.js can wrap it, as it wraps _resume.
		_wake() {
			if (this.exited) return;
			try {
				this._inst.exports.go_scheduler();
			} catch (e) {
				if (e !== wasmExit) throw e;
			}
		}

		_resume() {
			if (this.exited) {
				throw new Error("Go program has already exited");
			}
			try {
				this._inst.exports.resume();
			} catch (e) {
				if (e !== wasmExit) throw e;
			}
			if (this.exited) {
				this._resolveExitPromise();
			}
		}

		_makeFuncWrapper(id) {
			const go = this;
			return function() {
				const event = { id: id, this: this, args: arguments };
				go._pendingEvent = event;
				go._resume();
				return event.result;
			};
		}
	}
})();
