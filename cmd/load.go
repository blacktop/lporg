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
	"github.com/apex/log"
	"github.com/blacktop/lporg/internal/command"
	"github.com/spf13/cobra"
)

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "Load launchpad settings config from `FILE`",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		if Verbose {
			log.SetLevel(log.DebugLevel)
		}

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

		log.Info("Loading launchpad settings")
		return command.LoadConfig(&command.Config{
			File:     Config,
			Cloud:    UseICloud,
			LogLevel: setLogLevel(Verbose),
		})
	},
}

func init() {
	rootCmd.AddCommand(loadCmd)

	loadCmd.Flags().BoolP("backup", "b", false, "Backup current launchpad settings")
}
