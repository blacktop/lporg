package dock

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	plist "github.com/DHowett/go-plist"
)

// Plist is a dock plist object
type Plist struct {
	PersistentApps              []PAItem `plist:"persistent-apps"`
	PersistentOthers            []POItem `plist:"persistent-others"`
	AutoHide                    bool     `plist:"autohide"`
	Magnification               bool     `plist:"magnification"`
	MinimizeToApplication       bool     `plist:"minimize-to-application"`
	LastMessagetraceStamp       float64  `plist:"last-messagetrace-stamp"`
	ModCount                    int      `plist:"mod-count"`
	Orientation                 string   `plist:"orientation"`
	ShowAppExposeGestureEnabled bool     `plist:"showAppExposeGestureEnabled"`
	TileSize                    float64  `plist:"tilesize"`
	TrashFull                   bool     `plist:"trash-full"`
	Version                     int      `plist:"version"`
	WvousBlCorner               int      `plist:"wvous-bl-corner"`
	WvousBlModifier             int      `plist:"wvous-bl-modifier"`
	WvousTlCorner               int      `plist:"wvous-tl-corner"`
	WvousTlModifier             int      `plist:"wvous-tl-modifier"`
	WvousTrCorner               int      `plist:"wvous-tr-corner"`
	WvousTrModifier             int      `plist:"wvous-tr-modifier"`
}

// PAItem is a dock plist persistent-apps item object
type PAItem struct {
	GUID     int      `plist:"GUID"`
	TileType string   `plist:"tile-type"`
	TileData TileData `plist:"tile-data"`
}

// POItem is a dock plist persistent-others item object
type POItem struct {
	GUID     int        `plist:"GUID"`
	TileType string     `plist:"tile-type"`
	TileData POTileData `plist:"tile-data"`
}

// POTileData is a persistent-others item title-data object
type POTileData struct {
	ShowAs            int      `plist:"showas"`
	FileType          int      `plist:"file-type"`
	ParentModDate     int      `plist:"parent-mod-date"`
	Book              []byte   `plist:"book"`
	FileData          FileData `plist:"file-data"`
	DisplayAs         int      `plist:"displayas"`
	FileLabel         string   `plist:"file-label"`
	FileModDate       int      `plist:"file-mod-date"`
	Arrangement       int      `plist:"arrangement"`
	PreferredItemSize int      `plist:"preferreditemsize"`
}

// TileData is a item title-data object
type TileData struct {
	DockExtra        bool     `plist:"dock-extra "`
	ParentModDate    int      `plist:"parent-mod-date"`
	FileType         int      `plist:"file-type"`
	Book             []byte   `plist:"book"`
	FileData         FileData `plist:"file-data"`
	FileLabel        string   `plist:"file-label"`
	FileModDate      int      `plist:"file-mod-date"`
	BundleIdentifier string   `plist:"bundle-identifier"`
}

// FileData is a tile-data file-data object
type FileData struct {
	URLString     string `plist:"_CFURLString"`
	URLStringType int    `plist:"_CFURLStringType"`
}

// LoadDockPlist loads the dock plist into struct
func LoadDockPlist() (Plist, error) {

	var dPlist Plist

	// get current user
	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	// read users dock plist
	pfile, err := os.Open(filepath.Join(user.HomeDir, "/Library/Preferences/com.apple.dock.plist"))
	if err != nil {
		return Plist{}, err
	}
	defer pfile.Close()

	// decode plist into struct
	decoder := plist.NewDecoder(pfile)
	err = decoder.Decode(&dPlist)
	if err != nil {
		return Plist{}, err
	}

	return dPlist, nil
}

func main() {

	dPlist, err := LoadDockPlist()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("===> Your Current Dock Layout:\n\n")
	for _, item := range dPlist.PersistentApps {
		fmt.Printf("%#v\n", item.TileData.FileLabel)
	}
	fmt.Println("=======================")
	for _, item := range dPlist.PersistentOthers {
		fmt.Printf("%#v\n", item.TileData.FileLabel)
	}

	// plist, err := plist.MarshalIndent(dPlist, plist.XMLFormat, "\t")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(string(plist))
}
