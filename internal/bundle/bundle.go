package bundle

import "mini-fhir/internal/fhir/dstu3"

type Bundle struct {
	ResourceType string  `json:"resourceType"`
	Type         string  `json:"type"`
	Total        int     `json:"total,omitempty"`
	Entry        []Entry `json:"entry,omitempty"`
}

type Entry struct {
	FullURL  string         `json:"fullUrl,omitempty"`
	Resource dstu3.Resource `json:"resource,omitempty"`
	Search   *EntrySearch   `json:"search,omitempty"`
	Response *EntryResponse `json:"response,omitempty"`
}

type EntrySearch struct {
	Mode string `json:"mode,omitempty"`
}

type EntryResponse struct {
	Status string `json:"status,omitempty"`
}

func NewSearchBundle(total int) *Bundle {
	return &Bundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Total:        total,
	}
}

func NewBatchResponseBundle() *Bundle {
	return &Bundle{
		ResourceType: "Bundle",
		Type:         "batch-response",
	}
}
