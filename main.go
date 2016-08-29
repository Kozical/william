// +build windows

package main

import (
	"flag"
	"fmt"

	"github.com/Kozical/william/app"
)

func main() {
	path := flag.String("config", "config.yaml", "Specify the path to the configuration yaml file. (ie: C:\\my\\path\\config.yaml) (default: config.yaml)")
	flag.Parse()

	config, err := app.Init(*path)

	if err != nil {
		panic(err)
	}

	manager, err := app.New(config)
	if err != nil {
		panic(err)
	}

	defer manager.Close()
	if err := manager.Run(); err != nil {
		fmt.Printf("Error occurred while running Manager -> %s\n", err)
	}
}
