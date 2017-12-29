package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/mitchellh/mapstructure"
	yaml "gopkg.in/yaml.v2"
)

// Config is the Launchpad config
type Config struct {
	Apps      Apps     `yaml:"apps" json:"apps,omitempty"`
	Widgets   Apps     `yaml:"widgets" json:"widgets,omitempty"`
	DockItems []string `yaml:"dock_items" json:"dock_items,omitempty"`
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

func testYAML() {
	var config Config

	data, err := ioutil.ReadFile("launchpad-test.yaml")
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t:\n%#v\n\n", config)

	for _, page := range config.Apps.Pages {
		for _, item := range page.Items {
			switch i := item.(type) {
			case string:
				fmt.Println("string", i)
			default:
				fmt.Printf("--- t:\n%#v\n\n", item)
				fmt.Println(i)
				var result AppFolder
				err = mapstructure.Decode(item, &result)
				if err != nil {
					log.Fatal(err)
				}

				fmt.Printf("%#v", result)
			}
		}
	}

	// // try JSON too
	// configJSON, err := json.Marshal(config)
	// if err != nil {
	// 	log.Fatalf("error: %v", err)
	// }
	// fmt.Println(string(configJSON))
}

func main() {
	testYAML()
}
