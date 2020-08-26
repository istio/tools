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

func main() {
	var oldBranch, newBranch, notesDir, templatesDir, notesFile string
	var validateOnly bool
	flag.StringVar(&oldBranch, "oldBranch", "a", "branch to compare against")
	flag.StringVar(&newBranch, "newBranch", "b", "branch containing new files")
	flag.StringVar(&notesDir, "notes", "./notes", "the directory containing release notes")
	flag.StringVar(&notesFile, "notesFile", "", "a specific notes file to parse")
	flag.StringVar(&templatesDir, "templates", "./templates", "the directory containing release note templates")
	flag.BoolVar(&validateOnly, "validateOnly", false, "set to true to perform validation only")
	flag.Parse()

	var releaseNoteFiles []string
	if notesFile != "" {
		fmt.Printf("Parsing %s\n", notesFile)
		releaseNoteFiles = []string{notesFile}
	} else {
		fmt.Printf("Looking for release notes in %s.\n", notesDir)
		var err error
		releaseNoteFiles, err = getNewFilesInBranch(oldBranch, newBranch, notesDir, "releasenotes/notes")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to list files: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Printf("Found %d files.\n\n", len(releaseNoteFiles))
	}

	if len(releaseNoteFiles) < 1 {
		fmt.Fprintf(os.Stderr, "failed to find any release notes files.\n")
		os.Exit(1)
	}

	fmt.Printf("Parsing release notes\n")
	releaseNotes, err := parseReleaseNotesFiles(notesDir, releaseNoteFiles)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read release notes: %s\n", err.Error())
		os.Exit(1)
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
		output, err := populateTemplate(templatesDir, filename, releaseNotes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse template: %s\n", err.Error())
			os.Exit(1)
		}

		if err := writeAsMarkdown(filename, output); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write markdown: %s\n", err.Error())
		} else {
			fmt.Printf("Wrote markdown to %s\n", filename)
		}

		if err := writeAsHTML(filename, output); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write HTML: %s\n", err.Error())
		} else {
			fmt.Printf("Wrote markdown to %s.html\n", filename)
		}
	}
}

//writeAsHTML generates HTML from markdown before writing it to a file
func writeAsHTML(filename string, markdown string) error {
	output := string(blackfriday.Run([]byte(markdown)))
	if err := ioutil.WriteFile(filename+".html", []byte(output), 0644); err != nil {
		return err
	}
	return nil
}

//writeAsMarkdown writes markdown to a file
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
			parsedNotes = append(parsedNotes, note.getReleaseNotes(template.Area, template.Action)...)
		} else if template.Type == "upgradeNotes" {
			parsedNotes = append(parsedNotes, note.getUpgradeNotes()...)
		} else if template.Type == "securityNotes" {
			parsedNotes = append(parsedNotes, note.getSecurityNotes()...)
		}
	}
	return parsedNotes
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

func populateTemplate(filepath string, filename string, releaseNotes []Note) (string, error) {
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
		contents, err := parseTemplateFormat(releaseNotes, result)
		if err != nil {
			return "", fmt.Errorf("unable to parse templates: %s", err.Error())
		}
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

	outFiles := strings.Split(string(out), "\n")
	return outFiles[:len(outFiles)-1], nil
}
