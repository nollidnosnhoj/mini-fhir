package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/labstack/echo/v4"

	"mini-fhir/internal/bundle"
	"mini-fhir/internal/fhir/dstu3"
	"mini-fhir/internal/store"
	"mini-fhir/internal/validation"
)

func (s *Server) handleMetadata(c echo.Context) error {
	capability := map[string]any{
		"resourceType": "CapabilityStatement",
		"status":       "active",
		"fhirVersion":  "3.0.2",
		"format":       []string{"json"},
		"rest": []map[string]any{
			{
				"mode":     "server",
				"resource": s.capabilityResources(),
				"interaction": []map[string]string{
					{"code": "batch"},
					{"code": "transaction"},
				},
			},
		},
	}
	return c.JSON(http.StatusOK, capability)
}

func (s *Server) capabilityResources() []map[string]any {
	resources := make([]map[string]any, 0)
	resourceTypes := s.Registry.ResourceTypes()
	sort.Strings(resourceTypes)
	for _, resourceType := range resourceTypes {
		resources = append(resources, map[string]any{
			"type": resourceType,
			"interaction": []map[string]string{
				{"code": "read"},
				{"code": "vread"},
				{"code": "update"},
				{"code": "delete"},
				{"code": "create"},
				{"code": "history-instance"},
				{"code": "search-type"},
			},
			"searchParam": []map[string]string{
				{"name": "_include"},
				{"name": "_include:iterate"},
				{"name": "_profile"},
				{"name": "_count"},
				{"name": "_sort"},
			},
		})
	}
	return resources
}

func (s *Server) handleCreate(c echo.Context) error {
	resource, err := s.decodeBody(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, validation.NewOutcomeIssue("error", "invalid", err.Error()))
	}
	if resource.GetResourceType() != c.Param("type") {
		return c.JSON(http.StatusBadRequest, validation.NewOutcomeIssue("error", "invalid", "resourceType does not match URL"))
	}
	if resource.GetID() == "" {
		return c.JSON(http.StatusBadRequest, validation.NewOutcomeIssue("error", "required", "id is required"))
	}
	if outcome := s.Validator.Validate(resource, ""); outcome != nil {
		return c.JSON(http.StatusUnprocessableEntity, outcome)
	}
	entry, err := s.Store.Create(resource)
	if err != nil {
		return c.JSON(http.StatusConflict, validation.NewOutcomeIssue("error", "conflict", err.Error()))
	}
	return c.JSON(http.StatusCreated, entry.Resource)
}

func (s *Server) handleRead(c echo.Context) error {
	entry, err := s.Store.Get(c.Param("type"), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, validation.NewOutcomeIssue("error", "not-found", err.Error()))
	}
	return c.JSON(http.StatusOK, entry.Resource)
}

func (s *Server) handleUpdate(c echo.Context) error {
	resource, err := s.decodeBody(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, validation.NewOutcomeIssue("error", "invalid", err.Error()))
	}
	if resource.GetResourceType() != c.Param("type") {
		return c.JSON(http.StatusBadRequest, validation.NewOutcomeIssue("error", "invalid", "resourceType does not match URL"))
	}
	resource.SetID(c.Param("id"))
	if outcome := s.Validator.Validate(resource, ""); outcome != nil {
		return c.JSON(http.StatusUnprocessableEntity, outcome)
	}
	entry, err := s.Store.Update(resource)
	if err != nil {
		return c.JSON(http.StatusBadRequest, validation.NewOutcomeIssue("error", "invalid", err.Error()))
	}
	return c.JSON(http.StatusOK, entry.Resource)
}

