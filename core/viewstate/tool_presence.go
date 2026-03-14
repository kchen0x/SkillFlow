package viewstate

import "sort"

type AgentPresenceInput struct {
	Name        string
	Fingerprint string
}

type AgentPresenceSnapshot struct {
	AgentFingerprints map[string]string   `json:"agentFingerprints"`
	KeysByAgent       map[string][]string `json:"keysByAgent"`
	AgentsByKey       map[string][]string `json:"agentsByKey"`
}

func RebuildAgentPresence(previous AgentPresenceSnapshot, inputs []AgentPresenceInput, scan func(agentName string) ([]string, error)) (AgentPresenceSnapshot, error) {
	next := AgentPresenceSnapshot{
		AgentFingerprints: map[string]string{},
		KeysByAgent:       map[string][]string{},
		AgentsByKey:       map[string][]string{},
	}

	for _, input := range inputs {
		next.AgentFingerprints[input.Name] = input.Fingerprint

		keys := previous.KeysByAgent[input.Name]
		if previous.AgentFingerprints[input.Name] != input.Fingerprint {
			scannedKeys, err := scan(input.Name)
			if err != nil {
				return AgentPresenceSnapshot{}, err
			}
			keys = scannedKeys
		}

		normalizedKeys := uniqueSorted(keys)
		next.KeysByAgent[input.Name] = normalizedKeys
		for _, key := range normalizedKeys {
			next.AgentsByKey[key] = append(next.AgentsByKey[key], input.Name)
		}
	}

	for key := range next.AgentsByKey {
		next.AgentsByKey[key] = uniqueSorted(next.AgentsByKey[key])
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
