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
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strings"
)

type Note struct {
	Kind          string         `json:"kind"`
	Area          string         `json:"area"`
	Docs          []string       `json:"docs,omitempty"`
	Issues        []string       `json:"issue,omitempty"`
	ReleaseNotes  []releaseNote  `json:"releaseNotes"`
	UpgradeNotes  []upgradeNote  `json:"upgradeNotes"`
	SecurityNotes []securityNote `json:"securityNotes"`
	File          string
}

func (note Note) getIssues() string {
	issueString := ""
	for _, issue := range note.Issues {
		if issueString != "" {
			issueString += ","
		}
		if strings.Contains(issue, "github.com") {
			issueNumber := path.Base(issue)
			issueString += fmt.Sprintf("([Issue #%s](%s))", issueNumber, issue)
		} else {
			issueString += fmt.Sprintf("([Issue #%s](https://github.com/istio/istio/issues/%s))", issue, issue)
		}
	}
	return issueString
}

func (note Note) getDocs() string {
	docsString := ""
	for _, docsEntry := range note.Docs {
		entryParts := strings.SplitN(docsEntry[1:], "]", 2)
		if len(entryParts) != 2 {
			continue
		}
		docsString += fmt.Sprintf("([%s](%s))", entryParts[0], entryParts[1])
	}
	return docsString
}

func filterNote(templateFilter string, noteFilter string) bool {
	if templateFilter == "" {
		return true
	} else if templateFilter == noteFilter {
		return true
	} else if templateFilter[0] == '!' && templateFilter[1:] != noteFilter {
		return true
	}
	return false

}

func (note Note) getReleaseNotes(kind string, area string, action string) []string {
	notes := make([]string, 0)

	for _, releaseNote := range note.ReleaseNotes {
		if filterNote(kind, note.Kind) &&
			filterNote(area, note.Area) &&
			filterNote(action, releaseNote.Action) {
			noteEntry := fmt.Sprintf("%s %s %s\n", releaseNote, note.getDocs(), note.getIssues())
			if noteEntry != "" {
				notes = append(notes, noteEntry)
			}
		}
	}
	return notes
}

func (note Note) getUpgradeNotes() []string {
	notes := make([]string, 0)
	for _, upgradeNote := range note.UpgradeNotes {
		notes = append(notes, upgradeNote.String())
	}
	return notes
}

func (note Note) getSecurityNotes() []string {
	notes := make([]string, 0)
	for _, securityNote := range note.SecurityNotes {
		notes = append(notes, securityNote.String())
	}
	return notes
}

type upgradeNote struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (note *upgradeNote) UnmarshalJSON(data []byte) error {
	type noteIntType upgradeNote
	var noteInt noteIntType
	if err := json.Unmarshal(data, &noteInt); err != nil {
		return err
	}

	if noteInt.Title == "" {
		return fmt.Errorf("upgrade note title cannot be empty")
	}
	note.Title = noteInt.Title

	if noteInt.Content == "" {
		return fmt.Errorf("upgrade note body cannot be empty")
	}
	note.Content = noteInt.Content
	return nil

}

func (note upgradeNote) String() string {
	return fmt.Sprintf("## %s\n%s", note.Title, note.Content)
}

type releaseNote struct {
	Value  string
	Action string
	Issues []string
}

func (note *releaseNote) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &note.Value); err != nil {
		return err
	}
	if note.Value == "" {
		return fmt.Errorf("value missing for note: %s", note.Value)
	}

	note.Action = note.getAction(note.Value)
	if note.Action == "" {
		return fmt.Errorf("unable to determine action for note: %s; notes must start with an action and be of the form"+
			"**Action** {text} with an action listed here: https://github.com/istio/istio/tree/master/releasenotes#release-notes", note.Value)
	}

	//TODO: Externalize this... we should validate this and action. However, they should not live in code.
	if note.Action != "Added" && note.Action != "Deprecated" && note.Action != "Enabled" &&
		note.Action != "Fixed" && note.Action != "Optimized" && note.Action != "Improved" &&
		note.Action != "Removed" && note.Action != "Upgraded" && note.Action != "Updated" && note.Action != "Promoted" {
		return fmt.Errorf("action %s is not allowed;reference "+
			"https://github.com/istio/istio/tree/master/releasenotes#release-notes for a list of allowed actions", note.Action)
	}

	return nil
}

func (note releaseNote) getAction(line string) string {
	action := ""
	actionRegexp := regexp.MustCompile(`\*\*[A-Z][a-zA-Z]*\*\*`)
	if match := actionRegexp.FindString(line); match != "" {
		action = match[2 : len(match)-2]
	}
	return action
}

func (note releaseNote) String() string {
	return fmt.Sprintf("- %s", note.Value)
}

type securityNote string

func (note securityNote) String() string {
	return fmt.Sprintf("- %s", string(note))
}
