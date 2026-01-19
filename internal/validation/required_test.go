package validation

import (
	"testing"

	"mini-fhir/internal/fhir/dstu3"
)

func TestMissingRequiredNestedPath(t *testing.T) {
	rules := &RuleSet{
		ResourceType:  "Patient",
		RequiredPaths: []string{"gender", "name.family"},
	}
	patient := &dstu3.Patient{ResourceBase: dstu3.ResourceBase{ResourceType: "Patient"}}
	if missing := missingRequired(patient, rules); len(missing) == 0 {
		t.Fatalf("expected missing required fields")
	}

	patient.Gender = "female"
	patient.Name = []dstu3.HumanName{{Family: []string{"Smith"}}}
	missing := missingRequired(patient, rules)
	if len(missing) != 0 {
		t.Fatalf("expected no missing fields, got %v", missing)
	}
}

func TestMissingRequiredChoice(t *testing.T) {
	rules := &RuleSet{
		ResourceType: "Observation",
		Choices: []ChoiceRule{
			{BasePath: "effective[x]", Choices: []string{"effectiveDateTime", "effectivePeriod"}},
		},
	}

	obs := &dstu3.Observation{ResourceBase: dstu3.ResourceBase{ResourceType: "Observation"}}
	missing := missingRequired(obs, rules)
	if len(missing) == 0 {
		t.Fatalf("expected missing choice fields")
	}

	effective := "2024-01-01T00:00:00Z"
	obs.EffectiveDateTime = &effective
	missing = missingRequired(obs, rules)
	if len(missing) != 0 {
		t.Fatalf("expected choice to be satisfied, got %v", missing)
	}
}
