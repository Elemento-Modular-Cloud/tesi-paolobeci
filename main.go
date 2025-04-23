package main

// TEST FILE

import (
	"fmt"
	"github.com/Elemento-Modular-Cloud/tesi-paolobeci/ecloud"
)

func main() {
	client, err := ecloud.NewClient(
		"APP-NAME",
		"APP-VERSION",
	)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Welcome %s!\n", client)
}