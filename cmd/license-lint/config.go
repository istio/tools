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
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
)

type rawConfig struct {
	// definitely ok to use and modify.
	UnrestrictedLicenses []string `json:"unrestricted_licenses"`

	// can be used but not modified
	ReciprocalLicenses []string `json:"reciprocal_licenses"`

	// cannot be used
	RestrictedLicenses []string `json:"restricted_licenses"`

	// modules that get completely ignored during analysis
	AllowlistedModules []string `json:"allowlisted_modules"`
}

type config struct {
	// definitely ok to use and modify.
	unrestrictedLicenses map[string]bool

	// can be used but not modified
	reciprocalLicenses map[string]bool

	// cannot be used
	restrictedLicenses map[string]bool

	// modules that get completely ignored during analysis
	allowlistedModules map[string]bool
}

func newConfig() config {
	return config{
		unrestrictedLicenses: make(map[string]bool),
		reciprocalLicenses:   make(map[string]bool),
		restrictedLicenses:   make(map[string]bool),
		allowlistedModules:   make(map[string]bool),
	}
}

func readConfig(path string) (config, error) {
	var b []byte
	var err error
	if b, err = ioutil.ReadFile(path); err != nil {
		return config{}, fmt.Errorf("unable to read configuration file %s: %v", path, err)
	}

	var rc rawConfig
	if err = yaml.Unmarshal(b, &rc); err != nil {
		return config{}, fmt.Errorf("unable to parse configuration file %s: %v", path, err)
	}

	c := newConfig()

	for _, s := range rc.UnrestrictedLicenses {
		c.unrestrictedLicenses[s] = true
	}

	for _, s := range rc.ReciprocalLicenses {
		c.reciprocalLicenses[s] = true
	}

	for _, s := range rc.RestrictedLicenses {
		c.restrictedLicenses[s] = true
	}

	for _, s := range rc.AllowlistedModules {
		c.allowlistedModules[s] = true
	}

	return c, nil
}
