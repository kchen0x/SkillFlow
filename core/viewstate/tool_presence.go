package viewstate

import "sort"

type ToolPresenceInput struct {
	Name        string
	Fingerprint string
}

type ToolPresenceSnapshot struct {
	ToolFingerprints map[string]string   `json:"toolFingerprints"`
	KeysByTool       map[string][]string `json:"keysByTool"`
	ToolsByKey       map[string][]string `json:"toolsByKey"`
}

func RebuildToolPresence(previous ToolPresenceSnapshot, inputs []ToolPresenceInput, scan func(toolName string) ([]string, error)) (ToolPresenceSnapshot, error) {
	next := ToolPresenceSnapshot{
		ToolFingerprints: map[string]string{},
		KeysByTool:       map[string][]string{},
		ToolsByKey:       map[string][]string{},
	}

	for _, input := range inputs {
		next.ToolFingerprints[input.Name] = input.Fingerprint

		keys := previous.KeysByTool[input.Name]
		if previous.ToolFingerprints[input.Name] != input.Fingerprint {
			scannedKeys, err := scan(input.Name)
			if err != nil {
				return ToolPresenceSnapshot{}, err
			}
			keys = scannedKeys
		}

		normalizedKeys := uniqueSorted(keys)
		next.KeysByTool[input.Name] = normalizedKeys
		for _, key := range normalizedKeys {
			next.ToolsByKey[key] = append(next.ToolsByKey[key], input.Name)
		}
	}

	for key := range next.ToolsByKey {
		next.ToolsByKey[key] = uniqueSorted(next.ToolsByKey[key])
	}

	return next, nil
}

func uniqueSorted(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
