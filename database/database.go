package database

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"sort"

	"github.com/apex/log"
	"github.com/blacktop/lporg/database/utils"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// Config is the Launchpad config
type Config struct {
	Apps    Apps `yaml:"apps" json:"apps,omitempty"`
	Widgets Apps `yaml:"widgets" json:"widgets,omitempty"`
}

// Apps is the launchpad apps config object
type Apps struct {
	Pages []Page `yaml:"pages" json:"pages,omitempty"`
}

// Page is a launchpad page object
type Page struct {
	Number    int      `yaml:"number" json:"number"`
	FlatItems []string `yaml:"flat_items" json:"flat_items,omitempty"`
	Folders   []Folder `yaml:"folders" json:"folders,omitempty"`
}

// Folder is a launchpad folder object
type Folder struct {
	Name  string       `yaml:"name" json:"name,omitempty"`
	Pages []FolderPage `yaml:"pages" json:"pages,omitempty"`
}

// FolderPage is a launchpad folder page object
type FolderPage struct {
	Number int      `yaml:"number" json:"number"`
	Items  []string `yaml:"items" json:"items,omitempty"`
}

// LoadConfig loads the Launchpad config from the config file
func LoadConfig(filename string) (Config, error) {
	var conf Config

	utils.Indent(log.WithField("path", filename).Info)("parsing launchpad config YAML")
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		utils.DoubleIndent(log.WithError(err).WithField("path", filename).Fatal)("config file not found")
		return conf, err
	}

	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		utils.DoubleIndent(log.WithError(err).WithField("path", filename).Fatal)("unmarshalling yaml failed")
		return conf, err
	}

	return conf, nil
}

// GetMissing returns a list of the rest of the apps not in the config
func (lp *LaunchPad) GetMissing(apps Apps, appType int) ([]string, error) {

	missing := []string{}
	configApps := []string{}

	// get all apps from config file
	for _, page := range apps.Pages {
		for _, item := range page.FlatItems {
			configApps = append(configApps, item)
		}
		for _, folder := range page.Folders {
			for _, fpage := range folder.Pages {
				for _, fitem := range fpage.Items {
					configApps = append(configApps, fitem)
				}
			}
		}
	}

	switch appType {
	case ApplicationType:
		var apps []App
		err := lp.DB.Table("apps").Select("apps.item_id, apps.title").Joins("left join items on items.rowid = apps.item_id").Scan(&apps).Error
		if err != nil {
			return nil, err
		}
		for _, app := range apps {
			if !utils.StringInSlice(app.Title, configApps) {
				missing = utils.AppendIfMissing(missing, app.Title)
			}
		}
	case WidgetType:
		fallthrough
	default:
		utils.DoubleIndent(log.WithField("type", appType).Error)("bad type")
	}

	sort.Strings(missing)

	return missing, nil
}

// ClearGroups clears out items related to groups
func (lp *LaunchPad) ClearGroups() error {
	utils.Indent(log.Info)("clear out groups")

	var items []Item
	return lp.DB.Where("type in (?)", []int{RootType, FolderRootType, PageType}).Delete(&items).Error
}

// AddRootsAndHoldingPages adds back in the RootPage and HoldingPage defaults
func (lp *LaunchPad) AddRootsAndHoldingPages() error {
	utils.Indent(log.Info)("add root and holding pages")

	items := []Item{
		Item{ID: 1, UUID: "ROOTPAGE", Type: RootType, ParentID: 0, Ordering: 0},
		Item{ID: 2, UUID: "HOLDINGPAGE", Type: PageType, ParentID: 1, Ordering: 0},
		Item{ID: 3, UUID: "ROOTPAGE_DB", Type: RootType, ParentID: 0, Ordering: 0},
		Item{ID: 4, UUID: "HOLDINGPAGE_DB", Type: PageType, ParentID: 3, Ordering: 0},
		Item{ID: 5, UUID: "ROOTPAGE_VERS", Type: RootType, ParentID: 0, Ordering: 0},
		Item{ID: 6, UUID: "HOLDINGPAGE_VERS", Type: PageType, ParentID: 5, Ordering: 0},
	}

	for _, item := range items {
		group := Group{ID: item.ID}

		// if !lp.DB.NewRecord(item) {
		// 	log.Error("create new record failed")
		// }
		if err := lp.DB.Create(&item).Error; err != nil {
			return errors.Wrap(err, "db insert item failed")
		}
		if err := lp.DB.Create(&group).Error; err != nil {
			return errors.Wrap(err, "db insert group failed")
		}
	}

	return nil
}

