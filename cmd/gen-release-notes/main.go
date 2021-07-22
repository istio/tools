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

// golang flags don't accept arrays by default. This adds it.
type flagStrings []string

func (flagString *flagStrings) String() string {
	return strings.Join(*flagString, ",")
}

func (flagString *flagStrings) Set(value string) error {
	*flagString = append(*flagString, value)
	return nil
}

func main() {
	var oldBranch, newBranch, templatesDir, outDir, oldRelease, newRelease, pullRequest string
	var validateOnly bool
	var notesDirs flagStrings

	flag.StringVar(&oldBranch, "oldBranch", "a", "branch to compare against")
	flag.StringVar(&newBranch, "newBranch", "b", "branch containing new files")
	flag.StringVar(&pullRequest, "pullRequest", "", "the pull request to check. Either this or oldBranch & newBranch are required.")
	flag.Var(&notesDirs, "notes", "the directory containing release notes. Repeat for multiple notes directories")
	flag.StringVar(&templatesDir, "templates", "./templates", "the directory containing release note templates")
	flag.StringVar(&outDir, "outDir", ".", "the directory containing release notes")
	flag.BoolVar(&validateOnly, "validateOnly", false, "set to true to perform validation only")
	flag.StringVar(&oldRelease, "oldRelease", "x.y.(z-1)", "old release")
	flag.StringVar(&newRelease, "newRelease", "x.y.z", "new release")
	flag.Parse()

	var releaseNotes []Note
	for _, notesDir := range notesDirs {
		var releaseNoteFiles []string

		fmt.Printf("Looking for release notes in %s.\n", notesDir)
		var err error
		releaseNoteFiles, err = getNewFilesInBranch(oldBranch, newBranch, pullRequest, notesDir, "releasenotes/notes")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to list files: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Printf("Found %d files.\n\n", len(releaseNoteFiles))

		fmt.Printf("Parsing release notes\n")
		releaseNotesEntries, err := parseReleaseNotesFiles(notesDir, releaseNoteFiles)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to read release notes: %s\n", err.Error())
			os.Exit(1)
		}
		releaseNotes = append(releaseNotes, releaseNotesEntries...)
	}

	if len(releaseNotes) < 1 {
		fmt.Fprintf(os.Stderr, "failed to find any release notes.\n")
		// maps to EX_NOINPUT, but more importantly lets us differentiate between no files found and other errors
		os.Exit(66)
	}

	if validateOnly {
		return
	}

	fmt.Printf("\nLooking for markdown templates in %s.\n", templatesDir)
	templateFiles, err := getFilesWithExtension(templatesDir, "md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to list files: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Printf("Found %d files.\n\n", len(templateFiles))

	for _, filename := range templateFiles {
		output, err := populateTemplate(templatesDir, filename, releaseNotes, oldRelease, newRelease)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse template: %s\n", err.Error())
			os.Exit(1)
		}

		if err := createDirIfNotExists(outDir); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create our dir: %s\n", err.Error())
		}
		if err := writeAsMarkdown(path.Join(outDir, filename), output); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write markdown: %s\n", err.Error())
		} else {
			fmt.Printf("Wrote markdown to %s\n", filename)
		}

		if err := writeAsHTML(path.Join(outDir, filename), output); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write HTML: %s\n", err.Error())
		} else {
			fmt.Printf("Wrote markdown to %s.html\n", filename)
		}
	}
}

func createDirIfNotExists(path string) error {
	err := os.MkdirAll(path, 0755)
	if os.IsExist(err) {
		return nil
	}
	return err
}

// writeAsHTML generates HTML from markdown before writing it to a file
func writeAsHTML(filename string, markdown string) error {
	output := string(blackfriday.Run([]byte(markdown)))
	if err := ioutil.WriteFile(filename+".html", []byte(output), 0644); err != nil {
		return err
	}
	return nil
}

// writeAsMarkdown writes markdown to a file
func writeAsMarkdown(filename string, markdown string) error {
	if err := ioutil.WriteFile(filename, []byte(markdown), 0644); err != nil {
		return err
	}
	return nil
}

func parseTemplateFormat(releaseNotes []Note, format string) ([]string, error) {
	template, err := ParseTemplate(format)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %s", err.Error())
	}
	return getNotesForTemplateFormat(releaseNotes, template), nil
}

func getNotesForTemplateFormat(notes []Note, template Template) []string {
	parsedNotes := make([]string, 0)

	for _, note := range notes {
		if template.Type == "releaseNotes" {
			parsedNotes = append(parsedNotes, note.getReleaseNotes(template.Kind, template.Area, template.Action)...)
		} else if template.Type == "upgradeNotes" {
			parsedNotes = append(parsedNotes, note.getUpgradeNotes()...)
		} else if template.Type == "securityNotes" {
			parsedNotes = append(parsedNotes, note.getSecurityNotes()...)
		}
	}
	return parsedNotes
}

// getFilesWithExtension returns the files from filePath with extension extension
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

func parseReleaseNotesFiles(filePath string, files []string) ([]Note, error) {
	notes := make([]Note, 0)
	for _, file := range files {
		file = path.Join(filePath, file)
		contents, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("unable to open file %s: %s", file, err.Error())
		}

		var note Note
		if err = yaml.Unmarshal(contents, &note); err != nil {
			return nil, fmt.Errorf("unable to parse release note %s:%s", file, err.Error())
		}
		note.File = file
		notes = append(notes, note)
		fmt.Printf("found %d upgrade notes, %d release notes, and %d security notes in %s\n", len(note.UpgradeNotes),
			len(note.ReleaseNotes), len(note.SecurityNotes), note.File)

	}
	return notes, nil
}

func populateTemplate(filepath string, filename string, releaseNotes []Note, oldRelease string, newRelease string) (string, error) {
	filename = path.Join(filepath, filename)
	fmt.Printf("Processing %s\n", filename)

	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("unable to open file %s: %s", filename, err.Error())
	}

	comment := regexp.MustCompile("<!--(.*)-->")
	output := string(contents)

	output = strings.Replace(output, "<!--oldRelease-->", oldRelease, -1)
	output = strings.Replace(output, "<!--newRelease-->", newRelease, -1)

	results := comment.FindAllString(output, -1)

	for _, result := range results {
		contents, err := parseTemplateFormat(releaseNotes, result)
		if err != nil {
			return "", fmt.Errorf("unable to parse templates: %s", err.Error())
		}
		joinedContents := strings.Join(contents, "\n")
		output = strings.Replace(output, result, joinedContents, -1)
	}

	return output, nil
}

func getNewFilesInBranch(oldBranch string, newBranch string, pullRequest string, path string, notesSubpath string) ([]string, error) {
	cmd := ""
	if pullRequest != "" {
		cmd = fmt.Sprintf("cd %s; gh pr view %s --json files | jq -r '.files[].path' | grep -E '^%s'", path, pullRequest, notesSubpath)
	} else {
		cmd = fmt.Sprintf("cd %s; git diff-tree -r --diff-filter=AMR --name-only --relative=%s '%s' '%s'", path, notesSubpath, oldBranch, newBranch)
	}
	fmt.Printf("Executing: %s\n", cmd)

	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return nil, err
	}

	outFiles := strings.Split(string(out), "\n")
	return outFiles[:len(outFiles)-1], nil
}
