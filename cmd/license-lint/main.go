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
	"os"
)

func main() {
	var report bool
	var csv bool
	var config string
	flag.BoolVar(&report, "report", false, "Generate a report of all license usage.")
	flag.BoolVar(&csv, "csv", false, "Generate a report of all license usage in CSV format.")
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

	modules, _ := getLicenses()

	var noLicense []*moduleInfo
	var unknownLicenses []*licenseInfo
	var unrestrictedLicenses []*licenseInfo
	var reciprocalLicenses []*licenseInfo
	var restrictedLicenses []*licenseInfo

	// categorize the modules
	for _, module := range modules {
		// if we're not producing a report, then exclude any module on the whitelist
		if !report && !csv {
			if cfg.whitelistedModules[module.moduleName] {
				continue
			}
		}

		if len(module.licenses) == 0 {
			// no license found
			noLicense = append(noLicense, module)
		} else {
			for _, l := range module.licenses {
				if cfg.unrestrictedLicenses[l.analysis.licenseName] {
					unrestrictedLicenses = append(unrestrictedLicenses, l)
				} else if cfg.reciprocalLicenses[l.analysis.licenseName] {
					reciprocalLicenses = append(reciprocalLicenses, l)
				} else if cfg.restrictedLicenses[l.analysis.licenseName] {
					restrictedLicenses = append(restrictedLicenses, l)
				} else {
					unknownLicenses = append(unknownLicenses, l)
				}
			}
		}
	}

	// now do the real work

	if csv {
		// produce a csv report

		fmt.Printf("Module Name,Module Path,Whitelisted,License Path,License Name,Confidence,Exact Match,State\n")
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

				fmt.Printf(",%s,%s,%s,%v,%s", l.path, l.analysis.licenseName, l.analysis.confidence, l.analysis.exactMatch, state)
			}
			fmt.Printf("\n")
		}
	} else if report {
		// produce a human-friendly report

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
		if len(unknownLicenses) == 0 {
			fmt.Printf("  <none>\n")
		} else {
			for _, l := range unknownLicenses {
				fmt.Printf("  %s: %s, %s confidence\n", l.module.moduleName, l.analysis.licenseName, l.analysis.confidence)
			}
		}
		fmt.Printf("\n")

		fmt.Printf("Modules with no discernible license:\n")
		if len(noLicense) == 0 {
			fmt.Printf("  <none>\n")
		} else {
			for _, m := range noLicense {
				fmt.Printf("  %s\n", m.moduleName)
			}
		}

	} else {
		// lint

		fail := false
		if len(restrictedLicenses) > 0 {
			_, _ = fmt.Fprint(os.Stderr, "Some modules use restricted licenses:\n")
			for _, rl := range restrictedLicenses {
				_, _ = fmt.Fprintf(os.Stderr, "  Module %s, license %s\n", rl.module.moduleName, rl.analysis.licenseName)
			}
			fail = true
		}

		if len(unknownLicenses) > 0 {
			_, _ = fmt.Fprint(os.Stderr, "Some modules have unrecognized licenses:\n")
			for _, rl := range unknownLicenses {
				_, _ = fmt.Fprintf(os.Stderr, "  Module %s\n", rl.module.moduleName)
			}
			fail = true
		}

		if len(noLicense) > 0 {
			_, _ = fmt.Fprint(os.Stderr, "Some modules have no discernible license:\n")
			for _, m := range noLicense {
				_, _ = fmt.Fprintf(os.Stderr, "  Module %s\n", m.moduleName)
			}
			fail = true
		}

		if fail {
			os.Exit(1)
		}
	}
}
