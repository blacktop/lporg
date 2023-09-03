package database

import (
	"os"

	"github.com/apex/log"
	"github.com/blacktop/lporg/internal/utils"
	yaml "gopkg.in/yaml.v3"
)

// Config is the Launchpad config
type Config struct {
	Apps    Apps    `yaml:"apps" json:"apps,omitempty"`
	Widgets Apps    `yaml:"widgets" json:"widgets,omitempty"`
	Dock    Dock    `yaml:"dock_items" json:"dock_items,omitempty"  mapstructure:"dock_items"`
	Desktop Desktop `yaml:"desktop" json:"desktop,omitempty"  mapstructure:"desktop"`
}

// Apps is the launchpad apps config object
type Apps struct {
	Pages []Page `yaml:"pages" json:"pages,omitempty"`
}

// Page is a launchpad page object
type Page struct {
	Number int           `yaml:"number" json:"number"`
	Items  []interface{} `yaml:"items,omitempty" json:"items,omitempty"`
}

// AppFolder is a launchpad folder object
type AppFolder struct {
	Name  string       `yaml:"folder,omitempty" json:"folder,omitempty" mapstructure:"folder"`
	Pages []FolderPage `yaml:"pages,omitempty" json:"pages,omitempty"`
}

// FolderPage is a launchpad folder page object
type FolderPage struct {
	Number int      `yaml:"number,omitempty" json:"number"`
	Items  []string `yaml:"items,omitempty" json:"items,omitempty"`
}

// Desktop is the desktop object
type Desktop struct {
	Image string `yaml:"image,omitempty" json:"image,omitempty"`
}

// Dock is the launchpad dock config object
type Dock struct {
	Apps   []string `yaml:"apps,omitempty" json:"apps,omitempty"`
	Others []string `yaml:"others,omitempty" json:"others,omitempty"`
}

// LoadConfig loads the Launchpad config from the config file
func LoadConfig(filename string) (Config, error) {
	var conf Config

	utils.Indent(log.WithField("path", filename).Info)("parsing launchpad config YAML")
	data, err := os.ReadFile(filename)
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
