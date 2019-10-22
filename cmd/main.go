package main

import "github.com/pufferpanel/apufferi/v4/logging"

func main() {
	defer logging.Close()

	Execute()
}
