// +build avr cortexm

package interrupt

// Register is used to declare an interrupt. You should not normally call this
// function: it is only for telling the compiler about the mapping between an
// interrupt number and the interrupt handler name.
func Register(id int, handlerName string) int
