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
			lp.dbApps = append(lp.dbApps, app.Title)
		}
	default:
		return fmt.Errorf("GetMissing: unsupported app type: %d", appType)
	}

	sort.Strings(lp.dbApps)

	// get all apps from config file
	for _, page := range apps.Pages {
		for _, item := range page.Items {
			switch item.(type) {
			case string:
				lp.confApps = append(lp.confApps, item.(string))
			default:
				var folder AppFolder
				if err := mapstructure.Decode(item, &folder); err != nil {
					return fmt.Errorf("mapstructure unable to decode config folder: %w", err)
				}
				lp.confFolders = append(lp.confFolders, folder.Name)
				for _, fpage := range folder.Pages {
					for _, fitem := range fpage.Items {
						lp.confApps = append(lp.confApps, fitem)
					}
				}
			}
		}
	}

	sort.Strings(lp.confApps)

	for _, app := range lp.dbApps {
		if !slices.Contains(lp.confApps, app) {
			utils.Indent(log.WithField("app", app).Warn, 3)("found installed apps that are not in supplied config")
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
		tmp := []any{}
		for _, item := range page.Items {
			switch item.(type) {
			case string:
				if !slices.Contains(lp.dbApps, item.(string)) {
					utils.Indent(log.WithField("app", item.(string)).Warn, 3)("found app in config that are is not on system")
				} else {
					tmp = append(tmp, item)
				}
			default:
				var folder AppFolder
				if err := mapstructure.Decode(item, &folder); err != nil {
					return fmt.Errorf("mapstructure unable to decode config folder: %w", err)
				}
				for fpIdx, fpage := range folder.Pages {
					ftmp := []any{}
					for _, fitem := range fpage.Items {
						if !slices.Contains(lp.dbApps, fitem) {
							utils.Indent(log.WithField("app", fitem).Warn, 3)("found app in config that are is not on system")
						} else {
							ftmp = append(ftmp, fitem)
						}
					}
					item.(map[string]any)["pages"].([]any)[fpIdx].(map[string]any)["items"] = ftmp
				}
				tmp = append(tmp, item)
			}
		}
		apps.Pages[idx].Items = tmp
	}

	return nil
}

// ClearGroups clears out items related to groups
func (lp *LaunchPad) ClearGroups() error {
	utils.Indent(log.Info, 2)("clear out groups")
	var items []Item
	if err := lp.DB.Where("type in (?)", []int{RootType, FolderRootType, PageType}).Delete(&items).Error; err != nil {
		return fmt.Errorf("delete items associted with groups failed: %w", err)
	}
	// return lp.DB.Exec("DELETE FROM groups;").Error
	return nil
}

// FlattenApps sets all the apps to the root page
func (lp *LaunchPad) FlattenApps() error {
	var apps []App

	if err := lp.DB.Find(&apps).Error; err != nil {
		return fmt.Errorf("query all apps failed: %w", err)
	}

	lp.DisableTriggers()

	utils.Indent(log.Info, 2)("flattening out apps")
	for idx, app := range apps {
		if err := lp.updateItem(app.Title, ApplicationType, lp.rootPage, idx); err != nil {
			return fmt.Errorf("failed to update app '%s': %w", app.Title, err)
		}
	}

	lp.EnableTriggers()

	return nil
}

// AddRootsAndHoldingPages adds back in the RootPage and HoldingPage defaults
func (lp *LaunchPad) AddRootsAndHoldingPages() error {

	items := []Item{
		{ID: 1, UUID: "ROOTPAGE", Type: RootType, ParentID: 0, Ordering: 0},
		{ID: 2, UUID: "HOLDINGPAGE", Type: PageType, ParentID: 1, Ordering: 0},
		// {ID: 3, UUID: "ROOTPAGE_DB", Type: PageType, ParentID: 0, Ordering: 0},
		{ID: 4, UUID: "HOLDINGPAGE_DB", Type: PageType, ParentID: 3, Ordering: 0},
		{ID: 5, UUID: "ROOTPAGE_VERS", Type: RootType, ParentID: 0, Ordering: 0},
		{ID: 6, UUID: "HOLDINGPAGE_VERS", Type: PageType, ParentID: 5, Ordering: 0},
	}

	utils.Indent(log.Info, 2)("add root and holding pages")
	for _, item := range items {
		if err := lp.DB.Create(&item).Error; err != nil {
			return errors.Wrap(err, "db insert item failed")
		}
		if err := lp.DB.Create(&Group{ID: item.ID}).Error; err != nil {
			return errors.Wrap(err, "db insert group failed")
		}
	}

	return nil
}

