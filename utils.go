package percook

// return first non-empty string in list or defaultValue when all elements are empty
func stringCoalesceWithDefault(defaultValue string, choices ...string) string {
	for _, choice := range choices {
		if len(choice) > 0 {
			return choice
		}
	}
	return defaultValue
}

func stringMin(choices ...string) string {
	if len(choices) == 0 {
		return ""
	}

	best := choices[0]
	for _, choice := range choices {
		if len(choice) < len(best) {
			best = choice
		}
	}
	return best
}
