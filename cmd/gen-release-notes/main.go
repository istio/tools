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
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/russross/blackfriday/v2"
)

type ReleaseNote struct {
	Kind          string        `json:"kind"`
	Area          string        `json:"area"`
	Issue         []string      `json:"issue,omitempty"`
	ReleaseNotes  []string      `json:"releaseNotes"`
	UpgradeNotes  []upgradeNote `json:"upgradeNotes"`
	SecurityNotes []string      `json:"securityNotes"`
	File          string
}

type upgradeNote struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func main() {
	var oldBranch, newBranch, notesDir, templatesDir string
	flag.StringVar(&oldBranch, "oldBranch", "a", "branch to compare against")
	flag.StringVar(&newBranch, "newBranch", "b", "branch containing new files")
	flag.StringVar(&notesDir, "notes", "./notes", "the directory containing release notes")
	flag.StringVar(&templatesDir, "templates", "./templates", "the directory containing release note templates")
	flag.Parse()

	fmt.Printf("Looking for release notes in %s.\n", notesDir)
	releaseNoteFiles, err := getNewFilesInBranch(oldBranch, newBranch, notesDir, "releasenotes/notes")
	if err != nil {
		fmt.Printf("failed to list files: %s", err.Error())
		return
	}
	fmt.Printf("Found %d files.\n", len(releaseNoteFiles))

	fmt.Printf("Looking for markdown templates in %s.\n", templatesDir)
	templateFiles, err := getFilesWithExtension(templatesDir, "md")
	if err != nil {
		fmt.Printf("failed to list files: %s", err.Error())
		return
	}
	fmt.Printf("Found %d files.\n\n", len(templateFiles))

	fmt.Printf("Parsing release notes\n")
	releaseNotes, err := parseReleaseNotesFiles(notesDir, releaseNoteFiles)
	if err != nil {
		fmt.Printf("Unable to read release notes: %s\n", err.Error())
	}

	processTemplates(templatesDir, templateFiles, releaseNotes)
}

func processTemplates(templatesDir string, templateFiles []string, releaseNotes []ReleaseNote) {

	for _, file := range templateFiles {
		output, err := parseTemplate(templatesDir, file, releaseNotes)
		if err != nil {
			fmt.Printf("Failed to parse markdown template: %s", err.Error())
			return
		}

		if err := ioutil.WriteFile(file, []byte(output), 0644); err != nil {
			fmt.Printf("Failed to write markdown: %s", err.Error())
		} else {
			fmt.Printf("Wrote markdown to %s\n", file)
		}

		if err := ioutil.WriteFile(file+".html", []byte(markdownToHTML(output)), 0644); err != nil {
			fmt.Printf("Failed to write HTML: %s", err.Error())
		} else {
			fmt.Printf("Wrote markdown to %s.html\n", file)
		}
	}
}

func parseTemplateFormat(releaseNotes []ReleaseNote, format string) []string {
	noteType := "releaseNotes"
	if strings.Contains(format, "upgradeNotes") {
		noteType = "upgradeNotes"
	} else if strings.Contains(format, "securityNotes") {
		noteType = "securityNotes"
	}

	area := getNoteArea(format)
	action := getActionFromFormat(format)
	fmt.Printf("Notes format: %s type: %s area: %s action:%s\n", format, noteType, area, action)

	return getNotesForTemplateFormat(releaseNotes, noteType, area, action)
}

func getNoteArea(format string) string {
	area := ""
	areaRegexp := regexp.MustCompile("area:[a-zA-Z-]*")
	if match := areaRegexp.FindString(format); match != "" {
		sections := strings.Split(match, ":")
		area = sections[1]
	}
	return area
}

func getActionFromFormat(format string) string {
	action := ""
	actionRegexp := regexp.MustCompile("action:[a-zA-Z]*")
	if match := actionRegexp.FindString(format); match != "" {
		sections := strings.Split(match, ":")
		action = sections[1]
	}
	return action
}

func getNoteAction(note string) string {
	action := ""
	actionRegexp := regexp.MustCompile(`\*\*[A-Z][a-zA-Z]*\*\*`)
	if match := actionRegexp.FindString(note); match != "" {
		action = match[2 : len(match)-2]
	}
	return action
}

