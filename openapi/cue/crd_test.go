package main

import (
	"fmt"
	"testing"

	"cuelang.org/go/encoding/openapi"
)

func TestOverrideFieldValue(t *testing.T) {
	info := openapi.OrderedMap{}
	info.SetAll([]openapi.KeyValue{{"title", "test"}, {"version", "v1"}})
	test2 := openapi.OrderedMap{}
	test2.Set("c1", "d1")
	testKeyValuePairs := []openapi.KeyValue{
		{
			Key:   "a",
			Value: "b",
		},
		{
			Key:   "c",
			Value: test2,
		},
	}
	testinfo := openapi.OrderedMap{}
	testinfo.SetAll(testKeyValuePairs)

	v := overrideFieldValue("c.c1", info, testinfo)
	fmt.Println(v)
}
