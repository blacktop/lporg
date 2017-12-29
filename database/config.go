package database

import (
	"io/ioutil"

	"github.com/apex/log"
	"github.com/blacktop/lporg/database/utils"
	yaml "gopkg.in/yaml.v2"
)

// Config is the Launchpad config
type Config struct {
	Apps    Apps `yaml:"apps" json:"apps,omitempty"`
	Widgets Apps `yaml:"widgets" json:"widgets,omitempty"`
}

// Apps is the launchpad apps config object
type Apps struct {
	Pages []Page `yaml:"pages" json:"pages,omitempty"`
}

// Page is a launchpad page object
type Page struct {
	Number    int      `yaml:"number" json:"number"`
	FlatItems []string `yaml:"flat_items,omitempty" json:"flat_items,omitempty"`
	Folders   []Folder `yaml:"folders,omitempty" json:"folders,omitempty"`
}

// Folder is a launchpad folder object
type Folder struct {
	Name  string       `yaml:"name,omitempty" json:"name,omitempty"`
	Pages []FolderPage `yaml:"pages,omitempty" json:"pages,omitempty"`
}

// FolderPage is a launchpad folder page object
type FolderPage struct {
	Number int      `yaml:"number,omitempty" json:"number"`
	Items  []string `yaml:"items,omitempty" json:"items,omitempty"`
}

// LoadConfig loads the Launchpad config from the config file
func LoadConfig(filename string) (Config, error) {
	var conf Config

	utils.Indent(log.WithField("path", filename).Info)("parsing launchpad config YAML")
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		utils.DoubleIndent(log.WithError(err).WithField("path", filename).Fatal)("config file not found")
		return conf, err
	}

	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		utils.DoubleIndent(log.WithError(err).WithField("path", filename).Fatal)("unmarshalling yaml failed")
		return conf, err
	}

	return conf, nil
}
