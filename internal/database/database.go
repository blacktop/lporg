// Package database provides launchpad database functions
package database

import (
	"fmt"
	"sort"

	"github.com/apex/log"
	"github.com/blacktop/lporg/internal/utils"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

// GetMissing returns a list of the rest of the apps not in the config
func (lp *LaunchPad) GetMissing(apps *Apps, appType int) error {

	var (
		dbApps     []string
		configApps []string
	)

	// get all apps from database
	switch appType {
	case ApplicationType:
		var apps []App
		err := lp.DB.Table("apps").
			Select("apps.item_id, apps.title").
			Joins("left join items on items.rowid = apps.item_id").
			Not("items.parent_id = ?", 6).
			Scan(&apps).Error
		if err != nil {
			return fmt.Errorf("query all apps failed: %w", err)
		}
		for _, app := range apps {
			dbApps = append(dbApps, app.Title)
		}
	default:
		return fmt.Errorf("GetMissing: unsupported app type: %d", appType)
	}

	sort.Strings(dbApps)

	// get all apps from config file
	for _, page := range apps.Pages {
		for _, item := range page.Items {
			switch item.(type) {
			case string:
				configApps = append(configApps, item.(string))
			default:
				var folder AppFolder
				if err := mapstructure.Decode(item, &folder); err != nil {
					return fmt.Errorf("mapstructure unable to decode config folder: %w", err)
				}
				for _, fpage := range folder.Pages {
					for _, fitem := range fpage.Items {
						configApps = append(configApps, fitem)
					}
				}
			}
		}
	}

	sort.Strings(configApps)

	for _, app := range dbApps {
		if !slices.Contains(configApps, app) {
			utils.DoubleIndent(log.WithField("app", app).Warn)("found installed apps that are not in supplied config")
			if len(apps.Pages[len(apps.Pages)-1].Items) < 35 {
				apps.Pages[len(apps.Pages)-1].Items = append(apps.Pages[len(apps.Pages)-1].Items, app)
			} else {
				newPage := Page{
					Number: len(apps.Pages) + 1,
					Items:  []any{app},
				}
				apps.Pages = append(apps.Pages, newPage)
			}
		}
	}

	// check all apps from config file exist on system
	for idx, page := range apps.Pages {
		for iidx, item := range page.Items {
			switch item.(type) {
			case string:
				if !slices.Contains(dbApps, item.(string)) {
					utils.DoubleIndent(log.WithField("app", item.(string)).Warn)("found app in config that are is not on system")
					apps.Pages[idx].Items = append(apps.Pages[idx].Items[:iidx], apps.Pages[idx].Items[iidx+1:]...)
				}
			default:
				var folder AppFolder
				if err := mapstructure.Decode(item, &folder); err != nil {
					return fmt.Errorf("mapstructure unable to decode config folder: %w", err)
				}
				for fpIdx, fpage := range folder.Pages {
					for fpiIdx, fitem := range fpage.Items {
						if !slices.Contains(dbApps, fitem) {
							utils.DoubleIndent(log.WithField("app", fitem).Warn)("found app in config that are is not on system")
							apps.Pages[idx].Items[iidx].(map[string]any)["pages"].([]any)[fpIdx].(map[string]any)["items"] = append(
								apps.Pages[idx].Items[iidx].(map[string]any)["pages"].([]any)[fpIdx].(map[string]any)["items"].([]any)[:fpiIdx],
								apps.Pages[idx].Items[iidx].(map[string]any)["pages"].([]any)[fpIdx].(map[string]any)["items"].([]any)[fpiIdx+1:]...)
						}
					}
				}
			}
		}
	}

	return nil
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
		{ID: 1, UUID: "ROOTPAGE", Type: RootType, ParentID: 0, Ordering: 0},
		{ID: 2, UUID: "HOLDINGPAGE", Type: PageType, ParentID: 1, Ordering: 0},
		{ID: 3, UUID: "ROOTPAGE_DB", Type: RootType, ParentID: 0, Ordering: 0},
		{ID: 4, UUID: "HOLDINGPAGE_DB", Type: PageType, ParentID: 3, Ordering: 0},
		{ID: 5, UUID: "ROOTPAGE_VERS", Type: RootType, ParentID: 0, Ordering: 0},
		{ID: 6, UUID: "HOLDINGPAGE_VERS", Type: PageType, ParentID: 5, Ordering: 0},
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
		UUID:     uuid.New().String(),
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
		UUID:     uuid.New().String(),
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

	utils.DoubleIndent(log.WithField("group", group.Title).Info)("folder added")

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
		UUID:     uuid.New().String(),
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

// updateItem will add the apps/widgets to the correct page/folder
func (lp *LaunchPad) updateItem(item string, ordering, groupID, itemType int) error {

	var (
		i Item
		a App
		w Widget
	)

	i = Item{}
	a = App{}
	w = Widget{}

	switch itemType {
	case ApplicationType:

		if result := lp.DB.Where("title = ?", item).First(&a); result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
			utils.DoubleIndent(log.WithField("app", item).Warn)("app not installed. SKIPPING...")
			return nil
		}
		if err := lp.DB.Where("rowid = ?", a.ID).First(&i).Error; err != nil {
			return errors.Wrap(err, "item query failed for app: "+item)
		}

		lp.DB.Model(&i).Association("App").Find(&i.App)
	case WidgetType:
		if result := lp.DB.Where("title = ?", item).First(&w); result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
			utils.DoubleIndent(log.WithField("app", item).Warn)("widget not installed. SKIPPING...")
			return nil
		}
		if err := lp.DB.Where("rowid = ?", w.ID).First(&i).Error; err != nil {
			return errors.Wrap(err, "item query failed for widget: "+item)
		}

		lp.DB.Model(&i).Association("Widget").Find(&i.Widget)
	default:
		utils.DoubleIndent(log.WithField("type", itemType).Error)("bad type")
	}

	newItem := Item{
		ID:       i.ID,
		UUID:     i.UUID,
		Flags:    i.Flags,
		Type:     itemType,
		ParentID: groupID,
		Ordering: ordering,
	}

	// if !lp.DB.NewRecord(newItem) {
	// 	utils.DoubleIndent(log.WithField("item", newItem).Debug)("createItems - create new item record failed")
	// }
	return lp.DB.Save(&newItem).Error
}

// ApplyConfig places all the launchpad apps
func (lp *LaunchPad) ApplyConfig(config Apps, itemType, groupID, rootParentID int) (int, error) {

	for _, page := range config.Pages {
		// create a new page
		groupID++
		err := lp.createNewPage(page.Number, groupID, rootParentID)
		if err != nil {
			return groupID, errors.Wrap(err, "createNewPage")
		}

		pageParentID := groupID

		for idx, item := range page.Items {
			switch item.(type) {
			case string:
				// add a flat item
				if err := lp.updateItem(item.(string), idx, pageParentID, itemType); err != nil {
					return groupID, errors.Wrap(err, "updateItem")
				}
			default:
				var folder AppFolder
				if err := mapstructure.Decode(item, &folder); err != nil {
					return groupID, errors.Wrap(err, "mapstructure unable to decode config folder")
				}

				// create a new folder
				groupID++
				err := lp.createNewFolder(folder.Name, idx, groupID, pageParentID)
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
					for fidx, fitem := range fpage.Items {
						if err := lp.updateItem(fitem, fidx, groupID, itemType); err != nil {
							return groupID, errors.Wrap(err, "updateItem")
						}
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
// deprecated
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
