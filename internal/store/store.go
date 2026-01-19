package store

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"mini-fhir/internal/fhir/dstu3"
)

type ResourceEntry struct {
	Resource    dstu3.Resource
	VersionID   string
	LastUpdated string
	History     []dstu3.Resource
}

type Store struct {
	mu        sync.RWMutex
	resources map[string]map[string]*ResourceEntry
}

func NewStore() *Store {
	return &Store{
		resources: map[string]map[string]*ResourceEntry{},
	}
}

func (s *Store) Create(resource dstu3.Resource) (*ResourceEntry, error) {
	if resource == nil {
		return nil, fmt.Errorf("resource is nil")
	}
	if resource.GetID() == "" {
		return nil, fmt.Errorf("resource id is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	resourceType := resource.GetResourceType()
	if _, ok := s.resources[resourceType]; !ok {
		s.resources[resourceType] = map[string]*ResourceEntry{}
	}
	if _, exists := s.resources[resourceType][resource.GetID()]; exists {
		return nil, fmt.Errorf("resource already exists")
	}

	entry := newEntry(resource, "1")
	s.resources[resourceType][resource.GetID()] = entry
	return cloneEntry(entry), nil
}

func (s *Store) Update(resource dstu3.Resource) (*ResourceEntry, error) {
	if resource == nil {
		return nil, fmt.Errorf("resource is nil")
	}
	if resource.GetID() == "" {
		return nil, fmt.Errorf("resource id is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	resourceType := resource.GetResourceType()
	if _, ok := s.resources[resourceType]; !ok {
		s.resources[resourceType] = map[string]*ResourceEntry{}
	}

	entry, exists := s.resources[resourceType][resource.GetID()]
	if !exists {
		entry = newEntry(resource, "1")
		s.resources[resourceType][resource.GetID()] = entry
		return cloneEntry(entry), nil
	}

	entry.History = append(entry.History, entry.Resource)
	nextVersion := fmt.Sprintf("%d", len(entry.History)+1)
	entry.Resource = resource
	entry.VersionID = nextVersion
	entry.LastUpdated = time.Now().UTC().Format(time.RFC3339)
	applyMeta(entry)
	return cloneEntry(entry), nil
}

func (s *Store) Delete(resourceType, id string) error {
	if resourceType == "" || id == "" {
		return fmt.Errorf("resource type and id are required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.resources[resourceType]; !ok {
		return fmt.Errorf("resource not found")
	}
	if _, ok := s.resources[resourceType][id]; !ok {
		return fmt.Errorf("resource not found")
	}
	delete(s.resources[resourceType], id)
	return nil
}

func (s *Store) Get(resourceType, id string) (*ResourceEntry, error) {
	if resourceType == "" || id == "" {
		return nil, fmt.Errorf("resource type and id are required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.resources[resourceType][id]
	if !ok {
		return nil, fmt.Errorf("resource not found")
	}
	return cloneEntry(entry), nil
}

func (s *Store) List(resourceType string) ([]*ResourceEntry, error) {
	if resourceType == "" {
		return nil, fmt.Errorf("resource type is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, ok := s.resources[resourceType]
	if !ok {
		return nil, nil
	}
	ids := make([]string, 0, len(entries))
	for id := range entries {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := make([]*ResourceEntry, 0, len(ids))
	for _, id := range ids {
		result = append(result, cloneEntry(entries[id]))
	}
	return result, nil
}

func (s *Store) History(resourceType, id string) ([]dstu3.Resource, error) {
	entry, err := s.Get(resourceType, id)
	if err != nil {
		return nil, err
	}
	return entry.History, nil
}

func (s *Store) SystemHistory() []dstu3.Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := []dstu3.Resource{}
	for _, items := range s.resources {
		for _, entry := range items {
			result = append(result, entry.Resource)
		}
	}
	return result
}

func cloneEntry(entry *ResourceEntry) *ResourceEntry {
	clone := &ResourceEntry{
		VersionID:   entry.VersionID,
		LastUpdated: entry.LastUpdated,
		History:     make([]dstu3.Resource, 0, len(entry.History)),
	}
	if entry.Resource != nil {
		res, err := entry.Resource.Clone()
		if err == nil {
			clone.Resource = res
		} else {
			clone.Resource = entry.Resource
		}
	}
	for _, item := range entry.History {
		res, err := item.Clone()
		if err == nil {
			clone.History = append(clone.History, res)
		} else {
			clone.History = append(clone.History, item)
		}
	}
	return clone
}

func newEntry(resource dstu3.Resource, version string) *ResourceEntry {
	entry := &ResourceEntry{
		Resource:  resource,
		VersionID: version,
	}
	entry.LastUpdated = time.Now().UTC().Format(time.RFC3339)
	applyMeta(entry)
	return entry
}

func applyMeta(entry *ResourceEntry) {
	meta := entry.Resource.GetMeta()
	if meta == nil {
		meta = &dstu3.Meta{}
	}
	meta.VersionID = entry.VersionID
	meta.LastUpdated = entry.LastUpdated
	entry.Resource.SetMeta(meta)
}
