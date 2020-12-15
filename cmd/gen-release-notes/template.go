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
	Kind   string
}

func (tmpl Template) parseAction(line string) string {
	actionRegexp := regexp.MustCompile("action:[a-zA-Z]*")
	return parseField(line, *actionRegexp)
}

func (tmpl Template) parseArea(line string) string {
	areaRegexp := regexp.MustCompile("area:[!a-zA-Z-]*")
	return parseField(line, *areaRegexp)
}

func (tmpl Template) parseKind(line string) string {
	areaRegexp := regexp.MustCompile("kind:[!a-zA-Z-]*")
	return parseField(line, *areaRegexp)
}

func parseField(line string, regex regexp.Regexp) string {
	field := ""
	if match := regex.FindString(line); match != "" {
		sections := strings.Split(match, ":")
		field = sections[1]
	}
	return field
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
	tmpl.Kind = tmpl.parseKind(line)
	tmpl.Action = tmpl.parseAction(line)
	tmpl.Type = tmpl.parseType(line)

	if tmpl.Type != "" {
		fmt.Printf("Processed template %s. Kind: %s Area:%s action:%s type:%s\n", line, tmpl.Kind, tmpl.Area, tmpl.Action, tmpl.Type)
	} else {
		return Template{}, fmt.Errorf("unable to process template: %s; ignoring", line)
	}

	return tmpl, nil
}
