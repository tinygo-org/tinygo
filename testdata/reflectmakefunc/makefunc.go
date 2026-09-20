package reflectmakefunc

import "reflect"

func Make(typ reflect.Type) any {
	return reflect.MakeFunc(typ, func(args []reflect.Value) []reflect.Value {
		return []reflect.Value{reflect.ValueOf(int(args[0].Int()) + 1)}
	}).Interface()
}
