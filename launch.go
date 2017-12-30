package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/apex/log"
	clihander "github.com/apex/log/handlers/cli"
	"github.com/blacktop/lporg/database"
	"github.com/blacktop/lporg/database/utils"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	survey "gopkg.in/AlecAivazis/survey.v1"
	yaml "gopkg.in/yaml.v2"
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

// add missing apps to pages at 30 apps per page
func parseMissing(missing []string, pages []database.Page) []database.Page {
	if len(missing) > 0 {
		for _, chunk := range split(missing, 30) {
			p := database.Page{
				Number: len(pages) + 1,
			}

			// because you can't assign a []string to an []interface{} we must copy in one at a time
			chunkInterface := make([]interface{}, len(chunk))
			for i, v := range chunk {
				chunkInterface[i] = v
			}

			p.Items = chunkInterface
			pages = append(pages, p)
			for _, smallerChunk := range split(chunk, 5) {
				msg := fmt.Sprintf("adding missing apps to page=%d", p.Number)
				utils.DoubleIndent(log.WithField("apps", smallerChunk).Warn)(msg)
			}
		}
	}

	return pages
}

func parsePages(root int, parentMapping map[int][]database.Item) (database.Apps, error) {
	var apps database.Apps

	for pageNum, page := range parentMapping[root] {

		log.Infof("page number: %d", pageNum+1)

		p := database.Page{Number: pageNum + 1}

		for _, item := range parentMapping[page.ID] {
			switch item.Type {
			case database.ApplicationType:
				utils.Indent(log.WithField("title", item.App.Title).Info)("found app")
				p.Items = append(p.Items, item.App.Title)
			case database.WidgetType:
				utils.Indent(log.WithField("title", item.Widget.Title).Info)("found widget")
				p.Items = append(p.Items, item.Widget.Title)
			case database.FolderRootType:

				utils.Indent(log.WithField("title", item.Group.Title).Info)("found folder")

				f := database.AppFolder{Name: item.Group.Title}

				if len(parentMapping[item.ID]) < 1 {
					return database.Apps{}, errors.New("did not find folder page item in page")
				}

				for fpIndex, fpage := range parentMapping[item.ID] {
					utils.DoubleIndent(log.WithField("number", fpIndex+1).Info)("found folder page")

					fp := database.FolderPage{Number: fpIndex + 1}

					for _, folder := range parentMapping[fpage.ID] {
						utils.TripleIndent(log.WithField("title", folder.App.Title).Info)("found app")
						fp.Items = append(fp.Items, folder.App.Title)
					}

					f.Pages = append(f.Pages, fp)
				}

				if len(f.Pages) > 0 && len(f.Pages[0].Items) > 0 {
					p.Items = append(p.Items, f)
				} else {
					utils.DoubleIndent(log.WithField("folder", item.Group.Title).Error)("empty folder")
				}

			case database.PageType:
				utils.Indent(log.WithField("parent_id", item.ParentID).Info)("found page")
			default:
				utils.Indent(log.WithField("type", item.Type).Error)("found ?")
			}
		}
		apps.Pages = append(apps.Pages, p)
	}
	return apps, nil
}

