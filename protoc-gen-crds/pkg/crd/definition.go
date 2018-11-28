package crd

import (
	"encoding/json"

	"istio.io/tools/protoc-gen-crds/pkg/crd/openapi"
)

// Scope of a CRD.
type Scope string

const (
	// Namespaced scope
	Namespaced Scope = "Namespaced"

	// Cluster scope
	Cluster Scope = "Cluster"
)

// ResourceDefinition represents a CRD.
type ResourceDefinition struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   Metadata               `json:"metadata"`
	Spec       ResourceDefinitionSpec `json:"spec"`
}

// Metadata portion of the CRD
type Metadata struct {
	Name string `json:"name"`
}

// ResourceDefinitionSpec is a CRD spec
type ResourceDefinitionSpec struct {
	Group      string       `json:"group"`
	Version    string       `json:"version"`
	Scope      Scope        `json:"scope"`
	Names      Names        `json:"names"`
	Validation openapi.Type `json:"validation"`
}

// Names section of Spec.
type Names struct {
	Plural   string `json:"plural,omitempty"`
	Singular string `json:"singular,omitempty"`
	Kind     string `json:"kind"`
}

// String implements Stringer
func (d *ResourceDefinition) String() string {
	// TODO: Use yaml to serialize
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(b)
}
