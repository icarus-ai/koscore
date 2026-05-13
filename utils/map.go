package utils

// AssociateBy 将切片转为 map, keySelector 用于提取 key
func AssociateBy[T any, K comparable](list []T, keySelector func(T) K) map[K]T {
	result := make(map[K]T, len(list))
	for _, item := range list {
		result[keySelector(item)] = item
	}
	return result
}