// createNewPage creates a new page
func (lp *LaunchPad) createNewPage(rowID, pageParentID, pageNumber int) error {

	item := Item{
		ID:       rowID,
		UUID:     uuid.New().String(),
		Flags:    2,
		Type:     PageType,
		ParentID: pageParentID,
		Ordering: pageNumber,
	}

	if err := lp.DB.Create(&item).Error; err != nil {
		return fmt.Errorf("failed to create page item with ID=%d: %w", rowID, err)
	}

	utils.Indent(log.WithField("number", pageNumber).Info, 3)("page added")
	if err := lp.DB.Create(&Group{ID: rowID}).Error; err != nil {
		return fmt.Errorf("failed to create group for page with ID=%d: %w", rowID, err)
	}

	return nil
}

// createNewFolder creates a new app folder
func (lp *LaunchPad) createNewFolder(folderName string, rowID, folderParentID, folderNumber int) error {

	item := Item{
		ID:       rowID,
		UUID:     uuid.New().String(),
		Flags:    0,
		Type:     FolderRootType,
		ParentID: folderParentID,
		Ordering: folderNumber,
	}

	if folderName == "Utilities" {
		item.Flags = 1
	}

	if err := lp.DB.Create(&item).Error; err != nil {
		return fmt.Errorf("failed to create folder '%s' item with ID=%d: %w", folderName, rowID, err)
	}

	utils.Indent(log.WithField("group", folderName).Info, 4)("folder added")
	if err := lp.DB.Create(&Group{
		ID:    rowID,
		Title: folderName,
	}).Error; err != nil {
		return fmt.Errorf("failed to create group for folder '%s' with ID=%d: %w", folderName, rowID, err)
	}

	return nil
}

// createNewFolderPage creates a new folder page
func (lp *LaunchPad) createNewFolderPage(rowID, folderPageParentID, folderPageNumber int) error {

	item := Item{
		ID:       rowID,
		UUID:     uuid.New().String(),
		Flags:    2,
		Type:     PageType,
		ParentID: folderPageParentID,
		Ordering: folderPageNumber,
	}

	if err := lp.DB.Create(&item).Error; err != nil {
		return fmt.Errorf("failed to create folder page item with ID=%d: %w", rowID, err)
	}
	utils.Indent(log.WithField("number", folderPageNumber).Info, 5)("folder page added")
	if err := lp.DB.Create(&Group{ID: rowID}).Error; err != nil {
		return fmt.Errorf("failed to create group for folder page with ID=%d: %w", rowID, err)
	}

	return nil
}

// updateItem will add the apps/widgets to the correct page/folder
func (lp *LaunchPad) updateItem(item string, itemType, parentID, ordering int) error {

	i := Item{}
	a := App{}
	w := Widget{}

	switch itemType {
	case ApplicationType:
		if err := lp.DB.Where("title = ?", item).First(&a).Error; err != nil {
			return fmt.Errorf("app query failed for '%s': %w", item, err)
		}
		if err := lp.DB.Where("rowid = ?", a.ID).First(&i).Error; err != nil {
			return fmt.Errorf("item query failed for app ID %d: %w", a.ID, err)
		}
		lp.DB.Model(&i).Association("App").Find(&i.App)
	case WidgetType:
		if result := lp.DB.Where("title = ?", item).First(&w); result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
			utils.Indent(log.WithField("app", item).Warn, 3)("widget not installed. SKIPPING...")
			return nil
		}
		if err := lp.DB.Where("rowid = ?", w.ID).First(&i).Error; err != nil {
			return fmt.Errorf("item query failed for wiget ID %d: %w", a.ID, err)
		}
		lp.DB.Model(&i).Association("Widget").Find(&i.Widget)
	default:
		return fmt.Errorf("failed to update item: unknown item type: %d", itemType)
	}

	newItem := Item{
		ID:       i.ID,
		UUID:     i.UUID,
		Flags:    i.Flags,
		Type:     itemType,
		ParentID: parentID,
		Ordering: ordering,
	}

	return lp.DB.Save(&newItem).Error
}

