package compiler

// This file implements volatile loads/stores in runtime/volatile.LoadT and
// runtime/volatile.StoreT as compiler builtins.

// createVolatileLoad is the implementation of the intrinsic function
// runtime/volatile.LoadT().
func (b *builder) createVolatileLoad() {
	b.createFunctionStart(true)
	addr := b.getValue(b.fn.Params[0])
	b.createNilCheck(b.fn.Params[0], addr, "deref")
	val := b.CreateLoad(addr, "")
	val.SetVolatile(true)
	b.CreateRet(val)
}

// createVolatileStore is the implementation of the intrinsic function
// runtime/volatile.StoreT().
func (b *builder) createVolatileStore() {
	b.createFunctionStart(true)
	addr := b.getValue(b.fn.Params[0])
	val := b.getValue(b.fn.Params[1])
	b.createNilCheck(b.fn.Params[0], addr, "deref")
	store := b.CreateStore(val, addr)
	store.SetVolatile(true)
	b.CreateRetVoid()
}
