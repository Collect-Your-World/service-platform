package util

func Int64Ptr(i int64) *int64 {
	return &i
}

func Int32Ptr(i int32) *int32 {
	return &i
}

func GetOrDefault[T any](ptr *T, defaultValue T) T {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
