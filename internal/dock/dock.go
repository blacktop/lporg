// Package dock provides functions for manipulating the macOS dock
package dock

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/blacktop/lporg/internal/database"
	"github.com/blacktop/lporg/internal/utils"
	"howett.net/plist"
)

const dockPlistPath = "/Library/Preferences/com.apple.dock.plist"

// Plist is a dock plist object
type Plist struct {
	PersistentApps              []PAItem `plist:"persistent-apps"`
	PersistentOthers            []POItem `plist:"persistent-others"`
	AutoHide                    bool     `plist:"autohide"`
	Largesize                   any      `plist:"largesize"`
	Loc                         string   `plist:"loc"`
	Magnification               bool     `plist:"magnification"`
	MinimizeToApplication       bool     `plist:"minimize-to-application"`
	LastMessagetraceStamp       float64  `plist:"last-messagetrace-stamp"`
	LastShowIndicatorTime       float64  `plist:"lastShowIndicatorTime"`
	ModCount                    int      `plist:"mod-count"`
	MruSpaces                   bool     `plist:"mru-spaces"`
	Orientation                 string   `plist:"orientation"`
	RecentApps                  []any    `plist:"recent-apps"`
	Region                      string   `plist:"region"`
	ShowRecents                 bool     `plist:"show-recents"`
	ShowAppExposeGestureEnabled bool     `plist:"showAppExposeGestureEnabled"`
	TileSize                    any      `plist:"tilesize"`
	TrashFull                   bool     `plist:"trash-full"`
	Version                     int      `plist:"version"`
	WvousBlCorner               int      `plist:"wvous-bl-corner,omitempty"`
	WvousBlModifier             int      `plist:"wvous-bl-modifier,omitempty"`
	WvousTlCorner               int      `plist:"wvous-tl-corner,omitempty"`
	WvousTlModifier             int      `plist:"wvous-tl-modifier,omitempty"`
	WvousTrCorner               int      `plist:"wvous-tr-corner,omitempty"`
	WvousTrModifier             int      `plist:"wvous-tr-modifier,omitempty"`
}

// FileData is a tile-data file-data object
type FileData struct {
	URLString     string `plist:"_CFURLString"`
	URLStringType int    `plist:"_CFURLStringType"`
}

// TileData is a item title-data object
type TileData struct {
	BundleIdentifier string   `plist:"bundle-identifier,omitempty"`
	Book             []byte   `plist:"book,omitempty"`
	DockExtra        bool     `plist:"dock-extra,omitempty"`
	FileData         FileData `plist:"file-data"`
	FileLabel        string   `plist:"file-label"`
	FileModDate      int64    `plist:"file-mod-date,omitempty"`
	FileType         int      `plist:"file-type"`
	IsBeta           bool     `plist:"is-beta,omitempty"`
	ParentModDate    int64    `plist:"parent-mod-date,omitempty"`
}

func (d TileData) GetPath() string {
	out := strings.TrimPrefix(d.FileData.URLString, "file://")
	out = strings.TrimSuffix(out, "/")
	return strings.Replace(out, "%20", " ", -1)
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
	Arrangement       int      `plist:"arrangement"`
	DisplayAs         int      `plist:"displayas"`
	ShowAs            int      `plist:"showas"`
	FileData          FileData `plist:"file-data"`
	FileLabel         string   `plist:"file-label"`
	FileType          int      `plist:"file-type"`
	FileModDate       int64    `plist:"file-mod-date,omitempty"`
	IsBeta            bool     `plist:"is-beta,omitempty"`
	ParentModDate     int64    `plist:"parent-mod-date,omitempty"`
	PreferredItemSize int      `plist:"preferreditemsize,omitempty"`
	Book              []byte   `plist:"book,omitempty"`
	Directory         int      `plist:"directory,omitempty"`
}

func (d POTileData) GetPath() string {
	out := strings.TrimPrefix(d.FileData.URLString, "file://")
	out = strings.TrimSuffix(out, "/")
	return strings.Replace(out, "%20", " ", -1)
}

