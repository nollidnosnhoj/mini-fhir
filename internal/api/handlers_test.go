package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"mini-fhir/internal/fhir/dstu3"
	"mini-fhir/internal/search"
	"mini-fhir/internal/store"
	"mini-fhir/internal/validation"
)

type outcomeResponse struct {
	ResourceType string `json:"resourceType"`
	Issue        []struct {
		Severity string `json:"severity"`
		Code     string `json:"code"`
	} `json:"issue"`
}

func setupTestServer() (*echo.Echo, *dstu3.Registry) {
	registry := dstu3.NewRegistry()
	profileStore := validation.NewProfileStore("", 0, validation.CacheVersion)
	for _, resourceType := range registry.ResourceTypes() {
		info, ok := registry.Info(resourceType)
		if !ok || info.ProfileSource == "" {
			continue
		}
		profileStore.Add(info.ProfileSource, &validation.RuleSet{ResourceType: resourceType})
	}
	validator := validation.NewValidator(registry, profileStore)
	store := store.NewStore()
	searcher := search.NewSearcher(registry, store)
	e := echo.New()
	RegisterRoutes(e, registry, validator, store, searcher)
	return e, registry
}

func TestValidateOperation(t *testing.T) {
	e, _ := setupTestServer()
	payload := []byte(`{"resourceType":"Patient","id":"pat-1"}`)
	request := httptest.NewRequest(http.MethodPost, "/Patient/$validate", bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	e.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var outcome outcomeResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("decode outcome failed: %v", err)
	}
	if outcome.ResourceType != "OperationOutcome" {
		t.Fatalf("expected OperationOutcome, got %s", outcome.ResourceType)
	}
}

func TestBatchTransactionCreatesResources(t *testing.T) {
	e, _ := setupTestServer()
	payload := []byte(`{"resourceType":"Bundle","type":"batch","entry":[{"resource":{"resourceType":"Patient","id":"pat-1"}}]}`)
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	e.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var bundle struct {
		Entry []struct {
			Response struct {
				Status string `json:"status"`
			} `json:"response"`
		} `json:"entry"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if len(bundle.Entry) != 1 || bundle.Entry[0].Response.Status != "200" {
		t.Fatalf("expected status 200, got %+v", bundle.Entry)
	}
}

func TestCreateRejectsUnknownFields(t *testing.T) {
	e, _ := setupTestServer()
	payload := []byte(`{"resourceType":"Patient","id":"pat-1","unknownField":"nope"}`)
	request := httptest.NewRequest(http.MethodPost, "/Patient", bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	e.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}
