package utils

func FlattenMap(keyMap map[string]any) []any {
	attributes := make([]any, 0)

	for key, value := range keyMap {
		attributes = append(attributes, key, value)
	}

	return attributes
}
