package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
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
	FlatItems []string `yaml:"flat_items" json:"flat_items,omitempty"`
	Folders   []Folder `yaml:"folders" json:"folders,omitempty"`
}

// Folder is a launchpad folder object
type Folder struct {
	Name  string       `yaml:"name" json:"name,omitempty"`
	Pages []FolderPage `yaml:"pages" json:"pages,omitempty"`
}

// FolderPage is a launchpad folder page object
type FolderPage struct {
	Number int      `yaml:"number" json:"number"`
	Items  []string `yaml:"items" json:"items,omitempty"`
}

func testYAML() {
	var config Config

	data, err := ioutil.ReadFile("launchpad.yaml")
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	// fmt.Printf("--- t:\n%v\n\n", config)
	// try JSON too
	configJSON, _ := json.Marshal(config)
	fmt.Println(string(configJSON))
}

// func main() {
// 	testYAML()
// }
