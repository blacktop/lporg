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
	"os"

	"github.com/apex/log"
	clihander "github.com/apex/log/handlers/cli"
	"github.com/spf13/cobra"
)

var (
	// Verbose boolean flag for verbose logging
	Verbose bool
	// Config stores the path to the config file
	Config string
	// UseICloud boolean flag for using iCloud config
	UseICloud bool
	// AppVersion stores the plugin's version
	AppVersion string
	// AppBuildTime stores the plugin's build time
	AppBuildTime string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "lporg",
	Short: "Organize Your Launchpad",
}

func setLogLevel(verbose bool) int {
	if verbose {
		return int(log.DebugLevel)
	}
	return int(log.WarnLevel)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func init() {
	log.SetHandler(clihander.Default)

	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "V", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&Config, "config", "c", "", "config file (default is $CONFIG/lporg/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&UseICloud, "icloud", false, "use iCloud for config")
	// Settings
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
}
