package collection

func Map[T, U any](data []T, f func(T) U) []U {
	r := make([]U, 0, len(data))
	for _, e := range data {
		r = append(r, f(e))
	}
	return r
}

func Filter[T any](data []T, f func(T) bool) []T {
	r := make([]T, 0, len(data))
	for _, e := range data {
		if f(e) {
			r = append(r, e)
		}
	}
	return r
}

func FindOne[T any](data []T, f func(T) bool) *T {
	for _, e := range data {
		if f(e) {
			return &e
		}
	}
	return nil
}

func Contains[T any](data []T, f func(T) bool) bool {
	for _, e := range data {
		if f(e) {
			return true
		}
	}
	return false
}
