package reflect

import "internal/reflectlite"

func DeepEqual(x, y interface{}) bool {
	return reflectlite.DeepEqual(x, y)
}
