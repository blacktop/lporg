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

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
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

		fmt.Println(command.PorgASCIIArt)

		backup, _ := cmd.Flags().GetBool("backup")
		yes, _ := cmd.Flags().GetBool("yes")

		conf := &command.Config{
			Cmd:      cmd.Use,
			File:     Config,
			Cloud:    UseICloud,
			Backup:   backup,
			LogLevel: setLogLevel(Verbose),
		}

		if err := conf.Verify(); err != nil {
			return err
		}

		if backup {
			log.Debug("Backing up current launchpad settings")
			if err := command.SaveConfig(conf); err != nil {
				return err
			}
		}

		if !yes {
			prompt := &survey.Confirm{
				Message: "Organize launchpad with default config?",
			}
			if err := survey.AskOne(prompt, &yes); err == terminal.InterruptErr {
				log.Warn("Exiting...")
				return nil
			}
			if !yes {
				return nil
			}
		}

		log.Info("Apply default launchpad settings")
		return command.DefaultOrg(conf)
	},
}

func init() {
	rootCmd.AddCommand(defaultCmd)

	defaultCmd.Flags().BoolP("yes", "y", false, "Do not prompt user for confirmation")
	defaultCmd.Flags().BoolP("backup", "b", false, "Backup current launchpad settings")
	defaultCmd.SetHelpFunc(func(c *cobra.Command, s []string) {
		rootCmd.PersistentFlags().MarkHidden("config")
		rootCmd.PersistentFlags().MarkHidden("icloud")
		c.Parent().HelpFunc()(c, s)
	})
}
