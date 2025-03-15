package main

// ONLY FOR TESTING PURPOSES: this code should be inside kops

import (
	"fmt"
	"github.com/Elemento-Modular-Cloud/tesi-paolobeci/ecloud"
)

func main() {
	client, _ := ecloud.NewClient(
		"ovh-eu",
		"YOUR_APPLICATION_KEY",
		"YOUR_APPLICATION_SECRET",
		"YOUR_CONSUMER_KEY",
	)
	res, err := client.Get("/me")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Welcome %s!\n", res)
}