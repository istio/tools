// Copyright 2018 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pmodel

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"istio.io/tools/kubernetes/resource"
)

func GetOptionBool(options *descriptor.MessageOptions, desc *proto.ExtensionDesc) (bool, error) {
	ex, err := proto.GetExtension(options, desc)
	if err != nil {
		if err == proto.ErrMissingExtension {
			return false, nil
		}
		return false, err
	}

	pval, ok := ex.(*bool)
	if !ok {
		return false, fmt.Errorf("option is not a boolean: %v (%v)", ex, desc)
	}

	return *pval, nil
}

func GetOptionScope(options *descriptor.MessageOptions, desc *proto.ExtensionDesc) (resource.Scope, error) {
	ex, err := proto.GetExtension(options, desc)
	if err != nil {
		if err == proto.ErrMissingExtension {
			return 0, nil
		}
		return 0, err
	}

	pval, ok := ex.(*resource.Scope)
	if !ok {
		return 0, fmt.Errorf("option is not a Scope resource: %v (%v)", ex, desc)
	}

	return *pval, nil
}

func GetOptionInt(options *descriptor.MessageOptions, desc *proto.ExtensionDesc) (int, error) {
	ex, err := proto.GetExtension(options, desc)
	if err != nil {
		if err == proto.ErrMissingExtension {
			return 0, nil
		}
		return 0, err
	}

	pval, ok := ex.(*int)
	if !ok {
		return 0, fmt.Errorf("option is not an int: %v (%v)", ex, desc)
	}

	return *pval, nil
}

func GetOptionString(options *descriptor.MessageOptions, desc *proto.ExtensionDesc) (string, error) {
	ex, err := proto.GetExtension(options, desc)
	if err != nil {
		if err == proto.ErrMissingExtension {
			return "", nil
		}
		return "", err
	}

	pval, ok := ex.(*string)
	if !ok {
		return "", fmt.Errorf("option is not a string: %v (%v)", ex, desc)
	}

	return *pval, nil
}
