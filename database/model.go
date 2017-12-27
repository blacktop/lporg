package database

import "github.com/jinzhu/gorm"

// Types
const (
	RootType = iota
	FolderRootType
	PageType
	ApplicationType
	DownloadingAppType
	WidgetType
)

// LaunchPad is a LaunchPad struct
type LaunchPad struct {
	DB     *gorm.DB
	File   string
	Folder string
}

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

// Widget - CREATE TABLE widgets (item_id INTEGER PRIMARY KEY, title VARCHAR, bundleid VARCHAR, storeid VARCHAR,category_id INTEGER, moddate REAL, bookmark BLOB)
type Widget struct {
	ItemID     int    `gorm:"column:item_id;primary_key"`
	Title      string `gorm:"column:title"`
	BundleID   string `gorm:"column:bundleid"`
	StoreID    string `gorm:"column:storeid"`
	CategoryID int    `gorm:"column:category_id"`
	Category   Category
	Moddate    float64 `gorm:"column:moddate"`
	Bookmark   []byte  `gorm:"column:bookmark"`
}

// TableName set DBInfo's table name to be `dbinfo`
func (DBInfo) TableName() string {
	return "dbinfo"
}
