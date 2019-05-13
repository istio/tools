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

func TestNoOneof(t *testing.T) {
	obj := &generated.NoOneof{
		FieldA: 1,
		FieldB: "test",
	}
	testSerialization(t, obj)
}

func TestSimpleOneof(t *testing.T) {
	obj := &generated.SimpleOneof{
		FieldA: &generated.SimpleOneof_Name{Name: "test"},
	}
	testSerialization(t, obj)

	obj = &generated.SimpleOneof{
		FieldA: &generated.SimpleOneof_Number{Number: 1},
	}
	testSerialization(t, obj)
}

func TestNestedOneof(t *testing.T) {
	obj := &generated.NestedOneof{
		FieldA: 1,
		FieldB: &generated.NestedOneof_Nested{OneOf: &generated.NestedOneof_Nested_Name{Name: "test"}},
	}
	testSerialization(t, obj)

	obj = &generated.NestedOneof{
		FieldA: 2,
		FieldB: &generated.NestedOneof_Nested{OneOf: &generated.NestedOneof_Nested_Number{Number: 3}},
	}
	testSerialization(t, obj)
}

func TestReferencedOneof(t *testing.T) {
	obj := &generated.ReferencedOneof{
		FieldA: "Test1",
		FieldB: &generated.NestedOneof_Nested{OneOf: &generated.NestedOneof_Nested_Name{Name: "test"}},
	}
	testSerialization(t, obj)

	obj = &generated.ReferencedOneof{
		FieldA: "Test2",
		FieldB: &generated.NestedOneof_Nested{OneOf: &generated.NestedOneof_Nested_Number{Number: 3}},
	}
	testSerialization(t, obj)
}

func TestImportedReference(t *testing.T) {
	obj := &generated.ImportedReference{
		FieldA: 1,
		FieldB: &generated.ExternalOneof{FieldC: 2, FieldD: &generated.ExternalOneof_Name{Name: "test"}},
	}
	testSerialization(t, obj)

	obj = &generated.ImportedReference{
		FieldA: 3,
		FieldB: &generated.ExternalOneof{FieldC: 4, FieldD: &generated.ExternalOneof_Number{Number: 5}},
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
