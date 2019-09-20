// Copyright 2019 Istio Authors
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

type mover struct {
	// set when creating the object
	config hugoConfig

	// work space for individual move operations
	srcFile      string
	srcLink      string
	dstFile      string
	dstLink      string
	srcLinkRegex *regexp.Regexp
}

// The regex we use to isolate front matter in a merkdown file
var frontMatterRE = regexp.MustCompile("(?ms)^---$(.*)^---$")

func newMover(config hugoConfig) *mover {
	return &mover{
		config: config,
	}
}

// Does all the work needed to move from src to dst
func (m *mover) move(srcFile string, dstFile string) error {
	var err error

	if m.srcFile, m.srcLink, err = m.validate(srcFile); err != nil {
		return err
	}

	if m.dstFile, m.dstLink, err = m.validate(dstFile); err != nil {
		return err
	}

	if _, err = os.Stat(m.dstFile); err == nil {
		return fmt.Errorf("destination '%s' already exists", m.dstFile)
	}

	re := `\((` + regexp.QuoteMeta(m.srcLink) + `(/|))(#.*|)\)`
	if m.srcLinkRegex, err = regexp.Compile(re); err != nil {
		return err
	}

	// update the content
	for _, contentDir := range m.config.contentRootPerLanguage {
		if err = m.updateContentDir(contentDir); err != nil {
			return err
		}
	}

	// create the target directory
	dstDir := filepath.Dir(m.dstFile)
	if err = os.MkdirAll(dstDir, 01755); err != nil {
		return err
	}

	// move the file
	if err = os.Rename(m.srcFile, m.dstFile); err != nil {
		return fmt.Errorf("unable to move file '%s' to '%s': %v", m.srcFile, m.dstFile, err)
	}

	// all done
	return nil
}

// Ensures the file is a markdown file and lies within a content directory
func (m *mover) validate(file string) (absFile string, link string, err error) {
	if filepath.Ext(file) != ".md" {
		return "", "", fmt.Errorf("'%s' is not a markdown file", file)
	}

	if absFile, err = filepath.Abs(file); err != nil {
		return "", "", fmt.Errorf("unable to access file '%s': %v", file, err)
	}

	absDir := filepath.Dir(absFile)
	for lang, contentDir := range m.config.contentRootPerLanguage {
		// see if the file is within the particular content directory
		if strings.HasPrefix(absFile, contentDir) {
			// now generate the link to the file by using the relative path from the content directory
			if lang == m.config.defaultContentLanguage {
				link = strings.TrimPrefix(absDir, contentDir)
			} else {
				link = "/" + lang + strings.TrimPrefix(absDir, contentDir)
			}

			return
		}
	}

	return "", "", fmt.Errorf("'%s' is not located in a known content root", absFile)
}

// Walk the directory and update links from old page to new page, and add an alias to the page being moved.
// The alias keeps links to the old page location working.
func (m *mover) updateContentDir(contentDir string) error {
	return filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		input, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("unable to read file '%s': %v", path, err)
		}

		rep := "(" + regexp.QuoteMeta(m.dstLink) + "$2$3)"
		output := m.srcLinkRegex.ReplaceAll(input, []byte(rep))

		if path == m.srcFile {
			// the file being moved gets an alias
			if output, err = m.addAliasToFrontMatter(output); err != nil {
				return err
			}
		}

		if err = ioutil.WriteFile(path, output, 0644); err != nil {
			return fmt.Errorf("unable to write file '%s': %v", path, err)
		}

		return nil
	})
}

func (m *mover) addAliasToFrontMatter(input []byte) (output []byte, err error) {
	// extract the existing front matter
	var fm yaml.MapSlice
	if err = yaml.Unmarshal(frontMatterRE.Find(input), &fm); err != nil {
		return nil, fmt.Errorf("unable to decode front matter for file '%s': %v", m.srcFile, err)
	}

	// find the aliases stanza
	var aliases *yaml.MapItem
	for i, v := range fm {
		name := v.Key.(string)
		if name == "aliases" {
			aliases = &fm[i]
			break
		}
	}

	if aliases == nil {
		// no aliases stanza, so add one
		fm = append(fm, yaml.MapItem{
			Key: "aliases",
		})
		aliases = &fm[len(fm)-1]
	}

	var entries []interface{}
	if aliases.Value != nil {
		entries = aliases.Value.([]interface{})
	} else {
		entries = make([]interface{}, 0)
	}
	entries = append(entries, m.srcLink)
	aliases.Value = entries

	tmp, err := yaml.Marshal(fm)
	if err != nil {
		return nil, fmt.Errorf("unable to update front mtter for file '%s': %v", m.srcFile, err)
	}

	// update the front matter
	return frontMatterRE.ReplaceAllLiteral(input, []byte("---\n"+string(tmp)+"---")), nil
}
