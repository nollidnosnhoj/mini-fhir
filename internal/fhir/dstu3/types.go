package dstu3

import "encoding/json"

type Resource interface {
	GetResourceType() string
	GetID() string
	SetID(string)
	GetMeta() *Meta
	SetMeta(*Meta)
	References() []Reference
	Clone() (Resource, error)
}

type ResourceBase struct {
	ResourceType string            `json:"resourceType"`
	ID           string            `json:"id,omitempty"`
	Meta         *Meta             `json:"meta,omitempty"`
	Text         *Narrative        `json:"text,omitempty"`
	Contained    []json.RawMessage `json:"contained,omitempty"`
}

func (r *ResourceBase) GetResourceType() string { return r.ResourceType }
func (r *ResourceBase) GetID() string           { return r.ID }
func (r *ResourceBase) SetID(id string)         { r.ID = id }
func (r *ResourceBase) GetMeta() *Meta          { return r.Meta }
func (r *ResourceBase) SetMeta(meta *Meta)      { r.Meta = meta }

func cloneResource[T any](resource T) (Resource, error) {
	data, err := json.Marshal(resource)
	if err != nil {
		return nil, err
	}
	var out T
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	res, ok := any(&out).(Resource)
	if !ok {
		return nil, nil
	}
	return res, nil
}

// Common datatypes

type Meta struct {
	VersionID   string   `json:"versionId,omitempty"`
	LastUpdated string   `json:"lastUpdated,omitempty"`
	Profile     []string `json:"profile,omitempty"`
}

type Narrative struct {
	Status string `json:"status,omitempty"`
	Div    string `json:"div,omitempty"`
}

type Identifier struct {
	Use    string           `json:"use,omitempty"`
	Type   *CodeableConcept `json:"type,omitempty"`
	System string           `json:"system,omitempty"`
	Value  string           `json:"value,omitempty"`
}

type CodeableConcept struct {
	Coding []Coding `json:"coding,omitempty"`
	Text   string   `json:"text,omitempty"`
}

type Coding struct {
	System  string `json:"system,omitempty"`
	Code    string `json:"code,omitempty"`
	Display string `json:"display,omitempty"`
}

type HumanName struct {
	Use    string   `json:"use,omitempty"`
	Text   string   `json:"text,omitempty"`
	Family []string `json:"family,omitempty"`
	Given  []string `json:"given,omitempty"`
}

type Address struct {
	Use        string   `json:"use,omitempty"`
	Type       string   `json:"type,omitempty"`
	Text       string   `json:"text,omitempty"`
	Line       []string `json:"line,omitempty"`
	City       string   `json:"city,omitempty"`
	State      string   `json:"state,omitempty"`
	PostalCode string   `json:"postalCode,omitempty"`
	Country    string   `json:"country,omitempty"`
}

type Period struct {
	Start string `json:"start,omitempty"`
	End   string `json:"end,omitempty"`
}

type Quantity struct {
	Value  *float64 `json:"value,omitempty"`
	Unit   string   `json:"unit,omitempty"`
	System string   `json:"system,omitempty"`
	Code   string   `json:"code,omitempty"`
}

type Annotation struct {
	AuthorString string `json:"authorString,omitempty"`
	Time         string `json:"time,omitempty"`
	Text         string `json:"text,omitempty"`
}

type ContactPoint struct {
	System string `json:"system,omitempty"`
	Value  string `json:"value,omitempty"`
	Use    string `json:"use,omitempty"`
}

type Extension struct {
	URL          string      `json:"url,omitempty"`
	ValueString  string      `json:"valueString,omitempty"`
	ValueBoolean *bool       `json:"valueBoolean,omitempty"`
	ValueCode    string      `json:"valueCode,omitempty"`
	ValueCoding  *Coding     `json:"valueCoding,omitempty"`
	ValuePeriod  *Period     `json:"valuePeriod,omitempty"`
	ValueRef     *Reference  `json:"valueReference,omitempty"`
	Extension    []Extension `json:"extension,omitempty"`
}

type Reference struct {
	Reference string `json:"reference,omitempty"`
	Display   string `json:"display,omitempty"`
}

type Signature struct {
	Type         []Coding   `json:"type,omitempty"`
	When         string     `json:"when,omitempty"`
	Who          *Reference `json:"who,omitempty"`
	TargetFormat string     `json:"targetFormat,omitempty"`
	SigFormat    string     `json:"sigFormat,omitempty"`
	Data         string     `json:"data,omitempty"`
}

