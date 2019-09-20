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

	"istio.io/tools/cmd/protoc-gen-deepcopy/test/generated"
)

func TestNoTagType(t *testing.T) {
	if checkNoTagTypeDeepCopy(&generated.NoTagType{}) {
		t.Fail()
	}
}

func checkNoTagTypeDeepCopy(value interface{}) bool {
	type NoTagTypeDeepCopy interface {
		DeepCopyInto(*generated.NoTagType)
	}
	_, ok := value.(NoTagTypeDeepCopy)
	return ok
}

func TestTagType(t *testing.T) {
	if !checkTagTypeDeepCopy(&generated.TagType{}) {
		t.Fail()
	}
}

func checkTagTypeDeepCopy(value interface{}) bool {
	type TagTypeDeepCopy interface {
		DeepCopyInto(*generated.TagType)
	}
	_, ok := value.(TagTypeDeepCopy)
	return ok
}
func TestSeparatedTagType(t *testing.T) {
	if !checkSeparatedTagTypeDeepCopy(&generated.SeparatedTagType{}) {
		t.Fail()
	}
}

func checkSeparatedTagTypeDeepCopy(value interface{}) bool {
	type SeparatedTagTypeDeepCopy interface {
		DeepCopyInto(*generated.SeparatedTagType)
	}
	_, ok := value.(SeparatedTagTypeDeepCopy)
	return ok
}
