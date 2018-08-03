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

package graph

import (
	"errors"
	"fmt"

	"istio.io/tools/isotope/convert/pkg/graph/script"
)

// validate returns nil if g is valid.
// g is valid if a ServiceGraph:
// - Each of its services only makes requests to other defined services.
// - ConcurrentCommands do not contain other ConcurrentCommands.
func validate(g ServiceGraph) (err error) {
	svcNames := map[string]bool{}
	for _, svc := range g.Services {
		svcNames[svc.Name] = true
	}
	for _, svc := range g.Services {
		err = validateCommands(svc.Script, svcNames)
		if err != nil {
			return
		}
	}
	return
}

func validateCommands(cmds []script.Command, svcNames map[string]bool) error {
	for _, cmd := range cmds {
		switch cmd := cmd.(type) {
		case script.RequestCommand:
			if !svcNames[cmd.ServiceName] {
				return ErrRequestToUndefinedService{cmd.ServiceName}
			}
		case script.ConcurrentCommand:
			err := validateCommands(cmd, svcNames)
			if err != nil {
				return err
			}
			if containsConcurrentCommand([]script.Command(cmd)) {
				return ErrNestedConcurrentCommand
			}
		}
	}
	return nil
}

func containsConcurrentCommand(cmds []script.Command) bool {
	for _, cmd := range cmds {
		if _, ok := cmd.(script.ConcurrentCommand); ok {
			return true
		}
	}
	return false
}

// ErrRequestToUndefinedService is returned when a RequestCommand has a
// ServiceName that is not the name of a defined service.
type ErrRequestToUndefinedService struct {
	ServiceName string
}

func (e ErrRequestToUndefinedService) Error() string {
	return fmt.Sprintf(`cannot call undefined service "%s"`, e.ServiceName)
}

// ErrNestedConcurrentCommand is returned when a ConcurrentCommand contains
// a ConcurrentCommand.
var ErrNestedConcurrentCommand = errors.New(
	"concurrent commands may not be nested")
