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
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"

	"istio.io/tools/cmd/protoc-gen-jsonshim/test/generated"
)

func TestSimpleCase(t *testing.T) {
	obj := &generated.Simple{
		FieldA: 1,
		FieldB: "test",
		FieldC: &generated.Simple_Name{
			Name: "test",
		},
	}
	testSerialization(t, obj)
}

func TestSimpleCaseWithMap(t *testing.T) {
	obj := &generated.SimpleWithMap{
		FieldA: 1,
		FieldB: "test",
		FieldC: map[string]string{
			"test1": "test1",
		},
	}
	testSerialization(t, obj)
}

func TestNestedMap(t *testing.T) {
	obj := &generated.SimpleWithMap{
		FieldA: 1,
		FieldD: &generated.SimpleWithMap_Nested{
			NestedFieldD: map[string]string{
				"test1": "test1",
			},
		},
	}
	testSerialization(t, obj)
}

func TestReferencedMap(t *testing.T) {
	obj := &generated.ReferencedMap{
		FieldA: "Test1",
		FieldB: &generated.SimpleWithMap_Nested{
			NestedFieldD: map[string]string{
				"test1": "test1",
			},
		},
	}
	testSerialization(t, obj)
}

func TestImportedReference(t *testing.T) {
	obj := &generated.ImportedReference{
		FieldA: 1,
		FieldB: &generated.ExternalSimple{
			FieldC: 1,
			FieldD: &generated.ExternalSimple_ExternalNested{
				FieldA: map[string]string{
					"test1": "test1",
				},
			},
		},
	}
	testSerialization(t, obj)
}

func testSerialization(t *testing.T, obj proto.Message) {
	pbm := &jsonpb.Marshaler{}
	pbBuf := &bytes.Buffer{}
	jsonBytes, _ := json.Marshal(obj)
	pbm.Marshal(pbBuf, obj)
	if !reflect.DeepEqual(pbBuf.Bytes(), jsonBytes) {
		t.Errorf("jsonbp and json marshaled different output: %s vs %s", pbBuf, jsonBytes)
	}

	out1 := newObject(obj)
	pbum := &jsonpb.Unmarshaler{}
	err := pbum.Unmarshal(bytes.NewReader(jsonBytes), out1)
	if err != nil || !reflect.DeepEqual(out1, obj) {
		t.Errorf("jsonpb.Unmarshal() objects not equal: %v vs %v", out1, obj)
		t.Errorf("serialized json: %s", jsonBytes)
		t.Errorf("error: %v", err)
	}

	out2 := newObject(obj)
	err = json.Unmarshal(pbBuf.Bytes(), out2)
	if err != nil || !reflect.DeepEqual(out2, obj) {
		t.Errorf("json.Unmarshal() objects not equal: %v vs %v", out2, obj)
		t.Errorf("serialized jsonpb: %s", pbBuf)
		t.Errorf("error: %v", err)
	}
}

func newObject(source proto.Message) proto.Message {
	return reflect.New(reflect.TypeOf(source).Elem()).Interface().(proto.Message)
}
