package openapi

import "encoding/json"

// Array type
type Array struct {
	ElementType Type
}

var _ Type = &Array{}

// ToNode implementation
func (a *Array) ToNode() *Node {
	return &Node{
		Type:  "array",
		Items: a.ElementType.ToNode(),
	}
}

// MarshalJSON implementation
func (a *Array) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.ToNode())
}
