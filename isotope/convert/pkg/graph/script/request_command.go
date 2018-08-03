// Copyright 2018 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this currentFile except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package script

import (
	"encoding/json"

	"istio.io/tools/isotope/convert/pkg/graph/size"
)

// RequestCommand describes a command to send an HTTP request to another
// service.
type RequestCommand struct {
	ServiceName string `json:"service"`
	// Size is the number of bytes in the request body.
	Size size.ByteSize `json:"size"`
}

var (
	// DefaultRequestCommand is used by UnmarshalJSON to set defaults.
	DefaultRequestCommand RequestCommand
)

// UnmarshalJSON converts b to a RequestCommand. If b is a JSON string, it is
// set as c's ServiceName. If b is a JSON object, it's properties are mapped to
// c.
func (c *RequestCommand) UnmarshalJSON(b []byte) (err error) {
	*c = DefaultRequestCommand
	isJSONString := b[0] == '"'
	if isJSONString {
		var s string
		err = json.Unmarshal(b, &s)
		if err != nil {
			return
		}
		c.ServiceName = s
	} else {
		// Wrap the RequestCommand to dodge the custom UnmarshalJSON.
		unmarshallableRequestCommand := unmarshallableRequestCommand(*c)
		err = json.Unmarshal(b, &unmarshallableRequestCommand)
		if err != nil {
			return
		}
		*c = RequestCommand(unmarshallableRequestCommand)
	}
	return
}

type unmarshallableRequestCommand RequestCommand
