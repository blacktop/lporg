package main

import (
	"crypto/rand"
	"fmt"
	"io"

	"github.com/apex/log"
	"github.com/pkg/errors"
)

// Types
const (
	Root = iota
	FolderRoot
	Page
	Application
	DownloadingApp
	Widget
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
	CategoryID int    `gorm:"column:category_id;default:null"`
	Title      string `gorm:"column:title;default:null"`
}

// Item - CREATE TABLE items (rowid INTEGER PRIMARY KEY ASC, uuid VARCHAR, flags INTEGER, type INTEGER, parent_id INTEGER NOT NULL, ordering INTEGER)
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

// DBInfo - CREATE TABLE dbinfo (key VARCHAR, value VARCHAR)
type DBInfo struct {
	Key   string
	Value string
}

// TableName set DBInfo's table name to be `dbinfo`
func (DBInfo) TableName() string {
	return "dbinfo"
}

// ClearGroups clears out items related to groups
func (lp *LaunchPad) ClearGroups() error {
	log.Info("clear out groups")

	var items []Item
	return lp.DB.Where("type in (?)", []int{Root, FolderRoot, Page}).Delete(&items).Error
}

// AddRootsAndHoldingPages adds back in the RootPage and HoldingPage defaults
func (lp *LaunchPad) AddRootsAndHoldingPages() error {
	log.Info("add root and holding pages")

	items := []Item{
		Item{RowID: 1, UUID: "ROOTPAGE", Type: Root, ParentID: 0},
		Item{RowID: 2, UUID: "HOLDINGPAGE", Type: Page, ParentID: 1},
		Item{RowID: 3, UUID: "ROOTPAGE_DB", Type: Root, ParentID: 0},
		Item{RowID: 4, UUID: "HOLDINGPAGE_DB", Type: Page, ParentID: 3},
		Item{RowID: 5, UUID: "ROOTPAGE_VERS", Type: Root, ParentID: 0},
		Item{RowID: 6, UUID: "HOLDINGPAGE_VERS", Type: Page, ParentID: 5},
	}

	for item := range items {
		// if success := lp.DB.NewRecord(item); success == false {
		// 	log.Error("create new record failed")
		// }
		if err := lp.DB.Create(&item).Error; err != nil {
			return errors.Wrap(err, "db insert failed")
		}
	}

	return nil
}

// CreateAppFolders creates all the launchpad app folders
//    group_id = setup_items(conn, Types.APP, app_layout, app_mapping, group_id, root_parent_id=1)
func (lp *LaunchPad) CreateAppFolders(config map[string][]map[string][]string, groupID int) error {
	log.Infof(bold, "creating app folders and adding apps to them")

	// for index, page := range config["pages"] {
	// 	// Start a new page (note that the ordering starts at 1 instead of 0 as there is a holding page at an ordering of 0)
	// 	groupID++

	// 	item := Item{
	// 		RowID:    groupID,
	// 		UUID:     newUUID(),
	// 		Flags:    Page,
	// 		ParentID: 1,
	// 		Ordering: index + 1,
	// 	}

	// 	if success := lp.DB.NewRecord(item); success == false {
	// 		log.Error("create new record failed")
	// 	}
	// 	if err := lp.DB.Create(&item).Error; err != nil {
	// 		return err
	// 	}

	// group := Group{ID: groupID} // omitting fields makes them null

	// 	if success := lp.DB.NewRecord(group); success == false {
	// 		log.Error("create new record failed")
	// 	}
	// 	if err := lp.DB.Create(&group).Error; err != nil {
	// 		return err
	// 	}
	// 	// Capture the group id of the page to be used for child items
	// 	pageParentID := groupID

	// 	// Iterate through items
	// 	itemOrdering := 0
	// 	for folder := range page {
	// 		// Start a new folder
	// 		groupID++

	// 		item := Item{
	// 			RowID:    groupID,
	// 			UUID:     newUUID(),
	// 			Flags:    FolderRoot,
	// 			ParentID: pageParentID,
	// 			Ordering: itemOrdering,
	// 		}

	// 		if success := lp.DB.NewRecord(item); success == false {
	// 			log.Error("create new record failed")
	// 		}
	// 		if err := lp.DB.Create(&item).Error; err != nil {
	// 			return err
	// 		}

	// 		group := Group{
	// 			ID:         groupID,
	// 			CategoryID: 0,
	// 			Title:      FolderName,
	// 		}

	// 		if success := lp.DB.NewRecord(group); success == false {
	// 			log.Error("create new record failed")
	// 		}
	// 		if err := lp.DB.Create(&group).Error; err != nil {
	// 			return err
	// 		}

	// 		itemOrdering++
	// 	}
	// }

	return nil
}

// DisableTriggers disables item update triggers
func (lp *LaunchPad) DisableTriggers() error {

	log.Info("disabling SQL update triggers")

	if err := lp.DB.Exec("UPDATE dbinfo SET value = 1 WHERE key = 'ignore_items_update_triggers';").Error; err != nil {
		return errors.Wrap(err, "counld not update `ignore_items_update_triggers` to 1")
	}

	return nil
}

// EnableTriggers enables item update triggers
func (lp *LaunchPad) EnableTriggers() error {

	log.Info("enabling SQL update triggers")

	if err := lp.DB.Exec("UPDATE dbinfo SET value = 0 WHERE key = 'ignore_items_update_triggers';").Error; err != nil {
		return errors.Wrap(err, "counld not update `ignore_items_update_triggers` to 0")
	}

	return nil
}

func (lp *LaunchPad) getMaxAppID() int {
	var apps []App

	if err := lp.DB.Find(&apps).Error; err != nil {
		log.WithError(err).Error("query all apps failed")
	}

	maxID := 0
	for _, app := range apps {
		if app.ItemID > maxID {
			maxID = app.ItemID
		}
	}

	return maxID
}

func (lp *LaunchPad) getMaxWidgetID() int {
	// var apps []App

	// if err := lp.DB.Find(&apps).Error; err != nil {
	// 	log.WithError(err).Error("query all apps failed")
	// }

	// maxID := 0
	// for _, app := range apps {
	// 	if app.ItemID > maxID {
	// 		maxID = app.ItemID
	// 	}
	// }

	// return maxID
	return 0
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
