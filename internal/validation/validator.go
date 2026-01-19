package validation

import (
	"fmt"
	"strings"

	"mini-fhir/internal/fhir/dstu3"
)

type Validator struct {
	registry *dstu3.Registry
	profiles *ProfileStore
}

func NewValidator(registry *dstu3.Registry, profiles *ProfileStore) *Validator {
	return &Validator{registry: registry, profiles: profiles}
}

func (v *Validator) Validate(resource dstu3.Resource, profile string) *OperationOutcome {
	if resource == nil {
		return NewOutcomeIssue("error", "invalid", "resource is nil")
	}
	resourceType := resource.GetResourceType()
	if _, ok := v.registry.Info(resourceType); !ok {
		return NewOutcomeIssue("error", "invalid", fmt.Sprintf("unsupported resource type: %s", resourceType))
	}
	meta := resource.GetMeta()
	if meta != nil && len(meta.Profile) > 0 {
		for _, p := range meta.Profile {
			if strings.TrimSpace(p) == "" {
				return NewOutcomeIssue("error", "invalid", "meta.profile contains empty value")
			}
		}
	}
	if profile != "" {
		if strings.TrimSpace(profile) == "" {
			return NewOutcomeIssue("error", "invalid", "profile must not be empty")
		}
		if err := v.applyProfile(resource, profile); err != nil {
			return NewOutcomeIssue("error", "invalid", err.Error())
		}
	}
	if err := v.applyBaseProfile(resource); err != nil {
		return NewOutcomeIssue("error", "invalid", err.Error())
	}
	return nil
}

func (v *Validator) applyBaseProfile(resource dstu3.Resource) error {
	info, ok := v.registry.Info(resource.GetResourceType())
	if !ok {
		return fmt.Errorf("unsupported resource type: %s", resource.GetResourceType())
	}
	if info.ProfileSource == "" {
		return nil
	}
	return v.applyProfile(resource, info.ProfileSource)
}

func (v *Validator) applyProfile(resource dstu3.Resource, profileURL string) error {
	if v.profiles == nil {
		return fmt.Errorf("profiles not loaded")
	}
	rules, ok := v.profiles.Get(profileURL)
	if !ok {
		return fmt.Errorf("profile not loaded: %s", profileURL)
	}
	missing := missingRequired(resource, rules)
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}
	return nil
}
