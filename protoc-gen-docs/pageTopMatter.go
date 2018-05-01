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

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

type pageTopMatter struct {
	title        string
	overview     string
	homeLocation string
	frontMatter  []string
}

const (
	titleTag       = "$title: "
	overviewTag    = "$overview: "
	locationTag    = "$location: "
	frontMatterTag = "$front_matter: "
)

func checkSingle(name string, old string, line string, tag string) string {
	result := line[len(tag):]
	if old != "" {
		fmt.Fprint(os.Stderr, "%v has more than one %v: %v\n", name, tag, result)
	}
	return result
}

func makeTopMatter(name string, loc *descriptor.SourceCodeInfo_Location) *pageTopMatter {
	hasTopMatter := false
	title := ""
	overview := ""
	homeLocation := ""
	var frontMatter []string = nil
	for _, para := range loc.LeadingDetachedComments {
		lines := strings.Split(para, "\n")
		for _, l := range lines {
			l = strings.Trim(l, " ")

			empty := false
			if strings.HasPrefix(l, titleTag) {
				title = checkSingle(name, title, l, titleTag)
			} else if strings.HasPrefix(l, overview) {
				overview = checkSingle(name, overview, l, overviewTag)
			} else if strings.HasPrefix(l, locationTag) {
				homeLocation = checkSingle(name, homeLocation, l, locationTag)
			} else if strings.HasPrefix(l, frontMatterTag) {
				additional := l[len(frontMatterTag):]
				frontMatter = append(frontMatter, additional)
			} else {
				empty = true
			}
			if !empty {
				hasTopMatter = true
			}
		}
	}
	if hasTopMatter {
		return &pageTopMatter{
			title:        title,
			overview:     overview,
			homeLocation: homeLocation,
			frontMatter:  frontMatter,
		}
	}
	return nil
}
