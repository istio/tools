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

// TODO:
//  * Optionally use "git mv" to ensure git tracks the renamed file
//  * Move images next to the markdown file to the new location, updating any links accordingly
//  * Support moving a whole directory to a new location

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := cobra.Command{
		Use:   "mvpage <old markdown file> <new markdown file>",
		Short: "Moves a content page within a Hugo site.",
		Long: "Moves a content page within a Hugo site, updating all links to the page\n" +
			"and adding an alias so that old bookmarks to the page continue working",
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("expecting two arguments")
			}
			cmd.SilenceUsage = true

			srcFile := args[0]
			dstFile := args[1]

			cfg, err := readHugoConfig()
			if err != nil {
				return fmt.Errorf("unable to read config: %v", err)
			}

			mover := newMover(cfg)

			if err = mover.move(srcFile, dstFile); err != nil {
				return fmt.Errorf("unable to move: %v", err)
			}

			return nil
		},
	}

	flag.CommandLine.VisitAll(func(gf *flag.Flag) {
		rootCmd.PersistentFlags().AddGoFlag(gf)
	})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
