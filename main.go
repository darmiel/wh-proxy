package main

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
)

var wids = make(map[string][]A)

func main() {
	log.SetHandler(cli.Default)

	f, err := ParseFile("examples/test.yaml")
	if err != nil {
		log.WithError(err).Error("cannot parse file")
		return
	}

	wids[f.ID] = []A{*f}

	if err = Serve(); err != nil {
		log.WithError(err).Error("cannot serve")
	}
}