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
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"sigs.k8s.io/yaml"

	"istio.io/tools/pkg/markdown"
	"istio.io/tools/pkg/schemavalidation"
)

//go:embed release_notes_schema.json
var schema []byte

//go:embed templates/*.md
var rawTemplates embed.FS

var templates = func() fs.ReadDirFS {
	s, _ := fs.Sub(rawTemplates, "templates")
	return s.(fs.ReadDirFS)
}()

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
	var oldBranch, newBranch, outDir, oldRelease, newRelease string
	var validateOnly, checkLabel bool
	var notesDirs flagStrings

	flag.StringVar(&oldBranch, "oldBranch", "a", "branch to compare against")
	flag.StringVar(&newBranch, "newBranch", "b", "branch containing new files")
	flag.Var(&notesDirs, "notes", "the directory containing release notes. Repeat for multiple notes directories")
	flag.StringVar(&outDir, "outDir", ".", "the directory containing release notes")
	flag.BoolVar(&validateOnly, "validateOnly", false, "set to true to perform validation only")
	flag.BoolVar(&checkLabel, "checkLabel", false, "set to true to check PR has release notes OR a release-notes-none label")
	flag.StringVar(&oldRelease, "oldRelease", "x.y.(z-1)", "old release")
	flag.StringVar(&newRelease, "newRelease", "x.y.z", "new release")
	flag.Parse()

	// Detect if we are in CI, if so we are checking a single PR...
	RepoOwner := os.Getenv("REPO_OWNER")
	RepoName := os.Getenv("REPO_NAME")
	PullRequest := os.Getenv("PULL_NUMBER")
	var pullRequest string
	if RepoOwner != "" && RepoName != "" && PullRequest != "" {
		pullRequest = fmt.Sprintf("https://github.com/%s/%s/pull/%s", RepoOwner, RepoName, PullRequest)
	}

	if len(notesDirs) == 0 {
		notesDirs = []string{"."}
	}

	var releaseNotes []Note
	for _, notesDir := range notesDirs {

		log.Printf("Looking for release notes in %q.\n", notesDir)

		releaseNotesDir := "releasenotes/notes"
		if _, err := os.Stat(notesDir); os.IsNotExist(err) {
			log.Printf("Could not find repository -- directory %s does not exist.\n", notesDir)
			os.Exit(1)
		}

		if _, err := os.Stat(filepath.Join(notesDir, releaseNotesDir)); os.IsNotExist(err) {
			log.Printf("Could not find release notes directory -- %s does not exist.\n", filepath.Join(notesDir, releaseNotesDir))
			os.Exit(2)
		}

		branchInfo, err := getNewFilesInBranch(oldBranch, newBranch, pullRequest, notesDir, releaseNotesDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to list files: %s\n", err.Error())
			os.Exit(1)
		}
		log.Printf("Found %d release note files.\n", len(branchInfo.ReleaseNoteFiles))

		if checkLabel {
			log.Println("Checking label or release notes are present...")
			if err := checkReleaseNotesLabel(branchInfo); err != nil {
				os.Exit(1)
			}
		}

		log.Printf("Parsing release notes\n")
		releaseNotesEntries, err := parseReleaseNotesFiles(notesDir, branchInfo.ReleaseNoteFiles)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to read release notes: %s\n", err.Error())
			os.Exit(1)
		}
		releaseNotes = append(releaseNotes, releaseNotesEntries...)
	}

	if validateOnly {
		return
	}

	templateFiles, err := templates.ReadDir(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to list files: %s\n", err.Error())
		os.Exit(1)
	}
	log.Printf("Found %d files.\n\n", len(templateFiles))

	if err := createDirIfNotExists(outDir); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create our dir: %s\n", err.Error())
	}
	for _, f := range templateFiles {
		filename := f.Name()
		output, err := populateTemplate(filename, releaseNotes, oldRelease, newRelease)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse template: %s\n", err.Error())
			os.Exit(1)
		}

		if err := writeAsMarkdown(path.Join(outDir, filename), output); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write markdown: %s\n", err.Error())
		} else {
			log.Printf("Wrote markdown to %s\n", filename)
		}

		if err := writeAsHTML(path.Join(outDir, filename), output); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write HTML: %s\n", err.Error())
		} else {
			log.Printf("Wrote markdown to %s.html\n", filename)
		}
	}
}

// Check we have release-notes-none label OR we have release notes
func checkReleaseNotesLabel(info PRInfo) error {
	if info.HasReleaseNoteNoneLabel {
		log.Printf("Found %q label. This pull request will not include release notes.\n", ReleaseNoteNone)
		return nil
	}
	if len(info.ReleaseNoteFiles) > 0 {
		log.Printf("%d release notes found.\n", len(info.ReleaseNoteFiles))
		return nil
	}
	newURL := fmt.Sprintf("https://github.com/%s/%s/new/%s/releasenotes/notes", info.Author, os.Getenv("PULL_HEAD_REF"), os.Getenv("PULL_BASE_REF"))
	// nolint: lll
	log.Printf(`
ERROR: Missing release notes and missing %q label.

If this pull request contains user facing changes, please create a release note based on the template: https://github.com/istio/istio/blob/master/releasenotes/template.yaml by going here: %s.

Release notes documentation can be found here: https://github.com/istio/istio/tree/master/releasenotes.

If this pull request has no user facing changes, please add the %qlabel to the pull request. Note that the test will have to be manually retriggered (/retest) after adding the label.
`, ReleaseNoteNone, newURL, ReleaseNoteNone)
	return fmt.Errorf("missing release notes and missing %q label", ReleaseNoteNone)
}

