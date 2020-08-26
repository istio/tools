// Copyright 2020 Istio Authors
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
	"fmt"
	"regexp"
	"strings"
)

type Template struct {
	Area   string
	Type   string
	Action string
}

func (tmpl Template) parseAction(line string) string {
	action := ""
	actionRegexp := regexp.MustCompile("action:[a-zA-Z]*")
	if match := actionRegexp.FindString(line); match != "" {
		sections := strings.Split(match, ":")
		action = sections[1]
	}
	return action
}

func (tmpl Template) parseArea(line string) string {
	area := ""
	areaRegexp := regexp.MustCompile("area:[a-zA-Z-]*")
	if match := areaRegexp.FindString(line); match != "" {
		sections := strings.Split(match, ":")
		area = sections[1]
	}
	return area
}

func (tmpl Template) parseType(line string) string {
	if strings.Contains(line, "releaseNotes") {
		return "releaseNotes"
	} else if strings.Contains(line, "upgradeNotes") {
		return "upgradeNotes"
	} else if strings.Contains(line, "securityNotes") {
		return "securityNotes"
	}
	return ""
}

func ParseTemplate(line string) (Template, error) {
	var tmpl Template
	tmpl.Area = tmpl.parseArea(line)
	tmpl.Action = tmpl.parseAction(line)
	tmpl.Type = tmpl.parseType(line)

	if tmpl.Type != "" {
		fmt.Printf("Processed template %s. Area:%s action:%s type:%s\n", line, tmpl.Area, tmpl.Action, tmpl.Type)
	} else {
		return Template{}, fmt.Errorf("unable to process template: %s; ignoring", line)
	}

	return tmpl, nil
}
