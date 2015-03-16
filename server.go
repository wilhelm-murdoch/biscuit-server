package main

import (
	"flag"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/wilhelm-murdoch/biscuit"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

const (
	// Name defines the name of this utility
	Name = "biscuit-server"
	// Version defines this utility's current version
	Version = "1.0.0"
)

var (
	port    = flag.Int("p", 8001, "server port assignment")
	version = flag.Bool("v", false, "current version of this server")
	dir     = flag.String("d", "./corpora/*.csv", "glob path pointing to stored tables")
	bodies  = []string{}
	models  = make(map[string]*biscuit.Model)
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: biscuitserver [flags]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func init() {
	flag.Usage = usage
	flag.Parse()

	if *version {
		fmt.Println(Name, Version)
		os.Exit(0)
	}

	files, err := filepath.Glob(*dir)
	if err != nil {
		log.Println("Could not load bodies from path:", err)
		os.Exit(1)
	}
	var wg sync.WaitGroup

	log.Printf("LOADING %d MODEL(S) ...", len(files))
	for _, file := range files {
		wg.Add(1)

		go loadModel(file, &wg)
	}

	wg.Wait()
}

func loadModel(file string, wg *sync.WaitGroup) {
	defer wg.Done()

	label := filepath.Base(file)[:2]
	bodies = append(bodies, label)

	model, err := biscuit.NewModelFromFile(label, file, 3)
	if err != nil {
		log.Println("Could not create model from body:", err)
		os.Exit(1)
	}

	log.Println("... loaded model:", label)

	models[label] = model
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var form = `
	<!DOCTYPE html>
	<h2>Language Detector</h2>
	<p>Add some text and select the models to test against.</p>
	<form method="post">
		<textarea rows="10" cols="80" name="text"></textarea><br />
		{{ range .Bodies }}
			<label><input type="checkbox" name="bodies" value="{{ . }}" checked="checked" />{{ . }}</label><br />
		{{ end }}
		<hr />
		<input type="submit" value="Process ..." />
	</form>`

	data := struct {
		Bodies []string
	}{
		bodies,
	}

	t := template.Must(template.New("form").Parse(form))
	t.Execute(w, data)
}

func Process(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	log.Println(r.Form.Get("bodies"))
}

func main() {
	router := httprouter.New()
	router.GET("/", Index)
	router.POST("/", Process)

	log.Printf("Running %s v.%s on %d", Name, Version, *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), router))
}
