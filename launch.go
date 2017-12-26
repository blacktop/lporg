package main

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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
	ParentID int   `gorm:"not null;unique;column:parent_id"`
	Ordering int   `gorm:"column:ordering"`
}

type AppGroups struct {
	Groups map[string][]App
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func writeTomlFile(filename string, data interface{}) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	err = toml.NewEncoder(f).Encode(data)

	return err
}

func init() { log.SetLevel(log.DebugLevel) }

func main() {

	var items []Item
	var group Group
	appGroups := make(map[string][]App)

	// Older macOS
	// $HOME/Library/Application\ Support/Dock/*.db

	// High Sierra
	// $TMPDIR../0/com.apple.dock.launchpad/db/db

	// find launchpad database
	tmpDir := os.Getenv("TMPDIR")
	launchDB, err := filepath.Glob(tmpDir + "../0/com.apple.dock.launchpad/db/db")
	if err != nil {
		log.Fatal(err)
	}
	if len(launchDB) == 0 {
		log.Fatal(errors.New("launchpad DB not found"))
	}
	log.Infoln("Found Launchpad DB: ", launchDB[0])
	// open launchpad database
	db, err := gorm.Open("sqlite3", launchDB[0])
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// db.LogMode(true)

	if err := db.Where("type = ?", "4").Find(&items).Error; err != nil {
		log.Error(err)
	}

	for _, item := range items {
		group = Group{}
		db.Model(&item).Related(&item.App)
		db.Model(&item.App).Related(&item.App.Category)
		log.Debugln("item.ParentID-1 = ", item.ParentID-1)
		if err := db.First(&group, item.ParentID-1).Error; err != nil {
			log.Error(err)
		}
		item.Group = group
		log.Debugf("%+v\n", group)
		if len(group.Title) > 0 {
			appGroups[group.Title] = append(appGroups[group.Title], item.App)
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
	checkError(writeTomlFile("./groups.toml", appGroups))
}
