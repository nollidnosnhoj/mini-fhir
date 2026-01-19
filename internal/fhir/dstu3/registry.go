package dstu3

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type ResourceFactory func() Resource

type ResourceInfo struct {
	Type          string
	Factory       ResourceFactory
	ProfileURL    string
	ProfileSource string
}

type Registry struct {
	resources map[string]ResourceInfo
}

func NewRegistry() *Registry {
	resources := map[string]ResourceInfo{}
	add := func(resourceType string, factory ResourceFactory, profileSource string) {
		resources[resourceType] = ResourceInfo{
			Type:          resourceType,
			Factory:       factory,
			ProfileURL:    fmt.Sprintf("http://hl7.org/fhir/StructureDefinition/%s", resourceType),
			ProfileSource: profileSource,
		}
	}

	add("Patient", func() Resource { return &Patient{ResourceBase: ResourceBase{ResourceType: "Patient"}} }, "https://hl7.org/fhir/STU3/patient.profile.json")
	add("Practitioner", func() Resource { return &Practitioner{ResourceBase: ResourceBase{ResourceType: "Practitioner"}} }, "https://hl7.org/fhir/STU3/practitioner.profile.json")
	add("PractitionerRole", func() Resource {
		return &PractitionerRole{ResourceBase: ResourceBase{ResourceType: "PractitionerRole"}}
	}, "https://hl7.org/fhir/STU3/practitionerrole.profile.json")
	add("Organization", func() Resource { return &Organization{ResourceBase: ResourceBase{ResourceType: "Organization"}} }, "https://hl7.org/fhir/STU3/organization.profile.json")
	add("Observation", func() Resource { return &Observation{ResourceBase: ResourceBase{ResourceType: "Observation"}} }, "https://hl7.org/fhir/STU3/observation.profile.json")
	add("Flag", func() Resource { return &Flag{ResourceBase: ResourceBase{ResourceType: "Flag"}} }, "https://hl7.org/fhir/STU3/flag.profile.json")
	add("Consent", func() Resource { return &Consent{ResourceBase: ResourceBase{ResourceType: "Consent"}} }, "https://hl7.org/fhir/STU3/consent.profile.json")
	add("AdvanceDirective", func() Resource {
		return &AdvanceDirective{ResourceBase: ResourceBase{ResourceType: "AdvanceDirective"}}
	}, "https://hl7.org/fhir/STU3/advancedirective.profile.json")
	add("Location", func() Resource { return &Location{ResourceBase: ResourceBase{ResourceType: "Location"}} }, "https://hl7.org/fhir/STU3/location.profile.json")
	add("Task", func() Resource { return &Task{ResourceBase: ResourceBase{ResourceType: "Task"}} }, "https://hl7.org/fhir/STU3/task.profile.json")

	return &Registry{resources: resources}
}

func (r *Registry) ResourceTypes() []string {
	out := make([]string, 0, len(r.resources))
	for key := range r.resources {
		out = append(out, key)
	}
	return out
}

func (r *Registry) Info(resourceType string) (ResourceInfo, bool) {
	info, ok := r.resources[resourceType]
	return info, ok
}

func (r *Registry) NewResource(resourceType string) (Resource, error) {
	info, ok := r.resources[resourceType]
	if !ok {
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
	return info.Factory(), nil
}

func (r *Registry) DecodeResource(data []byte) (Resource, error) {
	resourceType, err := DetectResourceType(data)
	if err != nil {
		return nil, err
	}
	info, ok := r.resources[resourceType]
	if !ok {
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
	resource := info.Factory()
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(resource); err != nil {
		return nil, err
	}
	return resource, nil
}

func DetectResourceType(data []byte) (string, error) {
	var raw struct {
		ResourceType string `json:"resourceType"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return "", err
	}
	if raw.ResourceType == "" {
		return "", fmt.Errorf("missing resourceType")
	}
	return raw.ResourceType, nil
}
