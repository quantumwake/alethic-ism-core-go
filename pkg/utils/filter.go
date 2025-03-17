package utils

// MapReduce is a generic function that takes a map of any key/value types
// and returns a slice of output type after applying transform function and optional filter
func MapReduce[K comparable, V any, R any](
	input map[K]V,
	transform func(K, V) R,
	filter func(K, V) bool,
) []R {
	result := make([]R, 0, len(input))

	for k, v := range input {
		if filter == nil || filter(k, v) {
			result = append(result, transform(k, v))
		}
	}

	return result
}

// MapValues is a simplified version that just transforms map values without filtering
func MapValues[K comparable, V any, R any](
	input map[K]V,
	transform func(V) R,
) []R {
	return MapReduce(
		input,
		func(_ K, v V) R { return transform(v) },
		nil,
	)
}

// SliceToMap converts a slice of values into a map using the provided key function
func SliceToMap[T any, K comparable](
	slice []T,
	keyFn func(T) K,
) map[K]T {
	result := make(map[K]T, len(slice))

	for _, item := range slice {
		key := keyFn(item)
		result[key] = item
	}

	return result
}

func SliceToGroupMap[T any, K comparable](
	slice []T,
	keyFn func(T) K,
) map[K][]T {
	result := make(map[K][]T)

	for _, item := range slice {
		key := keyFn(item)
		result[key] = append(result[key], item)
	}

	return result
}

// FilterMap returns a new map containing only the key-value pairs that pass the filter
func FilterMap[K comparable, V any](
	input map[K]V,
	filter func(K, V) bool,
) map[K]V {
	result := make(map[K]V)

	for k, v := range input {
		if filter(k, v) {
			result[k] = v
		}
	}

	return result
}
