package interrupt

// A checkpoint is a setjmp like buffer, that can be used as a flag for
// interrupts.
//
// It can be used as follows:
//
//	// global var
//	var c Checkpoint
//
//	// to set up the checkpoint and wait for it
//	if c.Save() {
//		setupInterrupt()
//		for {
//			waitForInterrupt()
//		}
//	}
//
//	// Inside the interrupt handler:
//	if c.Saved() {
//		c.Jump()
//	}
type Checkpoint struct {
	jumpSP uintptr
	jumpPC uintptr
}

// Save the execution state in the given checkpoint, overwriting a previous
// saved checkpoint.
//
// This function returns twice: once the normal way after saving (returning
// true) and once after jumping (returning false).
//
// This function is a compiler intrinsic, it is not implemented in Go.
func (c *Checkpoint) Save() bool

// Returns whether a jump point was saved (and not erased due to a jump).
func (c *Checkpoint) Saved() bool {
	return c.jumpPC != 0
}

// Jump to the point where the execution state was saved, and erase the saved
// jump point. This must *only* be called from inside an interrupt.
//
// This method does not return in the conventional way, it resumes execution at
// the last point a checkpoint was saved.
func (c *Checkpoint) Jump() {
	if !c.Saved() {
		panic("runtime/interrupt: no checkpoint was saved")
	}
	jumpPC := c.jumpPC
	jumpSP := c.jumpSP
	c.jumpPC = 0
	c.jumpSP = 0
	if jumpPC == 0 {
		panic("jumping to 0")
	}
	checkpointJump(jumpSP, jumpPC)
}

//export tinygo_checkpointJump
func checkpointJump(jumpSP, jumpPC uintptr)
