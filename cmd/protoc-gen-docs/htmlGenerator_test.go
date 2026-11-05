// Copyright Istio Authors
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

package main

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"istio.io/tools/pkg/protomodel"
)

func TestGenerateCommentElidesHTMLComments(t *testing.T) {
	cases := []struct {
		name    string
		comment string
		want    string
	}{
		{
			name:    "inline comment",
			comment: "foo <!-- hidden --> bar",
			want:    "<p>foo  bar</p>\n",
		},
		{
			name:    "inline comment after long text",
			comment: "A long description of the field <!-- hidden --> tail",
			want:    "<p>A long description of the field  tail</p>\n",
		},
		{
			name:    "two inline comments",
			comment: "a <!-- one --> b <!-- two --> c",
			want:    "<p>a  b  c</p>\n",
		},
		{
			name:    "multi-line comment",
			comment: "foo <!-- hidden\nstill hidden --> bar",
			want:    "<p>foo\nbar</p>\n",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := &htmlGenerator{model: &protomodel.Model{}}
			loc := protomodel.LocationDescriptor{
				SourceCodeInfo_Location: &descriptor.SourceCodeInfo_Location{
					LeadingComments: proto.String(c.comment),
				},
			}
			g.generateComment(loc, "test")
			got := strings.TrimSuffix(g.buffer.String(), "\n")
			if got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}