// Resource types

type Patient struct {
	ResourceBase
	Identifier           []Identifier   `json:"identifier,omitempty"`
	Name                 []HumanName    `json:"name,omitempty"`
	Telecom              []ContactPoint `json:"telecom,omitempty"`
	Gender               string         `json:"gender,omitempty"`
	BirthDate            string         `json:"birthDate,omitempty"`
	Address              []Address      `json:"address,omitempty"`
	ManagingOrganization *Reference     `json:"managingOrganization,omitempty"`
	GeneralPractitioner  []Reference    `json:"generalPractitioner,omitempty"`
}

func (p *Patient) References() []Reference {
	refs := make([]Reference, 0)
	if p.ManagingOrganization != nil {
		refs = append(refs, *p.ManagingOrganization)
	}
	refs = append(refs, p.GeneralPractitioner...)
	return refs
}

func (p *Patient) Clone() (Resource, error) { return cloneResource(*p) }

// Practitioner

type Practitioner struct {
	ResourceBase
	Identifier []Identifier   `json:"identifier,omitempty"`
	Name       []HumanName    `json:"name,omitempty"`
	Telecom    []ContactPoint `json:"telecom,omitempty"`
	Address    []Address      `json:"address,omitempty"`
}

func (p *Practitioner) References() []Reference  { return nil }
func (p *Practitioner) Clone() (Resource, error) { return cloneResource(*p) }

// PractitionerRole

type PractitionerRole struct {
	ResourceBase
	Practitioner      *Reference  `json:"practitioner,omitempty"`
	Organization      *Reference  `json:"organization,omitempty"`
	Location          []Reference `json:"location,omitempty"`
	HealthcareService []Reference `json:"healthcareService,omitempty"`
}

func (p *PractitionerRole) References() []Reference {
	refs := make([]Reference, 0)
	if p.Practitioner != nil {
		refs = append(refs, *p.Practitioner)
	}
	if p.Organization != nil {
		refs = append(refs, *p.Organization)
	}
	refs = append(refs, p.Location...)
	refs = append(refs, p.HealthcareService...)
	return refs
}

func (p *PractitionerRole) Clone() (Resource, error) { return cloneResource(*p) }

// Organization

type Organization struct {
	ResourceBase
	Identifier []Identifier   `json:"identifier,omitempty"`
	Name       string         `json:"name,omitempty"`
	Telecom    []ContactPoint `json:"telecom,omitempty"`
	Address    []Address      `json:"address,omitempty"`
	PartOf     *Reference     `json:"partOf,omitempty"`
}

func (o *Organization) References() []Reference {
	if o.PartOf == nil {
		return nil
	}
	return []Reference{*o.PartOf}
}

func (o *Organization) Clone() (Resource, error) { return cloneResource(*o) }

// Observation

type Observation struct {
	ResourceBase
	Status            string          `json:"status,omitempty"`
	Code              CodeableConcept `json:"code,omitempty"`
	Subject           *Reference      `json:"subject,omitempty"`
	Performer         []Reference     `json:"performer,omitempty"`
	Encounter         *Reference      `json:"encounter,omitempty"`
	Specimen          *Reference      `json:"specimen,omitempty"`
	Device            *Reference      `json:"device,omitempty"`
	EffectiveDateTime *string         `json:"effectiveDateTime,omitempty"`
	EffectivePeriod   *Period         `json:"effectivePeriod,omitempty"`
	Issued            string          `json:"issued,omitempty"`
}

func (o *Observation) References() []Reference {
	refs := make([]Reference, 0)
	if o.Subject != nil {
		refs = append(refs, *o.Subject)
	}
	refs = append(refs, o.Performer...)
	if o.Encounter != nil {
		refs = append(refs, *o.Encounter)
	}
	if o.Specimen != nil {
		refs = append(refs, *o.Specimen)
	}
	if o.Device != nil {
		refs = append(refs, *o.Device)
	}
	return refs
}

func (o *Observation) Clone() (Resource, error) { return cloneResource(*o) }

// Flag

type Flag struct {
	ResourceBase
	Status    string          `json:"status,omitempty"`
	Category  CodeableConcept `json:"category,omitempty"`
	Code      CodeableConcept `json:"code,omitempty"`
	Subject   *Reference      `json:"subject,omitempty"`
	Encounter *Reference      `json:"encounter,omitempty"`
	Author    *Reference      `json:"author,omitempty"`
}