func createDirIfNotExists(path string) error {
	err := os.MkdirAll(path, 0o755)
	if os.IsExist(err) {
		return nil
	}
	return err
}

// writeAsHTML generates HTML from markdown before writing it to a file
func writeAsHTML(filename string, input string) error {
	output := string(markdown.Run([]byte(input)))
	if err := os.WriteFile(filename+".html", []byte(output), 0o644); err != nil {
		return err
	}
	return nil
}

// writeAsMarkdown writes markdown to a file
func writeAsMarkdown(filename string, markdown string) error {
	if err := os.WriteFile(filename, []byte(markdown), 0o644); err != nil {
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

func parseReleaseNotesFiles(filePath string, files []string) ([]Note, error) {
	notes := make([]Note, 0)
	for _, file := range files {
		file = path.Join(filePath, file)
		contents, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("unable to open file %s: %s", file, err.Error())
		}

		if err := schemavalidation.Validate(contents, schema); err != nil {
			return nil, err
		}

		var note Note
		if err = yaml.Unmarshal(contents, &note); err != nil {
			return nil, fmt.Errorf("unable to parse release note %s:%s", file, err.Error())
		}
		note.File = file
		notes = append(notes, note)
		log.Printf("found %d upgrade notes, %d release notes, and %d security notes in %s\n", len(note.UpgradeNotes),
			len(note.ReleaseNotes), len(note.SecurityNotes), note.File)
	}
	return notes, nil
}

func populateTemplate(filename string, releaseNotes []Note, oldRelease string, newRelease string) (string, error) {
	log.Printf("Processing %s\n", filename)

	contents, err := fs.ReadFile(templates, filename)
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

type prView struct {
	Files []struct {
		Path string `json:"path"`
	} `json:"files"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
}

func ReadGithubPR(path string, pullRequest string, notesSubpath string) (PRInfo, error) {
	c := exec.Command(
		"gh",
		"pr",
		"view",
		pullRequest,
		"--json=files,labels,author",
	)
	c.Dir = path
	log.Printf("Executing: %s\n", strings.Join(c.Args, " "))
	out, err := c.CombinedOutput()
	if err != nil {
		return PRInfo{}, fmt.Errorf("received error running GH: %s", err.Error())
	}

	var prResults prView
	if err := json.Unmarshal(out, &prResults); err != nil {
		return PRInfo{}, fmt.Errorf("failed to parse GH results: %s", err.Error())
	}

	var results []string
	for _, val := range prResults.Files {
		if strings.Contains(val.Path, notesSubpath) {
			// Only add file if it exists. May have been deleted in this PR.
			if _, err := os.Stat(val.Path); !os.IsNotExist(err) {
				results = append(results, val.Path)
			}
		}
	}

	info := PRInfo{
		Author:                  prResults.Author.Login,
		ReleaseNoteFiles:        results,
		HasReleaseNoteNoneLabel: false,
	}
	for _, l := range prResults.Labels {
		if l.Name == ReleaseNoteNone {
			info.HasReleaseNoteNoneLabel = true
			break
		}
	}

	return info, nil
}

const ReleaseNoteNone = "release-notes-none"

type PRInfo struct {
	Author                  string
	ReleaseNoteFiles        []string
	HasReleaseNoteNoneLabel bool
}

func getNewFilesInBranch(oldBranch string, newBranch string, pullRequest string, path string, notesSubpath string) (PRInfo, error) {
	// if there's a pull request, we can just get the changed files from GitHub. If not, we have to do it manually.
	if pullRequest != "" {
		return ReadGithubPR(path, pullRequest, notesSubpath)
	}

	c := exec.Command(
		"git",
		"diff-tree",
		"-r",
		"--diff-filter=AMR",
		"--name-only",
		"--relative="+notesSubpath,
		oldBranch,
		newBranch,
	)
	c.Dir = path
	log.Printf("Executing: %s\n", strings.Join(c.Args, " "))
	out, err := c.CombinedOutput()
	if err != nil {
		return PRInfo{}, err
	}
	outFiles := strings.Split(string(out), "\n")

	// the ReadGithubPR(path, pullRequest, notesSubpath) method returns file names which are relative to the repo path.
	// the git diff-tree is relative to the notesSupbpath, so we need to add the subpath back to the filenames.
	outFileswithPath := []string{}
	for _, f := range outFiles[:len(outFiles)-1] { // skip the last file which is empty
		outFileswithPath = append(outFileswithPath, filepath.Join(notesSubpath, f))
	}

	return PRInfo{ReleaseNoteFiles: outFileswithPath}, nil
}
