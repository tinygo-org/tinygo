// +build avr riscv,baremetal cortexm

package interrupt

// New is a compiler intrinsic that creates a new Interrupt object. You may call
// it only once, and must pass constant parameters to it. That means that the
// interrupt ID must be a Go constant and that the handler must be a simple
// function: closures are not supported.
func New(id int, handler func(Interrupt)) Interrupt

// Register is used to declare an interrupt. You should not normally call this
// function: it is only for telling the compiler about the mapping between an
// interrupt number and the interrupt handler name.
func Register(id int, handlerName string) int

type handle struct {
	handler func(Interrupt)
	Interrupt
}
