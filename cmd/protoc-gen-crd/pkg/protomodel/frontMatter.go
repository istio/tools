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

package protomodel

import (
	"fmt"
	"os"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

type FrontMatter struct {
	Title        string
	Overview     string
	Description  string
	HomeLocation string
	Extra        []string
	Location     LocationDescriptor
}

const (
	titleTag       = "$title: "
	overviewTag    = "$overview: "
	descriptionTag = "$description: "
	locationTag    = "$location: "
	frontMatterTag = "$front_matter: "
)

func checkSingle(name string, old string, line string, tag string) string {
	result := line[len(tag):]
	if old != "" {
		_, _ = fmt.Fprintf(os.Stderr, "%v has more than one %v: %v\n", name, tag, result)
	}
	return result
}

func extractFrontMatter(name string, loc *descriptor.SourceCodeInfo_Location, file *FileDescriptor) FrontMatter {
	title := ""
	overview := ""
	description := ""
	homeLocation := ""
	var extra []string

	for _, para := range loc.LeadingDetachedComments {
		lines := strings.Split(para, "\n")
		for _, l := range lines {
			l = strings.Trim(l, " ")

			if strings.HasPrefix(l, "$") {
				if strings.HasPrefix(l, titleTag) {
					title = checkSingle(name, title, l, titleTag)
				} else if strings.HasPrefix(l, overviewTag) {
					overview = checkSingle(name, overview, l, overviewTag)
				} else if strings.HasPrefix(l, descriptionTag) {
					description = checkSingle(name, description, l, descriptionTag)
				} else if strings.HasPrefix(l, locationTag) {
					homeLocation = checkSingle(name, homeLocation, l, locationTag)
				} else if strings.HasPrefix(l, frontMatterTag) {
					// old way to specify custom front-matter
					extra = append(extra, l[len(frontMatterTag):])
				} else {
					extra = append(extra, l[1:])
				}
			}
		}
	}

	return FrontMatter{
		Title:        title,
		Overview:     overview,
		Description:  description,
		HomeLocation: homeLocation,
		Extra:        extra,
		Location:     newLocationDescriptor(loc, file),
	}
}
