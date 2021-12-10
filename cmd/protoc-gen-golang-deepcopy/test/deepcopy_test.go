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

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"istio.io/tools/cmd/protoc-gen-deepcopy/test/generated"
)

func TestTagType(t *testing.T) {
	if !checkTagTypeDeepCopy(&generated.TagType{}) {
		t.Fail()
	}
}

func checkTagTypeDeepCopy(value interface{}) bool {
	type TagTypeDeepCopy interface {
		DeepCopyInto(*generated.TagType)
		DeepCopy() *generated.TagType
	}
	_, ok := value.(TagTypeDeepCopy)
	return ok
}

func TestTypeWithRepeatedField(t *testing.T) {
	in := &generated.RepeatedFieldType{
		Ns: []string{"ns-1", "ns-2"},
	}
	out := &generated.RepeatedFieldType{}
	in.DeepCopyInto(out)
	if !cmp.Equal(in, out, protocmp.Transform()) {
		t.Fatalf("Deepcopy of proto(DeepCopyInto) is not equal. got: %v, want: %v", out, in)
	}

	out = in.DeepCopy()
	if !cmp.Equal(in, out, protocmp.Transform()) {
		t.Fatalf("Deepcopy of proto(DeepCopy) is not equal. got: %v, want: %v", out, in)
	}

	outInterface := in.DeepCopyInterface()
	outPb, ok := outInterface.(*generated.RepeatedFieldType)
	if !ok {
		t.Fatalf("DeepCopyInterface was not a proto message, was %T", outInterface)
	}
	if !cmp.Equal(in, outPb, protocmp.Transform()) {
		t.Fatalf("Deepcopy of proto(DeepCopy) is not equal. got: %v, want: %v", outPb, in)
	}
}
