package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/apex/log"
	clihander "github.com/apex/log/handlers/cli"
	"github.com/blacktop/lporg/database"
	"github.com/blacktop/lporg/database/utils"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/urfave/cli"
)

var (
	// Version stores the plugin's version
	Version string
	// BuildTime stores the plugin's build time
	BuildTime string
	// for log output
	bold = "\033[1m%s\033[0m"
	// lpad is the main object
	lpad database.LaunchPad
)

// CmdDefaultOrg will organize your launchpad by the app default categories
func CmdDefaultOrg(verbose bool) error {
	log.Info("IMPLIMENT DEFAULT ORG HERE <=================")
	return nil
}

// CmdSaveConfig will save your launchpad settings to a config file
func CmdSaveConfig(verbose bool) error {
	log.Info("IMPLIMENT SAVING TO CONFIG YAML HERE <=================")
	// var items []Item
	// var group Group
	// appGroups := make(map[string][]string)

	// if err := db.Where("type = ?", "4").Find(&items).Error; err != nil {
	// 	log.WithError(err).Error("find item of type=4 failed")
	// }

	// for _, item := range items {
	// 	group = Group{}
	// 	db.Model(&item).Related(&item.App)
	// 	db.Model(&item.App).Related(&item.App.Category)
	// 	log.WithFields(log.Fields{
	// 		"app_id":    item.App.ItemID,
	// 		"app_name":  item.App.Title,
	// 		"parent_id": item.ParentID - 1,
	// 	}).Debug("parsing item")
	// 	if err := db.First(&group, item.ParentID-1).Error; err != nil {
	// 		log.WithError(err).WithFields(log.Fields{"ParentID": item.ParentID - 1}).Debug("find group failed")
	// 		continue
	// 	}
	// 	log.WithFields(log.Fields{
	// 		"group_id":   group.ID,
	// 		"group_name": group.Title,
	// 	}).Debug("parsing group")
	// 	item.Group = group
	// 	if len(group.Title) > 0 {
	// 		appGroups[group.Title] = append(appGroups[group.Title], item.App.Title)
	// 	}
	// }

	// fmt.Println("--------------------------------------------------------")

	// fmt.Printf("item========================================> %+v\n", item)
	// itemJSON, _ := json.Marshal(item)
	// fmt.Println(string(itemJSON))
	// fmt.Printf("%+v\n", item.App)

	// write out to TOML
	// checkError(writeTomlFile("./launchpad.toml", item))
	// if err := db.Find(&groups).Error; err != nil {
	// 	log.Error(err)
	// }
	// g := make(map[string][]Group)
	// g["Groups"] = groups

	////////////////////////////////////////////////// TODO: write config out to YAML
	// d, err := yaml.Marshal(&appGroups)
	// if err != nil {
	// 	return errors.Wrap(err, "unable to marshall YAML")
	// }

	// if err = ioutil.WriteFile("launchpad.yaml", d, 0644); err != nil {
	// 	return err
	// }

	log.Infof(bold, strings.ToUpper("successfully wrote launchpad.yaml"))

	return nil
}

