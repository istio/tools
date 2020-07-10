// Copyright 2018 Istio Authors. All Rights Reserved.
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

package checker

import (
	"log"
	"path/filepath"
)

// Allowlist determines if rules are allowlisted for the given paths.
type Allowlist struct {
	// Map from path to allowlisted rules.
	ruleAllowlist map[string][]string
}

// NewAllowlist creates and returns an Allowlist object.
func NewAllowlist(ruleAllowlist map[string][]string) *Allowlist {
	return &Allowlist{ruleAllowlist: ruleAllowlist}
}

// Apply returns true if the given rule is allowlisted for the given path.
func (wl *Allowlist) Apply(path string, rule Rule) bool {
	for _, skipRule := range wl.getAllowlistedRules(path) {
		if skipRule == rule.GetID() {
			return true
		}
	}
	return false
}

// getAllowlistedRules returns the allowlisted rule given the path
func (wl *Allowlist) getAllowlistedRules(path string) []string {
	// Check whether path is allowlisted
	for wp, allowlistedRules := range wl.ruleAllowlist {
		// filepath.Match is needed for canonical matching
		matched, err := filepath.Match(wp, path)
		if err != nil {
			log.Printf("file match returns error: %v", err)
		}
		if matched {
			return allowlistedRules
		}
	}
	return []string{}
}
