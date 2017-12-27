package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

func main() {
	var pages map[string][]map[string][]string

	data, err := ioutil.ReadFile("launchpad.yaml")
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(data, &pages)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	// fmt.Printf("--- t:\n%v\n\n", pages)
	itemJSON, _ := json.Marshal(pages)
	fmt.Println(string(itemJSON))
}