// CmdLoadConfig will load your launchpad settings from a config file
func CmdLoadConfig(verbose bool, configFile string) error {

	log.Infof(bold, "[PARSE LAUCHPAD DATABASE]")

	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	// Older macOS ////////////////////////////////
	// $HOME/Library/Application\ Support/Dock/*.db

	// High Sierra //////////////////////////////
	// $TMPDIR../0/com.apple.dock.launchpad/db/db

	// find launchpad database
	tmpDir := os.Getenv("TMPDIR")
	lpad.Folder = filepath.Join(tmpDir, "../0/com.apple.dock.launchpad/db")
	lpad.File = filepath.Join(lpad.Folder, "db")
	// launchpadDB = "./launchpad.db"
	if _, err := os.Stat(lpad.File); os.IsNotExist(err) {
		utils.Indent(log.WithError(err).WithField("path", lpad.File).Fatal)("launchpad DB not found")
	}
	utils.Indent(log.WithFields(log.Fields{"database": lpad.File}).Info)("found launchpad database")

	// start from a clean slate
	err := removeOldDatabaseFiles(lpad.Folder)
	if err != nil {
		return err
	}

	// open launchpad database
	db, err := gorm.Open("sqlite3", lpad.File)
	if err != nil {
		return err
	}
	defer db.Close()

	lpad.DB = db

	if verbose {
		db.LogMode(true)
	}

	// Disable the update triggers
	if err := lpad.DisableTriggers(); err != nil {
		log.WithError(err).Fatal("DisableTriggers failed")
	}
	// Clear all items related to groups so we can re-create them
	if err := lpad.ClearGroups(); err != nil {
		log.WithError(err).Fatal("ClearGroups failed")
	}
	// Add root and holding pages to items and groups
	if err := lpad.AddRootsAndHoldingPages(); err != nil {
		log.WithError(err).Fatal("AddRootsAndHoldingPagesfailed")
	}

	// We will begin our group records using the max ids found (groups always appear after apps and widgets)
	groupID := math.Max(float64(lpad.GetMaxAppID()), float64(lpad.GetMaxWidgetID()))

	// Read in Config file
	config, err := database.LoadConfig(configFile)
	if err != nil {
		log.WithError(err).Fatal("database.LoadConfig")
	}

	// Create App Folders
	if err := lpad.CreateAppFolders(config, int(groupID)); err != nil {
		log.WithError(err).Fatal("CreateAppFolders")
	}
	// Re-enable the update triggers
	if err := lpad.EnableTriggers(); err != nil {
		log.WithError(err).Fatal("EnableTriggers failed")
	}

	return restartDock()
}

func init() {
	log.SetHandler(clihander.Default)
}

var appHelpTemplate = `Usage: {{.Name}} {{if .Flags}}[OPTIONS] {{end}}COMMAND [arg...]

{{.Usage}}

Version: {{.Version}}{{if or .Author .Email}}
Author:{{if .Author}} {{.Author}}{{if .Email}} - <{{.Email}}>{{end}}{{else}}
  {{.Email}}{{end}}{{end}}
{{if .Flags}}
Options:
  {{range .Flags}}{{.}}
  {{end}}{{end}}
Commands:
  {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
  {{end}}
Run '{{.Name}} COMMAND --help' for more information on a command.
`

func main() {

	cli.AppHelpTemplate = appHelpTemplate
	app := cli.NewApp()

	app.Name = "lporg"
	app.Author = "blacktop"
	app.Email = "https://github.com/blacktop"
	app.Version = Version + ", BuildTime: " + BuildTime
	app.Compiled, _ = time.Parse("20060102", BuildTime)
	app.Usage = "Organize Your Launchpad"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, V",
			Usage: "verbose output",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "default",
			Usage: "Organize by Categories",
			Action: func(c *cli.Context) error {
				fmt.Println(porg)
				return CmdDefaultOrg(c.GlobalBool("verbose"))
			},
		},
		{
			Name:  "save",
			Usage: "Save Current Launchpad Settings Config",
			Action: func(c *cli.Context) error {
				return CmdSaveConfig(c.GlobalBool("verbose"))
			},
		},
		{
			Name:  "load",
			Usage: "Load Launchpad Settings Config From File",
			Action: func(c *cli.Context) error {
				if c.Args().Present() {
					// user supplied launchpad config YAML
					err := CmdLoadConfig(c.Bool("verbose"), c.Args().First())
					if err != nil {
						return err
					}
				} else {
					cli.ShowAppHelp(c)
				}
				return nil
			},
		},
	}
	app.Action = func(c *cli.Context) error {
		if !c.Args().Present() {
			cli.ShowAppHelp(c)
		}
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.WithError(err).Fatal("failed")
	}
}
