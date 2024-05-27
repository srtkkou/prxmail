package main

import (
	"os"

	app "github.com/srtkkou/prxmail"
)

var (
	// GITのコミットID
	Revision string
)

func main() {
	os.Exit(app.AppMain(Revision))
}
