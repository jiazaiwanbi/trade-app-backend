package textutil

func NormalizeStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}

	normalized := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		normalized = append(normalized, value)
	}
	return normalized
}
