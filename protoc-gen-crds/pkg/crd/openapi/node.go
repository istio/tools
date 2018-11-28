package openapi

// Node is a basic AST node for marshaling Open API spec.
type Node struct {
	Type       string           `json:"type"`
	Required   []string         `json:"required,omitempty"`
	Items      *Node            `json:"items,omitempty"`
	Properties map[string]*Node `json:"properties,omitempty"`
}