func (lp *LaunchPad) ApplyConfig(config Apps, groupID, rootParentID int) error {

	for _, page := range config.Pages {
		groupID++
		// create a new page
		err := lp.createNewPage(groupID, rootParentID, page.Number)
		if err != nil {
			return errors.Wrap(err, "createNewPage")
		}

		if page.Number == 1 {
			lp.rootPage = groupID
		}

		pageParentID := groupID

		for idx, item := range page.Items {
			switch item.(type) {
			case string:
				// add a flat item
				if err := lp.updateItem(item.(string), ApplicationType, pageParentID, idx); err != nil {
					return errors.Wrap(err, "updateItem")
				}
			default:
				var folder AppFolder
				if err := mapstructure.Decode(item, &folder); err != nil {
					return errors.Wrap(err, "mapstructure unable to decode config folder")
				}

				// create a new folder
				groupID++
				err := lp.createNewFolder(folder.Name, groupID, pageParentID, idx)
				if err != nil {
					return errors.Wrap(err, "createNewFolder")
				}

				folderParentID := groupID

				for _, fpage := range folder.Pages {
					// create a new folder page
					groupID++
					if err := lp.createNewFolderPage(groupID, folderParentID, fpage.Number); err != nil {
						return errors.Wrap(err, "createNewFolderPage")
					}

					// add all folder page items
					for fidx, fitem := range fpage.Items {
						if err := lp.updateItem(fitem, ApplicationType, groupID, fidx); err != nil {
							return errors.Wrap(err, "updateItem")
						}
					}
				}
			}
		}
	}

	return nil
}

// // ApplyConfig places all the launchpad apps
// func (lp *LaunchPad) ApplyConfig(config Apps, startingID, rootParentID int) error {

// 	var (
// 		rowID          int
// 		pageID         int
// 		pageParentID   int
// 		folderID       int
// 		folderParentID int
// 		folderPageID   int
// 	)

// 	rowID = startingID

// 	for _, page := range config.Pages {
// 		rowID++
// 		pageID = rowID

// 		// create a new page
// 		if err := lp.createNewPage(pageID, rootParentID, page.Number); err != nil {
// 			return fmt.Errorf("createNewPage failed: %w", err)
// 		}

// 		if page.Number == 1 { // flatten out apps to first page
// 			lp.rootPage = pageID
// 			// lp.FlattenApps()
// 		}

// 		pageParentID = pageID

// 		for idx, item := range page.Items {
// 			switch item.(type) {
// 			case string:
// 				// add a folder-less app to the current page
// 				if err := lp.updateItem(item.(string), ApplicationType, pageParentID, idx); err != nil {
// 					return fmt.Errorf("failed to update folder-less app '%s': %w", item, err)
// 				}
// 			default:
// 				var folder AppFolder
// 				if err := mapstructure.Decode(item, &folder); err != nil {
// 					return fmt.Errorf("mapstructure unable to decode config folder: %w", err)
// 				}

// 				// create a new folder
// 				rowID++
// 				folderID = rowID
// 				if err := lp.createNewFolder(folder.Name, folderID, pageParentID, idx); err != nil {
// 					return fmt.Errorf("failed to create folder '%s' with ID=%d: %w", folder.Name, folderID, err)
// 				}

// 				folderParentID = folderID

// 				for fpidx, fpage := range folder.Pages {
// 					rowID++
// 					folderPageID = rowID
// 					// create a new folder page
// 					if err := lp.createNewFolderPage(folderPageID, folderParentID, fpidx); err != nil {
// 						return fmt.Errorf("failed to create page #%d for folder '%s' failed: %w", fpage.Number, folder.Name, err)
// 					}
// 					// add all folder page items
// 					for fpiIDX, fpItem := range fpage.Items {
// 						if err := lp.updateItem(fpItem, ApplicationType, folderPageID, fpiIDX); err != nil {
// 							return fmt.Errorf("failed to update folder app '%s': %w", fpItem, err)
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}

// 	return nil
// }

func (lp *LaunchPad) addToFolder(appName, folderName string) error {
	var a App
	if err := lp.DB.Where("title = ?", appName).First(&a).Error; err != nil {
		return fmt.Errorf("app query failed for '%s': %w", appName, err)
	}
	var app Item
	if err := lp.DB.Where("rowid = ?", a.ID).First(&app).Error; err != nil {
		return fmt.Errorf("item query failed for app ID %d: %w", a.ID, err)
	}

	var g Group
	if err := lp.DB.Where("title = ?", folderName).First(&g).Error; err != nil {
		return fmt.Errorf("group query failed for '%s': %w", folderName, err)
	}
	var folder Item
	if err := lp.DB.Where("rowid = ?", g.ID).First(&folder).Error; err != nil {
		return fmt.Errorf("item query failed for folder ID %d: %w", a.ID, err)
	}
	var pages []Item
	if err := lp.DB.Where("parent_id = ?", folder.ID).Find(&pages).Error; err != nil {
		return fmt.Errorf("failed to find pages for folder '%s': %w", folderName, err)
	}
	if len(pages) == 0 {
		return fmt.Errorf("folder '%s' has no pages", folderName)
	}
	var folderApps []Item
	if err := lp.DB.Where("parent_id = ?", pages[0].ID).Find(&folderApps).Error; err != nil {
		return fmt.Errorf("failed to find apps for page '%d': %w", pages[0].ID, err)
	}

	if err := lp.updateItem(appName, ApplicationType, pages[0].ID, len(folderApps)); err != nil {
		return fmt.Errorf("failed to add app '%s' to folder '%s': %w", appName, folderName, err)
	}

	return nil
}