func (s *Server) handleDelete(c echo.Context) error {
	if err := s.Store.Delete(c.Param("type"), c.Param("id")); err != nil {
		return c.JSON(http.StatusNotFound, validation.NewOutcomeIssue("error", "not-found", err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (s *Server) handleHistory(c echo.Context) error {
	resources, err := s.Store.History(c.Param("type"), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, validation.NewOutcomeIssue("error", "not-found", err.Error()))
	}
	searchBundle := bundle.NewSearchBundle(len(resources))
	for _, res := range resources {
		searchBundle.Entry = append(searchBundle.Entry, bundle.Entry{Resource: res})
	}
	return c.JSON(http.StatusOK, searchBundle)
}

func (s *Server) handleSystemHistory(c echo.Context) error {
	resources := s.Store.SystemHistory()
	searchBundle := bundle.NewSearchBundle(len(resources))
	for _, res := range resources {
		searchBundle.Entry = append(searchBundle.Entry, bundle.Entry{Resource: res})
	}
	return c.JSON(http.StatusOK, searchBundle)
}

func (s *Server) handleSearch(c echo.Context) error {
	resourceType := c.Param("type")
	if _, ok := s.Registry.Info(resourceType); !ok {
		return c.JSON(http.StatusNotFound, validation.NewOutcomeIssue("error", "not-found", "resource type not supported"))
	}
	result, err := s.Searcher.Search(resourceType, c.QueryParams())
	if err != nil {
		return c.JSON(http.StatusBadRequest, validation.NewOutcomeIssue("error", "invalid", err.Error()))
	}
	bundleResp := bundle.NewSearchBundle(result.Count)
	for _, entry := range result.Entries {
		bundleResp.Entry = append(bundleResp.Entry, bundle.Entry{Resource: entry.Resource})
	}
	for _, entry := range result.Included {
		bundleResp.Entry = append(bundleResp.Entry, bundle.Entry{Resource: entry.Resource, Search: &bundle.EntrySearch{Mode: "include"}})
	}
	return c.JSON(http.StatusOK, bundleResp)
}

func (s *Server) handleValidate(c echo.Context) error {
	resource, err := s.decodeBody(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, validation.NewOutcomeIssue("error", "invalid", err.Error()))
	}
	profile := c.QueryParam("profile")
	outcome := s.Validator.Validate(resource, profile)
	if outcome != nil {
		return c.JSON(http.StatusUnprocessableEntity, outcome)
	}
	return c.JSON(http.StatusOK, validation.NewOutcomeIssue("information", "informational", "validation succeeded"))
}

func (s *Server) handleBatchTransaction(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, validation.NewOutcomeIssue("error", "invalid", err.Error()))
	}
	var bundleReq struct {
		ResourceType string           `json:"resourceType"`
		Type         string           `json:"type"`
		Entry        []map[string]any `json:"entry"`
	}
	if err := json.Unmarshal(body, &bundleReq); err != nil {
		return c.JSON(http.StatusBadRequest, validation.NewOutcomeIssue("error", "invalid", err.Error()))
	}
	if bundleReq.ResourceType != "Bundle" {
		return c.JSON(http.StatusBadRequest, validation.NewOutcomeIssue("error", "invalid", "expected Bundle"))
	}

	responseBundle := bundle.NewBatchResponseBundle()
	for _, entry := range bundleReq.Entry {
		resp := bundle.Entry{Response: &bundle.EntryResponse{Status: "400"}}
		if resourceRaw, ok := entry["resource"]; ok {
			resourceBytes, _ := json.Marshal(resourceRaw)
			resource, err := s.Registry.DecodeResource(resourceBytes)
			if err == nil {
				if outcome := s.Validator.Validate(resource, ""); outcome == nil {
					if _, err := s.Store.Update(resource); err == nil {
						resp.Response.Status = "200"
					} else {
						resp.Response.Status = "400"
					}
				}
			}
		}
		responseBundle.Entry = append(responseBundle.Entry, resp)
	}
	return c.JSON(http.StatusOK, responseBundle)
}

func LoadSeed(pattern string, strict bool, registry *dstu3.Registry, validator *validation.Validator, store *store.Store) error {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return fmt.Errorf("no seed files match pattern")
	}
	for _, file := range matches {
		data, err := readFile(file)
		if err != nil {
			return err
		}
		if strings.Contains(string(data), "\"resourceType\":\"Bundle\"") {
			var seedBundle struct {
				Entry []struct {
					Resource json.RawMessage `json:"resource"`
				} `json:"entry"`
			}
			if err := json.Unmarshal(data, &seedBundle); err != nil {
				return err
			}
			for _, entry := range seedBundle.Entry {
				if err := loadSeedResource(entry.Resource, strict, registry, validator, store); err != nil {
					return err
				}
			}
			continue
		}
		if err := loadSeedResource(data, strict, registry, validator, store); err != nil {
			return err
		}
	}
	return nil
}

func loadSeedResource(data []byte, strict bool, registry *dstu3.Registry, validator *validation.Validator, store *store.Store) error {
	resource, err := registry.DecodeResource(data)
	if err != nil {
		if strict {
			return err
		}
		return nil
	}
	if outcome := validator.Validate(resource, ""); outcome != nil {
		if strict {
			return fmt.Errorf("validation failed: %s", outcome.Issue[0].Diagnostics)
		}
		return nil
	}
	if resource.GetID() == "" {
		if strict {
			return fmt.Errorf("resource id is required")
		}
		return nil
	}
	if _, err := store.Update(resource); err != nil {
		if strict {
			return err
		}
	}
	return nil
}

func (s *Server) decodeBody(c echo.Context) (dstu3.Resource, error) {
	data, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return nil, err
	}
	return s.Registry.DecodeResource(data)
}

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
