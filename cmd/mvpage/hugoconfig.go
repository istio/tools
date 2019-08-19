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
	"path/filepath"

	"github.com/spf13/viper"
)

type hugoConfig struct {
	defaultContentLanguage string
	contentRootPerLanguage map[string]string
}

// Reads interesting bits of state from a Hugo config file
func readHugoConfig() (hugoConfig, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return hugoConfig{}, fmt.Errorf("unable to read Hugo config file: %v", err)
	}

	contentRoots := make(map[string]string)

	languages := viper.GetStringMapString("languages")
	for lang := range languages {
		dir := viper.GetString("languages." + lang + ".contentDir")

		absDir, err := filepath.Abs(dir)
		if err != nil {
			return hugoConfig{}, fmt.Errorf("unable to find Hugo content directory '%s': %v", dir, err)
		}
		contentRoots[lang] = absDir
	}

	return hugoConfig{
		contentRootPerLanguage: contentRoots,
		defaultContentLanguage: viper.GetString("defaultContentLanguage"),
	}, nil
}
