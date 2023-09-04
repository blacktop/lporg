// Package command provides the command line interface functionality for lporg.
package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/apex/log"
	"github.com/blacktop/lporg/internal/database"
	"github.com/blacktop/lporg/internal/desktop"
	"github.com/blacktop/lporg/internal/dock"
	"github.com/blacktop/lporg/internal/utils"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const bold = "\033[1m%s\033[0m"

// Config is the command config
type Config struct {
	Cmd      string
	File     string
	Cloud    bool
	Backup   bool
	LogLevel int
}

// Verify will verify the command config
func (c *Config) Verify() error {
	if c.Cloud && len(c.File) > 0 {
		return fmt.Errorf("cannot use --config with --icloud")
	}

	switch c.Cmd {
	case "revert":
		if c.Cloud {
			iCloudPath, err := getiCloudDrivePath()
			if err != nil {
				return fmt.Errorf("get iCloud drive path failed")
			}
			host, err := os.Hostname()
			if err != nil {
				return fmt.Errorf("failed to get hostname")
			}
			c.File = filepath.Join(iCloudPath, ".config", "lporg", strings.TrimRight(host, ".local")+".yml.bak")
		} else {
			if len(c.File) == 0 { // set DEFAULT config file
				confDir, err := os.UserConfigDir()
				if err != nil {
					return fmt.Errorf("failed to get user config dir")
				}
				c.File = filepath.Join(confDir, "lporg", "config.yml.bak")
			}
		}
	case "load":
		if len(c.File) == 0 && !c.Cloud {
			return fmt.Errorf("must supply --config file OR use --icloud")
		}
		fallthrough
	default:
		if c.Cloud { // use iCloud to store config
			iCloudPath, err := getiCloudDrivePath()
			if err != nil {
				return fmt.Errorf("get iCloud drive path failed")
			}
			host, err := os.Hostname()
			if err != nil {
				return fmt.Errorf("failed to get hostname")
			}
			c.File = filepath.Join(iCloudPath, ".config", "lporg", strings.TrimRight(host, ".local")+".yml")
		} else {
			if len(c.File) == 0 { // set DEFAULT config file
				confDir, err := os.UserConfigDir()
				if err != nil {
					return fmt.Errorf("failed to get user config dir")
				}
				c.File = filepath.Join(confDir, "lporg", "config.yml")
			}
		}
	}

	if err := os.MkdirAll(filepath.Dir(c.File), 0750); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	log.Info("using config file: " + c.File)

	return nil
}

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
			// case database.WidgetType:
			// 	utils.Indent(log.WithField("title", item.Widget.Title).Info)("found widget")
			// 	p.Items = append(p.Items, item.Widget.Title)
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

