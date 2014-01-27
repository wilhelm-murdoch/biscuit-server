package main

import (
	"github.com/codegangsta/martini"
	_ "github.com/wilhelm-murdoch/biscuit"
)

func init() {
	// Command line arguments will be read here.
	// Corpora for different languages will be read and parsed here before startup.
	// This is memory intensive, so, to increase startup times, I'm thinking a go
	// routine for each corpus of data that feeds into a map channel. Once the
	// number of specified bodies equals the number of entries in the map, we're done.
	// Essentially, block until all bodies are loaded into memory.
}

func main() {
	m := martini.Classic()
	m.Get("/", func() string {
		return "A form for manual entry will go here ..."
	})
	m.Post("/", func() string {
		return "JSON response containing scores, best match and processing time."
	})
	m.Run()
}
