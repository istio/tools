package openapi

import "encoding/json"

// PrimitiveType is the underlying type of a primitive
type PrimitiveType string

const (
	// TypeString is the primitive type string
	TypeString PrimitiveType = "string"

	// TypeInt32 is the primitive type int32
	TypeInt32 PrimitiveType = "int32"

	// TypeBool is the primitive type bool
	TypeBool PrimitiveType = "boolean"
)

// Primitive type
type Primitive struct {
	Type PrimitiveType
}

var _ Type = &Primitive{}

// NewString initializes a new String instance
func NewString() *Primitive {
	return &Primitive{
		Type: TypeString,
	}
}

// NewBool initializes a new Bool instance
func NewBool() *Primitive {
	return &Primitive{
		Type: TypeBool,
	}
}

// NewInt32 initializes a new Int32 instance
func NewInt32() *Primitive {
	return &Primitive{
		Type: TypeInt32,
	}
}

// ToNode implementation
func (p *Primitive) ToNode() *Node {
	return &Node{
		Type: string(p.Type),
	}
}

// MarshalJSON implementation
func (p *Primitive) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.ToNode())
}
