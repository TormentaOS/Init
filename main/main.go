package main

import (
	"fmt"
	"initializer/core"
	"os"
)

func main() {
	manager, err := core.NewServiceManager(os.Args[1])

	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	//manager.ConnectSignals()
	manager.Start()
}
