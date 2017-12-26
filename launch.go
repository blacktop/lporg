package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

var (
	// ctx *log.Entry

	// Version stores the plugin's version
	Version string
	// BuildTime stores the plugin's build time
	BuildTime string
)

// App CREATE TABLE apps (item_id INTEGER PRIMARY KEY, title VARCHAR, bundleid VARCHAR, storeid VARCHAR,category_id INTEGER, moddate REAL, bookmark BLOB)
type App struct {
	ItemID     int    `gorm:"column:item_id;primary_key"`
	Title      string `gorm:"column:title"`
	BundleID   string `gorm:"column:bundleid"`
	StoreID    string `gorm:"column:storeid"`
	CategoryID int    `gorm:"column:category_id"`
	Category   Category
	Moddate    float64 `gorm:"column:moddate"`
	Bookmark   []byte  `gorm:"-"`
	// Bookmark   []byte         `gorm:"column:bookmark"`
}

// Category CREATE TABLE categories (rowid INTEGER PRIMARY KEY ASC, uti VARCHAR)
type Category struct {
	ID  uint   `gorm:"column:rowid;primary_key"`
	UTI string `gorm:"column:uti"`
}

// Group CREATE TABLE groups (item_id INTEGER PRIMARY KEY, category_id INTEGER, title VARCHAR)
type Group struct {
	// gorm.Model
	ID         int    `gorm:"column:item_id;primary_key"`
	CategoryID int    `gorm:"column:category_id"`
	Title      string `gorm:"column:title"`
}

// Item CREATE TABLE items (rowid INTEGER PRIMARY KEY ASC, uuid VARCHAR, flags INTEGER, type INTEGER, parent_id INTEGER NOT NULL, ordering INTEGER)
type Item struct {
	RowID int `gorm:"column:rowid;primary_key"`
	App   App
	UUID  string `gorm:"column:uuid"`
	Flags int    `gorm:"column:flags"`
	Type  int    `gorm:"column:type"`
	// ParentID Group         `db:"parent_id"`
	Group    Group `gorm:"ForeignKey:ParentID"`
	ParentID int   `gorm:"not null;column:parent_id"`
	Ordering int   `gorm:"column:ordering"`
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
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
		log.Error(err)
	}
	if err := db.Where("rowid = ?", app.ItemID).First(&item).Error; err != nil {
		log.Error(err)
	}
	if err := db.Where("title = ?", groupName).First(&group).Error; err != nil {
		log.Error(err)
	}
	return updateItemGroup(db, group.ID+1, &item)
}

// CREATE TRIGGER update_item_parent AFTER UPDATE OF parent_id ON items WHEN 0 == (SELECT value FROM dbinfo WHERE key='ignore_items_update_triggers') BEGIN UPDATE dbinfo SET value=1 WHERE key='ignore_items_update_triggers'; UPDATE items SET ordering = (SELECT ifnull(MAX(ordering),0)+1 FROM items WHERE parent_id=new.parent_id AND ROWID!=old.rowid) WHERE ROWID=old.rowid; UPDATE items SET ordering = ordering - 1 WHERE parent_id = old.parent_id and ordering > old.ordering; UPDATE dbinfo SET value=0 WHERE key='ignore_items_update_triggers'; END
func updateItemGroup(db *gorm.DB, groupID int, item *Item) error {
	item.ParentID = groupID
	// item.Ordering = 0
	return db.Save(&item).Error
	// return db.Model(&item).Update("parent_id", groupID).Error
}

// CmdDefaultOrg will organize your launchpad by the app default categories
func CmdDefaultOrg(verbose bool) error {

	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	var items []Item
	var group Group
	appGroups := make(map[string][]string)

	// Older macOS
	// $HOME/Library/Application\ Support/Dock/*.db

	// High Sierra
	// $TMPDIR../0/com.apple.dock.launchpad/db/db

	// find launchpad database
	tmpDir := os.Getenv("TMPDIR")
	launchDB, err := filepath.Glob(tmpDir + "../0/com.apple.dock.launchpad/db/db")
	if err != nil {
		return err
	}
	if len(launchDB) == 0 {
		log.Fatal(errors.New("launchpad DB not found"))
	}
	log.Infoln("Found Launchpad DB: ", launchDB[0])
	// open launchpad database
	db, err := gorm.Open("sqlite3", launchDB[0])
	if err != nil {
		return err
	}
	defer db.Close()

	if verbose {
		db.LogMode(true)
	}

	// grp, err := createNewGroup(db, "Porg")
	// checkError(err)
	// checkError(addAppToGroup(db, "Atom", grp.Title))
	// checkError(addAppToGroup(db, "Brave", grp.Title))
	// checkError(addAppToGroup(db, "iTerm", grp.Title))

	if err := db.Where("type = ?", "4").Find(&items).Error; err != nil {
		log.Error(err)
	}

	for _, item := range items {
		group = Group{}
		db.Model(&item).Related(&item.App)
		db.Model(&item.App).Related(&item.App.Category)
		log.Debugf("App: %s, item.ParentID=%d\n", item.App.Title, item.ParentID-1)
		if err := db.First(&group, item.ParentID-1).Error; err != nil {
			log.WithFields(log.Fields{"ParentID": item.ParentID - 1}).Debug(errors.Wrap(err, "find group with item.ParentID-1 failed"))
			continue
		}
		item.Group = group
		log.Debugf("%+v\n", group)
		if len(group.Title) > 0 {
			appGroups[group.Title] = append(appGroups[group.Title], item.App.Title)
		}
	}

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
	d, err := yaml.Marshal(&appGroups)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("launchpad.yaml", d, 0644)
	return err
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
				return CmdDefaultOrg(c.Bool("verbose"))
			},
		},
	}
	app.Action = func(c *cli.Context) error {

		if c.Bool("verbose") {
			log.SetLevel(log.DebugLevel)
		}

		if c.Args().Present() {
			// user supplied launchpad config YAML
			log.Infoln("IMPLIMENT LOADING FROM CONFIG YAML HERE <=================")
			fmt.Println(porg)
		} else {
			cli.ShowAppHelp(c)
		}
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err.Error())
	}
}
