package validation

type OperationOutcome struct {
	ResourceType string           `json:"resourceType"`
	Issue        []OperationIssue `json:"issue"`
}

type OperationIssue struct {
	Severity    string `json:"severity"`
	Code        string `json:"code"`
	Diagnostics string `json:"diagnostics,omitempty"`
}

func NewOutcomeIssue(severity, code, diagnostics string) *OperationOutcome {
	return &OperationOutcome{
		ResourceType: "OperationOutcome",
		Issue: []OperationIssue{
			{
				Severity:    severity,
				Code:        code,
				Diagnostics: diagnostics,
			},
		},
	}
}
