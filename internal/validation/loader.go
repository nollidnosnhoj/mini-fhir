package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"mini-fhir/internal/fhir/dstu3"
)

type ProfileStore struct {
	profiles map[string]*RuleSet
	cacheDir string
	cacheTTL time.Duration
	version  int
}

type cacheEnvelope struct {
	Version int      `json:"version"`
	Rules   *RuleSet `json:"rules"`
}

const CacheVersion = 1 // Bump to invalidate cached rule sets

func NewProfileStore(cacheDir string, cacheTTL time.Duration, version int) *ProfileStore {
	if version <= 0 {
		version = CacheVersion
	}
	return &ProfileStore{profiles: map[string]*RuleSet{}, cacheDir: cacheDir, cacheTTL: cacheTTL, version: version}
}

func (p *ProfileStore) Add(profileURL string, rules *RuleSet) {
	p.profiles[profileURL] = rules
}

func (p *ProfileStore) Get(profileURL string) (*RuleSet, bool) {
	rules, ok := p.profiles[profileURL]
	return rules, ok
}

func (p *ProfileStore) LoadDefaults(ctx context.Context, registry *dstu3.Registry) error {
	client := defaultHTTPClient()
	resourceTypes := registry.ResourceTypes()
	sort.Strings(resourceTypes)
	for _, resourceType := range resourceTypes {
		info, ok := registry.Info(resourceType)
		if !ok {
			continue
		}
		if info.ProfileSource == "" {
			continue
		}
		rules, err := p.loadProfileRules(ctx, client, info.ProfileSource)
		if err != nil {
			return fmt.Errorf("load profile %s: %w", info.ProfileSource, err)
		}
		p.Add(info.ProfileSource, rules)
	}
	return nil
}

func (p *ProfileStore) loadProfileRules(ctx context.Context, client *http.Client, profileURL string) (*RuleSet, error) {
	if p.cacheDir != "" {
		if cached, ok := p.readCache(profileURL); ok {
			return cached, nil
		}
	}
	profile, err := loadProfile(ctx, client, profileURL)
	if err != nil {
		return nil, err
	}
	if p.cacheDir != "" {
		if err := p.writeCache(profileURL, profile); err != nil {
			return nil, err
		}
	}
	return profile, nil
}

func (p *ProfileStore) readCache(profileURL string) (*RuleSet, bool) {
	path := p.cachePath(profileURL)
	info, err := os.Stat(path)
	if err != nil {
		return nil, false
	}
	if p.cacheTTL > 0 && time.Since(info.ModTime()) > p.cacheTTL {
		return nil, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var envelope cacheEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, false
	}
	if envelope.Version != p.version || envelope.Rules == nil {
		return nil, false
	}
	if envelope.Rules.ResourceType == "" {
		return nil, false
	}
	return envelope.Rules, true
}

func (p *ProfileStore) writeCache(profileURL string, rules *RuleSet) error {
	path := p.cachePath(profileURL)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(cacheEnvelope{Version: p.version, Rules: rules})
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (p *ProfileStore) cachePath(profileURL string) string {
	sanitized := strings.NewReplacer("://", "_", "/", "_", "\\", "_").Replace(profileURL)
	return filepath.Join(p.cacheDir, sanitized+".json")
}
