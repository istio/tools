// Copyright Istio Authors
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
	"path"
)

func main() {
	var report bool
	var dump bool
	var csv bool
	var mirror bool
	var config string

	flag.BoolVar(&report, "report", false, "Generate a report of all license usage.")
	flag.BoolVar(&dump, "dump", false, "Generate a dump of all licenses used.")
	flag.BoolVar(&csv, "csv", false, "Generate a report of all license usage in CSV format.")
	flag.BoolVar(&mirror, "mirror", false, "Creates a 'licenses' directory with the licenses of all dependencies.")
	flag.StringVar(&config, "config", "", "Path to config file.")
	flag.Parse()

	cfg := newConfig()
	if config != "" {
		var err error
		if cfg, err = readConfig(config); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(1)
		}
	}

	modules, err := getLicenses()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	// now do the real work

	if csv {
		// produce a csv report

		fmt.Printf("Module Name,Module Path,Whitelisted,License Path,License Name,Confidence,Exact Match,Similar To,Similarity Confidence,State\n")
		for _, module := range modules {
			fmt.Printf("%s,%s,%v", module.moduleName, module.path, cfg.whitelistedModules[module.moduleName])
			for _, l := range module.licenses {

				state := "unrecognized"
				if cfg.unrestrictedLicenses[l.analysis.licenseName] {
					state = "unrestricted"
				} else if cfg.reciprocalLicenses[l.analysis.licenseName] {
					state = "reciprocal"
				} else if cfg.restrictedLicenses[l.analysis.licenseName] {
					state = "restricted"
				}

				fmt.Printf(",%s,%s,%s,%v,%s,%s,%s", l.path, l.analysis.licenseName, l.analysis.confidence, l.analysis.exactMatch, l.analysis.similarLicense,
					l.analysis.similarityConfidence, state)
			}
			fmt.Printf("\n")
		}
	} else if mirror {
		var basePath = "licenses"
		_ = os.MkdirAll(basePath, 0755)
		for _, module := range modules {
			p := path.Join(basePath, module.moduleName)
			_ = os.MkdirAll(p, 0755)

			if len(module.licenses) > 0 {
				for _, license := range module.licenses {
					fp := path.Join(p, path.Base(license.path))
					err := ioutil.WriteFile(fp, []byte(license.text), 0644)
					if err != nil {
						_, _ = fmt.Fprintf(os.Stderr, "ERROR: unable to write license file to %s: %v\n", fp, err)
						os.Exit(1)
					}
				}
			} else {
				fp := path.Join(p, "NONE")
				err := ioutil.WriteFile(fp, []byte("NO LICENSE FOUND\n"), 0644)
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "ERROR: unable to write file to %s: %v\n", fp, err)
					os.Exit(1)
				}
			}
		}
	} else {
		var unlicensedModules []*moduleInfo
		var unrecognizedLicenses []*licenseInfo
		var unrestrictedLicenses []*licenseInfo
		var reciprocalLicenses []*licenseInfo
		var restrictedLicenses []*licenseInfo

		// categorize the modules
		for _, module := range modules {
			if !report && !dump {
				// if we're not producing a report, then exclude any module on the whitelist
				if cfg.whitelistedModules[module.moduleName] {
					continue
				}
			}

			if len(module.licenses) == 0 {
				// no license found
				unlicensedModules = append(unlicensedModules, module)
			} else {
				for _, l := range module.licenses {
					if cfg.unrestrictedLicenses[l.analysis.licenseName] {
						unrestrictedLicenses = append(unrestrictedLicenses, l)
					} else if cfg.reciprocalLicenses[l.analysis.licenseName] {
						reciprocalLicenses = append(reciprocalLicenses, l)
					} else if cfg.restrictedLicenses[l.analysis.licenseName] {
						restrictedLicenses = append(restrictedLicenses, l)
					} else {
						unrecognizedLicenses = append(unrecognizedLicenses, l)
					}
				}
			}
		}

		if report {
			fmt.Printf("Modules with unrestricted licenses:\n")
			if len(unrestrictedLicenses) == 0 {
				fmt.Printf("  <none>\n")
			} else {
				for _, l := range unrestrictedLicenses {
					fmt.Printf("  %s: %s, %s confidence\n", l.module.moduleName, l.analysis.licenseName, l.analysis.confidence)
				}
			}
			fmt.Printf("\n")

			fmt.Printf("Modules with reciprocal licenses:\n")
			if len(unrestrictedLicenses) == 0 {
				fmt.Printf("  <none>\n")
			} else {
				for _, l := range reciprocalLicenses {
					fmt.Printf("  %s: %s, %s confidence\n", l.module.moduleName, l.analysis.licenseName, l.analysis.confidence)
				}
			}
			fmt.Printf("\n")

			fmt.Printf("Modules with restricted licenses:\n")
			if len(restrictedLicenses) == 0 {
				fmt.Printf("  <none>\n")
			} else {
				for _, l := range restrictedLicenses {
					fmt.Printf("  %s: %s, %s confidence\n", l.module.moduleName, l.analysis.licenseName, l.analysis.confidence)
				}
			}
			fmt.Printf("\n")

			fmt.Printf("Modules with unrecognized licenses:\n")
			if len(unrecognizedLicenses) == 0 {
				fmt.Printf("  <none>\n")
			} else {
				for _, l := range unrecognizedLicenses {
					if l.analysis.licenseName != "" {
						fmt.Printf("  %s: similar to %s, %s confidence, path '%s'\n", l.module.moduleName, l.analysis.licenseName, l.analysis.confidence, l.path)
					} else if l.analysis.similarLicense != "" {
						fmt.Printf("  %s: similar to %s, %s confidence, path '%s'\n", l.module.moduleName, l.analysis.similarLicense, l.analysis.similarityConfidence, l.path)
					} else {
						fmt.Printf("  %s: path '%s'\n", l.module.moduleName, l.path)
					}
				}
			}
			fmt.Printf("\n")

			fmt.Printf("Modules with no discernible license:\n")
			if len(unlicensedModules) == 0 {
				fmt.Printf("  <none>\n")
			} else {
				for _, m := range unlicensedModules {
					fmt.Printf("  %s\n", m.moduleName)
				}
			}
		} else if dump {
			for _, l := range unrestrictedLicenses {
				fmt.Printf("MODULE: %s\n%s\n", l.module.moduleName, l.text)
			}

			for _, l := range reciprocalLicenses {
				fmt.Printf("MODULE: %s\n%s\n", l.module.moduleName, l.text)
			}

			for _, l := range restrictedLicenses {
				fmt.Printf("MODULE: %s\n%s\n", l.module.moduleName, l.text)
			}

			for _, l := range unrecognizedLicenses {
				fmt.Printf("MODULE: %s\n%s\n", l.module.moduleName, l.text)
			}

			for _, m := range unlicensedModules {
				fmt.Printf("MODULE: %s\n%s\n", m.moduleName, "<none>")
			}
		} else {
			failLint := false

			if len(unrecognizedLicenses) > 0 {
				failLint = true
				fmt.Printf("ERROR: Some modules have unrecognized licenses:\n")
				for _, l := range unrecognizedLicenses {
					if l.analysis.licenseName != "" {
						fmt.Printf("  %s: similar to %s, %s confidence, path '%s'\n", l.module.moduleName, l.analysis.licenseName, l.analysis.confidence, l.path)
					} else if l.analysis.similarLicense != "" {
						fmt.Printf("  %s: similar to %s, %s confidence, path '%s'\n", l.module.moduleName, l.analysis.similarLicense, l.analysis.similarityConfidence, l.path)
					} else {
						fmt.Printf("  %s: path '%s'\n", l.module.moduleName, l.path)
					}
				}
				fmt.Printf("\n")
			}

			if len(unlicensedModules) > 0 {
				failLint = true
				fmt.Printf("ERROR: Some modules have no discernible license:\n")
				for _, m := range unlicensedModules {
					fmt.Printf("  %s\n", m.moduleName)
				}
			}

			if failLint {
				os.Exit(1)
			}
		}
	}
}
