package main

import "github.com/pufferpanel/apufferi/v3/logging"

func main() {
	defer logging.Close()

	Execute()
}
