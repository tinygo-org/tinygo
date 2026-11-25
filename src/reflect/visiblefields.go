package reflect

import "internal/reflectlite"

func VisibleFields(t Type) []StructField {
	fields := reflectlite.VisibleFields(toRawType(t))
	result := make([]StructField, len(fields))
	for i, field := range fields {
		result[i] = toStructField(field)
	}
	return result
}
