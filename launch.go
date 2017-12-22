package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// App CREATE TABLE apps (item_id INTEGER PRIMARY KEY, title VARCHAR, bundleid VARCHAR, storeid VARCHAR,category_id INTEGER, moddate REAL, bookmark BLOB)
type App struct {
	ItemID     int            `db:"item_id"`
	Title      string         `db:"title"`
	BundleID   string         `db:"bundleid"`
	StoreID    sql.NullString `db:"storeid"`
	CategoryID sql.NullInt64  `db:"category_id"`
	Moddate    float64        `db:"moddate"`
	Bookmark   []byte         `db:"bookmark"`
}

// Categorie CREATE TABLE categories (rowid INTEGER PRIMARY KEY ASC, uti VARCHAR)
type Categorie struct {
	RowID int    `db:"rowid"`
	UTI   string `db:"uti"`
}

// Group CREATE TABLE groups (item_id INTEGER PRIMARY KEY, category_id INTEGER, title VARCHAR)
type Group struct {
	ItemID     int            `db:"item_id"`
	CategoryID sql.NullInt64  `db:"category_id"`
	Title      sql.NullString `db:"title"`
}

// Item CREATE TABLE items (rowid INTEGER PRIMARY KEY ASC, uuid VARCHAR, flags INTEGER, type INTEGER, parent_id INTEGER NOT NULL, ordering INTEGER)
type Item struct {
	RowID    int           `db:"rowid"`
	UUID     string        `db:"uuid"`
	Flags    sql.NullInt64 `db:"flags"`
	Type     sql.NullInt64 `db:"type"`
	ParentID int           `db:"parent_id"`
	Ordering sql.NullInt64 `db:"ordering"`
}

func main() {
	// $TMPDIR../0/com.apple.dock.launchpad/db/db
	db, err := sqlx.Connect("sqlite3", "./db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Queryx("select * from groups")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	fmt.Println(cols)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		app := Group{}
		err = rows.StructScan(&app)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%#v\n", app)
		// fmt.Println(itemID, title, bundleid)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}
