/*
Copyright © 2023 blacktop

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"

	"github.com/apex/log"
	"github.com/blacktop/lporg/internal/command"
	"github.com/spf13/cobra"
)

// defaultCmd represents the default command
var defaultCmd = &cobra.Command{
	Use:   "default",
	Short: "Organize by default Apple app categories",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		if Verbose {
			log.SetLevel(log.DebugLevel)
		}

		fmt.Println(command.PorgAsciiArt)

		backup, _ := cmd.Flags().GetBool("backup")

		if backup {
			log.Debug("Backing up current launchpad settings")
			if err := command.SaveConfig(&command.Config{
				File:     Config,
				Cloud:    UseICloud,
				Backup:   true,
				LogLevel: setLogLevel(Verbose),
			}); err != nil {
				return err
			}
		}

		log.Info("Apply default launchpad settings")
		return command.DefaultOrg(&command.Config{
			File:     Config,
			Cloud:    UseICloud,
			LogLevel: setLogLevel(Verbose),
		})
	},
}

func init() {
	rootCmd.AddCommand(defaultCmd)

	defaultCmd.Flags().BoolP("backup", "b", false, "Backup current launchpad settings")
}
