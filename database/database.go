package database

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/apex/log"
	"github.com/blacktop/lporg/database/utils"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// Config is the Launchpad config
type Config map[string][]map[string][]string

// LoadConfig loads the Launchpad config from the config file
func LoadConfig(filename string) (Config, error) {
	var pages Config

	utils.Indent(log.WithField("path", filename).Info)("parsing launchpad config YAML")
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		utils.DoubleIndent(log.WithError(err).WithField("path", filename).Fatal)("config file not found")
		return nil, err
	}

	err = yaml.Unmarshal(data, &pages)
	if err != nil {
		utils.DoubleIndent(log.WithError(err).WithField("path", filename).Fatal)("unmarshalling yaml failed")
		return nil, err
	}

	return pages, nil
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

		// if success := lp.DB.NewRecord(item); success == false {
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

// CreateAppFolders creates all the launchpad app folders
//    group_id = setup_items(conn, Types.APP, app_layout, app_mapping, group_id, root_parent_id=1)
func (lp *LaunchPad) CreateAppFolders(config map[string][]map[string][]string, groupID int) error {
	utils.Indent(log.Info)("creating app folders and adding apps to them")

	for index, page := range config["pages"] {
		// Start a new page (note that the ordering starts at 1 instead of 0 as there is a holding page at an ordering of 0)
		groupID++

		item := Item{
			ID:       groupID,
			UUID:     newUUID(),
			Flags:    PageType,
			ParentID: 1,
			Ordering: index + 1,
		}

		if success := lp.DB.NewRecord(item); success == false {
			utils.DoubleIndent(log.WithField("item", item).Debug)("create new item record failed")
		}
		if err := lp.DB.Create(&item).Error; err != nil {
			return err
		}

		group := Group{ID: groupID} // omitting fields makes them null

		if success := lp.DB.NewRecord(group); success == false {
			utils.Indent(log.WithField("group", group).Debug)("create new group record failed")
		}
		if err := lp.DB.Create(&group).Error; err != nil {
			return err
		}
		// Capture the group id of the page to be used for child items
		pageParentID := groupID

		// Iterate through items
		itemOrdering := 0
		for folderName, items := range page {
			// Start a new folder
			groupID++

			item := Item{
				ID:       groupID,
				UUID:     newUUID(),
				Flags:    FolderRootType,
				ParentID: pageParentID,
				Ordering: itemOrdering,
			}

			if success := lp.DB.NewRecord(item); success == false {
				utils.DoubleIndent(log.WithField("item", item).Debug)("create new item record failed")
			}
			if err := lp.DB.Create(&item).Error; err != nil {
				return err
			}

			group := Group{
				ID:         groupID,
				CategoryID: 0,
				Title:      folderName,
			}
			utils.Indent(log.WithField("group", group).Info)("group being added")
			if success := lp.DB.NewRecord(group); success == false {
				utils.Indent(log.WithField("group", group).Debug)("create new group record failed")
			}
			if err := lp.DB.Create(&group).Error; err != nil {
				return err
			}

			itemOrdering++
			// Capture the group id of the folder root to be used for child items
			folderRootParentID := groupID
			for index, item := range items {

				utils.DoubleIndent(log.WithField("item", item).Info)("item being added")

				// Start a new folder page
				groupID++

				item := Item{
					ID:       groupID,
					UUID:     newUUID(),
					Flags:    PageType,
					ParentID: folderRootParentID,
					Ordering: index,
				}

				if success := lp.DB.NewRecord(item); success == false {
					utils.DoubleIndent(log.WithField("item", item).Debug)("create new item record failed")
				}
				if err := lp.DB.Create(&item).Error; err != nil {
					return err
				}

				group := Group{ID: groupID}

				if success := lp.DB.NewRecord(group); success == false {
					utils.Indent(log.WithField("group", group).Debug)("create new group record failed")
				}
				if err := lp.DB.Create(&group).Error; err != nil {
					return err
				}
			}
		}
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

// EnableTriggers enables item update triggers
func (lp *LaunchPad) EnableTriggers() error {

	utils.Indent(log.Info)("enabling SQL update triggers")

	if err := lp.DB.Exec("UPDATE dbinfo SET value = 0 WHERE key = 'ignore_items_update_triggers';").Error; err != nil {
		return errors.Wrap(err, "counld not update `ignore_items_update_triggers` to 0")
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
