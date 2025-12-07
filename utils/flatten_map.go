package utils

import "sort"

func FlattenMap(keyMap map[string]any) []any {
	attributes := make([]any, 0)
	keys := make([]string, 0, len(keyMap))
	for k := range keyMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := keyMap[key]
		attributes = append(attributes, key, value)
	}

	return attributes
}
