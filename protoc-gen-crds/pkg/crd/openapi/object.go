package openapi

import "encoding/json"

// Object type
type Object struct {
	Name   string
	Fields []*Field
}

var _ Type = &Object{}

// Field of an object
type Field struct {
	Name string
	Type Type
}

// ToNode implementation
func (o *Object) ToNode() *Node {
	properties := make(map[string]*Node)
	for _, f := range o.Fields {
		properties[f.Name] = f.Type.ToNode()
	}

	return &Node{
		Type:       "object",
		Properties: properties,
	}
}

// MarshalJSON implementation
func (o *Object) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.ToNode())
}
