package reflect

type SelectDir int

const (
	_             SelectDir = iota
	SelectSend              // case Chan <- Send
	SelectRecv              // case <-Chan:
	SelectDefault           // default
)

type SelectCase struct {
	Dir  SelectDir // direction of case
	Chan Value     // channel to use (for send or receive)
	Send Value     // value to send (for send)
}

func Select(cases []SelectCase) (chosen int, recv Value, recvOK bool) {
	panic("reflect.Select: unimplemented")
}

func (v Value) Send(x Value) {
	panic("reflect.Value.Send(): unimplemented")

}

func (v Value) Close() {
	panic("reflect.Value.Close(): unimplemented")

}
