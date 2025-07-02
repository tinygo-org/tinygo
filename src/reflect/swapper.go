package reflect

import "internal/reflectlite"

func Swapper(slice interface{}) func(i, j int) {
	return reflectlite.Swapper(slice)
}
