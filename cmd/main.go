package main

import "github.com/pufferpanel/apufferi/v3/logging"

func main() {
	defer logging.Close()

	if err := Execute(); err != nil {
		logging.Exception("Error running process", err)
	}
}
