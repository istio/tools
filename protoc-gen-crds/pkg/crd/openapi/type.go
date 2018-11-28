package openapi

import "encoding/json"

// Type represents a basic Open API type.
type Type interface {
	json.Marshaler
	ToNode() *Node
}