func fileNameWithoutExtTrimSuffix(fileName string) string {
	fileName = filepath.Base(fileName)
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

// LoadDockPlist loads the dock plist into struct
func LoadDockPlist(path ...string) (*Plist, error) {
	var dpath string
	if len(path) > 0 {
		dpath = path[0]
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %v", err)
		}
		dpath = filepath.Join(home, dockPlistPath)
	}

	// read users dock plist
	pfile, err := os.Open(dpath)
	if err != nil {
		return nil, err
	}
	defer pfile.Close()

	// decode plist into struct
	var dPlist Plist
	if err := plist.NewDecoder(pfile).Decode(&dPlist); err != nil {
		return nil, err
	}

	return &dPlist, nil
}

// AddApp adds an app to the dock plist
func (p *Plist) AddApp(appPath string) error {

	papp := PAItem{
		GUID:     rand.Intn(9999999999),
		TileType: "file-tile",
		TileData: TileData{
			FileData: FileData{
				URLString:     appPath,
				URLStringType: 0,
			},
			FileLabel: fileNameWithoutExtTrimSuffix(appPath),
			FileType:  41,
		},
	}

	p.PersistentApps = append(p.PersistentApps, papp)

	return nil
}

// AddOther adds an other to the dock plist
func (p *Plist) AddOther(other database.Folder) error {
	pother := POItem{
		GUID:     rand.Intn(9999999999),
		TileType: "directory-tile",
		TileData: POTileData{
			Directory:   1,
			Arrangement: int(other.Sort),
			DisplayAs:   int(other.Display),
			ShowAs:      int(other.View),
			FileData: FileData{
				URLString:     other.Path,
				URLStringType: 0,
			},
			FileLabel: fileNameWithoutExtTrimSuffix(other.Path),
			FileType:  2,
		},
	}

	p.PersistentOthers = append(p.PersistentOthers, pother)

	return nil
}

// Save saves the dock plist from struct
func (p *Plist) Save() error {

	p.ModCount++

	// backup previous plist
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}
	pfile, err := os.Open(filepath.Join(home, dockPlistPath))
	if err != nil {
		return fmt.Errorf("failed to open plist: %w", err)
	}
	defer pfile.Close()
	bak, err := os.Create(filepath.Join(home, dockPlistPath) + ".bak")
	if err != nil {
		return fmt.Errorf("failed to create backup plist: %w", err)
	}
	defer bak.Close()
	if _, err := io.Copy(bak, pfile); err != nil {
		return fmt.Errorf("failed to backup plist: %w", err)
	}

	// write dock plist to temp file
	tmp, err := os.CreateTemp("", "dock.plist")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	utils.DoubleIndent(log.WithField("plist", tmp.Name()).Info)("writing temp dock plist")
	if err := plist.NewBinaryEncoder(tmp).Encode(p); err != nil {
		return fmt.Errorf("failed to decode plist: %w", err)
	}
	tmp.Close()

	// import plist and restart dock
	if err := p.importPlist(tmp.Name()); err != nil {
		return fmt.Errorf("failed to import plist: %w", err)
	}
	return p.kickstart()
}

func (p *Plist) importPlist(path string) error {
	utils.DoubleIndent(log.Info)("importing dock plist")
	out, err := utils.RunCommand(context.Background(), "/usr/bin/defaults", "import", "com.apple.dock", path)
	if err != nil {
		return fmt.Errorf("failed to defaults import dock plist '%s': %v", path, err)
	}
	fmt.Println(out)
	return nil
}

func (p *Plist) kickstart() error {
	utils.DoubleIndent(log.Info)("restarting com.apple.Dock.agent service")
	out, err := utils.RunCommand(context.Background(), "/bin/launchctl", "kickstart", "-k", fmt.Sprintf("gui/%d/com.apple.Dock.agent", os.Getuid()))
	if err != nil {
		return fmt.Errorf("failed to kickstart dock: %v", err)
	}
	fmt.Println(out)
	return nil
}
