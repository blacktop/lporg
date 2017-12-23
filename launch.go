package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// App CREATE TABLE apps (item_id INTEGER PRIMARY KEY, title VARCHAR, bundleid VARCHAR, storeid VARCHAR,category_id INTEGER, moddate REAL, bookmark BLOB)
type App struct {
	ItemID     int            `gorm:"column:item_id;primary_key"`
	Title      string         `gorm:"column:title"`
	BundleID   string         `gorm:"column:bundleid"`
	StoreID    sql.NullString `gorm:"column:storeid"`
	CategoryID sql.NullInt64  `gorm:"column:category_id"`
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
	ID         int            `gorm:"column:item_id;primary_key"`
	CategoryID sql.NullInt64  `gorm:"column:category_id"`
	Title      sql.NullString `gorm:"column:title"`
}

// Item CREATE TABLE items (rowid INTEGER PRIMARY KEY ASC, uuid VARCHAR, flags INTEGER, type INTEGER, parent_id INTEGER NOT NULL, ordering INTEGER)
type Item struct {
	RowID int `gorm:"column:rowid;primary_key"`
	App   App
	UUID  string        `gorm:"column:uuid"`
	Flags sql.NullInt64 `gorm:"column:flags"`
	Type  sql.NullInt64 `gorm:"column:type"`
	// ParentID Group         `db:"parent_id"`
	Group    Group         `gorm:"ForeignKey:ParentID"`
	ParentID int           `gorm:"not null;unique;column:parent_id"`
	Ordering sql.NullInt64 `gorm:"column:ordering"`
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	var item Item
	var group Group

	// Older macOS
	// $HOME/Library/Application\ Support/Dock/*.db

	// High Sierra
	// $TMPDIR../0/com.apple.dock.launchpad/db/db

	tmpDir := os.Getenv("TMPDIR")
	launchDB, err := filepath.Glob(tmpDir + "../0/com.apple.dock.launchpad/db/db")
	if err != nil {
		log.Fatal(err)
	}
	if len(launchDB) == 0 {
		log.Fatal(errors.New("launchpad DB not found"))
	}
	db, err := gorm.Open("sqlite3", launchDB[0])
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// db.LogMode(true)

	if err := db.First(&item, 144).Error; err != nil {
		log.Error(err)
	}

	db.Model(&item).Related(&item.App)
	db.Model(&item.App).Related(&item.App.Category)
	if err := db.First(&group, item.ParentID+1).Error; err != nil {
		log.Error(err)
	}
	// fmt.Println("--------------------------------------------------------")
	item.Group = group
	// fmt.Printf("item========================================> %+v\n", item)
	itemJSON, _ := json.Marshal(item)
	fmt.Println(string(itemJSON))
	// fmt.Printf("%+v\n", item.App)

	// db, err := sqlx.Connect("sqlite3", "./launchpad.db")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer db.Close()

	// checkError(db.Select(&apps, "SELECT * FROM apps ORDER BY item_id ASC"))
	// checkError(db.Select(&categories, "SELECT * FROM categories ORDER BY rowid ASC"))
	// checkError(db.Select(&groups, "SELECT * FROM groups ORDER BY item_id ASC"))
	// checkError(db.Select(&items, "SELECT * FROM items ORDER BY rowid ASC"))

	// for _, g := range apps {
	// 	if g.CategoryID.Valid {
	// 		fmt.Printf("%+v\n", g)
	// 	} else {
	// 		log.Error("we got some bad hombres")
	// 	}
	// }

	// for _, g := range categories {
	// 	fmt.Printf("%+v\n", g)
	// }

	// for _, g := range groups {
	// 	if g.Title.Valid {
	// 		fmt.Printf("%+v\n", g)
	// 	} else {
	// 		log.Error("we got some bad hombres")
	// 	}
	// }

	// for _, g := range items {
	// 	fmt.Printf("%+v\n", g)
	// }

	// rows, err := db.Queryx("select * from groups")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer rows.Close()

	// cols, err := rows.Columns()
	// fmt.Println(cols)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// for rows.Next() {
	// 	app := Group{}
	// 	err = rows.StructScan(&app)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Printf("%#v\n", app)
	// 	// fmt.Println(itemID, title, bundleid)
	// }
	// err = rows.Err()
	// if err != nil {
	// 	log.Fatal(err)
	// }
}