// createNewPage creates a new page
func (lp *LaunchPad) createNewPage(pageNumber, groupID, pageParentID int) error {

	item := Item{
		ID:       groupID,
		UUID:     newUUID(),
		Flags:    2,
		Type:     PageType,
		ParentID: pageParentID,
		Ordering: pageNumber, // TODO: check if I should use 0 base index or 1 (what I'm doing now)
	}

	// if !lp.DB.NewRecord(item) {
	// 	utils.DoubleIndent(log.WithField("item", item).Debug)("createNewPage - create new item record failed")
	// }
	if err := lp.DB.Create(&item).Error; err != nil {
		return errors.Wrap(err, "createNewPage")
	}

	group := Group{ID: groupID} // omitting fields makes them null

	// if !lp.DB.NewRecord(group) {
	// 	utils.Indent(log.WithField("group", group).Debug)("createNewPage - create new group record failed")
	// }
	if err := lp.DB.Create(&group).Error; err != nil {
		return errors.Wrap(err, "createNewPage")
	}

	return nil
}

// createNewFolder creates a new app folder
func (lp *LaunchPad) createNewFolder(folderName string, folderNumber, groupID, folderParentID int) error {

	item := Item{
		ID:       groupID,
		UUID:     newUUID(),
		Flags:    0,
		Type:     FolderRootType,
		ParentID: folderParentID,
		Ordering: folderNumber,
	}

	// if !lp.DB.NewRecord(item) {
	// 	utils.DoubleIndent(log.WithField("item", item).Debug)("createNewFolder - create new item record failed")
	// }
	if err := lp.DB.Create(&item).Error; err != nil {
		return errors.Wrap(err, "createNewFolder")
	}

	group := Group{
		ID:    groupID,
		Title: folderName,
	}

	utils.Indent(log.WithField("group", group).Info)("group being added")

	// if !lp.DB.NewRecord(group) {
	// 	utils.Indent(log.WithField("group", group).Debug)("createNewFolder - create new group record failed")
	// }
	if err := lp.DB.Create(&group).Error; err != nil {
		return errors.Wrap(err, "createNewFolder")
	}

	return nil
}

// createNewFolderPage creates a new folder page
func (lp *LaunchPad) createNewFolderPage(folderPageNumber, groupID, folderPageParentID int) error {

	item := Item{
		ID:       groupID,
		UUID:     newUUID(),
		Flags:    2,
		Type:     PageType,
		ParentID: folderPageParentID,
		Ordering: folderPageNumber,
	}

	// if !lp.DB.NewRecord(item) {
	// 	utils.DoubleIndent(log.WithField("item", item).Debug)("createNewFolderPage - create new item record failed")
	// }
	if err := lp.DB.Create(&item).Error; err != nil {
		return errors.Wrap(err, "createNewFolderPage")
	}

	group := Group{ID: groupID}
	// if !lp.DB.NewRecord(group) {
	// 	utils.Indent(log.WithField("group", group).Debug)("createNewFolderPage - create new group record failed")
	// }
	if err := lp.DB.Create(&group).Error; err != nil {
		return errors.Wrap(err, "createNewFolderPage")
	}

	return nil
}

// updateItems will add the apps/widgets to the correct page/folder
func (lp *LaunchPad) updateItems(items []string, groupID, itemType int) error {

	var (
		i Item
		a App
		w Widget
	)

	for iidx, item := range items {

		i = Item{}
		a = App{}
		w = Widget{}

		switch itemType {
		case ApplicationType:
			if lp.DB.Where("title = ?", item).First(&a).RecordNotFound() {
				utils.DoubleIndent(log.WithField("app", item).Warn)("app not installed. SKIPPING...")
				continue
			}
			if err := lp.DB.Where("rowid = ?", a.ID).First(&i).Error; err != nil {
				return errors.Wrap(err, "createItems")
			}

			lp.DB.Model(&i).Related(&i.App)
		case WidgetType:
			if lp.DB.Where("title = ?", item).First(&w).RecordNotFound() {
				utils.DoubleIndent(log.WithField("app", item).Warn)("widget not installed. SKIPPING...")
				continue
			}
			if err := lp.DB.Where("rowid = ?", w.ID).First(&i).Error; err != nil {
				return errors.Wrap(err, "createItems")
			}

			lp.DB.Model(&i).Related(&i.Widget)
		default:
			utils.DoubleIndent(log.WithField("type", itemType).Error)("bad type")
		}

		newItem := Item{
			ID:       i.ID,
			UUID:     i.UUID,
			Flags:    i.Flags,
			ParentID: groupID,
			Ordering: iidx,
		}

		// if !lp.DB.NewRecord(newItem) {
		// 	utils.DoubleIndent(log.WithField("item", newItem).Debug)("createItems - create new item record failed")
		// }
		if err := lp.DB.Save(&newItem).Error; err != nil {
			return err
		}
	}

	return nil
}