// CmdDefaultOrg will organize your launchpad by the app default categories
func CmdDefaultOrg(verbose bool) error {

	log.Infof(bold, "USING DEFAULT LAUNCHPAD ORGANIZATION")

	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	// find launchpad database
	tmpDir := os.Getenv("TMPDIR")
	lpad.Folder = filepath.Join(tmpDir, "../0/com.apple.dock.launchpad/db")
	lpad.File = filepath.Join(lpad.Folder, "db")
	// lpad.File = "./launchpad.db"
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

	// Clear all items related to groups so we can re-create them
	if err := lpad.ClearGroups(); err != nil {
		log.WithError(err).Fatal("ClearGroups failed")
	}

	// Disable the update triggers
	if err := lpad.DisableTriggers(); err != nil {
		log.WithError(err).Fatal("DisableTriggers failed")
	}

	// Add root and holding pages to items and groups
	if err := lpad.AddRootsAndHoldingPages(); err != nil {
		log.WithError(err).Fatal("AddRootsAndHoldingPagesfailed")
	}

	// We will begin our group records using the max ids found (groups always appear after apps and widgets)
	groupID := int(math.Max(float64(lpad.GetMaxAppID()), float64(lpad.GetMaxWidgetID())))

	utils.Indent(log.Info)("creating folders out of app categories")

	// Create default config file
	var apps []database.App
	var categories []database.Category
	var conf database.Config

	page := database.Page{Number: 1}

	if err := db.Find(&categories).Error; err != nil {
		log.WithError(err).Error("categories query failed")
	}

	for _, category := range categories {
		folderName := strings.Title(strings.Replace(strings.TrimPrefix(category.UTI, "public.app-category."), "-", " ", 1))
		folder := database.AppFolder{Name: folderName}
		folderPage := database.FolderPage{Number: 1}
		utils.DoubleIndent(log.WithField("folder", folderName).Info)("adding folder")
		if err := db.Where("category_id = ?", category.ID).Find(&apps).Error; err != nil {
			log.WithError(err).Error("categories query failed")
		}
		for _, app := range apps {
			utils.TripleIndent(log.WithField("app", app.Title).Info)("adding app to category folder")
			folderPage.Items = utils.AppendIfMissing(folderPage.Items, app.Title)
		}
		folder.Pages = append(folder.Pages, folderPage)
		page.Items = append(page.Items, folder)
	}

	conf.Apps.Pages = append(conf.Apps.Pages, page)

	////////////////////////////////////////////////////////////////////
	// Place Widgets ///////////////////////////////////////////////////
	utils.Indent(log.Info)("creating Widget folders and adding widgets to them")
	missing, err := lpad.GetMissing(conf.Widgets, database.WidgetType)
	if err != nil {
		log.WithError(err).Fatal("Default GetMissing=>Widgets")
	}

	conf.Widgets.Pages = parseMissing(missing, conf.Widgets.Pages)
	groupID, err = lpad.ApplyConfig(conf.Widgets, database.WidgetType, groupID, 3)
	if err != nil {
		log.WithError(err).Fatal("Default ApplyConfig=>Widgets")
	}

	/////////////////////////////////////////////////////////////////////
	// Place Apps ///////////////////////////////////////////////////////
	utils.Indent(log.Info)("creating App folders and adding apps to them")
	missing, err = lpad.GetMissing(conf.Apps, database.ApplicationType)
	if err != nil {
		log.WithError(err).Fatal("Default GetMissing=>Apps")
	}

	conf.Apps.Pages = parseMissing(missing, conf.Apps.Pages)
	groupID, err = lpad.ApplyConfig(conf.Apps, database.ApplicationType, groupID, 1)
	if err != nil {
		log.WithError(err).Fatal("Default ApplyConfig==>Apps")
	}

	// Re-enable the update triggers
	if err := lpad.EnableTriggers(); err != nil {
		log.WithError(err).Fatal("EnableTriggers failed")
	}

	return restartDock()
}

// CmdSaveConfig will save your launchpad settings to a config file
func CmdSaveConfig(verbose bool, configFile string) error {

	log.Infof(bold, "SAVING LAUNCHPAD DATABASE")

	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	var (
		launchpadRoot int
		dashboardRoot int
		items         []database.Item
		dbinfo        []database.DBInfo
		conf          database.Config
	)

	// find launchpad database
	tmpDir := os.Getenv("TMPDIR")
	lpad.Folder = filepath.Join(tmpDir, "../0/com.apple.dock.launchpad/db")
	lpad.File = filepath.Join(lpad.Folder, "db")
	// lpad.File = "./launchpad.db"
	if _, err := os.Stat(lpad.File); os.IsNotExist(err) {
		utils.Indent(log.WithError(err).WithField("path", lpad.File).Fatal)("launchpad DB not found")
	}
	utils.Indent(log.WithFields(log.Fields{"database": lpad.File}).Info)("found launchpad database")

	// open launchpad database
	db, err := gorm.Open("sqlite3", lpad.File)
	if err != nil {
		return err
	}
	defer db.Close()

	if verbose {
		db.LogMode(true)
	}

	// get launchpad and dashboard roots
	if err := db.Where("key in (?)", []string{"launchpad_root", "dashboard_root"}).Find(&dbinfo).Error; err != nil {
		log.WithError(err).Error("dbinfo query failed")
	}
	for _, info := range dbinfo {
		switch info.Key {
		case "launchpad_root":
			launchpadRoot, _ = strconv.Atoi(info.Value)
		case "dashboard_root":
			dashboardRoot, _ = strconv.Atoi(info.Value)
		default:
			log.WithField("key", info.Key).Error("bad key")
		}
	}

	// get all the relavent items
	if err := db.Not("uuid in (?)", []string{"ROOTPAGE", "HOLDINGPAGE", "ROOTPAGE_DB", "HOLDINGPAGE_DB", "ROOTPAGE_VERS", "HOLDINGPAGE_VERS"}).
		Order("items.parent_id, items.ordering").
		Find(&items).Error; err != nil {
		log.WithError(err).Error("items query failed")
	}

	// create parent mapping object
	log.Info("collecting launchpad/dashboard pages")
	parentMapping := make(map[int][]database.Item)
	for _, item := range items {
		db.Model(&item).Related(&item.App)
		db.Model(&item).Related(&item.Widget)
		db.Model(&item).Related(&item.Group)

		parentMapping[item.ParentID] = append(parentMapping[item.ParentID], item)
	}

	log.Info("interating over launchpad pages")
	conf.Apps, err = parsePages(launchpadRoot, parentMapping)
	if err != nil {
		return errors.Wrap(err, "unable to parse launchpad pages")
	}

	log.Info("interating over dashboard pages")
	conf.Widgets, err = parsePages(dashboardRoot, parentMapping)
	if err != nil {
		return errors.Wrap(err, "unable to parse dashboard pages")
	}

	// write out config YAML file
	d, err := yaml.Marshal(&conf)
	if err != nil {
		return errors.Wrap(err, "unable to marshall YAML")
	}

	if len(configFile) == 0 {
		configFile = "launchpad.yaml"
	}
	if err = ioutil.WriteFile(configFile, d, 0644); err != nil {
		return errors.Wrap(err, "unable to write YAML")
	}

	log.Infof(bold, "successfully wrote launchpad.yaml")

	return nil
}

