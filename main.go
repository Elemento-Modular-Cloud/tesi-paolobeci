package main

// TEST FILE

import (
	"fmt"
	"github.com/Elemento-Modular-Cloud/tesi-paolobeci/ecloud"
)

func main() {
	client, err := ecloud.NewClient(
		"ovh-eu",
		"YOUR_APPLICATION_KEY",
		"YOUR_APPLICATION_SECRET",
		"YOUR_CONSUMER_KEY",
	)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Welcome %s!\n", client)
}