// ApplyConfig places all the launchpad apps
func (lp *LaunchPad) ApplyConfig(config Apps, itemType, groupID, rootParentID int) (int, error) {

	utils.Indent(log.Info)("creating app folders and adding apps to them")

	for _, page := range config.Pages {
		// create a new page
		groupID++
		err := lp.createNewPage(page.Number, groupID, rootParentID)
		if err != nil {
			return groupID, errors.Wrap(err, "createNewPage")
		}

		pageParentID := groupID

		if len(page.FlatItems) > 0 {
			// add all the flat items
			if err := lp.updateItems(page.FlatItems, pageParentID, itemType); err != nil {
				return groupID, errors.Wrap(err, "createItems")
			}
		}

		if len(page.Folders) > 0 {
			for fidx, folder := range page.Folders {
				// create a new folder
				groupID++
				err := lp.createNewFolder(folder.Name, fidx, groupID, pageParentID)
				if err != nil {
					return groupID, errors.Wrap(err, "createNewFolder")
				}

				folderParentID := groupID

				for _, fpage := range folder.Pages {
					// create a new folder page
					groupID++
					if err := lp.createNewFolderPage(fpage.Number, groupID, folderParentID); err != nil {
						return groupID, errors.Wrap(err, "createNewFolderPage")
					}

					// add all folder page items
					if err := lp.updateItems(fpage.Items, groupID, itemType); err != nil {
						return groupID, errors.Wrap(err, "createItems")
					}
				}
			}
		}
	}

	return groupID, nil
}

// EnableTriggers enables item update triggers
func (lp *LaunchPad) EnableTriggers() error {

	utils.Indent(log.Info)("enabling SQL update triggers")

	if err := lp.DB.Exec("UPDATE dbinfo SET value = 0 WHERE key = 'ignore_items_update_triggers';").Error; err != nil {
		return errors.Wrap(err, "counld not update `ignore_items_update_triggers` to 0")
	}

	return nil
}

// DisableTriggers disables item update triggers
func (lp *LaunchPad) DisableTriggers() error {

	utils.Indent(log.Info)("disabling SQL update triggers")

	if err := lp.DB.Exec("UPDATE dbinfo SET value = 1 WHERE key = 'ignore_items_update_triggers';").Error; err != nil {
		return errors.Wrap(err, "counld not update `ignore_items_update_triggers` to 1")
	}

	return nil
}

// GetMaxAppID returns the maximum App ItemID
func (lp *LaunchPad) GetMaxAppID() int {
	var apps []App

	if err := lp.DB.Find(&apps).Error; err != nil {
		utils.Indent(log.WithError(err).Error)("query all apps failed")
	}

	maxID := 0
	for _, app := range apps {
		if app.ID > maxID {
			maxID = app.ID
		}
	}

	return maxID
}

// GetMaxWidgetID returns the maximum Widget ItemID
func (lp *LaunchPad) GetMaxWidgetID() int {
	var widgets []Widget

	if err := lp.DB.Find(&widgets).Error; err != nil {
		utils.Indent(log.WithError(err).Error)("query all widgets failed")
	}

	maxID := 0
	for _, widget := range widgets {
		if widget.ID > maxID {
			maxID = widget.ID
		}
	}

	return maxID
}

// newUUID generates a random UUID according to RFC 4122
func newUUID() string {
	uuid := make([]byte, 16)
	_, _ = io.ReadFull(rand.Reader, uuid)
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
