package reflect_test

import (
	"reflect"
	"testing"
)

func TestMethodByNameNotFound(t *testing.T) {
	v := reflect.ValueOf(struct{ X int }{X: 1})
	if m, ok := v.Type().MethodByName("NoSuchMethod"); ok || m.Name != "" {
		t.Fatalf("Type.MethodByName returned (%+v, %v), want (Method{}, false)", m, ok)
	}
	if mv := v.MethodByName("NoSuchMethod"); mv.IsValid() {
		t.Fatalf("Value.MethodByName returned valid Value %v, want zero Value", mv)
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on (reflect.Value{}).MethodByName, got nil")
		}
		ve, ok := r.(*reflect.ValueError)
		if !ok || ve.Kind != reflect.Invalid {
			t.Fatalf("recover() = %T (%v), want *reflect.ValueError with Kind Invalid", r, r)
		}
	}()
	(reflect.Value{}).MethodByName("NoSuchMethod")
}
