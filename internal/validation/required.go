package validation

import (
	"encoding/json"
	"fmt"
	"strings"

	"mini-fhir/internal/fhir/dstu3"
)

func missingRequired(resource dstu3.Resource, rules *RuleSet) []string {
	if rules == nil {
		return nil
	}
	data, err := json.Marshal(resource)
	if err != nil {
		return []string{"unable to marshal resource"}
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return []string{"unable to inspect resource"}
	}

	missing := []string{}
	for _, field := range rules.RequiredPaths {
		if !hasPath(raw, strings.Split(field, ".")) {
			missing = append(missing, fmt.Sprintf("%s.%s", rules.ResourceType, field))
		}
	}
	for _, choice := range rules.Choices {
		found := false
		for _, option := range choice.Choices {
			if hasPath(raw, strings.Split(option, ".")) {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, fmt.Sprintf("%s.%s", rules.ResourceType, choice.BasePath))
		}
	}
	return missing
}

func hasPath(raw map[string]any, path []string) bool {
	if len(path) == 0 {
		return true
	}
	value, ok := raw[path[0]]
	if !ok {
		return false
	}
	return hasPathValue(value, path[1:])
}

func hasPathValue(value any, path []string) bool {
	if len(path) == 0 {
		return true
	}
	switch typed := value.(type) {
	case map[string]any:
		return hasPath(typed, path)
	case []any:
		for _, item := range typed {
			if hasPathValue(item, path) {
				return true
			}
		}
		return false
	default:
		return false
	}
}
