package search

import (
	"net/url"
	"testing"

	"mini-fhir/internal/fhir/dstu3"
	"mini-fhir/internal/store"
)

func TestObservationSortDateDescending(t *testing.T) {
	registry := dstu3.NewRegistry()
	store := store.NewStore()
	searcher := NewSearcher(registry, store)

	first := &dstu3.Observation{
		ResourceBase: dstu3.ResourceBase{ResourceType: "Observation", ID: "obs-1"},
		Issued:       "2023-01-01T00:00:00Z",
	}
	second := &dstu3.Observation{
		ResourceBase: dstu3.ResourceBase{ResourceType: "Observation", ID: "obs-2"},
		Issued:       "2024-01-01T00:00:00Z",
	}
	if _, err := store.Update(first); err != nil {
		t.Fatalf("store update failed: %v", err)
	}
	if _, err := store.Update(second); err != nil {
		t.Fatalf("store update failed: %v", err)
	}

	query := url.Values{"_sort": []string{"-date"}}
	result, err := searcher.Search("Observation", query)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 results")
	}
	if result.Entries[0].Resource.GetID() != "obs-2" {
		t.Fatalf("expected obs-2 first, got %s", result.Entries[0].Resource.GetID())
	}
}

func TestIncludeExpand(t *testing.T) {
	registry := dstu3.NewRegistry()
	store := store.NewStore()
	searcher := NewSearcher(registry, store)

	org := &dstu3.Organization{ResourceBase: dstu3.ResourceBase{ResourceType: "Organization", ID: "org-1"}}
	patient := &dstu3.Patient{ResourceBase: dstu3.ResourceBase{ResourceType: "Patient", ID: "pat-1"}, ManagingOrganization: &dstu3.Reference{Reference: "Organization/org-1"}}
	if _, err := store.Update(org); err != nil {
		t.Fatalf("store update failed: %v", err)
	}
	if _, err := store.Update(patient); err != nil {
		t.Fatalf("store update failed: %v", err)
	}

	query := url.Values{"_include": []string{"*"}}
	result, err := searcher.Search("Patient", query)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(result.Included) != 1 {
		t.Fatalf("expected 1 included resource, got %d", len(result.Included))
	}
}
