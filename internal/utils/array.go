package utils

// Map applies the given function to each element of the given array and returns
func Map[T any, R any](data []T, fn func(*T) R) []R {
	res := make([]R, len(data))
	for i, v := range data {
		res[i] = fn(&v)
	}
	return res
}

// CopyMap returns a copy of the given map.
func CopyMap[K comparable, V any](m map[K]V) map[K]V {
	result := make(map[K]V)
	for k, v := range m {
		result[k] = v
	}
	return result
}
