// Ensure Go WASM runtime class is defined on globalThis
export function ensureGoRuntime(): void {
  if (typeof (globalThis as any).Go !== 'undefined') {
    return;
  }

  const enosys = () => {
    const err = new Error("not implemented");
    (err as any).code = "ENOSYS";
    return err;
  };

  if (!(globalThis as any).fs) {
    let outputBuf = "";
    (globalThis as any).fs = {
      constants: { O_WRONLY: -1, O_RDWR: -1, O_CREAT: -1, O_TRUNC: -1, O_APPEND: -1, O_EXCL: -1, O_DIRECTORY: -1 },
      writeSync(fd: number, buf: Uint8Array) {
        outputBuf += decoder.decode(buf);
        const nl = outputBuf.lastIndexOf("\n");
        if (nl != -1) {
          console.log(outputBuf.substring(0, nl));
          outputBuf = outputBuf.substring(nl + 1);
        }
        return buf.length;
      },
      write(fd: number, buf: Uint8Array, offset: number, length: number, position: any, callback: any) {
        if (offset !== 0 || length !== buf.length || position !== null) {
          callback(enosys());
          return;
        }
        const n = this.writeSync(fd, buf);
        callback(null, n);
      },
      chmod(path: any, mode: any, callback: any) { callback(enosys()); },
      chown(path: any, uid: any, gid: any, callback: any) { callback(enosys()); },
      close(fd: any, callback: any) { callback(enosys()); },
      fchmod(fd: any, mode: any, callback: any) { callback(enosys()); },
      fchown(fd: any, uid: any, gid: any, callback: any) { callback(enosys()); },
      fstat(fd: any, callback: any) { callback(enosys()); },
      fsync(fd: any, callback: any) { callback(null); },
      ftruncate(fd: any, length: any, callback: any) { callback(enosys()); },
      lchown(path: any, uid: any, gid: any, callback: any) { callback(enosys()); },
      link(path: any, link: any, callback: any) { callback(enosys()); },
      lstat(path: any, callback: any) { callback(enosys()); },
      mkdir(path: any, perm: any, callback: any) { callback(enosys()); },
      open(path: any, flags: any, mode: any, callback: any) { callback(enosys()); },
      read(fd: any, buffer: any, offset: any, length: any, position: any, callback: any) { callback(enosys()); },
      readdir(path: any, callback: any) { callback(enosys()); },
      readlink(path: any, callback: any) { callback(enosys()); },
      rename(from: any, to: any, callback: any) { callback(enosys()); },
      rmdir(path: any, callback: any) { callback(enosys()); },
      stat(path: any, callback: any) { callback(enosys()); },
      symlink(path: any, link: any, callback: any) { callback(enosys()); },
      truncate(path: any, length: any, callback: any) { callback(enosys()); },
      unlink(path: any, callback: any) { callback(enosys()); },
      utimes(path: any, atime: any, mtime: any, callback: any) { callback(enosys()); },
    };
  }

  if (!(globalThis as any).process) {
    (globalThis as any).process = {
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
    };
  }

  if (!(globalThis as any).path) {
    (globalThis as any).path = {
      resolve(...pathSegments: string[]) {
        return pathSegments.join("/");
      }
    };
  }

  const encoder = new TextEncoder();
  const decoder = new TextDecoder("utf-8");

  (globalThis as any).Go = class Go {
    argv: string[];
    env: Record<string, string>;
    exit: (code: number) => void;
    importObject: Record<string, any>;
    private _exitCode: number;
    private _values: any[];
    private _goRefCounts: number[];
    private _ids: Map<any, number>;
    private _idPool: number[];
    private _inst?: WebAssembly.Instance;
    private _uint8Array?: Uint8Array;

    constructor() {
      this.argv = ["js"];
      this.env = {};
      this.exit = (code: number) => {
        if (code !== 0) {
          console.warn("exit code:", code);
        }
      };
      this._exitCode = 0;
      this._values = [
        NaN,
        0,
        null,
        true,
        false,
        globalThis,
        this,
      ];
      this._goRefCounts = new Array(this._values.length).fill(Infinity);
      this._ids = new Map<any, number>([
        [0, 1],
        [null, 2],
        [true, 3],
        [false, 4],
        [globalThis, 5],
        [this, 6],
      ]);
      this._idPool = [];

      const mem = () => {
        return new DataView((this._inst!.exports.mem as any).buffer as ArrayBuffer);
      };

      const setInt64 = (addr: number, v: number) => {
        mem().setUint32(addr + 0, v, true);
        mem().setUint32(addr + 4, Math.floor(v / 4294967296), true);
      };

      const getInt64 = (addr: number) => {
        const low = mem().getUint32(addr + 0, true);
        const high = mem().getInt32(addr + 4, true);
        return low + high * 4294967296;
      };

      const loadValue = (addr: number) => {
        const f = mem().getFloat64(addr, true);
        if (f === 0) return undefined;
        if (!isNaN(f)) return f;

        const id = mem().getUint32(addr, true);
        return this._values[id];
      };

      const storeValue = (addr: number, v: any) => {
        const nanHead = 0x7FF80000;

        if (typeof v === "number" && v !== 0) {
          if (isNaN(v)) {
            mem().setUint32(addr + 4, nanHead, true);
            mem().setUint32(addr + 0, 0, true);
            return;
          }
          mem().setFloat64(addr, v, true);
          return;
        }

        if (v === undefined) {
          mem().setFloat64(addr, 0, true);
          return;
        }

        let id = this._ids.get(v);
        if (id === undefined) {
          ref: {
            id = this._idPool.pop();
            if (id === undefined) {
              id = this._values.length;
            }
            this._values[id] = v;
            this._goRefCounts[id] = 0;
            this._ids.set(v, id);
          }
        }
        this._goRefCounts[id]++;
        let typeFlag = 0;
        switch (typeof v) {
          case "object": if (v !== null) typeFlag = 1; break;
          case "string": typeFlag = 2; break;
          case "symbol": typeFlag = 3; break;
          case "function": typeFlag = 4; break;
        }
        mem().setUint32(addr + 4, nanHead | typeFlag, true);
        mem().setUint32(addr + 0, id, true);
      };

      const loadSlice = (addr: number) => {
        const array = getInt64(addr + 0);
        const len = getInt64(addr + 8);
        return new Uint8Array((this._inst!.exports.mem as WebAssembly.Memory).buffer, array, len);
      };

      const loadString = (addr: number) => {
        const saddr = getInt64(addr + 0);
        const len = getInt64(addr + 8);
        return decoder.decode(new Uint8Array((this._inst!.exports.mem as WebAssembly.Memory).buffer, saddr, len));
      };

      const timeOrigin = Date.now() - performance.now();

      this.importObject = {
        go: {
          "runtime.wasmExit": (sp: number) => {
            sp >>>= 0;
            const code = mem().getInt32(sp + 8, true);
            this._exitCode = code;
            delete (this as any)._inst;
            delete (this as any)._values;
            delete (this as any)._goRefCounts;
            delete (this as any)._ids;
            delete (this as any)._idPool;
            this.exit(code);
          },
          "runtime.wasmWrite": (sp: number) => {
            sp >>>= 0;
            const fd = getInt64(sp + 8);
            const p = getInt64(sp + 16);
            const n = mem().getInt32(sp + 24, true);
            (globalThis as any).fs.writeSync(fd, new Uint8Array((this._inst!.exports.mem as WebAssembly.Memory).buffer, p, n));
          },
          "runtime.resetMemoryDataView": () => {
            // Memory view reset
          },
          "runtime.nanotime1": (sp: number) => {
            sp >>>= 0;
            setInt64(sp + 8, (timeOrigin + performance.now()) * 1000000);
          },
          "runtime.walltime": (sp: number) => {
            sp >>>= 0;
            const msec = Date.now();
            setInt64(sp + 8, Math.floor(msec / 1000));
            mem().setInt32(sp + 16, (msec % 1000) * 1000000, true);
          },
          "runtime.scheduleTimeoutEvent": (sp: number) => {
            sp >>>= 0;
            const id = mem().getInt32(sp + 8, true);
            setTimeout(() => {
              (this._inst!.exports as any).resume();
            }, getInt64(sp + 16));
          },
          "runtime.clearTimeoutEvent": (sp: number) => {
            // Unused
          },
          "runtime.getRandomData": (sp: number) => {
            sp >>>= 0;
            crypto.getRandomValues(loadSlice(sp + 8));
          },
          "syscall/js.finalizeRef": (sp: number) => {
            sp >>>= 0;
            const id = mem().getUint32(sp + 8, true);
            this._goRefCounts[id]--;
            if (this._goRefCounts[id] === 0) {
              const v = this._values[id];
              this._values[id] = 0;
              this._ids.delete(v);
              this._idPool.push(id);
            }
          },
          "syscall/js.stringVal": (sp: number) => {
            sp >>>= 0;
            storeValue(sp + 24, loadString(sp + 8));
          },
          "syscall/js.valueGet": (sp: number) => {
            sp >>>= 0;
            const result = Reflect.get(loadValue(sp + 8), loadString(sp + 16));
            sp = (this as any)._inst.exports.getsp();
            storeValue(sp + 32, result);
          },
          "syscall/js.valueSet": (sp: number) => {
            sp >>>= 0;
            Reflect.set(loadValue(sp + 8), loadString(sp + 16), loadValue(sp + 32));
          },
          "syscall/js.valueDelete": (sp: number) => {
            sp >>>= 0;
            Reflect.deleteProperty(loadValue(sp + 8), loadString(sp + 16));
          },
          "syscall/js.valueIndex": (sp: number) => {
            sp >>>= 0;
            storeValue(sp + 24, Reflect.get(loadValue(sp + 8), getInt64(sp + 16)));
          },
          "syscall/js.valueSetIndex": (sp: number) => {
            sp >>>= 0;
            Reflect.set(loadValue(sp + 8), getInt64(sp + 16), loadValue(sp + 24));
          },
          "syscall/js.valueCall": (sp: number) => {
            sp >>>= 0;
            try {
              const v = loadValue(sp + 8);
              const m = Reflect.get(v, loadString(sp + 16));
              const args = loadSlice(sp + 32);
              const result = Reflect.apply(m, v, (args as any));
              sp = (this as any)._inst.exports.getsp();
              storeValue(sp + 56, result);
              mem().setUint8(sp + 64, 1);
            } catch (err) {
              sp = (this as any)._inst.exports.getsp();
              storeValue(sp + 56, err);
              mem().setUint8(sp + 64, 0);
            }
          },
          "syscall/js.valueInvoke": (sp: number) => {
            sp >>>= 0;
            try {
              const v = loadValue(sp + 8);
              const args = loadSlice(sp + 16);
              const result = Reflect.apply(v, undefined, (args as any));
              sp = (this as any)._inst.exports.getsp();
              storeValue(sp + 40, result);
              mem().setUint8(sp + 48, 1);
            } catch (err) {
              sp = (this as any)._inst.exports.getsp();
              storeValue(sp + 40, err);
              mem().setUint8(sp + 48, 0);
            }
          },
          "syscall/js.valueNew": (sp: number) => {
            sp >>>= 0;
            try {
              const v = loadValue(sp + 8);
              const args = loadSlice(sp + 16);
              const result = Reflect.construct(v, (args as any));
              sp = (this as any)._inst.exports.getsp();
              storeValue(sp + 40, result);
              mem().setUint8(sp + 48, 1);
            } catch (err) {
              sp = (this as any)._inst.exports.getsp();
              storeValue(sp + 40, err);
              mem().setUint8(sp + 48, 0);
            }
          },
          "syscall/js.valueLength": (sp: number) => {
            sp >>>= 0;
            setInt64(sp + 16, loadValue(sp + 8).length);
          },
          "syscall/js.valuePrepareString": (sp: number) => {
            sp >>>= 0;
            const str = encoder.encode(String(loadValue(sp + 8)));
            storeValue(sp + 16, str);
            setInt64(sp + 24, str.length);
          },
          "syscall/js.valueLoadString": (sp: number) => {
            sp >>>= 0;
            const str = loadValue(sp + 8);
            loadSlice(sp + 16).set(str);
          },
          "syscall/js.copyBytesToGo": (sp: number) => {
            sp >>>= 0;
            const dst = loadSlice(sp + 8);
            const src = loadValue(sp + 32);
            if (!(src instanceof Uint8Array || src instanceof Uint8ClampedArray)) {
              mem().setUint8(sp + 48, 0);
              return;
            }
            const toCopy = src.subarray(0, dst.length);
            dst.set(toCopy);
            setInt64(sp + 40, toCopy.length);
            mem().setUint8(sp + 48, 1);
          },
          "syscall/js.copyBytesToJS": (sp: number) => {
            sp >>>= 0;
            const dst = loadValue(sp + 8);
            const src = loadSlice(sp + 24);
            if (!(dst instanceof Uint8Array || dst instanceof Uint8ClampedArray)) {
              mem().setUint8(sp + 48, 0);
              return;
            }
            const toCopy = src.subarray(0, dst.length);
            dst.set(toCopy);
            setInt64(sp + 40, toCopy.length);
            mem().setUint8(sp + 48, 1);
          },
        }
      };
    }

    async run(instance: WebAssembly.Instance) {
      this._inst = instance;
      this._uint8Array = new Uint8Array((this._inst.exports.mem as WebAssembly.Memory).buffer);
      (this._inst.exports as any).run(this.argv.length, 0);
    }
  };
}