// FixOther moves all apps in the 'Other' group to the root page
func (lp *LaunchPad) FixOther() error {
	if slices.Contains(lp.confFolders, "Other") { // config contain Other folder (no need to fix)
		return nil
	}

	var other Group
	if err := lp.DB.Where("title = ?", "Other").Find(&other).Error; err != nil {
		return fmt.Errorf("failed to find group 'Other': %w", err)
	}
	if other.Title != "Other" {
		return fmt.Errorf("group 'Other' not found")
	}

	var pages []Item
	if err := lp.DB.Where("parent_id = ?", other.ID).Find(&pages).Error; err != nil {
		return fmt.Errorf("failed to find pages for group 'Other': %w", err)
	}

	var apps []App
	for _, page := range pages {
		var pageItems []Item
		if err := lp.DB.Where("parent_id = ?", page.ID).Find(&pageItems).Error; err != nil {
			return fmt.Errorf("failed to find apps for page '%s': %w", page.UUID, err)
		}
		for _, pageItem := range pageItems {
			lp.DB.Model(&pageItem).Association("App").Find(&pageItem.App)
			apps = append(apps, pageItem.App)
		}
	}

	// move apps to root page
	for _, app := range apps {
		utils.Indent(log.WithField("app", app.Title).Warn, 3)("moving app from Other folder")
		if cfolder, err := lp.Config.GetFolderContainingApp(app.Title); err == nil {
			if err := lp.addToFolder(app.Title, cfolder); err != nil { // add to folder it SHOULD have been in
				return err
			}
		} else {
			if err := lp.updateItem(app.Title, ApplicationType, lp.rootPage, -1); err != nil { // add to end of root page
				return fmt.Errorf("failed to move app '%s' from Other to root: %w", app.Title, err)
			}
		}

	}

	// remove all traces of Other
	for _, page := range pages {
		if err := lp.DB.Delete(&page).Error; err != nil {
			return fmt.Errorf("failed to delete page '%s': %w", page.UUID, err)
		}
		if err := lp.DB.Delete(&Group{ID: page.ID}).Error; err != nil {
			return fmt.Errorf("failed to delete group for page '%s': %w", page.UUID, err)
		}
	}
	if err := lp.DB.Delete(&other).Error; err != nil {
		return fmt.Errorf("failed to delete group 'Other': %w", err)
	}
	if err := lp.DB.Delete(&Item{ID: other.ID}).Error; err != nil {
		return fmt.Errorf("failed to delete item for 'Other': %w", err)
	}

	return nil
}

// EnableTriggers enables item update triggers
func (lp *LaunchPad) EnableTriggers() error {
	utils.Indent(log.Info, 2)("enabling SQL update triggers")
	if err := lp.DB.Exec("UPDATE dbinfo SET value=0 WHERE key='ignore_items_update_triggers';").Error; err != nil {
		return errors.Wrap(err, "counld not update `ignore_items_update_triggers` to 0")
	}
	return nil
}

// DisableTriggers disables item update triggers
func (lp *LaunchPad) DisableTriggers() error {
	utils.Indent(log.Info, 2)("disabling SQL update triggers")
	if err := lp.DB.Exec("UPDATE dbinfo SET value=1 WHERE key='ignore_items_update_triggers';").Error; err != nil {
		return errors.Wrap(err, "counld not update `ignore_items_update_triggers` to 1")
	}
	return nil
}

// TriggersDisabled returns true if triggers are disabled
func (lp *LaunchPad) TriggersDisabled() bool {
	var dbinfo DBInfo
	if err := lp.DB.Where("key in (?)", []string{"ignore_items_update_triggers"}).Find(&dbinfo).Error; err != nil {
		log.WithError(err).Error("dbinfo query failed")
	}
	if dbinfo.Value == "1" {
		return true
	}
	return false
}

// GetMaxAppID returns the maximum App ItemID
func (lp *LaunchPad) GetMaxAppID() int {
	var apps []App

	if err := lp.DB.Find(&apps).Error; err != nil {
		utils.Indent(log.WithError(err).Error, 2)("query all apps failed")
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
		utils.Indent(log.WithError(err).Error, 2)("query all widgets failed")
	}

	maxID := 0
	for _, widget := range widgets {
		if widget.ID > maxID {
			maxID = widget.ID
		}
	}

	return maxID
}
