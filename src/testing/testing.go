package testing

import (
	"fmt"
)

type T struct {

}

func (t *T) Fatal(args ...string) {
	fmt.Println(args)
}