// DefaultOrg will organize your launchpad by the app default categories
func DefaultOrg(c *Config) (err error) {
	var lpad database.LaunchPad

	log.Infof(bold, "USING DEFAULT LAUNCHPAD ORGANIZATION")

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
	err = removeOldDatabaseFiles(lpad.Folder)
	if err != nil {
		return err
	}

	// open launchpad database
	lpad.DB, err = gorm.Open("sqlite3", lpad.File)
	if err != nil {
		return err
	}
	defer lpad.DB.Close()

	if c.LogLevel > 0 {
		lpad.DB.LogMode(true)
	}

	// // open launchpad database
	// lpad.DB, err = gorm.Open(sqlite.Open(lpad.File), &gorm.Config{
	// 	Logger: logger.Default.LogMode(logger.LogLevel(c.LogLevel)),
	// })
	// if err != nil {
	// 	return err
	// }
	// defer func() {
	// 	db, err := lpad.DB.DB()
	// 	if err != nil {
	// 		err = errors.Wrap(err, "unable to get db when trying to close")
	// 	}
	// 	err = db.Close()
	// 	if err != nil {
	// 		err = errors.Wrap(err, "unable to close db")
	// 	}
	// }()

	// Clear all items related to groups so we can re-create them
	if err := lpad.ClearGroups(); err != nil {
		return fmt.Errorf("failed to ClearGroups: %w", err)
	}

	// Disable the update triggers
	if err := lpad.DisableTriggers(); err != nil {
		return fmt.Errorf("failed to DisableTriggers: %w", err)
	}

	// Add root and holding pages to items and groups
	if err := lpad.AddRootsAndHoldingPages(); err != nil {
		return fmt.Errorf("failed to AddRootsAndHoldingPagesfailed: %w", err)
	}

	// We will begin our group records using the max ids found (groups always appear after apps and widgets)
	// groupID := int(math.Max(float64(lpad.GetMaxAppID()), float64(lpad.GetMaxWidgetID())))
	groupID := int(float64(lpad.GetMaxAppID())) // widgets are no longer supported

	utils.Indent(log.Info)("creating folders out of app categories")

	// Create default config file
	var apps []database.App
	var categories []database.Category
	var dbconf database.Config

	page := database.Page{Number: 1}

	if err := lpad.DB.Find(&categories).Error; err != nil {
		log.WithError(err).Error("categories query failed")
	}

	for _, category := range categories {
		folderName := strings.Title(strings.Replace(strings.TrimPrefix(category.UTI, "public.app-category."), "-", " ", 1))
		folder := database.AppFolder{Name: folderName}
		folderPage := database.FolderPage{Number: 1}
		utils.DoubleIndent(log.WithField("folder", folderName).Info)("adding folder")
		if err := lpad.DB.Where("category_id = ?", category.ID).Find(&apps).Error; err != nil {
			log.WithError(err).Error("categories query failed")
		}
		for _, app := range apps {
			utils.TripleIndent(log.WithField("app", app.Title).Info)("adding app to category folder")
			folderPage.Items = utils.AppendIfMissing(folderPage.Items, app.Title)
		}
		folder.Pages = append(folder.Pages, folderPage)
		page.Items = append(page.Items, folder)
	}

	dbconf.Apps.Pages = append(dbconf.Apps.Pages, page)

	////////////////////////////////////////////////////////////////////
	// Place Widgets ///////////////////////////////////////////////////
	// utils.Indent(log.Info)("creating Widget folders and adding widgets to them")
	// missing, err := lpad.GetMissing(conf.Widgets, database.WidgetType)
	// if err != nil {
	// 	log.WithError(err).Fatal("Default GetMissing=>Widgets")
	// }

	// conf.Widgets.Pages = parseMissing(missing, conf.Widgets.Pages)
	// groupID, err = lpad.ApplyConfig(conf.Widgets, database.WidgetType, groupID, 3)
	// if err != nil {
	// 	log.WithError(err).Fatal("Default ApplyConfig=>Widgets")
	// }

	/////////////////////////////////////////////////////////////////////
	// Place Apps ///////////////////////////////////////////////////////
	utils.Indent(log.Info)("creating App folders and adding apps to them")
	missing, err := lpad.GetMissing(dbconf.Apps, database.ApplicationType)
	if err != nil {
		return fmt.Errorf("failed to GetMissing=>Apps: %v", err)
	}

	dbconf.Apps.Pages = parseMissing(missing, dbconf.Apps.Pages)
	groupID, err = lpad.ApplyConfig(dbconf.Apps, database.ApplicationType, groupID, 1)
	if err != nil {
		return fmt.Errorf("failed to ApplyConfig: %v", err)
	}

	// Re-enable the update triggers
	if err := lpad.EnableTriggers(); err != nil {
		return fmt.Errorf("failed to EnableTriggers: %v", err)
	}

	return restartDock()
}

