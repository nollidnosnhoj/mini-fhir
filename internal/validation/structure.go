package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type RuleSet struct {
	ResourceType  string       `json:"resourceType"`
	RequiredPaths []string     `json:"requiredPaths"`
	Choices       []ChoiceRule `json:"choices"`
}

type ChoiceRule struct {
	BasePath string   `json:"basePath"`
	Choices  []string `json:"choices"`
}

type structureDefinition struct {
	ResourceType string `json:"resourceType"`
	URL          string `json:"url"`
	Type         string `json:"type"`
	Snapshot     struct {
		Element []elementDefinition `json:"element"`
	} `json:"snapshot"`
}

type elementDefinition struct {
	Path string `json:"path"`
	Min  int    `json:"min"`
}

func loadProfile(ctx context.Context, client *http.Client, profileURL string) (*RuleSet, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, profileURL, nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("profile fetch failed: %s", response.Status)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var profile structureDefinition
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, err
	}
	if profile.ResourceType != "StructureDefinition" {
		return nil, fmt.Errorf("unexpected resourceType: %s", profile.ResourceType)
	}
	resourceType := profile.Type
	if resourceType == "" {
		return nil, fmt.Errorf("profile missing type")
	}

	required := map[string]struct{}{}
	choices := map[string]map[string]struct{}{}
	prefix := resourceType + "."
	for _, element := range profile.Snapshot.Element {
		if element.Min <= 0 {
			continue
		}
		if !strings.HasPrefix(element.Path, prefix) {
			continue
		}
		remaining := strings.TrimPrefix(element.Path, prefix)
		if remaining == "" {
			continue
		}
		if strings.Contains(remaining, "[x]") {
			choices[remaining] = map[string]struct{}{}
			continue
		}
		required[remaining] = struct{}{}
	}
	for _, element := range profile.Snapshot.Element {
		if !strings.HasPrefix(element.Path, prefix) {
			continue
		}
		remaining := strings.TrimPrefix(element.Path, prefix)
		if remaining == "" {
			continue
		}
		for base := range choices {
			basePrefix := strings.TrimSuffix(base, "[x]")
			if strings.HasPrefix(remaining, basePrefix) && remaining != base {
				if sameDepth(basePrefix, remaining) {
					choices[base][remaining] = struct{}{}
				}
			}
		}
	}
	fields := make([]string, 0, len(required))
	for field := range required {
		fields = append(fields, field)
	}
	choiceRules := make([]ChoiceRule, 0, len(choices))
	for base, options := range choices {
		values := make([]string, 0, len(options))
		for option := range options {
			values = append(values, option)
		}
		choiceRules = append(choiceRules, ChoiceRule{BasePath: base, Choices: values})
	}
	return &RuleSet{ResourceType: resourceType, RequiredPaths: fields, Choices: choiceRules}, nil
}

func defaultHTTPClient() *http.Client {
	return &http.Client{Timeout: 15 * time.Second}
}

func sameDepth(base string, path string) bool {
	return strings.Count(base, ".") == strings.Count(path, ".")
}
