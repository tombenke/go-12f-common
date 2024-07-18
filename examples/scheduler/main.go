package main

import (
	"github.com/tombenke/go-12f-common/apprun"
	"github.com/tombenke/go-12f-common/must"
)

func main() {

	// Make and run an application via ApplicationRunner
	must.Must(apprun.MakeAndRun(&Config{}, NewApplication))
}