// SaveConfig will save your launchpad settings to a config file
func SaveConfig(c *Config) (err error) {
	var (
		lpad          database.LaunchPad
		launchpadRoot int
		dashboardRoot int
		items         []database.Item
		dbinfo        []database.DBInfo
		conf          database.Config
	)

	log.Infof(bold, "SAVING LAUNCHPAD DATABASE")

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
	lpad.DB, err = gorm.Open("sqlite3", lpad.File)
	if err != nil {
		return err
	}
	defer lpad.DB.Close()

	if c.LogLevel > 0 {
		lpad.DB.LogMode(true)
	}

	// // open launchpad database
	// lpad.DB, err = gorm.Open(sqlite.Open(lpad.File), &gorm.Config{
	// 	Logger: logger.Default.LogMode(logger.LogLevel(c.LogLevel)),
	// })
	// if err != nil {
	// 	return err
	// }
	// defer func() {
	// 	db, err := lpad.DB.DB()
	// 	if err != nil {
	// 		err = errors.Wrap(err, "unable to get db when trying to close")
	// 	}
	// 	err = db.Close()
	// 	if err != nil {
	// 		err = errors.Wrap(err, "unable to close db")
	// 	}
	// }()

	// get launchpad and dashboard roots
	if err := lpad.DB.Where("key in (?)", []string{"launchpad_root", "dashboard_root"}).Find(&dbinfo).Error; err != nil {
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
	if err := lpad.DB.Not("uuid in (?)", []string{"ROOTPAGE", "HOLDINGPAGE", "ROOTPAGE_DB", "HOLDINGPAGE_DB", "ROOTPAGE_VERS", "HOLDINGPAGE_VERS"}).
		Order("items.parent_id, items.ordering").
		Find(&items).Error; err != nil {
		log.WithError(err).Error("items query failed")
	}

	// create parent mapping object
	log.Info("collecting launchpad/dashboard pages")
	parentMapping := make(map[int][]database.Item)
	for _, item := range items {
		lpad.DB.Model(&item).Related(&item.App)
		// db.Model(&item).Related(&item.Widget)
		lpad.DB.Model(&item).Related(&item.Group)

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

	log.Info("interating over dock apps")
	dPlist, err := dock.LoadDockPlist()
	if err != nil {
		return errors.Wrap(err, "unable to load dock plist")
	}
	for _, item := range dPlist.PersistentApps {
		conf.Dock.Apps = append(conf.Dock.Apps, item.TileData.GetPath())
	}
	for _, item := range dPlist.PersistentOthers {
		conf.Dock.Others = append(conf.Dock.Others, item.TileData.GetPath())
	}

	// write out config YAML file
	d, err := yaml.Marshal(&conf)
	if err != nil {
		return errors.Wrap(err, "unable to marshall YAML")
	}

	if c.Backup {
		c.File += ".bak"
	}

	if err = os.WriteFile(c.File, d, 0644); err != nil {
		return errors.Wrap(err, "unable to write YAML")
	}

	if c.Backup {
		log.Infof(bold, "successfully backed up current settings!")
	} else {
		log.Infof(bold, "successfully wrote: "+c.File)
	}

	return nil
}

// LoadConfig will load your launchpad settings from a config file
func LoadConfig(c *Config) error {
	var lpad database.LaunchPad

	// Read in Config file
	config, err := database.LoadConfig(c.File)
	if err != nil {
		return fmt.Errorf("failed to load config file: %v", err)
	}

	log.Infof(bold, "PARSE LAUCHPAD DATABASE")

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
	err = removeOldDatabaseFiles(lpad.Folder)
	if err != nil {
		return err
	}

	// open launchpad database
	lpad.DB, err = gorm.Open("sqlite3", lpad.File)
	if err != nil {
		return err
	}
	defer lpad.DB.Close()

	if c.LogLevel > 0 {
		lpad.DB.LogMode(true)
	}

	// // open launchpad database
	// lpad.DB, err = gorm.Open(sqlite.Open(lpad.File), &gorm.Config{
	// 	Logger: logger.Default.LogMode(logger.LogLevel(c.LogLevel)),
	// })
	// if err != nil {
	// 	return err
	// }
	// defer func() {
	// 	db, err := lpad.DB.DB()
	// 	if err != nil {
	// 		err = errors.Wrap(err, "unable to get db when trying to close")
	// 	}
	// 	err = db.Close()
	// 	if err != nil {
	// 		err = errors.Wrap(err, "unable to close db")
	// 	}
	// }()

	// Clear all items related to groups so we can re-create them
	if err := lpad.ClearGroups(); err != nil {
		return fmt.Errorf("failed to ClearGroups: %v", err)
	}

	// Disable the update triggers
	if err := lpad.DisableTriggers(); err != nil {
		return fmt.Errorf("failed to DisableTriggers: %v", err)
	}

	// Add root and holding pages to items and groups
	if err := lpad.AddRootsAndHoldingPages(); err != nil {
		return fmt.Errorf("failed to AddRootsAndHoldingPagesfailed: %v", err)
	}

	// We will begin our group records using the max ids found (groups always appear after apps and widgets)
	// groupID := int(math.Max(float64(lpad.GetMaxAppID()), float64(lpad.GetMaxWidgetID())))
	groupID := int(float64(lpad.GetMaxAppID())) // widgets are no longer supported

	////////////////////////////////////////////////////////////////////
	// Place Widgets ///////////////////////////////////////////////////
	// utils.Indent(log.Info)("creating Widget folders and adding widgets to them")
	// missing, err := lpad.GetMissing(config.Widgets, database.WidgetType)
	// if err != nil {
	// 	log.WithError(err).Fatal("GetMissing=>Widgets")
	// }

	// config.Widgets.Pages = parseMissing(missing, config.Widgets.Pages)
	// groupID, err = lpad.ApplyConfig(config.Widgets, database.WidgetType, groupID, 3)
	// if err != nil {
	// 	log.WithError(err).Fatal("ApplyConfig=>Widgets")
	// }

	/////////////////////////////////////////////////////////////////////
	// Place Apps ///////////////////////////////////////////////////////
	utils.Indent(log.Info)("creating App folders and adding apps to them")
	missing, err := lpad.GetMissing(config.Apps, database.ApplicationType)
	if err != nil {
		return fmt.Errorf("failed to GetMissing=>Apps: %v", err)
	}

	config.Apps.Pages = parseMissing(missing, config.Apps.Pages)
	groupID, err = lpad.ApplyConfig(config.Apps, database.ApplicationType, groupID, 1)
	if err != nil {
		return fmt.Errorf("failed to ApplyConfig=>Apps: %v", err)
	}

	// Re-enable the update triggers
	if err := lpad.EnableTriggers(); err != nil {
		return fmt.Errorf("failed to EnableTriggers: %v", err)
	}

	if len(config.Desktop.Image) > 0 {
		utils.Indent(log.WithField("image", config.Desktop.Image).Info)("setting desktop background image")
		desktop.SetDesktopImage(config.Desktop.Image)
	}

	if len(config.Dock.Apps) > 0 || len(config.Dock.Others) > 0 {
		utils.Indent(log.Info)("setting dock apps")
		dPlist, err := dock.LoadDockPlist()
		if err != nil {
			return errors.Wrap(err, "unable to load dock plist")
		}
		if len(dPlist.PersistentApps) > 0 {
			dPlist.PersistentApps = nil // remove all apps from dock
		}
		for _, app := range config.Dock.Apps {
			utils.DoubleIndent(log.WithField("app", app).Info)("adding to dock")
			dPlist.AddApp(app)
		}
		if len(dPlist.PersistentOthers) > 0 {
			dPlist.PersistentOthers = nil // remove all folders from dock
		}
		for _, other := range config.Dock.Others {
			utils.DoubleIndent(log.WithField("other", other).Info)("adding to dock")
			dPlist.AddOther(other)
		}
		if err := dPlist.Save(); err != nil {
			return fmt.Errorf("failed to save dock plist: %w", err)
		}
	}

	return restartDock()
}
