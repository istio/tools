// Copyright 2019 Istio Authors
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package test

import (
	"testing"

	"google.golang.org/protobuf/proto"

	v1 "istio.io/tools/cmd/protoc-gen-alias/test/v1"
	"istio.io/tools/cmd/protoc-gen-alias/test/v1alpha"
)

func TestSimpleCase(t *testing.T) {
	concrete := &v1.Simple{
		FieldA: 1,
		FieldB: "test",
		FieldC: &v1.Simple_Name{
			Name: "test",
		},
	}
	alias := &v1alpha.Simple{
		FieldA: 1,
		FieldB: "test",
		FieldC: &v1alpha.Simple_Name{
			Name: "test",
		},
	}
	mixedAliasFirst := &v1alpha.Simple{
		FieldA: 1,
		FieldB: "test",
		FieldC: &v1.Simple_Name{
			Name: "test",
		},
	}
	mixedConcreteFirst := &v1.Simple{
		FieldA: 1,
		FieldB: "test",
		FieldC: &v1alpha.Simple_Name{
			Name: "test",
		},
	}
	// Test we can do proto operations
	proto.Equal(concrete, alias)
	proto.Equal(concrete, mixedConcreteFirst)
	proto.Equal(concrete, mixedAliasFirst)
	if proto.MessageName(mixedConcreteFirst).Name() != "Simple" {
		t.Errorf("proto name should be Simple")
	}
	if proto.MessageName(mixedAliasFirst).Name() != "Simple" {
		t.Errorf("proto name should be Simple")
	}
}
