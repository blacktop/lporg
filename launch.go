package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/apex/log"
	clihander "github.com/apex/log/handlers/cli"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

var (
	// Version stores the plugin's version
	Version string
	// BuildTime stores the plugin's build time
	BuildTime string
	// for log output
	bold = "\033[1m%s\033[0m"
	// lpad is the main object
	lpad LaunchPad
)

// LaunchPad is a LaunchPad struct
type LaunchPad struct {
	DB     *gorm.DB
	File   string
	Folder string
}

func checkError(err error) {
	if err != nil {
		log.WithError(err).Fatal("failed")
	}
}

// RunCommand runs cmd on file
func RunCommand(ctx context.Context, cmd string, args ...string) (string, error) {

	var c *exec.Cmd

	if ctx != nil {
		c = exec.CommandContext(ctx, cmd, args...)
	} else {
		c = exec.Command(cmd, args...)
	}

	output, err := c.Output()
	if err != nil {
		return string(output), err
	}

	// check for exec context timeout
	if ctx != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("command %s timed out", cmd)
		}
	}

	return string(output), nil
}

func restartDock() error {
	ctx := context.Background()

	log.Info("restarting Dock")
	if _, err := RunCommand(ctx, "killall", "Dock"); err != nil {
		return errors.Wrap(err, "killing Dock process failed")
	}

	// let system settle
	time.Sleep(5 * time.Second)

	return nil
}

func removeOldDatabaseFiles(dbpath string) error {

	paths := []string{
		filepath.Join(dbpath, "db"),
		filepath.Join(dbpath, "db-shm"),
		filepath.Join(dbpath, "db-wal"),
	}

	for _, path := range paths {
		if err := os.Remove(path); err != nil {
			return errors.Wrap(err, "removing file failed")
		}
		log.WithField("path", path).Info("removed old file")
	}

	return restartDock()
}

func createNewGroup(db *gorm.DB, title string) (Group, error) {

	group := Group{Title: title}

	success := db.NewRecord(group) // => returns `true` as primary key is blank
	log.WithFields(log.Fields{"success": success}).Debug("create new group record")

	err := db.Create(&group).Error
	if err != nil {
		return group, errors.Wrap(err, "create new entry failed")
	}

	emptyGroup := Group{}
	success = db.NewRecord(emptyGroup)
	log.WithFields(log.Fields{"success": success}).Debug("create new empty group record")

	err = db.Create(&emptyGroup).Error
	if err != nil {
		return group, errors.Wrap(err, "create new EMPTY entry failed")
	}

	return group, err
}

func addAppToGroup(db *gorm.DB, appName, groupName string) error {

	var (
		app   App
		item  Item
		group Group
	)

	if err := db.Where("title = ?", appName).First(&app).Error; err != nil {
		log.WithError(err).Error("find app failed")
	}
	if err := db.Where("rowid = ?", app.ItemID).First(&item).Error; err != nil {
		log.WithError(err).Error("find item failed")
	}
	if err := db.Where("title = ?", groupName).First(&group).Error; err != nil {
		log.WithError(err).Error("find group failed")
	}
	return updateItemGroup(db, group.ID+1, &item)
}

// CREATE TRIGGER update_item_parent AFTER UPDATE OF parent_id ON items WHEN 0 == (SELECT value FROM dbinfo WHERE key='ignore_items_update_triggers') BEGIN UPDATE dbinfo SET value=1 WHERE key='ignore_items_update_triggers'; UPDATE items SET ordering = (SELECT ifnull(MAX(ordering),0)+1 FROM items WHERE parent_id=new.parent_id AND ROWID!=old.rowid) WHERE ROWID=old.rowid; UPDATE items SET ordering = ordering - 1 WHERE parent_id = old.parent_id and ordering > old.ordering; UPDATE dbinfo SET value=0 WHERE key='ignore_items_update_triggers'; END
func updateItemGroup(db *gorm.DB, groupID int, item *Item) error {
	var dbinfo DBInfo

	if err := db.Where("key = ?", "ignore_items_update_triggers").First(&dbinfo).Error; err != nil {
		log.WithError(err).Error("find dbinfo failed")
	}
	err := db.Model(&item).Update("key", "1").Error
	if err != nil {
		log.WithError(err).Error("counld not update ignore_items_update_triggers to 1")
	}

	// item.ParentID = groupID
	// item.Ordering = 0
	// return db.Save(&item).Error
	err = db.Model(&item).Update("parent_id", groupID).Error
	if err != nil {
		log.WithError(err).Error("counld not update item's group")
	}

	err = db.Model(&item).Update("key", "0").Error
	if err != nil {
		log.WithError(err).Error("counld not update ignore_items_update_triggers to 0")
	}

	return nil
}

// CmdDefaultOrg will organize your launchpad by the app default categories
func CmdDefaultOrg(verbose bool) error {

	log.Infof(bold, "[PARSE LAUCHPAD DATABASE]")
	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	// var items []Item
	// var group Group
	// appGroups := make(map[string][]string)

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
		log.WithError(err).WithField("path", lpad.File).Fatal("launchpad DB not found")
	}
	log.WithFields(log.Fields{"database": lpad.File}).Info("found launchpad database")

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

	if err := lpad.DisableTriggers(); err != nil {
		log.WithError(err).Error("DisableTriggers failed")
	}
	// Clear all items related to groups so we can re-create them
	if err := lpad.ClearGroups(); err != nil {
		log.WithError(err).Error("ClearGroups failed")
	}
	// Add root and holding pages to items and groups
	if err := lpad.AddRootsAndHoldingPages(); err != nil {
		log.WithError(err).Error("AddRootsAndHoldingPagesfailed")
	}

	groupID := math.Max(float64(lpad.getMaxAppID()), float64(lpad.getMaxWidgetID()))

	var pages map[string][]map[string][]string

	data, err := ioutil.ReadFile("launchpad.yaml")
	if err != nil {
		log.WithError(err).WithField("path", lpad.File).Fatal("launchpad.yaml not found")
		return err
	}

	err = yaml.Unmarshal(data, &pages)
	if err != nil {
		log.WithError(err).WithField("path", lpad.File).Fatal("unmarshalling yaml failed")
		return err
	}

	// Create App Folders
	if err := lpad.CreateAppFolders(pages, int(groupID)); err != nil {
		log.WithError(err).Error("CreateAppFolders")
	}

	// grp, err := createNewGroup(db, "Porg")
	// checkError(err)
	// checkError(addAppToGroup(db, "Atom", grp.Title))
	// checkError(addAppToGroup(db, "Brave", grp.Title))
	// checkError(addAppToGroup(db, "iTerm", grp.Title))

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

	log.Infof(bold, "successfully wrote launchpad.yaml")
	lpad.EnableTriggers()
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
				return CmdDefaultOrg(c.GlobalBool("verbose"))
			},
		},
		{
			Name:  "save",
			Usage: "Save Current Launchpad App Config",
			Action: func(c *cli.Context) error {
				log.Info("IMPLIMENT SAVING TO CONFIG YAML HERE <=================")
				return nil
			},
		},
		{
			Name:  "load",
			Usage: "Load Launchpad App Config From File",
			Action: func(c *cli.Context) error {
				if c.Args().Present() {
					// user supplied launchpad config YAML
					log.Info("IMPLIMENT LOADING FROM CONFIG YAML HERE <=================")
					fmt.Println(porg)
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