func getNotesForTemplateFormat(releaseNotes []ReleaseNote, noteType string, area string, action string) []string {
	parsedNotes := make([]string, 0)
	for _, note := range releaseNotes {
		formatted := ""
		if noteType == "releaseNotes" && note.ReleaseNotes != nil && (note.Area == area || area == "") {
			for _, releaseNote := range note.ReleaseNotes {
				if getNoteAction(releaseNote) == action || action == "" {
					formatted += fmt.Sprintf("- %s %s\n", releaseNote, issuesListToString(note.Issue))
				}
			}
		} else if noteType == "upgradeNotes" {
			for _, upgradeNote := range note.UpgradeNotes {
				if upgradeNote.Content == "" || upgradeNote.Title == "" {
					fmt.Printf("Upgrade note in %s is missing either content or title. Skipping.\n", note.File)
				} else {
					formatted += fmt.Sprintf("## %s\n%s", upgradeNote.Title, upgradeNote.Content)
				}
			}
		} else if noteType == "securityNotes" && note.SecurityNotes != nil {
			for _, securityNote := range note.SecurityNotes {
				formatted += fmt.Sprintf("- %s", securityNote)
			}
		}

		if formatted != "" {
			parsedNotes = append(parsedNotes, formatted)
		}
	}

	return parsedNotes
}

//Bavery_TODO: rewrite
func issuesListToString(issues []string) string {
	issueString := ""
	for _, issue := range issues {
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

//getFilesWithExtension returns the files from filePath with extension extension
func getFilesWithExtension(filePath string, extension string) ([]string, error) {
	directory, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open directory: %s", err.Error())
	}
	defer directory.Close()

	var files []string
	files, err = directory.Readdirnames(0)
	if err != nil {
		return nil, fmt.Errorf("unable to list files for directory %s: %s", filePath, err.Error())
	}

	filesWithExtension := make([]string, 0)
	for _, fileName := range files {
		if strings.HasSuffix(fileName, extension) {
			filesWithExtension = append(filesWithExtension, fileName)
		}
	}

	return filesWithExtension, nil

}

func parseReleaseNotesFiles(filePath string, files []string) ([]ReleaseNote, error) {
	releaseNotes := make([]ReleaseNote, 0)
	for _, file := range files {
		file = path.Join(filePath, file)
		contents, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("unable to open file %s: %s", file, err.Error())
		}

		var releaseNote ReleaseNote
		if err = yaml.Unmarshal(contents, &releaseNote); err != nil {
			return nil, fmt.Errorf("unable to parse release note %s:%s", file, err.Error())
		}
		releaseNote.File = file
		releaseNotes = append(releaseNotes, releaseNote)

	}
	return releaseNotes, nil

}

//markdownToHTML is a wrapper around the blackfriday library to generate HTML previews from markdown
func markdownToHTML(markdown string) string {
	return string(blackfriday.Run([]byte(markdown)))
}

func parseTemplate(filepath string, filename string, releaseNotes []ReleaseNote) (string, error) {
	filename = path.Join(filepath, filename)
	fmt.Printf("Processing %s\n", filename)

	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("unable to open file %s: %s", filename, err.Error())
	}

	comment := regexp.MustCompile("<!--(.*)-->")
	output := string(contents)
	results := comment.FindAllString(string(contents), -1)

	for _, result := range results {
		contents := parseTemplateFormat(releaseNotes, result)
		joinedContents := strings.Join(contents, "\n")
		output = strings.Replace(output, result, joinedContents, -1)
	}

	return output, nil
}

func getNewFilesInBranch(oldBranch string, newBranch string, path string, notesSubpath string) ([]string, error) {
	cmd := fmt.Sprintf("cd %s; git diff-tree -r --diff-filter=AR --name-only --relative=%s '%s' '%s'", path, notesSubpath, oldBranch, newBranch)
	fmt.Printf("Executing: %s\n", cmd)

	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return nil, err
	}

	fmt.Printf("Release notes files: %s\n", out)

	outFiles := strings.Split(string(out), "\n")
	return outFiles[:len(outFiles)-1], nil
}