func (f *Flag) References() []Reference {
	refs := make([]Reference, 0)
	if f.Subject != nil {
		refs = append(refs, *f.Subject)
	}
	if f.Encounter != nil {
		refs = append(refs, *f.Encounter)
	}
	if f.Author != nil {
		refs = append(refs, *f.Author)
	}
	return refs
}

func (f *Flag) Clone() (Resource, error) { return cloneResource(*f) }

// Consent

type Consent struct {
	ResourceBase
	Status          string         `json:"status,omitempty"`
	Patient         *Reference     `json:"patient,omitempty"`
	Actor           []ConsentActor `json:"actor,omitempty"`
	Organization    []Reference    `json:"organization,omitempty"`
	SourceReference *Reference     `json:"sourceReference,omitempty"`
}

type ConsentActor struct {
	Role      CodeableConcept `json:"role,omitempty"`
	Reference Reference       `json:"reference,omitempty"`
}

func (c *Consent) References() []Reference {
	refs := make([]Reference, 0)
	if c.Patient != nil {
		refs = append(refs, *c.Patient)
	}
	for _, actor := range c.Actor {
		if actor.Reference.Reference != "" {
			refs = append(refs, actor.Reference)
		}
	}
	refs = append(refs, c.Organization...)
	if c.SourceReference != nil {
		refs = append(refs, *c.SourceReference)
	}
	return refs
}

func (c *Consent) Clone() (Resource, error) { return cloneResource(*c) }

// AdvanceDirective

type AdvanceDirective struct {
	ResourceBase
	Patient *Reference  `json:"patient,omitempty"`
	Author  []Reference `json:"author,omitempty"`
	Source  *Reference  `json:"sourceReference,omitempty"`
}

func (a *AdvanceDirective) References() []Reference {
	refs := make([]Reference, 0)
	if a.Patient != nil {
		refs = append(refs, *a.Patient)
	}
	refs = append(refs, a.Author...)
	if a.Source != nil {
		refs = append(refs, *a.Source)
	}
	return refs
}

func (a *AdvanceDirective) Clone() (Resource, error) { return cloneResource(*a) }

// Task

type Task struct {
	ResourceBase
	Status          string           `json:"status,omitempty"`
	Intent          string           `json:"intent,omitempty"`
	Priority        string           `json:"priority,omitempty"`
	Description     string           `json:"description,omitempty"`
	Focus           *Reference       `json:"focus,omitempty"`
	For             *Reference       `json:"for,omitempty"`
	Requester       *Reference       `json:"requester,omitempty"`
	Owner           *Reference       `json:"owner,omitempty"`
	ExecutionPeriod *Period          `json:"executionPeriod,omitempty"`
	ReasonCode      *CodeableConcept `json:"reasonCode,omitempty"`
	BasedOn         []Reference      `json:"basedOn,omitempty"`
}

func (t *Task) References() []Reference {
	refs := make([]Reference, 0)
	if t.Focus != nil {
		refs = append(refs, *t.Focus)
	}
	if t.For != nil {
		refs = append(refs, *t.For)
	}
	if t.Requester != nil {
		refs = append(refs, *t.Requester)
	}
	if t.Owner != nil {
		refs = append(refs, *t.Owner)
	}
	refs = append(refs, t.BasedOn...)
	return refs
}

func (t *Task) Clone() (Resource, error) { return cloneResource(*t) }

// Location

type Location struct {
	ResourceBase
	Status               string           `json:"status,omitempty"`
	Name                 string           `json:"name,omitempty"`
	Description          string           `json:"description,omitempty"`
	Mode                 string           `json:"mode,omitempty"`
	Type                 CodeableConcept  `json:"type,omitempty"`
	Telecom              []ContactPoint   `json:"telecom,omitempty"`
	Address              *Address         `json:"address,omitempty"`
	PhysicalType         *CodeableConcept `json:"physicalType,omitempty"`
	ManagingOrganization *Reference       `json:"managingOrganization,omitempty"`
	PartOf               *Reference       `json:"partOf,omitempty"`
}

func (l *Location) References() []Reference {
	refs := make([]Reference, 0)
	if l.ManagingOrganization != nil {
		refs = append(refs, *l.ManagingOrganization)
	}
	if l.PartOf != nil {
		refs = append(refs, *l.PartOf)
	}
	return refs
}

func (l *Location) Clone() (Resource, error) { return cloneResource(*l) }
