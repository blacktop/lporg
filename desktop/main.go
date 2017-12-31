package main

import (
	"fmt"
	"log"
	"os/user"
	"path/filepath"

	"github.com/blacktop/lporg/desktop/background"
)

func main() {
	// get current user
	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	out, err := background.SetDesktopImage(filepath.Join(user.HomeDir, "Downloads/Pics/milo.jpg"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(out)
}
