package search

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"mini-fhir/internal/fhir/dstu3"
	"mini-fhir/internal/store"
)

type Searcher struct {
	registry *dstu3.Registry
	store    *store.Store
}

type SearchResult struct {
	Entries      []*store.ResourceEntry
	Included     []*store.ResourceEntry
	Count        int
	IncludeDepth int
}

func NewSearcher(registry *dstu3.Registry, store *store.Store) *Searcher {
	return &Searcher{registry: registry, store: store}
}

func (s *Searcher) Search(resourceType string, query url.Values) (*SearchResult, error) {
	entries, err := s.store.List(resourceType)
	if err != nil {
		return nil, err
	}

	entries = filterByProfile(entries, query.Get("_profile"))
	entries = sortEntries(entries, query.Get("_sort"))
	count := len(entries)
	if countParam := query.Get("_count"); countParam != "" {
		parsed, err := parseCount(countParam)
		if err != nil {
			return nil, err
		}
		if parsed < len(entries) {
			entries = entries[:parsed]
		}
	}

	includes, err := s.expandIncludes(entries, query)
	if err != nil {
		return nil, err
	}

	return &SearchResult{Entries: entries, Included: includes, Count: count, IncludeDepth: 2}, nil
}

func filterByProfile(entries []*store.ResourceEntry, profile string) []*store.ResourceEntry {
	if profile == "" {
		return entries
	}
	filtered := make([]*store.ResourceEntry, 0, len(entries))
	for _, entry := range entries {
		meta := entry.Resource.GetMeta()
		if meta == nil {
			continue
		}
		for _, p := range meta.Profile {
			if p == profile {
				filtered = append(filtered, entry)
				break
			}
		}
	}
	return filtered
}

func sortEntries(entries []*store.ResourceEntry, sortParam string) []*store.ResourceEntry {
	if sortParam == "" {
		return entries
	}
	fields := strings.Split(sortParam, ",")
	if len(fields) == 0 {
		return entries
	}
	field := strings.TrimSpace(fields[0])
	if field == "" {
		return entries
	}
	desc := false
	if strings.HasPrefix(field, "-") {
		desc = true
		field = strings.TrimPrefix(field, "-")
	}

	sort.SliceStable(entries, func(i, j int) bool {
		left := sortValue(entries[i], field)
		right := sortValue(entries[j], field)
		if desc {
			return left > right
		}
		return left < right
	})
	return entries
}

func sortValue(entry *store.ResourceEntry, field string) string {
	switch field {
	case "_lastUpdated":
		return entry.LastUpdated
	case "id":
		return entry.Resource.GetID()
	case "date":
		if obs, ok := entry.Resource.(*dstu3.Observation); ok {
			if obs.EffectiveDateTime != nil {
				return *obs.EffectiveDateTime
			}
			if obs.EffectivePeriod != nil {
				return obs.EffectivePeriod.Start
			}
			return obs.Issued
		}
		return ""
	default:
		return ""
	}
}

func parseCount(count string) (int, error) {
	var parsed int
	_, err := fmt.Sscanf(count, "%d", &parsed)
	if err != nil || parsed < 0 {
		return 0, fmt.Errorf("invalid _count value")
	}
	return parsed, nil
}

func (s *Searcher) expandIncludes(entries []*store.ResourceEntry, query url.Values) ([]*store.ResourceEntry, error) {
	include := query["_include"]
	includeIterate := query["_include:iterate"]
	if len(include) == 0 && len(includeIterate) == 0 {
		return nil, nil
	}
	seen := map[string]struct{}{}
	result := []*store.ResourceEntry{}
	if err := s.expand(entries, 1, 2, seen, &result); err != nil {
		return nil, err
	}
	if len(includeIterate) > 0 {
		if err := s.expand(result, 2, 2, seen, &result); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (s *Searcher) expand(entries []*store.ResourceEntry, depth int, maxDepth int, seen map[string]struct{}, result *[]*store.ResourceEntry) error {
	if depth > maxDepth {
		return nil
	}
	for _, entry := range entries {
		for _, ref := range entry.Resource.References() {
			parts := strings.Split(ref.Reference, "/")
			if len(parts) != 2 {
				continue
			}
			key := parts[0] + "/" + parts[1]
			if _, ok := seen[key]; ok {
				continue
			}
			included, err := s.store.Get(parts[0], parts[1])
			if err != nil {
				continue
			}
			seen[key] = struct{}{}
			*result = append(*result, included)
		}
	}
	return nil
}

func parseFHIRTime(value string) (time.Time, bool) {
	if value == "" {
		return time.Time{}, false
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, true
	}
	return time.Time{}, false
}
