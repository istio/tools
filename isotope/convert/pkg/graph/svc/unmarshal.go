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

package svc

import (
	"encoding/json"
	"errors"

	"istio.io/tools/isotope/convert/pkg/graph/svctype"
)

var (
	// DefaultService is used by UnmarshalJSON and describes the default settings.
	DefaultService = Service{Type: svctype.ServiceHTTP, NumReplicas: 1}
)

// UnmarshalJSON converts b to a Service, applying the default values from
// DefaultService.
func (svc *Service) UnmarshalJSON(b []byte) (err error) {
	unmarshallable := unmarshallableService(DefaultService)
	err = json.Unmarshal(b, &unmarshallable)
	if err != nil {
		return
	}
	*svc = Service(unmarshallable)
	if svc.Name == "" {
		err = ErrEmptyName
		return
	}
	if svc.Version == "" {
		svc.Version = "v1"
	}
	return
}

type unmarshallableService Service

// ErrEmptyName is returned when attempting to parse JSON without an empty name
// field.
var ErrEmptyName = errors.New("services must have a name")
