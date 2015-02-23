package main

import (
	"flag"
	"fmt"
	"github.com/julienschmidt/httprouter"
	_ "github.com/wilhelm-murdoch/biscuit"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	// Name defines the name of this utility
	Name = "biscuit-server"
	// Version defines this utility's current version
	Version = "1.0.0"
)

var (
	port    = flag.Int("p", 8001, "server port assignment")
	support = flag.Bool("s", false, "lists all supported bodies of text")
	version = flag.Bool("v", false, "current version of this server")
	load    = flag.String("l", "", "comma separated list of bodies to load (all by default)")
	dir     = flag.String("d", "./corpora/*.csv", "glob path pointing to stored tables")
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: biscuitserver [flags]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func init() {
	flag.Usage = usage
	flag.Parse()

	// Command line arguments will be read here.
	// Corpora for different languages will be read and parsed here before startup.
	// This is memory intensive, so, to increase startup times, I'm thinking a go
	// routine for each corpus of data that feeds into a map channel. Once the
	// number of specified bodies equals the number of entries in the map, we're done.
	// Essentially, block until all bodies are loaded into memory.

	if *version {
		fmt.Println(Name, Version)
		os.Exit(0)
	}

	if *support {
		bodies, err := getListOfSupportedBodies(*dir)
		if err != nil {
			fmt.Println("None found ... Maybe check your path?")
			os.Exit(1)
		}

		if len(bodies) == 0 {
			fmt.Println("None found ... Maybe check your path?")
		} else {
			fmt.Print(len(bodies))
			fmt.Println(" Found:")
			fmt.Println("- " + strings.Join(bodies, "\t\n- "))
		}

		os.Exit(0)
	}

	if *load != "" {
		bodies, err := getListOfSupportedBodies(*dir)
		if err != nil {
			fmt.Printf("could not load bodies from path `%s`\n", *dir)
			os.Exit(1)
		}

		for _, body := range strings.Split(*load, ",") {
			if strings.TrimSpace(body) != "" && indexOfStringSlice(body, bodies) == -1 {
				fmt.Printf("`%s` not a supported body\n", body)
				os.Exit(1)
			}
		}
	}
}

func indexOfStringSlice(value string, slice []string) int {
	for p, v := range slice {
		if v == value {
			return p
		}
	}
	return -1
}

func getListOfSupportedBodies(path string) ([]string, error) {
	samples, _ := filepath.Glob(path)
	bodies := []string{}

	for _, file := range samples {
		bodies = append(bodies, filepath.Base(file)[:2])
	}

	return bodies, nil
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	bodies, err := getListOfSupportedBodies(*dir)
	if err != nil {
		fmt.Fprint(w, "could not load bodies from path `%s`\n", *dir)
		return
	}

	fmt.Fprint(w, "Supported bodies: "+strings.Join(bodies, ", "))
}

func main() {
	router := httprouter.New()
	router.GET("/", Index)

	log.Printf("running %s v.%s on %d", Name, Version, *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), router))
}