// CmdLoadConfig will load your launchpad settings from a config file
func CmdLoadConfig(verbose bool, configFile string) error {

	log.Infof(bold, "PARSE LAUCHPAD DATABASE")

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
	// lpad.File = "./launchpad-test.db"
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

	// Clear all items related to groups so we can re-create them
	if err := lpad.ClearGroups(); err != nil {
		log.WithError(err).Fatal("ClearGroups failed")
	}

	// Disable the update triggers
	if err := lpad.DisableTriggers(); err != nil {
		log.WithError(err).Fatal("DisableTriggers failed")
	}

	// Add root and holding pages to items and groups
	if err := lpad.AddRootsAndHoldingPages(); err != nil {
		log.WithError(err).Fatal("AddRootsAndHoldingPagesfailed")
	}

	// We will begin our group records using the max ids found (groups always appear after apps and widgets)
	groupID := int(math.Max(float64(lpad.GetMaxAppID()), float64(lpad.GetMaxWidgetID())))

	// Read in Config file
	config, err := database.LoadConfig(configFile)
	if err != nil {
		log.WithError(err).Fatal("database.LoadConfig")
	}

	////////////////////////////////////////////////////////////////////
	// Place Widgets ///////////////////////////////////////////////////
	utils.Indent(log.Info)("creating Widget folders and adding widgets to them")
	missing, err := lpad.GetMissing(config.Widgets, database.WidgetType)
	if err != nil {
		log.WithError(err).Fatal("GetMissing=>Widgets")
	}

	config.Widgets.Pages = parseMissing(missing, config.Widgets.Pages)
	groupID, err = lpad.ApplyConfig(config.Widgets, database.WidgetType, groupID, 3)
	if err != nil {
		log.WithError(err).Fatal("ApplyConfig=>Widgets")
	}

	/////////////////////////////////////////////////////////////////////
	// Place Apps ///////////////////////////////////////////////////////
	utils.Indent(log.Info)("creating App folders and adding apps to them")
	missing, err = lpad.GetMissing(config.Apps, database.ApplicationType)
	if err != nil {
		log.WithError(err).Fatal("GetMissing=>Apps")
	}

	config.Apps.Pages = parseMissing(missing, config.Apps.Pages)
	groupID, err = lpad.ApplyConfig(config.Apps, database.ApplicationType, groupID, 1)
	if err != nil {
		log.WithError(err).Fatal("ApplyConfig=>Apps")
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

func main() {

	cli.AppHelpTemplate = appHelpTemplate
	app := cli.NewApp()

	app.Name = "lporg"
	app.Author = "blacktop"
	app.Email = "https://github.com/blacktop"
	complTime, _ := time.Parse(time.RFC3339, BuildTime)
	app.Version = Version + ", BuildTime: " + complTime.Format("20060102")
	app.Compiled = complTime
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
			Usage: "organize by default app categories",
			Action: func(c *cli.Context) error {
				fmt.Println(porg)

				backup := false
				prompt := &survey.Confirm{
					Message: "Backup your current Launchpad settings?",
				}
				survey.AskOne(prompt, &backup, nil)

				if backup {
					err := CmdSaveConfig(c.GlobalBool("verbose"), "launchpad.BACKUP.yaml")
					if err != nil {
						return err
					}
					log.Infof(bold, "successfully backed up current settings!")
					fmt.Println()
				}

				return CmdDefaultOrg(c.GlobalBool("verbose"))
			},
		},
		{
			Name:  "save",
			Usage: "save current launchpad settings",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config, c",
					Usage: "Save configuration to `FILE`",
				},
			},
			Action: func(c *cli.Context) error {
				return CmdSaveConfig(c.GlobalBool("verbose"), c.String("config"))
			},
		},
		{
			Name:  "load",
			Usage: "load launchpad settings config from `FILE`",
			Action: func(c *cli.Context) error {
				if c.Args().Present() {

					backup := false
					prompt := &survey.Confirm{
						Message: "Backup your current Launchpad settings?",
					}
					survey.AskOne(prompt, &backup, nil)

					if backup {
						err := CmdSaveConfig(c.GlobalBool("verbose"), "launchpad.BACKUP.yaml")
						if err != nil {
							return err
						}
						log.Infof(bold, "successfully backed up current settings!")
						fmt.Println()
					}

					// user supplied launchpad config YAMLdep
					err := CmdLoadConfig(c.GlobalBool("verbose"), c.Args().First())
					if err != nil {
						return err
					}
				} else {
					log.Fatal("please supply a config file to load")
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
