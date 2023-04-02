package utils

// Map applies the given function to each element of the given array and returns
func Map[T any, R any](data []T, fn func(*T) R) []R {
	res := make([]R, len(data))
	for i, v := range data {
		res[i] = fn(&v)
	}
	return res
}
